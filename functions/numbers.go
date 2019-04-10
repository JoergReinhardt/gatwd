package functions

import (
	"math/big"

	d "github.com/joergreinhardt/gatwd/data"
)

type TyArriOp int8

func (t TyArriOp) Eval(...d.Native) d.Native { return t }
func (t TyArriOp) TypeNat() d.TyNative       { return d.Flag }
func (t TyArriOp) Flag() d.BitFlag           { return d.BitFlag(t) }
func (t TyArriOp) Uint() uint                { return d.BitFlag(t).Uint() }

//go:generate stringer -type=TyArriOp
const (
	Add       TyArriOp = 1
	Substract TyArriOp = 1 << iota
	Multiply
	Divide
	Modulo
)

type (
	CounterFnc  func() d.IntVal
	NumberCons  func() Numeral
	OperatorFnc func(a, b d.Native) Numeral
)

func (o OperatorFnc) Eval(args ...d.Native) d.Native {

	var a, b d.Native
	var ok bool

	if len(args) > 0 {
		if a, ok = args[0].(d.Native); ok {
			if len(args) > 1 {
				if b, ok = args[1].(d.Native); ok {
					return o(a, b)
				}
			}
			return o(a, b)
		}
		return o(a, d.IntVal(0))
	}
	return o(d.IntVal(0), d.IntVal(0))
}

func (o OperatorFnc) Call(args ...Callable) Callable {

	var a, b d.Native
	var ok bool

	if len(args) > 0 {
		if a, ok = args[0].Eval().(d.Native); ok {
			if len(args) > 1 {
				if b, ok = args[1].Eval().(d.Native); ok {
					return NewNative(o(a, b))
				}
			}
		}
	}

	return NewNative(o(a, d.IntVal(0)))
}

func (o OperatorFnc) TypeNat() d.TyNative { return d.Numbers }
func (o OperatorFnc) TypeFnc() TyFnc      { return Operator }
func (o OperatorFnc) String() string      { return "arrithmetic operator" }

// create counter, starting with the 'count' parameter, applying current count
// and step to the operator
func NewSeries(opty TyArriOp, args ...d.IntVal) CounterFnc {

	// allocate operator function
	var operator func(a d.IntVal) d.IntVal
	// initialize running count, step size & max value
	var count, step, max d.IntVal = 0, 0, 0

	// extract & assign arguments
	if len(args) > 0 {
		count = args[0]
		if len(args) > 1 {
			step = args[1]
			if len(args) > 2 {
				max = args[2]
			}
		}
	}

	// assign operator function, based on the passed operator type
	switch opty {
	case Add:
		operator = func(n d.IntVal) d.IntVal {
			return n + step
		}
	case Substract:
		operator = func(n d.IntVal) d.IntVal {
			return n - step
		}
	case Multiply:
		operator = func(n d.IntVal) d.IntVal {
			return n * step
		}
	case Modulo:
		operator = func(n d.IntVal) d.IntVal {
			return n % step
		}
	case Divide:
		operator = func(n d.IntVal) d.IntVal {
			return n / step
		}
	}

	// return the counter
	return CounterFnc(func() d.IntVal {
		var result d.IntVal
		if count < max {
			result = count
			count = operator(count)
		}
		return result
	})
}

// create generalized number from callable
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
			big.NewRat(
				int64(0),
				int64(1),
			),
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
