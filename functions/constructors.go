/*
DATA CONSTRUCTORS

  implementations of 'precedence types', ake functional base-/ and collection types
*/
package functions

import (
	d "github.com/JoergReinhardt/godeep/data"
)

type (
	ConstFnc  func() d.Data
	UnaryFnc  func(d.Data) d.Data
	BinaryFnc func(a, b d.Data) d.Data
	NaryFnc   func(...d.Data) d.Data
	VecFnc    func() []d.Data
	TupleFnc  func() (Vectorized, []d.BitFlag)
	RecordFnc func() (Tupled, []Paired)
	ListFnc   func() (d.Data, Recursive)
)

// ONSTANT
// constant also conains immutable data that may be an instance of a type of
// the data package, or result of a function call guarantueed to allways return
// the same value.
func NewConstant(dat d.Data) ConstFnc      { return ConstFnc(func() d.Data { return dat }) }
func (c ConstFnc) Kind() BitFlag           { return Value.Flag() }
func (c ConstFnc) Flag() d.BitFlag         { return c().Flag() }
func (c ConstFnc) Ident() d.Data           { return c }
func (c ConstFnc) Eval() d.Data            { return c() }
func (c ConstFnc) Call(d ...d.Data) d.Data { return c() }

func (u UnaryFnc) Call(d ...d.Data) d.Data {
	if len(d) > 0 {
		return u(d[0])
	}
	return nil
}
func (b BinaryFnc) Call(d ...d.Data) d.Data {
	if len(d) > 1 {
		return b(d[0], d[1])
	}
	return nil
}
func (n NaryFnc) Call(d ...d.Data) d.Data {
	if len(d) > 0 {
		return n(d...)
	}
	return nil
}

// VECTOR
// vector keeps a slice of data instances
func conVec(vec Vectorized, dd ...d.Data) Vectorized {
	return vectorConstructor(append(vec.Slice(), dd...)...)
}
func vectorConstructor(dd ...d.Data) Vectorized {
	return VecFnc(func() []d.Data { return dd })
}
func NewVector(dd ...d.Data) Vectorized {
	return VecFnc(func() (vec []d.Data) {
		for _, d := range dd {
			vec = append(vec, NewValue(d))
		}
		return vec
	})
}

// base implementation functions/sliceable interface
func (v VecFnc) Head() d.Data {
	if v.Len() > 0 {
		return v.Vector()[0]
	}
	return nil
}
func (v VecFnc) Len() int { return len(v()) }
func (v VecFnc) Empty() bool {
	if len(v()) > 0 {
		for _, dat := range v() {
			if !d.Nil.Flag().Match(dat.Flag()) {
				return false
			}
		}
	}
	return true
}
func (v VecFnc) Flag() d.BitFlag {
	return d.Vector.Flag() | v.Head().Flag()
}
func (v VecFnc) Kind() BitFlag     { return Vector.Flag() }
func (v VecFnc) Ident() Functional { return v }
func (v VecFnc) Eval() d.Data      { return NewVector(v()...) }
func (v VecFnc) Tail() []d.Data {
	if v.Len() > 1 {
		return v.Vector()[1:]
	}
	return nil
}
func (v VecFnc) DeCap() (d.Data, []d.Data) {
	return v.Head(), v.Tail()
}
func (v VecFnc) Vector() []d.Data { return v() }
func (v VecFnc) Slice() []d.Data  { return v() }
func (v VecFnc) Call(d ...d.Data) d.Data {
	if len(d) > 0 {
		conVec(v, d...)
	}
	return v
}

// RECURSIVE
// base implementation of linked lists
func conRecurse(rec Recursive, dd ...d.Data) Recursive {
	if len(dd) > 0 {
		if len(dd) > 1 {
			return ListFnc(func() (d.Data, Recursive) {
				return dd[0], conRecurse(rec, dd...)
			})
		}
		return ListFnc(func() (d.Data, Recursive) {
			return dd[0], rec
		})
	}
	return nil
}
func NewRecursiveList(dd ...d.Data) Recursive {
	if len(dd) > 0 {
		if len(dd) > 1 {
			return ListFnc(func() (d.Data, Recursive) { return dd[0], NewRecursiveList(dd[1:]...) })
		}
		return ListFnc(func() (d.Data, Recursive) { return dd[0], nil })
	}
	return nil
}
func (l ListFnc) Ident() d.Data   { return l }
func (l ListFnc) Eval() d.Data    { return NewPair(l.Head(), l.Tail()) }
func (l ListFnc) Head() d.Data    { h, _ := l(); return h }
func (l ListFnc) Tail() Recursive { _, t := l(); return t }
func (l ListFnc) Call(d ...d.Data) d.Data {
	if len(d) > 0 {
		return conRecurse(l, d...)
	}
	return l
}
func (l ListFnc) DeCap() (d.Data, Recursive) { return l() }
func (l ListFnc) Kind() BitFlag              { return List.Flag() }
func (l ListFnc) Flag() d.BitFlag            { return d.List.Flag() | l.Head().Flag() }
func (l ListFnc) Empty() bool {
	var h, _ = l()
	if h != nil {
		return false
	}
	return true
}
func (l ListFnc) Len() int {
	var _, t = l()
	if t != nil {
		return 1 + t.Len()
	}
	return 1
}

