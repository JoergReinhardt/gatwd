package functions

import (
	"fmt"
	"math"
	"math/big"
	"math/cmplx"
)

// This file is the specification of gatwd's numeric tower in code: the helpers that
// give every numeric operation ONE canonical answer type, whatever machine type the
// operands happen to be stored in.
//
// # The logical types
//
// The language has four number kinds, ordered by what they can represent:
//
//	Int ⊂ Ratio ⊂ Float(approx) ⊂ Complex
//
// (Ratio, exact fractions, is in ratio.go.) An `Int` is unbounded and
// exact. Behind that single logical type the runtime keeps a fast path — the Go
// integer types, narrowed to the smallest that fits when a literal is read — and a
// *big.Int for everything beyond int64. The two are never visible to a program: an
// operation on integers answers with an `int` when the result fits int64 and with a
// *big.Int only when it does not (CanonInt), so `(2 ** 64) / (2 ** 64)` is the plain
// `1` again, not a big integer that merely holds one. `BigInt` and `Uint` stay
// accepted type names (aliases of Int) so existing `type X = BigInt` lines parse.

// CanonInt returns the canonical Int for a big integer: an `int` when it fits int64 and
// the *big.Int itself otherwise. The argument is never mutated and, when it is returned,
// never aliased with a value the caller still owns mutably.
func CanonInt(b *big.Int) any {
	if b.IsInt64() {
		return fitInt(b.Int64())
	}
	return b
}

// isIntegerValue reports whether v is one of the integer representations: the Go
// integer types or a *big.Int.
func isIntegerValue(v any) bool {
	switch v.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, *big.Int:
		return true
	}
	return false
}

// intOperand reads any integer representation as a *big.Int (a copy for a *big.Int, so
// the caller may mutate the result freely).
func intOperand(v any) (*big.Int, bool) {
	if !isIntegerValue(v) {
		return nil, false
	}
	return ToBigInt(v)
}

// floatOperand reads an Int or a Float as a float64. An integer too large for a double is
// reported as not convertible (the caller turns that into an error instead of letting
// +Inf leak out of a conversion).
func floatOperand(v any) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case float32:
		return float64(x), true
	case *big.Float:
		f, _ := x.Float64()
		return f, true
	case *Ratio:
		return ratioFloat(x.rat())
	case *big.Int:
		f, _ := new(big.Float).SetInt(x).Float64()
		if math.IsInf(f, 0) {
			return 0, false
		}
		return f, true
	}
	if isIntegerValue(v) {
		return ToFloat64(v)
	}
	return 0, false
}

// Negate is unary minus: the additive inverse of any number. Integers stay exact
// (negating the smallest int64 promotes to a big integer and negating a big integer
// that lands on it demotes again), floats flip their sign, complex numbers flip both
// parts. Anything that is not a number is an error naming its type.
func Negate(v any) (any, error) {
	switch x := v.(type) {
	case float64:
		return -x, nil
	case float32:
		return -float64(x), nil
	case complex128:
		return -x, nil
	case complex64:
		return -complex128(x), nil
	case *big.Float:
		return new(big.Float).Neg(x), nil
	case *Ratio:
		return CanonRat(new(big.Rat).Neg(x.rat())), nil
	case Duration:
		if x == Duration(minInt64) {
			return nil, fmt.Errorf("Duration overflow: -%s is outside about ±292 years", x)
		}
		return -x, nil
	}
	if b, ok := intOperand(v); ok {
		return CanonInt(b.Neg(b)), nil
	}
	return nil, fmt.Errorf("operator '-' is not defined for %s; it negates numbers", KindName(v))
}

// maxPowBits bounds the size of an exact integer power: 2**4194304 has four million
// bits (about 500 KB). A larger request is an error, never an out-of-memory.
const maxPowBits = 1 << 22

