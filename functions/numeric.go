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
//	Int ⊂ Float(approx) ⊂ Complex
//
// (Ratio, exact fractions, joins between Int and Float.) An `Int` is unbounded and
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
