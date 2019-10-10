package functions

import (
	"sort"

	d "github.com/joergreinhardt/gatwd/data"
)

type (
	//// VALUE PAIRS
	ValPair   func(...Expression) (Expression, Expression)
	KeyPair   func(...Expression) (Expression, string)
	TypePair  func(...Expression) (Expression, Typed)
	IndexPair func(...Expression) (Expression, int)

	//// ENUMERABLE
	EnumType func(d.Integer) (EnumVal, d.Typed, d.Typed)
	EnumVal  func(...Expression) (Expression, d.Integer, EnumType)

	//// COLLECTIONS
	VecVal   func(...Expression) []Expression
	ListVal  func(...Expression) (Expression, ListVal)
	PairVec  func(...Paired) []Paired
	PairList func(...Paired) (Paired, PairList)
	MapVal   func(...Expression) (Expression, map[string]Expression)
)

///////////////////////////////////////////////////////////////////////////////
//// PAIRS OF VALUES
///
// pairs can be created empty, key & value may be constructed later
func NewEmptyPair() ValPair {
	return func(args ...Expression) (a, b Expression) {
		if len(args) > 0 {
			if len(args) > 1 {
				return args[0], args[1]
			}
			return args[0], NewNone()
		}
		return NewNone(), NewNone()
	}
}

// new pair from two callable instances
func NewPair(l, r Expression) ValPair {
	return func(args ...Expression) (Expression, Expression) {
		if len(args) > 0 {
			if len(args) > 1 {
				return args[0], args[1]
			}
			return args[0], r
		}
		return l, r
	}
}
func (p ValPair) Pair() Paired                   { return p }
func (p ValPair) Both() (Expression, Expression) { return p() }
func (p ValPair) Left() Expression               { l, _ := p(); return l }
func (p ValPair) Right() Expression              { _, r := p(); return r }
func (p ValPair) Swap() (Expression, Expression) { l, r := p(); return r, l }
func (p ValPair) SwappedPair() Paired            { return NewPair(p.Right(), p.Left()) }
func (p ValPair) Slice() []Expression            { return []Expression{p.Left(), p.Right()} }
func (p ValPair) Key() Expression                { return p.Left() }
func (p ValPair) Value() Expression              { return p.Right() }
func (p ValPair) TypeFnc() TyFnc                 { return Pair }
func (p ValPair) TypeElem() TyPattern {
	if p.Right() != nil {
		return p.Right().Type()
	}
	return Def(None, Pair, None)
}
func (p ValPair) TypeKey() d.Typed {
	if p.Left() != nil {
		return p.Left().Type()
	}
	return None
}
func (p ValPair) TypeValue() d.Typed {
	if p.Right() != nil {
		return p.Right().Type()
	}
	return None
}
func (p ValPair) Type() TyPattern {
	if p.TypeKey().Match(None) && p.TypeValue().Match(None) {
		return Def(Pair, None)
	}
	return Def(Pair, Def(p.TypeKey(), p.TypeValue()))
}

func (p ValPair) Empty() bool {
	if p.Left() == nil || (!p.Left().TypeFnc().Flag().Match(None) &&
		(p.Right() == nil || (!p.Right().TypeFnc().Flag().Match(None)))) {
		return true
	}
	return false
}

func (p ValPair) Call(args ...Expression) Expression {
	return NewPair(p.Key(), p.Value().Call(args...))
}

///////////////////////////////////////////////////////////////////////////////
//// ASSOCIATIVE PAIRS
///
// pair composed of a string key and a functional value
func NewKeyPair(key string, val Expression) KeyPair {
	return func(...Expression) (Expression, string) { return val, key }
}

func (a KeyPair) KeyStr() string                     { _, key := a(); return key }
func (a KeyPair) Value() Expression                  { val, _ := a(); return val }
func (a KeyPair) Left() Expression                   { return a.Value() }
func (a KeyPair) Right() Expression                  { return DecData(d.StrVal(a.KeyStr())) }
func (a KeyPair) Both() (Expression, Expression)     { return a.Left(), a.Right() }
func (a KeyPair) Pair() Paired                       { return NewPair(a.Both()) }
func (a KeyPair) Pairs() []Paired                    { return []Paired{NewPair(a.Both())} }
func (a KeyPair) Key() Expression                    { return a.Right() }
func (a KeyPair) Call(args ...Expression) Expression { return a.Value().Call(args...) }
func (a KeyPair) TypeValue() d.Typed                 { return a.Value().Type() }
func (a KeyPair) TypeKey() d.Typed                   { return Key }
func (a KeyPair) TypeFnc() TyFnc                     { return Key | Pair }
func (p KeyPair) Type() TyPattern {
	if p.TypeKey().Match(None) && p.TypeValue().Match(None) {
		return Def(Key|Pair, None)
	}
	return Def(Key|Pair, Def(p.TypeKey(), p.TypeValue()))
}

