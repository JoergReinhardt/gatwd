/*
DATA CONSTRUCTORS

  implementations of 'precedence types', ake functional base-/ and collection types
*/
package functions

import (
	d "github.com/JoergReinhardt/gatwd/data"
)

type (
	// CLOSURES OVER NATIVE DATA INSTANCE
	DataCon func() Native
	// CLOSURES OVER COLLENCTIONS OF NATIVES
	PairVal     func() (a, b Value)
	AssocVecFnc func() []Paired
	VecFnc      func() []Value
	ListFnc     func() (Value, ListFnc)
)

func New(ifs ...interface{}) DataCon {
	return DataCon(func() Native { return d.New(ifs...) })
}

func NewFromData(data ...d.Native) Value {
	if len(data) > 1 {
		var values = []Value{}
		for _, native := range data {
			values = append(values, DataCon(func() Native { return native }))
		}
		return NewVector(values...)
	}
	if len(data) == 1 {
		return DataCon(func() Native { return data[0] })
	}
	return NewNone()
}
func (c DataCon) Ident() Value                { return c }
func (c DataCon) TypeFnc() TyFnc              { return Data }
func (c DataCon) TypeNat() d.TyNative         { return c().TypeNat() }
func (c DataCon) String() string              { return c().String() }
func (c DataCon) Eval(p ...d.Native) d.Native { return c().Eval() }
func (c DataCon) Call(d ...Value) Value       { return c }

/////////////////////////////////////////////////////////
// PAIR
func NewPair(l, r Value) PairVal {
	return PairVal(func() (Value, Value) { return l, r })
}
func NewPairFromInterface(l, r interface{}) PairVal {
	return PairVal(func() (Value, Value) { return New(d.New(l)), New(d.New(r)) })
}
func NewPairFromData(l, r d.Native) PairVal {
	return PairVal(func() (Value, Value) { return New(l), New(r) })
}
func (p PairVal) Both() (Value, Value)        { return p() }
func (p PairVal) Pair() Value                 { return p }
func (p PairVal) Left() Value                 { l, _ := p(); return l }
func (p PairVal) Right() Value                { _, r := p(); return r }
func (p PairVal) Acc() Value                  { return p.Left() }
func (p PairVal) Arg() Value                  { return p.Right() }
func (p PairVal) AccType() d.BitFlag          { return p.Left().TypeNat().Flag() }
func (p PairVal) ArgType() d.BitFlag          { return p.Right().TypeNat().Flag() }
func (p PairVal) Ident() Value                { return p }
func (p PairVal) Call(...Value) Value         { return p }
func (p PairVal) Eval(a ...d.Native) d.Native { return d.NewPair(p.Left().Eval(), p.Right().Eval()) }
func (p PairVal) TypeFnc() TyFnc              { return Pair | Function }
func (p PairVal) TypeNat() d.TyNative {
	return d.Pair.TypeNat() | p.Left().TypeNat() | p.Right().TypeNat()
}
func (p PairVal) Empty() bool {
	return ElemEmpty(p.Left()) && ElemEmpty(p.Right())
}

///////////////////////////////////////////////////////////////////////
//// ASSOCIATIVE VECTOR (VECTOR OF PAIRS)
///
// associative array that uses pairs left field as accessor for sort & search
func conAssocVec(vec Accessable, pp ...Paired) AssocVecFnc {
	return conAssocVecFromPairs(append(vec.Pairs(), pp...)...)
}
func conAssocVecFromPairs(pp ...Paired) AssocVecFnc {
	return AssocVecFnc(func() []Paired { return pp })
}
func NewAssocVector(pp ...Paired) AssocVecFnc {
	return AssocVecFnc(func() (pairs []Paired) {
		for _, pair := range pp {
			pairs = append(pairs, pair)
		}
		return pairs
	})
}
func (v AssocVecFnc) Len() int                   { return len(v()) }
func (v AssocVecFnc) Empty() bool                { return ElemEmpty(v.Head()) && (len(v.Tail()) == 0) }
func (v AssocVecFnc) GetVal(praed Value) Paired  { return newPairSorter(v()...).Get(praed) }
func (v AssocVecFnc) Range(praed Value) []Paired { return newPairSorter(v()...).Range(praed) }
func (v AssocVecFnc) Search(praed Value) int     { return newPairSorter(v()...).Search(praed) }
func (v AssocVecFnc) Pairs() []Paired            { return v() }
func (v AssocVecFnc) Slice() []Value {
	var fncs = []Value{}
	for _, pair := range v() {
		fncs = append(fncs, NewPair(pair.Left(), pair.Right()))
	}
	return fncs
}
func (v AssocVecFnc) Head() Paired {
	if v.Len() > 0 {
		return v.Pairs()[0]
	}
	return nil
}
func (v AssocVecFnc) Tail() []Paired {
	if v.Len() > 1 {
		return v.Pairs()[1:]
	}
	return nil
}
func (v AssocVecFnc) Eval(p ...d.Native) d.Native {
	var slice = d.DataSlice{}
	for _, pair := range v() {
		d.SliceAppend(slice, d.NewPair(pair.Left(), pair.Right()))
	}
	return slice
}
func (v AssocVecFnc) SetVal(key, value Value) Associative {
	if idx := v.Search(key); idx >= 0 {
		var pairs = v.Pairs()
		pairs[idx] = NewPair(key, value)
		return NewAssocVector(pairs...)
	}
	return NewAssocVector(append(v.Pairs(), NewPair(key, value))...)
}
func (v AssocVecFnc) Sort(flag d.TyNative) {
	var ps = newPairSorter(v()...)
	ps.Sort(flag)
	v = NewAssocVector(ps...)
}

