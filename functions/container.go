/*
  FUNCTIONAL CONTAINERS

  containers implement enumeration of functional types, aka lists, vectors sets, pairs, tuples‥.
*/
package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

type (
	//// DATA
	NativeVal func(args ...interface{}) Callable
	DataVal   func(args ...d.Native) d.Native

	//// EXPRESSION
	ConstantExpr func() Callable
	UnaryExpr    func(Callable) Callable
	BinaryExpr   func(a, b Callable) Callable
	NaryExpr     func(...Callable) Callable

	//// COLLECTION
	PairVal   func(...Callable) (Callable, Callable)
	AssocPair func(...Callable) (string, Callable)
	IndexPair func(...Callable) (int, Callable)
	ListVal   func(...Callable) (Callable, ListVal)
	VecVal    func(...Callable) []Callable
	TupleVal  func(...Callable) []Callable
	AssocVec  func(...AssocPair) []AssocPair
	SetVal    func(...AssocPair) d.Mapped
)

// reverse arguments
func RevArgs(args ...Callable) []Callable {
	var rev = []Callable{}
	for i := len(args) - 1; i > 0; i-- {
		rev = append(rev, args[i])
	}
	return rev
}

// convert native to functional values
func NatToFnc(args ...d.Native) []Callable {
	var result = []Callable{}
	for _, arg := range args {
		result = append(result, NewFromData(arg))
	}
	return result
}

// convert functional to native values
func FncToNat(args ...Callable) []d.Native {
	var result = []d.Native{}
	for _, arg := range args {
		result = append(result, arg.Eval())
	}
	return result
}

//// DATA
func New(inf ...interface{}) Callable { return NewFromData(d.New(inf...)) }

func NewFromData(data ...d.Native) DataVal {
	var eval func(...d.Native) d.Native
	for _, val := range data {
		eval = val.Eval
	}
	return func(args ...d.Native) d.Native { return eval(args...) }
}

func NewDataVal() DataVal {
	return DataVal(func(args ...d.Native) d.Native {
		if len(args) > 1 {
			return d.NewSlice(args...)
		}
		if len(args) > 0 {
			return args[0]
		}
		return d.NilVal{}
	})
}
func (n DataVal) Eval(args ...d.Native) d.Native { return n().Eval(args...) }

func (n DataVal) Call(vals ...Callable) Callable {
	var results = NewVector()
	for _, val := range vals {
		// evaluate arguments to yield contained natives
		results = ConsVector(
			results,
			DataVal(func(arguments ...d.Native) d.Native {
				return val.Eval(arguments...)
			}),
		)
	}
	return results
}

func (n DataVal) TypeFnc() TyFnc   { return Data }
func (n DataVal) TypeNat() d.TyNat { return n().TypeNat() }
func (n DataVal) String() string   { return n().String() }

func NewNativeVal() NativeVal {
	return func(args ...interface{}) Callable {
		if len(args) > 0 {
			return New(args...)
		}
		return NewNone()
	}
}
func (n NativeVal) String() string   { return n().String() }
func (n NativeVal) TypeNat() d.TyNat { return n().TypeNat() }
func (n NativeVal) TypeFnc() TyFnc   { return Native }
func (n NativeVal) Call(args ...Callable) Callable {
	return NewFromData(n()).Call(args...)
}
func (n NativeVal) Eval(args ...d.Native) d.Native {
	return n().Eval(args...)
}

//// STATIC EXPRESSIONS
///
// CONSTANT EXPRESSION
func NewConstant(fnc Callable) Callable          { return fnc }
func (c ConstantExpr) Ident() Callable           { return c() }
func (c ConstantExpr) TypeFnc() TyFnc            { return Expression }
func (c ConstantExpr) TypeNat() d.TyNat          { return c().TypeNat() }
func (c ConstantExpr) Call(...Callable) Callable { return c() }
func (c ConstantExpr) Eval(...d.Native) d.Native { return c().Eval() }

