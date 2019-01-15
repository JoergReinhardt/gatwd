/*
DATA CONSTRUCTORS

  corpose base functional data constructors (functions.go) with type flags, a
  monoid definition to construct, argument-/ and return pattern
  sets(patterns.go), to form a constructor, or other type of callable higher
  order function declaration.

  implemets a monoidal runtime defined type according to is definition, which
  it is referenced by.V the implementation maps the higher order arguments and
  return values according to the parameter definition of the implementing
  function, that is referenced by a higher order type as one possible
  implementation to chose from, depending on input types pattern.

  when called with a completed set of parameters (single call, or after
  consequtive curry/, the embedded funtion will be called passig those
  parameters and a yield the resulting value according to the return value
  definition. local value declarations, as well as names, and/or positions of
  parameters and return values are enclosed by the constructor.
*/
package functions

import (
	d "github.com/JoergReinhardt/godeep/data"
)

type (
	constant func() Data // <- guarantueed to allways evaluate identicly
	tuple    func(...Data) (Data, []Data)
	unary    func(Data) Data
	binary   func(a, b Data) Data
	nary     func(...Data) Data
	vector   func() []Data // <- indexable native golang slice of data instances
)

// CONSTANT
// constant also conains immutable data, but it may be the result of a constant experssion
func newConstant(dat Data) Data    { return constant(func() Data { return dat.(Functional).Eval() }) }
func (c constant) Flag() d.BitFlag { return Constant.Flag() }
func (c constant) Type() Flag      { return newFlag(Constant, c().Flag()) }
func (c constant) String() string  { return c().(d.Data).String() }
func (c constant) Eval() Data      { return c }

// TUPLE
func newTuple(tail ...Data) Reduceable {
	var head Data
	switch len(tail) {
	case 0:
		head, tail = nil, nil
	case 1:
		head, tail = tail[0], nil
	case 2:
		head, tail = tail[0], []Data{tail[1]}
	default:
		head, tail = tail[0], tail[1:]
	}
	return tuple(func(d ...Data) (Data, []Data) {
		if len(d) > 0 {
			newTuple(d...)
		}
		return head, tail
	})
}
func (tup tuple) Flag() d.BitFlag {
	da, _ := tup()
	return da.Flag() |
		d.Parameter.Flag() |
		Accessor.Flag()
}
func (tup tuple) Len() int {
	var l int
	var h, t = tup()
	if !elemEmpty(h) {
		l = l + 1
	}
	l = l + len(t)
	return l
}
func (tup tuple) Type() Flag   { d, _ := tup(); return newFlag(Tuple, d.Flag()) }
func (tup tuple) Head() Data   { h, _ := tup(); return h }
func (tup tuple) Tail() []Data { _, t := tup(); return t }
func (tup tuple) Slice() []d.Data {
	var head, tail = tup()
	var slice = []d.Data{head}
	for _, t := range tail {
		slice = append(slice, t)
	}
	return slice
}
func (tup tuple) Shift() Reduceable {
	var dat Reduceable
	_, t := tup()
	switch len(t) {
	case 0:
		dat = newTuple(nil, nil)
	case 1:
		dat = newTuple(tup.Tail()[0], nil)
	default:
		dat = newTuple(tup.Tail()...)
	}
	return dat
}

func (tup tuple) String() string {
	dat, _ := tup()
	return dat.String() + " " + d.StringSlice("∙", "[", "]", tup.Slice()...)
}
func (tup tuple) Empty() bool {

	return true
}

// VECTOR
// vector keeps a slice of data instances
func vectorConstructor(dd ...Data) Quantified {
	return vector(func() []Data { return dd })
}
func newVector(dd ...d.Data) Quantified {
	return vector(func() (vec []Data) {
		for _, d := range dd {
			vec = append(vec, newData(d))
		}
		return vec
	})
}

// implements functions/sliceable interface
func (v vector) Head() Data {
	if v.Len() > 0 {
		return v.Vector()[0]
	}
	return nil
}
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