// Pow is `**`: exponentiation, right-associative in the parser. An Int raised to a
// non-negative Int is exact (unbounded; a result over maxPowBits is an error), an Int
// raised to a negative Int is the Float reciprocal power (like `/`, which is exact
// only when it can be), a Float on either side gives a Float, and a Complex on either
// side gives a Complex. `0 ** negative` is a division by zero and a negative base with a
// fractional exponent is an error (the result is not a real number; use a Complex).
func Pow(base, exp any) (any, error) {
	if isComplexValue(base) || isComplexValue(exp) {
		b, bok := ToComplex128(base)
		e, eok := ToComplex128(exp)
		if !bok || !eok {
			return nil, powTypeError(base, exp)
		}
		if b == 0 && real(e) < 0 {
			return nil, fmt.Errorf("division by zero: 0 ** a negative power")
		}
		return cmplx.Pow(b, e), nil
	}
	if bi, ok := intOperand(base); ok {
		if ei, ok := intOperand(exp); ok {
			return intPow(bi, ei)
		}
	}
	if br, ok := base.(*Ratio); ok {
		if ei, ok := intOperand(exp); ok {
			return ratPow(br.rat(), ei)
		}
	}
	bf, bok := floatOperand(base)
	ef, eok := floatOperand(exp)
	if !bok || !eok {
		return nil, powTypeError(base, exp)
	}
	return floatPow(bf, ef)
}

func isComplexValue(v any) bool {
	switch v.(type) {
	case complex64, complex128:
		return true
	}
	return false
}

func powTypeError(base, exp any) error {
	return fmt.Errorf("operator '**' is not defined for %s and %s", KindName(base), KindName(exp))
}

func intPow(base, exp *big.Int) (any, error) {
	switch {
	case exp.Sign() == 0:
		return fitInt(1), nil
	case base.Sign() == 0:
		if exp.Sign() < 0 {
			return nil, fmt.Errorf("division by zero: 0 ** a negative power")
		}
		return fitInt(0), nil
	case base.IsInt64() && base.Int64() == 1:
		return fitInt(1), nil
	case base.IsInt64() && base.Int64() == -1:
		if exp.Bit(0) == 0 {
			return fitInt(1), nil
		}
		return fitInt(-1), nil
	}
	if !exp.IsInt64() {
		return nil, fmt.Errorf("exponent %s is too large", exp)
	}
	n := exp.Int64()
	if n < 0 {
		// Inexact by nature: the reciprocal is a Float, like `/` on integers that do not divide.
		bf, ok := floatOperand(CanonInt(base))
		if !ok {
			return nil, fmt.Errorf("integer too large to convert to Float for a negative power")
		}
		return floatPow(bf, float64(n))
	}
	if int64(base.BitLen())*n > maxPowBits {
		return nil, fmt.Errorf("result of ** is too large (more than %d bits)", maxPowBits)
	}
	return CanonInt(new(big.Int).Exp(base, exp, nil)), nil
}

// ratPow is a fraction raised to an Int: exact (a negative exponent takes the reciprocal);
// a result over maxPowBits is an error.
func ratPow(base *big.Rat, exp *big.Int) (any, error) {
	if exp.Sign() == 0 {
		return fitInt(1), nil
	}
	if !exp.IsInt64() {
		return nil, fmt.Errorf("exponent %s is too large", exp)
	}
	n := exp.Int64()
	if n < 0 {
		if base.Sign() == 0 {
			return nil, fmt.Errorf("division by zero: 0 ** a negative power")
		}
		base = new(big.Rat).Inv(base)
		n = -n
	}
	bits := int64(base.Num().BitLen()) + int64(base.Denom().BitLen())
	if bits*n > maxPowBits {
		return nil, fmt.Errorf("result of ** is too large (more than %d bits)", maxPowBits)
	}
	num := new(big.Int).Exp(base.Num(), big.NewInt(n), nil)
	den := new(big.Int).Exp(base.Denom(), big.NewInt(n), nil)
	return CanonRat(new(big.Rat).SetFrac(num, den)), nil
}

