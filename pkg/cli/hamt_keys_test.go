package cli

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/functions"
)

// Pins written BEFORE the HAMT key change (hardening, user decision 2026-10-13: a String
// stops equalling the printed form of a collection). They hold on both sides of it.

// conwayDef is the committed pure-DSL Game of Life rule (examples/game_of_life_pure_dsl).
func conwayDef(t *testing.T) string {
	t.Helper()
	src, err := os.ReadFile("../../examples/game_of_life_pure_dsl/game_of_life.gatwd")
	if err != nil {
		t.Fatal(err)
	}
	text := string(src)
	// conway_step (the match wrapper) depends on the whole helper chain defined
	// above it in the file (wrap, offsets, neighbours, tally, births, survivors,
	// next_from, next_gen) — the extraction must start before all of that, not at
	// "def conway_step" itself, or conway_step's own match arm (which calls
	// next_gen) sees an interpreter with next_gen undefined.
	start := strings.Index(text, "type Cell = [Int]")
	end := strings.Index(text, "-- Iterate conway_step")
	if start < 0 || end < 0 {
		t.Fatal("conway_step moved in the example script")
	}
	// the definition spans lines; the interpreter's statement joiner reads it as one
	return text[start:end]
}

func printedCells(t *testing.T, v any) []string {
	t.Helper()
	items, ok := functions.ToSliceAny(v)
	if !ok {
		t.Fatalf("want a list of cells, got %#v", v)
	}
	out := make([]string, len(items))
	for i, it := range items {
		out[i] = fmt.Sprint(it)
	}
	sort.Strings(out)
	return out
}

// The blinker flips between a horizontal and a vertical bar; the glider keeps five cells.
func TestGameOfLifeRulePinned(t *testing.T) {
	in, _ := quietInterpreter()
	in.EvaluateScript(conwayDef(t))
	got := printedCells(t, evalOne(t, in, `[[[2, 1], [2, 2], [2, 3]], 8, 8] |> conway_step`))
	if want := []string{"[1 2]", "[2 2]", "[3 2]"}; strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("blinker after one generation = %v, want %v", got, want)
	}
	got = printedCells(t, evalOne(t, in, `[[[2, 1], [3, 2], [1, 3], [2, 3], [3, 3]], 8, 8] |> conway_step`))
	if want := []string{"[1 2]", "[2 3]", "[2 4]", "[3 2]", "[3 3]"}; strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("glider after one generation = %v, want %v", got, want)
	}
}

// The committed example keeps its output: both invariants hold after 20 and 40 generations, and the clause twin agrees.
func TestGameOfLifeExampleOutputPinned(t *testing.T) {
	in, buf := quietInterpreter()
	if _, err := in.Run(RunOptions{Script: "../../examples/game_of_life_pure_dsl/game_of_life.gatwd"}); err != nil {
		t.Fatal(err)
	}
	var vals []string
	for _, line := range strings.Split(buf.String(), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			vals = append(vals, line)
		}
	}
	if len(vals) == 0 || strings.Contains(buf.String(), "[ERROR]") {
		t.Fatalf("the example printed %q", buf.String())
	}
	if got := strings.Join(vals, "|"); !strings.HasSuffix(got, "∷ 3|∷ 5|∷ 3|∷ 5|∷ [5 5]|∷ true") {
		t.Fatalf("match-form counts 3, 5, clause-twin counts 3, 5 and parity [5 5], true expected at the end, got %q", got)
	}
}

// String keys — every record read from JSON/YAML/manifests — stay exactly themselves.
func TestHAMTStringKeysUnchanged(t *testing.T) {
	in := NewInterpreter()
	for src, want := range map[string]any{
		`hamt_get "a" (hamt_set "a" 1 hamt_empty)`:          1,
		`hamt_has "a" (hamt_set "a" 1 hamt_empty)`:          true,
		`hamt_has "b" (hamt_set "a" 1 hamt_empty)`:          false,
		`hamt_keys (hamt_set "a" 5 hamt_empty)`:             nil,
		`hamt_get_or 0 "zz" (hamt_from [["a", 1]])`:         0,
		`length (hamt_from [["a", 1], ["a", 2], ["b", 3]])`: 2,
		`hamt_get "a" (hamt_from [["a", 1], ["a", 2]])`:     2,
		`hamt_get "ab" (hamt_set "ab" 2 hamt_empty)`:        2,
	} {
		got := evalOne(t, in, src)
		if want != nil && !functions.StructuralEqual(got, want) {
			t.Errorf("%s = %#v, want %#v", src, got, want)
		}
	}
	keys := evalOne(t, in, `hamt_keys (hamt_set "a" 5 hamt_empty)`)
	items, _ := functions.ToSliceAny(keys)
	if len(items) != 1 {
		t.Fatalf("hamt_keys = %#v", keys)
	}
	if s, ok := items[0].(string); !ok || s != "a" {
		t.Errorf("a String key reads back as the String, got %#v", items[0])
	}
}

