package data

import "math/big"

type TyOps uint8

const (
	Add       TyOps = 1
	Substract TyOps = 1 << iota
	Multiply
	Divide
	Concat
)

// arithmetic operations for native numerals take two numeral operands and an
// operation type. it switches on the operands type and sub-switches on
// operation type to perform the propper operation.
func arithmeticOps(a, b Numeral, ops TyOps) Native {

	var typ TyNat // allocate common operand type
	// cast operands to be of common type
	a, b, typ = castNumeralsGreaterType(a, b)

	switch { // switch on operands type
	case typ.Match(Naturals): // natural arithmetics
		switch ops {
		case Add:
			return UintVal(a.Uint() + b.Uint())
		case Substract:
			var ua, ub = a.Uint(), b.Uint()
			// if negative result is expected‥.
			if ua < ub {
				// convert to integer
				var ia, ib = a.Int(), b.Int()
				return IntVal(ia - ib)
			}
			//‥.otherwise return natural number
			return UintVal(ua - ub)
		case Multiply:
			return UintVal(a.Uint() * b.Uint())
		case Divide:
			var rat = RatioVal(
				*big.NewRat(
					int64(a.Int()),
					int64(b.Int())))
			return &rat
		}
	case typ.Match(Integers): // integer arithmetics
		switch ops {
		case Add:
			return IntVal(a.Int() + b.Int())
		case Substract:
			return IntVal(a.Int() - b.Int())
		case Multiply:
			return IntVal(a.Int() * b.Int())
		case Divide:
			var rat = RatioVal(
				*big.NewRat(
					int64(a.Int()),
					int64(b.Int())))
			return &rat
		}
	case typ.Match(Reals): // real arithmetics
		switch ops {
		case Add:
			return FltVal(a.Float() + b.Float())
		case Substract:
			return FltVal(a.Float() - b.Float())
		case Multiply:
			return FltVal(a.Float() * b.Float())
		case Divide:
			return FltVal(a.Float() / b.Float())

		}
	case typ.Match(Rationals): // rational arithmetics
		var rat RatioVal
		var ratA, ratB = a.Rat(), b.Rat()
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
			return ImagVal(a.Imag() + b.Imag())
		case Substract:
			return ImagVal(a.Imag() - b.Imag())
		case Multiply:
			return ImagVal(a.Imag() * b.Imag())
		case Divide:
			return ImagVal(a.Imag() / b.Imag())
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
		b = castNumberAs(
			b.(Numeral),
			a.TypeNat(),
		).(Numeral)

	}
	if a.TypeNat().Flag() < b.TypeNat().Flag() {
		// reset return type to be b's native type
		typ = b.TypeNat()
		// convert a's type to match b's type‥.
		a = castNumberAs(
			a.(Numeral),
			b.TypeNat(),
		).(Numeral)
	}
	// both values are of the same type now‥.
	return a, b, typ
}
