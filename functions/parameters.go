/*
PARAMETERS

  compose base function data types (functions.go) with type flags, a monoid,
  argument-/ and return patterns (patterns.go), to form a constructor, or other
  type of callable higher order function declaration. implemets a runtime
  defined type and is referenced by it's definition. the implementation maps
  the higher order arguments and return values to the according parameters of
  the implementing functions, and references those. when called with a
  completed set of parameters (single call, or after consequtive curry/, the
  embedded funtion will be called passig those parameters and a resulting value
  will be  yielded according to the return values definition. local value
  declarations, as well as names, and/or positions of parameters and return
  values are enclosed.
*/
package functions

import (
	d "github.com/JoergReinhardt/godeep/data"
)

type (
	unary  func(Data) Data
	binary func(a, b Data) Data
	nary   func(...Data) Data
	tuple  func() (Data, Quantified)
	vector func() []Data // <- indexable native golang slice of data instances
)

// TUPLE
func (tup tuple) Flag() d.BitFlag {
	da, _ := tup()
	return da.Flag() |
		d.Parameter.Flag() |
		Accessor.Flag()
}
func (tup tuple) Type() Flag     { d, _ := tup(); return newFlag(Tuple, d.Flag()) }
func (tup tuple) String() string { d, c := tup(); return d.String() + " " + c.(vector).String() }

// VECTOR
// vector keeps a slice of data instances
func newVector(dd ...d.Data) Quantified {
	return vector(func() (vec []Data) {
		for _, d := range dd {
			vec = append(vec, newData(d))
		}
		return vec
	})
}

// implements functions/sliceable interface
func (v vector) Len() int { return len(v()) }
func (v vector) Empty() bool {
	if len(v()) > 0 {
		for _, dat := range v() {
			if !d.Nil.Flag().Match(dat.Flag()) {
				return false
			}
		}
	}
	return true
}
func (v vector) Flag() d.BitFlag {
	var flag d.BitFlag
	for _, dat := range v() {
		flag = flag.Concat(dat.Flag())
	}
	return flag | Vector.Flag()
}
func (v vector) Type() Flag {
	return newFlag(Vector,
		d.Slice.Flag()|
			d.Parameter.Flag()|
			v.Flag())
}
func (v vector) String() string {
	var slice []d.Data
	for _, dat := range v() {
		slice = append(slice, dat)
	}
	return d.StringSlice("∙", "[", "]", slice...)
}
func (v vector) Eval() Data     { return v }
func (v vector) Vector() []Data { return v() }
func (v vector) Slice() []Data  { return v() }