// ---- step 2: a HAMT key is a value ---------------------------------------------------
//
// hamt_from/hamt_set/hamt_get/hamt_has/hamt_del/hamt_get_or, the step forms, to_hamt and
// frequencies all encode a key with evaluator.EncodeKey, so two keys are one entry exactly
// when `==`, and hamt_keys/hamt_pairs hand back values, not text.

const keyHAMT = `hamt_from [[[1, 2], "list"], ["[1 2]", "string"], [3, "int"], ["3", "text"], [0.5, "half"]]`

func TestHAMTCollectionKeysAreValues(t *testing.T) {
	in := NewInterpreter()
	for src, want := range map[string]any{
		`length (` + keyHAMT + `)`:                    5, // a list, its printed form, 3, "3" and 0.5 are five keys
		`hamt_get [1, 2] (` + keyHAMT + `)`:           "list",
		`hamt_get (pair 1 2) (` + keyHAMT + `)`:       "list", // a Pair is == the list, so the same entry
		`hamt_get "[1 2]" (` + keyHAMT + `)`:          "string",
		`hamt_get 3 (` + keyHAMT + `)`:                "int",
		`hamt_get 3.0 (` + keyHAMT + `)`:              "int", // Int and Float keys are one entry exactly as == says
		`hamt_get "3" (` + keyHAMT + `)`:              "text",
		`hamt_get 1r2 (` + keyHAMT + `)`:              "half", // 1r2 == 0.5
		`hamt_has [2, 1] (` + keyHAMT + `)`:           false,
		`hamt_has "[2 1]" (` + keyHAMT + `)`:          false,
		`hamt_get_or "none" 4 (` + keyHAMT + `)`:      "none",
		`length (hamt_del [1, 2] (` + keyHAMT + `))`:  4,
		`length (hamt_del "[1 2]" (` + keyHAMT + `))`: 4,
	} {
		if got := evalOne(t, in, src); !functions.StructuralEqual(got, want) || functions.KindName(got) != functions.KindName(want) {
			t.Errorf("%s = %#v, want %#v", src, got, want)
		}
	}
	// deleting the list key leaves the String key of the same print (and the reverse)
	if got := evalOne(t, in, `hamt_has "[1 2]" (hamt_del [1, 2] (`+keyHAMT+`))`); got != true {
		t.Errorf("deleting the list key must keep the String key, got %v", got)
	}
	if got := evalOne(t, in, `hamt_has [1, 2] (hamt_del "[1 2]" (`+keyHAMT+`))`); got != true {
		t.Errorf("deleting the String key must keep the list key, got %v", got)
	}
	// 1 and "1" never meet either
	if got := evalOne(t, in, `length (hamt_from [[1, "a"], ["1", "b"], [1.0, "c"]])`); got != 2 {
		t.Errorf("Int, String and an equal Float are two keys, got %v", got)
	}
}

// hamt_keys / hamt_pairs return the keys as the values they were (raw memory types).
func TestHAMTKeysReadBackAsValues(t *testing.T) {
	in := NewInterpreter()
	keys, _ := functions.ToSliceAny(evalOne(t, in, `hamt_keys (hamt_from [[[1, 2], 1]])`))
	if len(keys) != 1 {
		t.Fatalf("keys = %#v", keys)
	}
	if _, ok := keys[0].(*collections.List[any]); !ok {
		t.Errorf("a list key reads back as a *List, got %T", keys[0])
	}
	if got := evalOne(t, in, `map (\k -> k == [1, 2]) (hamt_keys (hamt_from [[[1, 2], 1]]))`); !functions.StructuralEqual(got, []any{true}) {
		t.Errorf("the key read back is == the list it was made from, got %v", got)
	}
	ints, _ := functions.ToSliceAny(evalOne(t, in, `hamt_keys (hamt_from [[7, 1]])`))
	if len(ints) != 1 || functions.KindName(ints[0]) != "Int" {
		t.Errorf("an Int key reads back as an Int, got %#v", ints)
	}
	// pairs feed back in: hamt_from (hamt_pairs h) is h
	if got := evalOne(t, in, `length (hamt_from (hamt_pairs (`+keyHAMT+`)))`); got != 5 {
		t.Errorf("hamt_from (hamt_pairs h) keeps all five keys, got %v", got)
	}
	if got := evalOne(t, in, `hamt_get [1, 2] (hamt_from (hamt_pairs (`+keyHAMT+`)))`); got != "list" {
		t.Errorf("a round-tripped HAMT finds the list key, got %v", got)
	}
}

