/*
DATA CONSTRUCTORS

  implementations of 'precedence types', ake functional base-/ and collection types
*/
package functions

import (
	d "github.com/JoergReinhardt/godeep/data"
)

type (
	ConstFnc  func() Functional
	UnaryFnc  func(Functional) Functional
	BinaryFnc func(a, b Functional) Functional
	NaryFnc   func(...Functional) Functional
	VecFnc    func() []Functional
	TupleFnc  func() (Vectorized, []d.BitFlag)
	RecordFnc func() (Tupled, []Paired)
	ListFnc   func() (Functional, Recursive)
)

// ONSTANT
// constant also conains immutable data that may be an instance of a type of
// the data package, or result of a function call guarantueed to allways return
// the same value.
func NewConstant(fnc func(...Functional) Functional) ConstFnc {
	return ConstFnc(func() Functional { return fnc() })
}
func (c ConstFnc) Kind() d.BitFlag                 { return Value.Flag() }
func (c ConstFnc) Flag() d.BitFlag                 { return c().Flag() }
func (c ConstFnc) Ident() Functional               { return c }
func (c ConstFnc) Eval() d.Data                    { return c() }
func (c ConstFnc) Call(d ...Functional) Functional { return c() }
func (c ConstFnc) String() string                  { return "ϝ → т" }

func NewBinaryFnc(fnc func(a, b Functional) Functional) BinaryFnc {
	return BinaryFnc(func(a, b Functional) Functional { return fnc(a, b) })
}

func (u UnaryFnc) Kind() d.BitFlag                 { return Value.Flag() }
func (u UnaryFnc) Flag() d.BitFlag                 { return d.Function.Flag() }
func (u UnaryFnc) Ident() Functional               { return u }
func (u UnaryFnc) Eval() d.Data                    { return u }
func (u UnaryFnc) Call(d ...Functional) Functional { return u(d[0]) }
func (u UnaryFnc) String() string                  { return "т → ϝ → т" }

func (b BinaryFnc) Kind() d.BitFlag                 { return Value.Flag() }
func (b BinaryFnc) Flag() d.BitFlag                 { return d.Function.Flag() }
func (b BinaryFnc) Ident() Functional               { return b }
func (b BinaryFnc) Eval() d.Data                    { return b }
func (b BinaryFnc) Call(d ...Functional) Functional { return b(d[0], d[1]) }
func (b BinaryFnc) String() string                  { return "т → т → ϝ → т" }

func (n NaryFnc) Kind() d.BitFlag                 { return Value.Flag() }
func (n NaryFnc) Flag() d.BitFlag                 { return d.Function.Flag() }
func (n NaryFnc) Ident() Functional               { return n }
func (n NaryFnc) Eval() d.Data                    { return n }
func (n NaryFnc) Call(d ...Functional) Functional { return n(d...) }
func (n NaryFnc) String() string                  { return "[т...] → ϝ → т" }

// VECTOR
// vector keeps a slice of data instances
func conVec(vec Vectorized, dd ...Functional) Vectorized {
	return vectorConstructor(append(vec.Slice(), dd...)...)
}
func vectorConstructor(dd ...Functional) Vectorized {
	return VecFnc(func() []Functional { return dd })
}
func NewVector(dd ...Functional) Vectorized {
	return VecFnc(func() (vec []Functional) {
		for _, dat := range dd {
			vec = append(vec, New(dat))
		}
		return vec
	})
}

// base implementation functions/sliceable interface
func (v VecFnc) Head() Functional {
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
	if len(v()) > 0 {
		return d.Vector.Flag() | v.Head().Flag()
	}
	return d.Vector.Flag() | d.Nil.Flag()
}
func (v VecFnc) Kind() d.BitFlag   { return Vector.Flag() }
func (v VecFnc) Ident() Functional { return v }
func (v VecFnc) Eval() d.Data      { return NewVector(v()...) }
func (v VecFnc) Tail() []Functional {
	if v.Len() > 1 {
		return v.Vector()[1:]
	}
	return nil
}
func (v VecFnc) DeCap() (Functional, []Functional) {
	return v.Head(), v.Tail()
}
func (v VecFnc) Vector() []Functional { return v() }
func (v VecFnc) Slice() []Functional  { return v() }
func (v VecFnc) Call(d ...Functional) Functional {
	if len(d) > 0 {
		conVec(v, d...)
	}
	return v
}