// implement swappable
func (p KeyPair) Swap() (Expression, Expression) {
	l, r := p()
	return DecData(d.StrVal(r)), l
}
func (p KeyPair) SwappedPair() Paired { return NewPair(p.Right(), p.Left()) }

func (a KeyPair) Empty() bool {
	if a.Key() != nil && a.Value() != nil && a.Value().TypeFnc() != None {
		return false
	}
	return true
}

///////////////////////////////////////////////////////////////////////////////
//// INDEX PAIR
///
// pair composed of an integer and a functional value
func NewIndexPair(idx int, val Expression) IndexPair {
	return func(...Expression) (Expression, int) { return val, idx }
}
func (a IndexPair) Index() int                         { _, idx := a(); return idx }
func (a IndexPair) Value() Expression                  { val, _ := a(); return val }
func (a IndexPair) Left() Expression                   { return a.Value() }
func (a IndexPair) Right() Expression                  { return DecData(d.IntVal(a.Index())) }
func (a IndexPair) Both() (Expression, Expression)     { return a.Left(), a.Right() }
func (a IndexPair) Pair() Paired                       { return a }
func (a IndexPair) Pairs() []Paired                    { return []Paired{NewPair(a.Both())} }
func (a IndexPair) Key() Expression                    { return a.Right() }
func (a IndexPair) Call(args ...Expression) Expression { return a.Value().Call(args...) }
func (a IndexPair) TypeFnc() TyFnc                     { return Index | Pair }
func (a IndexPair) TypeKey() d.Typed                   { return Index }
func (a IndexPair) TypeValue() d.Typed                 { return a.Value().Type() }
func (a IndexPair) Type() TyPattern {
	if a.TypeKey().Match(None) && a.TypeValue().Match(None) {
		return Def(Index|Pair, None)
	}
	return Def(Index|Pair, Def(a.TypeKey(), a.TypeValue()))
}

// implement swappable
func (p IndexPair) Swap() (Expression, Expression) {
	l, r := p()
	return DecData(d.New(r)), l
}
func (p IndexPair) SwappedPair() Paired { return NewPair(p.Right(), p.Left()) }
func (a IndexPair) Empty() bool {
	if a.Index() >= 0 && a.Value() != nil && a.Value().TypeFnc() != None {
		return true
	}
	return false
}

//// ENUM TYPE
///
//
var (
	// check argument expression to implement data.integers interface
	isInt = NewTest(func(args ...Expression) bool {
		for _, arg := range args {
			if arg.Type().Match(Data) {
				if nat, ok := args[0].(Native); ok {
					if nat.Eval().Type().Match(d.Integers) {
						continue
					}
				}
			}
			return false
		}
		return true
	})
	// creates low/high bound type argument and lesser/greater bounds
	// checks, if no bound arguments where given, they will be set to minus
	// infinity to infinity and always check out true.
	createBounds = func(bounds ...d.Integer) (low, high d.Typed, lesser, greater func(idx d.Integer) bool) {
		if len(bounds) == 0 {
			low, high = Def(Lex_Negative, Lex_Infinite), Lex_Infinite
			lesser = func(idx d.Integer) bool { return true }
			greater = func(idx d.Integer) bool { return true }
		}
		if len(bounds) > 0 {
			var minBound = bounds[0].(d.Native)
			low = DefValNative(minBound)
			// bound argument could be instance of type big int
			if minBound.Type().Match(d.BigInt) {
				lesser = func(arg d.Integer) bool {
					if minBound.(*d.BigIntVal).GoBigInt().Cmp(
						arg.(*d.BigIntVal).GoBigInt()) < 0 {
						return true
					}
					return false
				}
			} else {
				lesser = func(arg d.Integer) bool {
					if minBound.(d.Integer).Int() >
						arg.(Native).Eval().(d.Integer).Int() {
						return true
					}
					return false
				}
			}
		}

		if len(bounds) > 1 {
			var maxBound = bounds[1].(d.Native)
			high = DefValNative(maxBound)
			if maxBound.Type().Match(d.BigInt) {
				greater = func(arg d.Integer) bool {
					if maxBound.(*d.BigIntVal).GoBigInt().Cmp(
						arg.(*d.BigIntVal).GoBigInt()) > 0 {
						return true
					}
					return false
				}
			} else {
				greater = func(arg d.Integer) bool {
					if arg.(d.Integer).Int() >
						maxBound.(d.Integer).Int() {
						return true
					}
					return false
				}
			}
		}
		return low, high, lesser, greater
	}

	inBound = func(lesser, greater func(d.Integer) bool, ints ...d.Integer) bool {
		for _, i := range ints {
			if !lesser(i) && !greater(i) {
				return true
			}
		}
		return false
	}
)

