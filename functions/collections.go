package functions

import d "github.com/joergreinhardt/gatwd/data"

type (
	//// COLLECTION
	ListCol func(...Expression) (Expression, ListCol)
	VecCol  func(...Expression) []Expression

	PairVal   func(...Expression) (Expression, Expression)
	KeyPair   func(...Expression) (Expression, string)
	TypedPair func(...Expression) (Expression, Typed)
	IndexPair func(...Expression) (Expression, int)

	PairList func(...Paired) (Paired, PairList)
	PairVec  func(...Paired) []Paired
)

//// RECURSIVE LIST OF VALUES
///
// base implementation of recursively linked lists
func ConList(list ListCol, elems ...Expression) ListCol {
	return list.Con(elems...)
}

func ConcatLists(a, b ListCol) ListCol {
	return ListCol(func(args ...Expression) (Expression, ListCol) {
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

func NewList(elems ...Expression) ListCol {
	return func(args ...Expression) (Expression, ListCol) {
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
func (l ListCol) Tail() Consumeable                  { _, t := l(); return t }
func (l ListCol) Head() Expression                   { h, _ := l(); return h }
func (l ListCol) TailList() ListCol                  { _, t := l(); return t }
func (l ListCol) Consume() (Expression, Consumeable) { return l() }
func (l ListCol) TypeFnc() TyFnc                     { return List }
func (l ListCol) Null() ListCol                      { return NewList() }
func (l ListCol) Type() TyPattern                    { return Def(List, l.TypeElem()) }
func (l ListCol) TypeElem() d.Typed {
	if l.Len() > 0 {
		return l.Head().Type()
	}
	return None
}

func (l ListCol) ConsumeList() (Expression, ListCol) {
	return l.Head(), l.TailList()
}

func (l ListCol) Con(elems ...Expression) ListCol {
	return ListCol(func(args ...Expression) (Expression, ListCol) {
		return l(append(elems, args...)...)
	})
}

func (l ListCol) Push(elems ...Expression) ListCol {
	return ConcatLists(NewList(elems...), l)
}

func (l ListCol) Slice() []Expression {
	var (
		vec  = NewVector()
		head Expression
		tail Consumeable
	)
	head, tail = l.Head(), l.Tail()
	for head != nil {
		vec = vec.Append(head)
		head, tail = tail.Consume()
	}
	return vec.Slice()
}

func (l ListCol) Call(args ...Expression) Expression {
	if len(args) > 0 {
		return l.Con(args...)
	}
	return l.Head()
}

func (l ListCol) GetIdx(n int) Expression {
	var head, list = l()
	for i := 0; i < n; i++ {
		head, list = list()
		if head == nil {
			return NewNone()
		}
	}
	return head
}

func (l ListCol) Empty() bool {
	if l.Head() != nil {
		if !None.Flag().Match(l.Head().TypeFnc()) {
			return false
		}
	}

	return true
}

func (l ListCol) Len() int {
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
func NewEmptyPair() PairVal {
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
func NewPair(l, r Expression) PairVal {
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
func (p PairVal) Pair() Paired                       { return p }
func (p PairVal) Both() (Expression, Expression)     { return p() }
func (p PairVal) Left() Expression                   { l, _ := p(); return l }
func (p PairVal) Right() Expression                  { _, r := p(); return r }
func (p PairVal) Swap() (Expression, Expression)     { l, r := p(); return r, l }
func (p PairVal) SwappedPair() Paired                { return NewPair(p.Right(), p.Left()) }
func (p PairVal) TypeElem() d.Typed                  { return p.Value().Type() }
func (p PairVal) Slice() []Expression                { return []Expression{p.Left(), p.Right()} }
func (p PairVal) Key() Expression                    { return p.Left() }
func (p PairVal) Value() Expression                  { return p.Right() }
func (p PairVal) KeyType() d.Typed                   { return p.Left().Type() }
func (p PairVal) ValType() d.Typed                   { return p.Right().Type() }
func (p PairVal) TypeFnc() TyFnc                     { return Pair }
func (p PairVal) Consume() (Expression, Consumeable) { return p.Head(), p.Tail() }
func (p PairVal) Head() Expression                   { return p.Key() }
func (p PairVal) Tail() Consumeable {
	var value = p.Value()
	if value.Type().Match(SumTypes) {
		if cons, ok := value.(Consumeable); ok {
			return cons
		}
	}
	return NewList(value)
}
func (p PairVal) Type() TyPattern { return Def(Pair, Def(p.KeyType(), p.ValType())) }

func (p PairVal) Empty() bool {
	if p.Left() == nil || (!p.Left().TypeFnc().Flag().Match(None) &&
		(p.Right() == nil || (!p.Right().TypeFnc().Flag().Match(None)))) {
		return true
	}
	return false
}

func (p PairVal) Call(args ...Expression) Expression {
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
func (a KeyPair) Right() Expression                  { return NewData(d.StrVal(a.KeyStr())) }
func (a KeyPair) Both() (Expression, Expression)     { return a.Left(), a.Right() }
func (a KeyPair) Pair() Paired                       { return NewPair(a.Both()) }
func (a KeyPair) Pairs() []Paired                    { return []Paired{NewPair(a.Both())} }
func (a KeyPair) Key() Expression                    { return a.Right() }
func (a KeyPair) Call(args ...Expression) Expression { return a.Value().Call(args...) }
func (a KeyPair) ValType() d.Typed                   { return a.Value().Type() }
func (a KeyPair) KeyType() d.Typed                   { return Key }
func (a KeyPair) TypeFnc() TyFnc                     { return Key }
func (a KeyPair) TypeNat() d.TyNat                   { return d.Function }
func (p KeyPair) Type() TyPattern {
	return Def(Def(Pair, Key), Def(p.KeyType(), p.ValType()))
}

// implement swappable
func (p KeyPair) Swap() (Expression, Expression) {
	l, r := p()
	return NewData(d.StrVal(r)), l
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
func (a IndexPair) Right() Expression                  { return NewData(d.IntVal(a.Index())) }
func (a IndexPair) Both() (Expression, Expression)     { return a.Left(), a.Right() }
func (a IndexPair) Pair() Paired                       { return a }
func (a IndexPair) Pairs() []Paired                    { return []Paired{NewPair(a.Both())} }
func (a IndexPair) Key() Expression                    { return a.Right() }
func (a IndexPair) Call(args ...Expression) Expression { return a.Value().Call(args...) }
func (a IndexPair) TypeFnc() TyFnc                     { return Index }
func (a IndexPair) TypeNat() d.TyNat                   { return d.Function }
func (a IndexPair) KeyType() d.Typed                   { return Index }
func (a IndexPair) ValType() d.Typed                   { return a.Value().Type() }
func (a IndexPair) Type() TyPattern {
	return Def(Def(Pair, Index), Def(a.KeyType(), a.ValType()))
}

// implement swappable
func (p IndexPair) Swap() (Expression, Expression) {
	l, r := p()
	return NewData(New(r)), l
}
func (p IndexPair) SwappedPair() Paired { return NewPair(p.Right(), p.Left()) }
func (a IndexPair) Empty() bool {
	if a.Index() >= 0 && a.Value() != nil && a.Value().TypeFnc() != None {
		return true
	}
	return false
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
func (l PairList) Type() TyPattern {
	return Def(Def(List, Pair), l.TypeElem())
}
func (l PairList) TypeFnc() TyFnc   { return List }
func (l PairList) TypeNat() d.TyNat { return d.Function }
func (l PairList) Null() PairList   { return NewPairList() }

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

func (l PairList) TypeElem() d.Typed {
	if l.Len() > 0 {
		return l.Head().Type()
	}
	return None.Type()
}

func (l PairList) KeyType() d.Typed {
	return l.Head().(PairVal).KeyType()
}

func (l PairList) ValType() d.Typed {
	return l.Head().(Paired).ValType()
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
func NewEmptyVector(init ...Expression) VecCol { return NewVector() }

func NewVector(init ...Expression) VecCol {
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

func ConVector(vec Vectorized, args ...Expression) VecCol {
	return NewVector(append(vec.Slice(), args...)...)
}

func AppendVectors(vec Vectorized, args ...Expression) VecCol {
	return NewVector(append(vec.Slice(), args...)...)
}

func AppendArgToVector(init ...Expression) VecCol {
	return func(args ...Expression) []Expression {
		return append(init, args...)
	}
}

func (v VecCol) Len() int            { return len(v()) }
func (v VecCol) Vector() VecCol      { return v }
func (v VecCol) Slice() []Expression { return v() }

func (v VecCol) Append(args ...Expression) VecCol {
	return NewVector(append(v(), args...)...)
}

func (v VecCol) Prepend(args ...Expression) VecCol {
	return NewVector(append(args, v()...)...)
}

func (v VecCol) Reverse(args ...Expression) VecCol {
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
func (v VecCol) TypeFnc() TyFnc   { return Vector }
func (v VecCol) TypeNat() d.TyNat { return d.Function }
func (v VecCol) Type() TyPattern  { return Def(Vector, v.TypeElem()) }
func (v VecCol) TypeElem() d.Typed {
	if v.Len() > 0 {
		return v.Head().Type()
	}
	return None.Type()
}

func (v VecCol) Con(args ...Expression) VecCol {
	return ConVector(v, args...)
}

func (v VecCol) Call(d ...Expression) Expression {
	return NewVector(v(d...)...)
}

func (v VecCol) Last() Expression {
	if v.Len() > 0 {
		return v()[v.Len()-1]
	}
	return nil
}

func (v VecCol) Head() Expression {
	if v.Len() > 0 {
		return v()[0]
	}
	return nil
}

func (v VecCol) Tail() Consumeable {
	if v.Len() > 1 {
		return NewVector(v()[1:]...)
	}
	return NewEmptyVector()
}

func (v VecCol) Consume() (Expression, Consumeable) {
	return v.Head(), v.Tail()
}

func (v VecCol) TailVec() VecCol {
	if v.Len() > 1 {
		return NewVector(v.Tail().(VecCol)()...)
	}
	return NewEmptyVector()
}

func (v VecCol) ConsumeVec() (Expression, VecCol) {
	return v.Head(), v.TailVec()
}

func (v VecCol) Clear() VecCol { return NewVector() }

func (v VecCol) Empty() bool {
	if len(v()) > 0 {
		for _, val := range v() {
			if !val.TypeFnc().Flag().Match(None) {
				return false
			}
		}
	}
	return true
}
func (v VecCol) Get(i int) (Expression, bool) {
	if i < v.Len() {
		return v()[i], true
	}
	return NewNone(), false
}

func (v VecCol) Set(i int, val Expression) (Vectorized, bool) {
	if i < v.Len() {
		var slice = v()
		slice[i] = val
		return VecCol(
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

func ConPairListFromArgs(rec PairVec, args ...Expression) PairVec {
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
	return Def(Def(Vector, Pair), v.TypeElem())
}
func (v PairVec) TypeFnc() TyFnc   { return Vector }
func (v PairVec) TypeNat() d.TyNat { return d.Function }

func (v PairVec) ConPairs(pairs ...Paired) PairVec {
	return ConPairVec(v, pairs...)
}

func (v PairVec) Con(args ...Expression) PairVec {
	return ConPairVecFromArgs(v, args...)
}

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
func (v PairVec) TypeElem() d.Typed {
	if v.Len() > 0 {
		v()[0].Type()
	}
	return None
}
func (v PairVec) KeyType() d.Typed {
	if v.Len() > 0 {
		return v.Pairs()[0].Left().Type()
	}
	return None.TypeFnc()
}
func (v PairVec) ValType() d.Typed {
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
