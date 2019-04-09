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
	PairFnc   func(elems ...Callable) (Callable, Callable)
	ListFnc   func(elems ...Callable) (Callable, ListFnc)
	VecFnc    func(elems ...Callable) []Callable
	RecordFnc func(pairs ...Applicable) []Applicable
	SetFnc    func(pairs ...Applicable) d.Mapped
)

//// PAIR
///
//
func NewPair(l, r Callable) PairFnc {
	return func(pairs ...Callable) (Callable, Callable) {
		return l, r
	}
}
func NewEmptyPair() PairFnc {
	return func(pairs ...Callable) (a, b Callable) {
		return NewNoOp(), NewNoOp()
	}
}
func NewPairFromInterface(l, r interface{}) PairFnc {
	return func(Pairs ...Callable) (Callable, Callable) {
		return New(d.New(l)), New(d.New(r))
	}
}
func NewPairFromData(l, r d.Native) PairFnc {
	return func(pairs ...Callable) (Callable, Callable) {
		return New(l), New(r)
	}
}
func (p PairFnc) Both() (Callable, Callable) {
	return p()
}

func (p PairFnc) DeCap() (Callable, Consumeable) {
	l, r := p()
	return l, NewList(r)
}

func (p PairFnc) Apply(args ...Callable) (Callable, ApplicapleFnc) {
	var head, tail = p.DeCap()
	var appl = NewApplicaple(tail)
	if head != nil {
		if len(args) > 0 {
			return head.Call(args...), appl
		}
		return head.(Applicable), appl
	}
	return nil, appl
}

func (p PairFnc) Fold(fold BinaryFnc, ilem Callable) Callable {
	return fold(ilem, p)
}

func (p PairFnc) MapF(fmap UnaryFnc) FunctorFnc {
	return NewFunctor(fmap(p).(PairFnc))
}

func (p PairFnc) Pair() Callable { return p }

func (p PairFnc) Head() Callable { l, _ := p(); return l }

func (p PairFnc) Tail() Consumeable { return p.Tail() }

func (p PairFnc) Left() Callable { l, _ := p(); return l }

func (p PairFnc) Right() Callable { _, r := p(); return r }

func (p PairFnc) Empty() bool {
	return p.Left() == nil && p.Right() == nil
}

func (p PairFnc) Acc() Callable { return p.Left() }

func (p PairFnc) Arg() Callable { return p.Right() }

func (p PairFnc) AccType() d.TyNative { return p.Left().TypeNat() }

func (p PairFnc) ArgType() d.TyNative { return p.Right().TypeNat() }

func (p PairFnc) Ident() Callable { return p }

func (p PairFnc) Call(...Callable) Callable { return p }

func (p PairFnc) Eval(a ...d.Native) d.Native { return d.NewPair(p.Left().Eval(), p.Right().Eval()) }

func (p PairFnc) TypeFnc() TyFnc { return Pair }

func (p PairFnc) TypeNat() d.TyNative {
	return d.Pair.TypeNat() | p.Left().TypeNat() | p.Right().TypeNat()
}

///// RECURSIVE LIST
////
/// base implementation of recursively linked lists
//
// generate empty lists out of thin air
func NewList(args ...Callable) ListFnc {
	return ConList(EmptyList(), args...)
}

func EmptyList() ListFnc {
	return func(elems ...Callable) (Callable, ListFnc) {
		if len(elems) == 0 {
			return nil, EmptyList()
		}
		if len(elems) == 1 {
			return elems[0], EmptyList()
		}
		return elems[0], ConList(EmptyList(), elems[1:]...)
	}
}

// concat elements to list step wise
func ConList(list ListFnc, initials ...Callable) ListFnc {
	// return empty list if no parameter got passed
	if len(initials) == 0 {
		return list
	}

	// allocate head from first element passed
	var head = initials[0]

	// if only head element has been passed
	if len(initials) == 1 {
		// return a function, that returns‥.
		return func(elems ...Callable) (Callable, ListFnc) {
			// either head element and the initial list (which
			// would be a list with the head element as it's only
			// element)
			if len(elems) == 0 {
				return head, list
			}
			// or return the initial list followed by the elements
			// passed to the inner function, followed by the
			// initial head
			return ConList(list, append(elems, head)...)()
		}
	}

	// if more elements have been passed, lazy concat them with the initial list
	return func(elems ...Callable) (Callable, ListFnc) {
		// no elements → return head and list
		if len(elems) == 0 {
			return head, ConList(list, initials[1:]...)
		}
		// elements got passed, append to list. to get order of passed
		// elements & head right, concat all and call resutling list,
		// to yield new head & tail list.
		return ConList(list, append(elems, initials...)...)()
	}
}