// TUPLE
func conTuple(tup Tupled, dat ...d.Data) Tupled {
	return NewTuple(append(tup.Slice(), dat...)...)
}
func NewTuple(dat ...d.Data) Tupled {
	var flags []d.BitFlag
	for _, data := range dat {
		flags = append(flags, data.Flag())
	}
	var vec = vectorConstructor(dat...)
	return TupleFnc(func() (Vectorized, []d.BitFlag) {
		return vec, flags
	})
}
func (t TupleFnc) Flags() []d.BitFlag        { _, f := t(); return f }
func (t TupleFnc) DeCap() (d.Data, []d.Data) { v, _ := t(); return v.DeCap() }
func (t TupleFnc) Slice() []d.Data           { v, _ := t(); return v.Slice() }
func (t TupleFnc) Head() d.Data              { v, _ := t(); return v.Head() }
func (t TupleFnc) Tail() []d.Data            { v, _ := t(); return v.Tail() }
func (t TupleFnc) Empty() bool               { v, _ := t(); return v.Empty() }
func (t TupleFnc) Len() int                  { v, _ := t(); return v.Len() }
func (t TupleFnc) Flag() d.BitFlag {
	var flag d.BitFlag
	for _, elem := range t.Slice() {
		flag = flag | elem.Flag()
	}
	return flag | d.Tuple.Flag()
}
func (t TupleFnc) Kind() BitFlag     { return Tuple.Flag() }
func (t TupleFnc) Eval() d.Data      { return NewVector(t.Slice()...) }
func (t TupleFnc) Ident() Functional { return t }
func (t TupleFnc) Call(d ...d.Data) d.Data {
	if len(d) > 0 {
		return conTuple(t, d...)
	}
	return t
}

// RECORD
func conRecord(rec Recorded, pairs ...Paired) Recorded {
	tup, ps := rec.(RecordFnc)()
	if len(pairs) > 0 {
		return RecordFnc(func() (Tupled, []Paired) {
			return tup, pairs
		})
	}
	return RecordFnc(func() (Tupled, []Paired) {
		return tup, ps
	})
}
func NewRecord(pairs ...Paired) Recorded {
	var sig = []Paired{}
	var dat = []d.Data{}
	for _, pair := range pairs {
		sig = append(sig, NewPair(pair.Left(), pair.Right().Flag()))
		dat = append(dat, pair)
	}
	var tup = NewTuple(dat...)
	return RecordFnc(func() (Tupled, []Paired) {
		return tup, sig
	})
}
func (r RecordFnc) Ident() d.Data { return r }
func (r RecordFnc) Eval() d.Data  { return r.Tuple() }
func (r RecordFnc) Call(d ...d.Data) d.Data {
	if len(d) > 0 {
		var pairs = []Paired{}
		for _, pair := range d {
			pairs = append(pairs, pair.(Paired))
		}
		return conRecord(r, pairs...)
	}
	return r
}
func (r RecordFnc) ArgSig() []Paired          { _, pairs := r(); return pairs }
func (r RecordFnc) Tuple() Tupled             { tup, _ := r(); return tup }
func (r RecordFnc) DeCap() (d.Data, []d.Data) { return r.Tuple().DeCap() }
func (r RecordFnc) Head() d.Data              { return r.Tuple().Head() }
func (r RecordFnc) Tail() []d.Data            { return r.Tuple().Tail() }
func (r RecordFnc) Slice() []d.Data           { return r.Tuple().Slice() }
func (r RecordFnc) Empty() bool               { return r.Tuple().Empty() }
func (r RecordFnc) Len() int                  { return r.Tuple().Len() }
func (r RecordFnc) Get(p Functional) Paired {
	_, pairs := r()
	ps := newPairSorter(pairs...)
	ps.Sort(d.Symbolic)
	idx := ps.Search(p)
	if idx != -1 {
		return ps[idx]
	}
	return nil
}
func (r RecordFnc) Flag() d.BitFlag { return d.Record.Flag() }
func (r RecordFnc) Kind() BitFlag   { return Record.Flag() }
