/*
DATA CONSTRUCTORS

  implementations of 'precedence types', ake functional base-/ and collection types
*/
package functions

import (
	"strings"

	d "github.com/JoergReinhardt/gatwd/data"
	"github.com/olekukonko/tablewriter"
)

type (
	// FIXED ARITY FUNCTIONS
	ConstFnc  func() Functional
	UnaryFnc  func(Functional) Functional
	BinaryFnc func(a, b Functional) Functional
	NaryFnc   func(...Functional) Functional

	// CLOSURES OVER COLLECTIONS
	ListFnc     func() (Functional, ListFnc)
	PairFnc     func() (a, b Functional)
	VecFnc      func() []Functional
	AssocVecFnc func() []Paired
	AssocSetFnc func() d.Mapped
)

func ElemEmpty(val Functional) bool {
	if val != nil {
		if !val.TypeFnc().Flag().Match(None) {
			return false
		}
	}
	return true
}

// CONSTANT
//
// constant also conains immutable data that may be an instance of a type of
// the data package, or result of a function call guarantueed to allways return
// the same value.
func NewConstant(fnc func() Functional) ConstFnc {
	return ConstFnc(func() Functional { return fnc() })
}

func (c ConstFnc) Ident() Functional               { return c() }
func (c ConstFnc) TypeFnc() TyFnc                  { return Function }
func (c ConstFnc) TypeNat() d.TyNative             { return c().TypeNat() }
func (c ConstFnc) Eval(p ...d.Native) d.Native     { return c().Eval() }
func (c ConstFnc) Call(d ...Functional) Functional { return c() }

///// UNARY FUNCTION
func NewUnaryFnc(fnc func(f Functional) Functional) UnaryFnc {
	return UnaryFnc(func(f Functional) Functional { return fnc(f) })
}
func (u UnaryFnc) TypeNat() d.TyNative         { return d.Function.TypeNat() }
func (u UnaryFnc) TypeFnc() TyFnc              { return Function }
func (u UnaryFnc) Ident() Functional           { return u }
func (u UnaryFnc) Eval(p ...d.Native) d.Native { return u }
func (u UnaryFnc) Call(d ...Functional) Functional {
	return u(d[0])
}

///// BINARY FUNCTION
func NewBinaryFnc(fnc func(a, b Functional) Functional) BinaryFnc {
	return BinaryFnc(func(a, b Functional) Functional { return fnc(a, b) })
}
func (b BinaryFnc) TypeNat() d.TyNative             { return d.Function.TypeNat() }
func (b BinaryFnc) TypeFnc() TyFnc                  { return Function }
func (b BinaryFnc) Ident() Functional               { return b }
func (b BinaryFnc) Eval(p ...d.Native) d.Native     { return b }
func (b BinaryFnc) Call(d ...Functional) Functional { return b(d[0], d[1]) }

///// NARY FUNCTION
func NewNaryFnc(fnc func(f ...Functional) Functional) NaryFnc {
	return NaryFnc(func(f ...Functional) Functional { return fnc(f...) })
}
func (n NaryFnc) TypeNat() d.TyNative             { return d.Function.TypeNat() }
func (n NaryFnc) TypeFnc() TyFnc                  { return Function }
func (n NaryFnc) Ident() Functional               { return n }
func (n NaryFnc) Eval(p ...d.Native) d.Native     { return n }
func (n NaryFnc) Call(d ...Functional) Functional { return n(d...) }

//////////////////////////////////////////////////////////////////////////////////////////////
///// RECURSIVE LIST
////
/// base implementation of recursively linked lists
func NewRecursiveList(dd ...Functional) ListFnc {
	if len(dd) > 0 {
		if len(dd) > 1 {
			return ListFnc(func() (Functional, ListFnc) {
				return dd[0], NewRecursiveList(dd[1:]...)
			})
		}
		return ListFnc(func() (Functional, ListFnc) { return dd[0], nil })
	}
	return nil
}
func (l ListFnc) Ident() Functional           { return l }
func (l ListFnc) Head() Functional            { h, _ := l(); return h }
func (l ListFnc) Tail() Consumeable           { _, t := l(); return t }
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
func (l ListFnc) Call(d ...Functional) Functional {
	if len(d) > 0 {
		var head, tail = l()
		return NewPair(head, tail)
	}
	return l
}
func (l ListFnc) DeCap() (Functional, Consumeable) {
	var head, rec = l()
	l = ListFnc(func() (Functional, ListFnc) { return l() })
	return head, rec
}
func (l ListFnc) Con(val Functional) ListFnc {
	return ListFnc(func() (Functional, ListFnc) { return val, l })
}