/// UNARY EXPRESSION
func NewUnaryExpr(fnc func(Callable) Callable) UnaryExpr { return fnc }
func (u UnaryExpr) Ident() Callable                      { return u }
func (u UnaryExpr) TypeFnc() TyFnc                       { return Expression }
func (u UnaryExpr) TypeNat() d.TyNat                     { return d.Expression.TypeNat() }
func (u UnaryExpr) Call(arg ...Callable) Callable        { return u(arg[0]) }
func (u UnaryExpr) Eval(arg ...d.Native) d.Native        { return u(NewFromData(arg...)) }

/// BINARY EXPRESSION
func NewBinaryExpr(fnc func(l, r Callable) Callable) BinaryExpr {
	return func(left, right Callable) Callable { return fnc(left, right) }
}

func (b BinaryExpr) Ident() Callable                { return b }
func (b BinaryExpr) TypeFnc() TyFnc                 { return Expression }
func (b BinaryExpr) TypeNat() d.TyNat               { return d.Expression.TypeNat() }
func (b BinaryExpr) Call(args ...Callable) Callable { return b(args[0], args[1]) }
func (b BinaryExpr) Eval(args ...d.Native) d.Native {
	return b(NewFromData(args[0]), NewFromData(args[1]))
}

/// NARY EXPRESSION
func NewNaryExpr(fnc func(...Callable) Callable) NaryExpr { return fnc }
func (n NaryExpr) Ident() Callable                        { return n }
func (n NaryExpr) TypeFnc() TyFnc                         { return Expression }
func (n NaryExpr) TypeNat() d.TyNat                       { return d.Expression.TypeNat() }
func (n NaryExpr) Call(d ...Callable) Callable            { return n(d...) }
func (n NaryExpr) Eval(args ...d.Native) d.Native {
	var params = []Callable{}
	for _, arg := range args {
		params = append(params, NewFromData(arg))
	}
	return n(params...)
}

/// PAIRS OF VALUES
func NewEmptyPair() PairVal {
	return func(args ...Callable) (a, b Callable) {
		if len(args) > 0 {
			if len(args) > 1 {
				return args[0], args[1]
			}
			return args[0], NewNone()
		}
		return NewNone(), NewNone()
	}
}

func NewPair(l, r Callable) PairVal {
	return func(args ...Callable) (Callable, Callable) {
		if len(args) > 0 {
			if len(args) > 1 {
				return args[0], args[1]
			}
			return args[0], r
		}
		return l, r
	}
}

func NewPairFromData(l, r d.Native) PairVal {
	return func(args ...Callable) (Callable, Callable) {
		if len(args) > 0 {
			if len(args) > 1 {
				// return pointers to natives eval functions
				return DataVal(args[0].Eval), DataVal(args[1].Eval)
			}

			return DataVal(args[0].Eval), NewNone()
		}

		return DataVal(l.Eval), DataVal(r.Eval)
	}
}

func NewPairFromLiteral(l, r interface{}) PairVal {

	return func(args ...Callable) (Callable, Callable) {

		if len(args) > 0 {

			if len(args) > 1 {

				// return values eval methods as continuations
				return DataVal(
						d.New(args[0]).Eval,
					),
					DataVal(
						d.New(args[1]).Eval,
					)
			}

			return DataVal(d.New(args[0]).Eval), NewNone()
		}

		return DataVal(d.New(l).Eval), DataVal(d.New(r).Eval)
	}
}

func (p PairVal) Ident() Callable { return p }
func (p PairVal) Pair() Paired    { return p }

// construct value pairs from any consumeable assuming keys and values alter
func ConsPair(list Consumeable) (PairVal, Consumeable) {

	var first, tail = list.DeCap()

	if first != nil {

		var second Callable
		second, tail = tail.DeCap()

		if tail != nil {
			// walk list generate a pair every second step
			// recursively.
			return NewPair(first, second), tail
		}
		// if number of elements in list is not dividable by two, last
		// element will contain an empty list as its right element
		return NewPair(first, tail), nil
	}
	// argument consumeable vanished, return nil for left and right
	return nil, nil
}

