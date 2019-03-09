/*
DATA CONSTRUCTORS

  implementations of 'precedence types', ake functional base-/ and collection types
*/
package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

type (
	// FUNCTIONAL COLLECTIONS (depend on enclosed data
	PairFnc     func(elems ...Paired) (Parametric, Parametric)
	ListFnc     func(elems ...Parametric) (Parametric, ListFnc)
	VecFnc      func(elems ...Parametric) []Parametric
	RecordFnc   func(pairs ...Paired) []Paired
	AssocSetFnc func(pairs ...Paired) d.Mapped
)

///// RECURSIVE LIST
////
/// base implementation of recursively linked lists
//
// generate empty lists out of thin air
func NewList(args ...Parametric) ListFnc {
	return conList(emptyList(), args...)
}

func emptyList() ListFnc {
	return func(elems ...Parametric) (Parametric, ListFnc) {
		if len(elems) == 0 {
			return nil, emptyList()
		}
		if len(elems) == 1 {
			return elems[0], emptyList()
		}
		return elems[0], conList(emptyList(), elems[1:]...)
	}
}

func conList(list ListFnc, initials ...Parametric) ListFnc {
	if len(initials) == 0 {
		return list
	}

	var head = initials[0]

	if len(initials) == 1 {
		return func(elems ...Parametric) (Parametric, ListFnc) {
			if len(elems) == 0 {
				return head, list
			}
			return conList(list, append(elems, head)...)()
		}
	}

	return func(elems ...Parametric) (Parametric, ListFnc) {
		// no elements → return head and list
		if len(elems) == 0 {
			return head, conList(list, initials[1:]...)
		}
		// elements got passed, append to list. to get order of passed
		// elements & head right, concat all and call resutling list,
		// to yield new head & tail list.
		return conList(list, append(elems, head)...)()
	}
}

func (l ListFnc) Ident() Parametric { return l }

func (l ListFnc) Tail() Consumeable { _, t := l(); return t }

func (l ListFnc) Head() Parametric { h, _ := l(); return h }

func (l ListFnc) TypeFnc() TyFnc { return List | Functor }

func (l ListFnc) Eval(p ...d.Native) d.Native { return NewPair(l.Head(), l.Tail()) }

func (l ListFnc) TypeNat() d.TyNative { return d.List.TypeNat() | l.Head().TypeNat() }

func (l ListFnc) Empty() bool {
	if l.Head() == nil {
		return true
	}
	return false
}

func (l ListFnc) Len() int {
	var length int
	var head, tail = l()
	if head != nil {
		length += 1 + tail.Len()
	}
	return length
}

func (l ListFnc) Call(d ...Parametric) Parametric {
	var head Parametric
	head, l = l(d...)
	return head
}

///////////////////////////////////////////////////
//// VECTOR
///
// vector is a list backed by a slice.
func conVec(vec Vectorized, fncs ...Parametric) VecFnc {
	return conVecFromFunctionals(append(vec.Slice(), fncs...)...)
}

func conVecFromFunctionals(fncs ...Parametric) VecFnc {
	return VecFnc(func(elems ...Parametric) []Parametric { return fncs })
}

func newEmptyVector() VecFnc {
	return VecFnc(func(elems ...Parametric) []Parametric {
		return []Parametric{}
	})
}

func NewVector(fncs ...Parametric) VecFnc {
	return func(elems ...Parametric) (vec []Parametric) {
		for _, dat := range fncs {
			vec = append(vec, New(dat))
		}
		return vec
	}
}

func (v VecFnc) TypeFnc() TyFnc { return Vector | Functor }

func (v VecFnc) Ident() Parametric { return v }

func (v VecFnc) Eval(p ...d.Native) d.Native { return NewVector(v()...) }

func (v VecFnc) TypeNat() d.TyNative {
	if len(v()) > 0 {
		return d.Vector.TypeNat() | v.Head().TypeNat()
	}
	return d.Vector.TypeNat() | d.Nil.TypeNat()
}