func floatPow(b, e float64) (any, error) {
	if b == 0 && e < 0 {
		return nil, fmt.Errorf("division by zero: 0 ** a negative power")
	}
	res := math.Pow(b, e)
	switch {
	case math.IsNaN(res) && !math.IsNaN(b) && !math.IsNaN(e):
		return nil, fmt.Errorf("%v ** %v is not a real number (a negative base with a fractional exponent); use a Complex base", b, e)
	case math.IsInf(res, 0) && !math.IsInf(b, 0) && !math.IsInf(e, 0):
		return nil, fmt.Errorf("%v ** %v overflows a Float", b, e)
	}
	return res, nil
}

// ---- the promotion lattice -----------------------------------------------------------

// numClass is a number's position in the lattice Int ⊂ Ratio ⊂ Float ⊂ Complex. An operation on
// two numbers is carried out in the join (the larger class) of the operands.
type numClass int

const (
	clsNone numClass = iota
	clsInt
	clsRatio
	clsFloat
	clsComplex
)

// classOf places a value in the lattice. A *big.Float is a Float (the arbitrary-precision
// form an overflowing Float operation produces); Char, Bool and String are not numbers.
func classOf(v any) numClass {
	switch v.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, *big.Int:
		return clsInt
	case *Ratio:
		return clsRatio
	case float32, float64, *big.Float:
		return clsFloat
	case complex64, complex128:
		return clsComplex
	}
	return clsNone
}

// IsNumber reports whether v is a number of the tower (Int, Float or Complex in any of
// their machine representations).
func IsNumber(v any) bool { return classOf(v) != clsNone }

// arith carries out +, -, *, /, // and % on two numbers in the join of their classes. The
// second result is false when the operator is not an arithmetic one (the caller goes on).
//
// The rules, in one place:
//
//   - Int ⊕ Int is exact and canonical (CanonInt): int64 fast path, big integers beyond,
//     demoted whenever the result fits again. `/` is exact when it divides and a Float
//     (correctly rounded, even for huge integers) otherwise; `//` floors; `%` is the
//     modulo that takes the sign of the divisor, so a == b*(a//b) + a%b always holds.
//   - Ratio involved (and no Float): exact fractions, demoted to an Int when whole; `//` is
//     the Int floor of the quotient and `%` the remainder a - b*(a//b), a Ratio.
//   - Float involved: the other operand converts to Float (an integer too large for a
//     double is an error, never +Inf); `//` is the Int floor of the quotient, `%` a Float.
//   - Complex involved: + - * /; `//` and `%` are errors (no floor on a plane).
//   - Division, floor division and modulo by zero are errors in every class.
func arith(op OpKind, left, right any) (any, bool, error) {
	switch op {
	case OpAdd, OpSub, OpMul, OpDiv, OpIDiv, OpMod:
	default:
		return nil, false, nil
	}
	cl, cr := classOf(left), classOf(right)
	if cl == clsNone || cr == clsNone {
		return nil, false, nil
	}
	class := cl
	if cr > class {
		class = cr
	}
	var res any
	var err error
	switch class {
	case clsInt:
		res, err = intArith(op, left, right)
	case clsRatio:
		res, err = ratioArith(op, left, right)
	case clsFloat:
		res, err = floatArith(op, left, right)
	default:
		res, err = complexArith(op, left, right)
	}
	return res, true, err
}

func divZero(op OpKind) error {
	if op == OpMod {
		return fmt.Errorf("modulo by zero")
	}
	return fmt.Errorf("division by zero")
}

