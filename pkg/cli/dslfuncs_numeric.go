package cli

import (
	"fmt"

	"github.com/joergreinhardt/gatwd/functions"
)

// registerNumericFuncs adds the function forms of the numeric tower (hardening W3). Each
// is a curried, data-last closure over a generic function of package functions, so the
// step form (`x |> floor`, every builtin is also a step) and the function form (`floor x`
// inside a lambda) are the same code path. The tower itself — the Int/Float/Complex
// lattice, promotion, demotion, exact comparison — is specified in functions/numeric.go
// and docs/dsl_guide.md.
//
//	pow base exp (the function form of `**`) · sqrt x · isqrt n
//	floor x · ceil x · trunc x · round x · round_even x     (a number to an Int)
//	sign x (-1, 0 or 1) · abs x · negate x
//	gcd a b · lcm a b · divmod a b (floor) · quot a b · rem a b (truncated)
//	to_int x (exact or an error) · to_float x · is_nan x · is_inf x
//
// and the Ratio (exact fraction) row, functions/ratio.go:
//
//	ratio n d · to_ratio x (shortest decimal) · to_ratio_exact x (the binary value)
//	to_ratio_approx x max_den · numer r · denom r · ratio_parts r · ratio_mixed r
//	str_to_ratio s (a Maybe) · ratio_cf r · ratio_from_cf xs · ratio_mediant a b
func registerNumericFuncs(m map[string]any) {
	unary := func(name string, f func(any) (any, error)) {
		defn(m, name, "x", func(a []any) (any, error) { return f(a[0]) })
	}
	defn(m, "pow", "base exp", func(a []any) (any, error) { return functions.Pow(a[0], a[1]) })
	unary("sqrt", functions.Sqrt)
	unary("isqrt", functions.ISqrt)
	unary("sign", functions.Sign)
	unary("abs", functions.Abs)
	unary("negate", functions.Negate)
	unary("to_int", functions.ToInt)
	unary("to_float", functions.ToFloat)
	for name, mode := range map[string]functions.RoundingMode{
		"floor": functions.RoundFloor, "ceil": functions.RoundCeil, "trunc": functions.RoundTrunc,
		"round": functions.RoundHalfAway, "round_even": functions.RoundHalfEven,
	} {
		mode := mode
		unary(name, func(v any) (any, error) { return functions.RoundToInt(mode, v) })
	}
	defn(m, "gcd", "a b", func(a []any) (any, error) { return functions.GCD(a[0], a[1]) })
	defn(m, "lcm", "a b", func(a []any) (any, error) { return functions.LCM(a[0], a[1]) })
	defn(m, "is_nan", "x", func(a []any) (any, error) { return functions.IsNaN(a[0]) })
	defn(m, "is_inf", "x", func(a []any) (any, error) { return functions.IsInf(a[0]) })
	defn(m, "quot", "a b", func(a []any) (any, error) {
		q, _, err := functions.QuotRem(a[0], a[1])
		return q, err
	})
	defn(m, "rem", "a b", func(a []any) (any, error) {
		_, r, err := functions.QuotRem(a[0], a[1])
		return r, err
	})
	defn(m, "divmod", "a b", func(a []any) (any, error) {
		q, err := functions.EvalBinaryOp(functions.OpIDiv, a[0], a[1])
		if err != nil {
			return nil, err
		}
		r, err := functions.EvalBinaryOp(functions.OpMod, a[0], a[1])
		if err != nil {
			return nil, err
		}
		return functions.Pair[any, any]{Fst: q, Snd: r}, nil
	})

	registerRatioFuncs(m)
}

func registerRatioFuncs(m map[string]any) {
	defn(m, "ratio", "n d", func(a []any) (any, error) { return functions.MakeRatio(a[0], a[1]) })
	defn(m, "to_ratio", "x", func(a []any) (any, error) { return functions.ToRatio(a[0]) })
	defn(m, "to_ratio_exact", "x", func(a []any) (any, error) { return functions.ToRatioExact(a[0]) })
	defn(m, "to_ratio_approx", "x max_den", func(a []any) (any, error) { return functions.ToRatioApprox(a[0], a[1]) })
	defn(m, "numer", "r", func(a []any) (any, error) { return functions.Numer(a[0]) })
	defn(m, "denom", "r", func(a []any) (any, error) { return functions.Denom(a[0]) })
	defn(m, "ratio_parts", "r", func(a []any) (any, error) {
		n, err := functions.Numer(a[0])
		if err != nil {
			return nil, err
		}
		d, _ := functions.Denom(a[0])
		return functions.Pair[any, any]{Fst: n, Snd: d}, nil
	})
	defn(m, "ratio_mixed", "r", func(a []any) (any, error) {
		w, f, err := functions.RatioMixed(a[0])
		if err != nil {
			return nil, err
		}
		return functions.Pair[any, any]{Fst: w, Snd: f}, nil
	})
	defn(m, "str_to_ratio", "s", func(a []any) (any, error) {
		s, err := asString(a[0])
		if err != nil {
			return nil, err
		}
		if r, ok := functions.ParseRatioText(s); ok {
			return functions.Just[any](r), nil
		}
		return functions.Nothing[any](), nil
	})
	defn(m, "ratio_cf", "r", func(a []any) (any, error) {
		cf, err := functions.RatioCF(a[0])
		if err != nil {
			return nil, err
		}
		return listOf(cf), nil
	})
	defn(m, "ratio_from_cf", "xs", func(a []any) (any, error) {
		items, ok := seqOf(a[0])
		if !ok {
			return nil, fmt.Errorf("want a list of Ints, got %T", a[0])
		}
		return functions.RatioFromCF(items)
	})
	defn(m, "ratio_mediant", "a b", func(a []any) (any, error) { return functions.RatioMediant(a[0], a[1]) })
}
