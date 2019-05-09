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
	BindExpr func(fm, gm Consumeable, args ...Consumeable) Consumeable

	/// FUNCTORS
	// all functors can be applyed to map & fold by implementing the
	// consumeable interface. that renders all consumeables to be functors
	FunctorVal func(...Callable) (Callable, Consumeable)

	/// APPLICAPLE
	// applicables enclose a functor value and an apply function that is
	// called and passed the functor and optional arguments to, whenever
	// the applicable value is evalueated, or called
	ApplicableVal func(...Callable) (Callable, ApplicableVal)

	/// MONADIC
	// monadic values provide mappings between two or more functor types by
	// taking functors as arguments and returning a functor value as result
	// and a new instance of the monadic value type to compute the next
	// result from.
	MonadicVal func(...Consumeable) (Consumeable, MonadicVal)
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
// new functor encloses a flat callable expression to implement consumeable so
// that it can be mapped over to return new results depending solely on the
// passed arguments for each consequtive call. the wrapping is ommited, should
// the passed expression implement the consumeable interface already and the
// expression will be type asserted and returned instead.
func NewFunctor(expr Callable) FunctorVal {
	if expr.TypeFnc().Match(Functors) {
		return func(args ...Callable) (Callable, Consumeable) {
			return expr.Call(args...), expr.(Consumeable)
		}
	}
	return func(args ...Callable) (Callable, Consumeable) {
		if len(args) > 0 {
			if len(args) > 1 {
				return expr.Call(args...), NewFunctor(expr)
			}
			return expr.Call(args[0]), NewFunctor(expr)
		}
		return expr, NewFunctor(expr)
	}
}

func (c FunctorVal) Call(args ...Callable) Callable {
	var head, _ = c()
	if len(args) > 0 {
		if len(args) > 1 {
			return head.Call(args...)
		}
		return head.Call(args[0])
	}
	return head
}
func (c FunctorVal) Eval(args ...d.Native) d.Native {
	var head, _ = c()
	if len(args) > 0 {
		if len(args) > 1 {
			return head.Eval(args...)
		}
		return head.Eval(args[0])
	}
	return head.Eval()
}
func (c FunctorVal) Ident() Callable                  { return c }
func (c FunctorVal) Consume() (Callable, Consumeable) { return c() }
func (c FunctorVal) Head() Callable                   { h, _ := c(); return h }
func (c FunctorVal) Tail() Consumeable                { _, t := c(); return t }
func (c FunctorVal) TypeFnc() TyFnc                   { return Functor | c.Head().TypeFnc() }
func (c FunctorVal) TypeNat() d.TyNat                 { return c.Head().TypeNat() }
func (c FunctorVal) String() string                   { return c.Head().String() }

// APPLY FUNCTION

// APPLICATIVE
// appliccable encloses over a consumeable-/ and an apply expression. whenn
// called, expression and optional arguments are passed to the apply function
// and the yielded result will be returned
func NewApplicable(cons Consumeable, apply ApplyF) ApplicableVal {
	return func(args ...Callable) (Callable, ApplicableVal) {
		var head, tail = cons.Consume()
		if head == nil {
			return nil, NewApplicable(tail, apply)
		}
		if len(args) > 0 {
			if len(args) > 1 {
				return apply(head.Call, args...),
					NewApplicable(tail, apply)
			}
			return apply(head.Call, args[0]),
				NewApplicable(tail, apply)
		}
		return apply(head.Call),
			NewApplicable(tail, apply)

	}
}