///////////////////////////////////////////////////
// VECTOR
// vector keeps a slice of data instances
func conVec(vec Vectorized, dd ...Value) Vectorized {
	return conVecFromValues(append(vec.Slice(), dd...)...)
}
func conVecFromValues(dd ...Value) Vectorized {
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
func (v AssocVecFnc) TypeFnc() TyFnc         { return AssocVec }
func (v VecFnc) TypeFnc() TyFnc              { return Vector }
func (v VecFnc) Ident() Value                { return v }
func (v VecFnc) Eval(p ...d.Native) d.Native { return NewVector(v()...) }
func (v VecFnc) TypeNat() d.TyNative {
	if len(v()) > 0 {
		return d.Vector.TypeNat() | v.Head().TypeNat()
	}
	return d.Vector.TypeNat() | d.Nil.TypeNat()
}

func (v AssocVecFnc) Call(d ...Value) Value {
	if len(d) > 0 {
		for _, val := range d {
			if pair, ok := val.(Paired); ok {
				v = v.Con(pair)
			}
		}
	}
	return v
}
func (v AssocVecFnc) Con(p ...Paired) AssocVecFnc {
	return v.Con(p...)
}
func (v AssocVecFnc) DeCap() (Paired, []Paired) {
	return v.Head(), v.Tail()
}
func (v AssocVecFnc) TypeNat() d.TyNative {
	if len(v()) > 0 {
		return d.Vector | v.Head().TypeNat()
	}
	return d.Vector | d.Nil.TypeNat()
}

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
			if !d.Nil.TypeNat().Flag().Match(dat.TypeNat().Flag()) {
				return false
			}
		}
	}
	return true
}
func (v VecFnc) DeCap() (Value, []Value) {
	return v.Head(), v.Tail()
}
func (v VecFnc) Vector() []Value          { return v() }
func (v VecFnc) Slice() []Value           { return v() }
func (v VecFnc) Con(arg ...Value) []Value { return append(v(), arg...) }
func (v VecFnc) Call(d ...Value) Value {
	if len(d) > 0 {
		conVec(v, d...)
	}
	return v
}
func (v VecFnc) Set(i int, val Value) Vectorized {
	if i < v.Len() {
		var slice = v()
		slice[i] = val
		return VecFnc(func() []Value { return slice })

	}
	return nil
}
func (v VecFnc) Get(i int) Value {
	if i < v.Len() {
		return v()[i]
	}
	return nil
}
func (v VecFnc) Search(praed Value) int { return newDataSorter(v()...).Search(praed) }
func (v VecFnc) Sort(flag d.TyNative) {
	var ps = newDataSorter(v()...)
	ps.Sort(flag)
	v = NewVector(ps...).(VecFnc)
}

//////////////////////////////////////////////////////////////////////////////////////////////
///// RECURSIVE LIST
////
/// base implementation of linked lists
func NewRecursiveList(dd ...Value) ListFnc {
	if len(dd) > 0 {
		if len(dd) > 1 {
			return ListFnc(func() (Value, ListFnc) {
				return dd[0], NewRecursiveList(dd[1:]...)
			})
		}
		return ListFnc(func() (Value, ListFnc) { return dd[0], nil })
	}
	return nil
}
func (l ListFnc) Ident() Value                { return l }
func (l ListFnc) Head() Value                 { h, _ := l(); return h }
func (l ListFnc) Tail() ListFnc               { _, t := l(); return t }
func (l ListFnc) TypeFnc() TyFnc              { return List }
func (l ListFnc) Eval(p ...d.Native) d.Native { return NewPair(l.Head(), l.Tail()) }
func (l ListFnc) TypeNat() d.TyNative         { return d.List.TypeNat() | l.Head().TypeNat() }
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
func (l ListFnc) Call(d ...Value) Value {
	if len(d) > 0 {
		var head, tail = l()
		return NewPair(head, tail)
	}
	return l
}
func (l ListFnc) DeCap() (Value, ListFnc) {
	var head, rec = l()
	l = ListFnc(func() (Value, ListFnc) { return l() })
	return head, rec
}
func (l ListFnc) Con(val Value) ListFnc {
	return ListFnc(func() (Value, ListFnc) { return val, l })
}
