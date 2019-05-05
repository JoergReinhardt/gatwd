package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

// type system implementation
type (
	//// CURRY FUNCTION
	///
	Curry func(...Callable) Callable

	//// APPLY FUNCTION
	///
	ApplyF func(NaryExpr, ...Callable) Callable
	ApplyP func(NaryExpr, ...Paired) Paired

	//// MAP FUNCTION
	///
	MapFExpr func(...Callable) Callable
	MapPExpr func(...Paired) Paired

	//// FOLD FUNCTION
	///
	FoldFExpr func(Callable, Callable, ...Callable) Callable

	//// FILTER FUNCTION
	///
	FilterFExpr func(Callable, ...Callable) bool
	FilterPExpr func(Paired, ...Paired) bool

	//// ZIP FUNCTION
	///
	ZipExpr func(l, r Callable) Paired

	//// SPLIT FUNCTION
	///
	SplitFExpr func(Consumeable) (Paired, Consumeable)
	SplitLExpr func(ListVal) (Paired, ListVal)

	//// BIND
	///
	// bind operator (>>=) binds the return value of one monad to be the
	// argument of another
	BindMExpr func(fm, gm MonadicCons) MonadicCons
	BindFExpr func(f, g Callable) Consumeable

	//// FUNCTOR
	///
	// all functors apply to map-, foldl
	FunctorCons func(...Callable) (Callable, Consumeable)

	//// APPLICAPLE
	///
	// applicables are functors to be applyd on a list of boxed values
	PairFunctorCons func(args ...Paired) (Callable, PairFunctorCons)

	//// MONADIC
	///
	// monadic functions provide transformations between two functor types
	MonadicCons func(...Callable) (Callable, Consumeable)
)

//// CURRY
func ConsCurry(exprs ...Callable) Callable {
	if len(exprs) == 0 {
		return NaryExpr(ConsCurry)
	}
	if len(exprs) == 1 {
		return exprs[0]
	}
	return exprs[0].Call(ConsCurry(exprs[1:]...))
}

// FUNCTOR
// evaluats list elements by applying passed parameters lazy, to generate the
// new list on demand
func NewFunctor(cons Consumeable) FunctorCons {
	return func(args ...Callable) (Callable, Consumeable) {
		var head, tail = cons.DeCap()
		if head == nil {
			return nil, NewFunctor(tail)
		}
		if len(args) > 0 {
			if len(args) > 1 {
				return head.Call(args...), NewFunctor(tail)
			}
			return head.Call(args[0]), NewFunctor(tail)
		}
		return head.Call(), NewFunctor(tail)
	}
}

func (c FunctorCons) Call(args ...Callable) Callable { h, _ := c(args...); return h }
func (c FunctorCons) Eval(args ...d.Native) d.Native { return c.Head().Eval(args...) }

func (c FunctorCons) Ident() Callable                { return c }
func (c FunctorCons) DeCap() (Callable, Consumeable) { return c() }
func (c FunctorCons) Head() Callable                 { h, _ := c(); return h }
func (c FunctorCons) Tail() Consumeable              { _, t := c(); return t }
func (c FunctorCons) TypeFnc() TyFnc                 { return Functor | c.Head().TypeFnc() }
func (c FunctorCons) TypeNat() d.TyNat               { return c.Head().TypeNat() }
func (c FunctorCons) String() string                 { return c.Head().String() }

// MONADIC
func NewMonad(cons Consumeable) MonadicCons {
	return func(args ...Callable) (Callable, Consumeable) {
		var head, tail = cons.DeCap()
		if head == nil {
			return nil, NewMonad(tail)
		}
		if len(args) > 0 {
			if len(args) > 1 {
				return head.Call(args...), NewMonad(tail)
			}
			return head.Call(args[0]), NewMonad(tail)
		}
		return head.Call(), NewMonad(tail)
	}
}

func (c MonadicCons) Call(args ...Callable) Callable { h, _ := c(args...); return h }
func (c MonadicCons) Eval(args ...d.Native) d.Native { return c.Head().Eval(args...) }
func (c MonadicCons) String() string                 { return c.Head().String() }
func (c MonadicCons) Ident() Callable                { return c }
func (c MonadicCons) DeCap() (Callable, Consumeable) { return c() }
func (c MonadicCons) Head() Callable                 { h, _ := c(); return h }
func (c MonadicCons) Tail() Consumeable              { _, t := c(); return t }
func (c MonadicCons) TypeNat() d.TyNat               { return c.Head().TypeNat() }
func (c MonadicCons) TypeFnc() TyFnc                 { return Monad | c.Head().TypeFnc() }

