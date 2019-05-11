package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

// type system implementation
type (
	Apply  func(NaryExpr, ...Callable) Callable
	Bind   func(f, g Callable) Callable
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
	MonadicExpr    func(...Callable) (Callable, MonadicExpr)
)

//// CURRY
func Curry(f, g NaryExpr, args ...Callable) Callable {
	if len(args) > 0 {
		return f(g(args...))
	}
	return f(g())
}

//// CURRY-N
func CurryN(args ...Callable) Callable {
	if len(args) > 0 {
		var f = args[0].Call
		if len(args) > 1 {
			var g = args[1].Call
			if len(args) > 2 {
				return f(g(
					CurryN(
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

//// COLLECTION
func NewCollection(expr Callable) Collection {
	if expr.TypeFnc().Match(Consumeables) {
		return func(args ...Callable) (Callable, Collection) {
			return expr.Call(args...), NewCollection(expr)
		}
	}
	return func(args ...Callable) (Callable, Collection) {
		if len(args) > 0 {
			if len(args) > 1 {
				return expr.Call(args...), NewCollection(expr)
			}
			return expr.Call(args[0]), NewCollection(expr)
		}
		return expr, NewCollection(expr)
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

//// COLLECTION OF PAIRS
func NewPairCollection(expr Paired) PairCollection {
	return func(args ...Callable) (Callable, PairCollection) {
		var pair Callable
		var arg = expr.Call(args...)
		switch {
		case arg.TypeFnc().Match(Pair):
			if val, ok := arg.(Paired); ok {
				pair = val
			}

		case arg.TypeFnc().Match(Collections):
			if col, ok := arg.(Collection); ok {
				var left, right Callable
				left, arg = col.Consume()
				right, arg = col.Consume()
				pair = NewPair(left, right)
			}

		default:
			pair = NewPair(arg.TypeFnc(), arg)
		}
		return pair,
			NewPairCollection(expr)
	}
}

func (c PairCollection) Call(args ...Callable) Callable {
	var head, _ = c()
	if len(args) > 0 {
		if len(args) > 1 {
			return head.Call(args...)
		}
		return head.Call(args[0])
	}
	return head
}
func (c PairCollection) Eval(args ...d.Native) d.Native {
	var head, _ = c()
	if len(args) > 0 {
		if len(args) > 1 {
			return head.Eval(args...)
		}
		return head.Eval(args[0])
	}
	return head.Eval()
}
func (c PairCollection) Ident() Callable {
	return c
}
func (c PairCollection) Consume() (Callable, Consumeable) {
	return c.Head(), c.Tail()
}
func (c PairCollection) Head() Callable {
	h, _ := c()
	return h
}
func (c PairCollection) Tail() Consumeable {
	_, t := c()
	return t
}
func (c PairCollection) TypeFnc() TyFnc {
	return Functor | Pair | c.Head().TypeFnc()
}
func (c PairCollection) TypeNat() d.TyNat {
	return c.Head().TypeNat()
}
func (c PairCollection) String() string {
	return c.Head().String()
}

//// MONAD
func NewMonad(mon Consumeable, bind Bind) MonadicExpr {
	return MonadicExpr(func(args ...Callable) (Callable, MonadicExpr) {
		var head Callable
		head, mon = mon.Consume()
		if head != nil {
			return head, NewMonad(mon, bind)
		}
		return nil, NewMonad(mon, bind)
	})
}
func (m MonadicExpr) Consume() (Callable, Consumeable) {
	var head Callable
	head, m = m()
	if head == nil {
		return nil, m
	}
	return head, m
}
func (m MonadicExpr) Head() Callable {
	var head, _ = m()
	return head
}
func (m MonadicExpr) Tail() Consumeable {
	return m
}
func (m MonadicExpr) TypeFnc() TyFnc {
	return Monad
}
func (m MonadicExpr) TypeNat() d.TyNat {
	return m.Head().TypeNat()
}
func (m MonadicExpr) Call(args ...Callable) Callable {
	return m.Head().Call(args...)
}
func (m MonadicExpr) Eval(args ...d.Native) d.Native {
	return m.Head().Eval(args...)
}
func (m MonadicExpr) String() string {
	return m.Head().String()
}

//// MAP
func MapC(cons Consumeable, fmap Map) Collection {
	return Collection(func(args ...Callable) (Callable, Collection) {
		var head Callable
		head, cons = cons.Consume()
		if head == nil {
			return nil, NewCollection(cons)
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

//// BIND
func BindL(fl, gl ListVal, bind Bind) MonadicExpr {
	return MonadicExpr(func(args ...Callable) (Callable, MonadicExpr) {
		var f, g Callable
		f, fl = fl()
		g, gl = gl()
		if f != nil {
			if g != nil {
				return bind(f, g), BindF(fl, gl, bind)
			}
		}
		return nil, BindL(fl, gl, bind)
	})
}

func BindF(fa, ga Consumeable, bind Bind) MonadicExpr {
	return MonadicExpr(func(args ...Callable) (Callable, MonadicExpr) {
		var f, g Callable
		f, fa = fa.Consume()
		g, ga = ga.Consume()
		if f != nil {
			if g != nil {
				return bind(f, g), BindF(fa, ga, bind)
			}
		}
		return nil, BindF(fa, ga, bind)
	})
}

//// FOLD
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

//// FILTER FUNCTOR
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

//// ZIP
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
