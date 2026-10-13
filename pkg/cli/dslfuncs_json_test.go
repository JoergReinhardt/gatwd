package cli

import (
	"math"
	"math/big"
	"strings"
	"testing"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/functions"
)

func justValue(t *testing.T, src string) any {
	t.Helper()
	mb, ok := evalOne(t, NewInterpreter(), src).(functions.Maybe[any])
	if !ok {
		t.Fatalf("%s is not a Maybe", src)
	}
	v, isJust := mb.MaybeValue()
	if !isJust {
		t.Fatalf("%s = Nothing, want a value", src)
	}
	return v
}

func TestJSONParseStructure(t *testing.T) {
	obj, ok := justValue(t, `json_parse "{\"a\": 1, \"b\": [true, null, 2.5, \"x\"], \"c\": {\"d\": \"e\"}}"`).(*collections.HAMT[string, any])
	if !ok {
		t.Fatal("an object must parse to a *HAMT[string, any]")
	}
	a, _ := obj.Get("a")
	if !functions.StructuralEqual(a, 1) || functions.KindName(a) != "Int" {
		t.Errorf("a = %#v, want the Int 1", a)
	}
	b, _ := obj.Get("b")
	items := itemsOf(t, b)
	if len(items) != 4 || items[0] != true || items[1] != nil || items[2] != 2.5 || items[3] != "x" {
		t.Errorf("b = %#v", items)
	}
	c, _ := obj.Get("c")
	d, _ := c.(*collections.HAMT[string, any]).Get("d")
	if d != "e" {
		t.Errorf("c.d = %#v", d)
	}
	if big, ok := justValue(t, `json_parse "12345678901234567890123"`).(*big.Int); !ok || big.String() != "12345678901234567890123" {
		t.Errorf("an integer beyond int64 must be a *big.Int")
	}
	if f := justValue(t, `json_parse "1e3"`); f != 1000.0 {
		t.Errorf("1e3 = %#v, want the Float 1000", f)
	}
	if f := justValue(t, `json_parse "-5"`); !functions.StructuralEqual(f, -5) || functions.KindName(f) != "Int" {
		t.Errorf("-5 = %#v", f)
	}
	if v := justValue(t, `json_parse "null"`); v != nil {
		t.Errorf("null = %#v, want nil (and a Just, not Nothing)", v)
	}
	if v := justValue(t, `json_parse "  [1, 2]  "`); len(itemsOf(t, v)) != 2 {
		t.Errorf("surrounding whitespace is fine, got %#v", v)
	}
}

func TestJSONParseMalformedIsNothing(t *testing.T) {
	in := NewInterpreter()
	for _, bad := range []string{`""`, `"nope"`, `"[1] extra"`, `"[1,]"`, `"{'a': 1}"`, `"{\"a\": }"`, `"[1, 2"`, `"1 2"`, `"\"unterminated"`, `"NaN"`, `"{\"a\":1}}"`} {
		mb, ok := evalOne(t, in, `json_parse `+bad).(functions.Maybe[any])
		if !ok || mb.IsJust() {
			t.Errorf("json_parse %s must be Nothing, got %v", bad, mb)
		}
	}
	if res := in.evaluatePipelineStr(`json_parse 5`); len(res) != 1 || !strings.Contains(res[0], "want a String") {
		t.Errorf("a non-String argument is an error, got %v", res)
	}
}