///////////////////////////////////////////////////
//// VECTOR
///
// vector is a list backed by a slice.
func conVec(vec Vectorized, dd ...Functional) VecFnc {
	return conVecFromFunctionals(append(vec.Slice(), dd...)...)
}
func conVecFromFunctionals(dd ...Functional) VecFnc {
	return VecFnc(func() []Functional { return dd })
}
func NewVector(dd ...Functional) VecFnc {
	return VecFnc(func() (vec []Functional) {
		for _, dat := range dd {
			vec = append(vec, New(dat))
		}
		return vec
	})
}
func (v AssocVecFnc) TypeFnc() TyFnc         { return AssocVec }
func (v VecFnc) TypeFnc() TyFnc              { return Vector }
func (v VecFnc) Ident() Functional           { return v }
func (v VecFnc) Eval(p ...d.Native) d.Native { return NewVector(v()...) }
func (v VecFnc) TypeNat() d.TyNative {
	if len(v()) > 0 {
		return d.Vector.TypeNat() | v.Head().TypeNat()
	}
	return d.Vector.TypeNat() | d.Nil.TypeNat()
}

func (v AssocVecFnc) Call(d ...Functional) Functional {
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
func (v AssocVecFnc) DeCap() (Functional, Consumeable) {
	return v.Head(), v.Tail()
}
func (v AssocVecFnc) TypeNat() d.TyNative {
	if len(v()) > 0 {
		return d.Vector | v.Head().TypeNat()
	}
	return d.Vector | d.Nil.TypeNat()
}

// base implementation functions/sliceable interface
func (v VecFnc) Head() Functional {
	if v.Len() > 0 {
		return v.Vector()[0]
	}
	return nil
}
func (v VecFnc) Tail() Consumeable {
	if v.Len() > 1 {
		return conVecFromFunctionals(v.Vector()[1:]...)
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
func (v VecFnc) DeCap() (Functional, Consumeable) {
	return v.Head(), v.Tail()
}
func (v VecFnc) Vector() []Functional               { return v() }
func (v VecFnc) Slice() []Functional                { return v() }
func (v VecFnc) Con(arg ...Functional) []Functional { return append(v(), arg...) }
func (v VecFnc) Call(d ...Functional) Functional {
	if len(d) > 0 {
		conVec(v, d...)
	}
	return v
}
func (v VecFnc) Set(i int, val Functional) Vectorized {
	if i < v.Len() {
		var slice = v()
		slice[i] = val
		return VecFnc(func() []Functional { return slice })

	}
	return nil
}
func (v VecFnc) Get(i int) Functional {
	if i < v.Len() {
		return v()[i]
	}
	return nil
}
func (v VecFnc) Search(praed Functional) int { return newDataSorter(v()...).Search(praed) }
func (v VecFnc) Sort(flag d.TyNative) {
	var ps = newDataSorter(v()...)
	ps.Sort(flag)
	v = NewVector(ps...)
}

/////////////////////////////////////////////////////////
// PAIR
func NewPair(l, r Functional) PairFnc {
	return PairFnc(func() (Functional, Functional) { return l, r })
}
func NewPairFromInterface(l, r interface{}) PairFnc {
	return PairFnc(func() (Functional, Functional) { return New(d.New(l)), New(d.New(r)) })
}
func NewPairFromData(l, r d.Native) PairFnc {
	return PairFnc(func() (Functional, Functional) { return New(l), New(r) })
}
func (p PairFnc) Both() (Functional, Functional) { return p() }
func (p PairFnc) Pair() Functional               { return p }
func (p PairFnc) Left() Functional               { l, _ := p(); return l }
func (p PairFnc) Right() Functional              { _, r := p(); return r }
func (p PairFnc) Acc() Functional                { return p.Left() }
func (p PairFnc) Arg() Functional                { return p.Right() }
func (p PairFnc) AccType() d.TyNative            { return p.Left().TypeNat() }
func (p PairFnc) ArgType() d.TyNative            { return p.Right().TypeNat() }
func (p PairFnc) Ident() Functional              { return p }
func (p PairFnc) Call(...Functional) Functional  { return p }
func (p PairFnc) Eval(a ...d.Native) d.Native    { return d.NewPair(p.Left().Eval(), p.Right().Eval()) }
func (p PairFnc) TypeFnc() TyFnc                 { return Pair | Function }
func (p PairFnc) TypeNat() d.TyNative {
	return d.Pair.TypeNat() | p.Left().TypeNat() | p.Right().TypeNat()
}
func (p PairFnc) Empty() bool {
	return ElemEmpty(p.Left()) && ElemEmpty(p.Right())
}

///////////////////////////////////////////////////////////////////////
//// ASSOCIATIVE VECTOR (VECTOR OF PAIRS)
///
// associative array that uses pairs left field as accessor for sort & search
func conAssocVec(vec Associative, pp ...Paired) AssocVecFnc {
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
func (v AssocVecFnc) Len() int                        { return len(v()) }
func (v AssocVecFnc) Empty() bool                     { return ElemEmpty(v.Head()) && (len(v()) == 0) }
func (v AssocVecFnc) GetVal(praed Functional) Paired  { return newPairSorter(v()...).Get(praed) }
func (v AssocVecFnc) Range(praed Functional) []Paired { return newPairSorter(v()...).Range(praed) }
func (v AssocVecFnc) Search(praed Functional) int     { return newPairSorter(v()...).Search(praed) }
func (v AssocVecFnc) Pairs() []Paired                 { return v() }
func (v AssocVecFnc) SetVal(key, value Functional) Associative {
	if idx := v.Search(key); idx >= 0 {
		var pairs = v.Pairs()
		pairs[idx] = NewPair(key, value)
		return NewAssocVector(pairs...)
	}
	return NewAssocVector(append(v.Pairs(), NewPair(key, value))...)
}
func (v AssocVecFnc) Slice() []Functional {
	var fncs = []Functional{}
	for _, pair := range v() {
		fncs = append(fncs, NewPair(pair.Left(), pair.Right()))
	}
	return fncs
}
func (v AssocVecFnc) Head() Functional {
	if v.Len() > 0 {
		return v.Pairs()[0]
	}
	return nil
}
func (v AssocVecFnc) Tail() Consumeable {
	if v.Len() > 1 {
		return conAssocVecFromPairs(v.Pairs()[1:]...)
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
func (v AssocVecFnc) Sort(flag d.TyNative) {
	var ps = newPairSorter(v()...)
	ps.Sort(flag)
	v = NewAssocVector(ps...)
}

///////////////////////////////////////////////////////////////////////
//// ASSOCIATIVE MAP (HASH MAP OF VALUES)
///
// associative array that uses pairs left field as accessor for sort & search
func conAssocSetFromPairs(pairs ...PairFnc) {
}
func NewAssocSet(pairs ...PairFnc) AssocSetFnc {

	var kt d.TyNative
	var set d.Mapped

	// OR concat all accessor types
	for _, pair := range pairs {
		kt = kt | pair.AccType()
	}
	// if accessors are of mixed type‥.
	if kt.Flag().Count() > 1 {
		set = d.SetVal{}
	} else {
		var ktf = kt.Flag()
		switch {
		case ktf.Match(d.Int):
			set = d.SetInt{}
		case ktf.Match(d.Uint):
			set = d.SetUint{}
		case ktf.Match(d.Flag):
			set = d.SetFlag{}
		case ktf.Match(d.Float):
			set = d.SetFloat{}
		case ktf.Match(d.String):
			set = d.SetString{}
		}
	}
	return AssocSetFnc(func() d.Mapped { return set })
}
func (v AssocSetFnc) Split() (VecFnc, VecFnc) {
	var keys, vals = []Functional{}, []Functional{}
	for _, pair := range v.Pairs() {
		keys = append(keys, pair.Left())
		vals = append(vals, pair.Right())
	}
	return NewVector(keys...), NewVector(vals...)
}
func (v AssocSetFnc) Pairs() []Paired {
	var pairs = []Paired{}
	for _, field := range v().Fields() {
		pairs = append(
			pairs,
			NewPairFromData(
				field.Left(),
				field.Right()))
	}
	return pairs
}
func (v AssocSetFnc) Keys() VecFnc { k, _ := v.Split(); return k }
func (v AssocSetFnc) Data() VecFnc { _, d := v.Split(); return d }
func (v AssocSetFnc) Len() int     { return v().Len() }
func (v AssocSetFnc) GetVal(praed Functional) Paired {
	var val Functional
	var nat, ok = v().Get(praed)
	if val, ok = nat.(Functional); !ok {
		val = NewFromData(val)
	}
	return NewPair(praed, val)
}
func (v AssocSetFnc) SetVal(key, value Functional) Associative {
	var m = v()
	m.Set(key, value)
	return AssocSetFnc(func() d.Mapped { return m })
}
func (v AssocSetFnc) Empty() bool {
	if v.Len() > 0 {
		for _, pair := range v.Pairs() {
			if !ElemEmpty(pair.Left()) || !ElemEmpty(pair.Right()) {
				return false
			}
		}
	}
	return true
}
func (v AssocSetFnc) Slice() []Functional {
	var pairs = []Functional{}
	for _, pair := range v.Pairs() {
		pairs = append(pairs, pair)
	}
	return pairs
}
func (v AssocSetFnc) Call(f ...Functional) Functional { return v }
func (v AssocSetFnc) Eval(p ...d.Native) d.Native {
	var slice = d.DataSlice{}
	for _, pair := range v().Fields() {
		d.SliceAppend(slice, d.NewPair(pair.Left(), pair.Right()))
	}
	return slice
}
func (v AssocSetFnc) String() string {
	var strb = &strings.Builder{}
	var tab = tablewriter.NewWriter(strb)

	for _, pair := range v.Pairs() {
		var row = []string{pair.Left().String(), pair.Right().String()}
		tab.Append(row)
	}
	tab.Render()
	return strb.String()
}
func (v AssocSetFnc) TypeFnc() TyFnc      { return MuliSet | Accessor }
func (v AssocSetFnc) TypeNat() d.TyNative { return d.Set | d.Function }
