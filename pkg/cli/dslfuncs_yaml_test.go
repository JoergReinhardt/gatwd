package cli

import (
	"math/big"
	"strings"
	"testing"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/functions"
)

func isNothing(t *testing.T, src string) bool {
	t.Helper()
	mb, ok := evalOne(t, NewInterpreter(), src).(functions.Maybe[any])
	if !ok {
		t.Fatalf("%s is not a Maybe", src)
	}
	_, isJust := mb.MaybeValue()
	return !isJust
}

func TestYAMLParseStructure(t *testing.T) {
	doc := "a: 1\nb: [true, null, 2.5, x]\nc:\n  d: e\nbig: 12345678901234567890123\nhex: 0x1F\n"
	obj, ok := justValue(t, `yaml_parse "`+strings.ReplaceAll(doc, "\n", `\n`)+`"`).(*collections.HAMT[string, any])
	if !ok {
		t.Fatal("a mapping must parse to a *HAMT[string, any]")
	}
	if a, _ := obj.Get("a"); !functions.StructuralEqual(a, 1) || functions.KindName(a) != "Int" {
		t.Errorf("a = %#v, want the Int 1", a)
	}
	b, _ := obj.Get("b")
	if items := itemsOf(t, b); len(items) != 4 || items[0] != true || items[1] != nil || items[2] != 2.5 || items[3] != "x" {
		t.Errorf("b = %#v", items)
	}
	c, _ := obj.Get("c")
	if d, _ := c.(*collections.HAMT[string, any]).Get("d"); d != "e" {
		t.Errorf("c.d = %#v", d)
	}
	if v, _ := obj.Get("big"); func() bool { b, ok := v.(*big.Int); return !ok || b.String() != "12345678901234567890123" }() {
		t.Errorf("an integer beyond int64 must be a *big.Int, got %#v", v)
	}
	if v, _ := obj.Get("hex"); !functions.StructuralEqual(v, 31) {
		t.Errorf("0x1F = %#v, want 31", v)
	}
}

// Nothing is altered silently: a timestamp and a quoted number keep their text; anchors,
// aliases and merge keys expand, an explicit key beating a merged one.
func TestYAMLParseKeepsTextAndExpandsAliases(t *testing.T) {
	obj := justValue(t, `yaml_parse "when: 2026-10-13\nq: \"5\"\nbase: &b {x: 1, y: 2}\nuse:\n  <<: *b\n  y: 9\n"`).(*collections.HAMT[string, any])
	if v, _ := obj.Get("when"); v != "2026-10-13" {
		t.Errorf("an unquoted date must stay the text it was written as, got %#v", v)
	}
	if v, _ := obj.Get("q"); v != "5" {
		t.Errorf("a quoted number is a String, got %#v", v)
	}
	use, _ := obj.Get("use")
	u := use.(*collections.HAMT[string, any])
	if x, _ := u.Get("x"); !functions.StructuralEqual(x, 1) {
		t.Errorf("merged x = %#v, want 1", x)
	}
	if y, _ := u.Get("y"); !functions.StructuralEqual(y, 9) {
		t.Errorf("explicit y = %#v, want 9 (explicit keys beat merged ones)", y)
	}
	if k := justValue(t, `yaml_parse "1: a"`).(*collections.HAMT[string, any]); func() bool { _, ok := k.Get("1"); return !ok }() {
		t.Error(`a non-string scalar key becomes the String it was written as ("1")`)
	}
}