// RECURSIVE
// base implementation of linked lists
func conRecurse(rec Recursive, dd ...Functional) Recursive {
	if len(dd) > 0 {
		if len(dd) > 1 {
			return ListFnc(func() (Functional, Recursive) {
				return dd[0], conRecurse(rec, dd...)
			})
		}
		return ListFnc(func() (Functional, Recursive) {
			return dd[0], rec
		})
	}
	return nil
}
func NewRecursiveList(dd ...Functional) Recursive {
	if len(dd) > 0 {
		if len(dd) > 1 {
			return ListFnc(func() (Functional, Recursive) { return dd[0], NewRecursiveList(dd[1:]...) })
		}
		return ListFnc(func() (Functional, Recursive) { return dd[0], nil })
	}
	return nil
}
func (l ListFnc) Ident() Functional { return l }
func (l ListFnc) Eval() d.Data      { return NewPair(l.Head(), l.Tail()) }
func (l ListFnc) Head() Functional  { h, _ := l(); return h }
func (l ListFnc) Tail() Recursive   { _, t := l(); return t }
func (l ListFnc) Call(d ...Functional) Functional {
	if len(d) > 0 {
		return conRecurse(l, d...)
	}
	return l
}
func (l ListFnc) DeCap() (Functional, Recursive) { return l() }
func (l ListFnc) Kind() d.BitFlag                { return List.Flag() }
func (l ListFnc) Flag() d.BitFlag                { return d.List.Flag() | l.Head().Flag() }
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
func conTuple(tup Tupled, dat ...Functional) Tupled {
	return NewTuple(append(tup.Slice(), dat...)...)
}
func NewTuple(dat ...Functional) Tupled {
	var flags []d.BitFlag
	for _, data := range dat {
		flags = append(flags, data.Flag())
	}
	var vec = vectorConstructor(dat...)
	return TupleFnc(func() (Vectorized, []d.BitFlag) {
		return vec, flags
	})
}
func (t TupleFnc) Flags() []d.BitFlag                { _, f := t(); return f }
func (t TupleFnc) DeCap() (Functional, []Functional) { v, _ := t(); return v.DeCap() }
func (t TupleFnc) Slice() []Functional               { v, _ := t(); return v.Slice() }
func (t TupleFnc) Head() Functional                  { v, _ := t(); return v.Head() }
func (t TupleFnc) Tail() []Functional                { v, _ := t(); return v.Tail() }
func (t TupleFnc) Empty() bool                       { v, _ := t(); return v.Empty() }
func (t TupleFnc) Len() int                          { v, _ := t(); return v.Len() }
func (t TupleFnc) Flag() d.BitFlag {
	var flag d.BitFlag
	for _, elem := range t.Slice() {
		flag = flag | elem.Flag()
	}
	return flag | d.Tuple.Flag()
}
func (t TupleFnc) Kind() d.BitFlag   { return Tuple.Flag() }
func (t TupleFnc) Eval() d.Data      { return NewVector(t.Slice()...) }
func (t TupleFnc) Ident() Functional { return t }
func (t TupleFnc) Call(d ...Functional) Functional {
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
	var dat = []Functional{}
	for _, pair := range pairs {
		sig = append(sig, NewPair(pair.Left(), New(pair.Right().Flag())))
		dat = append(dat, pair)
	}
	var tup = NewTuple(dat...)
	return RecordFnc(func() (Tupled, []Paired) {
		return tup, sig
	})
}
func (r RecordFnc) Ident() Functional { return r }
func (r RecordFnc) Eval() d.Data      { return r.Tuple() }
func (r RecordFnc) Call(d ...Functional) Functional {
	if len(d) > 0 {
		var pairs = []Paired{}
		for _, pair := range d {
			pairs = append(pairs, pair.(Paired))
		}
		return conRecord(r, pairs...)
	}
	return r
}
func (r RecordFnc) ArgSig() []Paired                  { _, pairs := r(); return pairs }
func (r RecordFnc) Tuple() Tupled                     { tup, _ := r(); return tup }
func (r RecordFnc) DeCap() (Functional, []Functional) { return r.Tuple().DeCap() }
func (r RecordFnc) Head() Functional                  { return r.Tuple().Head() }
func (r RecordFnc) Tail() []Functional                { return r.Tuple().Tail() }
func (r RecordFnc) Slice() []Functional               { return r.Tuple().Slice() }
func (r RecordFnc) Empty() bool                       { return r.Tuple().Empty() }
func (r RecordFnc) Len() int                          { return r.Tuple().Len() }
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
func (r RecordFnc) Kind() d.BitFlag { return Record.Flag() }
