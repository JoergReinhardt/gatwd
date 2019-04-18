/*
DATA CONSTRUCTORS

  implementations of 'precedence types', ake functional base-/ and collection types
*/
package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

type (
	// NATIVE DATA
	Native func() d.Native

	/// STATIC EXPRESSIONS
	ConstFnc  func() Callable
	UnaryFnc  func(Callable) Callable
	BinaryFnc func(a, b Callable) Callable
	NaryFnc   func(...Callable) Callable

	// FUNCTIONAL COLLECTIONS
	PairVal        func(...Callable) (Callable, Callable)
	ListVal        func(...Callable) (Callable, ListVal)
	VecVal         func(...Callable) []Callable
	TupleVal       func(...Callable) []Callable
	AccociativeVal func(...PairVal) []PairVal
	SetVal         func(...PairVal) d.Mapped

	// MONADIC VALUES
	NoOp     func()
	TruthVal func() bool
)

//// DATA INSTANCIATION
///
// 'new' instanciates all kinds of value instances as well as literals
// automagicaly figuring out what type seems to be appropriate.
func New(inf ...interface{}) Callable { return NewFromData(d.New(inf...)) }

// 'new from data' expects an instanciate implementing the 'data/Native'
// interface and wraps it in a function to implement the Callable interface and
// return the enclosedndata as instance implementing the data/Native interface
func NewFromData(data ...d.Native) Native {
	var result d.Native
	if len(data) == 1 {
		result = data[0]
	} else {
		result = d.DataSlice(data)
	}
	return func() d.Native { return result }
}

func (n Native) TypeFnc() TyFnc                 { return Data }
func (n Native) String() string                 { return n().String() }
func (n Native) Eval(args ...d.Native) d.Native { return n().Eval(args...) }
func (n Native) TypeNat() d.TyNative            { return n().TypeNat() }
func (n Native) Call(vals ...Callable) Callable { return n }
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

//////////////////////////////////////////////////////////////////////////////
//// STATIC EXPRESSIONS
///
// static function expressions allways yield results of the same type,
// independent from passed arguments. expressions don't take any arguments,
// take one, two, or multiple, arguments are quite common and have dedicated
// signatures for performance reasons.

// CONSTANT FUNCTION
func (c ConstFnc) Ident() Callable           { return c() }
func (c ConstFnc) TypeFnc() TyFnc            { return Function }
func (c ConstFnc) TypeNat() d.TyNative       { return c().TypeNat() }
func (c ConstFnc) Call(...Callable) Callable { return c() }
func (c ConstFnc) Eval(...d.Native) d.Native { return c().Eval() }

///// UNARY FUNCTION
func (u UnaryFnc) Ident() Callable               { return u }
func (u UnaryFnc) TypeFnc() TyFnc                { return Function }
func (u UnaryFnc) TypeNat() d.TyNative           { return d.Function.TypeNat() }
func (u UnaryFnc) Call(arg ...Callable) Callable { return u(arg[0]) }
func (u UnaryFnc) Eval(arg ...d.Native) d.Native { return u(NewFromData(arg...)) }

///// BINARY FUNCTION
func (b BinaryFnc) Ident() Callable                { return b }
func (b BinaryFnc) TypeFnc() TyFnc                 { return Function }
func (b BinaryFnc) TypeNat() d.TyNative            { return d.Function.TypeNat() }
func (b BinaryFnc) Call(args ...Callable) Callable { return b(args[0], args[1]) }
func (b BinaryFnc) Eval(args ...d.Native) d.Native {
	return b(NewFromData(args[0]), NewFromData(args[1]))
}

///// NARY FUNCTION
func (n NaryFnc) Ident() Callable             { return n }
func (n NaryFnc) TypeFnc() TyFnc              { return Function }
func (n NaryFnc) TypeNat() d.TyNative         { return d.Function.TypeNat() }
func (n NaryFnc) Call(d ...Callable) Callable { return n(d...) }
func (n NaryFnc) Eval(args ...d.Native) d.Native {
	var params = []Callable{}
	for _, arg := range args {
		params = append(params, NewFromData(arg))
	}
	return n(params...)
}

//////////////////////////////////////////////////////////////////////////////
//// PAIRS OF VALUES
func NewEmptyPair() PairVal {
	return func(args ...Callable) (a, b Callable) {
		return NewNoOp(), NewNoOp()
	}
}