func NewEnumType(fnc func(...d.Integer) Expression, limits ...d.Integer) EnumType {
	var low, high, lesser, greater = createBounds(limits...)
	return func(idx d.Integer) (EnumVal, d.Typed, d.Typed) {
		return func(args ...Expression) (Expression, d.Integer, EnumType) {
			if inBound(lesser, greater, idx) {
				if len(args) > 0 {
					return fnc(idx).Call(args...), idx, NewEnumType(fnc, limits...)
				}
				return fnc(idx), idx, NewEnumType(fnc, limits...)
			}
			return NewNone(), idx, NewEnumType(fnc, limits...)
		}, low, high
	}
}
func (e EnumType) Expr() Expression {
	var expr, _, _ = e(d.IntVal(0))
	return expr
}
func (e EnumType) Limits() (min, max d.Typed) {
	_, min, max = e(d.IntVal(0))
	return min, max
}
func (e EnumType) Low() d.Typed {
	var min, _ = e.Limits()
	return min
}
func (e EnumType) High() d.Typed {
	var _, max = e.Limits()
	return max
}
func (e EnumType) InBound(ints ...d.Integer) bool {
	var _, _, lesser, greater = createBounds(
		e.Low().(d.Integer),
		e.High().(d.Integer),
	)
	return inBound(lesser, greater, ints...)
}
func (e EnumType) Null() Expression {
	var result, _, _ = e(d.IntVal(0))
	return result
}
func (e EnumType) Unit() Expression {
	var result, _, _ = e(d.IntVal(1))
	return result
}
func (e EnumType) Type() TyPattern { return Def(Enum, e.Unit().Type()) }
func (e EnumType) TypeFnc() TyFnc  { return Enum | e.Unit().TypeFnc() }
func (e EnumType) String() string  { return e.Type().TypeName() }
func (e EnumType) Call(args ...Expression) Expression {
	if len(args) > 0 {
		return e.Expr().Call(args...)
	}
	return e.Expr().Call()
}

//// ENUM VALUE
///
//
func (e EnumVal) Expr() Expression {
	var expr, _, _ = e()
	return expr
}
func (e EnumVal) Index() d.Integer {
	var _, idx, _ = e()
	return idx
}
func (e EnumVal) EnumType() EnumType {
	var _, _, et = e()
	return et
}
func (e EnumVal) Next() EnumVal {
	var result, _, _ = e.EnumType()(e.Index().Int() + d.IntVal(1))
	return result
}
func (e EnumVal) Previous() EnumVal {
	var result, _, _ = e.EnumType()(e.Index().Int() - d.IntVal(1))
	return result
}
func (e EnumVal) String() string                     { return e.Expr().String() }
func (e EnumVal) Type() TyPattern                    { return e.EnumType().Type() }
func (e EnumVal) TypeFnc() TyFnc                     { return e.EnumType().TypeFnc() }
func (e EnumVal) Call(args ...Expression) Expression { return e.Expr().Call(args...) }

//////////////////////////////////////////////////////////////////////////////////////////
//// VECTORS (SLICES) OF VALUES
///
// helper function to reverse argument sets
func reverse(args []Expression) (rev []Expression) {
	if len(args) > 1 {
		var l = len(args)
		rev = make([]Expression, l, l)
		for i, arg := range args {
			rev[l-1-i] = arg
		}
		return rev
	}
	return args
}

