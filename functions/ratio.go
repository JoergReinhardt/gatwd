package functions

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
)

// Ratio is an exact fraction: the language's rational number type (`3r4`, `0.75r`).
//
// # The concept
//
// Floating point cannot represent one tenth: `0.1 + 0.2 == 0.3` is false in every IEEE
// language. A rational number is a pair of integers (numerator, denominator) and its
// arithmetic is exact: `1r10 + 2r10 == 3r10`. The rationals sit between the integers and
// the reals (Int ⊂ Ratio ⊂ Float(approx) ⊂ Complex), and Haskell's Rational, Clojure's
// Ratio and Python's Fraction are the models.
//
// # Why a named type, and the canonical form
//
// Ratio is `type Ratio big.Rat`, used as *Ratio, so it can print as the literal `3r4`
// (a bare *big.Rat prints `3/4`, which does not lex back, and lists print their elements
// with %v). A Ratio is always in lowest terms with a positive denominator and is never
// mutated. A fraction whose denominator is 1 is not a Ratio at all: it DEMOTES to an Int
// (CanonRat), so `1r2 + 1r2` is the plain `1` — the same rule as big integers that fit
// int64 — and Int ⊂ Ratio holds without a `4r1` that is "almost an Int".
//
// # Example
//
//	r := functions.NewRatio(big.NewInt(6), big.NewInt(8)) // 3r4
//	r.String()                                            // "3r4"
type Ratio big.Rat

func (r *Ratio) rat() *big.Rat { return (*big.Rat)(r) }

// String renders the literal form `3r4`, which the lexer reads back to the same Ratio.
func (r *Ratio) String() string {
	return r.rat().Num().String() + "r" + r.rat().Denom().String()
}

// Num is the numerator (carries the sign) as a fresh big integer.
func (r *Ratio) Num() *big.Int { return new(big.Int).Set(r.rat().Num()) }

// Denom is the positive denominator as a fresh big integer.
func (r *Ratio) Denom() *big.Int { return new(big.Int).Set(r.rat().Denom()) }

// Rat returns an independent *big.Rat holding the same value, for computation.
func (r *Ratio) Rat() *big.Rat { return new(big.Rat).Set(r.rat()) }

// CanonRat returns the canonical number for a fraction: an Int (CanonInt) when the
// denominator is 1, otherwise a *Ratio. The argument is copied, never aliased.
func CanonRat(x *big.Rat) any {
	if x.IsInt() {
		return CanonInt(new(big.Int).Set(x.Num()))
	}
	return (*Ratio)(new(big.Rat).Set(x))
}

// NewRatio is n/d in lowest terms, demoted to an Int when d divides n; d must not be zero.
func NewRatio(n, d *big.Int) (any, error) {
	if d.Sign() == 0 {
		return nil, fmt.Errorf("ratio with denominator zero")
	}
	return CanonRat(new(big.Rat).SetFrac(n, d)), nil
}

// maxRatioText bounds the text a Ratio is parsed from, and maxRatioExp the decimal
// exponent it may carry, so malformed input cannot ask for a gigantic number.
const (
	maxRatioText = 512
	maxRatioExp  = 1000
)

// ParseRatioText reads exact decimal or fraction text: `3/4`, `3r4`, `-0.75`, `1.5e-3`,
// `12`. It reports false for anything else (whitespace, `nan`, `inf`, hex, an exponent
// beyond ±1000, a zero denominator); the caller turns that into Nothing or an error.
func ParseRatioText(s string) (any, bool) {
	if s == "" || len(s) > maxRatioText || strings.TrimSpace(s) != s {
		return nil, false
	}
	s = strings.Replace(s, "r", "/", 1)
	if i := strings.IndexAny(s, "eE"); i >= 0 {
		exp, err := strconv.Atoi(strings.TrimPrefix(s[i+1:], "+"))
		if err != nil || exp > maxRatioExp || exp < -maxRatioExp {
			return nil, false
		}
	}
	for _, c := range s {
		if !strings.ContainsRune("0123456789+-./eE", c) {
			return nil, false
		}
	}
	r, ok := new(big.Rat).SetString(s)
	if !ok {
		return nil, false
	}
	return CanonRat(r), true
}