// implement consumeable
func (p PairVal) DeCap() (Callable, Consumeable) { l, r := p(); return l, NewList(r) }
func (p PairVal) Head() Callable                 { l, _ := p(); return l }
func (p PairVal) Tail() Consumeable              { _, r := p(); return NewPair(r, NewNone()) }

// implement swappable
func (p PairVal) Swap() (Callable, Callable) { l, r := p(); return r, l }
func (p PairVal) SwappedPair() PairVal       { return NewPair(p.Right(), p.Left()) }

// implement associated
func (p PairVal) Left() Callable             { l, _ := p(); return l }
func (p PairVal) Right() Callable            { _, r := p(); return r }
func (p PairVal) Both() (Callable, Callable) { return p() }

// implement sliced
func (p PairVal) Slice() []Callable { return []Callable{p.Left(), p.Right()} }

// associative implementing element access
func (p PairVal) Key() Callable   { return p.Left() }
func (p PairVal) Value() Callable { return p.Right() }

// key and values native and functional types
func (p PairVal) KeyType() TyFnc        { return p.Left().TypeFnc() }
func (p PairVal) KeyNatType() d.TyNat   { return p.Left().TypeNat() }
func (p PairVal) ValueType() TyFnc      { return p.Right().TypeFnc() }
func (p PairVal) ValueNatType() d.TyNat { return p.Right().TypeNat() }

// slightly different element types, since right value is a list now
func (p PairVal) HeadType() TyFnc { return p.Left().TypeFnc() }
func (p PairVal) TailType() TyFnc { return p.Right().TypeFnc() }

// composed functional type of a value pair
func (p PairVal) TypeFnc() TyFnc {
	return Pair | p.Left().TypeFnc() | p.Right().TypeFnc()
}

// composed native type of a value pair
func (p PairVal) TypeNat() d.TyNat {
	return p.Left().TypeNat() | p.Right().TypeNat()
}

// implements compose
func (p PairVal) Empty() bool {
	if (p.Left() == nil ||
		!p.Left().TypeFnc().Flag().Match(None) &&
			!p.Left().TypeNat().Flag().Match(d.Nil)) &&
		(p.Right() == nil ||
			!p.Right().TypeFnc().Flag().Match(None) &&
				!p.Right().TypeNat().Flag().Match(d.Nil)) {
		return true
	}
	return false
}

// call arguments are forwarded to the contained sub elements
func (p PairVal) Call(args ...Callable) Callable {
	return NewPair(p.Left().Call(args...), p.Right().Call(args...))
}

// evaluation arguments are forwarded to the contained sub elements
func (p PairVal) Eval(args ...d.Native) d.Native {
	return d.NewPair(p.Left().Eval(args...), p.Right().Eval(args...))
}

//// ASSOCIATIVE PAIRS
///
// pair composed of a string key and a functional value
func NewAssocPair(key string, val Callable) AssocPair {
	return func(...Callable) (string, Callable) { return key, val }
}

func (a AssocPair) KeyStr() string {
	var key, _ = a()
	return key
}

func (a AssocPair) Ident() Callable { return a }

func (a AssocPair) Pair() Paired    { return NewPair(a.Both()) }
func (a AssocPair) Pairs() []Paired { return []Paired{NewPair(a.Both())} }

func (a AssocPair) Empty() bool {
	if a.Left() != nil && a.Right() != nil {
		return false
	}
	return true
}
func (a AssocPair) Both() (Callable, Callable) {
	var key, val = a()
	return NewFromData(d.StrVal(key)), val
}

func (a AssocPair) GetVal(Callable) Callable { return a.Right() }
func (a AssocPair) SetVal(key, val Callable) Associative {
	return NewAssocPair(a.Left().String(), a.Right())
}
func (a AssocPair) Left() Callable {
	key, _ := a()
	return NewFromData(d.StrVal(key))
}