func intArith(op OpKind, left, right any) (any, error) {
	if li, ok := toInt64(left); ok {
		if ri, ok := toInt64(right); ok {
			switch op {
			case OpAdd:
				return AddWithPromotion(li, ri), nil
			case OpSub:
				return SubWithPromotion(li, ri), nil
			case OpMul:
				return MulWithPromotion(li, ri), nil
			}
			if ri == 0 {
				return nil, divZero(op)
			}
			if !(li == math.MinInt64 && ri == -1) { // the one int64 quotient that overflows
				switch op {
				case OpDiv:
					if li%ri == 0 {
						return fitInt(li / ri), nil
					}
					if li > -(1<<53) && li < 1<<53 && ri > -(1<<53) && ri < 1<<53 {
						return float64(li) / float64(ri), nil // both exact doubles: one correctly rounded division
					}
				case OpIDiv:
					return fitInt(floorDiv(li, ri)), nil
				case OpMod:
					return fitInt(floorMod(li, ri)), nil
				}
			}
		}
	}
	lb, _ := intOperand(left)
	rb, _ := intOperand(right)
	switch op {
	case OpAdd:
		return CanonInt(lb.Add(lb, rb)), nil
	case OpSub:
		return CanonInt(lb.Sub(lb, rb)), nil
	case OpMul:
		return CanonInt(lb.Mul(lb, rb)), nil
	}
	if rb.Sign() == 0 {
		return nil, divZero(op)
	}
	quo, rem := new(big.Int).QuoRem(lb, rb, new(big.Int))
	switch op {
	case OpDiv:
		if rem.Sign() == 0 {
			return CanonInt(quo), nil
		}
		f, _ := new(big.Rat).SetFrac(lb, rb).Float64()
		return f, nil
	case OpIDiv:
		if rem.Sign() != 0 && (rem.Sign() < 0) != (rb.Sign() < 0) {
			quo.Sub(quo, big.NewInt(1))
		}
		return CanonInt(quo), nil
	default: // OpMod
		if rem.Sign() != 0 && (rem.Sign() < 0) != (rb.Sign() < 0) {
			rem.Add(rem, rb)
		}
		return CanonInt(rem), nil
	}
}

// floorDiv is the quotient rounded toward negative infinity; b must be non-zero and the
// pair must not be (MinInt64, -1).
func ratioArith(op OpKind, left, right any) (any, error) {
	l, _ := exactOperand(left)
	r, _ := exactOperand(right)
	switch op {
	case OpAdd:
		return CanonRat(l.Add(l, r)), nil
	case OpSub:
		return CanonRat(l.Sub(l, r)), nil
	case OpMul:
		return CanonRat(l.Mul(l, r)), nil
	}
	if r.Sign() == 0 {
		return nil, divZero(op)
	}
	switch op {
	case OpDiv:
		return CanonRat(l.Quo(l, r)), nil
	}
	// floor(l / r): the quotient of two fractions is a fraction; its floor is the Int.
	q := new(big.Rat).Quo(l, r)
	fl := new(big.Int).Div(q.Num(), q.Denom()) // Euclidean = floor, the denominator is positive
	if op == OpIDiv {
		return CanonInt(fl), nil
	}
	rem := new(big.Rat).Sub(l, new(big.Rat).Mul(r, new(big.Rat).SetInt(fl)))
	return CanonRat(rem), nil
}

func floorDiv(a, b int64) int64 {
	q := a / b
	if a%b != 0 && (a < 0) != (b < 0) {
		q--
	}
	return q
}

// floorMod is the remainder that takes the sign of the divisor.
func floorMod(a, b int64) int64 {
	r := a % b
	if r != 0 && (r < 0) != (b < 0) {
		r += b
	}
	return r
}

func floatArith(op OpKind, left, right any) (any, error) {
	_, lBig := left.(*big.Float)
	_, rBig := right.(*big.Float)
	if lBig || rBig {
		return bigFloatArith(op, left, right)
	}
	lf, lok := floatOperand(left)
	rf, rok := floatOperand(right)
	if !lok || !rok {
		return nil, fmt.Errorf("an integer is too large to convert to Float in operator '%s'; convert it explicitly or keep the operation on integers", op)
	}
	switch op {
	case OpAdd:
		return AddFloatWithPromotion(lf, rf), nil
	case OpSub:
		return SubFloatWithPromotion(lf, rf), nil
	case OpMul:
		return MulFloatWithPromotion(lf, rf), nil
	}
	if rf == 0 {
		return nil, divZero(op)
	}
	switch op {
	case OpDiv:
		return lf / rf, nil
	case OpIDiv:
		q := math.Floor(lf / rf)
		if math.IsNaN(q) || math.IsInf(q, 0) {
			return nil, fmt.Errorf("floor division of %v by %v is not a finite number", lf, rf)
		}
		b, _ := new(big.Float).SetFloat64(q).Int(nil)
		return CanonInt(b), nil
	default: // OpMod
		r := math.Mod(lf, rf)
		if r != 0 && (r < 0) != (rf < 0) {
			r += rf
		}
		return r, nil
	}
}