// sequential vector provides random access to sequential data. appends
// arguments in the order they where passed in, at the end of slice, when
// called
func NewVector(elems ...Expression) VecVal {
	return func(args ...Expression) []Expression {
		if len(args) > 0 {
			return append(elems, args...)
		}
		return elems
	}
}

// default operation
func (v VecVal) Append(args ...Expression) Sequential {
	return NewVector(append(v(), args...)...)
}

// prepends arguments at head of list in reversed order, to emulate arguments
// added once at a time recursively.
func (v VecVal) Cons(args ...Expression) Sequential {
	return NewVector(append(reverse(args), v()...)...)
}

// appends arguments to the vector, or returns unaltered vector, when no
// arguments are passed.
func (v VecVal) Call(args ...Expression) Expression {
	if len(args) > 0 {
		return v.Append(args...)
	}
	return v
}
func (v VecVal) Slice() []Expression { return v() }
func (v VecVal) Len() int            { return len(v()) }
func (v VecVal) TypeFnc() TyFnc      { return Vector }
func (v VecVal) Type() TyPattern {
	if v.Len() > 0 {
		return Def(Vector, v.Head().Type())
	}
	return Def(Vector, None)
}
func (v VecVal) TypeElem() TyPattern {
	if v.Len() > 0 {
		return v.Head().Type()
	}
	return Def(None)
}

func (v VecVal) Head() Expression {
	if v.Len() > 0 {
		return v()[0]
	}
	return NewNone()
}

func (v VecVal) Tail() Sequential {
	if v.Len() > 1 {
		return NewVector(v()[1:]...)
	}
	return NewVector()
}

func (v VecVal) TailVec() VecVal {
	if v.Len() > 1 {
		return NewVector(v.Tail().(VecVal)()...)
	}
	return NewVector()
}

func (v VecVal) Consume() (Expression, Sequential) {
	return v.Head(), v.Tail()
}

func (v VecVal) ConsumeVec() (Expression, VecVal) {
	return v.Head(), v.TailVec()
}

func (v VecVal) First() Expression { return v.Head() }

func (v VecVal) Last() Expression {
	if v.Len() > 0 {
		return v()[v.Len()-1]
	}
	return NewNone()
}

func (v VecVal) Reverse() VecVal {
	return NewVector(reverse(v())...)
}

func (v VecVal) Empty() bool {
	if len(v()) > 0 {
		for _, val := range v() {
			if !val.TypeFnc().Flag().Match(None) {
				return false
			}
		}
	}
	return true
}

func (v VecVal) Clear() VecVal { return NewVector() }

func (v VecVal) Sequential() SequenceVal {
	return func(args ...Expression) (Expression, Sequential) {
		var head, tail = v.ConsumeVec()
		if len(args) > 0 {
			return head, NewVector(tail(args...)...).Sequential()
		}
		return head, tail.Sequential()
	}
}

func (v VecVal) Get(i int) (Expression, bool) {
	if i < v.Len() {
		return v()[i], true
	}
	return NewNone(), false
}

func (v VecVal) Set(i int, val Expression) (Vectorized, bool) {
	if i < v.Len() {
		var slice = v()
		slice[i] = val
		return VecVal(
			func(elems ...Expression) []Expression {
				return slice
			}), true

	}
	return v, false
}
func (v VecVal) Sort(less func(a, b Expression) bool) VecVal {
	var s = vecSort(func() ([]Expression, func(i, j Expression) bool) {
		return v(), less
	})
	sort.Sort(s)
	var vec, _ = s()
	return NewVector(vec...)
}
func (v VecVal) Search(
	less func(a, b Expression) bool,
	match func(arg Expression) bool,
) Expression {
	var s = vecSort(func() ([]Expression, func(i, j Expression) bool) {
		return v(), less
	})
	sort.Sort(s)
	var vec, _ = s()
	for _, elem := range vec {
		if match(elem) {
			return elem
		}
	}
	return NewNone()
}
func (v VecVal) SearchAll(
	less func(a, b Expression) bool,
	match func(arg Expression) bool,
) VecVal {
	var s = vecSort(func() ([]Expression, func(i, j Expression) bool) {
		return v(), less
	})
	sort.Sort(s)
	var vec, _ = s()
	var res = []Expression{}
	for _, elem := range vec {
		if match(elem) {
			res = append(res, elem)
		}
	}
	return NewVector(res...)
}

