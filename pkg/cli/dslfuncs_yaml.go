package cli

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	yaml "go.yaml.in/yaml/v3"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/functions"
)

// ---- YAML (hardening W2, tier 2; the library joined the Clean Core allowlist 2026-10-13) --
//
// yaml_parse reads YAML text into the same nested shape json_parse produces — a mapping is a
// HAMT (string keys), a sequence a List, an int an Int (a BigInt when it does not fit), a float
// a Float, a string a String, a bool a Bool and null the Go nil — in a Maybe; malformed text is
// Nothing, never an error. The reader walks the library's node tree instead of letting it
// decode to Go values, so that nothing is altered silently:
//
//   - a timestamp (`2026-10-13`) and a binary scalar stay the String they were written as;
//   - anchors and aliases are expanded, `<<` merge keys are applied (an explicit key wins);
//   - a scalar with a tag outside the YAML core schema (`!Ref x`), a duplicate mapping key
//     and a complex (non-scalar) key are malformed, i.e. Nothing;
//   - a non-string scalar key (`1: a`) becomes the String it was written as ("1");
//   - a document that expands to more than maxYAMLNodes nodes (an alias bomb) is Nothing.
//
// yaml_encode writes a value as block-style YAML with sorted mapping keys (via the same
// value-to-JSON walk as json_encode, so it fails loudly on the same values), without the final
// newline. yaml_parse_all / yaml_encode_all handle a multi-document stream (`---`).
//
//	yaml_parse s (Maybe) · yaml_parse_all s (Maybe of a List) · yaml_encode v · yaml_encode_all docs

// maxYAMLNodes bounds the size of an expanded document so an alias bomb (`&a [*b, *b]` …) is
// Nothing instead of exhausting memory.
const maxYAMLNodes = 1_000_000

type yamlReader struct{ budget int }

func (r *yamlReader) value(n *yaml.Node) (any, error) {
	if r.budget--; r.budget < 0 {
		return nil, fmt.Errorf("document expands to more than %d nodes", maxYAMLNodes)
	}
	switch n.Kind {
	case yaml.DocumentNode:
		if len(n.Content) != 1 {
			return nil, fmt.Errorf("empty document")
		}
		return r.value(n.Content[0])
	case yaml.AliasNode:
		return r.value(n.Alias)
	case yaml.SequenceNode:
		out := make([]any, len(n.Content))
		for i, c := range n.Content {
			v, err := r.value(c)
			if err != nil {
				return nil, err
			}
			out[i] = v
		}
		return listOf(out), nil
	case yaml.MappingNode:
		return r.mapping(n)
	case yaml.ScalarNode:
		return r.scalar(n)
	}
	return nil, fmt.Errorf("unexpected YAML node kind %d", n.Kind)
}

func (r *yamlReader) mapping(n *yaml.Node) (any, error) {
	h := collections.NewHAMT[string, any]()
	seen := map[string]bool{}
	var merges []*yaml.Node
	for i := 0; i+1 < len(n.Content); i += 2 {
		k, v := n.Content[i], n.Content[i+1]
		if k.Kind == yaml.ScalarNode && k.Tag == "!!merge" {
			merges = append(merges, v)
			continue
		}
		if k.Kind != yaml.ScalarNode {
			return nil, fmt.Errorf("line %d: a mapping key must be a scalar", k.Line)
		}
		if seen[k.Value] {
			return nil, fmt.Errorf("line %d: duplicate mapping key %q", k.Line, k.Value)
		}
		seen[k.Value] = true
		c, err := r.value(v)
		if err != nil {
			return nil, err
		}
		h = h.Set(k.Value, c)
	}
	// Merge keys supply defaults: every explicit key (above) wins; of several merged
	// mappings the first listed wins, as the YAML merge-key type specifies.
	for _, m := range merges {
		sources := []*yaml.Node{m}
		if m.Kind == yaml.SequenceNode {
			sources = m.Content
		}
		for _, s := range sources {
			for s.Kind == yaml.AliasNode {
				s = s.Alias
			}
			if s.Kind != yaml.MappingNode {
				return nil, fmt.Errorf("line %d: a merge key needs a mapping", s.Line)
			}
			sv, err := r.mapping(s)
			if err != nil {
				return nil, err
			}
			src := sv.(*collections.HAMT[string, any])
			for k, v := range src.Seq() {
				if !seen[k] {
					seen[k] = true
					h = h.Set(k, v)
				}
			}
		}
	}
	return h, nil
}

