package cli

import (
	"strings"
	"testing"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/functions"
)

// A BST is ordered by comparison, not hashed, and holds the elements themselves: `[3, 1, 2, 10, 9]`
// in order is the Ints 1 2 3 9 10 (the text order would put 10 before 2), in the step form
// (`build_bst`) and the function forms (`bst_from`, `bst_insert`) alike, all one *collections.BST[any].
func TestBSTKeepsValuesStepAndFunctionForm(t *testing.T) {
	in := NewInterpreter()
	step := evalOne(t, in, `spawn S: x |> emit 3, 1, 2, 10, 9 |> drop 1 |> build_bst`)
	fn := evalOne(t, in, `bst_from [3, 1, 2, 10, 9]`)
	viaInsert := evalOne(t, in, `bst_insert 9 (bst_insert 10 (bst_insert 2 (bst_insert 1 (bst_insert 3 bst_empty))))`)
	for name, v := range map[string]any{"build_bst": step, "bst_from": fn, "bst_insert": viaInsert} {
		tree, ok := v.(*collections.BST[any])
		if !ok {
			t.Errorf("%s is a %T, want *collections.BST[interface {}]", name, v)
			continue
		}
		var got []any
		for e := range tree.Seq() {
			got = append(got, e)
		}
		want := []any{1, 2, 3, 9, 10}
		if len(got) != len(want) {
			t.Errorf("%s in order: %#v", name, got)
			continue
		}
		for k := range want {
			if !functions.StructuralEqual(got[k], want[k]) || functions.KindName(got[k]) != "Int" {
				t.Errorf("%s in order %#v, want the Ints %#v", name, got, want)
				break
			}
		}
	}
	if !functions.StructuralEqual(step, fn) {
		t.Error("the step form and the function form must build equal trees")
	}
	// the function forms read the step's tree, and the elements are not text
	in.EvaluateScript("let t = spawn S: x |> emit 10, 9 |> drop 1 |> build_bst")
	for src, want := range map[string]any{
		`bst_member 10 t`: true, `bst_member 9 t`: true, `bst_member 11 t`: false, `length t`: 2,
		`fold (\a x -> a + x) 0 t`: 19, `length (bst_insert 100 t)`: 3,
	} {
		if got := evalOne(t, in, src); got != want {
			t.Errorf("%s = %#v, want %#v", src, got, want)
		}
	}
	// text elements still order as text; a Float and an Int order by value and 1, 1.0 are one member
	if got := itemsOf(t, evalOne(t, in, `to_list (bst_from ["10", "9", "b", "a"])`)); len(got) != 4 || got[0] != "10" || got[1] != "9" {
		t.Errorf("Strings: %#v", got)
	}
	if got := evalOne(t, in, `length (bst_from [1, 1.0, 2.5])`); got != 2 {
		t.Errorf("1 and 1.0 are one member, length %#v", got)
	}
	if got := itemsOf(t, evalOne(t, in, `to_list (bst_from [[2, 1], [1, 5]])`)); len(got) != 2 {
		t.Errorf("Lists order lexicographically: %#v", got)
	}
}

// Incomparable elements are a loud error naming the elements, in every form (Mandate 6).
func TestBSTIncomparableIsAnErrorNamingTheElements(t *testing.T) {
	in := NewInterpreter()
	for src, parts := range map[string][]string{
		`bst_from [1, "a"]`:                              {"bst_from", `"a"`, "BST"},
		`bst_insert "x" (bst_from [1, 2])`:               {"bst_insert", `"x"`},
		`bst_member "10" (bst_from [10, 9])`:             {"bst_member", `"10"`},
		`spawn S: x |> emit 1, a |> drop 1 |> build_bst`: {"build_bst", `"a"`},
		`bst_insert 1 [1, 2]`:                            {"want a BST"},
	} {
		got := in.evaluatePipelineStr(src)
		if len(got) != 1 || !strings.HasPrefix(got[0], "[ERROR] ") {
			t.Errorf("%s = %q, want one error", src, got)
			continue
		}
		for _, p := range parts {
			if !strings.Contains(got[0], p) {
				t.Errorf("%s: %q does not contain %s", src, got[0], p)
			}
		}
	}
}