// vector sorter with parametrizes less method
type vecSort func() ([]Expression, func(i, j Expression) bool)

func (v vecSort) Len() int {
	var s, _ = v()
	return len(s)
}
func (v vecSort) Less(i, j int) bool {
	var s, l = v()
	return l(s[i], s[j])
}
func (v vecSort) Swap(i, j int) {
	var s, l = v()
	s[i], s[j] = s[j], s[i]
	v = func() ([]Expression, func(Expression, Expression) bool) {
		return s, l
	}
}

//// RECURSIVE LIST OF VALUES
///
// lazy implementation of recursively linked list. backed by slice. returns
// last element put in as head. prepends arguments when called to become new
// head of list, one at a time, thereby reversing argument order.
func NewList(elems ...Expression) ListVal {
	return func(args ...Expression) (Expression, ListVal) {
		if len(args) > 0 {
			elems = append(elems, args...)
		}
		var l = len(elems)
		if l > 0 {
			var head = elems[l-1]
			if l > 1 {
				return head, NewList(elems[:l-1]...)
			}
			return head, NewList()
		}
		return NewNone(), NewList()
	}
}

// default operation
func (l ListVal) Cons(elems ...Expression) Sequential {
	if len(elems) == 0 {
		return l
	}
	return ListVal(func(args ...Expression) (Expression, ListVal) {
		if len(args) > 0 {
			return l(append(elems, args...)...)
		}
		return l(elems...)
	})
}

// appends elements at the end of list in the order they where passed.
func (l ListVal) Append(elems ...Expression) Sequential {
	return ListVal(func(args ...Expression) (Expression, ListVal) {
		if len(args) > 0 {
			var head, tail = l(args...)
			if tail.Empty() {
				return head, NewList(reverse(elems)...)
			}
			return head, tail.Append(elems...).(ListVal)
		}
		var head, tail = l()
		if tail.Empty() {
			return head, NewList(reverse(elems)...)
		}
		return head, tail.Append(elems...).(ListVal)
	})
}
func (l ListVal) Head() Expression                  { h, _ := l(); return h }
func (l ListVal) Tail() Sequential                  { _, t := l(); return t }
func (l ListVal) TailList() ListVal                 { _, t := l(); return t }
func (l ListVal) Consume() (Expression, Sequential) { return l() }
func (l ListVal) ConsumeList() (Expression, ListVal) {
	return l.Head(), l.TailList()
}
func (l ListVal) TypeFnc() TyFnc { return List }
func (l ListVal) Null() ListVal  { return NewList() }
func (l ListVal) TypeElem() TyPattern {
	if l.Len() > 0 {
		return l.Head().Type()
	}
	return Def(List, None)
}

func (l ListVal) Type() TyPattern {
	if l.Len() > 0 {
		return Def(List, l.Head().Type())
	}
	return Def(List, None)
}

func (l ListVal) Slice() []Expression {
	var (
		vec        = []Expression{}
		head, tail = l()
	)
	for !head.TypeFnc().Match(None) {
		vec = append(vec, head)
		head, tail = tail()
	}
	return vec
}

func (l ListVal) Call(args ...Expression) Expression {
	if len(args) > 0 {
		return l.Cons(args...)
	}
	return l
}

func (l ListVal) Empty() bool {
	if l.Head().TypeFnc().Match(None) {
		return true
	}
	return false
}

func (l ListVal) Len() int {
	var (
		length     int
		head, tail = l()
	)
	if !head.TypeFnc().Match(None) {
		length += 1 + tail.Len()
	}
	return length
}

///////////////////////////////////////////////////////////////////////////////
//// ASSOCIATIVE SLICE OF VALUE PAIRS
///
// list of associative pairs in sequential order associated, sorted and
// searched by left value of the pairs
func NewEmptyPairVec() PairVec {
	return PairVec(func(args ...Paired) []Paired {
		var pairs = []Paired{}
		if len(args) > 0 {
			pairs = append(pairs, args...)
		}
		return pairs
	})
}