func (r *yamlReader) scalar(n *yaml.Node) (any, error) {
	switch n.ShortTag() {
	case "!!null":
		return nil, nil
	case "!!str", "!!timestamp", "!!binary":
		return n.Value, nil
	case "!!bool":
		var b bool
		if err := n.Decode(&b); err != nil {
			return nil, err
		}
		return b, nil
	case "!!int":
		var i int64
		if err := n.Decode(&i); err == nil {
			return int(i), nil
		}
		if v, ok := functions.ParseNumberLiteral(strings.ReplaceAll(n.Value, "_", "")); ok {
			if _, isComplex := v.(complex128); !isComplex {
				return v, nil
			}
		}
		return nil, fmt.Errorf("line %d: integer %s is out of range", n.Line, n.Value)
	case "!!float":
		// The library resolves an integer too large for 64 bits as a float, losing digits.
		if isIntegerText(n.Value) {
			if v, ok := functions.ParseNumberLiteral(strings.ReplaceAll(n.Value, "_", "")); ok {
				return v, nil
			}
		}
		var f float64
		if err := n.Decode(&f); err != nil {
			return nil, err
		}
		return f, nil
	}
	return nil, fmt.Errorf("line %d: unsupported YAML tag %s", n.Line, n.Tag)
}

// parseYAMLDocuments reads every document of a stream; an empty stream has none.
func parseYAMLDocuments(text string) ([]any, error) {
	dec := yaml.NewDecoder(strings.NewReader(text))
	var docs []any
	r := &yamlReader{budget: maxYAMLNodes}
	for {
		var n yaml.Node
		err := dec.Decode(&n)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return docs, nil
			}
			return nil, err
		}
		v, err := r.value(&n)
		if err != nil {
			return nil, err
		}
		docs = append(docs, v)
	}
}

// yamlNodeOf encodes v as JSON (the one place that decides which values a text format can
// hold), reads that back as a node tree — YAML is a superset of JSON, so the numbers keep
// their exact text — and strips the flow/quoting styles so the library chooses block style
// and quotes a string only where it would otherwise read back as something else (including the YAML 1.1
// boolean words, so the text is safe for older readers too).
func yamlNodeOf(v any) (*yaml.Node, error) {
	var b strings.Builder
	if err := encodeJSON(&b, v, "$"); err != nil {
		return nil, err
	}
	var n yaml.Node
	if err := yaml.Unmarshal([]byte(b.String()), &n); err != nil {
		return nil, err
	}
	var strip func(*yaml.Node)
	strip = func(n *yaml.Node) {
		n.Style = 0
		if n.Kind == yaml.ScalarNode && n.ShortTag() == "!!str" && yaml11Bool[strings.ToLower(n.Value)] {
			n.Style = yaml.DoubleQuotedStyle // YAML 1.1 readers (kubectl, PyYAML) take these words as booleans
		}
		for _, c := range n.Content {
			strip(c)
		}
	}
	strip(&n)
	return n.Content[0], nil
}

var yaml11Bool = map[string]bool{"y": true, "n": true, "yes": true, "no": true, "on": true, "off": true}

func marshalYAML(n *yaml.Node) (string, error) {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(n); err != nil {
		return "", err
	}
	if err := enc.Close(); err != nil {
		return "", err
	}
	return strings.TrimRight(buf.String(), "\n"), nil
}

func registerYAMLFuncs(m map[string]any) {
	defn(m, "yaml_parse", "s", func(a []any) (any, error) {
		s, err := asString(a[0])
		if err != nil {
			return nil, err
		}
		docs, err := parseYAMLDocuments(s)
		if err != nil || len(docs) != 1 {
			return functions.Nothing[any](), nil
		}
		return functions.Just[any](docs[0]), nil
	})
	defn(m, "yaml_parse_all", "s", func(a []any) (any, error) {
		s, err := asString(a[0])
		if err != nil {
			return nil, err
		}
		docs, err := parseYAMLDocuments(s)
		if err != nil {
			return functions.Nothing[any](), nil
		}
		return functions.Just[any](listOf(docs)), nil
	})
	defn(m, "yaml_encode", "v", func(a []any) (any, error) {
		n, err := yamlNodeOf(a[0])
		if err != nil {
			return nil, err
		}
		return marshalYAML(n)
	})
	defn(m, "yaml_encode_all", "docs", func(a []any) (any, error) {
		parts := []string{}
		items, ok := functions.ToSliceAny(a[0])
		if !ok {
			return nil, fmt.Errorf("yaml_encode_all needs a collection of documents, got %s", functions.KindName(a[0]))
		}
		for _, d := range items {
			n, err := yamlNodeOf(d)
			if err != nil {
				return nil, err
			}
			s, err := marshalYAML(n)
			if err != nil {
				return nil, err
			}
			parts = append(parts, s)
		}
		return strings.Join(parts, "\n---\n"), nil
	})
}

// isIntegerText reports a decimal integer literal (optional sign, digits, `_` separators).
func isIntegerText(s string) bool {
	s = strings.TrimLeft(s, "+-")
	if s == "" {
		return false
	}
	for _, r := range s {
		if (r < '0' || r > '9') && r != '_' {
			return false
		}
	}
	return true
}