func bigFloatArith(op OpKind, left, right any) (any, error) {
	lbf, rbf := toBigFloat(left), toBigFloat(right)
	switch op {
	case OpAdd:
		return new(big.Float).Add(lbf, rbf), nil
	case OpSub:
		return new(big.Float).Sub(lbf, rbf), nil
	case OpMul:
		return new(big.Float).Mul(lbf, rbf), nil
	case OpDiv:
		if rbf.Sign() == 0 {
			return nil, divZero(op)
		}
		return new(big.Float).Quo(lbf, rbf), nil
	}
	return nil, fmt.Errorf("operator '%s' is not defined for a Float beyond the double range", op)
}

func complexArith(op OpKind, left, right any) (any, error) {
	lc, lok := ToComplex128(left)
	rc, rok := ToComplex128(right)
	if !lok || !rok {
		return nil, fmt.Errorf("an integer is too large to convert to Complex in operator '%s'", op)
	}
	switch op {
	case OpAdd:
		return AddComplexWithPromotion(lc, rc), nil
	case OpSub:
		return lc - rc, nil
	case OpMul:
		res := lc * rc
		if cmplx.IsInf(res) && !cmplx.IsInf(lc) && !cmplx.IsInf(rc) {
			r := new(big.Float).Sub(
				new(big.Float).Mul(big.NewFloat(real(lc)), big.NewFloat(real(rc))),
				new(big.Float).Mul(big.NewFloat(imag(lc)), big.NewFloat(imag(rc))),
			)
			im := new(big.Float).Add(
				new(big.Float).Mul(big.NewFloat(real(lc)), big.NewFloat(imag(rc))),
				new(big.Float).Mul(big.NewFloat(imag(lc)), big.NewFloat(real(rc))),
			)
			return fmt.Sprintf("(%s+%si)", r.Text('g', 10), im.Text('g', 10)), nil
		}
		return res, nil
	case OpDiv:
		if rc == 0 {
			return nil, fmt.Errorf("division by zero in complex arithmetic")
		}
		return lc / rc, nil
	}
	return nil, fmt.Errorf("operator '%s' is not defined for Complex numbers", op)
}

// ---- exact comparison ----------------------------------------------------------------

// CompareNumbers orders two real numbers EXACTLY, whatever their representations: an
// integer is never rounded to a double to be compared with a Float, so `9007199254740993
// > 9007199254740992.0` is true. The second result is false when the pair is unordered
// (either side NaN); ok is false when a or b is not a real number of the tower (Complex
// numbers have no order — use StructuralEqual for equality).
func CompareNumbers(a, b any) (cmp int, ordered, ok bool) {
	ca, cb := classOf(a), classOf(b)
	if ca == clsNone || cb == clsNone || ca == clsComplex || cb == clsComplex {
		return 0, false, false
	}
	if ca == clsInt && cb == clsInt {
		if x, xok := toInt64(a); xok {
			if y, yok := toInt64(b); yok {
				switch {
				case x < y:
					return -1, true, true
				case x > y:
					return 1, true, true
				}
				return 0, true, true
			}
		}
		x, _ := intOperand(a)
		y, _ := intOperand(b)
		return x.Cmp(y), true, true
	}
	_, aBigF := a.(*big.Float)
	_, bBigF := b.(*big.Float)
	if aBigF || bBigF {
		return toBigFloat(a).Cmp(toBigFloat(b)), true, true
	}
	if ca == clsFloat && cb == clsFloat {
		x, _ := floatOperand(a)
		y, _ := floatOperand(b)
		switch {
		case math.IsNaN(x) || math.IsNaN(y):
			return 0, false, true
		case x < y:
			return -1, true, true
		case x > y:
			return 1, true, true
		}
		return 0, true, true
	}
	// Int and Ratio against each other: exact fractions.
	if ca != clsFloat && cb != clsFloat {
		x, _ := ratOperand(a)
		y, _ := ratOperand(b)
		return x.Cmp(y), true, true
	}
	// An exact number (Int or Ratio) against a Float: exact, through the rational value of the double.
	f, i, flip := a, b, false
	if ca != clsFloat {
		f, i, flip = b, a, true
	}
	x, _ := floatOperand(f)
	if math.IsNaN(x) {
		return 0, false, true
	}
	var res int
	switch {
	case math.IsInf(x, 1):
		res = 1
	case math.IsInf(x, -1):
		res = -1
	default:
		ir, _ := ratOperand(i)
		res = new(big.Rat).SetFloat64(x).Cmp(ir)
	}
	if flip {
		res = -res
	}
	return res, true, true
}