func (l ListFnc) Ident() Callable { return l }

func (l ListFnc) Tail() Consumeable { _, t := l(); return t }

func (l ListFnc) Head() Callable { h, _ := l(); return h }

func (l ListFnc) DeCap() (Callable, Consumeable) { return l() }

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

func (l ListFnc) Call(d ...Callable) Callable {
	var head Callable
	head, l = l(d...)
	return head
}

/// LIST FUNCTIONS
//
// REVERSE LIST
func ReverseList(lfn ListFnc) ListFnc {
	var result = EmptyList()
	var head Callable
	head, lfn = lfn()
	for head != nil {
		result = ConList(result, head)
		head, lfn = lfn()
	}
	return result
}

//// VECTOR
///
// vector is a list backed by a slice.
func ConVector(vec Vectorized, fncs ...Callable) VecFnc {
	return ConVecFromCallable(append(vec.Slice(), fncs...)...)
}

func ConVecFromCallable(fncs ...Callable) VecFnc {
	return VecFnc(func(elems ...Callable) []Callable { return fncs })
}

func NewEmptyVector() VecFnc {
	return VecFnc(func(elems ...Callable) []Callable {
		return []Callable{}
	})
}

func NewVector(fncs ...Callable) VecFnc {
	return func(elems ...Callable) (vec []Callable) {
		for _, dat := range fncs {
			vec = append(vec, New(dat))
		}
		return vec
	}
}

func (v VecFnc) TypeFnc() TyFnc { return Vector | Functor }

func (v VecFnc) Ident() Callable { return v }

func (v VecFnc) Eval(p ...d.Native) d.Native { return NewVector(v()...) }

func (v VecFnc) TypeNat() d.TyNative {
	if len(v()) > 0 {
		return d.Vector.TypeNat() | v.Head().TypeNat()
	}
	return d.Vector.TypeNat() | d.Nil.TypeNat()
}

func (v VecFnc) Head() Callable {
	if v.Len() > 0 {
		return v.Vector()[0]
	}
	return nil
}

func (v VecFnc) Tail() Consumeable {
	if v.Len() > 1 {
		return ConVecFromCallable(v.Vector()[1:]...)
	}
	return NewEmptyVector()
}

func (v VecFnc) Empty() bool {
	if len(v()) > 0 {
		return false
	}
	return true
}

func (v VecFnc) Len() int { return len(v()) }

func (v VecFnc) DeCap() (Callable, Consumeable) {
	return v.Head(), v.Tail()
}
func (v VecFnc) Vector() []Callable { return v() }

func (v VecFnc) Slice() []Callable { return v() }

func (v VecFnc) Con(arg ...Callable) []Callable { return append(v(), arg...) }

func (v VecFnc) Call(d ...Callable) Callable {
	if len(d) > 0 {
		ConVector(v, d...)
	}
	return v
}

func (v VecFnc) Set(i int, val Callable) Vectorized {
	if i < v.Len() {
		var slice = v()
		slice[i] = val
		return VecFnc(func(elems ...Callable) []Callable { return slice })

	}
	return v
}

func (v VecFnc) Get(i int) Callable {
	if i < v.Len() {
		return v()[i]
	}
	return NewNoOp()
}
func (v VecFnc) Search(praed Callable) int { return newDataSorter(v()...).Search(praed) }
func (v VecFnc) Sort(flag d.TyNative) {
	var ps = newDataSorter(v()...)
	ps.Sort(flag)
	v = NewVector(ps...)
}

//// RECORD
///
//
func (v RecordFnc) Call(d ...Callable) Callable {
	if len(d) > 0 {
		for _, val := range d {
			if pair, ok := val.(Applicable); ok {
				v = v.Con(pair)
			}
		}
	}
	return v
}