func NewPair(l, r Callable) PairVal {
	return func(args ...Callable) (Callable, Callable) {
		if len(args) > 0 {
		}
		return l, r
	}
}

func NewPairFromData(l, r d.Native) PairVal {
	return func(args ...Callable) (Callable, Callable) {
		return New(l), New(r)
	}
}

func NewPairFromInterface(l, r interface{}) PairVal {
	return func(arg ...Callable) (Callable, Callable) {
		return New(d.New(l)), New(d.New(r))
	}
}

// con pairs from list of arguments
func ConPair(list Consumeable) (PairVal, Consumeable) {

	var left Callable

	left, list = list.DeCap()

	if left != nil {
		if list != nil {
			return NewPair(left, list.Head()), list.Tail()
		}
		return NewPair(left, left.TypeFnc()), nil
	}
	// all arguments depleted
	return nil, list
}

func (p PairVal) Both() (Callable, Callable)     { return p() }
func (p PairVal) Head() Callable                 { l, _ := p(); return l }
func (p PairVal) Tail() Consumeable              { _, r := p(); return NewPair(r, NewNoOp()) }
func (p PairVal) DeCap() (Callable, Consumeable) { l, r := p(); return l, NewList(r) }
func (p PairVal) Right() Callable                { _, r := p(); return r }
func (p PairVal) Left() Callable                 { l, _ := p(); return l }

func (p PairVal) Ident() Callable      { return p }
func (p PairVal) Pair() Callable       { return p }
func (p PairVal) Acc() Callable        { return p.Left() }
func (p PairVal) Arg() Callable        { return p.Right() }
func (p PairVal) AccType() d.TyNative  { return p.Left().TypeNat() }
func (p PairVal) ArgType() d.TyNative  { return p.Right().TypeNat() }
func (p PairVal) HeadType() d.TyNative { return p.Left().TypeNat() }
func (p PairVal) TailType() d.TyNative { return p.Right().TypeNat() }

func (p PairVal) TypeFnc() TyFnc {
	return Pair | p.Left().TypeFnc() | p.Right().TypeFnc()
}

func (p PairVal) Empty() bool {
	return p.Left() == nil && p.Right() == nil
}

func (p PairVal) TypeNat() d.TyNative {
	return p.Left().TypeNat() | p.Right().TypeNat()
}

func (p PairVal) Call(args ...Callable) Callable {
	return NewPair(p.Left().Call(args...), p.Right().Call(args...))
}

func (p PairVal) Eval(args ...d.Native) d.Native {
	return d.NewPair(p.Left().Eval(args...), p.Right().Eval(args...))
}

//////////////////////////////////////////////////////////////////////////////////////
///// RECURSIVE LIST OF VALUES
////
/// base implementation of recursively linked lists
//
// recursive list function holds a list of values on a late binding call by
// name base. when called without arguments, list function returns the current
// head element and a continuation, to fetch the preceeding one only, when
// called.
//
// when arguments are passed, list function concatenates them to the existing
// list lazyly. arguments stay unevaluated until they are returned. list
// function is a monad and base type of many other monads
func NewList(elems ...Callable) ListVal {

	// function litereal closes over initial list and needs to deal with arguments
	return func(args ...Callable) (Callable, ListVal) {

		var head Callable

		// pass additional arguments on to be returned as head in
		// preceeding calls until argument depletion.
		if len(args) > 0 {
			// take first argument as head to return
			head = args[0]
			// remaining arguments are parameters to generate the
			// list of continuations.
			if len(args) > 1 {
				// append previously existing elements to set
				// of args to generate a new list from.
				return head, NewList(append(args[1:], elems...)...)
			}
			// last argument is returned as head, return previously
			// existing list as tail to smoothly hand over (Krueger
			// industial smoothing‥. we don't care and it shows)
			return head, NewList(elems...)
		}

		// as long as there are elements
		if len(elems) > 0 {
			// assign first element to head
			head = elems[0]
			// if there are further elements, return a list
			// continuation to contain them.
			if len(elems) > 1 {
				return head, NewList(elems[1:]...)
			}
			// otherwise return last element and replace depleted
			// list, with an empty one for convienience
			return head, NewList()
		}
		// return neither head nor tail
		return nil, nil
	}
}

func (l ListVal) Ident() Callable { return l }

func (l ListVal) Null() ListVal { return NewList() }

func (l ListVal) Tail() Consumeable { _, t := l(); return t }

func (l ListVal) Head() Callable { h, _ := l(); return h }

