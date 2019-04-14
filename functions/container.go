/*
DATA CONSTRUCTORS

  implementations of 'precedence types', ake functional base-/ and collection types
*/
package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

type (

	// NATIVE DATA (aliased natives implementing parametric)
	Native func() d.Native

	/// PURE FUNCTIONS (sole dependece on argset)
	ConstFnc  func() Callable
	UnaryFnc  func(Callable) Callable
	BinaryFnc func(a, b Callable) Callable
	NaryFnc   func(...Callable) Callable

	// FUNCTIONAL COLLECTIONS (depend on enclosed data
	PairFnc   func(arg ...d.Native) (Callable, Callable)
	TupleFnc  func(arg ...d.Native) VecFnc
	ListFnc   func(elems ...Callable) (Callable, ListFnc)
	VecFnc    func(elems ...Callable) []Callable
	RecordFnc func(pairs ...PairFnc) []PairFnc
	SetFnc    func(pairs ...PairFnc) d.Mapped
)

//////////////////////////////////////////////////////////////////////////////
//// PLAIN FUNCTIONS
///
// CONSTANT FUNCTION
//
// constant also conains immutable data that may be an instance of a type of
// the data package, or result of a function call guarantueed to allways return
// the same value.
func (c ConstFnc) Ident() Callable             { return c() }
func (c ConstFnc) TypeFnc() TyFnc              { return Function }
func (c ConstFnc) TypeNat() d.TyNative         { return c().TypeNat() }
func (c ConstFnc) Eval(p ...d.Native) d.Native { return c().Eval() }
func (c ConstFnc) Call(d ...Callable) Callable { return c() }

///// UNARY FUNCTION
func (u UnaryFnc) TypeNat() d.TyNative         { return d.Function.TypeNat() }
func (u UnaryFnc) TypeFnc() TyFnc              { return Function }
func (u UnaryFnc) Ident() Callable             { return u }
func (u UnaryFnc) Eval(p ...d.Native) d.Native { return u }
func (u UnaryFnc) Call(d ...Callable) Callable {
	return u(d[0])
}

///// BINARY FUNCTION
func (b BinaryFnc) TypeNat() d.TyNative         { return d.Function.TypeNat() }
func (b BinaryFnc) TypeFnc() TyFnc              { return Function }
func (b BinaryFnc) Ident() Callable             { return b }
func (b BinaryFnc) Eval(p ...d.Native) d.Native { return b }
func (b BinaryFnc) Call(d ...Callable) Callable { return b(d[0], d[1]) }

///// NARY FUNCTION
func (n NaryFnc) TypeNat() d.TyNative         { return d.Function.TypeNat() }
func (n NaryFnc) TypeFnc() TyFnc              { return Function }
func (n NaryFnc) Ident() Callable             { return n }
func (n NaryFnc) Eval(p ...d.Native) d.Native { return n }
func (n NaryFnc) Call(d ...Callable) Callable { return n(d...) }

//// (RE-) INSTANCIATE PRIMARY DATA TO IMPLEMENT FUNCTIONS VALUE INTERFACE
///
//
func NewNative(nat d.Native) Native {
	return func() d.Native {
		return nat
	}
}

func (n Native) String() string                 { return n().String() }
func (n Native) Eval(args ...d.Native) d.Native { return n().Eval(args...) }
func (n Native) TypeNat() d.TyNative            { return n().TypeNat() }
func (n Native) TypeFnc() TyFnc                 { return Data }
func (n Native) Empty() bool {
	if n != nil {
		if !d.Nil.Flag().Match(n.TypeNat()) {
			if !None.Flag().Match(n.TypeFnc()) {
				return false
			}
		}
	}
	return true
}

func (n Native) Call(vals ...Callable) Callable { return n }

func New(inf ...interface{}) Callable { return NewFromData(d.New(inf...)) }

func NewFromData(data ...d.Native) Native {
	var result d.Native
	if len(data) == 1 {
		result = data[0]
	} else {
		result = d.DataSlice(data)
	}
	return func() d.Native { return result }
}

