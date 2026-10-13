package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/big"
	"sort"
	"strconv"
	"strings"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/evaluator"
	"github.com/joergreinhardt/gatwd/functions"
)

// ---- JSON (hardening W2, tier 2) ------------------------------------------------------
//
// json_parse reads JSON text into the same nested shape every resource has: an object is a
// HAMT (string keys), an array a List, a number an Int (an integer that fits) or a BigInt
// (an integer that does not) or a Float, a string a String, true/false a Bool and null the
// Go nil. Malformed text — or text with anything after the one value — is `Nothing`, never
// an error. json_encode is the reverse; it is deterministic (object keys are sorted) and
// fails loudly on a value JSON cannot hold (a Pair, a closure, NaN, Bytes: use base64_encode).
//
//	json_parse s (Maybe) · json_encode v · json_pretty v (two-space indent) · is_null x

func jsonToValue(v any) (any, error) {
	switch x := v.(type) {
	case nil:
		return nil, nil
	case bool, string:
		return x, nil
	case json.Number:
		if n, ok := functions.ParseNumberLiteral(x.String()); ok {
			if _, isComplex := n.(complex128); !isComplex {
				return n, nil
			}
		}
		f, err := x.Float64()
		if err != nil {
			return nil, fmt.Errorf("number %s is out of range", x)
		}
		return f, nil
	case []any:
		out := make([]any, len(x))
		for i, it := range x {
			c, err := jsonToValue(it)
			if err != nil {
				return nil, err
			}
			out[i] = c
		}
		return listOf(out), nil
	case map[string]any:
		h := collections.NewHAMT[string, any]()
		for k, it := range x {
			c, err := jsonToValue(it)
			if err != nil {
				return nil, err
			}
			h = h.Set(k, c)
		}
		return h, nil
	}
	return nil, fmt.Errorf("unexpected JSON value %T", v)
}

func parseJSON(text string) (any, bool) {
	dec := json.NewDecoder(strings.NewReader(text))
	dec.UseNumber()
	var raw any
	if err := dec.Decode(&raw); err != nil {
		return nil, false
	}
	if _, err := dec.Token(); err != io.EOF {
		return nil, false // trailing data after the value
	}
	v, err := jsonToValue(raw)
	return v, err == nil
}

// encodeJSON writes v as compact JSON, or reports the first value JSON cannot hold.
func encodeJSON(b *strings.Builder, v any, path string) error {
	str := func(s string) {
		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		enc.SetEscapeHTML(false)
		_ = enc.Encode(s)
		b.WriteString(strings.TrimSuffix(buf.String(), "\n"))
	}
	switch x := v.(type) {
	case nil:
		b.WriteString("null")
	case bool:
		b.WriteString(strconv.FormatBool(x))
	case string:
		str(x)
	case functions.Char:
		str(x.String())
	case functions.Time, functions.Date, functions.Duration:
		// A String, never a number and never an error: RFC 3339 for a Time (in its own zone),
		// ISO 8601 for a Date, the printed literal for a Duration (`1h30m`). The decoders do
		// not turn text back into these types; time_parse, date_parse and dur_parse do, on request.
		str(fmt.Sprint(x))
	case *big.Int:
		b.WriteString(x.String())
	case float32, float64:
		f, _ := functions.ToFloat64(x)
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return fmt.Errorf("%s: %v has no JSON form", path, f)
		}
		out, _ := json.Marshal(f)
		b.Write(out)
	case *collections.HAMT[string, any]:
		keys := make([]string, 0, x.Len())
		vals := map[string]any{}
		for k, val := range x.Seq() {
			// an object key is a String; a HAMT keyed by another value (a list, a number) has
			// no JSON form, and its encoded key must not leak out as text
			if _, isText := evaluator.DecodeKey(k).(string); !isText {
				return fmt.Errorf("%s: a HAMT with %s keys has no JSON form (object keys are Strings)", path, functions.KindName(evaluator.DecodeKey(k)))
			}
			keys = append(keys, k)
			vals[k] = val
		}
		sort.Strings(keys)
		b.WriteByte('{')
		for i, k := range keys {
			if i > 0 {
				b.WriteByte(',')
			}
			str(k)
			b.WriteByte(':')
			if err := encodeJSON(b, vals[k], path+"."+k); err != nil {
				return err
			}
		}
		b.WriteByte('}')
	default:
		if functions.IsBytes(v) {
			return fmt.Errorf("%s: Bytes has no JSON form (use base64_encode or hex_encode)", path)
		}
		if functions.IsNumeric(v) {
			if _, isComplex := v.(complex128); !isComplex {
				n, ok := functions.ToBigInt(v)
				if ok {
					b.WriteString(n.String())
					return nil
				}
			}
		}
		if _, isSet := v.(interface{ IsSet() bool }); isSet {
			return fmt.Errorf("%s: a Set has no JSON form (its order is not defined; use to_list and sort)", path)
		}
		items, ok := functions.ToSliceAny(v)
		if !ok || isMapLike(v) {
			return fmt.Errorf("%s: %s has no JSON form", path, functions.KindName(v))
		}
		b.WriteByte('[')
		for i, it := range items {
			if i > 0 {
				b.WriteByte(',')
			}
			if err := encodeJSON(b, it, path+"["+strconv.Itoa(i)+"]"); err != nil {
				return err
			}
		}
		b.WriteByte(']')
	}
	return nil
}

// isMapLike reports a value ToSliceAny would flatten although it is not a plain sequence
// (a Pair is two components, not a two-element array).
func isMapLike(v any) bool {
	_, isPair := v.(interface{ PairElements() (any, any) })
	return isPair
}

func registerJSONFuncs(m map[string]any) {
	defn(m, "json_parse", "s", func(a []any) (any, error) {
		s, err := asString(a[0])
		if err != nil {
			return nil, err
		}
		if v, ok := parseJSON(s); ok {
			return functions.Just[any](v), nil
		}
		return functions.Nothing[any](), nil
	})
	enc := func(a []any) (string, error) {
		var b strings.Builder
		if err := encodeJSON(&b, a[0], "$"); err != nil {
			return "", err
		}
		return b.String(), nil
	}
	defn(m, "json_encode", "v", func(a []any) (any, error) { return enc(a) })
	defn(m, "json_pretty", "v", func(a []any) (any, error) {
		s, err := enc(a)
		if err != nil {
			return nil, err
		}
		var out bytes.Buffer
		if err := json.Indent(&out, []byte(s), "", "  "); err != nil {
			return nil, err
		}
		return out.String(), nil
	})
	defn(m, "is_null", "x", func(a []any) (any, error) { return a[0] == nil, nil })
}
