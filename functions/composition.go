package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

// type system implementation
type (
	//// CONSUMER
	///
	// consumer function consumes consumeables typeagnosticly to provide a
	// common return type function of all endofunctor operators. that way
	// endofunctors don't need to be type converted
	ConsumerFnc func(...Callable) (Callable, Consumeable)

	//// MAP FUNCTION
	///
	MapFnc func(...Callable) Callable

	//// FOLD FUNCTION
	///
	FoldFnc func(Callable, Callable, ...Callable) Callable

	//// FILTER FUNCTION
	///
	FilterFnc func(Callable, ...Callable) bool

	//// JOIN FUNCTION
	///
	JoinFnc func(f, g NaryExpr, args ...Callable) Callable

	//// APPLY FUNCTION
	///
	ApplyFnc func(NaryExpr, ...Callable) Callable

	//// FUNCTOR
	///
	// all functors apply to map-, foldl
	FunctorCon func(...Callable) (Callable, Consumeable)

	//// APPLICAPLE
	///
	// applicables are functors to be applyd on a list of boxed values
	ApplicativeCon func(...Callable) (Callable, Consumeable)

	//// MONADIC
	///
	// monadic functions provide transformations between two functor types
	MonadicCon func(...Callable) (Callable, Consumeable)
)

//// CURRY
func Curry(exprs ...Callable) Callable {

	if len(exprs) == 0 {

		return NewNaryExpr(Curry)
	}

	if len(exprs) == 1 {

		return NewNaryExpr(exprs[0].Call)
	}

	return NewNaryExpr(

		func(args ...Callable) Callable {

			return exprs[0].Call(Curry(exprs[1:]...).Call(args...))
		})
}

//// CONSUME
///
// consumes an epifunctor and passes its arguments through, when called
func Consume(fnc func(...Callable) (Callable, Consumeable)) ConsumerFnc {

	return func(args ...Callable) (Callable, Consumeable) {

		return fnc(args...)
	}
}

func (c ConsumerFnc) Ident() Callable                { return c }
func (c ConsumerFnc) DeCap() (Callable, Consumeable) { return c() }
func (c ConsumerFnc) String() string                 { return "ϝ → Fₙ [F]" }
func (c ConsumerFnc) Head() Callable                 { h, _ := c(); return h }
func (c ConsumerFnc) Tail() Consumeable              { _, t := c(); return t }
func (c ConsumerFnc) Eval(args ...d.Native) d.Native { return c.Head().Eval(args...) }
func (c ConsumerFnc) TypeFnc() TyFnc                 { return Functor | c.Head().TypeFnc() }
func (c ConsumerFnc) TypeNat() d.TyNative {
	res, _ := c()
	return res.TypeNat()
}

func (c ConsumerFnc) Call(args ...Callable) Callable { result, _ := c(args...); return result }

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
func (c FunctorCon) TypeNat() d.TyNative            { return c.Head().TypeNat() }
func (c FunctorCon) String() string                 { return c.Head().String() }

/// APPLICATIVE
// applicative encloses over a function to be applyd the head element and any
// arguments given at each call
func NewApplicative(list Consumeable, applyFnc ApplyFnc) ApplicativeCon {

	var apply = applyFnc

	return func(args ...Callable) (Callable, Consumeable) {

		var head, tail = list.DeCap()

		if head == nil {
			return nil, NewApplicative(NewList(args...), apply)
		}

		if len(args) > 0 {

			return apply(head.Call, args...),
				NewApplicative(tail, apply)
		}

		return apply(head.Call), NewApplicative(NewList(), apply)
	}
}

func (c ApplicativeCon) Ident() Callable                { return c }
func (c ApplicativeCon) DeCap() (Callable, Consumeable) { return c() }
func (c ApplicativeCon) Head() Callable                 { h, _ := c(); return h }
func (c ApplicativeCon) Tail() Consumeable              { _, t := c(); return t }
func (c ApplicativeCon) Call(args ...Callable) Callable { return c.Head().Call(args...) }
func (c ApplicativeCon) Eval(args ...d.Native) d.Native { return c.Head().Eval(args...) }
func (c ApplicativeCon) TypeFnc() TyFnc                 { return Applicable | c.Head().TypeFnc() }
func (c ApplicativeCon) TypeNat() d.TyNative            { return c.Head().TypeNat() }
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
func (c MonadicCon) TypeNat() d.TyNative            { return c.Head().TypeNat() }
func (c MonadicCon) TypeFnc() TyFnc                 { return Monad | c.Head().TypeFnc() }

//// MAP FUNCTOR LATE BINDING
///
// expects a consumeable list and a mapping function to apply on each element.
// list elements are late bound per call to the resulting consumeable, passed
// arguments get concatenated to the yielded list element, when fmap is called.
func MapF(list Consumeable, fmap MapFnc) Consumeable {
	return ConsumerFnc(
		func(args ...Callable) (Callable, Consumeable) {
			// decapitate list to get head and list continuation
			var head, tail = list.DeCap()
			if head == nil { // return empty head
				return nil, list
			}
			// return result of applying arguments to fmap and the
			// list continuation
			return fmap(append(
					[]Callable{head},
					args...,
				)...),
				MapF(tail, fmap)
		})
}

// FOLD FUNCTOR LATE BINDING
//
// returns a list of continuations, yielding accumulated result & list of
// follow-up continuations. when the list is depleted, return result only.
func FoldF(list Consumeable, fold FoldFnc, ilem Callable) Consumeable {
	return ConsumerFnc(
		func(args ...Callable) (Callable, Consumeable) {
			var head, tail = list.DeCap()
			// return when list depleted
			if head == nil {
				return nil, list
			}
			// update the accumulated result by passing it to fold
			// followed by head and all elements passed to yield
			// the call
			ilem = fold(ilem, head, args...)
			// return result & continuation
			return ilem, FoldF(tail, fold, ilem)
		})
}

// FILTER FUNCTOR LATE BINDING
func FilterF(list ListVal, filter FilterFnc) Consumeable {
	return ConsumerFnc(
		func(args ...Callable) (Callable, Consumeable) {
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