func (l ListVal) DeCap() (Callable, Consumeable) { return l() }

func (l ListVal) TypeFnc() TyFnc { return List | Functor }

func (l ListVal) TypeNat() d.TyNative { return d.List.TypeNat() | l.Head().TypeNat() }

func (l ListVal) Eval(args ...d.Native) d.Native {
	var parms = []Callable{}

	for _, arg := range args {
		parms = append(parms, NewFromData(arg))
	}

	head, _ := l(parms...)
	return head
}

func (l ListVal) Empty() bool {
	if l.Head() == nil {
		return true
	}
	return false
}

func (l ListVal) Len() int {
	var length int
	var head, tail = l()
	if head != nil {
		length += 1 + tail.Len()
	}
	return length
}

func (l ListVal) Call(d ...Callable) Callable {
	var head Callable
	head, l = l(d...)
	return head
}

//////////////////////////////////////////////////////////////////////////////////////////
//// VECTORS (SLICES) OF VALUES
func NewEmptyVector(init ...Callable) VecVal { return NewVector() }

func NewVector(init ...Callable) VecVal {

	var vector = init

	return func(args ...Callable) []Callable {

		if len(args) > 0 {
			// add args to vector
			vector = append(
				vector,
				args...,
			)
		}
		// decapitate vector
		return vector
	}
}

func ConVector(vec Vectorized, args ...Callable) VecVal {
	return ConVecFromCallable(append(vec.Slice(), args...)...)
}

func ConVecFromCallable(init ...Callable) VecVal {

	return func(args ...Callable) []Callable {
		return init
	}
}

func (v VecVal) Ident() Callable { return v }

func (v VecVal) Call(d ...Callable) Callable { return NewVector(v(d...)...) }

func (v VecVal) Eval(args ...d.Native) d.Native {
	var parms = []Callable{}
	var result = []d.Native{}
	for _, arg := range args {
		parms = append(parms, NewFromData(arg))
	}
	for _, parm := range v(parms...) {
		result = append(result, parm.Eval())
	}
	return d.DataSlice(result)
}

func (v VecVal) TypeFnc() TyFnc { return Vector | Functor }

func (v VecVal) TypeNat() d.TyNative {
	if len(v()) > 0 {
		return d.Vector.TypeNat() | v.Head().TypeNat()
	}
	return d.Vector.TypeNat() | d.Nil.TypeNat()
}

func (v VecVal) Head() Callable {
	if v.Len() > 0 {
		return v.Vector()[0]
	}
	return nil
}

func (v VecVal) Tail() Consumeable {
	if v.Len() > 1 {
		return ConVecFromCallable(v.Vector()[1:]...)
	}
	return NewEmptyVector()
}

func (v VecVal) Empty() bool {
	if len(v()) > 0 {
		return false
	}
	return true
}

func (v VecVal) Len() int { return len(v()) }

func (v VecVal) DeCap() (Callable, Consumeable) {
	var l = len(v())
	if l > 0 {
		if l > 1 {
			return v.Head(), v.Tail()
		}
		return v.Head(), NewList()
	}
	return nil, NewList()
}

func (v VecVal) Vector() []Callable { return v() }

func (v VecVal) Slice() []Callable { return v() }

func (v VecVal) Con(arg ...Callable) []Callable { return append(v(), arg...) }

func (v VecVal) Set(i int, val Callable) Vectorized {
	if i < v.Len() {
		var slice = v()
		slice[i] = val
		return VecVal(func(elems ...Callable) []Callable { return slice })

	}
	return v
}

func (v VecVal) Get(i int) Callable {
	if i < v.Len() {
		return v()[i]
	}
	return NewNoOp()
}
func (v VecVal) Search(praed Callable) int { return newDataSorter(v()...).Search(praed) }

func (v VecVal) Sort(flag d.TyNative) {
	var ps = newDataSorter(v()...)
	ps.Sort(flag)
	v = NewVector(ps...)
}

/////////////////////////////////////////////////////////////////////////////////////
//// TUPLE TYPE VALUES
///
// tuples are sequences of values grouped in a predefined signature of types,
// number and order they are expected to appear in
func NewTupleType(init ...Callable) TupleVal { return TupleVal(NewVector(init...)) }

func (t TupleVal) Ident() Callable                { return t }
func (t TupleVal) Len() int                       { return len(t()) }
func (t TupleVal) DeCap() (Callable, Consumeable) { return VecVal(t).DeCap() }
func (t TupleVal) Head() Callable                 { return VecVal(t).Head() }
func (t TupleVal) Tail() Consumeable              { return VecVal(t).Tail() }

