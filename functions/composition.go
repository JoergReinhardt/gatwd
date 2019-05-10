package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

// type system implementation
type (
	Curry        func(Callable, ...Callable) Callable
	Apply        func(NaryExpr, ...Callable) Callable
	Map          func(...Callable) Callable
	MapPaired    func(...Callable) Paired
	Fold         func(Callable, Callable, ...Callable) Callable
	FoldPaired   func(Paired, Paired, ...Callable) Paired
	Filter       func(Callable, ...Callable) bool
	FilterPaired func(Paired, ...Callable) bool
	Zip          func(l, r Callable) Paired
	Split        func(Callable, ...Callable) Paired

	CollectionFnc func(...Callable) (Callable, CollectionFnc)
	PairedCollFnc func(...Callable) (Callable, PairedCollFnc)
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
func NewFunctor(expr Callable) CollectionFnc {
	if expr.TypeFnc().Match(Consumeables) {
		return func(args ...Callable) (Callable, CollectionFnc) {
			return expr.Call(args...), NewFunctor(expr)
		}
	}
	return func(args ...Callable) (Callable, CollectionFnc) {
		if len(args) > 0 {
			if len(args) > 1 {
				return expr.Call(args...), NewFunctor(expr)
			}
			return expr.Call(args[0]), NewFunctor(expr)
		}
		return expr, NewFunctor(expr)
	}
}

func (c CollectionFnc) Call(args ...Callable) Callable {
	var head, _ = c()
	if len(args) > 0 {
		if len(args) > 1 {
			return head.Call(args...)
		}
		return head.Call(args[0])
	}
	return head
}
func (c CollectionFnc) Eval(args ...d.Native) d.Native {
	var head, _ = c()
	if len(args) > 0 {
		if len(args) > 1 {
			return head.Eval(args...)
		}
		return head.Eval(args[0])
	}
	return head.Eval()
}
func (c CollectionFnc) Ident() Callable {
	return c
}
func (c CollectionFnc) Consume() (Callable, Consumeable) {
	return c.Head(), c.Tail()
}
func (c CollectionFnc) Head() Callable {
	h, _ := c()
	return h
}
func (c CollectionFnc) Tail() Consumeable {
	_, t := c()
	return t
}
func (c CollectionFnc) TypeFnc() TyFnc {
	return Functor | c.Head().TypeFnc()
}
func (c CollectionFnc) TypeNat() d.TyNat {
	return c.Head().TypeNat()
}
func (c CollectionFnc) String() string {
	return c.Head().String()
}

func MapC(cons Consumeable, fmap Map) CollectionFnc {
	return CollectionFnc(func(args ...Callable) (Callable, CollectionFnc) {
		// decapitate list to get head and list continuation
		var head Callable
		head, cons = cons.Consume()
		if head == nil { // return empty head
			return nil, NewFunctor(cons)
		}
		// return result of applying arguments to fmap and the
		// list continuation
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
		// decapitate list to get head and list continuation
		var head Callable
		head, list = list()
		if head == nil { // return empty head
			return nil, list
		}
		// return result of applying arguments to fmap and the
		// list continuation
		if len(args) > 0 {
			return mapf(head.Call(args...)), MapL(list, mapf)
		}
		return mapf(head), MapL(list, mapf)
	})
}

func MapF(fnc CollectionFnc, fmap Map) CollectionFnc {
	return CollectionFnc(func(args ...Callable) (Callable, CollectionFnc) {
		// decapitate list to get head and list continuation
		var head, tail = fnc()
		if head == nil { // return empty head
			return nil, tail
		}
		// return result of applying arguments to fmap and the
		// list continuation
		if len(args) > 0 {
			return fmap(head.Call(args...)),
				MapF(tail, fmap)
		}
		return fmap(head),
			MapF(tail, fmap)
	})
}

func MapP(pairs ConsumeablePairs, pmap MapPaired) PairedCollFnc {
	return PairedCollFnc(func(args ...Callable) (Callable, PairedCollFnc) {
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

func FoldF(cons Consumeable, elem Callable, fold Fold) CollectionFnc {
	return CollectionFnc(func(args ...Callable) (Callable, CollectionFnc) {
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

func FoldP(pairs ConsumeablePairs, elem Callable, fold Fold) PairedCollFnc {
	return PairedCollFnc(func(args ...Callable) (Callable, PairedCollFnc) {
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

func FilterF(cons Consumeable, filter Filter) CollectionFnc {
	return CollectionFnc(
		func(args ...Callable) (Callable, CollectionFnc) {
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

func FilterP(pairs ConsumeablePairs, filter Filter) PairedCollFnc {
	return PairedCollFnc(
		func(args ...Callable) (Callable, PairedCollFnc) {
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

func ZipF(lcons, rcons Consumeable, zip Zip) CollectionFnc {
	return CollectionFnc(
		func(args ...Callable) (Callable, CollectionFnc) {
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
