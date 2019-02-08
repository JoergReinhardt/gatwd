/*
DATA CONSTRUCTORS

  implementations of 'precedence types', ake functional base-/ and collection types
*/
package functions

import (
	d "github.com/JoergReinhardt/godeep/data"
)

type (
	// FUNCTION VALUES
	ConstFnc  func() Value
	UnaryFnc  func(Value) Value
	BinaryFnc func(a, b Value) Value
	NaryFnc   func(...Value) Value
	// DATA CONSTRUCTORS
	TupleFnc  func() (Vectorized, []d.BitFlag)
	RecordFnc func() (Tupled, []Paired)
	ListFnc   func() (Value, Recursive)
	VecFnc    func() []Value
	AccVecFnc func() []Paired
)

// ONSTANT
// constant also conains immutable data that may be an instance of a type of
// the data package, or result of a function call guarantueed to allways return
// the same value.
func NewConstant(fnc func(...Value) Value) ConstFnc {
	return ConstFnc(func() Value { return fnc() })
}
func (c ConstFnc) TypeHO() TyHigherOrder         { return Data }
func (c ConstFnc) TypePrim() d.TyPrimitive       { return c().TypePrim() }
func (c ConstFnc) Ident() Value                  { return c }
func (c ConstFnc) Eval(p ...d.Primary) d.Primary { return c() }
func (c ConstFnc) Call(d ...Value) Value {
	return c().(ConstFnc)()
}

///// UNARY FUNCTION
func NewUnaryFnc(fnc func(f Value) Value) UnaryFnc {
	return UnaryFnc(func(f Value) Value { return fnc(f) })
}
func (u UnaryFnc) TypeHO() TyHigherOrder         { return Data }
func (u UnaryFnc) TypePrim() d.TyPrimitive       { return d.Function.TypePrim() }
func (u UnaryFnc) Ident() Value                  { return u }
func (u UnaryFnc) Eval(p ...d.Primary) d.Primary { return u }
func (u UnaryFnc) Call(d ...Value) Value {
	return u(d[0])
}

///// BINARY FUNCTION
func NewBinaryFnc(fnc func(a, b Value) Value) BinaryFnc {
	return BinaryFnc(func(a, b Value) Value { return fnc(a, b) })
}
func (b BinaryFnc) TypeHO() TyHigherOrder         { return Data }
func (b BinaryFnc) TypePrim() d.TyPrimitive       { return d.Function.TypePrim() }
func (b BinaryFnc) Ident() Value                  { return b }
func (b BinaryFnc) Eval(p ...d.Primary) d.Primary { return b }
func (b BinaryFnc) Call(d ...Value) Value         { return b(d[0], d[1]) }

///// NARY FUNCTION
func NewNaryFnc(fnc func(f ...Value) Value) NaryFnc {
	return NaryFnc(func(f ...Value) Value { return fnc(f...) })
}
func (n NaryFnc) TypeHO() TyHigherOrder         { return Data }
func (n NaryFnc) TypePrim() d.TyPrimitive       { return d.Function.TypePrim() }
func (n NaryFnc) Ident() Value                  { return n }
func (n NaryFnc) Eval(p ...d.Primary) d.Primary { return n }
func (n NaryFnc) Call(d ...Value) Value         { return n(d...) }

/////////////////////////////////////////////////////////
// ASSOCIATIVE VECTOR (VECTOR OF PAIRS)
func conAccVec(vec Accessable, pp ...Paired) Accessable {
	return accVecConstructor(append(vec.Pairs(), pp...)...)
}
func accVecConstructor(pp ...Paired) Accessable {
	return AccVecFnc(func() []Paired { return pp })
}
func NewAccVector(pp ...Paired) Accessable {
	return AccVecFnc(func() (pairs []Paired) {
		for _, pair := range pp {
			pairs = append(pairs, pair)
		}
		return pairs
	})
}
func (v AccVecFnc) Slice() []Value {
	var fncs = []Value{}
	for _, pair := range v() {
		fncs = append(fncs, NewPair(pair.Left(), pair.Right()))
	}
	return fncs
}
func (v AccVecFnc) Pairs() []Paired { return v() }
func (v AccVecFnc) Head() Paired {
	if v.Len() > 0 {
		return v.Pairs()[0]
	}
	return nil
}
func (v AccVecFnc) Tail() []Paired {
	if v.Len() > 1 {
		return v.Pairs()[1:]
	}
	return nil
}
func (v AccVecFnc) Eval(p ...d.Primary) d.Primary {
	var slice = d.DataSlice{}
	for _, pair := range v() {
		d.SliceAppend(slice, d.NewPair(pair.Left(), pair.Right()))
	}
	return slice
}
func (v AccVecFnc) Len() int                   { return len(v()) }
func (v AccVecFnc) Empty() bool                { return ElemEmpty(v.Head()) && (len(v.Tail()) == 0) }
func (v AccVecFnc) Get(praed Value) Paired     { return newPairSorter(v()...).Get(praed) }
func (v AccVecFnc) Range(praed Value) []Paired { return newPairSorter(v()...).Range(praed) }
func (v AccVecFnc) Search(praed Value) int     { return newPairSorter(v()...).Search(praed) }
func (v AccVecFnc) Sort(flag d.TyPrimitive) {
	var ps = newPairSorter(v()...)
	ps.Sort(flag)
	v = NewAccVector(ps...).(AccVecFnc)
}