// functional type concatenates the functional types of all the subtypes
func (t TupleVal) TypeFnc() TyFnc {
	var ftype = TyFnc(0)
	for _, typ := range t() {
		ftype = ftype | typ.TypeFnc()
	}
	return ftype
}

// native type concatenates the native types of all the subtypes
func (t TupleVal) TypeNat() d.TyNative {
	var ntype = d.Tuple
	for _, typ := range t() {
		ntype = ntype | typ.TypeNat()
	}
	return ntype
}

// string representation of a tuple generates one row per sub type by
// concatenating each sub types native type, functional type and value.
func (t TupleVal) String() string { return t.Head().String() }

func (t TupleVal) Eval(args ...d.Native) d.Native { return t.Eval() }

func (t TupleVal) Call(args ...Callable) Callable { return t }

func (t TupleVal) ApplyPartial(args ...Callable) TupleVal {
	return NewTupleType(partialApplyTuple(t, args...)...)
}

func partialApplyTuple(tuple TupleVal, args ...Callable) []Callable {
	// fetch current tupple
	var result = tuple()
	var l = len(result)

	// range through arguments
	for i := 0; i < l; i++ {

		// pick argument by index
		var arg = args[i]

		// partial arguments can either be given by position, or in
		// pairs that contains the intendet position as integer value
		// in its left and the value itself in its right cell, so‥.
		if pair, ok := arg.(PairVal); ok {
			// ‥.and the left element is an integer‥.
			if pos, ok := pair.Left().(Integer); ok {
				// ‥.and that integer is within the range of indices‥.
				if l < pos.Int() {
					// ‥.and both types of the right element
					// match the corresponding result types
					// of the given index‥.
					if result[i].TypeFnc() == pair.Right().TypeFnc() &&
						result[i].TypeNat() == args[i].TypeNat() {
						// ‥.replace the value in
						// results, with right
						// element of pair.
						result[i] = pair.Right()
					}
				}
			}
		}
		// ‥.otherwise assume arguments are passed one element at a
		// time, altering between position & value and the current
		// index is expected to be the position, so if it's an uneven
		// index (positions)‥.
		if i%2 == 0 {
			var idx = i  // safe current index
			if i+1 < l { // check if next index is out of bounds
				i = i + 1 // advance loop counter by one
				// replace value in results at previous index
				// with value at index of the advanced loop
				// counter
				result[idx] = args[i]
			}
		}
	}

	// return altered result
	return result
}

///////////////////////////////////////////////////////////////////////////////////////////
//// ASSOCIATIVE SEQUENCE OF VALUES
///
// list of associative values in predefined order.
func ConAssociative(vec Associative, pfnc ...PairVal) AccociativeVal {
	return NewAssociativeFromPairFunction(append(vec.Pairs(), pfnc...)...)
}

func NewAssociativeFromPairFunction(ps ...PairVal) AccociativeVal {
	var pairs = []PairVal{}
	for _, pair := range ps {
		pairs = append(pairs, pair)
	}
	return AccociativeVal(func(pairs ...PairVal) []PairVal { return pairs })
}

func ConAssociativeFromPairs(pp ...PairVal) AccociativeVal {
	return AccociativeVal(func(pairs ...PairVal) []PairVal { return pp })
}

func NewEmptyAssociative() AccociativeVal {
	return AccociativeVal(func(pairs ...PairVal) []PairVal { return []PairVal{} })
}

func NewAssociative(pp ...PairVal) AccociativeVal {

	return func(pairs ...PairVal) []PairVal {
		for _, pair := range pp {
			pairs = append(pairs, pair)
		}
		return pairs
	}
}

func (v AccociativeVal) Call(d ...Callable) Callable {
	if len(d) > 0 {
		for _, val := range d {
			if pair, ok := val.(PairVal); ok {
				v = v.Con(pair)
			}
		}
	}
	return v
}

func (v AccociativeVal) Con(p ...Callable) AccociativeVal {

	var pairs = v.Pairs()

	return ConAssociativeFromPairs(pairs...)
}

func (v AccociativeVal) DeCap() (Callable, Consumeable) {
	return v.Head(), v.Tail()
}

func (v AccociativeVal) Empty() bool {
	if len(v()) > 0 {
		for _, pair := range v() {
			if !pair.Empty() {
				return false
			}
		}
	}
	return true
}