func (v VecFnc) Head() Parametric {
	if v.Len() > 0 {
		return v.Vector()[0]
	}
	return NewNoOp()
}

func (v VecFnc) Tail() Consumeable {
	if v.Len() > 1 {
		return conVecFromFunctionals(v.Vector()[1:]...)
	}
	return newEmptyVector()
}

func (v VecFnc) Empty() bool {
	if len(v()) > 0 {
		return false
	}
	return true
}

func (v VecFnc) Len() int { return len(v()) }

func (v VecFnc) DeCap() (Parametric, Consumeable) {
	var head, tail = v.Head(), v.Tail()
	if head == nil {
		head = NewNoOp()
	}
	if tail == nil {
		tail = newEmptyVector()
	}
	return head, tail
}
func (v VecFnc) Vector() []Parametric { return v() }

func (v VecFnc) Slice() []Parametric { return v() }

func (v VecFnc) Con(arg ...Parametric) []Parametric { return append(v(), arg...) }

func (v VecFnc) Call(d ...Parametric) Parametric {
	if len(d) > 0 {
		conVec(v, d...)
	}
	return v
}

func (v VecFnc) Set(i int, val Parametric) Vectorized {
	if i < v.Len() {
		var slice = v()
		slice[i] = val
		return VecFnc(func(elems ...Parametric) []Parametric { return slice })

	}
	return v
}

func (v VecFnc) Get(i int) Parametric {
	if i < v.Len() {
		return v()[i]
	}
	return NewNoOp()
}
func (v VecFnc) Search(praed Parametric) int { return newDataSorter(v()...).Search(praed) }
func (v VecFnc) Sort(flag d.TyNative) {
	var ps = newDataSorter(v()...)
	ps.Sort(flag)
	v = NewVector(ps...)
}

//// PAIR
///
//
func NewPair(l, r Parametric) PairFnc {
	return func(pairs ...Paired) (Parametric, Parametric) { return l, r }
}
func newEmptyPair() PairFnc {
	return func(pairs ...Paired) (a, b Parametric) {
		return NewNoOp(), NewNoOp()
	}
}
func NewPairFromInterface(l, r interface{}) PairFnc {
	return func(Pairs ...Paired) (Parametric, Parametric) { return New(d.New(l)), New(d.New(r)) }
}
func NewPairFromData(l, r d.Native) PairFnc {
	return func(pairs ...Paired) (Parametric, Parametric) { return New(l), New(r) }
}
func (p PairFnc) Both() (Parametric, Parametric) { return p() }

func (p PairFnc) Pair() Parametric { return p }

func (p PairFnc) Left() Parametric { l, _ := p(); return l }

func (p PairFnc) Right() Parametric { _, r := p(); return r }

func (p PairFnc) Empty() bool {
	return p.Left() == nil && p.Right() == nil
}

func (p PairFnc) Acc() Parametric { return p.Left() }

func (p PairFnc) Arg() Parametric { return p.Right() }

func (p PairFnc) AccType() d.TyNative { return p.Left().TypeNat() }

func (p PairFnc) ArgType() d.TyNative { return p.Right().TypeNat() }

func (p PairFnc) Ident() Parametric { return p }

func (p PairFnc) Call(...Parametric) Parametric { return p }

func (p PairFnc) Eval(a ...d.Native) d.Native { return d.NewPair(p.Left().Eval(), p.Right().Eval()) }

func (p PairFnc) TypeFnc() TyFnc { return Pair | Function }

func (p PairFnc) TypeNat() d.TyNative {
	return d.Pair.TypeNat() | p.Left().TypeNat() | p.Right().TypeNat()
}

//// RECORD
///
//
func (v RecordFnc) Call(d ...Parametric) Parametric {
	if len(d) > 0 {
		for _, val := range d {
			if pair, ok := val.(Paired); ok {
				v = v.Con(pair)
			}
		}
	}
	return v
}