/// PAIR FUNCTOR
// applicative encloses over a function to be applyd the head element and any
// arguments given at each call

func ApplyPairs(apply ApplyP, records ...Paired) PairFunctorCons {
	return func(args ...Paired) (Callable, PairFunctorCons) {
		var pair Paired
		var pairs = []Paired{}
		if len(records) > 0 {
			pair = records[0]
			if len(records) > 1 {
				pairs = records[1:]
			}
		}
		if pair == nil {
			return nil, NewPairFunctorFromPairs(pairs...)
		}
		if len(args) > 0 {
			return apply(pair.Call), ApplyPairs(apply, pairs...)
		}
		return apply(pair.Call), ApplyPairs(apply, pairs...)
	}
}

func NewPairFunctorFromPairs(pairs ...Paired) PairFunctorCons {
	return func(args ...Paired) (Callable, PairFunctorCons) {
		var pair Paired
		if len(pairs) > 0 {
			if len(pairs) > 1 {
				pair, pairs = pairs[0], pairs[1:]
			}
		}
		if pair == nil {
			return nil, NewPairFunctorFromPairs(pairs...)
		}
		if len(args) > 0 {
			return pair, NewPairFunctorFromPairs(append(pairs, args...)...)
		}
		return pair, NewPairFunctorFromPairs(pairs...)
	}
}

func NewPairFunctorFromList(pl ListVal) PairFunctorCons {
	return func(args ...Paired) (Callable, PairFunctorCons) {
		var pair, list = pl()
		if pair == nil {
			return nil, NewPairFunctorFromList(list)
		}
		if len(args) > 0 {
			for _, arg := range args {
				if pair, ok := arg.(Paired); ok {
					list = list.Cons(pair)
				}
			}
			return pair.(Paired), NewPairFunctorFromList(list)
		}
		return pair.(Paired), NewPairFunctorFromList(list)
	}
}

func (c PairFunctorCons) Ident() Callable                { return c }
func (c PairFunctorCons) DeCap() (Callable, Consumeable) { return c() }
func (c PairFunctorCons) Head() Callable                 { h, _ := c(); return h }
func (c PairFunctorCons) Tail() Consumeable              { _, t := c(); return t }
func (c PairFunctorCons) Call(args ...Callable) Callable { return c.Head().Call(args...) }
func (c PairFunctorCons) Eval(args ...d.Native) d.Native { return c.Head().Eval(args...) }
func (c PairFunctorCons) TypeFnc() TyFnc                 { return Applicable | c.Head().TypeFnc() }
func (c PairFunctorCons) TypeNat() d.TyNat               { return c.Head().TypeNat() }
func (c PairFunctorCons) String() string                 { return c.Head().String() }

//// MAP FUNCTOR LATE BINDING
///
// expects a consumeable list and a mapping function to apply on each element.
// list elements are late bound per call to the resulting consumeable, passed
// arguments get concatenated to the yielded list element, when fmap is called.
func MapL(list ListVal, mapf MapFExpr) ListVal {
	return ListVal(func(args ...Callable) (Callable, ListVal) {
		// decapitate list to get head and list continuation
		var head, tail = list()
		if head == nil { // return empty head
			return nil, list
		}
		// return result of applying arguments to fmap and the
		// list continuation
		if len(args) > 0 {
			return mapf(head).Call(args...), MapL(tail, mapf)
		}
		return mapf(head), MapL(tail, mapf)
	})
}

func MapF(cons Consumeable, fmap MapFExpr) Consumeable {
	return FunctorCons(func(args ...Callable) (Callable, Consumeable) {
		// decapitate list to get head and list continuation
		var head, tail = cons.DeCap()
		if head == nil { // return empty head
			return nil, cons
		}
		// return result of applying arguments to fmap and the
		// list continuation
		if len(args) > 0 {
			return fmap(head).Call(args...),
				MapF(tail, fmap)
		}
		return fmap(head),
			MapF(tail, fmap)
	})
}

func MapP(appl PairFunctorCons, mapa MapPExpr) PairFunctorCons {
	return PairFunctorCons(func(args ...Paired) (Callable, PairFunctorCons) {
		var head Callable
		head, appl = appl()
		if head == nil { // return empty head
			return nil, appl
		}
		var pair Paired
		if p, ok := head.(Paired); ok {
			pair = p
		}
		// return result of applying arguments to fmap and the
		// list continuation
		if len(args) > 0 {
			var pairs = []Paired{}
			for _, arg := range args {
				if pair, ok := arg.(Paired); ok {
					pairs = append(pairs, pair)
				}
			}
			return mapa(append(pairs, pair)...), MapP(appl, mapa)
		}
		return mapa(pair), MapP(appl, mapa)
	})
}