func (v AccociativeVal) KeyFncType() TyFnc {
	if v.Len() > 0 {
		return v.Pairs()[0].Left().TypeFnc()
	}
	return None
}

func (v AccociativeVal) KeyNatType() d.TyNative {
	if v.Len() > 0 {
		return v.Pairs()[0].Left().TypeNat()
	}
	return d.Nil
}

func (v AccociativeVal) ValFncType() TyFnc {
	if v.Len() > 0 {
		return v.Pairs()[0].Right().TypeFnc()
	}
	return None
}

func (v AccociativeVal) ValNatType() d.TyNative {
	if v.Len() > 0 {
		return v.Pairs()[0].Right().TypeNat()
	}
	return d.Nil
}

func (v AccociativeVal) TypeFnc() TyFnc { return Record | Functor }

func (v AccociativeVal) TypeNat() d.TyNative {
	if len(v()) > 0 {
		return d.Vector | v.Head().TypeNat()
	}
	return d.Vector | d.Nil.TypeNat()
}
func (v AccociativeVal) Len() int { return len(v()) }

func (v AccociativeVal) Get(idx int) PairVal {
	if idx < v.Len()-1 {
		return v()[idx]
	}
	return NewPair(NewNoOp(), NewNoOp())
}

func (v AccociativeVal) GetVal(praed Callable) PairVal {
	return newPairSorter(v()...).Get(praed)
}

func (v AccociativeVal) Range(praed Callable) []PairVal {
	return newPairSorter(v()...).Range(praed)
}

func (v AccociativeVal) Search(praed Callable) int {
	return newPairSorter(v()...).Search(praed)
}

func (v AccociativeVal) Pairs() []PairVal { return v() }

func (v AccociativeVal) DeCapPairWise() (PairVal, []PairVal) {

	var pairs = v()

	if len(pairs) > 0 {
		if len(pairs) > 1 {
			return pairs[0], pairs[1:]
		}
		return pairs[0], []PairVal{}
	}
	return nil, []PairVal{}
}

func (v AccociativeVal) SwitchedPairs() []PairVal {
	var switched = []PairVal{}
	for _, pair := range v() {
		switched = append(
			switched,
			NewPair(
				pair.Right(),
				pair.Left()))
	}
	return switched
}

func (v AccociativeVal) SetVal(key, value Callable) Associative {
	if idx := v.Search(key); idx >= 0 {
		var pairs = v.Pairs()
		pairs[idx] = NewPair(key, value)
		return NewAssociative(pairs...)
	}
	return NewAssociative(append(v.Pairs(), NewPair(key, value))...)
}

func (v AccociativeVal) Slice() []Callable {
	var fncs = []Callable{}
	for _, pair := range v() {
		fncs = append(fncs, NewPair(pair.Left(), pair.Right()))
	}
	return fncs
}

func (v AccociativeVal) Head() Callable {
	if v.Len() > 0 {
		return v.Pairs()[0]
	}
	return nil
}

func (v AccociativeVal) Tail() Consumeable {
	if v.Len() > 1 {
		return ConAssociativeFromPairs(v.Pairs()[1:]...)
	}
	return NewEmptyAssociative()
}

func (v AccociativeVal) Eval(p ...d.Native) d.Native {
	var slice = d.DataSlice{}
	for _, pair := range v() {
		d.SliceAppend(slice, d.NewPair(pair.Left(), pair.Right()))
	}
	return slice
}

func (v AccociativeVal) Sort(flag d.TyNative) {
	var ps = newPairSorter(v()...)
	ps.Sort(flag)
	v = NewAssociative(ps...)
}

//////////////////////////////////////////////////////////////////////////////
//// ASSOCIATIVE SET (HASH MAP OF VALUES)
///
// unordered associative set of values
func ConAssocSet(pairs ...PairVal) SetVal {
	var paired = []PairVal{}
	for _, pair := range pairs {
		paired = append(paired, pair)
	}
	return NewAssocSet(paired...)
}

func NewAssocSet(pairs ...PairVal) SetVal {

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
	return SetVal(func(pairs ...PairVal) d.Mapped { return set })
}

func (v SetVal) Split() (VecVal, VecVal) {
	var keys, vals = []Callable{}, []Callable{}
	for _, pair := range v.Pairs() {
		keys = append(keys, pair.Left())
		vals = append(vals, pair.Right())
	}
	return NewVector(keys...), NewVector(vals...)
}