func (v RecordFnc) Con(p ...Paired) RecordFnc {
	return v.Con(p...)
}

func (v RecordFnc) DeCap() (Parametric, Consumeable) {
	return v.Head(), v.Tail()
}

func (v RecordFnc) AccFncType() TyFnc {
	if v.Len() > 0 {
		return v.Pairs()[0].Left().TypeFnc()
	}
	return None
}

func (v RecordFnc) Empty() bool {
	if len(v()) > 0 {
		for _, pair := range v() {
			if !pair.(PairFnc).Empty() {
				return false
			}
		}
	}
	return true
}

func (v RecordFnc) AccNatType() d.TyNative {
	if v.Len() > 0 {
		return v.Pairs()[0].Left().TypeNat()
	}
	return d.Nil
}

func (v RecordFnc) ArgFncType() TyFnc {
	if v.Len() > 0 {
		return v.Pairs()[0].Right().TypeFnc()
	}
	return None
}

func (v RecordFnc) ArgNatType() d.TyNative {
	if v.Len() > 0 {
		return v.Pairs()[0].Right().TypeNat()
	}
	return d.Nil
}

func (v RecordFnc) TypeFnc() TyFnc { return Record | Accessor | Functor }

func (v RecordFnc) TypeNat() d.TyNative {
	if len(v()) > 0 {
		return d.Vector | v.Head().TypeNat()
	}
	return d.Vector | d.Nil.TypeNat()
}

///////////////////////////////////////////////////////////////////////
//// ASSOCIATIVE VECTOR (VECTOR OF PAIRS)
///
// associative array that uses pairs left field as accessor for sort & search
func conRecord(vec Associative, pp ...Paired) RecordFnc {
	return conRecordFromPairs(append(vec.Pairs(), pp...)...)
}

func newRecord(ps ...PairFnc) RecordFnc {
	var pairs = []Paired{}
	for _, pair := range ps {
		pairs = append(pairs, pair)
	}
	return RecordFnc(func(pairs ...Paired) []Paired { return pairs })
}

func conRecordFromPairs(pp ...Paired) RecordFnc {
	return RecordFnc(func(pairs ...Paired) []Paired { return pp })
}

func newEmptyRecord() RecordFnc {
	return RecordFnc(func(pairs ...Paired) []Paired { return []Paired{} })
}

func NewRecord(pp ...Paired) RecordFnc {
	return func(pairs ...Paired) []Paired {
		for _, pair := range pp {
			pairs = append(pairs, pair)
		}
		return pairs
	}
}

func (v RecordFnc) Len() int { return len(v()) }

func (v RecordFnc) Get(idx int) Paired {
	if idx < v.Len()-1 {
		return v()[idx]
	}
	return NewPair(NewNoOp(), NewNoOp())
}

func (v RecordFnc) GetVal(praed Parametric) Paired {
	return newPairSorter(v()...).Get(praed)
}

func (v RecordFnc) Range(praed Parametric) []Paired {
	return newPairSorter(v()...).Range(praed)
}

func (v RecordFnc) Search(praed Parametric) int {
	return newPairSorter(v()...).Search(praed)
}

func (v RecordFnc) Pairs() []Paired {
	return v()
}

func (v RecordFnc) SetVal(key, value Parametric) Associative {
	if idx := v.Search(key); idx >= 0 {
		var pairs = v.Pairs()
		pairs[idx] = NewPair(key, value)
		return NewRecord(pairs...)
	}
	return NewRecord(append(v.Pairs(), NewPair(key, value))...)
}

func (v RecordFnc) Slice() []Parametric {
	var fncs = []Parametric{}
	for _, pair := range v() {
		fncs = append(fncs, NewPair(pair.Left(), pair.Right()))
	}
	return fncs
}

func (v RecordFnc) Head() Parametric {
	if v.Len() > 0 {
		return v.Pairs()[0]
	}
	return NewNoOp()
}

