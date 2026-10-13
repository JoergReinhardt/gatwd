package cli

import (
	"strings"
	"testing"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/functions"
)

// The DSL's Set is value-typed (user decision 2026-10-13, plan 6.3): element identity is the
// language's `==`, never the printed form, and the step form and the function form build the
// same raw type, a *collections.Setoid[any].

func setOf(t *testing.T, in *Interpreter, src string) *collections.Setoid[any] {
	t.Helper()
	s, ok := evalOne(t, in, src).(*collections.Setoid[any])
	if !ok {
		t.Fatalf("%s: want a *Setoid[any]", src)
	}
	return s
}

// Every carrier a Set can be built from yields the same Set, and a Set built by the step form
// is the Set the function form builds from the same elements.
func TestToSetCarriersAndStepFormAgree(t *testing.T) {
	in := NewInterpreter()
	// elements: Int 1, Float 1.0 (one with it), String "1", two equal lists, a different list
	elems := `1, 1.0, "1", [1, 2], [1, 2], [2, 1]`
	want := setOf(t, in, `to_set [`+elems+`]`)
	if want.Len() != 4 {
		t.Fatalf("to_set over a List has %d members, want 4 (1, \"1\", [1 2], [2 1])", want.Len())
	}
	for name, src := range map[string]string{
		"Set (idempotent)":   `to_set (to_set [` + elems + `])`,
		"HAMT values":        `to_set (hamt_from [["a", 1], ["b", 1.0], ["c", "1"], ["d", [1, 2]], ["e", [1, 2]], ["f", [2, 1]]])`,
		"Tree":               `to_set (tree_node 1 [(tree_leaf 1.0), (tree_leaf "1"), (tree_leaf [1, 2]), (tree_leaf [1, 2]), (tree_leaf [2, 1])])`,
		"Chunk":              `let c = (spawn S: x |> emit 1, 1, 1 |> drop 1 |> to_chunk) in to_set (map (\v -> [v, 2]) c) ++ to_set [1, 1.0, "1", [2, 1]]`,
		"step form (stream)": `let xs = [` + elems + `] in xs |> unchunk |> to_set`,
		"set_add chain":      `fold (\s x -> set_add x s) set_empty [` + elems + `]`,
		"union of halves":    `set_union (to_set [1, 1.0, "1"]) (to_set [[1, 2], [1, 2], [2, 1]])`,
	} {
		got := setOf(t, in, src)
		if name == "Chunk" { // a different shape of input: only the carrier is under test
			if got.Len() != 4 {
				t.Errorf("%s: %d members, want 4", name, got.Len())
			}
			continue
		}
		if !got.Equal(want) {
			t.Errorf("%s: %v members, not the Set of the List form (%d)", name, got.Elements(), want.Len())
		}
	}
}

// The three questions of the user decision, in step form and function form side by side.
func TestSetIdentityStepFormEqualsFunctionForm(t *testing.T) {
	in := NewInterpreter()
	for _, c := range []struct {
		name, step, fn string
		want           any
	}{
		{"1 and \"1\" are two",
			`let xs = [1, "1"] in xs |> unchunk |> to_set |> length`, `length (to_set [1, "1"])`, 2},
		{"1 and 1.0 are one",
			`let xs = [1, 1.0] in xs |> unchunk |> to_set |> length`, `length (to_set [1, 1.0])`, 1},
		{"equal lists are one, a reordered list is another",
			`let xs = [[1, 2], [1, 2], [2, 1]] in xs |> unchunk |> to_set |> length`, `length (to_set [[1, 2], [1, 2], [2, 1]])`, 2},
		{"an instant in two zones is one",
			`let xs = [@2026-10-13T09:30:00Z, @2026-10-13T11:30:00+02:00] in xs |> unchunk |> to_set |> length`,
			`length (to_set [@2026-10-13T09:30:00Z, @2026-10-13T11:30:00+02:00])`, 1},
		{"a list is a member, its printed form is not",
			`let xs = [[1, 2]] in xs |> unchunk |> to_set |> set_contains "[1 2]"`, `set_contains "[1 2]" (to_set [[1, 2]])`, false},
		{"an equal list is a member",
			`let xs = [[1, 2]] in xs |> unchunk |> to_set |> set_contains [1 2]`, `set_contains [1, 2] (to_set [[1, 2]])`, true},
		{"1.0 is a member of a Set holding 1",
			`let xs = [1, 2] in xs |> unchunk |> to_set |> set_contains 1.0`, `set_contains 1.0 (to_set [1, 2])`, true},
		{"the String \"1\" is not",
			`let xs = [1, 2] in xs |> unchunk |> to_set |> set_contains "1"`, `set_contains "1" (to_set [1, 2])`, false},
		{"set_add of an equal member keeps the size",
			`let xs = [[1, 2], 3] in xs |> unchunk |> to_set |> set_add [1 2] |> length`, `length (set_add [1, 2] (to_set [[1, 2], 3]))`, 2},
		{"set_add of a printed list adds a member",
			`let xs = [[1, 2], 3] in xs |> unchunk |> to_set |> set_add "[1 2]" |> length`, `length (set_add "[1 2]" (to_set [[1, 2], 3]))`, 3},
		{"set_remove removes by ==",
			`let xs = [[1, 2], 3] in xs |> unchunk |> to_set |> set_remove [1 2] |> length`, `length (set_remove [1, 2] (to_set [[1, 2], 3]))`, 1},
		{"distinct keeps \"1\" apart from 1 and merges 1.0",
			`let xs = [1, 1.0, "1"] in xs |> unchunk |> distinct |> collect_list |> length`, `length (to_list (to_set [1, 1.0, "1"]))`, 2},
	} {
		step := evalOne(t, in, c.step)
		fn := evalOne(t, in, c.fn)
		if !functions.StructuralEqual(step, c.want) || !functions.StructuralEqual(fn, c.want) || !functions.StructuralEqual(step, fn) {
			t.Errorf("%s: step form = %v, function form = %v, want %v", c.name, step, fn, c.want)
		}
	}
}

