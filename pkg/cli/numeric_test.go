package cli

import (
	"testing"

	"github.com/joergreinhardt/gatwd/functions"
)

// W3: the F2 rows that used to be errors (`- 5`, `2 ** 3`) are values now, and the
// operator reads the same in a statement, a lambda, a pipeline section and a step.
func TestUnaryMinusAndPowerThroughTheInterpreter(t *testing.T) {
	in := NewInterpreter()
	in.EvaluateScript(`def w3_neg = \x -> -x
def w3_sq = \x -> x ** 2`)
	rows := []struct {
		src  string
		want any
	}{
		{`- 5`, -5},
		{`2 ** 3`, 8},
		{`-2 ** 2`, -4},
		{`abs -3`, 3},
		{`w3_neg 4`, -4},
		{`w3_neg -4`, 4},
		{`map w3_sq [1, 2, 3]`, []any{1, 4, 9}},
		{`5 |> ** 2`, 25},
		{`5 |> math_pow 2`, 25},
		{`5 |> - 2`, 3},
		{`[1, 2, 3] |> map (\x -> -x)`, []any{-1, -2, -3}},
	}
	for _, r := range rows {
		got := in.EvaluateToValues(r.src)
		if len(got) != 1 {
			t.Errorf("%s = %v, want one value", r.src, got)
			continue
		}
		if !functions.StructuralEqual(got[0], r.want) {
			t.Errorf("%s = %#v, want %#v", r.src, got[0], r.want)
		}
	}
}
