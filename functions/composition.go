package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

// type system implementation
type (
	/// CURRY FUNCTION
	Curry func(...Callable) Callable

	/// APPLY FUNCTION
	ApplyF func(NaryExpr, ...Callable) Callable
	ApplyP func(NaryExpr, ...Paired) Paired

	/// MAP FUNCTION
	MapFExpr func(...Callable) Callable
	MapPExpr func(...Paired) Paired

	/// FOLD FUNCTION
	FoldFExpr func(Callable, Callable, ...Callable) Callable

	/// FILTER FUNCTION
	FilterFExpr func(Callable, ...Callable) bool
	FilterPExpr func(Paired, ...Paired) bool

	/// ZIP FUNCTION
	ZipExpr func(l, r Callable) Paired

	/// SPLIT FUNCTION
	SplitExpr func(Callable, ...Callable) Paired

	/// BIND
	// bind operator (>>=) binds the return value of one monad to be the
	// argument of another
	BindFExpr func(f, g Callable, args ...Callable) Consumeable
	BindMExpr func(fm, gm Consumeable, args ...Callable) Consumeable

	/// FUNCTOR
	// all functors apply to map-, foldl
	FunctorCons func(...Callable) (Callable, Consumeable)

	/// APPLICAPLE
	// applicables are functors to be applyd on a list of boxed values
	ApplicableCons func(...Callable) (Callable, Consumeable)

	/// MONADIC
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
// functor encloses a simple expression to implement consumeable so that it can
// be mapped over.
func NewFunctor(expr Callable) FunctorCons {
	return func(args ...Callable) (Callable, Consumeable) {
		return expr.Call(args...), NewFunctor(expr)
	}
}

func (c FunctorCons) Call(args ...Callable) Callable   { h, _ := c(args...); return h }
func (c FunctorCons) Eval(args ...d.Native) d.Native   { return c.Head().Eval(args...) }
func (c FunctorCons) Ident() Callable                  { return c }
func (c FunctorCons) Consume() (Callable, Consumeable) { return c() }
func (c FunctorCons) Head() Callable                   { h, _ := c(); return h }
func (c FunctorCons) Tail() Consumeable                { _, t := c(); return t }
func (c FunctorCons) TypeFnc() TyFnc                   { return Functor | c.Head().TypeFnc() }
func (c FunctorCons) TypeNat() d.TyNat                 { return c.Head().TypeNat() }
func (c FunctorCons) String() string                   { return c.Head().String() }

// APPLICATIVE
// functor that encloses an functor and a function that knows how to apply
// passed arguments to that expression.
func NewApplicable(cons Consumeable, apply ApplyF) ApplicableCons {
	return func(args ...Callable) (Callable, Consumeable) {
		var head, tail = cons.Consume()
		if head == nil {
			return nil, NewApplicable(tail, apply)
		}
		var result = apply(head.Call, args...)
		return result, NewApplicable(tail, apply)
	}
}

func (c ApplicableCons) Call(args ...Callable) Callable   { h, _ := c(args...); return h }
func (c ApplicableCons) Eval(args ...d.Native) d.Native   { return c.Head().Eval(args...) }
func (c ApplicableCons) Ident() Callable                  { return c }
func (c ApplicableCons) Consume() (Callable, Consumeable) { return c() }
func (c ApplicableCons) Head() Callable                   { h, _ := c(); return h }
func (c ApplicableCons) Tail() Consumeable                { _, t := c(); return t }
func (c ApplicableCons) TypeFnc() TyFnc                   { return Applicable | c.Head().TypeFnc() }
func (c ApplicableCons) TypeNat() d.TyNat                 { return c.Head().TypeNat() }
func (c ApplicableCons) String() string                   { return c.Head().String() }

// MONADIC
func NewMonad(cons Consumeable) MonadicCons {
	return func(args ...Callable) (Callable, Consumeable) {
		var head, tail = cons.Consume()
		if head == nil {
			return nil, NewNone()
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

func (c MonadicCons) Call(args ...Callable) Callable   { h, _ := c(args...); return h }
func (c MonadicCons) Eval(args ...d.Native) d.Native   { return c.Head().Eval(args...) }
func (c MonadicCons) String() string                   { return c.Head().String() }
func (c MonadicCons) Ident() Callable                  { return c }
func (c MonadicCons) Consume() (Callable, Consumeable) { return c() }
func (c MonadicCons) Head() Callable                   { h, _ := c(); return h }
func (c MonadicCons) Tail() Consumeable                { _, t := c(); return t }
func (c MonadicCons) TypeNat() d.TyNat                 { return c.Head().TypeNat() }
func (c MonadicCons) TypeFnc() TyFnc                   { return Monad | c.Head().TypeFnc() }

//// MAP FUNCTOR LATE BINDING
///
// expects a consumeable list and a mapping function to apply on each element.
// list elements are late bound per call to the resulting consumeable, passed
// arguments get concatenated to the yielded list element, when fmap is called.
func MapF(cons Consumeable, fmap MapFExpr) Consumeable {
	return FunctorCons(func(args ...Callable) (Callable, Consumeable) {
		// decapitate list to get head and list continuation
		var head, tail = cons.Consume()
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

// bind f consumes the heads of each arguments and returns some value and a
// remaining consumeable at each call
func BindF(f Callable, g Callable, bind BindFExpr) Consumeable {
	return FunctorCons(
		func(args ...Callable) (Callable, Consumeable) {
			return bind(f, g, args...),
				BindF(f, g, bind)
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
			return bind(f, g, args...),
				BindL(fm, gm, bind)
		})
}

// bind m calls the bind expression with both monads as it's arguments and
// returns a single monadic result per call
func BindM(fm Consumeable, gm Consumeable, bind BindMExpr) MonadicCons {
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
		var head, tail = cons.Consume()
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
			// filter either returns true & head is returned, or
			// FilterL will be called recursively
			if !filter(head, args...) {
				return FilterL(tail, filter)(args...)
			}
			return head, FilterL(tail, filter)
		})
}

func FilterF(cons Consumeable, filter FilterFExpr) FunctorCons {
	return FunctorCons(
		func(args ...Callable) (Callable, Consumeable) {
			var head, tail = cons.Consume()
			if head == nil {
				return nil, cons
			}
			if !filter(head, args...) {
				return FilterF(tail, filter)(args...)
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
