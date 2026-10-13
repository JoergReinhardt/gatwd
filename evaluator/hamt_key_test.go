package evaluator

import (
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"testing"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/functions"
)

func keyList(items ...any) *collections.List[any] {
	l := collections.Nil[any]()
	for i := len(items) - 1; i >= 0; i-- {
		l = collections.Cons(items[i], l)
	}
	return l
}

func keyChunk(items ...any) *collections.Chunk[any] {
	c := collections.NewChunk[any](len(items) + 1)
	for _, it := range items {
		c.Append(it)
	}
	return c
}

func mustKey(t *testing.T, v any) string {
	t.Helper()
	k, err := EncodeKey(v)
	if err != nil {
		t.Fatalf("EncodeKey(%#v): %v", v, err)
	}
	return k
}

// keyPool is a spread of values over every key carrier, chosen so that the interesting
// pairs (equal across representations, and different but alike in print) both occur.
func keyPool() map[string]any {
	big1, _ := new(big.Int).SetString("123456789012345678901234567890", 10)
	r13, _ := functions.NewRatio(big.NewInt(1), big.NewInt(3))
	r12, _ := functions.NewRatio(big.NewInt(1), big.NewInt(2))
	r110, _ := functions.NewRatio(big.NewInt(1), big.NewInt(10))
	return map[string]any{
		"string empty":     "",
		"string 1":         "1",
		"string 1.0":       "1.0",
		"string a":         "a",
		"string [1 2]":     "[1 2]",
		"string [1, 2]":    "[1, 2]",
		"string true":      "true",
		"string mark":      "\x00",
		"string mark-l":    "\x00l0:",
		"string mark-i1":   "\x00i1",
		"string mark-s":    "\x00s",
		"char a":           functions.Char('a'),
		"char nul":         functions.Char(0),
		"int 0":            0,
		"int 1":            1,
		"int64 1":          int64(1),
		"uint8 1":          uint8(1),
		"int -1":           -1,
		"float 1.0":        1.0,
		"float 0.5":        0.5,
		"float 0.1":        0.1,
		"float -0.0":       math.Copysign(0, -1),
		"float +Inf":       math.Inf(1),
		"float -Inf":       math.Inf(-1),
		"float 1e30":       1e30,
		"big":              big1,
		"big again":        new(big.Int).Set(big1),
		"ratio 1/3":        r13,
		"ratio 1/2":        r12,
		"ratio 1/10":       r110,
		"complex 1+0i":     complex(1, 0),
		"complex 1+2i":     complex(1, 2),
		"complex 0.5+0i":   complex(0.5, 0),
		"bool true":        true,
		"bool false":       false,
		"list empty":       keyList(),
		"list 1 2":         keyList(1, 2),
		"list 2 1":         keyList(2, 1),
		"list str":         keyList("1", "2"),
		"list 1.0 2":       keyList(1.0, 2),
		"list nested a":    keyList(keyList(1), keyList(2)),
		"list nested b":    keyList(keyList(1, 2)),
		"list nested c":    keyList(keyList(), keyList(1, 2)),
		"list 12":          keyList(12),
		"list 1 str 2":     keyList(1, "2"),
		"list 1:2":         keyList("1:2"),
		"chunk 1 2":        keyChunk(1, 2),
		"slice 1 2":        []any{1, 2},
		"pair 1 2":         functions.Pair[any, any]{Fst: 1, Snd: 2},
		"pair str":         functions.Pair[any, any]{Fst: "1", Snd: "2"},
		"bytes 1 2":        functions.Bytes{1, 2},
		"set 1 2":          collections.NewSet[any](1, 2),
		"set 2 1":          collections.NewSet[any](2, 1),
		"set 1":            collections.NewSet[any](1),
		"set empty":        collections.NewSet[any](),
		"set str":          collections.NewSet[any]("1", "2"),
		"hamt empty":       collections.NewHAMT[string, any](),
		"hamt a1":          collections.NewHAMT[string, any]().Set("a", 1),
		"hamt a1 again":    collections.NewHAMT[string, any]().Set("a", 1),
		"hamt a str":       collections.NewHAMT[string, any]().Set("a", "1"),
		"hamt ab":          collections.NewHAMT[string, any]().Set("a", 1).Set("b", 2),
		"hamt ba":          collections.NewHAMT[string, any]().Set("b", 2).Set("a", 1),
		"tree leaf 1":      collections.Leaf[any](1),
		"tree 1 [2]":       collections.NewTree[any](1, collections.Cons(collections.Leaf[any](2), collections.Nil[*collections.Tree[any]]())),
		"tree 1 [2 3]":     collections.NewTree[any](1, collections.Cons(collections.Leaf[any](2), collections.Cons(collections.Leaf[any](3), collections.Nil[*collections.Tree[any]]()))),
		"tree 1 [[2] 3]":   collections.NewTree[any](1, collections.Cons(collections.NewTree[any](2, collections.Cons(collections.Leaf[any](3), collections.Nil[*collections.Tree[any]]())), collections.Nil[*collections.Tree[any]]())),
		"nothing":          functions.Nothing[any](),
		"just 1":           functions.Just[any](1),
		"just str":         functions.Just[any]("1"),
		"just list":        functions.Just[any](keyList(1)),
		"left 1":           functions.Left[any, any](1),
		"right 1":          functions.Right[any, any](1),
		"variant A":        dsltypes.NewVariant("A", 4, nil),
		"variant A nothng": dsltypes.NewVariant("A", 4, functions.Nothing[any]()),
		"variant A 1":      dsltypes.NewVariant("A", 4, 1),
		"variant B 1":      dsltypes.NewVariant("B", 4, 1),
		"variant A flag 8": dsltypes.NewVariant("A", 8, 1),
	}
}

