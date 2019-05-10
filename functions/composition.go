package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

// type system implementation
type (
	Curry  func(Callable, ...Callable) Callable
	Apply  func(NaryExpr, ...Callable) Callable
	Map    func(...Callable) Callable
	Fold   func(Callable, Callable, ...Callable) Callable
	Filter func(Callable, ...Callable) bool
	Zip    func(l, r Callable) Paired
	Split  func(Callable, ...Callable) Paired

	MapPaired    func(...Callable) Paired
	FoldPaired   func(Paired, Paired, ...Callable) Paired
	FilterPaired func(Paired, ...Callable) bool

	Collection     func(...Callable) (Callable, Collection)
	PairCollection func(...Callable) (Callable, PairCollection)
)

//// CURRY
func ConsCurry(f, g NaryExpr, args ...Callable) Callable {
	if len(args) > 0 {
		return f(g(args...))
	}
	return f(g())
}
func RecCurry(args ...Callable) Callable {
	if len(args) > 0 {
		var f = args[0].Call
		if len(args) > 1 {
			var g = args[1].Call
			if len(args) > 2 {
				return f(g(
					RecCurry(
						args[2:]...,
					),
				))
			}
			return f(g())
		}
		return f()
	}
	return NewNone()
}

// FUNCTOR
// new functor encloses a flat callable expression to implement consumeable so
// that it can be mapped over to return new results depending solely on the
// passed arguments for each consequtive call. the wrapping is ommited, should
// the passed expression implement the consumeable interface already and the
// expression will be type asserted and returned instead.
func NewFunctor(expr Callable) Collection {
	if expr.TypeFnc().Match(Consumeables) {
		return func(args ...Callable) (Callable, Collection) {
			return expr.Call(args...), NewFunctor(expr)
		}
	}
	return func(args ...Callable) (Callable, Collection) {
		if len(args) > 0 {
			if len(args) > 1 {
				return expr.Call(args...), NewFunctor(expr)
			}
			return expr.Call(args[0]), NewFunctor(expr)
		}
		return expr, NewFunctor(expr)
	}
}

func (c Collection) Call(args ...Callable) Callable {
	var head, _ = c()
	if len(args) > 0 {
		if len(args) > 1 {
			return head.Call(args...)
		}
		return head.Call(args[0])
	}
	return head
}
func (c Collection) Eval(args ...d.Native) d.Native {
	var head, _ = c()
	if len(args) > 0 {
		if len(args) > 1 {
			return head.Eval(args...)
		}
		return head.Eval(args[0])
	}
	return head.Eval()
}
func (c Collection) Ident() Callable {
	return c
}
func (c Collection) Consume() (Callable, Consumeable) {
	return c.Head(), c.Tail()
}
func (c Collection) Head() Callable {
	h, _ := c()
	return h
}
func (c Collection) Tail() Consumeable {
	_, t := c()
	return t
}
func (c Collection) TypeFnc() TyFnc {
	return Functor | c.Head().TypeFnc()
}
func (c Collection) TypeNat() d.TyNat {
	return c.Head().TypeNat()
}
func (c Collection) String() string {
	return c.Head().String()
}

func MapC(cons Consumeable, fmap Map) Collection {
	return Collection(func(args ...Callable) (Callable, Collection) {
		var head Callable
		head, cons = cons.Consume()
		if head == nil {
			return nil, NewFunctor(cons)
		}
		if len(args) > 0 {
			return fmap(head.Call(args...)),
				MapC(cons, fmap)
		}
		return fmap(head),
			MapC(cons, fmap)
	})
}

func MapL(list ListVal, mapf Map) ListVal {
	return ListVal(func(args ...Callable) (Callable, ListVal) {
		var head Callable
		head, list = list()
		if head == nil {
			return nil, list
		}
		if len(args) > 0 {
			return mapf(head.Call(args...)), MapL(list, mapf)
		}
		return mapf(head), MapL(list, mapf)
	})
}

func MapF(fnc Collection, fmap Map) Collection {
	return Collection(func(args ...Callable) (Callable, Collection) {
		var head, tail = fnc()
		if head == nil {
			return nil, tail
		}
		if len(args) > 0 {
			return fmap(head.Call(args...)),
				MapF(tail, fmap)
		}
		return fmap(head),
			MapF(tail, fmap)
	})
}