// ---- named functions of the tower ----------------------------------------------------
//
// Each is the function form of an operation the infix operators do not cover. They are
// generic (Layer 0); pkg/cli registers them as curried builtins.

func notNumber(fn string, v any) error {
	return fmt.Errorf("%s is not defined for %s; it takes a number", fn, KindName(v))
}

// Abs is the absolute value of a real number, in the same class: Int stays an exact Int
// (abs of the smallest int64 promotes), Float stays Float. A Complex number has a
// magnitude, not an absolute value; use a Complex-aware function for that.
func Abs(v any) (any, error) {
	if d, ok := v.(Duration); ok {
		if d < 0 {
			return Negate(d)
		}
		return d, nil
	}
	switch classOf(v) {
	case clsInt:
		b, _ := intOperand(v)
		return CanonInt(b.Abs(b)), nil
	case clsRatio:
		return CanonRat(new(big.Rat).Abs(v.(*Ratio).rat())), nil
	case clsFloat:
		if bf, ok := v.(*big.Float); ok {
			return new(big.Float).Abs(bf), nil
		}
		f, _ := floatOperand(v)
		return math.Abs(f), nil
	case clsComplex:
		return nil, fmt.Errorf("abs is not defined for Complex; a complex number has a magnitude, not a sign")
	}
	return nil, notNumber("abs", v)
}

// Sign is -1, 0 or 1 as an Int: the sign of a real number. A NaN has no sign (an error).
func Sign(v any) (any, error) {
	cmp, ordered, ok := CompareNumbers(v, 0)
	switch {
	case !ok && IsNumber(v):
		return nil, fmt.Errorf("sign is not defined for Complex")
	case !ok:
		return nil, notNumber("sign", v)
	case !ordered:
		return nil, fmt.Errorf("sign of NaN is undefined")
	}
	return fitInt(int64(cmp)), nil
}

// Sqrt is the square root as a Float. A negative real is an error that points to Complex
// (`sqrt (-4 + 0i)` is `2i`); a Complex argument gives a Complex.
func Sqrt(v any) (any, error) {
	switch classOf(v) {
	case clsComplex:
		c, _ := ToComplex128(v)
		return cmplx.Sqrt(c), nil
	case clsInt, clsRatio, clsFloat:
		f, ok := floatOperand(v)
		if !ok {
			return nil, fmt.Errorf("sqrt: the number is too large for a Float; use isqrt")
		}
		if f < 0 {
			return nil, fmt.Errorf("sqrt of a negative number is not real; use a Complex argument")
		}
		return math.Sqrt(f), nil
	}
	return nil, notNumber("sqrt", v)
}

// ISqrt is the integer square root: the largest Int whose square does not exceed n,
// exact for any size of Int. A negative argument or a non-Int is an error.
func ISqrt(v any) (any, error) {
	b, ok := intOperand(v)
	if !ok {
		return nil, fmt.Errorf("isqrt is not defined for %s; it takes an Int", KindName(v))
	}
	if b.Sign() < 0 {
		return nil, fmt.Errorf("isqrt of a negative number")
	}
	return CanonInt(b.Sqrt(b)), nil
}