///////////////////////////////////////////////////
// VECTOR
// vector keeps a slice of data instances
func conVec(vec Vectorized, dd ...Value) Vectorized {
	return vectorConstructor(append(vec.Slice(), dd...)...)
}
func vectorConstructor(dd ...Value) Vectorized {
	return VecFnc(func() []Value { return dd })
}
func NewVector(dd ...Value) Vectorized {
	return VecFnc(func() (vec []Value) {
		for _, dat := range dd {
			vec = append(vec, New(dat))
		}
		return vec
	})
}

func (v AccVecFnc) DeCap() (Paired, []Paired) {
	return v.Head(), v.Tail()
}
func (v AccVecFnc) TypePrim() d.TyPrimitive {
	if len(v()) > 0 {
		return d.Vector | v.Head().TypePrim()
	}
	return d.Vector | d.Nil.TypePrim()
}
func (v AccVecFnc) TypeHO() TyHigherOrder { return Vector }

// base implementation functions/sliceable interface
func (v VecFnc) Head() Value {
	if v.Len() > 0 {
		return v.Vector()[0]
	}
	return nil
}
func (v VecFnc) Tail() []Value {
	if v.Len() > 1 {
		return v.Vector()[1:]
	}
	return nil
}
func (v VecFnc) Len() int { return len(v()) }
func (v VecFnc) Empty() bool {
	if len(v()) > 0 {
		for _, dat := range v() {
			if !d.Nil.TypePrim().Flag().Match(dat.TypePrim().Flag()) {
				return false
			}
		}
	}
	return true
}
func (v VecFnc) TypePrim() d.TyPrimitive {
	if len(v()) > 0 {
		return d.Vector.TypePrim() | v.Head().TypePrim()
	}
	return d.Vector.TypePrim() | d.Nil.TypePrim()
}
func (v VecFnc) TypeHO() TyHigherOrder         { return Vector }
func (v VecFnc) Ident() Value                  { return v }
func (v VecFnc) Eval(p ...d.Primary) d.Primary { return NewVector(v()...) }
func (v VecFnc) DeCap() (Value, []Value) {
	return v.Head(), v.Tail()
}
func (v VecFnc) Vector() []Value { return v() }
func (v VecFnc) Slice() []Value  { return v() }
func (v VecFnc) Call(d ...Value) Value {
	if len(d) > 0 {
		conVec(v, d...)
	}
	return v
}
func (v VecFnc) Search(praed Value) int { return newDataSorter(v()...).Search(praed) }
func (v VecFnc) Sort(flag d.TyPrimitive) {
	var ps = newDataSorter(v()...)
	ps.Sort(flag)
	v = NewVector(ps...).(VecFnc)
}

