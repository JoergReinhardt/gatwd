package cli

import (
	"strings"
	"testing"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/functions"
)

func evalOne(t *testing.T, i *Interpreter, src string) any {
	t.Helper()
	vals := i.EvaluateToValues(src)
	if len(vals) != 1 {
		t.Fatalf("%s: want one value, got %v", src, vals)
	}
	return vals[0]
}

// TestMatrixCarrierPreservation asserts each Functor/Filterable function form returns
// the SAME carrier it was given, with element types untouched (ints stay ints).
func TestMatrixCarrierPreservation(t *testing.T) {
	i := NewInterpreter()
	if _, ok := evalOne(t, i, `map (\x -> x + 1) (to_set [1, 2])`).(*collections.Setoid[any]); !ok {
		t.Error("map over a Set must yield a *Setoid[any]")
	}
	if _, ok := evalOne(t, i, `map (\x -> x + 1) (hamt_from [["a", 1]])`).(*collections.HAMT[string, any]); !ok {
		t.Error("map over a HAMT must yield a *HAMT")
	}
	if _, ok := evalOne(t, i, `map (\x -> x + 1) (tree_leaf 1)`).(*collections.Tree[any]); !ok {
		t.Error("map over a Tree must yield a *Tree")
	}
	if _, ok := evalOne(t, i, `map (\x -> x + 1) [1, 2]`).(*collections.List[any]); !ok {
		t.Error("map over a List must yield a *List")
	}
	if p, ok := evalOne(t, i, `map (\x -> x + 1) (pair "k" 1)`).(functions.Pair[any, any]); !ok || p.Fst != "k" {
		t.Errorf("map over a Pair must keep fst and yield a Pair, got %v", p)
	}
	if mb, ok := evalOne(t, i, `map (\x -> x + 1) (just 1)`).(functions.Maybe[any]); !ok || !mb.IsJust() {
		t.Errorf("map over a Just must yield a Just, got %v", mb)
	}
	if mb, ok := evalOne(t, i, `map (\x -> x + 1) nothing`).(functions.Maybe[any]); !ok || !mb.IsNothing() {
		t.Errorf("map over Nothing must stay Nothing, got %v", mb)
	}
	// element types preserved: the set holds ints, not strings.
	s := evalOne(t, i, `to_set [1, 2, 3]`).(*collections.Setoid[any])
	for _, e := range s.ElementsAny() {
		if _, isStr := e.(string); isStr {
			t.Errorf("set element %v was stringified", e)
		}
	}
}

// TestMatrixVariantConstructor proves `variant` boxes a declared tag with its branch
// flag (exactly like spawn) and rejects an undeclared one.
func TestMatrixVariantConstructor(t *testing.T) {
	i := NewInterpreter()
	i.EvaluateLine(`type Shape = Circle | Square`)
	v, ok := evalOne(t, i, `variant "Circle" 3`).(dsltypes.Variant)
	if !ok || v.Tag != "Circle" || v.Flag == 0 {
		t.Fatalf("want a flagged Circle variant, got %#v", v)
	}
	if !functions.StructuralEqual(v.Value, 3) {
		t.Errorf("payload %v", v.Value)
	}
	out := i.evaluatePipelineStr(`variant "Triangle" 1`)
	if len(out) != 1 || !strings.Contains(out[0], "not a constructor") {
		t.Errorf("undeclared tag must error loudly, got %v", out)
	}
}

// TestMatrixCurrying shows fixed-arity builtins partially apply, including the
// pre-existing graph/HAMT row.
func TestMatrixCurrying(t *testing.T) {
	i := NewInterpreter()
	c, ok := evalOne(t, i, `set_add 3`).(*functions.Closure)
	if !ok || c.Arity() != 1 {
		t.Fatalf("set_add 3 must be a 1-ary closure, got %T", c)
	}
	h, ok := evalOne(t, i, `hamt_get "a"`).(*functions.Closure)
	if !ok || h.Arity() != 1 {
		t.Fatalf("hamt_get \"a\" must be a 1-ary closure, got %T", h)
	}
}

// TestMatrixErrorsAreValues checks bad calls surface as errors, not panics or no-ops.
func TestMatrixErrorsAreValues(t *testing.T) {
	i := NewInterpreter()
	for _, src := range []string{
		`map 3 [1, 2]`,                // non-callable
		`hamt_set "a" 1 [1, 2]`,       // wrong carrier
		`nth 9 [1, 2]`,                // out of range
		`head []`,                     // empty
		`tree_node 1 [2]`,             // children must be trees
		`filter (\x -> x + 1) [1, 2]`, // predicate must be bool
		`variant_tag 3`,               // not a variant
		`bimap (\a -> a) (\b -> b) 3`, // not a pair/either
		`range "a" 3`,                 // not an integer
		`compare (pair 1 2) 3`,        // incomparable
	} {
		out := i.evaluatePipelineStr(src)
		if len(out) != 1 || !strings.HasPrefix(out[0], "[ERROR]") {
			t.Errorf("%s: want an [ERROR] result, got %v", src, out)
		}
	}
}

// TestUnkeyableSetElementIsAnError: a Set member needs the language's equality, which a
// function has none of; the Set function forms report that as an error value and name the kind.
func TestUnkeyableSetElementIsAnError(t *testing.T) {
	fn := functions.NewClosure([]string{"x"}, func(a []any) (any, error) { return a[0], nil })
	if _, err := newAnySet([]any{fn}); err == nil || !strings.Contains(err.Error(), "cannot be a Set member") {
		t.Fatalf("a function as a Set member: err = %v, want one naming the Set member", err)
	}
	// A list is no longer "unhashable": it is a member by value.
	if s, err := newAnySet([]any{[]any{1}, []any{1}}); err != nil || s.Len() != 1 {
		t.Fatalf("two equal lists: Len = %v, err = %v, want one member", s.Len(), err)
	}
}
