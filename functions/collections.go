package functions

import (
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
	SetVal   func(...Expression) (Expression, map[string]Expression)
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
	makeBounds = func(bounds ...d.Integer) (low, high d.Typed, lesser, greater ComparatorType) {
		if len(bounds) == 0 {
			low, high = Lex_Infinite, Lex_Infinite
			lesser = func(...Expression) int { return 0 }
			greater = func(...Expression) int { return 0 }
		}

		if len(bounds) >= 1 {
			var minBound = bounds[0].(d.Native)
			low = DefValNative(minBound)
			lesser = NewComparator(func(args ...Expression) int {
				for _, arg := range args {
					if arg.Type().Match(Data) {
						var aint = arg.(Native).Eval()
						if aint.Type().Match(d.BigInt) {
							if minBound.Type().Match(d.BigInt) {
								if minBound.(d.BigIntVal).GoBigInt().Cmp(
									aint.(d.BigIntVal).GoBigInt()) < 0 {
									return -1
								}
							}
						}
						if aint.Type().Match(d.Integers) {
							if minBound.Type().Match(d.Integers) {
								if aint.(d.Integer).Int() < minBound.(d.Integer).Int() {
									return -1
								}
							}
						}
					}
				}
				return -2
			})
		}

		if len(bounds) >= 2 {
			var maxBound = bounds[1].(d.Native)
			high = DefValNative(maxBound)
			greater = NewComparator(func(args ...Expression) int {
				for _, arg := range args {
					if arg.Type().Match(Data) {
						var aint = arg.(Native).Eval()
						if maxBound.Type().Match(d.BigInt) {
							if aint.Type().Match(d.BigInt) {
								if maxBound.(d.BigIntVal).GoBigInt().Cmp(
									aint.(d.BigIntVal).GoBigInt()) > 0 {
									return 1
								}
							}
						}
						if aint.Type().Match(d.Integers) {
							if maxBound.Type().Match(d.Integers) {
								if aint.(d.Integer).Int() > maxBound.(d.Integer).Int() {
									return 1
								}
							}
						}
					}
				}
				return -2
			})
		}
		return low, high, lesser, greater
	}
	inBound = func(lesser, greater ComparatorType, args ...Expression) bool {
		for _, arg := range args {
			if isInt(arg) && lesser(arg) < 0 && greater(arg) > 0 {
				return false
			}
		}
		return true
	}
)