func (a AssocPair) Right() Callable {
	_, val := a()
	return val
}
func (a AssocPair) Key() Callable                  { return a.Left() }
func (a AssocPair) Value() Callable                { return a.Right() }
func (a AssocPair) Call(args ...Callable) Callable { return a.Right().Call(args...) }
func (a AssocPair) Eval(args ...d.Native) d.Native { return a.Right().Eval(args...) }

func (a AssocPair) KeyType() TyFnc        { return Pair }
func (a AssocPair) KeyNatType() d.TyNat   { return d.String }
func (a AssocPair) ValFncType() TyFnc     { return a.Right().TypeFnc() }
func (a AssocPair) ValNatType() d.TyNat   { return a.Right().TypeNat() }
func (a AssocPair) KeyFncType() TyFnc     { return a.Left().TypeFnc() }
func (a AssocPair) ValueType() TyFnc      { return a.Right().TypeFnc() }
func (a AssocPair) ValueNatType() d.TyNat { return a.Right().TypeNat() }

func (a AssocPair) TypeFnc() TyFnc   { return Pair | a.ValueType() }
func (a AssocPair) TypeNat() d.TyNat { return d.Pair | d.String | a.ValueNatType() }

/// pair composed of an integer and a functional value
func NewIndexPair(idx int, val Callable) IndexPair {
	return func(...Callable) (int, Callable) { return idx, val }
}

func (a IndexPair) Index() int {
	idx, _ := a()
	return idx
}

func (a IndexPair) Ident() Callable { return a }

func (a IndexPair) Both() (Callable, Callable) {
	var idx, val = a()
	return NewFromData(d.IntVal(idx)), val
}

func (a IndexPair) Pair() Paired { return a }

func (a IndexPair) Left() Callable {
	idx, _ := a()
	return NewFromData(d.IntVal(idx))
}

func (a IndexPair) Right() Callable {
	_, val := a()
	return val
}

func (a IndexPair) Key() Callable   { return a.Left() }
func (a IndexPair) Value() Callable { return a.Right() }

func (a IndexPair) Call(args ...Callable) Callable { return a.Right().Call(args...) }
func (a IndexPair) Eval(args ...d.Native) d.Native { return a.Right().Eval(args...) }

func (a IndexPair) KeyType() TyFnc        { return Pair }
func (a IndexPair) KeyNatType() d.TyNat   { return d.Int }
func (a IndexPair) ValueType() TyFnc      { return a.Right().TypeFnc() }
func (a IndexPair) ValueNatType() d.TyNat { return a.Right().TypeNat() }

func (a IndexPair) TypeFnc() TyFnc   { return Pair | a.ValueType() }
func (a IndexPair) TypeNat() d.TyNat { return d.Pair | d.Int | a.ValueNatType() }

///////////////////////////////////////////////////////////////////////////////
//// RECURSIVE LIST OF VALUES
///
// base implementation of recursively linked lists
func ConcatLists(a, b ListVal) ListVal {
	return ListVal(func(args ...Callable) (Callable, ListVal) {
		if len(args) > 0 {
			b = b.Cons(args...)
		}
		var head Callable
		if head, a = a(); head != nil {
			return head, ConcatLists(a, b)
		}
		return b()
	})
}
func NewList(elems ...Callable) ListVal {
	return func(args ...Callable) (Callable, ListVal) {
		if len(args) > 0 {
			elems = append(elems, args...)
		}
		if len(elems) > 0 {
			var head = elems[0]
			if len(elems) > 1 {
				return head, NewList(
					elems[1:]...,
				)
			}
			return head, NewList()
		}
		return nil, NewList()
	}
}
func (l ListVal) Cons(elems ...Callable) ListVal {
	return ListVal(func(args ...Callable) (Callable, ListVal) {
		return l(append(elems, args...)...)
	})
}
func (l ListVal) Push(elems ...Callable) ListVal {
	return ConcatLists(NewList(elems...), l)
}

