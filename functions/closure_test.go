package functions

import (
	"fmt"
	"testing"
)

func sub3() *Closure {
	return NewClosure([]string{"a", "b", "c"}, func(x []any) (any, error) {
		return x[0].(int) - x[1].(int) - x[2].(int), nil
	})
}

// TestClosureCurrying checks every arity split of a 3-ary closure yields the same
// result — the currying isomorphism (A×B×C → D) ≅ A → B → C → D.
func TestClosureCurrying(t *testing.T) {
	splits := [][][]any{
		{{10, 3, 2}},
		{{10}, {3, 2}},
		{{10, 3}, {2}},
		{{10}, {3}, {2}},
	}
	for _, split := range splits {
		var cur any = sub3()
		var err error
		for _, args := range split {
			if cur, err = Apply(cur, args); err != nil {
				t.Fatal(err)
			}
		}
		if cur != 5 {
			t.Errorf("split %v: got %v want 5", split, cur)
		}
	}
}

func TestClosureOverApplication(t *testing.T) {
	adder := NewClosure([]string{"x"}, func(a []any) (any, error) {
		k := a[0].(int)
		return NewClosure([]string{"y"}, func(b []any) (any, error) { return k + b[0].(int), nil }), nil
	})
	if v, err := adder.Apply([]any{1, 2}); err != nil || v != 3 {
		t.Fatalf("got %v %v", v, err)
	}
	flat := NewClosure([]string{"x"}, func(a []any) (any, error) { return a[0], nil })
	if _, err := flat.Apply([]any{1, 2}); err == nil {
		t.Fatal("over-applying a non-function result must error")
	}
}

func TestClosureErrorsAndCallables(t *testing.T) {
	boom := NewClosure([]string{"x"}, func([]any) (any, error) { return nil, fmt.Errorf("boom") })
	if _, err := boom.Apply([]any{1}); err == nil {
		t.Fatal("body error must propagate")
	}
	if !IsCallable(boom) || !IsCallable(DSLFunc(nil)) || IsCallable(3) {
		t.Fatal("IsCallable wrong")
	}
	if _, err := Apply("s", nil); err == nil {
		t.Fatal("applying a string must error")
	}
	if c, _ := boom.Apply(nil); c != boom {
		t.Fatal("zero-arg apply must return the closure")
	}
	if boom.String() != "<fn x>" {
		t.Fatalf("String: %s", boom.String())
	}
}
