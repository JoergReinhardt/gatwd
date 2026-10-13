package cli

import (
	"math/big"
	"strings"
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

// Every numeric function is a builtin AND a step (F18), and the two readings are the same
// code: `x |> f a` ≡ `f a x` (data-last), `floor x` inside a lambda ≡ `x |> floor`.
func TestNumericStepFormEqualsFunctionForm(t *testing.T) {
	in := NewInterpreter()
	in.EvaluateScript(`def w3_viaLambda = \x -> floor x`)
	rows := []struct{ step, fn string }{
		{`-3.5 |> floor`, `floor -3.5`},
		{`2.1 |> ceil`, `ceil 2.1`},
		{`2.5 |> round`, `round 2.5`},
		{`2.5 |> round_even`, `round_even 2.5`},
		{`-2.9 |> trunc`, `trunc -2.9`},
		{`-4 |> sign`, `sign -4`},
		{`-4 |> abs`, `abs -4`},
		{`4 |> negate`, `negate 4`},
		{`16 |> sqrt`, `sqrt 16`},
		{`17 |> isqrt`, `isqrt 17`},
		{`3 |> to_float`, `to_float 3`},
		{`3.0 |> to_int`, `to_int 3.0`},
		{`3 |> pow 2`, `pow 2 3`},
		{`18 |> gcd 12`, `gcd 12 18`},
		{`6 |> lcm 4`, `lcm 4 6`},
		{`2 |> divmod -7`, `divmod -7 2`},
		{`2 |> quot -7`, `quot -7 2`},
		{`2 |> rem -7`, `rem -7 2`},
		{`1.5 |> is_nan`, `is_nan 1.5`},
		{`1.5 |> is_inf`, `is_inf 1.5`},
	}
	for _, r := range rows {
		step, fn := in.EvaluateToValues(r.step), in.EvaluateToValues(r.fn)
		if len(step) != 1 || len(fn) != 1 {
			t.Errorf("%s -> %v, %s -> %v; want one value each", r.step, step, r.fn, fn)
			continue
		}
		if !functions.StructuralEqual(step[0], fn[0]) {
			t.Errorf("step %s = %#v but function %s = %#v", r.step, step[0], r.fn, fn[0])
		}
	}
	if got := in.EvaluateToValues(`w3_viaLambda -2.5`); len(got) != 1 || got[0] != -3 {
		t.Errorf("a function form inside a lambda = %v, want -3", got)
	}
}

// The numeric functions answer with errors, never silence, and the error names the function.
func TestNumericFunctionErrors(t *testing.T) {
	for stmt, want := range map[string]string{
		`floor "2.5"`:    "floor",
		`sqrt -4`:        "negative",
		`to_int 3.5`:     "fractional",
		`to_float "3"`:   "str_to_float",
		`gcd 1.5 2`:      "two Ints",
		`isqrt -1`:       "negative",
		`rem 1 0`:        "by zero",
		`divmod 1 0`:     "by zero",
		`sign "x"`:       "sign",
		`is_nan "x"`:     "is_nan",
		`abs "x"`:        "abs",
		`abs (3 + 4i)`:   "Complex",
		`negate "x"`:     "negates numbers",
		`1.5 // 0`:       "by zero",
		`10 ** 30 + 1.5`: "", // fine: a huge Int still fits a Float
	} {
		in := NewInterpreter()
		res := in.evaluatePipelineStr(stmt)
		joined := ""
		for _, r := range res {
			joined += r + "\n"
		}
		if want == "" {
			if len(res) == 0 || strings.Contains(joined, "[ERROR]") {
				t.Errorf("%s = %v, want a value", stmt, res)
			}
			continue
		}
		if !strings.Contains(joined, "[ERROR]") || !strings.Contains(joined, want) {
			t.Errorf("%s = %v, want an [ERROR] mentioning %q", stmt, res, want)
		}
	}
}

// F11: a big integer prints as plain digits (no leading '+'), as a result and inside a list,
// and demoted results print as the small integers they are.
func TestBigIntegersPrintWithoutPlus(t *testing.T) {
	in := NewInterpreter()
	for stmt, want := range map[string]string{
		`2 ** 100`:                            "1267650600228229401496703205376",
		`-(2 ** 100)`:                         "-1267650600228229401496703205376",
		`[2 ** 64, 1]`:                        "[18446744073709551616 1]",
		`2 ** 100 - 2 ** 100`:                 "0",
		`9223372036854775807 + 1 - 1`:         "9223372036854775807",
		`str_from (2 ** 70)`:                  "1180591620717411303424",
		`show (2 ** 70)`:                      "1180591620717411303424",
		`18446744073709551615 + 1`:            "18446744073709551616",
		`(18446744073709551615 + 1) // 2**64`: "1",
	} {
		got := in.evaluatePipelineStr(stmt)
		if len(got) != 1 || got[0] != want {
			t.Errorf("%s = %v, want [%s]", stmt, got, want)
		}
	}
	// the memory type: an Int that fits int64 is an int, the next one is a *big.Int
	small := in.EvaluateToValues(`2 ** 63 - 1`)
	if len(small) != 1 || small[0] != int(1<<63-1) {
		t.Errorf("2**63 - 1 = %#v, want the int 9223372036854775807", small)
	}
	large := in.EvaluateToValues(`2 ** 63`)
	if len(large) != 1 {
		t.Fatalf("2**63 = %v", large)
	}
	if _, ok := large[0].(*big.Int); !ok {
		t.Errorf("2**63 is %T, want *big.Int", large[0])
	}
}