func (v RecordFnc) Con(p ...Applicable) RecordFnc {
	return v.Con(p...)
}

func (v RecordFnc) DeCap() (Callable, Consumeable) {
	return v.Head(), v.Tail()
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

// extract signature of record type, including index position, native &
// functional type of each element.
type RecType func() (
	pos int,
	tnat d.TyNative,
	tfnc TyFnc,
)

func NewRecType(
	pos int,
	tnat d.TyNative,
	tfnc TyFnc,
) RecType {
	return func() (int, d.TyNative, TyFnc) {
		return pos, tnat, tfnc
	}
}

func (v RecordFnc) RecordType() []RecType {
	var rtype = []RecType{}
	for pos, rec := range v() {
		rtype = append(
			rtype,
			NewRecType(
				pos,
				rec.TypeNat(),
				rec.TypeFnc(),
			),
		)
	}
	return rtype
}

func (v RecordFnc) KeyFncType() TyFnc {
	if v.Len() > 0 {
		return v.Pairs()[0].Left().TypeFnc()
	}
	return None
}

func (v RecordFnc) KeyNatType() d.TyNative {
	if v.Len() > 0 {
		return v.Pairs()[0].Left().TypeNat()
	}
	return d.Nil
}

func (v RecordFnc) ValFncType() TyFnc {
	if v.Len() > 0 {
		return v.Pairs()[0].Right().TypeFnc()
	}
	return None
}

func (v RecordFnc) ValNatType() d.TyNative {
	if v.Len() > 0 {
		return v.Pairs()[0].Right().TypeNat()
	}
	return d.Nil
}

func (v RecordFnc) TypeFnc() TyFnc { return Record | Functor }

func (v RecordFnc) TypeNat() d.TyNative {
	if len(v()) > 0 {
		return d.Vector | v.Head().TypeNat()
	}
	return d.Vector | d.Nil.TypeNat()
}
func ConRecord(vec Associative, pp ...Applicable) RecordFnc {
	return ConRecordFromPairs(append(vec.Pairs(), pp...)...)
}

func NewRecordFromPairFunction(ps ...PairFnc) RecordFnc {
	var pairs = []Applicable{}
	for _, pair := range ps {
		pairs = append(pairs, pair)
	}
	return RecordFnc(func(pairs ...Applicable) []Applicable { return pairs })
}

func ConRecordFromPairs(pp ...Applicable) RecordFnc {
	return RecordFnc(func(pairs ...Applicable) []Applicable { return pp })
}

func NewEmptyRecord() RecordFnc {
	return RecordFnc(func(pairs ...Applicable) []Applicable { return []Applicable{} })
}

func NewRecord(pp ...Applicable) RecordFnc {
	return func(pairs ...Applicable) []Applicable {
		for _, pair := range pp {
			pairs = append(pairs, pair)
		}
		return pairs
	}
}

func (v RecordFnc) Len() int { return len(v()) }

func (v RecordFnc) Get(idx int) Applicable {
	if idx < v.Len()-1 {
		return v()[idx]
	}
	return NewPair(NewNoOp(), NewNoOp())
}

func (v RecordFnc) GetVal(praed Callable) Applicable {
	return newPairSorter(v()...).Get(praed)
}

func (v RecordFnc) Range(praed Callable) []Applicable {
	return newPairSorter(v()...).Range(praed)
}

func (v RecordFnc) Search(praed Callable) int {
	return newPairSorter(v()...).Search(praed)
}

func (v RecordFnc) Pairs() []Applicable {
	return v()
}

func (v RecordFnc) SwitchedPairs() []Applicable {
	var switched = []Applicable{}
	for _, pair := range v() {
		switched = append(
			switched,
			NewPair(
				pair.Right(),
				pair.Left()))
	}
	return switched
}

func (v RecordFnc) SetVal(key, value Callable) Associative {
	if idx := v.Search(key); idx >= 0 {
		var pairs = v.Pairs()
		pairs[idx] = NewPair(key, value)
		return NewRecord(pairs...)
	}
	return NewRecord(append(v.Pairs(), NewPair(key, value))...)
}

func (v RecordFnc) Slice() []Callable {
	var fncs = []Callable{}
	for _, pair := range v() {
		fncs = append(fncs, NewPair(pair.Left(), pair.Right()))
	}
	return fncs
}

func (v RecordFnc) Head() Callable {
	if v.Len() > 0 {
		return v.Pairs()[0]
	}
	return nil
}

func (v RecordFnc) Tail() Consumeable {
	if v.Len() > 1 {
		return ConRecordFromPairs(v.Pairs()[1:]...)
	}
	return NewEmptyRecord()
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

func (v RecordFnc) MapRecord(fnc Callable) Consumeable {
	return v
}

///////////////////////////////////////////////////////////////////////
//// ASSOCIATIVE SET (HASH MAP OF VALUES)
///
// associative array that uses pairs left field as accessor for sort & search
func ConAssocSet(pairs ...Applicable) SetFnc {
	var paired = []PairFnc{}
	for _, pair := range pairs {
		paired = append(paired, pair.(PairFnc))
	}
	return NewAssocSet(paired...)
}

func NewAssocSet(pairs ...PairFnc) SetFnc {

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
	return SetFnc(func(pairs ...Applicable) d.Mapped { return set })
}

func (v SetFnc) Split() (VecFnc, VecFnc) {
	var keys, vals = []Callable{}, []Callable{}
	for _, pair := range v.Pairs() {
		keys = append(keys, pair.Left())
		vals = append(vals, pair.Right())
	}
	return NewVector(keys...), NewVector(vals...)
}

func (v SetFnc) Pairs() []Applicable {
	var pairs = []Applicable{}
	for _, field := range v().Fields() {
		pairs = append(
			pairs,
			NewPairFromData(
				field.Left(),
				field.Right()))
	}
	return pairs
}

func (v SetFnc) Keys() VecFnc { k, _ := v.Split(); return k }

func (v SetFnc) Data() VecFnc { _, d := v.Split(); return d }

func (v SetFnc) Len() int { return v().Len() }

func (v SetFnc) Empty() bool {
	for _, pair := range v.Pairs() {
		if !pair.(PairFnc).Empty() {
			return false
		}
	}
	return true
}

func (v SetFnc) GetVal(praed Callable) Applicable {
	var val Callable
	var nat, ok = v().Get(praed)
	if val, ok = nat.(Callable); !ok {
		val = NewFromData(val)
	}
	return NewPair(praed, val)
}

func (v SetFnc) SetVal(key, value Callable) Associative {
	var m = v()
	m.Set(key, value)
	return SetFnc(func(pairs ...Applicable) d.Mapped { return m })
}

func (v SetFnc) Slice() []Callable {
	var pairs = []Callable{}
	for _, pair := range v.Pairs() {
		pairs = append(pairs, pair)
	}
	return pairs
}

func (v SetFnc) Call(f ...Callable) Callable { return v }

func (v SetFnc) Eval(p ...d.Native) d.Native {
	var slice = d.DataSlice{}
	for _, pair := range v().Fields() {
		d.SliceAppend(slice, d.NewPair(pair.Left(), pair.Right()))
	}
	return slice
}

func (v SetFnc) TypeFnc() TyFnc { return Set | Functor }

func (v SetFnc) TypeNat() d.TyNative { return d.Set | d.Function }

func (v SetFnc) KeyFncType() TyFnc {
	if v.Len() > 0 {
		return v.Pairs()[0].Left().TypeFnc()
	}
	return None
}

func (v SetFnc) KeyNatType() d.TyNative {
	if v.Len() > 0 {
		return v.Pairs()[0].Left().TypeNat()
	}
	return d.Nil
}

func (v SetFnc) ValFncType() TyFnc {
	if v.Len() > 0 {
		return v.Pairs()[0].Right().TypeFnc()
	}
	return None
}

func (v SetFnc) ValNatType() d.TyNative {
	if v.Len() > 0 {
		return v.Pairs()[0].Right().TypeNat()
	}
	return d.Nil
}

func (v SetFnc) DeCap() (Callable, Consumeable) {
	return v.Head(), v.Tail()
}

func (v SetFnc) Head() Callable {
	if v.Len() > 0 {
		return v.Pairs()[0]
	}
	return nil
}

func (v SetFnc) Tail() Consumeable {
	if v.Len() > 1 {
		return ConRecordFromPairs(v.Pairs()[1:]...)
	}
	return NewEmptyRecord()
}