// parseRatioLiteral reads the source literals `3r4` (numerator `r` denominator, unsigned
// integers) and `0.75r` / `12.5r` (an exact decimal with an `r` suffix). The error names
// the problem, so the lexer can report `1r0` as "denominator is zero".
func parseRatioLiteral(text string) (any, error) {
	if num, den, ok := strings.Cut(text, "r"); ok && den != "" {
		n, nok := new(big.Int).SetString(num, 10)
		d, dok := new(big.Int).SetString(den, 10)
		if !nok || !dok || strings.ContainsAny(num+den, "+-.") {
			return nil, fmt.Errorf("malformed ratio literal %q (want numerator r denominator, e.g. 3r4)", text)
		}
		if d.Sign() == 0 {
			return nil, fmt.Errorf("ratio literal %q: denominator is zero", text)
		}
		return NewRatio(n, d)
	}
	dec := strings.TrimSuffix(text, "r")
	if !strings.HasSuffix(text, "r") || !strings.Contains(dec, ".") {
		return nil, fmt.Errorf("malformed ratio literal %q (want 3r4 or an exact decimal such as 0.75r)", text)
	}
	r, ok := new(big.Rat).SetString(dec)
	if !ok {
		return nil, fmt.Errorf("malformed ratio literal %q", text)
	}
	return CanonRat(r), nil
}

// ParseRatioLiteral is the lexer's reader for a ratio literal: an error names what is wrong.
func ParseRatioLiteral(text string) (any, error) { return parseRatioLiteral(text) }

// IsRatioLiteralText reports whether a numeric token has the `r` suffix form.
func IsRatioLiteralText(text string) bool { return strings.Contains(text, "r") }

// ratOperand reads an Int, a Ratio or a finite Float EXACTLY as a fraction (a Float by its
// binary value); NaN and the infinities, and everything not a real number, are false.
func ratOperand(v any) (*big.Rat, bool) {
	switch x := v.(type) {
	case *Ratio:
		return x.Rat(), true
	case float64:
		if math.IsNaN(x) || math.IsInf(x, 0) {
			return nil, false
		}
		return new(big.Rat).SetFloat64(x), true
	case float32:
		return ratOperand(float64(x))
	case *big.Float:
		if x.IsInf() {
			return nil, false
		}
		r, _ := x.Rat(nil)
		return r, true
	}
	if b, ok := intOperand(v); ok {
		return new(big.Rat).SetInt(b), true
	}
	return nil, false
}

// ratioFloat converts a fraction to the nearest double (correctly rounded); false when the
// magnitude is beyond the double range.
func ratioFloat(r *big.Rat) (float64, bool) {
	f, _ := r.Float64()
	return f, !math.IsInf(f, 0)
}

// ---- conversions ---------------------------------------------------------------------

// MakeRatio is `ratio n d`: the exact quotient of two Ints or Ratios (so `ratio 1r2 3` is
// 1r6), demoted to an Int when it is whole. A zero denominator or a non-exact operand is an error.
func MakeRatio(n, d any) (any, error) {
	nr, nok := exactOperand(n)
	dr, dok := exactOperand(d)
	if !nok || !dok {
		return nil, fmt.Errorf("ratio takes Ints or Ratios, got %s and %s", KindName(n), KindName(d))
	}
	if dr.Sign() == 0 {
		return nil, fmt.Errorf("ratio with denominator zero")
	}
	return CanonRat(new(big.Rat).Quo(nr, dr)), nil
}

// exactOperand is ratOperand restricted to exact numbers (Int, Ratio): a Float is
// never converted implicitly.
func exactOperand(v any) (*big.Rat, bool) {
	switch v.(type) {
	case float32, float64, *big.Float:
		return nil, false
	}
	return ratOperand(v)
}