func NewPairVectorFromPairs(pairs ...Paired) PairVec {
	return PairVec(func(args ...Paired) []Paired {
		if len(args) > 0 {
			return append(pairs, args...)
		}
		return pairs
	})
}

func ConsPairVecFromPairs(rec PairVec, args ...Expression) PairVec {
	var pairs = []Paired{}
	for _, arg := range args {
		if pair, ok := arg.(Paired); ok {
			pairs = append(pairs, pair)
		}
	}
	return NewPairVectorFromPairs(append(pairs, rec()...)...)
}

func NewPairVec(args ...Paired) PairVec {
	return NewPairVectorFromPairs(args...)
}

func ConsPairVec(rec PairVec, pairs ...Paired) PairVec {
	return NewPairVectorFromPairs(append(pairs, rec()...)...)
}

func ConsPairFromArgs(pvec PairVec, args ...Expression) PairVec {
	var pairs = make([]Paired, 0, len(args))
	for _, arg := range args {
		if pair, ok := arg.(Paired); ok {
			pairs = append(pairs, pair)
		}
	}
	return PairVec(func(args ...Paired) []Paired {
		if len(args) > 0 {
			return append(append(
				args, pairs...),
				pvec()...)
		}
		return append(pairs, pvec()...)
	})
}
func (v PairVec) Len() int { return len(v()) }
func (v PairVec) Type() TyPattern {
	if v.Len() > 0 {
		return Def(Vector|Pair, v.TypeElem().TypeReturn())
	}
	return Def(None, Vector|Pair, None)
}
func (v PairVec) TypeFnc() TyFnc { return Vector }

func (v PairVec) ConsPairs(pairs ...Paired) PairVec {
	return NewPairVec(append(pairs, v()...)...)
}

func (v PairVec) Cons(elems ...Expression) Sequential {
	var pairs = make([]Paired, 0, len(elems))
	for _, elem := range elems {
		if elem.Type().Match(Pair) {
			if pair, ok := elem.(Paired); ok {
				pairs = append(pairs, pair)
			}
		}
	}
	return PairVec(func(args ...Paired) []Paired {
		if len(args) > 0 {
			return v.ConsPairs(pairs...).ConsPairs(args...)()
		}
		return v.ConsPairs(pairs...)()
	})
}

func (v PairVec) Append(args ...Expression) Sequential { return v.Cons(args...) }

func (v PairVec) Consume() (Expression, Sequential) {
	return v.Head(), v.Tail()
}

func (v PairVec) ConsumePairVec() (Paired, PairVec) {
	return v.HeadPair(), v.Tail().(PairVec)
}

func (v PairVec) Empty() bool {
	if len(v()) > 0 {
		for _, pair := range v() {
			if !pair.Empty() {
				return false
			}
		}
	}
	return true
}
func (v PairVec) TypeElem() TyPattern {
	if v.Len() > 0 {
		return Def(v.Head().TypeFnc(), Vector|Pair, v.Head().TypeFnc())
	}
	return Def(None, Vector|Pair, None)
}
func (v PairVec) TypeKey() d.Typed {
	if v.Len() > 0 {
		return v.Pairs()[0].Left().Type()
	}
	return None.TypeFnc()
}
func (v PairVec) TypeValue() d.Typed {
	if v.Len() > 0 {
		return v.Pairs()[0].Right().Type()
	}
	return None.TypeFnc()
}
func (v PairVec) Get(idx int) (Paired, bool) {
	if idx < v.Len()-1 {
		return v()[idx], true
	}
	return NewKeyPair("None", None), false
}

func (v PairVec) Pairs() []Paired {
	var pairs = []Paired{}
	for _, pair := range v() {
		pairs = append(pairs, pair)
	}
	return pairs
}

func (v PairVec) ConsumePair() (Paired, ConsumeablePaired) {
	var pairs = v()
	if len(pairs) > 0 {
		if len(pairs) > 1 {
			return pairs[0], NewPairVec(pairs[1:]...)
		}
		return pairs[0], NewPairVec()
	}
	return nil, NewPairVec()
}

func (v PairVec) SwitchedPairs() []Paired {
	var switched = []Paired{}
	for _, pair := range v() {
		switched = append(
			switched,
			pair,
		)
	}
	return switched
}