func (l ListVal) Call(d ...Callable) Callable {
	var head Callable
	head, l = l(d...)
	return head
}

// eval applys current heads eval method to passed arguments, or calle it empty
func (l ListVal) Eval(args ...d.Native) d.Native {
	return l.Head().Eval(args...)
}

func (l ListVal) Empty() bool {
	if l.Head() != nil {
		if !None.Flag().Match(l.Head().TypeFnc()) ||
			!d.Nil.Flag().Match(l.Head().TypeNat()) {
			return false
		}
	}

	return true
}

// to determine the length of a recursive function, it has to be fully unwound,
// so use with care! (and ask yourself, what went wrong to make the length of a
// list be of importance)
func (l ListVal) Len() int {
	var length int
	var head, tail = l()
	if head != nil {
		length += 1 + tail.Len()
	}
	return length
}

func (l ListVal) Ident() Callable                { return l }
func (l ListVal) Null() ListVal                  { return NewList() }
func (l ListVal) Tail() Consumeable              { _, t := l(); return t }
func (l ListVal) Head() Callable                 { h, _ := l(); return h }
func (l ListVal) DeCap() (Callable, Consumeable) { return l() }
func (l ListVal) TypeFnc() TyFnc                 { return List | Functor }
func (l ListVal) TypeNat() d.TyNat               { return l.Head().TypeNat() }

//////////////////////////////////////////////////////////////////////////////////////////
//// VECTORS (SLICES) OF VALUES
func NewEmptyVector(init ...Callable) VecVal { return NewVector() }

func NewVector(init ...Callable) VecVal {

	var vector = init

	return func(args ...Callable) []Callable {

		if len(args) > 0 {

			// append args to vector
			vector = append(
				vector,
				args...,
			)
		}

		// return slice vector
		return vector
	}
}

func ConsVector(vec Vectorized, args ...Callable) VecVal {

	return ConsVectorFromCallable(append(RevArgs(args...), vec.Slice()...)...)
}

func AppendVector(vec Vectorized, args ...Callable) VecVal {

	return ConsVectorFromCallable(append(vec.Slice(), args...)...)

}

func ConsVectorFromCallable(init ...Callable) VecVal {

	return func(args ...Callable) []Callable {

		return RevArgs(append(args, init...)...)
	}
}

func AppendVecFromCallable(init ...Callable) VecVal {

	return func(args ...Callable) []Callable {

		return append(init, args...)
	}
}

func (v VecVal) Ident() Callable { return v }

func (v VecVal) Call(d ...Callable) Callable { return NewVector(v(d...)...) }

func (v VecVal) Eval(args ...d.Native) d.Native {

	var result = []d.Native{}

	for _, arg := range args {
		result = append(result, arg)
	}

	return d.DataSlice(result)
}

func (v VecVal) TypeFnc() TyFnc {
	if len(v()) > 0 {
		return Vector | v.Head().TypeFnc()
	}
	return Vector | None
}

func (v VecVal) TypeNat() d.TyNat {
	if len(v()) > 0 {
		return d.Slice.TypeNat() | v.Head().TypeNat()
	}
	return d.Slice.TypeNat() | d.Nil
}

func (v VecVal) Head() Callable {
	if v.Len() > 0 {
		return v.Slice()[0]
	}
	return nil
}

func (v VecVal) Tail() Consumeable {
	if v.Len() > 1 {
		return NewVector(v.Slice()[1:]...)
	}
	return NewEmptyVector()
}

func (v VecVal) DeCap() (Callable, Consumeable) {
	return v.Head(), v.Tail()
}

func (v VecVal) Empty() bool {

	if len(v()) > 0 {

		for _, val := range v() {

			if !val.TypeNat().Flag().Match(d.Nil) &&
				!val.TypeFnc().Flag().Match(None) {

				return false
			}
		}
	}
	return true
}