func (v SetVal) Pairs() []PairVal {
	var pairs = []PairVal{}
	for _, field := range v().Fields() {
		pairs = append(
			pairs,
			NewPairFromData(
				field.Left(),
				field.Right()))
	}
	return pairs
}

func (v SetVal) Keys() VecVal { k, _ := v.Split(); return k }

func (v SetVal) Data() VecVal { _, d := v.Split(); return d }

func (v SetVal) Len() int { return v().Len() }

func (v SetVal) Empty() bool {
	for _, pair := range v.Pairs() {
		if !pair.Empty() {
			return false
		}
	}
	return true
}

func (v SetVal) GetVal(praed Callable) PairVal {
	var val Callable
	var nat, ok = v().Get(praed)
	if val, ok = nat.(Callable); !ok {
		val = NewFromData(val)
	}
	return NewPair(praed, val)
}

func (v SetVal) SetVal(key, value Callable) Associative {
	var m = v()
	m.Set(key, value)
	return SetVal(func(pairs ...PairVal) d.Mapped { return m })
}

func (v SetVal) Slice() []Callable {
	var pairs = []Callable{}
	for _, pair := range v.Pairs() {
		pairs = append(pairs, pair)
	}
	return pairs
}

func (v SetVal) Call(f ...Callable) Callable { return v }

func (v SetVal) Eval(p ...d.Native) d.Native {
	var slice = d.DataSlice{}
	for _, pair := range v().Fields() {
		d.SliceAppend(slice, d.NewPair(pair.Left(), pair.Right()))
	}
	return slice
}

func (v SetVal) TypeFnc() TyFnc { return Set | Functor }

func (v SetVal) TypeNat() d.TyNative { return d.Set | d.Function }

func (v SetVal) KeyFncType() TyFnc {
	if v.Len() > 0 {
		return v.Pairs()[0].Left().TypeFnc()
	}
	return None
}

func (v SetVal) KeyNatType() d.TyNative {
	if v.Len() > 0 {
		return v.Pairs()[0].Left().TypeNat()
	}
	return d.Nil
}

func (v SetVal) ValFncType() TyFnc {
	if v.Len() > 0 {
		return v.Pairs()[0].Right().TypeFnc()
	}
	return None
}

func (v SetVal) ValNatType() d.TyNative {
	if v.Len() > 0 {
		return v.Pairs()[0].Right().TypeNat()
	}
	return d.Nil
}

func (v SetVal) DeCap() (Callable, Consumeable) {
	return v.Head(), v.Tail()
}

func (v SetVal) Head() Callable {
	if v.Len() > 0 {
		return v.Pairs()[0]
	}
	return nil
}

func (v SetVal) Tail() Consumeable {
	if v.Len() > 1 {
		return ConAssociativeFromPairs(v.Pairs()[1:]...)
	}
	return NewEmptyAssociative()
}

////////////////////////////////////////////////////////////////////////////////
//// MONADIC VALUES
///

///// NOOP
//
// aka void, null, nada, none, niente, zero, nan, rien de vas plus, or whatever
// else you like to call the abscence of a value
func NewNoOp() NoOp                      { return func() {} }
func (n NoOp) Ident() Callable           { return n }
func (n NoOp) Maybe() bool               { return false }
func (n NoOp) Empty() bool               { return true }
func (n NoOp) Eval(...d.Native) d.Native { return nil }
func (n NoOp) Value() Callable           { return nil }
func (n NoOp) Call(...Callable) Callable { return nil }
func (n NoOp) String() string            { return "⊥" }
func (n NoOp) Len() int                  { return 0 }
func (n NoOp) TypeFnc() TyFnc            { return None }
func (n NoOp) TypeNat() d.TyNative       { return d.Nil }

//// TRUTH
//
// there is exactly one truth & the abscence there of, resulting in two
// possible variants of truth values
func NewTruth(truth bool) TruthVal {
	if truth {
		return func() bool { return true }
	}
	return func() bool { return false }
}

func (t TruthVal) Eval(...d.Native) d.Native { return d.BoolVal(t()) }
func (t TruthVal) Call(...Callable) Callable { return t }
func (t TruthVal) Ident() Callable           { return t }
func (t TruthVal) TypeNat() d.TyNative       { return d.Bool }
func (t TruthVal) TypeFnc() TyFnc {
	if t() {
		return True
	}
	return False
}

func (t TruthVal) String() string {
	if t() {
		return "True"
	}
	return "False"
}
