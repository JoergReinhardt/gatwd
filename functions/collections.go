package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

type (
	//// COLLECTION
	ListType func(...Expression) (Expression, ListType)
	VecType  func(...Expression) []Expression

	PairType      func(...Expression) (Expression, Expression)
	KeyPairType   func(...Expression) (Expression, string)
	TypePairType  func(...Expression) (Expression, Typed)
	IndexPairType func(...Expression) (Expression, int)

	PairListType func(...Paired) (Paired, PairListType)
	PairVecType  func(...Paired) []Paired

	SetType func(...Expression) (Expression, map[string]Expression)
)

//// RECURSIVE LIST OF VALUES
///
// base implementation of recursively linked lists
func ConList(list ListType, elems ...Expression) ListType {
	return list.Con(elems...)
}

func ConcatLists(a, b ListType) ListType {
	return ListType(func(args ...Expression) (Expression, ListType) {
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

func NewList(elems ...Expression) ListType {
	return func(args ...Expression) (Expression, ListType) {
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
func (l ListType) Tail() Consumeable                  { _, t := l(); return t }
func (l ListType) Head() Expression                   { h, _ := l(); return h }
func (l ListType) TailList() ListType                 { _, t := l(); return t }
func (l ListType) Consume() (Expression, Consumeable) { return l() }
func (l ListType) TypeFnc() TyFnc                     { return List }
func (l ListType) Null() ListType                     { return NewList() }
func (l ListType) TypeElem() TyPattern {
	if l.Len() > 0 {
		return l.Head().Type()
	}
	return Def(None, List, None)
}

func (l ListType) Type() TyPattern {
	if l.Len() > 0 {
		return Def(List, l.TypeElem().TypeReturn())
	}
	return Def(List, None)
}

func (l ListType) ConsumeList() (Expression, ListType) {
	return l.Head(), l.TailList()
}

func (l ListType) Append(elems ...Expression) Consumeable { _, l = l(elems...); return l }
func (l ListType) Con(elems ...Expression) ListType {
	return ListType(func(args ...Expression) (Expression, ListType) {
		return l(append(elems, args...)...)
	})
}

func (l ListType) Push(elems ...Expression) ListType {
	return ConcatLists(NewList(elems...), l)
}

func (l ListType) Slice() []Expression {
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

func (l ListType) Call(args ...Expression) Expression {
	if len(args) > 0 {
		return l.Con(args...)
	}
	return l.Head()
}

func (l ListType) GetIdx(n int) Expression {
	var head, list = l()
	for i := 0; i < n; i++ {
		head, list = list()
		if head == nil {
			return NewNone()
		}
	}
	return head
}

func (l ListType) Empty() bool {
	if l.Head() != nil {
		if !None.Flag().Match(l.Head().TypeFnc()) {
			return false
		}
	}

	return true
}

func (l ListType) Len() int {
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
//// PAIRS OF VALUES
///
// pairs can be created empty, key & value may be constructed later
func NewEmptyPair() PairType {
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
func NewPair(l, r Expression) PairType {
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
func (p PairType) Pair() Paired                   { return p }
func (p PairType) Both() (Expression, Expression) { return p() }
func (p PairType) Left() Expression               { l, _ := p(); return l }
func (p PairType) Right() Expression              { _, r := p(); return r }
func (p PairType) Swap() (Expression, Expression) { l, r := p(); return r, l }
func (p PairType) SwappedPair() Paired            { return NewPair(p.Right(), p.Left()) }
func (p PairType) Slice() []Expression            { return []Expression{p.Left(), p.Right()} }
func (p PairType) Key() Expression                { return p.Left() }
func (p PairType) Value() Expression              { return p.Right() }
func (p PairType) TypeFnc() TyFnc                 { return Pair }
func (p PairType) TypeElem() TyPattern {
	if p.Right() != nil {
		return p.Right().Type()
	}
	return Def(None, Pair, None)
}
func (p PairType) TypeKey() d.Typed {
	if p.Left() != nil {
		return p.Left().Type()
	}
	return None
}
func (p PairType) TypeValue() d.Typed {
	if p.Right() != nil {
		return p.Right().Type()
	}
	return None
}
func (p PairType) Type() TyPattern {
	if p.TypeKey().Match(None) && p.TypeValue().Match(None) {
		return Def(Pair, None)
	}
	return Def(Pair, Def(p.TypeKey(), p.TypeValue()))
}

func (p PairType) Empty() bool {
	if p.Left() == nil || (!p.Left().TypeFnc().Flag().Match(None) &&
		(p.Right() == nil || (!p.Right().TypeFnc().Flag().Match(None)))) {
		return true
	}
	return false
}

func (p PairType) Call(args ...Expression) Expression {
	return NewPair(p.Key(), p.Value().Call(args...))
}

///////////////////////////////////////////////////////////////////////////////
//// ASSOCIATIVE PAIRS
///
// pair composed of a string key and a functional value
func NewKeyPair(key string, val Expression) KeyPairType {
	return func(...Expression) (Expression, string) { return val, key }
}

func (a KeyPairType) KeyStr() string                     { _, key := a(); return key }
func (a KeyPairType) Value() Expression                  { val, _ := a(); return val }
func (a KeyPairType) Left() Expression                   { return a.Value() }
func (a KeyPairType) Right() Expression                  { return DecData(d.StrVal(a.KeyStr())) }
func (a KeyPairType) Both() (Expression, Expression)     { return a.Left(), a.Right() }
func (a KeyPairType) Pair() Paired                       { return NewPair(a.Both()) }
func (a KeyPairType) Pairs() []Paired                    { return []Paired{NewPair(a.Both())} }
func (a KeyPairType) Key() Expression                    { return a.Right() }
func (a KeyPairType) Call(args ...Expression) Expression { return a.Value().Call(args...) }
func (a KeyPairType) TypeValue() d.Typed                 { return a.Value().Type() }
func (a KeyPairType) TypeKey() d.Typed                   { return Key }
func (a KeyPairType) TypeFnc() TyFnc                     { return Key }
func (p KeyPairType) Type() TyPattern {
	if p.TypeKey().Match(None) && p.TypeValue().Match(None) {
		return Def(Key|Pair, None)
	}
	return Def(Key|Pair, Def(p.TypeKey(), p.TypeValue()))
}

// implement swappable
func (p KeyPairType) Swap() (Expression, Expression) {
	l, r := p()
	return DecData(d.StrVal(r)), l
}
func (p KeyPairType) SwappedPair() Paired { return NewPair(p.Right(), p.Left()) }

func (a KeyPairType) Empty() bool {
	if a.Key() != nil && a.Value() != nil && a.Value().TypeFnc() != None {
		return false
	}
	return true
}

///////////////////////////////////////////////////////////////////////////////
//// INDEX PAIR
///
// pair composed of an integer and a functional value
func NewIndexPair(idx int, val Expression) IndexPairType {
	return func(...Expression) (Expression, int) { return val, idx }
}
func (a IndexPairType) Index() int                         { _, idx := a(); return idx }
func (a IndexPairType) Value() Expression                  { val, _ := a(); return val }
func (a IndexPairType) Left() Expression                   { return a.Value() }
func (a IndexPairType) Right() Expression                  { return DecData(d.IntVal(a.Index())) }
func (a IndexPairType) Both() (Expression, Expression)     { return a.Left(), a.Right() }
func (a IndexPairType) Pair() Paired                       { return a }
func (a IndexPairType) Pairs() []Paired                    { return []Paired{NewPair(a.Both())} }
func (a IndexPairType) Key() Expression                    { return a.Right() }
func (a IndexPairType) Call(args ...Expression) Expression { return a.Value().Call(args...) }
func (a IndexPairType) TypeFnc() TyFnc                     { return Index }
func (a IndexPairType) TypeKey() d.Typed                   { return Index }
func (a IndexPairType) TypeValue() d.Typed                 { return a.Value().Type() }
func (a IndexPairType) Type() TyPattern {
	if a.TypeKey().Match(None) && a.TypeValue().Match(None) {
		return Def(Index|Pair, None)
	}
	return Def(Index|Pair, Def(a.TypeKey(), a.TypeValue()))
}

// implement swappable
func (p IndexPairType) Swap() (Expression, Expression) {
	l, r := p()
	return DecData(d.New(r)), l
}
func (p IndexPairType) SwappedPair() Paired { return NewPair(p.Right(), p.Left()) }
func (a IndexPairType) Empty() bool {
	if a.Index() >= 0 && a.Value() != nil && a.Value().TypeFnc() != None {
		return true
	}
	return false
}

///////////////////////////////////////////////////////////////////////////////
//// LIST OF PAIRS
func ConPairList(list PairListType, pairs ...Paired) PairListType {
	return list.Con(pairs...)
}
func ConcatPairLists(a, b PairListType) PairListType {
	return PairListType(func(args ...Paired) (Paired, PairListType) {
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
func NewPairList(elems ...Paired) PairListType {
	return func(pairs ...Paired) (Paired, PairListType) {
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

func (l PairListType) Tail() Consumeable                        { _, t := l(); return t }
func (l PairListType) TailPairs() ConsumeablePaired             { _, t := l(); return t }
func (l PairListType) TailPairList() PairListType               { _, t := l(); return t }
func (l PairListType) Head() Expression                         { h, _ := l(); return h }
func (l PairListType) HeadPair() Paired                         { p, _ := l(); return p }
func (l PairListType) Consume() (Expression, Consumeable)       { return l() }
func (l PairListType) ConsumePair() (Paired, ConsumeablePaired) { return l() }
func (l PairListType) ConsumePairList() (Paired, PairListType)  { return l() }
func (l PairListType) Append(args ...Expression) Consumeable {
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
func (l PairListType) TypeFnc() TyFnc     { return List }
func (l PairListType) Null() PairListType { return NewPairList() }
func (l PairListType) Type() TyPattern {
	if l.Len() > 0 {
		return Def(List|Pair, l.TypeElem().TypeReturn())
	}
	return Def(Pair|List, None)
}

func (l PairListType) Con(elems ...Paired) PairListType {
	return PairListType(func(args ...Paired) (Paired, PairListType) {
		return l(append(elems, args...)...)
	})
}

func (l PairListType) Push(elems ...Paired) PairListType {
	return ConcatPairLists(NewPairList(elems...), l)
}

func (l PairListType) Call(args ...Expression) Expression {
	var pairs = []Paired{}
	if len(args) > 0 {
		pairs = append(pairs, argsToPaired(args...)...)
	}
	var head Expression
	head, l = l(pairs...)
	return head
}

func (l PairListType) Empty() bool {
	if pair := l.HeadPair(); pair != nil {
		return pair.Empty()
	}
	return true
}

func (l PairListType) Len() int {
	var (
		length     int
		head, tail = l()
	)
	if head != nil {
		length += 1 + tail.Len()
	}
	return length
}

func (l PairListType) TypeElem() TyPattern {
	if l.Len() > 0 {
		return Def(l.Head().TypeFnc(), Def(Pair, List), l.Head().TypeFnc())
	}
	return Def(None, Def(Pair, List), None)
}

func (l PairListType) TypeKey() d.Typed {
	return l.Head().(PairType).TypeKey()
}

func (l PairListType) TypeValue() d.Typed {
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

//////////////////////////////////////////////////////////////////////////////////////////
//// VECTORS (SLICES) OF VALUES
func NewEmptyVector(init ...Expression) VecType { return NewVector() }

func NewVector(init ...Expression) VecType {
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

func ConVector(vec Vectorized, args ...Expression) VecType {
	return NewVector(append(vec.Slice(), args...)...)
}

func AppendVectors(vec Vectorized, args ...Expression) VecType {
	return NewVector(append(vec.Slice(), args...)...)
}

func AppendArgToVector(init ...Expression) VecType {
	return func(args ...Expression) []Expression {
		return append(init, args...)
	}
}

func (v VecType) Len() int            { return len(v()) }
func (v VecType) Vector() VecType     { return v }
func (v VecType) Slice() []Expression { return v() }

func (v VecType) Append(args ...Expression) Consumeable { return v.Con(args...) }

func (v VecType) Prepend(args ...Expression) VecType {
	return NewVector(append(args, v()...)...)
}

func (v VecType) Reverse(args ...Expression) VecType {
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
func (v VecType) TypeFnc() TyFnc { return Vector }
func (v VecType) Type() TyPattern {
	if v.Len() > 0 {
		return Def(Vector, v.TypeElem().TypeReturn())
	}
	return Def(Vector, None)
}
func (v VecType) TypeElem() TyPattern {
	if v.Len() > 0 {
		return Def(v.Head().TypeFnc(),
			Vector, v.Head().TypeFnc())
	}
	return Def(None, Vector, None)
}

func (v VecType) Con(args ...Expression) VecType {
	return ConVector(v, args...)
}

func (v VecType) Call(d ...Expression) Expression {
	return NewVector(v(d...)...)
}

func (v VecType) Last() Expression {
	if v.Len() > 0 {
		return v()[v.Len()-1]
	}
	return nil
}

func (v VecType) Head() Expression {
	if v.Len() > 0 {
		return v()[0]
	}
	return nil
}

func (v VecType) Tail() Consumeable {
	if v.Len() > 1 {
		return NewVector(v()[1:]...)
	}
	return NewEmptyVector()
}

func (v VecType) Consume() (Expression, Consumeable) {
	return v.Head(), v.Tail()
}

func (v VecType) TailVec() VecType {
	if v.Len() > 1 {
		return NewVector(v.Tail().(VecType)()...)
	}
	return NewEmptyVector()
}

func (v VecType) ConsumeVec() (Expression, VecType) {
	return v.Head(), v.TailVec()
}

func (v VecType) Clear() VecType { return NewVector() }

func (v VecType) Empty() bool {
	if len(v()) > 0 {
		for _, val := range v() {
			if !val.TypeFnc().Flag().Match(None) {
				return false
			}
		}
	}
	return true
}
func (v VecType) Get(i int) (Expression, bool) {
	if i < v.Len() {
		return v()[i], true
	}
	return NewNone(), false
}

func (v VecType) Set(i int, val Expression) (Vectorized, bool) {
	if i < v.Len() {
		var slice = v()
		slice[i] = val
		return VecType(
			func(elems ...Expression) []Expression {
				return slice
			}), true

	}
	return v, false
}

///////////////////////////////////////////////////////////////////////////////
//// ASSOCIATIVE SLICE OF VALUE PAIRS
///
// list of associative pairs in sequential order associated, sorted and
// searched by left value of the pairs
func NewEmptyPairVec() PairVecType {
	return PairVecType(func(args ...Paired) []Paired {
		var pairs = []Paired{}
		if len(args) > 0 {
			pairs = append(pairs, args...)
		}
		return pairs
	})
}

func NewPairVectorFromPairs(pairs ...Paired) PairVecType {
	return PairVecType(func(args ...Paired) []Paired {
		if len(args) > 0 {
			return append(pairs, args...)
		}
		return pairs
	})
}

func ConPairListFromArgs(rec PairVecType, args ...Expression) PairVecType {
	var pairs = []Paired{}
	for _, arg := range args {
		if pair, ok := arg.(Paired); ok {
			pairs = append(pairs, pair)
		}
	}
	return NewPairVectorFromPairs(append(rec(), pairs...)...)
}

func NewPairVec(args ...Paired) PairVecType {
	return NewPairVectorFromPairs(args...)
}

func ConPairVec(rec PairVecType, pairs ...Paired) PairVecType {
	return NewPairVectorFromPairs(append(rec(), pairs...)...)
}

func ConPairVecFromArgs(pvec PairVecType, args ...Expression) PairVecType {
	var pairs = pvec.Pairs()
	for _, arg := range args {
		if pair, ok := arg.(Paired); ok {
			pairs = append(pairs, pair)
		}
	}
	return PairVecType(func(args ...Paired) []Paired {
		if len(args) > 0 {
			return ConPairVec(pvec, args...)()
		}
		return append(pvec(), pairs...)
	})
}
func (v PairVecType) Len() int { return len(v()) }
func (v PairVecType) Type() TyPattern {
	if v.Len() > 0 {
		return Def(Vector|Pair, v.TypeElem().TypeReturn())
	}
	return Def(None, Vector|Pair, None)
}
func (v PairVecType) TypeFnc() TyFnc { return Vector }

func (v PairVecType) ConPairs(pairs ...Paired) PairVecType {
	return ConPairVec(v, pairs...)
}

func (v PairVecType) Con(args ...Expression) PairVecType {
	return ConPairVecFromArgs(v, args...)
}

func (v PairVecType) Append(args ...Expression) Consumeable { return v.Con(args...) }

func (v PairVecType) Consume() (Expression, Consumeable) {
	return v.Head(), v.Tail()
}

func (v PairVecType) ConsumePairVec() (Paired, PairVecType) {
	return v.HeadPair(), v.Tail().(PairVecType)
}

func (v PairVecType) Empty() bool {
	if len(v()) > 0 {
		for _, pair := range v() {
			if !pair.Empty() {
				return false
			}
		}
	}
	return true
}
func (v PairVecType) TypeElem() TyPattern {
	if v.Len() > 0 {
		return Def(v.Head().TypeFnc(), Vector|Pair, v.Head().TypeFnc())
	}
	return Def(None, Vector|Pair, None)
}
func (v PairVecType) TypeKey() d.Typed {
	if v.Len() > 0 {
		return v.Pairs()[0].Left().Type()
	}
	return None.TypeFnc()
}
func (v PairVecType) TypeValue() d.Typed {
	if v.Len() > 0 {
		return v.Pairs()[0].Right().Type()
	}
	return None.TypeFnc()
}
func (v PairVecType) Get(idx int) (Paired, bool) {
	if idx < v.Len()-1 {
		return v()[idx], true
	}
	return NewKeyPair("None", None), false
}

func (v PairVecType) Pairs() []Paired {
	var pairs = []Paired{}
	for _, pair := range v() {
		pairs = append(pairs, pair)
	}
	return pairs
}

func (v PairVecType) ConsumePair() (Paired, ConsumeablePaired) {
	var pairs = v()
	if len(pairs) > 0 {
		if len(pairs) > 1 {
			return pairs[0], NewPairVec(pairs[1:]...)
		}
		return pairs[0], NewPairVec()
	}
	return nil, NewPairVec()
}

func (v PairVecType) SwitchedPairs() []Paired {
	var switched = []Paired{}
	for _, pair := range v() {
		switched = append(
			switched,
			pair,
		)
	}
	return switched
}

func (v PairVecType) Slice() []Expression {
	var fncs = []Expression{}
	for _, pair := range v() {
		fncs = append(fncs, NewPair(pair.Left(), pair.Right()))
	}
	return fncs
}

func (v PairVecType) HeadPair() Paired {
	if v.Len() > 0 {
		return v()[0].(Paired)
	}
	return NewPair(NewNone(), NewNone())
}
func (v PairVecType) Head() Expression {
	if v.Len() > 0 {
		return v.Pairs()[0]
	}
	return nil
}

func (v PairVecType) TailPairs() ConsumeablePaired {
	if v.Len() > 1 {
		return NewPairVectorFromPairs(v.Pairs()[1:]...)
	}
	return NewEmptyPairVec()
}
func (v PairVecType) Tail() Consumeable {
	if v.Len() > 1 {
		return NewPairVectorFromPairs(v.Pairs()[1:]...)
	}
	return NewEmptyPairVec()
}

func (v PairVecType) Call(args ...Expression) Expression {
	return v.Con(args...)
}

/// DATA SET
func NewSet(pairs ...PairType) SetType {
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
func (s SetType) Dict() map[string]Expression {
	var _, set = s()
	return set
}
func (s SetType) Keys() []string {
	var keys = []string{}
	for key, _ := range s.Dict() {
		keys = append(keys, key)
	}
	return keys
}
func (s SetType) Values() []Expression {
	var vals = []Expression{}
	for _, val := range s.Dict() {
		vals = append(vals, val)
	}
	return vals
}
func (s SetType) Pairs() []Paired {
	var pairs = make([]Paired, 0, s.Len())
	for _, key := range s.Keys() {
		pairs = append(pairs, NewKeyPair(key, s.Get(key)))
	}
	return pairs
}
func (s SetType) KeyPairs() []KeyPairType {
	var pairs = make([]KeyPairType, 0, s.Len())
	for _, key := range s.Keys() {
		pairs = append(pairs, NewKeyPair(key, s.Get(key)))
	}
	return pairs
}
func (s SetType) Get(key string) Expression {
	if val, ok := s.Dict()[key]; ok {
		return val
	}
	return NewNone()
}
func (s SetType) Len() int                             { return len(s.Keys()) }
func (s SetType) TypeFnc() TyFnc                       { return Set }
func (s SetType) GetByData(key Native) Expression      { return s.Get(key.String()) }
func (s SetType) Set(key string, val Expression)       { s(NewKeyPair(key, val)) }
func (s SetType) SetByData(key Native, val Expression) { s(NewKeyPair(key.String(), val)) }
func (s SetType) Type() TyPattern                      { return Def(Set, s.TypeElem()) }
func (s SetType) TypeElem() TyPattern {
	if s.Len() > 0 {
		if val := s.Values()[0]; !val.Type().Match(None) {
			return val.Type()
		}
	}
	return None.Type()
}
func (s SetType) Call(args ...Expression) Expression {
	var expr, _ = s(args...)
	return expr
}