// RoundingMode names how a real number rounds to an Int.
type RoundingMode int

// The rounding modes: toward -Inf, toward +Inf, toward 0, half away from zero, half to even.
const (
	RoundFloor RoundingMode = iota
	RoundCeil
	RoundTrunc
	RoundHalfAway
	RoundHalfEven
)

func (m RoundingMode) name() string {
	return [...]string{"floor", "ceil", "trunc", "round", "round_even"}[m]
}

// RoundToInt rounds a real number to an Int by the given mode. An Int is its own
// rounding; a Float rounds and is converted exactly (a huge Float gives a big Int); NaN and
// the infinities have no integer and are errors.
func RoundToInt(mode RoundingMode, v any) (any, error) {
	switch classOf(v) {
	case clsInt:
		return v, nil
	case clsRatio:
		return CanonInt(roundRat(mode, v.(*Ratio).rat())), nil
	case clsFloat:
		if bf, ok := v.(*big.Float); ok {
			// beyond the double range every value is an integer already
			i, _ := bf.Int(nil)
			return CanonInt(i), nil
		}
		f, _ := floatOperand(v)
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return nil, fmt.Errorf("%s of %v: not a finite number", mode.name(), f)
		}
		switch mode {
		case RoundFloor:
			f = math.Floor(f)
		case RoundCeil:
			f = math.Ceil(f)
		case RoundTrunc:
			f = math.Trunc(f)
		case RoundHalfAway:
			f = math.Round(f)
		default:
			f = math.RoundToEven(f)
		}
		i, _ := new(big.Float).SetFloat64(f).Int(nil)
		return CanonInt(i), nil
	case clsComplex:
		return nil, fmt.Errorf("%s is not defined for Complex", mode.name())
	}
	return nil, notNumber(mode.name(), v)
}

// roundRat rounds a (non-integral) fraction to a big integer by mode. The denominator is
// positive, so Euclidean division is the floor.
func roundRat(mode RoundingMode, r *big.Rat) *big.Int {
	num, den := r.Num(), r.Denom()
	floor := new(big.Int).Div(num, den)
	switch mode {
	case RoundFloor:
		return floor
	case RoundCeil:
		return floor.Add(floor, big.NewInt(1)) // not integral, so the ceiling is floor + 1
	case RoundTrunc:
		if num.Sign() < 0 {
			return floor.Add(floor, big.NewInt(1))
		}
		return floor
	}
	// halves: compare the fractional part with 1/2
	frac := new(big.Rat).Sub(r, new(big.Rat).SetInt(floor))
	switch frac.Cmp(big.NewRat(1, 2)) {
	case -1:
		return floor
	case 1:
		return floor.Add(floor, big.NewInt(1))
	}
	// exactly half: away from zero, or to the even neighbour
	up := new(big.Int).Add(floor, big.NewInt(1))
	if mode == RoundHalfAway {
		if num.Sign() < 0 {
			return floor
		}
		return up
	}
	if floor.Bit(0) == 0 {
		return floor
	}
	return up
}

// GCD is the greatest common divisor of two Ints (always non-negative; gcd 0 0 is 0).
func GCD(a, b any) (any, error) {
	x, y, err := twoInts("gcd", a, b)
	if err != nil {
		return nil, err
	}
	return CanonInt(new(big.Int).GCD(nil, nil, x.Abs(x), y.Abs(y))), nil
}

// LCM is the least common multiple of two Ints (non-negative; 0 when either is 0).
func LCM(a, b any) (any, error) {
	x, y, err := twoInts("lcm", a, b)
	if err != nil {
		return nil, err
	}
	if x.Sign() == 0 || y.Sign() == 0 {
		return fitInt(0), nil
	}
	x.Abs(x)
	y.Abs(y)
	g := new(big.Int).GCD(nil, nil, x, y)
	return CanonInt(x.Div(x, g).Mul(x, y)), nil
}