// frequencies keys the stream's own elements; Game of Life reads them back as coordinates.
func TestFrequenciesKeysAreValues(t *testing.T) {
	in := NewInterpreter()
	const counted = `let nb = [[x, y] | [x, y] <- [[2, 1], [2, 1], [3, 3]]] in let m = nb |> unchunk |> frequencies in `
	if got := evalOne(t, in, counted+`hamt_get [2, 1] m`); got != 2 {
		t.Errorf("frequencies counts the coordinate list, got %v", got)
	}
	if got := evalOne(t, in, counted+`hamt_has "[2 1]" m`); got != false {
		t.Errorf("the printed form is not the key, got %v", got)
	}
	pairs, _ := functions.ToSliceAny(evalOne(t, in, counted+`m |> hamt_pairs |> collect_list`))
	if len(pairs) != 2 {
		t.Fatalf("pairs = %#v", pairs)
	}
	for _, p := range pairs {
		a, _ := p.(interface{ PairElements() (any, any) }).PairElements()
		if _, ok := a.(*collections.List[any]); !ok {
			t.Errorf("a frequencies key is the element (a *List), got %T", a)
		}
	}
}

// The Game of Life rule now compares real coordinates: its births are lists, not text.
func TestGameOfLifeBirthsAreCells(t *testing.T) {
	in, _ := quietInterpreter()
	in.EvaluateScript(conwayDef(t))
	cells, _ := functions.ToSliceAny(evalOne(t, in, `[[[2, 1], [2, 2], [2, 3]], 8, 8] |> conway_step`))
	if len(cells) != 3 {
		t.Fatalf("cells = %#v", cells)
	}
	for _, c := range cells {
		if _, ok := c.(*collections.List[any]); !ok {
			t.Errorf("a cell is a coordinate *List, got %T (%v)", c, c)
		}
	}
}

func TestHAMTStepFormKeysMatchFunctionForms(t *testing.T) {
	in := NewInterpreter()
	for src, want := range map[string]any{
		`hamt_empty |> hamt_set 5 five |> hamt_get 5`:                              "five",
		`hamt_empty |> hamt_set "5" five |> hamt_get "5"`:                          "five",
		`hamt_empty |> hamt_set a five |> hamt_get a`:                              "five",
		`hamt_get 5 (hamt_set 5 "five" hamt_empty)`:                                "five",
		`hamt_empty |> hamt_set 5 five |> hamt_del 5 |> hamt_keys |> collect_list`: nil,
	} {
		got := evalOne(t, in, src)
		if want != nil && !functions.StructuralEqual(got, want) {
			t.Errorf("%s = %#v, want %#v", src, got, want)
		}
	}
	// the Int key 5 and the String key "5" are different entries in both forms
	for _, src := range []string{
		`hamt_empty |> hamt_set 5 five |> hamt_get "5"`,
		`hamt_empty |> hamt_set "5" five |> hamt_get 5`,
		`hamt_get "5" (hamt_set 5 "five" hamt_empty)`,
	} {
		res := in.EvaluateToValues(src)
		if len(res) != 1 || !strings.Contains(fmt.Sprint(res[0]), "[ERROR]") && !strings.Contains(fmt.Sprint(res[0]), "not found") {
			t.Errorf("%s = %v, want a not-found error", src, res)
		}
	}
}

func TestHAMTKeyErrorsAreLoud(t *testing.T) {
	in := NewInterpreter()
	for src, want := range map[string]string{
		`hamt_set (\x -> x) 1 hamt_empty`:           "cannot be a HAMT key",
		`hamt_from [[(\x -> x), 1]]`:                "cannot be a HAMT key",
		`hamt_get [1, (\x -> x)] hamt_empty`:        "cannot be a HAMT key",
		`hamt_has (\x -> x) hamt_empty`:             "cannot be a HAMT key",
		`graph_add_node "c" [1, 2] graph_empty`:     "graph id or color must be a String",
		`graph_has_node [1, 2] graph_empty`:         "graph id or color must be a String",
		`graph_add_edge "c" 1 [1] "b" graph_empty`:  "graph id or color must be a String",
		`graph_add_node (pair 1 2) "n" graph_empty`: "graph id or color must be a String",
	} {
		res := in.EvaluateToValues(src)
		if len(res) != 1 {
			t.Errorf("%s = %v, want one error", src, res)
			continue
		}
		if _, isErr := res[0].(error); !isErr || !strings.Contains(fmt.Sprint(res[0]), want) {
			t.Errorf("%s = %v, want an error containing %q", src, res[0], want)
		}
	}
	// a number or a String names a node, as before
	if got := evalOne(t, in, `graph_nodes (graph_add_node "c" 7 graph_empty)`); !functions.StructuralEqual(got, []any{"7"}) {
		t.Errorf("a number id is its text, got %#v", got)
	}
}