// RECURSIVE LIST
// base implementation of linked lists
func conRecurse(rec Recursive, dd ...Value) Recursive {
	if len(dd) > 0 {
		if len(dd) > 1 {
			return ListFnc(func() (Value, Recursive) {
				return dd[0], conRecurse(rec, dd...)
			})
		}
		return ListFnc(func() (Value, Recursive) {
			return dd[0], rec
		})
	}
	return nil
}
func NewRecursiveList(dd ...Value) ListFnc {
	if len(dd) > 0 {
		if len(dd) > 1 {
			return ListFnc(func() (Value, Recursive) { return dd[0], NewRecursiveList(dd[1:]...) })
		}
		return ListFnc(func() (Value, Recursive) { return dd[0], nil })
	}
	return nil
}
func (l ListFnc) Ident() Value                  { return l }
func (l ListFnc) Eval(p ...d.Primary) d.Primary { return NewPair(l.Head(), l.Tail()) }
func (l ListFnc) Head() Value                   { h, _ := l(); return h }
func (l ListFnc) Tail() Recursive               { _, t := l(); return t }
func (l ListFnc) Call(d ...Value) Value {
	if len(d) > 0 {
		return conRecurse(l, d...)
	}
	return l
}
func (l ListFnc) DeCap() (Value, Recursive) { return l() }
func (l ListFnc) TypeHO() TyHigherOrder     { return List }
func (l ListFnc) TypePrim() d.TyPrimitive   { return d.List.TypePrim() | l.Head().TypePrim() }
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
func conTuple(tup Tupled, dat ...Value) Tupled {
	return NewTuple(append(tup.Slice(), dat...)...)
}
func NewTuple(dat ...Value) Tupled {
	var flags []d.BitFlag
	for _, data := range dat {
		flags = append(flags, data.TypePrim().Flag())
	}
	var vec = vectorConstructor(dat...)
	return TupleFnc(func() (Vectorized, []d.BitFlag) {
		return vec, flags
	})
}
func (t TupleFnc) Flags() []d.BitFlag      { _, f := t(); return f }
func (t TupleFnc) DeCap() (Value, []Value) { v, _ := t(); return v.DeCap() }
func (t TupleFnc) Slice() []Value          { v, _ := t(); return v.Slice() }
func (t TupleFnc) Head() Value             { v, _ := t(); return v.Head() }
func (t TupleFnc) Tail() []Value           { v, _ := t(); return v.Tail() }
func (t TupleFnc) Empty() bool             { v, _ := t(); return v.Empty() }
func (t TupleFnc) Len() int                { v, _ := t(); return v.Len() }
func (t TupleFnc) TypePrim() d.TyPrimitive {
	var flag d.BitFlag
	for _, elem := range t.Slice() {
		flag = flag | elem.TypePrim().Flag()
	}
	return d.TyPrimitive(flag | d.Tuple.Flag())
}
func (t TupleFnc) TypeHO() TyHigherOrder         { return Tuple }
func (t TupleFnc) Eval(p ...d.Primary) d.Primary { return NewVector(t.Slice()...) }
func (t TupleFnc) Ident() Value                  { return t }
func (t TupleFnc) Call(d ...Value) Value {
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
	var dat = []Value{}
	for _, pair := range pairs {
		sig = append(sig, NewPair(pair.Left(), New(pair.Right().TypePrim())))
		dat = append(dat, pair)
	}
	var tup = NewTuple(dat...)
	return RecordFnc(func() (Tupled, []Paired) {
		return tup, sig
	})
}
func (r RecordFnc) Ident() Value                  { return r }
func (r RecordFnc) Eval(p ...d.Primary) d.Primary { return r.Tuple() }
func (r RecordFnc) Call(d ...Value) Value {
	if len(d) > 0 {
		var pairs = []Paired{}
		for _, pair := range d {
			pairs = append(pairs, pair.(Paired))
		}
		return conRecord(r, pairs...)
	}
	return r
}
func (r RecordFnc) ArgSig() []Paired        { _, pairs := r(); return pairs }
func (r RecordFnc) Tuple() Tupled           { tup, _ := r(); return tup }
func (r RecordFnc) DeCap() (Value, []Value) { return r.Tuple().DeCap() }
func (r RecordFnc) Head() Value             { return r.Tuple().Head() }
func (r RecordFnc) Tail() []Value           { return r.Tuple().Tail() }
func (r RecordFnc) Slice() []Value          { return r.Tuple().Slice() }
func (r RecordFnc) Empty() bool             { return r.Tuple().Empty() }
func (r RecordFnc) Len() int                { return r.Tuple().Len() }
func (r RecordFnc) Get(p Value) Paired {
	_, pairs := r()
	ps := newPairSorter(pairs...)
	ps.Sort(d.Symbolic)
	idx := ps.Search(p)
	if idx != -1 {
		return ps[idx]
	}
	return nil
}
func (r RecordFnc) TypePrim() d.TyPrimitive { return d.Record }
func (r RecordFnc) TypeHO() TyHigherOrder   { return Record }