// set_union / set_intersect / set_diff, step and function form, over the DSL Set.
func TestSetAlgebraAgreesAcrossForms(t *testing.T) {
	in := NewInterpreter()
	in.EvaluateScript(`let a = (to_set [1, "1", [1, 2]])
let b = (to_set [1.0, [1, 2], 7])`)
	for _, c := range []struct {
		op   string
		want int
	}{{"union", 4}, {"intersect", 2}, {"diff", 1}} {
		fn := setOf(t, in, `set_`+c.op+` a b`)
		step := setOf(t, in, `let xs = [1, "1", [1, 2]] in xs |> unchunk |> to_set |> set_`+c.op+` b`)
		if fn.Len() != c.want || !step.Equal(fn) {
			t.Errorf("set_%s: function form %d members, step form %d, want %d", c.op, fn.Len(), step.Len(), c.want)
		}
	}
	// ++ on two Sets is the union
	if got := setOf(t, in, `a ++ b`); got.Len() != 4 {
		t.Errorf("a ++ b has %d members, want 4", got.Len())
	}
}

// map / filter / bind over a Set stay Sets and use the same equality; member and == agree.
func TestSetFunctorRowAndEquality(t *testing.T) {
	in := NewInterpreter()
	if got := setOf(t, in, `map (\x -> [x, 0]) (to_set [1, 1.0, 2])`); got.Len() != 2 {
		t.Errorf("map over a Set: %d members, want 2", got.Len())
	}
	if got := setOf(t, in, `map (\x -> 0) (to_set [[1], [2], [3]])`); got.Len() != 1 {
		t.Errorf("map collapsing every member: %d members, want 1", got.Len())
	}
	if got := setOf(t, in, `filter (\x -> x != [1]) (to_set [[1], [2], [3]])`); got.Len() != 2 {
		t.Errorf("filter: %d members, want 2", got.Len())
	}
	if got := setOf(t, in, `flatmap (\x -> to_set [x, x]) (to_set [1, 2])`); got.Len() != 2 {
		t.Errorf("bind: %d members, want 2", got.Len())
	}
	for src, want := range map[string]bool{
		`member [1, 2] (to_set [[1, 2]])`:        true,
		`member "[1 2]" (to_set [[1, 2]])`:       false,
		`to_set [1, 2] == to_set [2.0, 1]`:       true,
		`to_set [[1], [2]] == to_set [[2], [1]]`: true,
		`to_set [1] == to_set ["1"]`:             false,
		`[ (set_contains y (to_set [[1, 2]])) | y <- [[1, 2], [2, 1]] ]` + ` == [true, false]`: true,
	} {
		if got := evalOne(t, in, src); got != want {
			t.Errorf("%s = %v, want %v", src, got, want)
		}
	}
	// a Set as a HAMT key is still keyed by value, in any element order
	if got := evalOne(t, in, `hamt_get (to_set [1, 2]) (hamt_from [[(to_set [2, 1.0]), "s"]])`); got != "s" {
		t.Errorf("a Set key, got %v", got)
	}
}

// printedBy runs a script on a fresh interpreter and returns everything it printed.
func printedBy(script string) string {
	in, buf := quietInterpreter()
	in.EvaluateScript(script)
	return buf.String()
}

// Fail loud: an element with no equality, and a step handed a non-Set, are errors that name
// what failed — in the step form and the function form alike.
func TestSetNegativeCases(t *testing.T) {
	const fdef = "def f = \\x -> x\n"
	for src, want := range map[string]string{
		`to_set [(\x -> x)]`:                                  "a Function cannot be a Set member",
		`set_add (\x -> x) set_empty`:                         "a Function cannot be a Set member",
		`set_contains (\x -> x) set_empty`:                    "a Function cannot be a Set member",
		fdef + `let xs = [f] in xs |> unchunk |> to_set`:      "to_set: a Function cannot be a Set member",
		fdef + `let xs = [f, 1] in xs |> unchunk |> distinct`: "distinct: a Function cannot be a Set member",
		`let xs = [1] in xs |> unchunk |> set_contains 1`:     "set_contains: want a Set, got Int",
		`let xs = [1] in xs |> unchunk |> set_add 1`:          "set_add: want a Set, got Int",
		`let xs = [1] in xs |> unchunk |> set_remove 1`:       "set_remove: want a Set, got Int",
		`let xs = [1] in xs |> unchunk |> set_elements`:       "set_elements: want a Set, got Int",
		`set_contains 1 5`:                                    "want a Set",
		`set_union (to_set [1]) 5`:                            "want a Set",
	} {
		out := printedBy(src)
		if !strings.Contains(out, "[ERROR]") || !strings.Contains(out, want) {
			t.Errorf("%s: printed %q, want an [ERROR] containing %q", src, out, want)
		}
	}
}