//////////////////////////////////////////////////////////////////////////////
//// PAIR
///
//
func NewPair(l, r Callable) PairFnc {
	return func(args ...d.Native) (Callable, Callable) {
		if len(args) > 0 {
		}
		return l, r
	}
}
func NewEmptyPair() PairFnc {
	return func(args ...d.Native) (a, b Callable) {
		return NewNoOp(), NewNoOp()
	}
}
func NewPairFromInterface(l, r interface{}) PairFnc {
	return func(arg ...d.Native) (Callable, Callable) {
		return New(d.New(l)), New(d.New(r))
	}
}
func NewPairFromData(l, r d.Native) PairFnc {
	return func(args ...d.Native) (Callable, Callable) {
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

////////////////////////////////////////////////////////////
//// TUPLE
func NewTuple(data ...Callable) TupleFnc {
	return func(arg ...d.Native) VecFnc {
		if len(arg) > 0 {
		}
		return NewVector(data...)
	}
}

func (t TupleFnc) Ident() Callable { return t }
func (t TupleFnc) Call(args ...Callable) Callable {
	if len(args) > 0 {
	}
	return t()
}
func (t TupleFnc) Eval(args ...d.Native) d.Native {
	if len(args) > 0 {
	}
	return t().Eval()
}
func (t TupleFnc) TypeNat() d.TyNative { return d.Tuple }
func (t TupleFnc) TypeFnc() TyFnc      { return Tuple }
func (t TupleFnc) String() string      { return t().String() }

///// RECURSIVE LIST
////
/// base implementation of recursively linked lists
//
// generate empty lists out of thin air
func NewList(args ...Callable) ListFnc {
	return ConList(EmptyList(), args...)
}

func EmptyList() ListFnc {
	return func(args ...Callable) (Callable, ListFnc) {
		if len(args) == 0 {
			return nil, EmptyList()
		}
		if len(args) == 1 {
			return args[0], EmptyList()
		}
		return args[0], ConList(EmptyList(), args[1:]...)
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
		return func(args ...Callable) (Callable, ListFnc) {
			// either head element and the initial list (which
			// would be a list with the head element as it's only
			// element)
			if len(args) == 0 {
				return head, list
			}
			// or return the initial list followed by the elements
			// passed to the inner function, followed by the
			// initial head
			return ConList(list, append(args, head)...)()
		}
	}

	// if more elements have been passed, lazy concat them with the initial list
	return func(args ...Callable) (Callable, ListFnc) {
		// no elements → return head and list
		if len(args) == 0 {
			return head, ConList(list, initials[1:]...)
		}
		// elements got passed, append to list. to get order of passed
		// elements & head right, concat all and call resutling list,
		// to yield new head & tail list.
		return ConList(list, append(args, initials...)...)()
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
			if pair, ok := val.(PairFnc); ok {
				v = v.Con(pair)
			}
		}
	}
	return v
}

func (v RecordFnc) Con(p ...Callable) RecordFnc {
	return v.Con(p...)
}

func (v RecordFnc) DeCap() (Callable, Consumeable) {
	return v.Head(), v.Tail()
}

func (v RecordFnc) Empty() bool {
	if len(v()) > 0 {
		for _, pair := range v() {
			if !pair.Empty() {
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
func ConRecord(vec Associative, pfnc ...PairFnc) RecordFnc {
	return ConRecordFromPairs(append(vec.Pairs(), pfnc...)...)
}

func NewRecordFromPairFunction(ps ...PairFnc) RecordFnc {
	var pairs = []PairFnc{}
	for _, pair := range ps {
		pairs = append(pairs, pair)
	}
	return RecordFnc(func(pairs ...PairFnc) []PairFnc { return pairs })
}

func ConRecordFromPairs(pp ...PairFnc) RecordFnc {
	return RecordFnc(func(pairs ...PairFnc) []PairFnc { return pp })
}

func NewEmptyRecord() RecordFnc {
	return RecordFnc(func(pairs ...PairFnc) []PairFnc { return []PairFnc{} })
}

func NewRecord(pp ...PairFnc) RecordFnc {
	return func(pairs ...PairFnc) []PairFnc {
		for _, pair := range pp {
			pairs = append(pairs, pair)
		}
		return pairs
	}
}

func (v RecordFnc) Len() int { return len(v()) }

func (v RecordFnc) Get(idx int) PairFnc {
	if idx < v.Len()-1 {
		return v()[idx]
	}
	return NewPair(NewNoOp(), NewNoOp())
}

func (v RecordFnc) GetVal(praed Callable) PairFnc {
	return newPairSorter(v()...).Get(praed)
}

func (v RecordFnc) Range(praed Callable) []PairFnc {
	return newPairSorter(v()...).Range(praed)
}

func (v RecordFnc) Search(praed Callable) int {
	return newPairSorter(v()...).Search(praed)
}

func (v RecordFnc) Pairs() []PairFnc {
	return v()
}

func (v RecordFnc) SwitchedPairs() []PairFnc {
	var switched = []PairFnc{}
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

///////////////////////////////////////////////////////////////////////
//// ASSOCIATIVE SET (HASH MAP OF VALUES)
///
// associative array that uses pairs left field as accessor for sort & search
func ConAssocSet(pairs ...PairFnc) SetFnc {
	var paired = []PairFnc{}
	for _, pair := range pairs {
		paired = append(paired, pair)
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
	return SetFnc(func(pairs ...PairFnc) d.Mapped { return set })
}

func (v SetFnc) Split() (VecFnc, VecFnc) {
	var keys, vals = []Callable{}, []Callable{}
	for _, pair := range v.Pairs() {
		keys = append(keys, pair.Left())
		vals = append(vals, pair.Right())
	}
	return NewVector(keys...), NewVector(vals...)
}

func (v SetFnc) Pairs() []PairFnc {
	var pairs = []PairFnc{}
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
		if !pair.Empty() {
			return false
		}
	}
	return true
}

func (v SetFnc) GetVal(praed Callable) PairFnc {
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
	return SetFnc(func(pairs ...PairFnc) d.Mapped { return m })
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
