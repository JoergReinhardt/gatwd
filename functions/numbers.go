package functions

import (
	"math/big"

	d "github.com/joergreinhardt/gatwd/data"
)

type (
	NumberCons func() Numeral
)

func NewNumber(num Callable) NumberCons {
	var cons NumberCons
	switch num.TypeNat() {
	case d.Bool:
	}
	return cons
}

// CALLABLE INTERFACE
func (n NumberCons) Ident() Callable { return n }
func (n NumberCons) Call(arg ...Callable) Callable {
	return Native(func() d.Native { return n().Eval() })
}
func (n NumberCons) Eval(arg ...d.Native) d.Native { return n().Eval() }
func (n NumberCons) String() string                { return n().String() }
func (n NumberCons) TypeNat() d.TyNative           { return n().TypeNat() }
func (n NumberCons) TypeFnc() TyFnc                { return Number }

// NUMBER INTERFACE
func (n NumberCons) Nullable() d.Native {
	var nat Native
	switch n.TypeNat() {
	case d.Bool:
		return d.BoolVal(false)
	case d.Uint:
		return d.UintVal(0)
	case d.Int:
		return d.IntVal(0)
	case d.Ratio:
		return (*d.RatioVal)(
			big.NewRat(int64(0), int64(1)),
		)
	case d.Float:
		return d.FltVal(0)
	case d.Imag:
		return d.ImagVal(
			complex(0, 0),
		)
	}
	return nat
}

func (n NumberCons) Bool() bool {
	var nat bool
	return nat
}

func (n NumberCons) Uint() uint {
	var nat uint
	return nat
}

func (n NumberCons) Int() int {
	var nat int
	return nat
}

func (n NumberCons) Rat() *big.Rat {
	var nat *big.Rat
	return nat
}

func (n NumberCons) Float() float64 {
	var nat float64
	return nat
}

func (n NumberCons) Imag() complex128 {
	var nat complex128
	return nat
}