func (c ApplicableVal) Call(args ...Callable) Callable {
	var head Callable
	if len(args) > 0 {
		if len(args) > 1 {
			head, _ = c(args...)
			return head
		}
		head, _ = c(args[0])
		return head
	}
	head, _ = c()
	return head
}
func (c ApplicableVal) Eval(args ...d.Native) d.Native {
	var head, _ = c()
	if len(args) > 0 {
		if len(args) > 1 {
			return head.Eval(args...)
		}
		return head.Eval(args[0])
	}
	return head.Eval()
}
func (c ApplicableVal) Ident() Callable                  { return c }
func (c ApplicableVal) Consume() (Callable, Consumeable) { return c() }
func (c ApplicableVal) Head() Callable                   { h, _ := c(); return h }
func (c ApplicableVal) Tail() Consumeable                { _, t := c(); return t }
func (c ApplicableVal) TypeFnc() TyFnc                   { return Applicable | c.Head().TypeFnc() }
func (c ApplicableVal) TypeNat() d.TyNat                 { return c.Head().TypeNat() }
func (c ApplicableVal) String() string                   { return c.Head().String() }

// MONADIC
func NewMonad(cons Consumeable, bind BindExpr) MonadicVal {
	return func(monargs ...Consumeable) (Consumeable, MonadicVal) {
		if len(monargs) > 0 {
			if len(monargs) > 0 {
				bind(cons, monargs[0], monargs[1:]...)
			}
			bind(cons, monargs[0])
		}
		var head, tail = cons.Consume()
		if head.TypeFnc().Match(Functors) {
			return head.(Consumeable), NewMonad(tail, bind)
		}
		var result = NewFunctor(head)
		return result, NewMonad(tail, bind)
	}
}

func (c MonadicVal) Call(args ...Callable) Callable   { return c.Head().Call(args...) }
func (c MonadicVal) Eval(args ...d.Native) d.Native   { return c.Head().Eval(args...) }
func (c MonadicVal) String() string                   { return c.Head().String() }
func (c MonadicVal) Ident() Callable                  { return c }
func (c MonadicVal) Consume() (Callable, Consumeable) { return c() }
func (c MonadicVal) Head() Callable                   { h, _ := c(); return h }
func (c MonadicVal) Tail() Consumeable                { _, t := c(); return t }
func (c MonadicVal) TypeNat() d.TyNat                 { return c.Head().TypeNat() }
func (c MonadicVal) TypeFnc() TyFnc                   { return Monad | c.Head().TypeFnc() }

//// MAP FUNCTOR LATE BINDING
///
// expects a consumeable list and a mapping function to apply on each element.
// list elements are late bound per call to the resulting consumeable, passed
// arguments get concatenated to the yielded list element, when fmap is called.
func MapF(cons Consumeable, fmap MapFExpr) Consumeable {
	return FunctorVal(func(args ...Callable) (Callable, Consumeable) {
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

// bind m calls the bind expression with both monads as it's arguments and
// returns a single monadic result per call
func BindM(fm Consumeable, gm Consumeable, bind BindExpr) MonadicVal {
	return MonadicVal(
		func(args ...Consumeable) (Consumeable, MonadicVal) {
			if fm == nil || gm == nil {
				return nil, BindM(fm, gm, bind)
			}
			if len(args) > 0 {
				return bind(fm, gm, args...),
					BindM(fm, gm, bind)
			}
			return bind(fm, gm),
				BindM(fm, gm, bind)
		})
}

// bind l expects two lists f & g and a bind expression as its arguments & returns a monadic value.
func BindL(fm ListVal, gm ListVal, bind BindExpr) MonadicVal {
	return MonadicVal(
		func(args ...Consumeable) (Consumeable, MonadicVal) {
			var f, fm = fm()
			var g, gm = gm()
			if f == nil || g == nil {
				return nil, BindL(fm, gm, bind)
			}
			var fcon, gcon Consumeable
			if f.TypeFnc().Match(Functors) {
				fcon = f.(Consumeable)
			} else {
				fcon = NewFunctor(f)
			}
			if g.TypeFnc().Match(Functors) {
				gcon = g.(Consumeable)
			} else {
				gcon = NewFunctor(g)
			}
			return bind(fcon, gcon, args...),
				BindL(fm, gm, bind)
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
	return FunctorVal(func(args ...Callable) (Callable, Consumeable) {
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

func FilterF(cons Consumeable, filter FilterFExpr) FunctorVal {
	return FunctorVal(
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
	return FunctorVal(
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
