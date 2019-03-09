/*
DATA CONSTRUCTORS

  implementations of 'precedence types', ake functional base-/ and collection types
*/
package functions

import (
	"strings"

	d "github.com/joergreinhardt/gatwd/data"
	"github.com/olekukonko/tablewriter"
)

type (
	// FUNCTIONAL COLLECTIONS (depend on enclosed data
	ListFnc     func() (Parametric, ListFnc)
	PairFnc     func() (a, b Parametric)
	VecFnc      func() []Parametric
	RecordFnc   func() []Paired
	AssocSetFnc func() d.Mapped
)

func ElemEmpty(val Parametric) bool {
	if val != nil {
		if !val.TypeFnc().Flag().Match(None) {
			return false
		}
	}
	return true
}

///// RECURSIVE LIST
////
/// base implementation of recursively linked lists
func NewRecursiveList(fncs ...Parametric) ListFnc {
	if len(fncs) > 0 {
		if len(fncs) > 1 {
			return ListFnc(func() (Parametric, ListFnc) {
				return fncs[0], NewRecursiveList(fncs[1:]...)
			})
		}
		return ListFnc(func() (Parametric, ListFnc) { return fncs[0], nil })
	}
	return nil
}
func (l ListFnc) Ident() Parametric           { return l }
func (l ListFnc) Head() Parametric            { h, _ := l(); return h }
func (l ListFnc) Tail() Consumeable           { _, t := l(); return t }
func (l ListFnc) TypeFnc() TyFnc              { return List | Functor }
func (l ListFnc) Eval(p ...d.Native) d.Native { return NewPair(l.Head(), l.Tail()) }
func (l ListFnc) TypeNat() d.TyNative         { return d.List.TypeNat() | l.Head().TypeNat() }
func (l ListFnc) Len() int {
	var _, t = l()
	if t != nil {
		return 1 + t.Len()
	}
	return 1
}
func (l ListFnc) Empty() bool {
	var list Consumeable = l
	var elem Parametric
	for elem, list = list.DeCap(); !ElemEmpty(elem); {
		return false
	}
	return true
}
func (l ListFnc) Call(d ...Parametric) Parametric {
	if len(d) > 0 {
		var head, tail = l()
		return NewPair(head, tail)
	}
	return l
}
func (l ListFnc) DeCap() (Parametric, Consumeable) {
	var head, rec = l()
	l = ListFnc(func() (Parametric, ListFnc) { return l() })
	return head, rec
}
func (l ListFnc) Con(val Parametric) ListFnc {
	return ListFnc(func() (Parametric, ListFnc) { return val, l })
}
func (l ListFnc) Map(fnc Parametric) ListFnc {
	return MapList(fnc, l)
}
func (l ListFnc) FoldL(fnc Parametric, init Parametric) Parametric {
	return FoldLList(fnc, l, init)
}
func (l ListFnc) RFoldL(fnc Parametric, init Parametric) Parametric {
	return RFoldLList(fnc, l, init)
}

///////////////////////////////////////////////////
//// VECTOR
///
// vector is a list backed by a slice.
func conVec(vec Vectorized, fncs ...Parametric) VecFnc {
	return conVecFromFunctionals(append(vec.Slice(), fncs...)...)
}
func conVecFromFunctionals(fncs ...Parametric) VecFnc {
	return VecFnc(func() []Parametric { return fncs })
}
func NewVector(fncs ...Parametric) VecFnc {
	return VecFnc(func() (vec []Parametric) {
		for _, dat := range fncs {
			vec = append(vec, New(dat))
		}
		return vec
	})
}
func (v VecFnc) TypeFnc() TyFnc              { return Vector | Functor }
func (v VecFnc) Ident() Parametric           { return v }
func (v VecFnc) Eval(p ...d.Native) d.Native { return NewVector(v()...) }
func (v VecFnc) TypeNat() d.TyNative {
	if len(v()) > 0 {
		return d.Vector.TypeNat() | v.Head().TypeNat()
	}
	return d.Vector.TypeNat() | d.Nil.TypeNat()
}

func (v RecordFnc) Atomic() bool {
	if v.Len() > 0 {
		for _, pair := range v() {
			if !ElemEmpty(pair.Left()) || !ElemEmpty(pair.Right()) {
				return false
			}
		}
	}
	return true
}
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