// just like bind-f but works on recursive lists exclusively
func BindL(fm ListVal, gm ListVal, bind BindFExpr) ListVal {
	return ListVal(
		func(args ...Callable) (Callable, ListVal) {
			var f, fm = fm()
			var g, gm = gm()
			if f == nil || g == nil {
				return nil, BindL(fm, gm, bind)
			}
			if len(args) > 0 {
				return bind(f, g).Call(args...),
					BindL(fm, gm, bind)
			}
			return bind(f, g),
				BindL(fm, gm, bind)
		})
}

// bind f consumes the heads of each arguments and returns some value and a
// remaining consumeable at each call
func BindF(fm Consumeable, gm Consumeable, bind BindFExpr) Consumeable {
	return FunctorCons(
		func(args ...Callable) (Callable, Consumeable) {
			var f, fm = fm.DeCap()
			var g, gm = gm.DeCap()
			if f == nil || g == nil {
				return nil, BindF(fm, gm, bind)
			}
			if len(args) > 0 {
				return bind(f, g).Call(args...),
					BindF(fm, gm, bind)
			}
			return bind(f, g),
				BindF(fm, gm, bind)
		})
}

// bind m calls the bind expression with both monads as it's arguments and
// returns a single monadic result per call
func BindM(fm MonadicCons, gm MonadicCons, bind BindMExpr) MonadicCons {
	return MonadicCons(
		func(args ...Callable) (Callable, Consumeable) {
			if fm == nil || gm == nil {
				return nil, BindM(fm, gm, bind)
			}
			if len(args) > 0 {
				return bind(fm, gm).Call(args...),
					BindM(fm, gm, bind)
			}
			return bind(fm, gm),
				BindM(fm, gm, bind)
		})
}

// FOLD FUNCTOR LATE BINDING
//
// returns a list of continuations, yielding accumulated result & list of
// follow-up continuations. when the list is depleted, return result only.
func FoldL(list ListVal, elem Callable, fold FoldFExpr) ListVal {
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

func FoldF(cons Consumeable, elem Callable, fold FoldFExpr) Consumeable {
	return FunctorCons(func(args ...Callable) (Callable, Consumeable) {
		var head, tail = cons.DeCap()
		if head == nil {
			return nil, cons
		}
		if len(args) > 0 {
			elem = fold(elem, head, args...)
			return elem, FoldF(tail, elem, fold)
		}
		elem = fold(elem, head)
		return elem, FoldF(tail, elem, fold)
	})
}

// FILTER FUNCTOR LATE BINDING
func FilterL(list ListVal, filter FilterFExpr) ListVal {
	return ListVal(
		func(args ...Callable) (Callable, ListVal) {
			var head, tail = list()
			if head == nil {
				return nil, list
			}
			if !filter(head, args...) {
				return head.Call(args...), FilterL(tail, filter)
			}

			return head, FilterL(tail, filter)
		})
}

func FilterF(cons Consumeable, filter FilterFExpr) Consumeable {
	return FunctorCons(
		func(args ...Callable) (Callable, Consumeable) {
			var head, tail = cons.DeCap()
			if head == nil {
				return nil, cons
			}
			if !filter(head, args...) {
				return head.Call(args...), FilterF(tail, filter)
			}

			return head, FilterF(tail, filter)
		})
}

func ZipL(llist, rlist ListVal, zip ZipExpr) ListVal {
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

func ZipF(lcons, rcons Consumeable, zip ZipExpr) Consumeable {
	return FunctorCons(
		func(args ...Callable) (Callable, Consumeable) {
			var lhead, ltail = lcons.DeCap()
			var rhead, rtail = rcons.DeCap()
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

// split takes the whole list and might remove multiple elements to satisfy
// split fnc
func SplitL(list ListVal, split SplitLExpr) ListVal {
	return ListVal(
		func(args ...Callable) (Callable, ListVal) {
			var p Paired
			p, list = split(list)
			if p == nil {
				return nil, SplitL(NewList(args...), split)
			}
			if len(args) > 0 {
				return p.Call(args...), SplitL(list, split)
			}
			return p, SplitL(list, split)
		})
}

func SplitF(cons Consumeable, split SplitFExpr) Consumeable {
	return FunctorCons(
		func(args ...Callable) (Callable, Consumeable) {
			var p Paired
			p, cons = split(cons)
			if p == nil {
				return nil, SplitF(cons, split)
			}
			if len(args) > 0 {
				return p.Call(args...), SplitF(cons, split)
			}
			return p, SplitF(cons, split)
		})
}