func NewEnumType(fnc func(...d.Integer) Expression, limits ...d.Integer) EnumType {
	var low, high, lesser, greater = makeBounds(limits...)
	return func(idx d.Integer) (EnumVal, d.Typed, d.Typed) {
		return func(args ...Expression) (Expression, d.Integer, EnumType) {
			if inBound(lesser, greater, args...) {
				return fnc(idx).Call(args...), idx, NewEnumType(fnc, limits...)
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
func (e EnumType) InBound(args ...Expression) bool {
	var _, _, lesser, greater = makeBounds(
		e.Low().(d.Integer),
		e.High().(d.Integer),
	)
	return inBound(lesser, greater, args...)
}
func (e EnumType) String() string {
	return "Enum " + e.Null().Type().TypeName()
}
func (e EnumType) Null() Expression {
	var result, _, _ = e(d.IntVal(0))
	return result
}
func (e EnumType) Unit() Expression {
	var result, _, _ = e(d.IntVal(1))
	return result
}
func (e EnumType) Type() TyPattern { return e.Unit().Type() }
func (e EnumType) TypeFnc() TyFnc  { return e.Unit().TypeFnc() }
func (e EnumType) Call(args ...Expression) Expression {
	if len(args) > 0 {
		if len(args) == 1 {
			if isInt(args[0]) {
				var result, _, _ = e(args[0].(Native).Eval().(d.Integer))
				return result
			}
		}
		var enums = NewVector()
		for _, arg := range args {
			if isInt(arg) {
				var result, _, _ = e(arg.(Native).Eval().(d.Integer))
				var num = enums.Con(result)
				enums = enums.Con(num)
			}
		}
		return enums
	}
	return e
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
	var _, _, etype = e()
	return etype
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
func (e EnumVal) Type() TyPattern                    { return e.Expr().Type() }
func (e EnumVal) TypeFnc() TyFnc                     { return e.Expr().TypeFnc() }
func (e EnumVal) Call(args ...Expression) Expression { return e.Expr().Call(args...) }

//////////////////////////////////////////////////////////////////////////////////////////
//// VECTORS (SLICES) OF VALUES
func NewEmptyVector(init ...Expression) VecVal { return NewVector() }

func NewVector(init ...Expression) VecVal {
	var vector = init
	return func(args ...Expression) []Expression {
		if len(args) > 0 {
			vector = append(
				vector,
				args...,
			)
		}
		return vector
	}
}

func ConVector(vec Vectorized, args ...Expression) VecVal {
	return NewVector(append(vec.Slice(), args...)...)
}

func AppendVectors(vec Vectorized, args ...Expression) VecVal {
	return NewVector(append(vec.Slice(), args...)...)
}

func AppendArgToVector(init ...Expression) VecVal {
	return func(args ...Expression) []Expression {
		return append(init, args...)
	}
}

func (v VecVal) Len() int            { return len(v()) }
func (v VecVal) Vector() VecVal      { return v }
func (v VecVal) Slice() []Expression { return v() }

func (v VecVal) Append(args ...Expression) Consumeable { return v.Con(args...) }

func (v VecVal) Prepend(args ...Expression) VecVal {
	return NewVector(append(args, v()...)...)
}

func (v VecVal) Reverse(args ...Expression) VecVal {
	var slice []Expression
	if v.Len() > 1 {
		slice = []Expression{}
		var vector = v()
		for i := v.Len() - 1; i > 0; i-- {
			slice = append(slice, vector[i])
		}
	}
	if len(args) > 0 {
		for _, arg := range args {
			v = v.Prepend(arg)
		}
	}
	return NewVector(slice...)
}
func (v VecVal) TypeFnc() TyFnc { return Vector }
func (v VecVal) Type() TyPattern {
	if v.Len() > 0 {
		return Def(Vector, v.TypeElem().TypeReturn())
	}
	return Def(Vector, None)
}
func (v VecVal) TypeElem() TyPattern {
	if v.Len() > 0 {
		return Def(v.Head().TypeFnc(),
			Vector, v.Head().TypeFnc())
	}
	return Def(None, Vector, None)
}

func (v VecVal) Con(args ...Expression) VecVal {
	return ConVector(v, args...)
}

func (v VecVal) Call(d ...Expression) Expression {
	return NewVector(v(d...)...)
}

func (v VecVal) Last() Expression {
	if v.Len() > 0 {
		return v()[v.Len()-1]
	}
	return nil
}

func (v VecVal) Head() Expression {
	if v.Len() > 0 {
		return v()[0]
	}
	return nil
}

func (v VecVal) Tail() Consumeable {
	if v.Len() > 1 {
		return NewVector(v()[1:]...)
	}
	return NewEmptyVector()
}

func (v VecVal) Consume() (Expression, Consumeable) {
	return v.Head(), v.Tail()
}

func (v VecVal) TailVec() VecVal {
	if v.Len() > 1 {
		return NewVector(v.Tail().(VecVal)()...)
	}
	return NewEmptyVector()
}

func (v VecVal) ConsumeVec() (Expression, VecVal) {
	return v.Head(), v.TailVec()
}

func (v VecVal) Clear() VecVal { return NewVector() }

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

//// RECURSIVE LIST OF VALUES
///
// base implementation of recursively linked lists
func ConList(list ListVal, elems ...Expression) ListVal {
	return list.Con(elems...)
}

func ConcatLists(a, b ListVal) ListVal {
	return ListVal(func(args ...Expression) (Expression, ListVal) {
		if len(args) > 0 {
			b = b.Con(args...)
		}
		var head Expression
		if head, a = a(); head != nil {
			return head, ConcatLists(a, b)
		}
		return b()
	})
}

func NewList(elems ...Expression) ListVal {
	return func(args ...Expression) (Expression, ListVal) {
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
func (l ListVal) Tail() Consumeable                  { _, t := l(); return t }
func (l ListVal) Head() Expression                   { h, _ := l(); return h }
func (l ListVal) TailList() ListVal                  { _, t := l(); return t }
func (l ListVal) Consume() (Expression, Consumeable) { return l() }
func (l ListVal) TypeFnc() TyFnc                     { return List }
func (l ListVal) Null() ListVal                      { return NewList() }
func (l ListVal) TypeElem() TyPattern {
	if l.Len() > 0 {
		return l.Head().Type()
	}
	return Def(None, List, None)
}

func (l ListVal) Type() TyPattern {
	if l.Len() > 0 {
		return Def(List, l.TypeElem().TypeReturn())
	}
	return Def(List, None)
}

func (l ListVal) ConsumeList() (Expression, ListVal) {
	return l.Head(), l.TailList()
}

func (l ListVal) Append(elems ...Expression) Consumeable { _, l = l(elems...); return l }
func (l ListVal) Con(elems ...Expression) ListVal {
	return ListVal(func(args ...Expression) (Expression, ListVal) {
		return l(append(elems, args...)...)
	})
}

func (l ListVal) Push(elems ...Expression) ListVal {
	return ConcatLists(NewList(elems...), l)
}

func (l ListVal) Slice() []Expression {
	var (
		vec  = NewVector()
		head Expression
		tail Consumeable
	)
	head, tail = l.Head(), l.Tail()
	for head != nil {
		vec = vec.Con(head)
		head, tail = tail.Consume()
	}
	return vec.Slice()
}

func (l ListVal) Call(args ...Expression) Expression {
	if len(args) > 0 {
		return l.Con(args...)
	}
	return l.Head()
}

func (l ListVal) GetIdx(n int) Expression {
	var head, list = l()
	for i := 0; i < n; i++ {
		head, list = list()
		if head == nil {
			return NewNone()
		}
	}
	return head
}

func (l ListVal) Empty() bool {
	if l.Head() != nil {
		if !None.Flag().Match(l.Head().TypeFnc()) {
			return false
		}
	}

	return true
}

func (l ListVal) Len() int {
	var (
		length     int
		head, tail = l()
	)
	if head != nil {
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

func ConPairVecFromPairs(rec PairVec, args ...Expression) PairVec {
	var pairs = []Paired{}
	for _, arg := range args {
		if pair, ok := arg.(Paired); ok {
			pairs = append(pairs, pair)
		}
	}
	return NewPairVectorFromPairs(append(rec(), pairs...)...)
}

func NewPairVec(args ...Paired) PairVec {
	return NewPairVectorFromPairs(args...)
}

func ConPairVec(rec PairVec, pairs ...Paired) PairVec {
	return NewPairVectorFromPairs(append(rec(), pairs...)...)
}

func ConPairVecFromArgs(pvec PairVec, args ...Expression) PairVec {
	var pairs = pvec.Pairs()
	for _, arg := range args {
		if pair, ok := arg.(Paired); ok {
			pairs = append(pairs, pair)
		}
	}
	return PairVec(func(args ...Paired) []Paired {
		if len(args) > 0 {
			return ConPairVec(pvec, args...)()
		}
		return append(pvec(), pairs...)
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

func (v PairVec) ConPairs(pairs ...Paired) PairVec {
	return ConPairVec(v, pairs...)
}

func (v PairVec) Con(args ...Expression) PairVec {
	return ConPairVecFromArgs(v, args...)
}

func (v PairVec) Append(args ...Expression) Consumeable { return v.Con(args...) }

func (v PairVec) Consume() (Expression, Consumeable) {
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
func (v PairVec) Tail() Consumeable {
	if v.Len() > 1 {
		return NewPairVectorFromPairs(v.Pairs()[1:]...)
	}
	return NewEmptyPairVec()
}

func (v PairVec) Call(args ...Expression) Expression {
	return v.Con(args...)
}

///////////////////////////////////////////////////////////////////////////////
//// LIST OF PAIRS
func ConPairList(list PairList, pairs ...Paired) PairList {
	return list.Con(pairs...)
}
func ConcatPairLists(a, b PairList) PairList {
	return PairList(func(args ...Paired) (Paired, PairList) {
		if len(args) > 0 {
			b = b.Con(args...)
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

func (l PairList) Tail() Consumeable                        { _, t := l(); return t }
func (l PairList) TailPairs() ConsumeablePaired             { _, t := l(); return t }
func (l PairList) TailPairList() PairList                   { _, t := l(); return t }
func (l PairList) Head() Expression                         { h, _ := l(); return h }
func (l PairList) HeadPair() Paired                         { p, _ := l(); return p }
func (l PairList) Consume() (Expression, Consumeable)       { return l() }
func (l PairList) ConsumePair() (Paired, ConsumeablePaired) { return l() }
func (l PairList) ConsumePairList() (Paired, PairList)      { return l() }
func (l PairList) Append(args ...Expression) Consumeable {
	var pairs = make([]Paired, 0, len(args))
	for _, arg := range args {
		if arg.TypeFnc().Match(Pair) {
			if pair, ok := arg.(Paired); ok {
				pairs = append(pairs, pair)
			}
		}
	}
	return l.Con(pairs...)
}
func (l PairList) TypeFnc() TyFnc { return List }
func (l PairList) Null() PairList { return NewPairList() }
func (l PairList) Type() TyPattern {
	if l.Len() > 0 {
		return Def(List|Pair, l.TypeElem().TypeReturn())
	}
	return Def(Pair|List, None)
}

func (l PairList) Con(elems ...Paired) PairList {
	return PairList(func(args ...Paired) (Paired, PairList) {
		return l(append(elems, args...)...)
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
func NewSet(pairs ...ValPair) SetVal {
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
func (s SetVal) Dict() map[string]Expression {
	var _, set = s()
	return set
}
func (s SetVal) Keys() []string {
	var keys = []string{}
	for key, _ := range s.Dict() {
		keys = append(keys, key)
	}
	return keys
}
func (s SetVal) Values() []Expression {
	var vals = []Expression{}
	for _, val := range s.Dict() {
		vals = append(vals, val)
	}
	return vals
}
func (s SetVal) Pairs() []Paired {
	var pairs = make([]Paired, 0, s.Len())
	for _, key := range s.Keys() {
		pairs = append(pairs, NewKeyPair(key, s.Get(key)))
	}
	return pairs
}
func (s SetVal) KeyPairs() []KeyPair {
	var pairs = make([]KeyPair, 0, s.Len())
	for _, key := range s.Keys() {
		pairs = append(pairs, NewKeyPair(key, s.Get(key)))
	}
	return pairs
}
func (s SetVal) Get(key string) Expression {
	if val, ok := s.Dict()[key]; ok {
		return val
	}
	return NewNone()
}
func (s SetVal) Len() int                             { return len(s.Keys()) }
func (s SetVal) TypeFnc() TyFnc                       { return Set }
func (s SetVal) GetByData(key Native) Expression      { return s.Get(key.String()) }
func (s SetVal) Set(key string, val Expression)       { s(NewKeyPair(key, val)) }
func (s SetVal) SetByData(key Native, val Expression) { s(NewKeyPair(key.String(), val)) }
func (s SetVal) Type() TyPattern                      { return Def(Set, s.TypeElem()) }
func (s SetVal) TypeElem() TyPattern {
	if s.Len() > 0 {
		if val := s.Values()[0]; !val.Type().Match(None) {
			return val.Type()
		}
	}
	return None.Type()
}
func (s SetVal) Call(args ...Expression) Expression {
	var expr, _ = s(args...)
	return expr
}