// ToRatio converts to the Ratio a person MEANT: a Float becomes the exact fraction of its
// SHORTEST round-tripping decimal (`0.1` is 1r10, not the binary neighbour); an Int or
// Ratio is itself. NaN, ±Inf, Complex and text have no Ratio (an error).
func ToRatio(v any) (any, error) {
	switch x := v.(type) {
	case float32, float64:
		f, _ := floatOperand(x)
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return nil, fmt.Errorf("to_ratio: %v has no Ratio", f)
		}
		r, ok := new(big.Rat).SetString(strconv.FormatFloat(f, 'g', -1, 64))
		if !ok {
			return nil, fmt.Errorf("to_ratio: cannot convert %v", f)
		}
		return CanonRat(r), nil
	case *Ratio:
		return x, nil
	}
	if isIntegerValue(v) {
		return v, nil
	}
	return nil, fmt.Errorf("to_ratio is not defined for %s; use str_to_ratio to parse text", KindName(v))
}

// ToRatioExact converts a Float to the EXACT fraction of its binary value
// (`0.1` is 3602879701896397r36028797018963968); an Int or Ratio is itself.
func ToRatioExact(v any) (any, error) {
	switch x := v.(type) {
	case float32, float64:
		f, _ := floatOperand(x)
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return nil, fmt.Errorf("to_ratio_exact: %v has no Ratio", f)
		}
		return CanonRat(new(big.Rat).SetFloat64(f)), nil
	case *Ratio:
		return x, nil
	}
	if isIntegerValue(v) {
		return v, nil
	}
	return nil, fmt.Errorf("to_ratio_exact is not defined for %s", KindName(v))
}

// ToRatioApprox is the best rational approximation of x whose denominator is at most
// maxDen, by continued-fraction convergents (the algorithm of Python's
// Fraction.limit_denominator): `3.14159265` with 100 is 311r99. x is a Float (read as
// ToRatio does), an Int or a Ratio; maxDen is an Int of at least 1.
func ToRatioApprox(x, maxDen any) (any, error) {
	md, ok := intOperand(maxDen)
	if !ok || md.Sign() <= 0 {
		return nil, fmt.Errorf("to_ratio_approx: the maximum denominator must be an Int of at least 1, got %v", maxDen)
	}
	exact, err := ToRatio(x)
	if err != nil {
		return nil, fmt.Errorf("to_ratio_approx: %w", err)
	}
	r, _ := ratOperand(exact)
	if r.Denom().Cmp(md) <= 0 {
		return CanonRat(r), nil
	}
	p0, q0, p1, q1 := big.NewInt(0), big.NewInt(1), big.NewInt(1), big.NewInt(0)
	n, d := new(big.Int).Set(r.Num()), new(big.Int).Set(r.Denom())
	for {
		a := new(big.Int).Div(n, d)
		q2 := new(big.Int).Add(q0, new(big.Int).Mul(a, q1))
		if q2.Cmp(md) > 0 {
			break
		}
		p0, q0, p1, q1 = p1, q1, new(big.Int).Add(p0, new(big.Int).Mul(a, p1)), q2
		n, d = d, new(big.Int).Sub(n, new(big.Int).Mul(a, d))
	}
	k := new(big.Int).Div(new(big.Int).Sub(md, q0), q1)
	bound1 := new(big.Rat).SetFrac(new(big.Int).Add(p0, new(big.Int).Mul(k, p1)), new(big.Int).Add(q0, new(big.Int).Mul(k, q1)))
	bound2 := new(big.Rat).SetFrac(p1, q1)
	err1 := new(big.Rat).Abs(new(big.Rat).Sub(bound1, r))
	err2 := new(big.Rat).Abs(new(big.Rat).Sub(bound2, r))
	if err2.Cmp(err1) <= 0 {
		return CanonRat(bound2), nil
	}
	return CanonRat(bound1), nil
}

