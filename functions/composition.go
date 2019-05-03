package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

// type system implementation
type (
	//// CURRY FUNCTION
	///
	Curry func(...Callable) Callable
	//// MAP FUNCTION
	///
	MapFnc func(...Callable) Callable

	//// FOLD FUNCTION
	///
	FoldFnc func(Callable, Callable, ...Callable) Callable

	//// FILTER FUNCTION
	///
	FilterFnc func(Callable, ...Callable) bool

	//// ZIP FUNCTION
	///
	ZipFnc func(l, r Callable) PairVal

	//// SPLIT FUNCTION
	///
	SplitFnc func(ListVal) (PairVal, ListVal)

	//// APPLY FUNCTION
	///
	ApplyFnc func(NaryExpr, ...Callable) Callable

	//// BIND
	///
	// bind operator (>>=) binds the return value of one monad to be the
	// argument of another
	BindFnc func(fm, gm Callable) MonadicCon

	//// FUNCTOR
	///
	// all functors apply to map-, foldl
	FunctorCon func(...Callable) (Callable, Consumeable)

	//// APPLICAPLE
	///
	// applicables are functors to be applyd on a list of boxed values
	ApplicativeCon func(...Callable) (Callable, ListVal)

	//// MONADIC
	///
	// monadic functions provide transformations between two functor types
	MonadicCon func(...Callable) (Callable, Consumeable)
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

//// CONSUME
///
// consumes an epifunctor and passes its arguments through, when called

// FUNCTOR
// evaluats list elements by applying passed parameters lazy, to generate the
// new list on demand
func NewFunctor(list Consumeable) FunctorCon {

	return func(args ...Callable) (Callable, Consumeable) {

		var head, tail = list.DeCap()

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

func (c FunctorCon) Call(args ...Callable) Callable { h, _ := c(args...); return h }

func (c FunctorCon) Ident() Callable                { return c }
func (c FunctorCon) DeCap() (Callable, Consumeable) { return c() }
func (c FunctorCon) Head() Callable                 { h, _ := c(); return h }
func (c FunctorCon) Tail() Consumeable              { _, t := c(); return t }
func (c FunctorCon) Eval(args ...d.Native) d.Native { return c.Head().Eval(args...) }
func (c FunctorCon) TypeFnc() TyFnc                 { return Functor | c.Head().TypeFnc() }
func (c FunctorCon) TypeNat() d.TyNat               { return c.Head().TypeNat() }
func (c FunctorCon) String() string                 { return c.Head().String() }

/// APPLICATIVE
// applicative encloses over a function to be applyd the head element and any
// arguments given at each call

func NewApplicative(list ListVal, apply ApplyFnc) ApplicativeCon {
	return func(args ...Callable) (Callable, ListVal) {
		var head, tail = list()
		if head == nil {
			return nil, NewList(args...)
		}
		if len(args) > 0 {
			return apply(head.Call, args...), tail
		}
		return apply(head.Call), tail
	}
}
func ApplyArity(arity Arity, list ListVal, apply ApplyFnc) ApplicativeCon {
	return func(args ...Callable) (Callable, ListVal) {
		var l = int(arity)
		var params []Callable
		if len(args) > 0 {
			params = args[:l]
			args = args[l:]
		}
		var head, tail = list()
		if head == nil {
			return nil, NewList(args...)
		}
		return apply(head.Call, params...), tail.Cons(args...)
	}
}

func (c ApplicativeCon) Ident() Callable                { return c }
func (c ApplicativeCon) DeCap() (Callable, Consumeable) { return c() }
func (c ApplicativeCon) Head() Callable                 { h, _ := c(); return h }
func (c ApplicativeCon) Tail() Consumeable              { _, t := c(); return t }
func (c ApplicativeCon) Call(args ...Callable) Callable { return c.Head().Call(args...) }
func (c ApplicativeCon) Eval(args ...d.Native) d.Native { return c.Head().Eval(args...) }
func (c ApplicativeCon) TypeFnc() TyFnc                 { return Applicable | c.Head().TypeFnc() }
func (c ApplicativeCon) TypeNat() d.TyNat               { return c.Head().TypeNat() }
func (c ApplicativeCon) String() string                 { return c.Head().String() }

// MONADIC
func (c MonadicCon) Call(args ...Callable) Callable {
	var result Callable
	return result
}
func (c MonadicCon) String() string                 { return c.Head().String() }
func (c MonadicCon) Ident() Callable                { return c }
func (c MonadicCon) DeCap() (Callable, Consumeable) { return c() }
func (c MonadicCon) Head() Callable                 { h, _ := c(); return h }
func (c MonadicCon) Tail() Consumeable              { _, t := c(); return t }
func (c MonadicCon) Eval(args ...d.Native) d.Native { return c.Head().Eval(args...) }
func (c MonadicCon) TypeNat() d.TyNat               { return c.Head().TypeNat() }
func (c MonadicCon) TypeFnc() TyFnc                 { return Monad | c.Head().TypeFnc() }

//// MAP FUNCTOR LATE BINDING
///
// expects a consumeable list and a mapping function to apply on each element.
// list elements are late bound per call to the resulting consumeable, passed
// arguments get concatenated to the yielded list element, when fmap is called.
func MapF(list ListVal, fmap MapFnc) ListVal {
	return ListVal(func(args ...Callable) (Callable, ListVal) {
		// decapitate list to get head and list continuation
		var head, tail = list()
		if head == nil { // return empty head
			return nil, list
		}
		// return result of applying arguments to fmap and the
		// list continuation
		return fmap(append(args, head)...),
			MapF(tail, fmap)
	})
}

func BindF(mf, mg Consumeable, bind BindFnc) MonadicCon {
	return MonadicCon(
		func(args ...Callable) (Callable, Consumeable) {
			var fhead, ftail = mf.DeCap()
			var ghead, gtail = mg.DeCap()
			if fhead == nil || ghead == nil {
				return nil, BindF(ftail.(MonadicCon), gtail.(MonadicCon), bind)
			}
			if len(args) > 0 {
			}
			return bind(fhead, ghead), BindF(ftail, gtail, bind)
		})
}

// FOLD FUNCTOR LATE BINDING
//
// returns a list of continuations, yielding accumulated result & list of
// follow-up continuations. when the list is depleted, return result only.
func FoldF(list ListVal, fold FoldFnc, elem Callable) ListVal {
	return ListVal(func(args ...Callable) (Callable, ListVal) {
		var head, tail = list()
		// return when list depleted
		if head == nil {
			return nil, list
		}
		// update the accumulated result by passing it to fold
		// followed by head and all elements passed to yield
		// the call
		elem = fold(elem, head, args...)
		// return result & continuation
		return elem, FoldF(tail, fold, elem)
	})
}

// FILTER FUNCTOR LATE BINDING
func FilterF(list ListVal, filter FilterFnc) ListVal {
	return ListVal(
		func(args ...Callable) (Callable, ListVal) {
			var head, tail = list()
			// return when list depleted
			if head == nil {
				return nil, list
			}
			// if result is filtered out‥.
			if !filter(head, args...) {
				// progress by recursively passing on arguments, filter
				// & remaining tail
				return head, FilterF(tail, filter)
			}

			// otherwise return result & continuation on remaining
			// elements, possibly taking new arguments into
			// consideration, when called
			return head, FilterF(tail, filter)
		})
}

func StrideF(arity Arity, list ListVal, stride ApplyFnc) ListVal {

	var ari = int(arity)
	var parms = []Callable{}
	var head, tail = list()

	return ListVal(func(args ...Callable) (Callable, ListVal) {

		if head == nil {
			return nil, StrideF(arity, tail, stride)
		}
		if len(args) > 0 {
			if len(args) > ari {
				parms = append(parms, args[:ari]...)
				args = args[ari:]
				return stride(head.Call, parms...), StrideF(arity, tail.Cons(args...), stride)
			}
			return stride(head.Call, args...), StrideF(arity, tail.Cons(args...), stride)
		}
		return stride(head.Call), StrideF(arity, tail, stride)
	})
}

func ZipF(llist, rlist ListVal, zip ZipFnc) ListVal {
	return ListVal(
		func(args ...Callable) (Callable, ListVal) {
			var lhead, ltail = llist()
			var rhead, rtail = rlist()
			if lhead == nil || rhead == nil {
				return nil, ZipF(llist, rlist, zip)
			}
			if len(args) > 0 {
				return zip(lhead, rhead).Call(args...), ZipF(ltail, rtail, zip)
			}
			return zip(lhead, rhead), ZipF(ltail, rtail, zip)
		})
}

// split takes the whole list and might remove multiple elements to satisfy
// split fnc
func SplitF(list ListVal, split SplitFnc) ListVal {
	return ListVal(
		func(args ...Callable) (Callable, ListVal) {
			var p PairVal
			p, list = split(list)
			if p == nil {
				return nil, SplitF(NewList(args...), split)
			}
			if len(args) > 0 {
				return p.Call(args...), SplitF(list, split)
			}
			return p, SplitF(list, split)
		})
}

//////////////////////////////////////////////////////////////////////////////////////
/// EAGER MAP FUNCTOR
// applys each element of list to passed function and returns resulting vector
func Map(list Consumeable, fmap MapFnc) VecVal {
	var result = NewVector()
	var head, tail = list.DeCap()
	for head != nil {
		result = NewVector(result, fmap(head))
		head, tail = tail.DeCap()
	}
	return result
}

/// EAGER FOLD FUNCTOR
// fold takes a list, an initial element to pass on and a fold function that
// gets called for each list element expecting the initialy passed element to
// be passed on from call to call to accumulate results
func Fold(list Consumeable, fold FoldFnc, ilem Callable) Callable {
	var head, tail = list.DeCap()
	for head != nil {
		ilem = fold(ilem, head)
		head, tail = tail.DeCap()
	}
	return ilem
}

/// EAGER FILTER FUNCTOR
// applys each element to filter function and returns the list of elements that
// yielded true.
func Filter(list Consumeable, filter FilterFnc) VecVal {
	var result = NewVector()
	var head, tail = list.DeCap()
	for head != nil {
		if filter(head) {
			result = NewVector(result, head)
		}
		head, tail = tail.DeCap()
	}
	return result
}