// base implementation functions/sliceable interface
func (v VecFnc) Head() Parametric {
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
func (v VecFnc) DeCap() (Parametric, Consumeable) {
	return v.Head(), v.Tail()
}
func (v VecFnc) Vector() []Parametric               { return v() }
func (v VecFnc) Slice() []Parametric                { return v() }
func (v VecFnc) Con(arg ...Parametric) []Parametric { return append(v(), arg...) }
func (v VecFnc) Atomic() bool {
	if v.Len() > 0 {
		for _, elem := range v() {
			if !ElemEmpty(elem) {
				return false
			}
		}
	}
	return true
}
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
		return VecFnc(func() []Parametric { return slice })

	}
	return nil
}
func (v VecFnc) Get(i int) Parametric {
	if i < v.Len() {
		return v()[i]
	}
	return nil
}
func (v VecFnc) Search(praed Parametric) int { return newDataSorter(v()...).Search(praed) }
func (v VecFnc) Sort(flag d.TyNative) {
	var ps = newDataSorter(v()...)
	ps.Sort(flag)
	v = NewVector(ps...)
}
func (v VecFnc) Map(fnc Parametric) VecFnc {
	return MapVector(fnc, v)
}
func (v VecFnc) FoldL(fnc Parametric, init Parametric) Parametric {
	return FoldLVector(fnc, v, init)
}
func (v VecFnc) RFoldL(fnc Parametric, init Parametric) Parametric {
	return RFoldLVector(fnc, v, init)
}