func (v VecVal) Len() int          { return len(v()) }
func (v VecVal) Vector() VecVal    { return v }
func (v VecVal) Slice() []Callable { return v() }

func (v VecVal) Append(args ...Callable) VecVal {
	return NewVector(append(v(), args...)...)
}

func (v VecVal) Cons(args ...Callable) VecVal {
	return NewVector(append(RevArgs(args...), RevArgs(v()...)...)...)
}

func (v VecVal) Get(i int) Callable {
	if i < v.Len() {
		return v()[i]
	}
	return NewNone()
}

func (v VecVal) Set(i int, val Callable) Vectorized {
	if i < v.Len() {
		var slice = v()
		slice[i] = val
		return VecVal(func(elems ...Callable) []Callable { return slice })

	}
	return v
}

func (v VecVal) Sort(flag d.TyNat) {
	var ps = newDataSorter(v()...)
	ps.Sort(flag)
	v = NewVector(ps...)
}

func (v VecVal) Search(praed Callable) int { return newDataSorter(v()...).Search(praed) }

//// ASSOCIATIVE SLICE OF VALUE PAIRS
///
// list of associative pairs in sequential order associated, sorted and
// searched by left value of the pairs
func ConsAssociative(vec Associative, pfnc ...Paired) AssocVec {
	return NewAssociativeFromPair(append(vec.Pairs(), pfnc...)...)
}

func NewAssociativeFromPair(ps ...Paired) AssocVec {
	var pairs = []AssocPair{}
	for _, arg := range ps {
		if pair, ok := arg.(AssocPair); ok {
			pairs = append(pairs, pair)
		}
	}
	return AssocVec(func(pairs ...AssocPair) []AssocPair { return pairs })
}

func ConsAssociativeFromPairs(pp ...Paired) AssocVec {
	var pairs = []AssocPair{}
	for _, pair := range pp {
		if assoc, ok := pair.(AssocPair); ok {
			pairs = append(pairs, assoc)
		}
	}
	return AssocVec(func(pairs ...AssocPair) []AssocPair { return pairs })
}

func NewEmptyAssociative() AssocVec {
	return AssocVec(func(pairs ...AssocPair) []AssocPair { return []AssocPair{} })
}

func NewAssociative(pp ...AssocPair) AssocVec {

	return func(pairs ...AssocPair) []AssocPair {
		for _, pair := range pp {
			pairs = append(pairs, pair)
		}
		return pairs
	}
}

func (v AssocVec) Call(args ...Callable) Callable {
	return v.Cons(args...)
}

func (v AssocVec) Cons(p ...Callable) AssocVec {

	var pairs = v.Pairs()

	return ConsAssociativeFromPairs(pairs...)
}

func (v AssocVec) DeCap() (Callable, Consumeable) {
	return v.Head(), v.Tail()
}

func (v AssocVec) Empty() bool {

	if len(v()) > 0 {

		for _, pair := range v() {

			if !pair.Empty() {

				return false
			}
		}
	}
	return true
}

func (v AssocVec) KeyFncType() TyFnc {
	if v.Len() > 0 {
		return v.Pairs()[0].Left().TypeFnc()
	}
	return None
}

func (v AssocVec) KeyNatType() d.TyNat {
	if v.Len() > 0 {
		return v.Pairs()[0].Left().TypeNat()
	}
	return d.Nil
}

func (v AssocVec) ValFncType() TyFnc {
	if v.Len() > 0 {
		return v.Pairs()[0].Right().TypeFnc()
	}
	return None
}

func (v AssocVec) ValNatType() d.TyNat {
	if v.Len() > 0 {
		return v.Pairs()[0].Right().TypeNat()
	}
	return d.Nil
}

func (v AssocVec) TypeFnc() TyFnc { return Record | Functor }