func (v PairVec) Slice() []Expression {
	var fncs = []Expression{}
	for _, pair := range v() {
		fncs = append(fncs, NewPair(pair.Left(), pair.Right()))
	}
	return fncs
}

func (v PairVec) HeadPair() Paired {
	if v.Len() > 0 {
		return v()[0].(Paired)
	}
	return NewPair(NewNone(), NewNone())
}
func (v PairVec) Head() Expression {
	if v.Len() > 0 {
		return v.Pairs()[0]
	}
	return nil
}

func (v PairVec) TailPairs() ConsumeablePaired {
	if v.Len() > 1 {
		return NewPairVectorFromPairs(v.Pairs()[1:]...)
	}
	return NewEmptyPairVec()
}
func (v PairVec) Tail() Sequential {
	if v.Len() > 1 {
		return NewPairVectorFromPairs(v.Pairs()[1:]...)
	}
	return NewEmptyPairVec()
}

func (v PairVec) Call(args ...Expression) Expression {
	return v.Cons(args...)
}

///////////////////////////////////////////////////////////////////////////////
//// LIST OF PAIRS
func ConsPairList(list PairList, pairs ...Paired) PairList {
	return list.ConsFromPairs(pairs...).(PairList)
}
func ConcatPairLists(a, b PairList) PairList {
	return PairList(func(args ...Paired) (Paired, PairList) {
		if len(args) > 0 {
			b = b.ConsFromPairs(args...).(PairList)
		}
		var pair Paired
		if pair, a = a(); pair != nil {
			return pair, ConcatPairLists(a, b)
		}
		return b()
	})
}
func NewPairList(elems ...Paired) PairList {
	return func(pairs ...Paired) (Paired, PairList) {
		if len(pairs) > 0 {
			elems = append(elems, pairs...)
		}
		if len(elems) > 0 {
			var pair = elems[0]
			if len(elems) > 1 {
				return pair, NewPairList(
					elems[1:]...,
				)
			}
			return pair, NewPairList()
		}
		return nil, NewPairList()
	}
}

func (l PairList) Tail() Sequential                         { _, t := l(); return t }
func (l PairList) TailPairs() ConsumeablePaired             { _, t := l(); return t }
func (l PairList) TailPairList() PairList                   { _, t := l(); return t }
func (l PairList) Head() Expression                         { h, _ := l(); return h }
func (l PairList) HeadPair() Paired                         { p, _ := l(); return p }
func (l PairList) Consume() (Expression, Sequential)        { return l() }
func (l PairList) ConsumePair() (Paired, ConsumeablePaired) { return l() }
func (l PairList) ConsumePairList() (Paired, PairList)      { return l() }
func (l PairList) Append(args ...Expression) Sequential {
	var pairs = make([]Paired, 0, len(args))
	for _, arg := range args {
		if arg.TypeFnc().Match(Pair) {
			if pair, ok := arg.(Paired); ok {
				pairs = append(pairs, pair)
			}
		}
	}
	return l.ConsFromPairs(pairs...)
}
func (l PairList) TypeFnc() TyFnc { return List }
func (l PairList) Null() PairList { return NewPairList() }
func (l PairList) Type() TyPattern {
	if l.Len() > 0 {
		return Def(List|Pair, l.TypeElem().TypeReturn())
	}
	return Def(Pair|List, None)
}

func (l PairList) ConsFromPairs(pairs ...Paired) Sequential {
	return PairList(func(args ...Paired) (Paired, PairList) {
		if len(args) > 0 {
			pairs = append(args, pairs...)
		}
		if len(pairs) == 0 {
			return l()
		}
		if len(pairs) == 1 {
			return pairs[0], l
		}
		var head Paired
		head, pairs = pairs[0], pairs[1:]
		return head, l.ConsFromPairs(pairs...).(PairList)
	})
}

func (l PairList) Cons(elems ...Expression) Sequential {
	var pairs = make([]Paired, 0, len(elems))
	for _, elem := range elems {
		if elem.Type().Match(Pair) {
			if pair, ok := elem.(Paired); ok {
				pairs = append(pairs, pair)
			}
		}
	}
	return PairList(func(args ...Paired) (Paired, PairList) {
		return l(append(args, pairs...)...)
	})
}

func (l PairList) Push(elems ...Paired) PairList {
	return ConcatPairLists(NewPairList(elems...), l)
}

