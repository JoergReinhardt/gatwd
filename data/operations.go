package data

import (
	"math/big"
	"math/cmplx"
)

type TyOps uint8

const (
	Add       TyOps = 1
	Substract TyOps = 1 << iota
	Multiply
	Divide
	Concat
)

// arithmetic operations for numerals take two numeral operands and an
// operation type. casts operands to match in type, switches on the operands
// type and sub-switches on operation type passed, to perform the propper
// operation and return resulting value as instance of native.
func arithmetics(a, b Numeral, ops TyOps) Native {

	var typ TyNat // allocate common operand type
	// cast operands to be of common type
	a, b, typ = castNumeralsGreaterType(a, b)

	switch { // switch on operands type

	case typ.Match(Naturals): // natural arithmetics

		switch ops {

		case Add:
			return UintVal(a.GoUint() + b.GoUint())
		case Substract:
			var ua, ub = a.GoUint(), b.GoUint()
			// if negative result is expected‥.
			if ua < ub {
				// convert to integer
				var ia, ib = a.GoInt(), b.GoInt()
				return IntVal(ia - ib)
			}
			//‥.otherwise return natural number
			return UintVal(ua - ub)
		case Multiply:
			return UintVal(a.GoUint() * b.GoUint())
		case Divide:
			if a.GoUint() > 0 &&
				b.GoUint() > 0 {
				var rat = RatioVal(*big.NewRat(
					int64(a.GoUint()),
					int64(b.GoUint()),
				))
				return &rat
			}
		}

	case typ.Match(Integers): // integer arithmetics

		switch ops {

		case Add:
			return IntVal(a.GoInt() + b.GoInt())
		case Substract:
			return IntVal(a.GoInt() - b.GoInt())
		case Multiply:
			return IntVal(a.GoInt() * b.GoInt())
		case Divide:
			if a.GoInt() > 0 &&
				b.GoInt() > 0 {
				var rat = RatioVal(*big.NewRat(
					int64(a.GoInt()),
					int64(b.GoInt()),
				))
				return &rat
			}
		}

	case typ.Match(Reals): // real arithmetics

		switch ops {

		case Add:
			return FltVal(a.GoFlt() + b.GoFlt())
		case Substract:
			return FltVal(a.GoFlt() - b.GoFlt())
		case Multiply:
			return FltVal(a.GoFlt() * b.GoFlt())
		case Divide:
			if a.GoFlt() > 0 &&
				b.GoFlt() > 0 {
				return FltVal(a.GoFlt() / b.GoFlt())
			}

		}

	case typ.Match(Rationals): // rational arithmetics

		var rat RatioVal
		var ratA, ratB = a.GoRat(), b.GoRat()

		switch ops {

		case Add:
			rat = RatioVal(*ratA.Add(ratA, ratB))
			return &rat
		case Substract:
			rat = RatioVal(*ratA.Sub(ratA, ratB))
			return &rat
		case Multiply:
			rat = RatioVal(*ratA.Mul(ratA, ratB))
			return &rat
		case Divide:
			rat = RatioVal(*ratA.Quo(ratA, ratB))
			return &rat
		}

	case typ.Match(Imaginarys): // imaginary arithmetics

		switch ops {

		case Add:
			return ImagVal(a.GoImag() + b.GoImag())
		case Substract:
			return ImagVal(a.GoImag() - b.GoImag())
		case Multiply:
			return ImagVal(a.GoImag() * b.GoImag())
		case Divide:
			if cmplx.Abs(a.GoImag()) > 0 &&
				cmplx.Abs(b.GoImag()) > 0 {
				return ImagVal(a.GoImag() / b.GoImag())
			}
		}
	}
	// if no value has been computed, return nil instance
	return NilVal{}
}

// takes two numerals to compare their types. in case types don't match, lesser
// numeral is cast as the greater of both types.
func castNumeralsGreaterType(a, b Numeral) (Numeral, Numeral, TyNat) {
	// preset return type to be a's native type
	var typ = a.TypeNat()
	// if type of value a has higher precedence‥.
	if a.TypeNat().Flag() > b.TypeNat().Flag() {
		// convert b's type to match a's type‥.
		b = CastNumeral(
			b.(Numeral),
			a.TypeNat(),
		).(Numeral)

	}
	if a.TypeNat().Flag() < b.TypeNat().Flag() {
		// reset return type to be b's native type
		typ = b.TypeNat()
		// convert a's type to match b's type‥.
		a = CastNumeral(
			a.(Numeral),
			b.TypeNat(),
		).(Numeral)
	}
	// both values are of the same type now‥.
	return a, b, typ
}
