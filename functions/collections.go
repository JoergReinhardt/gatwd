package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

type (
	//// COLLECTION
	ColList func(...Expression) (Expression, ColList)
	ColVec  func(...Expression) []Expression

	PairVal   func(...Expression) (Expression, Expression)
	KeyPair   func(...Expression) (Expression, string)
	TypedPair func(...Expression) (Expression, Typed)
	IndexPair func(...Expression) (Expression, int)

	ColPairL func(...Paired) (Paired, ColPairL)
	ColPairV func(...Paired) []Paired
)

//// RECURSIVE LIST OF VALUES
///
// base implementation of recursively linked lists
func ConList(list ColList, elems ...Expression) ColList {
	return list.Con(elems...)
}

func ConcatLists(a, b ColList) ColList {
	return ColList(func(args ...Expression) (Expression, ColList) {
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

func NewList(elems ...Expression) ColList {
	return func(args ...Expression) (Expression, ColList) {
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
func (l ColList) Tail() Consumeable                  { _, t := l(); return t }
func (l ColList) Head() Expression                   { h, _ := l(); return h }
func (l ColList) TailList() ColList                  { _, t := l(); return t }
func (l ColList) Consume() (Expression, Consumeable) { return l() }
func (l ColList) TypeFnc() TyFnc                     { return List }
func (l ColList) Null() ColList                      { return NewList() }
func (l ColList) Type() TyPattern                    { return Def(List, l.TypeElem()) }
func (l ColList) TypeElem() d.Typed {
	if l.Len() > 0 {
		return l.Head().Type()
	}
	return None
}

func (l ColList) ConsumeList() (Expression, ColList) {
	return l.Head(), l.TailList()
}

func (l ColList) Append(elems ...Expression) Consumeable { _, l = l(elems...); return l }
func (l ColList) Con(elems ...Expression) ColList {
	return ColList(func(args ...Expression) (Expression, ColList) {
		return l(append(elems, args...)...)
	})
}

func (l ColList) Push(elems ...Expression) ColList {
	return ConcatLists(NewList(elems...), l)
}

func (l ColList) Slice() []Expression {
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

func (l ColList) Call(args ...Expression) Expression {
	if len(args) > 0 {
		return l.Con(args...)
	}
	return l.Head()
}

func (l ColList) GetIdx(n int) Expression {
	var head, list = l()
	for i := 0; i < n; i++ {
		head, list = list()
		if head == nil {
			return NewNone()
		}
	}
	return head
}

func (l ColList) Empty() bool {
	if l.Head() != nil {
		if !None.Flag().Match(l.Head().TypeFnc()) {
			return false
		}
	}

	return true
}

func (l ColList) Len() int {
	var (
		length     int
		head, tail = l()
	)
	if head != nil {
		length += 1 + tail.Len()
	}
	return length
}

func (l ColList) Map(fn Expression) ColList {
	return func(args ...Expression) (Expression, ColList) {
		if len(args) > 0 {
			var head, list = l(args...)
			return fn.Call(head), list.Map(fn)
		}
		var head, list = l()
		if head != nil {
			return fn.Call(head), list.Map(fn)
		}
		return nil, NewList().Map(fn)
	}
}

func (l ColList) Apply(exprs Consumeable) ColList {
	return func(args ...Expression) (Expression, ColList) {
		if len(args) > 0 {
			l = l.Con(args...)
		}
		var fn, fns = exprs.Consume()
		if fn != nil {
			return ConcatLists(l.Map(fn),
				l.Apply(fns))()
		}
		return nil, l
	}
}
func (l ColList) FoldL(acc, init Expression) ColList {
	return func(args ...Expression) (Expression, ColList) {
		if len(args) > 0 {
			l = l.Con(args...)
			return l()
		}
		var head, tail = l()
		if head != nil {
			init = acc.Call(init, head)
			return init, tail.FoldL(acc, init)
		}
		return nil, tail.FoldL(acc, init)
	}
}

func (l ColList) Filter(filter TestVal) ColList {
	return func(args ...Expression) (Expression, ColList) {
		if len(args) > 0 {
			l = l.Con(args...)
		}
		var head, tail = l()
		if head != nil {
			if filter(head) {
				return head, tail.Filter(filter)
			}
			return tail.Filter(filter)()
		}
		return nil, tail
	}
}

func (l ColList) TakeN(n int, elems ...Expression) (ColVec, ColList) {
	if len(elems) < n {
		var head, tail = l()
		for head != nil {
			elems = append(elems, head)
			return tail.TakeN(n, elems...)
		}
	}
	return NewVector(elems...), l
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
func (p PairVal) Pair() Paired                   { return p }
func (p PairVal) Both() (Expression, Expression) { return p() }
func (p PairVal) Left() Expression               { l, _ := p(); return l }
func (p PairVal) Right() Expression              { _, r := p(); return r }
func (p PairVal) Swap() (Expression, Expression) { l, r := p(); return r, l }
func (p PairVal) SwappedPair() Paired            { return NewPair(p.Right(), p.Left()) }
func (p PairVal) TypeElem() d.Typed              { return p.Value().Type() }
func (p PairVal) Slice() []Expression            { return []Expression{p.Left(), p.Right()} }
func (p PairVal) Key() Expression                { return p.Left() }
func (p PairVal) Value() Expression              { return p.Right() }
func (p PairVal) KeyType() d.Typed               { return p.Left().Type() }
func (p PairVal) ValType() d.Typed               { return p.Right().Type() }
func (p PairVal) TypeFnc() TyFnc                 { return Pair }
func (p PairVal) Type() TyPattern                { return Def(Pair, Def(p.KeyType(), p.ValType())) }

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
func ConPairList(list ColPairL, pairs ...Paired) ColPairL {
	return list.Con(pairs...)
}
func ConcatPairLists(a, b ColPairL) ColPairL {
	return ColPairL(func(args ...Paired) (Paired, ColPairL) {
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
func NewPairList(elems ...Paired) ColPairL {
	return func(pairs ...Paired) (Paired, ColPairL) {
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

func (l ColPairL) Tail() Consumeable                        { _, t := l(); return t }
func (l ColPairL) TailPairs() ConsumeablePaired             { _, t := l(); return t }
func (l ColPairL) TailPairList() ColPairL                   { _, t := l(); return t }
func (l ColPairL) Head() Expression                         { h, _ := l(); return h }
func (l ColPairL) HeadPair() Paired                         { p, _ := l(); return p }
func (l ColPairL) Consume() (Expression, Consumeable)       { return l() }
func (l ColPairL) ConsumePair() (Paired, ConsumeablePaired) { return l() }
func (l ColPairL) ConsumePairList() (Paired, ColPairL)      { return l() }
func (l ColPairL) Append(args ...Expression) Consumeable {
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
func (l ColPairL) Type() TyPattern {
	return Def(Def(List, Pair), l.TypeElem())
}
func (l ColPairL) TypeFnc() TyFnc   { return List }
func (l ColPairL) TypeNat() d.TyNat { return d.Function }
func (l ColPairL) Null() ColPairL   { return NewPairList() }

func (l ColPairL) Con(elems ...Paired) ColPairL {
	return ColPairL(func(args ...Paired) (Paired, ColPairL) {
		return l(append(elems, args...)...)
	})
}

func (l ColPairL) Push(elems ...Paired) ColPairL {
	return ConcatPairLists(NewPairList(elems...), l)
}

func (l ColPairL) Call(args ...Expression) Expression {
	var pairs = []Paired{}
	if len(args) > 0 {
		pairs = append(pairs, argsToPaired(args...)...)
	}
	var head Expression
	head, l = l(pairs...)
	return head
}

func (l ColPairL) Empty() bool {
	if pair := l.HeadPair(); pair != nil {
		return pair.Empty()
	}
	return true
}

func (l ColPairL) Len() int {
	var (
		length     int
		head, tail = l()
	)
	if head != nil {
		length += 1 + tail.Len()
	}
	return length
}

func (l ColPairL) TypeElem() d.Typed {
	if l.Len() > 0 {
		return l.Head().Type()
	}
	return None.Type()
}

func (l ColPairL) KeyType() d.Typed {
	return l.Head().(PairVal).KeyType()
}

func (l ColPairL) ValType() d.Typed {
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
func NewEmptyVector(init ...Expression) ColVec { return NewVector() }

func NewVector(init ...Expression) ColVec {
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

func ConVector(vec Vectorized, args ...Expression) ColVec {
	return NewVector(append(vec.Slice(), args...)...)
}

func AppendVectors(vec Vectorized, args ...Expression) ColVec {
	return NewVector(append(vec.Slice(), args...)...)
}

func AppendArgToVector(init ...Expression) ColVec {
	return func(args ...Expression) []Expression {
		return append(init, args...)
	}
}

func (v ColVec) Len() int            { return len(v()) }
func (v ColVec) Vector() ColVec      { return v }
func (v ColVec) Slice() []Expression { return v() }

func (v ColVec) Append(args ...Expression) Consumeable { return v.Con(args...) }

func (v ColVec) Prepend(args ...Expression) ColVec {
	return NewVector(append(args, v()...)...)
}

func (v ColVec) Reverse(args ...Expression) ColVec {
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
func (v ColVec) TypeFnc() TyFnc   { return Vector }
func (v ColVec) TypeNat() d.TyNat { return d.Function }
func (v ColVec) Type() TyPattern  { return Def(Vector, v.TypeElem()) }
func (v ColVec) TypeElem() d.Typed {
	if v.Len() > 0 {
		return v.Head().Type()
	}
	return None.Type()
}

func (v ColVec) Con(args ...Expression) ColVec {
	return ConVector(v, args...)
}

func (v ColVec) Call(d ...Expression) Expression {
	return NewVector(v(d...)...)
}

func (v ColVec) Last() Expression {
	if v.Len() > 0 {
		return v()[v.Len()-1]
	}
	return nil
}

func (v ColVec) Head() Expression {
	if v.Len() > 0 {
		return v()[0]
	}
	return nil
}

func (v ColVec) Tail() Consumeable {
	if v.Len() > 1 {
		return NewVector(v()[1:]...)
	}
	return NewEmptyVector()
}

func (v ColVec) Consume() (Expression, Consumeable) {
	return v.Head(), v.Tail()
}

func (v ColVec) TailVec() ColVec {
	if v.Len() > 1 {
		return NewVector(v.Tail().(ColVec)()...)
	}
	return NewEmptyVector()
}

func (v ColVec) ConsumeVec() (Expression, ColVec) {
	return v.Head(), v.TailVec()
}

func (v ColVec) Clear() ColVec { return NewVector() }

func (v ColVec) Empty() bool {
	if len(v()) > 0 {
		for _, val := range v() {
			if !val.TypeFnc().Flag().Match(None) {
				return false
			}
		}
	}
	return true
}
func (v ColVec) Get(i int) (Expression, bool) {
	if i < v.Len() {
		return v()[i], true
	}
	return NewNone(), false
}

func (v ColVec) Set(i int, val Expression) (Vectorized, bool) {
	if i < v.Len() {
		var slice = v()
		slice[i] = val
		return ColVec(
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
func NewEmptyPairVec() ColPairV {
	return ColPairV(func(args ...Paired) []Paired {
		var pairs = []Paired{}
		if len(args) > 0 {
			pairs = append(pairs, args...)
		}
		return pairs
	})
}

func NewPairVectorFromPairs(pairs ...Paired) ColPairV {
	return ColPairV(func(args ...Paired) []Paired {
		if len(args) > 0 {
			return append(pairs, args...)
		}
		return pairs
	})
}

func ConPairListFromArgs(rec ColPairV, args ...Expression) ColPairV {
	var pairs = []Paired{}
	for _, arg := range args {
		if pair, ok := arg.(Paired); ok {
			pairs = append(pairs, pair)
		}
	}
	return NewPairVectorFromPairs(append(rec(), pairs...)...)
}

func NewPairVec(args ...Paired) ColPairV {
	return NewPairVectorFromPairs(args...)
}

func ConPairVec(rec ColPairV, pairs ...Paired) ColPairV {
	return NewPairVectorFromPairs(append(rec(), pairs...)...)
}

func ConPairVecFromArgs(pvec ColPairV, args ...Expression) ColPairV {
	var pairs = pvec.Pairs()
	for _, arg := range args {
		if pair, ok := arg.(Paired); ok {
			pairs = append(pairs, pair)
		}
	}
	return ColPairV(func(args ...Paired) []Paired {
		if len(args) > 0 {
			return ConPairVec(pvec, args...)()
		}
		return append(pvec(), pairs...)
	})
}
func (v ColPairV) Len() int { return len(v()) }
func (v ColPairV) Type() TyPattern {
	return Def(Def(Vector, Pair), v.TypeElem())
}
func (v ColPairV) TypeFnc() TyFnc   { return Vector }
func (v ColPairV) TypeNat() d.TyNat { return d.Function }

func (v ColPairV) ConPairs(pairs ...Paired) ColPairV {
	return ConPairVec(v, pairs...)
}

func (v ColPairV) Con(args ...Expression) ColPairV {
	return ConPairVecFromArgs(v, args...)
}

func (v ColPairV) Append(args ...Expression) Consumeable { return v.Con(args...) }

func (v ColPairV) Consume() (Expression, Consumeable) {
	return v.Head(), v.Tail()
}

func (v ColPairV) ConsumePairVec() (Paired, ColPairV) {
	return v.HeadPair(), v.Tail().(ColPairV)
}

func (v ColPairV) Empty() bool {
	if len(v()) > 0 {
		for _, pair := range v() {
			if !pair.Empty() {
				return false
			}
		}
	}
	return true
}
func (v ColPairV) TypeElem() d.Typed {
	if v.Len() > 0 {
		v()[0].Type()
	}
	return None
}
func (v ColPairV) KeyType() d.Typed {
	if v.Len() > 0 {
		return v.Pairs()[0].Left().Type()
	}
	return None.TypeFnc()
}
func (v ColPairV) ValType() d.Typed {
	if v.Len() > 0 {
		return v.Pairs()[0].Right().Type()
	}
	return None.TypeFnc()
}
func (v ColPairV) Get(idx int) (Paired, bool) {
	if idx < v.Len()-1 {
		return v()[idx], true
	}
	return NewKeyPair("None", None), false
}

func (v ColPairV) Pairs() []Paired {
	var pairs = []Paired{}
	for _, pair := range v() {
		pairs = append(pairs, pair)
	}
	return pairs
}

func (v ColPairV) ConsumePair() (Paired, ConsumeablePaired) {
	var pairs = v()
	if len(pairs) > 0 {
		if len(pairs) > 1 {
			return pairs[0], NewPairVec(pairs[1:]...)
		}
		return pairs[0], NewPairVec()
	}
	return nil, NewPairVec()
}

func (v ColPairV) SwitchedPairs() []Paired {
	var switched = []Paired{}
	for _, pair := range v() {
		switched = append(
			switched,
			pair,
		)
	}
	return switched
}

func (v ColPairV) Slice() []Expression {
	var fncs = []Expression{}
	for _, pair := range v() {
		fncs = append(fncs, NewPair(pair.Left(), pair.Right()))
	}
	return fncs
}

func (v ColPairV) HeadPair() Paired {
	if v.Len() > 0 {
		return v()[0].(Paired)
	}
	return NewPair(NewNone(), NewNone())
}
func (v ColPairV) Head() Expression {
	if v.Len() > 0 {
		return v.Pairs()[0]
	}
	return nil
}

func (v ColPairV) TailPairs() ConsumeablePaired {
	if v.Len() > 1 {
		return NewPairVectorFromPairs(v.Pairs()[1:]...)
	}
	return NewEmptyPairVec()
}
func (v ColPairV) Tail() Consumeable {
	if v.Len() > 1 {
		return NewPairVectorFromPairs(v.Pairs()[1:]...)
	}
	return NewEmptyPairVec()
}

func (v ColPairV) Call(args ...Expression) Expression {
	return v.Con(args...)
}