func MapP(pairs ConsumeablePairs, pmap MapPaired) PairCollection {
	return PairCollection(func(args ...Callable) (Callable, PairCollection) {
		// decapitate list to get head and list continuation
		var pair Paired
		pair, pairs = pairs.ConsumePair()
		if pair == nil { // return empty head
			return nil, MapP(pairs, pmap)
		}
		// return result of applying arguments to fmap and the
		// list continuation
		if len(args) > 0 {
			return pmap(pair.Call(args...)),
				MapP(pairs, pmap)
		}
		return pmap(pair),
			MapP(pairs, pmap)
	})
}

func FoldL(list ListVal, elem Callable, fold Fold) ListVal {
	return ListVal(func(args ...Callable) (Callable, ListVal) {
		var head, tail = list()
		if head == nil {
			return nil, list
		}
		if len(args) > 0 {
			elem = fold(elem, head, args...)
			return elem, FoldL(tail, elem, fold)
		}
		elem = fold(elem, head)
		return elem, FoldL(tail, elem, fold)
	})
}

func FoldF(cons Consumeable, elem Callable, fold Fold) Collection {
	return Collection(func(args ...Callable) (Callable, Collection) {
		var head Callable
		head, cons = cons.Consume()
		if head == nil {
			return nil, FoldF(cons, elem, fold)
		}
		if len(args) > 0 {
			elem = fold(elem, head, args...)
			return elem, FoldF(cons, elem, fold)
		}
		elem = fold(elem, head)
		return elem, FoldF(cons, elem, fold)
	})
}

func FoldP(pairs ConsumeablePairs, elem Callable, fold Fold) PairCollection {
	return PairCollection(func(args ...Callable) (Callable, PairCollection) {
		var pair Paired
		pair, pairs = pairs.ConsumePair()
		if pair == nil {
			return nil, FoldP(pairs, elem, fold)
		}
		if len(args) > 0 {
			elem = fold(elem, pair, args...)
			return elem, FoldP(pairs, elem, fold)
		}
		elem = fold(elem, pair)
		return elem, FoldP(pairs, elem, fold)
	})
}

// FILTER FUNCTOR LATE BINDING
func FilterL(list ListVal, filter Filter) ListVal {
	return ListVal(
		func(args ...Callable) (Callable, ListVal) {
			var head, tail = list()
			if head == nil {
				return nil, list
			}
			// filter either returns true & head is returned, or
			// FilterL will be called recursively
			if !filter(head, args...) {
				return FilterL(tail, filter)(args...)
			}
			return head, FilterL(tail, filter)
		})
}

func FilterF(cons Consumeable, filter Filter) Collection {
	return Collection(
		func(args ...Callable) (Callable, Collection) {
			var head, tail = cons.Consume()
			if head == nil {
				return nil, FilterF(cons, filter)
			}
			if !filter(head, args...) {
				return FilterF(tail, filter)(args...)
			}
			return head, FilterF(tail, filter)
		})
}

func FilterP(pairs ConsumeablePairs, filter Filter) PairCollection {
	return PairCollection(
		func(args ...Callable) (Callable, PairCollection) {
			var pair Paired
			pair, pairs = pairs.ConsumePair()
			if pair == nil {
				return nil, FilterP(pairs, filter)
			}
			if !filter(pair, args...) {
				return FilterP(pairs, filter)(args...)
			}
			return pair, FilterP(pairs, filter)
		})
}

func ZipL(llist, rlist ListVal, zip Zip) ListVal {
	return ListVal(
		func(args ...Callable) (Callable, ListVal) {
			var lhead, ltail = llist()
			var rhead, rtail = rlist()
			if lhead == nil || rhead == nil {
				return nil, ZipL(llist, rlist, zip)
			}
			if len(args) > 0 {
				return zip(lhead, rhead).Call(args...), ZipL(ltail, rtail, zip)
			}
			return zip(lhead, rhead), ZipL(ltail, rtail, zip)
		})
}

func ZipF(lcons, rcons Consumeable, zip Zip) Collection {
	return Collection(
		func(args ...Callable) (Callable, Collection) {
			var lhead, ltail = lcons.Consume()
			var rhead, rtail = rcons.Consume()
			if lhead == nil || rhead == nil {
				return nil,
					ZipF(lcons, rcons, zip)
			}
			if len(args) > 0 {
				return zip(lhead, rhead).Call(args...),
					ZipF(ltail, rtail, zip)
			}
			return zip(lhead, rhead),
				ZipF(ltail, rtail, zip)
		})
}