// Malformed or unsupported text is Nothing, never an error and never a partial value.
func TestYAMLParseNothing(t *testing.T) {
	bomb := "a: &a [x, x, x, x, x, x, x, x, x, x]\n"
	prev := "a"
	for i := 'b'; i < 'k'; i++ {
		n := string(i)
		bomb += n + ": &" + n + " [*" + prev + ", *" + prev + ", *" + prev + ", *" + prev + ", *" + prev + ", *" + prev + ", *" + prev + ", *" + prev + ", *" + prev + ", *" + prev + "]\n"
		prev = n
	}
	rows := map[string]string{
		"unbalanced":      `yaml_parse "a: [1, 2"`,
		"tab indentation": `yaml_parse "a:\n\tb: 1"`,
		"duplicate key":   `yaml_parse "a: 1\na: 2"`,
		"custom tag":      `yaml_parse "a: !Ref x"`,
		"complex key":     `yaml_parse "? [1, 2]\n: v"`,
		"merge a scalar":  `yaml_parse "a: {<<: 5}"`,
		"empty":           `yaml_parse ""`,
		"two documents":   `yaml_parse "a: 1\n---\nb: 2"`,
		"alias bomb":      `yaml_parse "` + strings.ReplaceAll(bomb, "\n", `\n`) + `"`,
	}
	for name, src := range rows {
		if !isNothing(t, src) {
			t.Errorf("%s: %s must be Nothing", name, src)
		}
	}
	if isNothing(t, `yaml_parse "a: 1"`) {
		t.Error("negative control: a plain mapping must be a Just")
	}
}

func TestYAMLParseAllAndEncode(t *testing.T) {
	docs := itemsOf(t, justValue(t, `yaml_parse_all "a: 1\n---\n- 2\n---\nx"`))
	if len(docs) != 3 {
		t.Fatalf("three documents, got %d", len(docs))
	}
	if v := itemsOf(t, justValue(t, `yaml_parse_all ""`)); len(v) != 0 {
		t.Errorf("an empty stream has no documents, got %#v", v)
	}
	enc := func(src string) string {
		t.Helper()
		s, ok := evalOne(t, NewInterpreter(), src).(string)
		if !ok {
			t.Fatalf("%s did not yield a String", src)
		}
		return s
	}
	if got := enc(`yaml_encode (hamt_from [["b", 1], ["a", [1, "x"]]])`); got != "a:\n  - 1\n  - x\nb: 1" {
		t.Errorf("block style, sorted keys, no final newline; got %q", got)
	}
	if got := enc(`yaml_encode "5"`); got != `"5"` {
		t.Errorf("a String that reads back as a number must be quoted, got %q", got)
	}
	if got := enc(`yaml_encode "yes"`); got != `"yes"` {
		t.Errorf("YAML 1.1 readers take yes as a boolean, so it must be quoted; got %q", got)
	}
	if got := enc(`yaml_encode_all [[1], "x"]`); got != "- 1\n---\nx" {
		t.Errorf("documents joined by ---, got %q", got)
	}
}

func TestYAMLRoundTrip(t *testing.T) {
	in := justValue(t, `yaml_parse (yaml_encode (hamt_from [["s", "yes"], ["n", "007"], ["t", "true"], ["e", ""], ["l", [1, [2, 3]]], ["f", 2.5]]))`)
	obj := in.(*collections.HAMT[string, any])
	for k, want := range map[string]any{"s": "yes", "n": "007", "t": "true", "e": "", "f": 2.5} {
		if got, _ := obj.Get(k); got != want {
			t.Errorf("round trip %s = %#v, want %#v", k, got, want)
		}
	}
}

// The encoder reports what a text format cannot hold, exactly as json_encode does.
func TestYAMLEncodeFailsLoud(t *testing.T) {
	in := NewInterpreter()
	for _, tc := range []struct{ src, want string }{
		{`yaml_encode (pair 1 2)`, "Pair has no JSON form"},
		{`yaml_encode [1, (\x -> x)]`, "$[1]"},
		{`yaml_encode_all 5`, "collection of documents"},
		{`yaml_encode_all [(pair 1 2)]`, "Pair"},
	} {
		res := in.evaluatePipelineStr(tc.src)
		if len(res) != 1 || !strings.Contains(res[0], "[ERROR]") || !strings.Contains(res[0], tc.want) {
			t.Errorf("%s = %v, want an [ERROR] containing %q", tc.src, res, tc.want)
		}
	}
}