// TestEncodeKeyAgreesWithEquality is the defining property: two values have the same key
// exactly when `==` (StructuralEqual) says they are equal — over every pair of a pool that
// holds the equal-but-differently-represented and the alike-in-print-but-different cases.
func TestEncodeKeyAgreesWithEquality(t *testing.T) {
	pool := keyPool()
	keys := map[string]string{}
	for name, v := range pool {
		keys[name] = mustKey(t, v)
	}
	same, different := 0, 0
	for an, a := range pool {
		for bn, b := range pool {
			// `==` on Complex goes through doubles, so it is looser than the (exact) key:
			// documented in hamt_key.go; the pool avoids the cases where they part.
			eq := functions.StructuralEqual(a, b)
			ke := keys[an] == keys[bn]
			if eq && !ke && stringVsCollection(a, b) {
				continue // the printed-form fallback of StructuralEqual; removed in step 3
			}
			if eq != ke {
				t.Errorf("%s vs %s: == is %v but the keys %q and %q are %s", an, bn, eq, keys[an], keys[bn],
					map[bool]string{true: "equal", false: "different"}[ke])
			}
			if ke {
				same++
			} else {
				different++
			}
		}
	}
	if same <= len(pool) || different == 0 {
		t.Fatalf("the pool exercised nothing: %d equal pairs, %d different", same, different)
	}
}