func (l PairList) Call(args ...Expression) Expression {
	var pairs = []Paired{}
	if len(args) > 0 {
		pairs = append(pairs, argsToPaired(args...)...)
	}
	var head Expression
	head, l = l(pairs...)
	return head
}

func (l PairList) Empty() bool {
	if pair := l.HeadPair(); pair != nil {
		return pair.Empty()
	}
	return true
}

func (l PairList) Len() int {
	var (
		length     int
		head, tail = l()
	)
	if head != nil {
		length += 1 + tail.Len()
	}
	return length
}

func (l PairList) TypeElem() TyPattern {
	if l.Len() > 0 {
		return Def(l.Head().TypeFnc(), Def(Pair, List), l.Head().TypeFnc())
	}
	return Def(None, Def(Pair, List), None)
}

func (l PairList) TypeKey() d.Typed {
	return l.Head().(ValPair).TypeKey()
}

func (l PairList) TypeValue() d.Typed {
	return l.Head().(Paired).TypeValue()
}

func argsToPaired(args ...Expression) []Paired {
	var (
		pairs = []Paired{}
		alen  = len(args)
	)
	for i, arg := range args {
		if arg.TypeFnc().Match(Pair) {
			pairs = append(pairs, arg.(Paired))
		}
		if i < alen-2 {
			i = i + 1
			pairs = append(pairs, NewPair(arg, args[i]))
		}
		pairs = append(pairs, NewPair(arg, NewNone()))
	}
	return pairs
}

/// DATA SET
func NewSet(pairs ...ValPair) MapVal {
	var (
		set = make(map[string]Expression, len(pairs))
	)
	return func(args ...Expression) (Expression, map[string]Expression) {
		if len(args) > 0 {
			// access element by srting key
			if len(args) == 1 {
				var arg = args[0]
				if arg.Type().MatchArgs(DecNative("")) {
					if val, ok := set[arg.String()]; ok {
						return val, set
					}
				}
			}
			// add arguments to set
			for _, arg := range args {
				// argument implements paired → srting
				// representation of left field will be key
				if arg.TypeFnc().Match(Key | Pair | Index) {
					if pair, ok := arg.(Paired); ok {
						var val = pair.Value()
						set[pair.Left().String()] = val
						continue
					}
				}
				// argument does not implement paired → string
				// representation of value will be the key
				set[arg.String()] = arg
			}
		}
		return None, set
	}
}
func (s MapVal) Dict() map[string]Expression {
	var _, set = s()
	return set
}
func (s MapVal) Keys() []string {
	var keys = []string{}
	for key, _ := range s.Dict() {
		keys = append(keys, key)
	}
	return keys
}
func (s MapVal) Values() []Expression {
	var vals = []Expression{}
	for _, val := range s.Dict() {
		vals = append(vals, val)
	}
	return vals
}
func (s MapVal) Pairs() []Paired {
	var pairs = make([]Paired, 0, s.Len())
	for _, key := range s.Keys() {
		pairs = append(pairs, NewKeyPair(key, s.Get(key)))
	}
	return pairs
}
func (s MapVal) KeyPairs() []KeyPair {
	var pairs = make([]KeyPair, 0, s.Len())
	for _, key := range s.Keys() {
		pairs = append(pairs, NewKeyPair(key, s.Get(key)))
	}
	return pairs
}
func (s MapVal) Get(key string) Expression {
	if val, ok := s.Dict()[key]; ok {
		return val
	}
	return NewNone()
}
func (s MapVal) Len() int                             { return len(s.Keys()) }
func (s MapVal) TypeFnc() TyFnc                       { return Set }
func (s MapVal) GetByData(key Native) Expression      { return s.Get(key.String()) }
func (s MapVal) Set(key string, val Expression)       { s(NewKeyPair(key, val)) }
func (s MapVal) SetByData(key Native, val Expression) { s(NewKeyPair(key.String(), val)) }
func (s MapVal) Type() TyPattern                      { return Def(Set, s.TypeElem()) }
func (s MapVal) TypeElem() TyPattern {
	if s.Len() > 0 {
		if val := s.Values()[0]; !val.Type().Match(None) {
			return val.Type()
		}
	}
	return None.Type()
}
func (s MapVal) Call(args ...Expression) Expression {
	var expr, _ = s(args...)
	return expr
}