func (v AssocVec) TypeNat() d.TyNat {
	if len(v()) > 0 {
		return d.Slice | v.Head().TypeNat()
	}
	return d.Slice | d.Nil.TypeNat()
}

func (v AssocVec) Len() int { return len(v()) }

func (v AssocVec) Sort(flag d.TyNat) {
	var ps = newPairSorter(v.Pairs()...)
	ps.Sort(flag)
	v = NewAssociativeFromPair(ps...)
}

func (v AssocVec) Search(praed Callable) int {
	return newPairSorter(v.Pairs()...).Search(praed)
}

func (v AssocVec) Get(idx int) AssocPair {
	if idx < v.Len()-1 {
		return v()[idx]
	}
	return NewAssocPair("None", NewNone())
}

func (v AssocVec) GetVal(praed Callable) Callable {
	return NewAssociativeFromPair(newPairSorter(v.Pairs()...).Get(praed))
}

func (v AssocVec) Range(praed Callable) []Paired {
	return newPairSorter(v.Pairs()...).Range(praed)
}

func (v AssocVec) Pairs() []Paired {
	var pairs = []Paired{}
	for _, pair := range v() {
		pairs = append(pairs, pair)
	}
	return pairs
}

func (v AssocVec) DeCapPairWise() (AssocPair, []AssocPair) {
	var pairs = v()
	if len(pairs) > 0 {
		if len(pairs) > 1 {
			return pairs[0], pairs[1:]
		}
		return pairs[0], []AssocPair{}
	}
	return nil, []AssocPair{}
}

func (v AssocVec) SwitchedPairs() []Paired {
	var switched = []Paired{}
	for _, pair := range v() {
		switched = append(
			switched,
			pair,
		)
	}
	return switched
}

func (v AssocVec) SetVal(key, value Callable) Associative {
	if idx := v.Search(key); idx >= 0 {
		var pairs = v()
		pairs[idx] = NewAssocPair(key.String(), value)
		return NewAssociative(pairs...)
	}
	return NewAssociative(append(v(), NewAssocPair(key.String(), value))...)
}

func (v AssocVec) Slice() []Callable {
	var fncs = []Callable{}
	for _, pair := range v() {
		fncs = append(fncs, NewPair(pair.Left(), pair.Right()))
	}
	return fncs
}

func (v AssocVec) Head() Callable {
	if v.Len() > 0 {
		return v.Pairs()[0]
	}
	return nil
}

func (v AssocVec) Tail() Consumeable {
	if v.Len() > 1 {
		return ConsAssociativeFromPairs(v.Pairs()[1:]...)
	}
	return NewEmptyAssociative()
}

func (v AssocVec) Eval(p ...d.Native) d.Native {
	var slice = d.DataSlice{}
	for _, pair := range v() {
		d.SliceAppend(slice, d.NewPair(pair.Left(), pair.Right()))
	}
	return slice
}

///////////////////////////////////////////////////////////////////////////////
//// ASSOCIATIVE SET (HASH MAP OF VALUES)
///
// unordered associative set of key/value pairs that can be sorted, accessed
// and searched by the left (key) value of the pair
func ConsAssocSet(pairs ...Paired) SetVal {
	var paired = []Paired{}
	for _, pair := range pairs {
		paired = append(paired, pair)
	}
	return NewAssocSet(paired...)
}