// TestEncodeKeyEqualGroups pins the groups that must meet, by name, so a failure says which.
func TestEncodeKeyEqualGroups(t *testing.T) {
	pool := keyPool()
	groups := [][]string{
		{"int 1", "int64 1", "uint8 1", "float 1.0", "complex 1+0i"},
		{"int 0", "float -0.0"},
		{"big", "big again"},
		{"float 0.5", "ratio 1/2", "complex 0.5+0i"},
		{"list 1 2", "chunk 1 2", "slice 1 2", "pair 1 2", "bytes 1 2", "list 1.0 2"},
		{"set 1 2", "set 2 1"},
		{"hamt a1", "hamt a1 again"},
		{"hamt ab", "hamt ba"},
		{"char a", "string a"},
	}
	for _, g := range groups {
		want := mustKey(t, pool[g[0]])
		for _, n := range g[1:] {
			if got := mustKey(t, pool[n]); got != want {
				t.Errorf("%s and %s are == but their keys differ: %q vs %q", g[0], n, want, got)
			}
		}
	}
	// ... and the pairs that must never meet.
	apart := [][2]string{
		{"string [1 2]", "list 1 2"}, {"string [1, 2]", "pair 1 2"}, {"string 1", "int 1"},
		{"string 1.0", "float 1.0"}, {"string true", "bool true"}, {"list str", "list 1 2"},
		{"ratio 1/10", "float 0.1"}, {"set 1 2", "list 1 2"}, {"just 1", "list 12"},
		{"left 1", "right 1"}, {"variant A", "variant A nothng"}, {"string mark-l", "list empty"},
		{"string mark-i1", "int 1"}, {"list nested a", "list nested b"}, {"list 1:2", "list str"},
		{"variant A 1", "variant B 1"}, {"variant A 1", "variant A flag 8"}, {"hamt a1", "hamt a str"},
		{"tree 1 [2 3]", "tree 1 [[2] 3]"}, {"string empty", "list empty"},
	}
	for _, p := range apart {
		if mustKey(t, pool[p[0]]) == mustKey(t, pool[p[1]]) {
			t.Errorf("%s and %s must be different keys", p[0], p[1])
		}
	}
}

// TestKeyRoundTrip: DecodeKey returns a value `==` the original, and that value is a real
// value of the right shape, not text.
func TestKeyRoundTrip(t *testing.T) {
	for name, v := range keyPool() {
		got := DecodeKey(mustKey(t, v))
		if !functions.StructuralEqual(got, v) {
			t.Errorf("%s: DecodeKey(EncodeKey(v)) = %#v, want a value == %#v", name, got, v)
		}
	}
	// the memory types (Mandate: structural tests assert raw types)
	if _, ok := DecodeKey(mustKey(t, functions.Pair[any, any]{Fst: 2, Snd: 1})).(*collections.List[any]); !ok {
		t.Errorf("a Pair key decodes to a *List")
	}
	if _, ok := DecodeKey(mustKey(t, collections.NewSet[any](1))).(*collections.Setoid[any]); !ok {
		t.Errorf("a Set key decodes to the DSL's Set, a *Setoid")
	}
	if _, ok := DecodeKey(mustKey(t, collections.NewHAMT[string, any]().Set("a", 1))).(*collections.HAMT[string, any]); !ok {
		t.Errorf("a HAMT key decodes to a *HAMT")
	}
	if v := DecodeKey(mustKey(t, 1.0)); v != 1 {
		t.Errorf("a whole Float key decodes to the Int, got %#v", v)
	}
	if v := DecodeKey(mustKey(t, 0.5)); v != 0.5 {
		t.Errorf("0.5 decodes to the float, got %#v", v)
	}
	if v, ok := DecodeKey(mustKey(t, "plain")).(string); !ok || v != "plain" {
		t.Errorf("a plain String is its own key, got %#v", v)
	}
	if s := mustKey(t, "plain"); s != "plain" {
		t.Errorf("a String key must stay itself (records from JSON/YAML/manifests), got %q", s)
	}
	// a key that starts with the mark but was not made by EncodeKey is its own text
	if v := DecodeKey("\x00garbage"); v != "\x00garbage" {
		t.Errorf("malformed key decodes to itself, got %#v", v)
	}
}

func TestEncodeKeyErrors(t *testing.T) {
	g := functions.NewClosure([]string{"x"}, func(a []any) (any, error) { return a[0], nil })
	cases := map[string]any{
		"nil":              nil,
		"NaN":              math.NaN(),
		"NaN complex":      complex(math.NaN(), 1),
		"a closure":        g,
		"an error value":   fmt.Errorf("boom"),
		"list holding nil": keyList(1, nil),
		"list of closure":  keyList(g),
		"just of closure":  functions.Just[any](g),
		"a graph":          collections.NewGraph[string, string, any, any](),
		"a map":            collections.NewMap[string, any]().Set("a", 1),
	}
	for name, v := range cases {
		if k, err := EncodeKey(v); err == nil {
			t.Errorf("%s: want an error, got key %q", name, k)
		}
	}
}

