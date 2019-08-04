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

	ColVal func(...Expression) (Expression, Consumeable)
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

func (l ColList) takeN(n int) ColList {
	var init ColVec
	var acc = DeclareFunction(func(args ...Expression) Expression {
		if len(args) > 0 {
			init = args[0].(ColVec)
			if init.Len() < n {
				if len(args) > 1 {
					return init.Con(args[1:]...)
				}
			}
		}
		return init
	}, Def(Vector))
	return l.FoldL(acc, NewVector())
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