// Numer is the numerator of an Int (itself) or a Ratio (carries the sign).
func Numer(v any) (any, error) {
	if r, ok := v.(*Ratio); ok {
		return CanonInt(r.Num()), nil
	}
	if isIntegerValue(v) {
		return v, nil
	}
	return nil, fmt.Errorf("numer is not defined for %s; it takes an Int or a Ratio", KindName(v))
}

// Denom is the positive denominator of a Ratio, and 1 for an Int.
func Denom(v any) (any, error) {
	if r, ok := v.(*Ratio); ok {
		return CanonInt(r.Denom()), nil
	}
	if isIntegerValue(v) {
		return fitInt(1), nil
	}
	return nil, fmt.Errorf("denom is not defined for %s; it takes an Int or a Ratio", KindName(v))
}

// RatioMixed splits a Ratio into its whole part and proper fraction, truncating toward
// zero: 7r2 is (3, 1r2) and -7r2 is (-3, -1r2). An Int is (n, 0).
func RatioMixed(v any) (whole, frac any, err error) {
	r, ok := exactOperand(v)
	if !ok {
		return nil, nil, fmt.Errorf("ratio_mixed is not defined for %s; it takes an Int or a Ratio", KindName(v))
	}
	w := new(big.Int).Quo(r.Num(), r.Denom())
	return CanonInt(w), CanonRat(new(big.Rat).Sub(r, new(big.Rat).SetInt(w))), nil
}

// RatioCF is the continued-fraction coefficients [a0; a1, a2, …] of an Int or Ratio (a0
// floors, so the rest are positive): 7r3 is [2, 3].
func RatioCF(v any) ([]any, error) {
	r, ok := exactOperand(v)
	if !ok {
		return nil, fmt.Errorf("ratio_cf is not defined for %s; it takes an Int or a Ratio", KindName(v))
	}
	n, d := new(big.Int).Set(r.Num()), new(big.Int).Set(r.Denom())
	var out []any
	for d.Sign() != 0 {
		a, m := new(big.Int).DivMod(n, d, new(big.Int))
		out = append(out, CanonInt(a))
		n, d = d, m
	}
	return out, nil
}

// RatioFromCF is the inverse of RatioCF: [a0, a1, …] to the fraction a0 + 1/(a1 + 1/…). The
// list must be non-empty, of Ints, with every coefficient after the first positive.
func RatioFromCF(coeffs []any) (any, error) {
	if len(coeffs) == 0 {
		return nil, fmt.Errorf("ratio_from_cf: the coefficient list is empty")
	}
	acc := new(big.Rat)
	for i := len(coeffs) - 1; i >= 0; i-- {
		a, ok := intOperand(coeffs[i])
		if !ok {
			return nil, fmt.Errorf("ratio_from_cf: coefficient %d is %s, want an Int", i, KindName(coeffs[i]))
		}
		if i > 0 && a.Sign() <= 0 {
			return nil, fmt.Errorf("ratio_from_cf: coefficient %d must be positive, got %s", i, a)
		}
		if i == len(coeffs)-1 {
			acc.SetInt(a)
			continue
		}
		acc.Inv(acc)
		acc.Add(acc, new(big.Rat).SetInt(a))
	}
	return CanonRat(acc), nil
}

// RatioMediant is the Stern–Brocot mediant (a+c)/(b+d) of the fractions a/b and c/d:
// the simplest fraction strictly between them when they are neighbours. Both operands
// must be exact (Int or Ratio) with positive denominators — which a Ratio always has.
func RatioMediant(x, y any) (any, error) {
	a, aok := exactOperand(x)
	b, bok := exactOperand(y)
	if !aok || !bok {
		return nil, fmt.Errorf("ratio_mediant takes Ints or Ratios, got %s and %s", KindName(x), KindName(y))
	}
	num := new(big.Int).Add(a.Num(), b.Num())
	den := new(big.Int).Add(a.Denom(), b.Denom())
	return CanonRat(new(big.Rat).SetFrac(num, den)), nil
}
