package functions

import (
	"math/big"

	d "github.com/joergreinhardt/gatwd/data"
)

type (
	BoolFnc      func() d.BoolVal
	NaturalFnc   func() d.UintVal
	IntegerFnc   func() d.IntVal
	RationalFnc  func() *d.RatioVal
	RealFnc      func() d.FltVal
	ImaginaryFnc func() d.ImagVal
	// methods of all number functions
	NumberFnc func() d.Native
)

func (n BoolFnc) Bool() d.BoolVal { return n() }
func (n BoolFnc) Int() d.IntVal   { return d.IntVal(n.Uint()) }
func (n BoolFnc) Uint() d.UintVal {
	if n() {
		return 1
	}
	return 0
}

//////////
func (n NaturalFnc) Uint() d.UintVal { return n() }
func (n NaturalFnc) Int() d.IntVal   { return d.IntVal(n()) }
func (n NaturalFnc) Bool() d.BoolVal { return n() > 0 }

//////////
func (n IntegerFnc) Int() d.IntVal   { return n() }
func (n IntegerFnc) Uint() d.UintVal { return d.UintVal(n()) }
func (n IntegerFnc) Real() d.FltVal  { return d.FltVal(n()) }
func (n IntegerFnc) Bool() d.BoolVal { return n() > 0 }
func (n IntegerFnc) Imagine() ImaginaryFnc {
	return func() d.ImagVal { return d.ImagVal(complex(n.Real(), 0)) }
}
func (n IntegerFnc) Rationalize() RationalFnc {
	return func() *d.RatioVal {
		return (*d.RatioVal)(big.NewRat(int64(n()), 1))
	}
}

//////////
func (n RationalFnc) Real() d.FltVal {
	flt, ext := (*big.Rat)(n()).Float64()
	if ext {
		return d.FltVal(flt)
	}
	return 0
}
func (n RationalFnc) Int() d.IntVal {
	if (*big.Rat)(n()).IsInt() {
		return d.IntVal((*big.Rat)(n()).Num().Int64())
	}
	return 0
}

//////////
func (n RealFnc) Real() d.FltVal { return n() }
func (n RealFnc) Rationalize() *d.RatioVal {
	return (*d.RatioVal)(big.NewRat(1, 1).SetFloat64(float64(n())))
}
func (n RealFnc) Int() d.IntVal {
	return RationalFnc(
		func() *d.RatioVal {
			return n.Rationalize()
		}).Int()
}

//////////
func (n ImaginaryFnc) Imaginary() d.ImagVal { return n() }

//////////
func NewNumberFnc(num Number) NumberFnc {
	return NumberFnc(func() d.Native {
		return num
	})
}

//////////
func (n NumberFnc) Null() d.Native { return d.IntVal(0) }
func (n NumberFnc) Init() d.Native { return d.IntVal(1) }
func (n NumberFnc) Bool() d.BoolVal {
	var b d.BoolVal
	var f = n().TypeNat().Flag()
	switch {
	case f.Match(d.Booleans):
		return n().(d.BoolVal)
	case f.Match(d.Naturals):
	}
	return b
}
func (n NumberFnc) Uint() d.UintVal {
	var u d.UintVal
	return u
}
func (n NumberFnc) Int() d.IntVal {
	var i d.IntVal
	return i
}
func (n NumberFnc) Rat() *d.RatioVal {
	var r *d.RatioVal
	return r
}
func (n NumberFnc) Real() d.FltVal {
	var f d.FltVal
	return f
}
func (n NumberFnc) Imaginary() d.ImagVal {
	var i d.ImagVal
	return i
}
