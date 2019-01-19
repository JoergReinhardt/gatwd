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
  parameters and return values are enclosed by thetconstructor.
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
	tuple    func() (Vectorized, []d.BitFlag) // <- indexable native golang slice of fixed length & type signature
	record   func() (Tupled, []Paired)        // <- indexable native golang slice of fixed length, type signature & param keys
	list     func() (Data, Recursive)
)

// ONSTANT
// constant also conains immutable data, but it may be the result of a constant experssion
func newConstant(dat Data) constant    { return constant(func() Data { return dat }) }
func (c constant) Flag() d.BitFlag     { return d.Definition.Flag() }
func (c constant) Type() Flag          { return newFlag(0, Constant, c().Flag()) }
func (c constant) Ident() Data         { return c }
func (c constant) Call(d ...Data) Data { return c() }

func (u unary) Call(d ...Data) Data {
	if len(d) > 0 {
		return u(d[0])
	}
	return nil
}
func (b binary) Call(d ...Data) Data {
	if len(d) > 1 {
		return b(d[0], d[1])
	}
	return nil
}
func (n nary) Call(d ...Data) Data {
	if len(d) > 0 {
		return n(d...)
	}
	return nil
}

// TUPLE

// VECTOR
// vector keeps a slice of data instances
func conVec(vec Vectorized, dd ...Data) Vectorized {
	return vectorConstructor(append(vec.Slice(), dd...)...)
}
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
func (v vector) Flag() d.BitFlag { return d.Vector.Flag() }
func (v vector) Ident() Data     { return v }
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
func (v vector) Call(d ...Data) Data {
	if len(d) > 0 {
		conVec(v, d...)
	}
	return v
}

// LINKED LIST
// base implementation of linked lists
func conRecurse(rec Recursive, d ...Data) Recursive {
	if len(d) > 0 {
		if len(d) > 1 {
			return list(func() (Data, Recursive) {
				return d[0], conRecurse(rec, d...)
			})
		}
		return list(func() (Data, Recursive) {
			return d[0], rec
		})
	}
	return nil
}
func newRecursive(d ...Data) Recursive {
	if len(d) > 0 {
		if len(d) > 1 {
			return list(func() (Data, Recursive) { return d[0], newRecursive(d[1:]...) })
		}
		return list(func() (Data, Recursive) { return d[0], nil })
	}
	return nil
}
func (l list) Ident() Data     { return l }
func (l list) Head() Data      { h, _ := l(); return h }
func (l list) Tail() Recursive { _, t := l(); return t }
func (l list) Call(d ...Data) Data {
	if len(d) > 0 {
		return conRecurse(l, d...)
	}
	return l
}
func (l list) DeCap() (Data, Recursive) { return l() }
func (l list) Flag() d.BitFlag          { return d.List.Flag() }
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

// TUPLE
func conTuple(tup Tupled, dat ...Data) Tupled {
	return newTuple(append(tup.Slice(), dat...)...)
}
func newTuple(dat ...Data) Tupled {
	var flags []d.BitFlag
	for _, data := range dat {
		flags = append(flags, data.Flag())
	}
	var vec = vectorConstructor(dat...)
	return tuple(func() (Vectorized, []d.BitFlag) {
		return vec, flags
	})
}
func (t tuple) Arity() Arity          { _, f := t(); return Arity(len(f)) }
func (t tuple) Flags() []d.BitFlag    { _, f := t(); return f }
func (t tuple) DeCap() (Data, []Data) { v, _ := t(); return v.DeCap() }
func (t tuple) Slice() []Data         { v, _ := t(); return v.Slice() }
func (t tuple) Head() Data            { v, _ := t(); return v.Head() }
func (t tuple) Tail() []Data          { v, _ := t(); return v.Tail() }
func (t tuple) Empty() bool           { v, _ := t(); return v.Empty() }
func (t tuple) Len() int              { v, _ := t(); return v.Len() }
func (t tuple) Flag() d.BitFlag       { return d.Tuple.Flag() }
func (t tuple) Ident() Data           { return t }
func (t tuple) Call(d ...Data) Data {
	if len(d) > 0 {
		return conTuple(t, d...)
	}
	return t
}

// RECORD
func conRecord(rec Recorded, pairs ...Paired) Recorded {
	tup, ps := rec.(record)()
	if len(pairs) > 0 {
		return record(func() (Tupled, []Paired) {
			return tup, pairs
		})
	}
	return record(func() (Tupled, []Paired) {
		return tup, ps
	})
}
func newRecord(pairs ...Paired) Recorded {
	var sig = []Paired{}
	var dat = []Data{}
	for _, pair := range pairs {
		sig = append(sig, newPair(pair.Left(), pair.Right().Flag()))
		dat = append(dat, pair)
	}
	var tup = newTuple(dat...)
	return record(func() (Tupled, []Paired) {
		return tup, sig
	})
}
func (r record) Ident() Data { return r }
func (r record) Call(d ...Data) Data {
	if len(d) > 0 {
		var pairs = []Paired{}
		for _, pair := range d {
			pairs = append(pairs, pair.(Paired))
		}
		return conRecord(r, pairs...)
	}
	return r
}
func (r record) Arity() Arity          { _, pairs := r(); return Arity(len(pairs)) }
func (r record) ArgSig() []Paired      { _, pairs := r(); return pairs }
func (r record) Tuple() Tupled         { tup, _ := r(); return tup }
func (r record) DeCap() (Data, []Data) { return r.Tuple().DeCap() }
func (r record) Head() Data            { return r.Tuple().Head() }
func (r record) Tail() []Data          { return r.Tuple().Tail() }
func (r record) Slice() []Data         { return r.Tuple().Slice() }
func (r record) Empty() bool           { return r.Tuple().Empty() }
func (r record) Len() int              { return r.Tuple().Len() }
func (r record) Flag() d.BitFlag       { return d.Record.Flag() }
