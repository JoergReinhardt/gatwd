package cli

import (
	"fmt"
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

// D2, finished: a String is never a number — not in `==`, and not in the helpers that used to
// parse numeric text (`nth "1" xs`, `range "0" "2"`, `min "5" 3`, `compare "5" 3`).
func TestStringsAreNotNumbersInFunctionsAndEquality(t *testing.T) {
	in := NewInterpreter()
	for stmt, want := range map[string]string{
		`"5" == 5`:                              "false",
		`5 == "5"`:                              "false",
		`"5" != 5`:                              "true",
		`"10" == "10.0"`:                        "false",
		`"5" == "5"`:                            "true",
		`'a' == "a"`:                            "true",
		`maybe_default 0 (str_to_int "5") == 5`: "true",
	} {
		if got := in.evaluatePipelineStr(stmt); len(got) != 1 || got[0] != want {
			t.Errorf("%s = %v, want [%s]", stmt, got, want)
		}
	}
	for stmt, want := range map[string]string{
		`nth "1" [7, 8, 9]`: "str_to_int",
		`range "0" "2"`:     "str_to_int",
		`min "5" 3`:         "str_to_int",
		`max 3 "5"`:         "str_to_int",
		`compare "5" 3`:     "str_to_int",
		`floor "2.5"`:       "String",
		`abs "5"`:           "String",
	} {
		got := in.evaluatePipelineStr(stmt)
		if len(got) != 1 || !strings.Contains(got[0], "[ERROR]") || !strings.Contains(got[0], want) {
			t.Errorf("%s = %v, want an [ERROR] mentioning %q", stmt, got, want)
		}
	}
	// the legitimate spelling still works
	if got := in.EvaluateToValues(`nth (maybe_default 0 (str_to_int "1")) [7, 8, 9]`); len(got) != 1 || !functions.StructuralEqual(got[0], 8) {
		t.Errorf("nth with a parsed index = %v, want [8]", got)
	}
	// A String equals only a String: not a number, not a Char beyond the D1 one-character
	// rule, and not the printed form of a collection (HAMT keys carry values now).
	for stmt, want := range map[string]string{
		`"[2 1]" == [2, 1]`: "false", `[2, 1] == "[2 1]"`: "false", `"[2 1]" != [2, 1]`: "true",
		`"[2 1]" == "[2 1]"`: "true", `"5" == 5`: "false",
	} {
		if got := in.evaluatePipelineStr(stmt); len(got) != 1 || got[0] != want {
			t.Errorf("%s = %v, want [%s]", stmt, got, want)
		}
	}
}

// ---- Ratio through the interpreter ---------------------------------------------------

func TestRatioThroughTheInterpreter(t *testing.T) {
	in := NewInterpreter()
	in.EvaluateScript(`def rt_third = \x -> x / 3r1 + 1r3
def rt_sum = \xs -> fold (\a x -> a + x) 0r1 xs`)
	for stmt, want := range map[string]string{
		`3r4`:                                   "3r4",
		`0.75r`:                                 "3r4",
		`1r2 + 1r2`:                             "1",
		`1r10 + 2r10`:                           "3r10",
		`1r10 + 2r10 == 3r10`:                   "true",
		`0.1 + 0.2 == 0.3`:                      "false",
		`1r10 == 0.1`:                           "false",
		`to_ratio 0.1 == 1r10`:                  "true",
		`to_ratio 0.1`:                          "1r10",
		`to_ratio_exact 0.5`:                    "1r2",
		`to_ratio_approx 3.14159265 100`:        "311r99",
		`ratio 1 3`:                             "1r3",
		`3 |> ratio 1`:                          "1r3",
		`1r3 / 2`:                               "1r6",
		`7 / 2`:                                 "3.5", // D11: Int / Int is unchanged
		`1r2 + 0.5`:                             "1",   // D12: arithmetic with a Float is a Float
		`7r2 // 1r2`:                            "7",
		`7r2 % 1r3`:                             "1r6",
		`-3r4`:                                  "-3r4",
		`abs -3r4`:                              "3r4",
		`floor 7r2`:                             "3",
		`round 5r2`:                             "3",
		`round_even 5r2`:                        "2",
		`to_float 1r4`:                          "0.25",
		`numer 6r8`:                             "3",
		`denom 6r8`:                             "4",
		`str_from 3r4`:                          "3r4",
		`show -3r4`:                             "-3r4",
		`[1r2, 1r3]`:                            "[1r2 1r3]",
		`rt_sum [1r2, 1r3, 1r6]`:                "1",
		`rt_sum [1r2, 1r3]`:                     "5r6",
		`maybe_default 0 (str_to_ratio "3/4")`:  "3r4",
		`maybe_default 0 (str_to_ratio "-0.5")`: "-1r2",
		`is_just (str_to_ratio "nope")`:         "false",
		`5 |> + 1r2`:                            "11r2",
		`5 |> * 1r2`:                            "5r2",
		`2 |> ** 1r2`:                           "1.4142135623730951",
		`1r2 |> ** 3`:                           "1r8",
	} {
		if got := in.evaluatePipelineStr(stmt); len(got) != 1 || got[0] != want {
			t.Errorf("%s = %v, want [%s]", stmt, got, want)
		}
	}
	// memory types: *functions.Ratio for a fraction, a plain int when whole
	if v := in.EvaluateToValues(`1r3 + 1r6`); len(v) != 1 {
		t.Fatal(v)
	} else if r, ok := v[0].(*functions.Ratio); !ok || r.String() != "1r2" {
		t.Errorf("1r3 + 1r6 is %T %v, want *functions.Ratio 1r2", v[0], v[0])
	}
	if v := in.EvaluateToValues(`1r3 + 2r3`); len(v) != 1 || v[0] != 1 {
		t.Errorf("1r3 + 2r3 = %#v, want the int 1", v)
	}
}

func TestRatioErrorsAreLoud(t *testing.T) {
	in := NewInterpreter()
	for stmt, want := range map[string]string{
		`1r0`:                   "denominator is zero",
		`5r`:                    "malformed ratio",
		`ratio 1 0`:             "denominator zero",
		`ratio 1.5 2`:           "Ints or Ratios",
		`to_ratio nan`:          "", // nan is an undefined name: loud anyway
		`to_ratio "0.5"`:        "str_to_ratio",
		`to_int 1r2`:            "fractional",
		`1r2 // 0`:              "by zero",
		`1r2 % 0`:               "by zero",
		`1r2 / 0`:               "by zero",
		`0r1 ** -1`:             "",
		`0 ** -1`:               "by zero",
		`numer 1.5`:             "Int or a Ratio",
		`to_ratio_approx 0.5 0`: "at least 1",
		`ratio_from_cf [1, 0]`:  "positive",
		`1r2 + "a"`:             "not defined for Ratio and String",
		`1r2 < "a"`:             "not defined for Ratio and String",
	} {
		got := in.evaluatePipelineStr(stmt)
		joined := strings.Join(got, "\n")
		if !strings.Contains(joined, "[ERROR]") || !strings.Contains(joined, want) {
			t.Errorf("%s = %v, want an [ERROR] mentioning %q", stmt, got, want)
		}
	}
}

// A Ratio is a first-class value: usable as a let-binding, a -D constant and a pipeline operand.
func TestRatioAsValue(t *testing.T) {
	in := NewInterpreter()
	in.EvaluateLine(`let half = 1r2`)
	if got := in.evaluatePipelineStr(`half + half`); len(got) != 1 || got[0] != "1" {
		t.Errorf("let-bound Ratio: half + half = %v", got)
	}
	if got := in.evaluatePipelineStr(`half + 1r3`); len(got) != 1 || got[0] != "5r6" {
		t.Errorf("half + 1r3 = %v", got)
	}
	if got := in.evaluatePipelineStr(`[half, 1] |> map (\x -> x * 2)`); len(got) != 1 || got[0] != "[1 2]" {
		t.Errorf("map over Ratios = %v", got)
	}
}

// ---- prefix `!` (user decision 2026-10-13): a synonym of `not`, `!=` unchanged -----------

func TestPrefixBangThroughTheInterpreter(t *testing.T) {
	in := NewInterpreter()
	in.EvaluateScript(`def pb_inv = \b -> !b
let pb_t = true`)
	for stmt, want := range map[string]string{
		`!true`:                                 "false",
		`!false`:                                "true",
		`!(1 == 2)`:                             "true",
		`1 != 2`:                                "true",
		`1 != 1`:                                "false",
		`!pb_t`:                                 "false",
		`!!pb_t`:                                "true",
		`pb_inv true`:                           "false",
		`[true, false] |> map (\x -> !x)`:       "[false true]",
		`[true, false] |> map (\x -> not x)`:    "[false true]",
		`[1, 2, 3] |> filter (\x -> !(x == 2))`: "[1 3]",
		`[1, 2, 3] |> filter (\x -> x != 2)`:    "[1 3]",
		`!true || true`:                         "true",
		`not !true`:                             "true",
		`1 !=2`:                                 "true",
	} {
		if got := in.evaluatePipelineStr(stmt); len(got) != 1 || got[0] != want {
			t.Errorf("%s = %v, want [%s]", stmt, got, want)
		}
	}
	// `!` and `not` agree on every Bool and fail identically on everything else.
	for _, v := range []string{"true", "false", "1", `"x"`, "'c'", "[true]", "1r2", "2.5"} {
		bang := in.evaluatePipelineStr("!(" + v + ")")
		word := in.evaluatePipelineStr("not " + v)
		if len(bang) != 1 || len(word) != 1 {
			t.Errorf("! / not %s: %v / %v", v, bang, word)
			continue
		}
		if strings.Contains(bang[0], "[ERROR]") != strings.Contains(word[0], "[ERROR]") {
			t.Errorf("! and not disagree on %s: %v vs %v", v, bang, word)
		}
		if !strings.Contains(bang[0], "[ERROR]") && bang[0] != word[0] {
			t.Errorf("! and not disagree on %s: %v vs %v", v, bang, word)
		}
	}
	// statements that merely contain a '!' are untouched: text, exec arguments, lambdas that echo it
	for stmt, want := range map[string]string{
		`"a!" ++ "b"`:                        "a!b",
		`spawn S: x |> echo hi!`:             "hi!",
		`spawn S: x |> sh -c "echo wow!"`:    "wow!",
		`spawn S: x |> map (\s -> s ++ "!")`: "x!",
	} {
		if got := in.evaluatePipelineStr(stmt); len(got) != 1 || got[0] != want {
			t.Errorf("%s = %v, want [%s]", stmt, got, want)
		}
	}
}

// Step form ≡ operator form for every arithmetic operator over the tower: `a op b` and the
// pipeline section `a |> op b` answer the same value, or fail with the same message.
func TestOperatorSectionsEqualOperators(t *testing.T) {
	in := NewInterpreter()
	operands := []string{"7", "2", "(-3)", "2.5", "0.5", "3r2", "5r3", "1180591620717411303424", `"s"`, "true"}
	// a section's operand is a literal token: a negative number is written without the parentheses
	section := func(operand string) string { return strings.Trim(operand, "()") }
	ops := []string{"+", "-", "*", "/", "//", "%", "**"}
	checked := 0
	for _, a := range operands {
		for _, b := range operands {
			for _, op := range ops {
				infix := in.evaluatePipelineStr(fmt.Sprintf("%s %s %s", a, op, b))
				step := in.evaluatePipelineStr(fmt.Sprintf("%s |> %s %s", a, op, section(b)))
				if len(infix) != 1 || len(step) != 1 {
					t.Errorf("%s %s %s: infix %v, step %v; want one line each", a, op, b, infix, step)
					continue
				}
				checked++
				iErr, sErr := strings.Contains(infix[0], "[ERROR]"), strings.Contains(step[0], "[ERROR]")
				if iErr != sErr {
					t.Errorf("%s %s %s: infix %q but step %q disagree on failing", a, op, b, infix[0], step[0])
					continue
				}
				if !iErr && infix[0] != step[0] {
					t.Errorf("%s %s %s: infix %q but step %q", a, op, b, infix[0], step[0])
				}
			}
		}
	}
	if checked != len(operands)*len(operands)*len(ops) {
		t.Errorf("checked %d cells, want %d", checked, len(operands)*len(operands)*len(ops))
	}
}