func (v RecordFnc) Tail() Consumeable {
	if v.Len() > 1 {
		return conRecordFromPairs(v.Pairs()[1:]...)
	}
	return newEmptyRecord()
}

func (v RecordFnc) Eval(p ...d.Native) d.Native {
	var slice = d.DataSlice{}
	for _, pair := range v() {
		d.SliceAppend(slice, d.NewPair(pair.Left(), pair.Right()))
	}
	return slice
}

func (v RecordFnc) Sort(flag d.TyNative) {
	var ps = newPairSorter(v()...)
	ps.Sort(flag)
	v = NewRecord(ps...)
}

func (v RecordFnc) MapF(fnc Parametric) Consumeable {
	return v
}

///////////////////////////////////////////////////////////////////////
//// ASSOCIATIVE MAP (HASH MAP OF VALUES)
///
// associative array that uses pairs left field as accessor for sort & search
func conAssocSet(pairs ...Paired) AssocSetFnc {
	var paired = []PairFnc{}
	for _, pair := range pairs {
		paired = append(paired, pair.(PairFnc))
	}
	return NewAssocSet(paired...)
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
	return AssocSetFnc(func(pairs ...Paired) d.Mapped { return set })
}

func (v AssocSetFnc) Split() (VecFnc, VecFnc) {
	var keys, vals = []Parametric{}, []Parametric{}
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

func (v AssocSetFnc) Len() int { return v().Len() }

func (v AssocSetFnc) Empty() bool {
	for _, pair := range v.Pairs() {
		if !pair.(PairFnc).Empty() {
			return false
		}
	}
	return true
}

func (v AssocSetFnc) GetVal(praed Parametric) Paired {
	var val Parametric
	var nat, ok = v().Get(praed)
	if val, ok = nat.(Parametric); !ok {
		val = NewFromData(val)
	}
	return NewPair(praed, val)
}

func (v AssocSetFnc) SetVal(key, value Parametric) Associative {
	var m = v()
	m.Set(key, value)
	return AssocSetFnc(func(pairs ...Paired) d.Mapped { return m })
}

func (v AssocSetFnc) Slice() []Parametric {
	var pairs = []Parametric{}
	for _, pair := range v.Pairs() {
		pairs = append(pairs, pair)
	}
	return pairs
}

func (v AssocSetFnc) Call(f ...Parametric) Parametric { return v }

func (v AssocSetFnc) Eval(p ...d.Native) d.Native {
	var slice = d.DataSlice{}
	for _, pair := range v().Fields() {
		d.SliceAppend(slice, d.NewPair(pair.Left(), pair.Right()))
	}
	return slice
}

func (v AssocSetFnc) TypeFnc() TyFnc { return MuliSet | Accessor | Functor }

func (v AssocSetFnc) TypeNat() d.TyNative { return d.Set | d.Function }

func (v AssocSetFnc) AccFncType() TyFnc {
	if v.Len() > 0 {
		return v.Pairs()[0].Left().TypeFnc()
	}
	return None
}

func (v AssocSetFnc) AccNatType() d.TyNative {
	if v.Len() > 0 {
		return v.Pairs()[0].Left().TypeNat()
	}
	return d.Nil
}

func (v AssocSetFnc) ArgFncType() TyFnc {
	if v.Len() > 0 {
		return v.Pairs()[0].Right().TypeFnc()
	}
	return None
}

func (v AssocSetFnc) ArgNatType() d.TyNative {
	if v.Len() > 0 {
		return v.Pairs()[0].Right().TypeNat()
	}
	return d.Nil
}

func (v AssocSetFnc) DeCap() (Parametric, Consumeable) {
	return v.Head(), v.Tail()
}

func (v AssocSetFnc) Head() Parametric {
	if v.Len() > 0 {
		return v.Pairs()[0]
	}
	return NewNoOp()
}

func (v AssocSetFnc) Tail() Consumeable {
	if v.Len() > 1 {
		return conRecordFromPairs(v.Pairs()[1:]...)
	}
	return newEmptyRecord()
}
