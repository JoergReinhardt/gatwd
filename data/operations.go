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
func Ops(x, y Native, ops TyOps) Native {

	var typ = x.Type()

	x, y = Precedence(x, y)

	switch { // switch on operands type

	case typ.Match(Naturals): // natural arithmetics

		var a, b = x.(Numeral).Uint(), y.(Numeral).Uint()

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

		var a, b = x.(Numeral).Uint(), y.(Numeral).Uint()

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

		var a, b = x.(Numeral).Float(), y.(Numeral).Float()

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
		var ratA, ratB = x.(Numeral).GoRat(), y.(Numeral).GoRat()

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

		var a, b = x.(Numeral).Imag(), y.(Numeral).Imag()

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