func NewAssocSet(pairs ...Paired) SetVal {

	var kt d.TyNat
	var set d.Mapped

	// OR concat all accessor types
	for _, pair := range pairs {
		kt = kt | pair.Left().TypeNat()
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
	return SetVal(func(pairs ...AssocPair) d.Mapped { return set })
}

func (v SetVal) Split() (VecVal, VecVal) {
	var keys, vals = []Callable{}, []Callable{}
	for _, pair := range v.Pairs() {
		keys = append(keys, pair.Left())
		vals = append(vals, pair.Right())
	}
	return NewVector(keys...), NewVector(vals...)
}

func (v SetVal) Pairs() []Paired {
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

func (v SetVal) Keys() VecVal { k, _ := v.Split(); return k }

func (v SetVal) Data() VecVal { _, d := v.Split(); return d }

func (v SetVal) Len() int { return v().Len() }

func (v SetVal) Empty() bool {
	for _, pair := range v.Pairs() {
		if pair.Left() != nil && pair.Right() != nil {
			return false
		}
	}
	return true
}

func (v SetVal) GetVal(praed Callable) Callable {
	var val Callable
	var nat, ok = v().Get(praed)
	if val, ok = nat.(Callable); !ok {
		val = NewFromData(val)
	}
	return NewAssocPair(praed.String(), val)
}

func (v SetVal) SetVal(key, value Callable) Associative {
	var m = v()
	m.Set(key, value)
	return SetVal(func(pairs ...AssocPair) d.Mapped { return m })
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

func (v SetVal) TypeNat() d.TyNat { return d.Map | d.Expression }

func (v SetVal) KeyFncType() TyFnc {
	if v.Len() > 0 {
		return v.Pairs()[0].Left().TypeFnc()
	}
	return None
}

func (v SetVal) KeyNatType() d.TyNat {
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

func (v SetVal) ValNatType() d.TyNat {
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
		return ConsAssociativeFromPairs(v.Pairs()[1:]...)
	}
	return NewEmptyAssociative()
}

/////////////////////////////////////////////////////////////////////////////////////
//// TUPLE TYPE VALUES
///
// tuples are sequences of values grouped in a distinct sequence of distinct types,
func NewTuple(data ...Callable) TupleVal {

	return TupleVal(func(args ...Callable) []Callable {
		return data
	})
}

func (t TupleVal) Ident() Callable { return t }
func (t TupleVal) Len() int        { return len(t()) }

// pairs prepends annotates member values as pair values carrying this
// instances sub-type signature and tuple position in in the second field
func (t TupleVal) Pairs() []Paired {
	var pairs = []Paired{}
	for _, arg := range t() {
		pairs = append(
			pairs,
			NewPair(NewFromData(d.NewPair(
				arg.TypeNat(),
				arg.TypeFnc(),
			)),
				arg,
			))
	}
	return pairs
}

// implement consumeable
func (t TupleVal) DeCap() (Callable, Consumeable) {
	var list = NewList(t()...)
	return list()
}

func (t TupleVal) Head() Callable    { head, _ := t.DeCap(); return head }
func (t TupleVal) Tail() Consumeable { _, tail := t.DeCap(); return tail }

// functional type concatenates the functional types of all the subtypes
func (t TupleVal) TypeFnc() TyFnc {
	var ftype = TyFnc(0)
	for _, typ := range t() {
		ftype = ftype | typ.TypeFnc()
	}
	return ftype
}

// native type concatenates the native types of all the subtypes
func (t TupleVal) TypeNat() d.TyNat {
	var ntype = d.Slice
	for _, typ := range t() {
		ntype = ntype | typ.TypeNat()
	}
	return ntype
}

// string representation of a tuple generates one row per sub type by
// concatenating each sub types native type, functional type and value.
func (t TupleVal) String() string { return t.Head().String() }

func (t TupleVal) Eval(args ...d.Native) d.Native {
	var result = []d.Native{}
	for _, val := range t() {
		result = append(result, val.Eval(val))
	}
	return d.DataSlice(result)
}

func (t TupleVal) Call(args ...Callable) Callable {
	var result []Callable
	for _, val := range t() {
		result = append(result, val.Call(args...))
	}
	return NewVector(result...)
}

func (t TupleVal) ApplyPartial(args ...Callable) TupleVal {
	return NewTuple(partialApplyTuple(t, args...)()...)
}

func partialApplyTuple(tuple TupleVal, args ...Callable) TupleVal {
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
	return TupleVal(
		func(...Callable) []Callable {
			return result
		})
}
