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
	unary    func(Data) Data
	binary   func(a, b Data) Data
	nary     func(...Data) Data
	vector   func() []Data                    // <- indexable native golang slice of data instances
	tuple    func() (Vectorized, []d.BitFlag) // <- indexable native golang slice of fixed length and type signature
	list     func() (Data, Recursive)
)

// CONSTANT
// constant also conains immutable data, but it may be the result of a constant experssion
func newConstant(dat Data) Data    { return constant(func() Data { return dat.(Functional).Eval() }) }
func (c constant) Flag() d.BitFlag { return Constant.Flag() }
func (c constant) Type() Flag      { return newFlag(Constant, c().Flag()) }
func (c constant) String() string  { return c().(d.Data).String() }
func (c constant) Eval() Data      { return c }

// TUPLE

// VECTOR
// vector keeps a slice of data instances
func vectorConstructor(dd ...Data) Vectorized {
	return vector(func() []Data { return dd })
}
func newVector(dd ...d.Data) Vectorized {
	return vector(func() (vec []Data) {
		for _, d := range dd {
			vec = append(vec, newData(d))
		}
		return vec
	})
}

// base implementation functions/sliceable interface
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
func (v vector) Eval() Data { return v }
func (v vector) Tail() []Data {
	if v.Len() > 1 {
		return v.Vector()[1:]
	}
	return nil
}
func (v vector) DeCap() (Data, []Data) {
	return v.Head(), v.Tail()
}
func (v vector) Vector() []Data { return v() }
func (v vector) Slice() []Data  { return v() }

// LINKED LIST
// base implementation of linked lists
func conRecursive(d ...Data) Recursive {
	if len(d) > 0 {
		if len(d) > 1 {
			return list(func() (Data, Recursive) { return d[0], conRecursive(d[1:]...) })
		}
		return list(func() (Data, Recursive) { return d[0], nil })
	}
	return nil
}
func (l list) Eval() Data               { return l }
func (l list) Head() Data               { h, _ := l(); return h }
func (l list) Tail() Recursive          { _, t := l(); return t }
func (l list) DeCap() (Data, Recursive) { return l() }
func (l list) Flag() d.BitFlag          { return d.Function.Flag() }
func (l list) Empty() bool {
	var h, _ = l()
	if h != nil {
		return false
	}
	return true
}
func (l list) Len() int {
	var _, t = l()
	if t != nil {
		return 1 + t.Len()
	}
	return 1
}
func (l list) String() string {
	var h, t = l()
	if t != nil {
		return h.String() + ", " + t.String()
	}
	return h.String()
}
func (l list) Type() Flag {
	return newFlag(List,
		d.Slice.Flag()|
			d.Parameter.Flag()|
			l.Flag())
}

// TUPLE
func conTuple(dat ...Data) Tupled {
	var flags = []d.BitFlag{}
	for _, data := range dat {
		flags = append(flags, data.Flag())
	}
	var vec = vectorConstructor(dat...)
	return tuple(func() (Vectorized, []d.BitFlag) {
		return vec, flags
	})
}
func (t tuple) Arity() Arity          { _, f := t(); return Arity(len(f)) }
func (t tuple) Sig() []d.BitFlag      { _, f := t(); return f }
func (t tuple) DeCap() (Data, []Data) { v, _ := t(); return v.DeCap() }
func (t tuple) Slice() []Data         { v, _ := t(); return v.Slice() }
func (t tuple) Head() Data            { v, _ := t(); return v.Head() }
func (t tuple) Tail() []Data          { v, _ := t(); return v.Tail() }
func (t tuple) Empty() bool           { v, _ := t(); return v.Empty() }
func (t tuple) Len() int              { v, _ := t(); return v.Len() }
func (t tuple) Flag() d.BitFlag       { return d.Function.Flag() }
func (t tuple) Eval() Data            { return t }
func (t tuple) String() string {
	var slice []d.Data
	var v, _ = t()
	for _, dat := range v.(vector)() {
		slice = append(slice, dat)
	}
	return d.StringSlice("∙", "(", ")", slice...)
}
func (t tuple) Type() Flag {
	return newFlag(Tuple,
		d.Slice.Flag()|
			d.Parameter.Flag()|
			t.Flag())
}