func twoInts(fn string, a, b any) (*big.Int, *big.Int, error) {
	x, aok := intOperand(a)
	y, bok := intOperand(b)
	if !aok || !bok {
		return nil, nil, fmt.Errorf("%s is not defined for %s and %s; it takes two Ints", fn, KindName(a), KindName(b))
	}
	return x, y, nil
}

// ToInt converts a number to an Int without losing anything: an Int is itself and a
// Float with no fractional part converts; a Float with a fraction is an error that names
// the rounding functions (so a lossy conversion is never silent), as is NaN or ±Inf.
func ToInt(v any) (any, error) {
	switch classOf(v) {
	case clsInt:
		return v, nil
	case clsRatio:
		return nil, fmt.Errorf("to_int of %v loses its fractional part; use floor, ceil, round or trunc", v)
	case clsFloat:
		r, err := RoundToInt(RoundTrunc, v)
		if err != nil {
			return nil, fmt.Errorf("to_int: %w", err)
		}
		if cmp, ordered, _ := CompareNumbers(r, v); !ordered || cmp != 0 {
			return nil, fmt.Errorf("to_int of %v loses its fractional part; use floor, ceil, round or trunc", v)
		}
		return r, nil
	case clsComplex:
		return nil, fmt.Errorf("to_int is not defined for Complex")
	}
	return nil, fmt.Errorf("to_int is not defined for %s; use str_to_int to parse text", KindName(v))
}

// ToFloat converts a real number to the nearest Float (ties to even, correctly rounded
// for any size of integer). An integer beyond the double range is an error, never +Inf.
func ToFloat(v any) (any, error) {
	switch classOf(v) {
	case clsFloat:
		if f, ok := floatOperand(v); ok {
			return f, nil
		}
		return v, nil
	case clsInt, clsRatio:
		f, ok := floatOperand(v)
		if !ok {
			return nil, fmt.Errorf("to_float: the number is beyond the range of a Float")
		}
		return f, nil
	case clsComplex:
		return nil, fmt.Errorf("to_float is not defined for Complex")
	}
	return nil, fmt.Errorf("to_float is not defined for %s; use str_to_float to parse text", KindName(v))
}

// IsNaN reports whether a Float is not-a-number; every Int is a number. Anything that is
// not a real number is an error.
func IsNaN(v any) (bool, error) {
	if f, ok := v.(float64); ok {
		return math.IsNaN(f), nil
	}
	if cl := classOf(v); cl == clsInt || cl == clsRatio || cl == clsFloat {
		f, ok := floatOperand(v)
		return ok && math.IsNaN(f), nil
	}
	return false, notNumber("is_nan", v)
}

// IsInf reports whether a Float is +Inf or -Inf; an Int is never infinite.
func IsInf(v any) (bool, error) {
	if cl := classOf(v); cl == clsInt || cl == clsRatio || cl == clsFloat {
		f, ok := floatOperand(v)
		return ok && math.IsInf(f, 0), nil
	}
	return false, notNumber("is_inf", v)
}

// QuotRem is truncated division: the quotient rounded toward zero and the remainder that
// takes the sign of the dividend, for Ints (Go's / and %). `//` and `%` floor instead.
func QuotRem(a, b any) (q, r any, err error) {
	x, y, err := twoInts("quot/rem", a, b)
	if err != nil {
		return nil, nil, err
	}
	if y.Sign() == 0 {
		return nil, nil, fmt.Errorf("division by zero")
	}
	qq, rr := new(big.Int).QuoRem(x, y, new(big.Int))
	return CanonInt(qq), CanonInt(rr), nil
}

// Not is boolean negation, the function behind `not b` and the prefix operator `!b`. It is
// strict: only a Bool has a negation, a number is never "truthy" (Mandate 6).
func Not(v any) (any, error) {
	b, ok := v.(bool)
	if !ok {
		return nil, fmt.Errorf("want a Bool, got %v (%T)", v, v)
	}
	return !b, nil
}