// A HAMT whose keys are not Strings has no JSON object form; the encoded key never leaks as text.
func TestJSONRefusesNonStringHAMTKeys(t *testing.T) {
	in := NewInterpreter()
	for _, src := range []string{`json_encode (hamt_from [[[1, 2], 1]])`, `json_encode (hamt_from [[7, 1]])`} {
		res := in.EvaluateToValues(src)
		if len(res) != 1 {
			t.Fatalf("%s = %v", src, res)
		}
		if err, ok := res[0].(error); !ok || !strings.Contains(err.Error(), "keys has no JSON form") && !strings.Contains(err.Error(), "keys have no JSON form") {
			t.Errorf("%s = %v, want a loud error", src, res[0])
		}
	}
	if got := evalOne(t, in, `json_encode (hamt_from [["a", 1]])`); got != `{"a":1}` {
		t.Errorf("String keys still encode, got %v", got)
	}
}

// ---- step 3: a String never equals the printed form of a collection --------------------
//
// The W3 operator matrix (functions.TestNumericOperatorMatrix) covers String x number; this
// is its String x carrier half, built here because it needs every collection type.

func stringVsCarrierValues() map[string]any {
	chunk := collections.NewChunk[any](2)
	chunk.Append(1)
	chunk.Append(2)
	tree := collections.NewTree[any](1, collections.Cons(collections.Leaf[any](2), collections.Nil[*collections.Tree[any]]()))
	return map[string]any{
		"List":       listOf([]any{1, 2}),
		"EmptyList":  listOf(nil),
		"Chunk":      chunk,
		"Slice":      []any{1, 2},
		"Set":        collections.NewSet[any](1, 2),
		"HAMT":       collections.NewHAMT[string, any]().Set("a", 1),
		"Tree":       tree,
		"Pair":       functions.Pair[any, any]{Fst: 1, Snd: 2},
		"Just":       functions.Just[any](1),
		"Nothing":    functions.Nothing[any](),
		"Left":       functions.Left[any, any](1),
		"Right":      functions.Right[any, any](1),
		"Bytes":      functions.Bytes{1, 2},
		"Int":        12,
		"Float":      1.5,
		"Bool":       true,
		"Char":       functions.Char('x'),
		"StringText": "text",
	}
}

func TestStringVersusEveryCarrierIsNeverEqual(t *testing.T) {
	for name, v := range stringVsCarrierValues() {
		printed := fmt.Sprintf("%v", v)
		for _, s := range []string{printed, fmt.Sprint(v), "[1 2]", "[1, 2]", "", "true", "12", "x"} {
			if name == "StringText" || (name == "Char" && s == "x") {
				continue // String == String, and the D1 one-character rule, are the intended equalities
			}
			if functions.StructuralEqual(s, v) || functions.StructuralEqual(v, s) {
				t.Errorf("%s %#v == String %q must be false", name, v, s)
			}
		}
		// ordering across kinds is an error, never a comparison of printed forms
		if _, err := functions.Compare(printed, v); err == nil && name != "Char" && name != "StringText" {
			t.Errorf("Compare(String %q, %s) must be an error", printed, name)
		}
	}
	// the equalities that remain
	for _, tc := range [][2]any{{"[2 1]", "[2 1]"}, {"", ""}, {"x", functions.Char('x')}} {
		if !functions.StructuralEqual(tc[0], tc[1]) || !functions.StructuralEqual(tc[1], tc[0]) {
			t.Errorf("%#v == %#v must stay true", tc[0], tc[1])
		}
	}
}

// Through the interpreter, inside lambdas (function forms) and as the infix operator.
func TestStringVersusCollectionThroughTheInterpreter(t *testing.T) {
	in := NewInterpreter()
	for src, want := range map[string]any{
		`"[2 1]" == [2, 1]`: false, `[2, 1] == "[2 1]"`: false, `"[2 1]" != [2, 1]`: true,
		`"[2 1]" == "[2 1]"`: true, `[2, 1] == [2, 1]`: true,
		`map (\c -> c == "[2 1]") [[2, 1], "[2 1]"]`:                      []any{false, true},
		`map (\c -> "[2, 1]" == c) [(pair 2 1), [2, 1]]`:                  []any{false, false},
		`map (\c -> c == "{a: 1}") [(hamt_from [["a", 1]])]`:              []any{false},
		`map (\c -> c == "{1 2}") [(to_set [1, 2])]`:                      []any{false},
		`map (\c -> c == "Just(1)") [(just 1)]`:                           []any{false},
		`map (\c -> member c ["[1 2]"]) [[1, 2], "[1 2]"]`:                []any{false, true},
		`length (filter (\c -> c == [1, 2]) [[1, 2], "[1 2]", [1.0, 2]])`: 2,
	} {
		got := evalOne(t, in, src)
		if !functions.StructuralEqual(got, want) {
			t.Errorf("%s = %#v, want %#v", src, got, want)
		}
	}
}