func TestJSONEncode(t *testing.T) {
	in := NewInterpreter()
	for _, tc := range []struct{ src, want string }{
		{`json_encode (hamt_from [["b", 1], ["a", [1, 2]]])`, `{"a":[1,2],"b":1}`}, // keys sorted
		{`json_encode [1, "a", 'c', true, 2.5, 99999999999999999999]`, `[1,"a","c",true,2.5,99999999999999999999]`},
		{`json_encode "say \"hi\"\n"`, `"say \"hi\"\n"`},
		{`json_encode "é<>&"`, `"é<>&"`},
		{`json_encode []`, `[]`},
		{`json_encode (hamt_from [])`, `{}`},
		{`json_encode (maybe_default 0 (json_parse "null"))`, `null`},
		{`json_encode [[1, 2], [3]]`, `[[1,2],[3]]`},
		{`json_encode (hamt_from [["k", (hamt_from [["n", 1]])]])`, `{"k":{"n":1}}`},
		{`json_pretty (hamt_from [["a", [1, 2]]])`, "{\n  \"a\": [\n    1,\n    2\n  ]\n}"},
		{`is_null (maybe_default 5 (json_parse "null"))`, `true`},
		{`is_null 0`, `false`},
	} {
		if got := displayText(evalOne(t, in, tc.src)); got != tc.want {
			t.Errorf("%s = %q, want %q", tc.src, got, tc.want)
		}
	}
	// parse . encode round trips (normalised: sorted keys, no whitespace)
	for _, doc := range []string{`{"a":[1,2,{"b":null}],"c":"x"}`, `[1,2.5,"s",true,null]`, `"text"`, `12345678901234567890123`} {
		src := `maybe_default 0 (map (\v -> json_encode v) (json_parse "` + strings.ReplaceAll(doc, `"`, `\"`) + `"))`
		if got := displayText(evalOne(t, in, src)); got != doc {
			t.Errorf("round trip of %s = %q", doc, got)
		}
	}
}

func TestJSONEncodeFailsLoud(t *testing.T) {
	in := NewInterpreter()
	for _, tc := range []struct{ src, want string }{
		{`json_encode (pair 1 2)`, "Pair has no JSON form"},
		{`json_encode (utf8_encode "x")`, "base64_encode"},
		{`json_encode (to_set [1])`, "Set has no JSON form"},
		{`json_encode (\x -> x)`, "Function has no JSON form"},
		{`json_encode [1, (pair 1 2)]`, "$[1]: Pair"},
		{`json_encode (hamt_from [["a", (utf8_encode "x")]])`, "$.a: Bytes"},
		{`json_pretty (pair 1 2)`, "no JSON form"},
	} {
		res := in.evaluatePipelineStr(tc.src)
		if len(res) != 1 || !strings.Contains(res[0], "[ERROR]") || !strings.Contains(res[0], tc.want) {
			t.Errorf("%s = %v, want an [ERROR] containing %q", tc.src, res, tc.want)
		}
	}
}

// Parsed JSON is the same shape a resource has, so the existing hamt_path navigation works.
func TestJSONComposesWithHamtPath(t *testing.T) {
	in := NewInterpreter()
	src := `maybe_default "?" (map (\d -> hamt_path "metadata.name" d) (json_parse "{\"metadata\": {\"name\": \"web\"}}"))`
	if got := evalOne(t, in, src); got != "web" {
		t.Errorf("hamt_path over parsed JSON = %#v, want web", got)
	}
}

func TestJSONEncodeRejectsNaNAndInf(t *testing.T) {
	for _, f := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		var b strings.Builder
		if err := encodeJSON(&b, f, "$"); err == nil || !strings.Contains(err.Error(), "no JSON form") {
			t.Errorf("encodeJSON(%v) = %q, %v; want an error", f, b.String(), err)
		}
	}
}

// Object keys are emitted sorted whatever the HAMT's internal order is.
func TestJSONEncodeSortsKeys(t *testing.T) {
	h := collections.NewHAMT[string, any]()
	var want []string
	for _, k := range "zyxwvutsrqponmlkjihgfedcba0123456789" {
		h = h.Set(string(k), 1)
	}
	for _, k := range "0123456789abcdefghijklmnopqrstuvwxyz" {
		want = append(want, `"`+string(k)+`":1`)
	}
	var b strings.Builder
	if err := encodeJSON(&b, h, "$"); err != nil {
		t.Fatal(err)
	}
	if got, exp := b.String(), "{"+strings.Join(want, ",")+"}"; got != exp {
		t.Errorf("keys not sorted:\n got %s\nwant %s", got, exp)
	}
}
