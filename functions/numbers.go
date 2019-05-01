package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

type TyArriOp int8

func (t TyArriOp) Eval(...d.Native) d.Native { return t }
func (t TyArriOp) TypeNat() d.TyNat          { return d.Flag }
func (t TyArriOp) Flag() d.BitFlag           { return d.BitFlag(t) }
func (t TyArriOp) Uint() uint                { return d.BitFlag(t).Uint() }

//go:generate stringer -type=TyArriOp
const (
	Add       TyArriOp = 1
	Substract TyArriOp = 1 << iota
	Multiply
	Divide
	Modulo
	SquareRoot
)

type (
	CounterFnc func() d.IntVal
	NumberCons func() d.Native
)

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