// TestHAMTKeysSurviveTheTrie: the whole point — keys carried by a *HAMT come back as the
// values they were, found by any `==` representation, and distinct collections stay
// distinct entries.
func TestHAMTKeysSurviveTheTrie(t *testing.T) {
	h := collections.NewHAMT[string, any]()
	h = h.Set(mustKey(t, keyList(2, 1)), "list")
	h = h.Set(mustKey(t, "[2 1]"), "string")
	h = h.Set(mustKey(t, 1), "int")
	h = h.Set(mustKey(t, "1"), "text")
	if h.Len() != 4 {
		t.Fatalf("a list, its printed form, an Int and its text are four keys, got %d", h.Len())
	}
	if v, _ := h.Get(mustKey(t, functions.Pair[any, any]{Fst: 2, Snd: 1})); v != "list" {
		t.Errorf("a Pair finds the entry of the equal list, got %v", v)
	}
	if v, _ := h.Get(mustKey(t, 1.0)); v != "int" {
		t.Errorf("1.0 finds the entry of 1, got %v", v)
	}
	if v, _ := h.Get(mustKey(t, "[2 1]")); v != "string" {
		t.Errorf("the String finds its own entry, got %v", v)
	}
	for k := range h.Keys() {
		if s, isStr := DecodeKey(k).(string); isStr && k != s {
			t.Errorf("a String key must decode to exactly itself: %q vs %q", k, s)
		}
	}
}

// TestEncodeKeyRandomNested generates nested keys at random and checks the defining
// property on every pair: the keys are equal exactly when the values are `==`.
func TestEncodeKeyRandomNested(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	var gen func(depth int) any
	gen = func(depth int) any {
		n := rng.Intn(9)
		if depth == 0 {
			n = rng.Intn(5)
		}
		switch n {
		case 0:
			return rng.Intn(3)
		case 1:
			return float64(rng.Intn(3))
		case 2:
			return [...]string{"", "a", "1", "[1]", "\x00"}[rng.Intn(5)]
		case 3:
			return rng.Intn(2) == 0
		case 4:
			return [...]float64{0.5, 0.25, 1.5}[rng.Intn(3)]
		}
		k := rng.Intn(3)
		items := make([]any, k)
		for i := range items {
			items[i] = gen(depth - 1)
		}
		switch n {
		case 5:
			return keyList(items...)
		case 6:
			return keyChunk(items...)
		case 7:
			if k == 0 {
				return functions.Nothing[any]()
			}
			return functions.Just[any](items[0])
		}
		s := collections.NewSet[any]()
		for _, it := range items {
			s = s.Add(it)
		}
		return s
	}
	values := make([]any, 70)
	keys := make([]string, len(values))
	for i := range values {
		values[i] = gen(3)
		keys[i] = mustKey(t, values[i])
		if got := DecodeKey(keys[i]); !functions.StructuralEqual(got, values[i]) {
			t.Fatalf("round trip of %#v gave %#v", values[i], got)
		}
	}
	equalPairs := 0
	for i := range values {
		for j := range values {
			eq := functions.StructuralEqual(values[i], values[j])
			if eq != (keys[i] == keys[j]) {
				t.Fatalf("%#v vs %#v: == is %v, keys %q vs %q", values[i], values[j], eq, keys[i], keys[j])
			}
			if eq && i != j {
				equalPairs++
			}
		}
	}
	if equalPairs == 0 {
		t.Fatal("random pool had no equal pair; the test could not fail")
	}
}

// stringVsCollection is a String against a non-scalar: the one pair class on which `==`
// still answered through the printed form (StructuralEqual's fallback, removed in step 3).
func stringVsCollection(a, b any) bool {
	_, aStr := a.(string)
	_, bStr := b.(string)
	return aStr != bStr
}