/////////////////////////////////////////////////////////
// PAIR
func NewPair(l, r Parametric) PairFnc {
	return PairFnc(func() (Parametric, Parametric) { return l, r })
}
func NewPairFromInterface(l, r interface{}) PairFnc {
	return PairFnc(func() (Parametric, Parametric) { return New(d.New(l)), New(d.New(r)) })
}
func NewPairFromData(l, r d.Native) PairFnc {
	return PairFnc(func() (Parametric, Parametric) { return New(l), New(r) })
}
func (p PairFnc) Both() (Parametric, Parametric) { return p() }
func (p PairFnc) Pair() Parametric               { return p }
func (p PairFnc) Left() Parametric               { l, _ := p(); return l }
func (p PairFnc) Right() Parametric              { _, r := p(); return r }
func (p PairFnc) Acc() Parametric                { return p.Left() }
func (p PairFnc) Arg() Parametric                { return p.Right() }
func (p PairFnc) AccType() d.TyNative            { return p.Left().TypeNat() }
func (p PairFnc) ArgType() d.TyNative            { return p.Right().TypeNat() }
func (p PairFnc) Ident() Parametric              { return p }
func (p PairFnc) Call(...Parametric) Parametric  { return p }
func (p PairFnc) Eval(a ...d.Native) d.Native    { return d.NewPair(p.Left().Eval(), p.Right().Eval()) }
func (p PairFnc) TypeFnc() TyFnc                 { return Pair | Function }
func (p PairFnc) TypeNat() d.TyNative {
	return d.Pair.TypeNat() | p.Left().TypeNat() | p.Right().TypeNat()
}
func (p PairFnc) Empty() bool {
	return ElemEmpty(p.Left()) && ElemEmpty(p.Right())
}
func (p PairFnc) Map(fnc Parametric) Parametric {
	return fnc.Call(p.Right())
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
	return RecordFnc(func() []Paired { return pairs })
}
func conRecordFromPairs(pp ...Paired) RecordFnc {
	return RecordFnc(func() []Paired { return pp })
}
func NewRecord(pp ...Paired) RecordFnc {
	return RecordFnc(func() (pairs []Paired) {
		for _, pair := range pp {
			pairs = append(pairs, pair)
		}
		return pairs
	})
}
func (v RecordFnc) Len() int    { return len(v()) }
func (v RecordFnc) Empty() bool { return ElemEmpty(v.Head()) && (len(v()) == 0) }
func (v RecordFnc) Get(idx int) Paired {
	if idx < v.Len()-1 {
		return v()[idx]
	}
	return nil
}
func (v RecordFnc) GetVal(praed Parametric) Paired  { return newPairSorter(v()...).Get(praed) }
func (v RecordFnc) Range(praed Parametric) []Paired { return newPairSorter(v()...).Range(praed) }
func (v RecordFnc) Search(praed Parametric) int     { return newPairSorter(v()...).Search(praed) }
func (v RecordFnc) Pairs() []Paired                 { return v() }
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
	return nil
}
func (v RecordFnc) Tail() Consumeable {
	if v.Len() > 1 {
		return conRecordFromPairs(v.Pairs()[1:]...)
	}
	return nil
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
func (v RecordFnc) Map(fnc Parametric) RecordFnc {
	return MapRecord(fnc, v)
}
func (v RecordFnc) FoldL(fnc Parametric, init Parametric) Parametric {
	return FoldLRecord(fnc, v, init)
}
func (v RecordFnc) ZipPair(left, right PairFnc) RecordFnc {
	return ZipPairsToRecord(left, right)
}
func (v RecordFnc) ZipLists(left, right VecFnc) RecordFnc {
	return ZipListsToRecord(left, right)
}
func (v RecordFnc) ZipAlternatingList(list ListFnc) RecordFnc {
	return ZipLAlternatingListToRecord(list)
}
func (v RecordFnc) ZipAlternatingVector(vec VecFnc) RecordFnc {
	return ZipLAlternatingVecToRecord(vec)
}
func (v RecordFnc) ZipAlternatingParams(params ...Parametric) RecordFnc {
	return ZipLAlternatingParamsToRecord(params...)
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
	return AssocSetFnc(func() d.Mapped { return set })
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
func (v AssocSetFnc) Len() int     { return v().Len() }
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
	return AssocSetFnc(func() d.Mapped { return m })
}
func (v AssocSetFnc) Slice() []Parametric {
	var pairs = []Parametric{}
	for _, pair := range v.Pairs() {
		pairs = append(pairs, pair)
	}
	return pairs
}
func (v AssocSetFnc) Empty() bool {
	if v.Len() > 0 {
		for _, field := range v().Fields() {
			if field != nil {
				if field.Left() != nil && field.Right() != nil {
					if !field.
						Left().
						TypeNat().
						Flag().
						Match(
							d.Nil.Flag(),
						) ||
						!field.
							Right().
							TypeNat().
							Flag().
							Match(
								d.Nil.Flag(),
							) {
						return false
					}
				}
			}
		}
	}
	return true
}
func (v AssocSetFnc) Map(fnc Parametric) AssocSetFnc {
	return MapAssocSet(fnc, v)
}
func (v AssocSetFnc) FoldL(fnc Parametric, init Parametric) Parametric {
	return FoldLAssocSet(fnc, v, init)
}
func (v AssocSetFnc) ZipPair(left, right PairFnc) AssocSetFnc {
	return ZipPairsToSet(left, right)
}
func (v AssocSetFnc) ZipLists(left, right ListFnc) AssocSetFnc {
	return ZipListsToSet(left, right)
}
func (v AssocSetFnc) ZipAlternatingList(list ListFnc) AssocSetFnc {
	return ZipLAlternatingListToSet(list)
}
func (v AssocSetFnc) ZipAlternatingVector(vec VecFnc) AssocSetFnc {
	return ZipLAlternatingVecToSet(vec)
}
func (v AssocSetFnc) ZipAlternatingParams(params ...Parametric) AssocSetFnc {
	return ZipLAlternatingParamsToSet(params...)
}
func (v AssocSetFnc) Call(f ...Parametric) Parametric { return v }
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
func (v AssocSetFnc) TypeFnc() TyFnc      { return MuliSet | Accessor | Functor }
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

/// MAP, FOLD & ZIP
//
// map implements functional (aka functor) interface
func MapList(expr Parametric, list Consumeable) ListFnc {
	var elem Parametric
	var vec = []Parametric{}
	for elem, list = list.DeCap(); !ElemEmpty(elem); {
		vec = append(vec, expr.Call(elem))
	}
	return NewRecursiveList(vec...)
}

func MapVector(expr Parametric, vector VecFnc) VecFnc {
	var slice = []Parametric{}
	for _, elem := range vector.Slice() {
		if !ElemEmpty(elem) {
			slice = append(slice, expr.Call(elem))
		}
	}
	return NewVector(slice...)
}

func MapRecord(expr Parametric, record RecordFnc) RecordFnc {
	var pairs = []Paired{}
	for _, rec := range record.Pairs() {
		if pair, ok := expr.Call(rec).(Paired); ok {
			pairs = append(pairs, pair)
		}
	}
	return NewRecord(pairs...)
}

func MapAssocSet(expr Parametric, assocs AssocSetFnc) AssocSetFnc {
	var pairs = []Paired{}
	for _, pair := range assocs.Pairs() {
		pairs = append(pairs, pair)
	}
	return conAssocSet(pairs...)
}

/// FOLDL
// initial expression is called once expecting it's predeccessor & each
// succeeding element of the list as arguments.
func FoldLList(
	expr Parametric,
	list Consumeable,
	init Parametric,
) Parametric {
	var elem Parametric
	for elem, list = list.DeCap(); !ElemEmpty(elem); {
		init = init.Call(init, elem)
	}
	return init
}

func FoldLVector(
	expr Parametric,
	vector VecFnc,
	init Parametric,
) Parametric {
	for _, elem := range vector() {
		init = init.Call(init, elem)
	}
	return init
}

func FoldLRecord(
	expr Parametric,
	record RecordFnc,
	init Parametric,
) Parametric {
	for _, pair := range record() {
		init = init.Call(init, pair)
	}
	return init
}

func FoldLAssocSet(
	expr Parametric,
	assoc AssocSetFnc,
	init Parametric,
) Parametric {
	for _, pair := range assoc.Pairs() {
		init = init.Call(init, pair)
	}
	return init
}

/// RFOLDL
func RFoldLList(
	expr Parametric,
	list Consumeable,
	init Parametric,
) Parametric {
	var reverser = []Parametric{}
	var elem Parametric
	for elem, list = list.DeCap(); !ElemEmpty(elem); {
		reverser = append(reverser, elem)
	}
	return FoldLList(expr, NewRecursiveList(reverser...), init)
}

func RFoldLVector(
	expr Parametric,
	vector VecFnc,
	init Parametric,
) Parametric {
	for i := vector.Len() - 1; i > 0; i-- {
		init = init.Call(init, vector.Get(i))
	}
	return init
}

/// ZIP
// zip to set
func ZipPairsToSet(pairs ...PairFnc) AssocSetFnc {
	return NewAssocSet(pairs...)
}
func ZipListsToSet(llist, rlist Consumeable) AssocSetFnc {
	var left, right Parametric
	var pairs = []PairFnc{}
	for left, llist = llist.DeCap(); !ElemEmpty(left); {
		for right, rlist = rlist.DeCap(); !ElemEmpty(left); {
			pairs = append(pairs, NewPair(left, right))
		}
	}
	return ZipPairsToSet(pairs...)
}
func ZipLAlternatingParamsToSet(params ...Parametric) AssocSetFnc {
	return ZipLAlternatingVecToSet(NewVector(params...))
}
func ZipLAlternatingVecToSet(list VecFnc) AssocSetFnc {
	var pairs = []PairFnc{}
	var left Parametric
	for i, val := range list() {
		if i%2 == 0 { // append new pair from left & current, when
			// divideable by two
			pairs = append(pairs, NewPair(left, val))
		} else {
			left = val
		}
	}
	return ZipPairsToSet(pairs...)
}
func ZipLAlternatingListToSet(list Consumeable) AssocSetFnc {
	var pairs = []PairFnc{}
	var left Parametric
	var elem Parametric
	for elem, list = list.DeCap(); !ElemEmpty(elem); {
		if left != nil {
			pairs = append(pairs, NewPair(left, elem))
			left = nil
		} else {
			left = elem
		}
	}
	return NewAssocSet(pairs...)
}

// zip to record
func ZipPairsToRecord(pairs ...PairFnc) RecordFnc {
	return newRecord(pairs...)
}
func ZipListsToRecord(llist, rlist VecFnc) RecordFnc {
	var pairs = []PairFnc{}
	var l = llist.Len()
	var rl = rlist.Len()
	if rl < l {
		l = rl
	}
	for i := 0; i < l; i++ {
		pairs = append(pairs, NewPair(llist.Get(i), rlist.Get(i)))
	}
	return ZipPairsToRecord(pairs...)
}

func ZipLAlternatingParamsToRecord(params ...Parametric) RecordFnc {
	return ZipLAlternatingVecToRecord(NewVector(params...))
}

func ZipLAlternatingVecToRecord(list VecFnc) RecordFnc {
	var pairs = []PairFnc{}
	var left Parametric
	for i, val := range list() {
		if i%2 == 0 { // append new pair from left & current, when
			// divideable by two
			pairs = append(pairs, NewPair(left, val))
		} else {
			left = val
		}
	}
	return ZipPairsToRecord(pairs...)
}
func ZipLAlternatingListToRecord(list Consumeable) RecordFnc {
	var pairs = []PairFnc{}
	var left Parametric
	var elem Parametric
	for elem, list = list.DeCap(); !ElemEmpty(elem); {
		if left != nil {
			pairs = append(pairs, NewPair(left, elem))
			left = nil
		} else {
			left = elem
		}
	}
	return NewRecord()
}
