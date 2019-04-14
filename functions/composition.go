package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

// type system implementation
type (
	/// FUNCTOR
	FunctFnc func(args ...Callable) (Callable, FunctFnc)

	/// APPLY & APPLICAPLE
	AppliFnc func(...Callable) (Callable, AppliFnc)

	/// MONOID
	MonaFnc func(args ...Callable) (Callable, MonaFnc)

	/// RESSOURCE
	GenericFnc func(args ...Callable) (Callable, GenericFnc)
)

func (c GenericFnc) String() string                 { return "ϝ → Fₙ [F]" }
func (c GenericFnc) Ident() Callable                { return c }
func (c GenericFnc) DeCap() (Callable, Consumeable) { return c() }
func (c GenericFnc) Head() Callable                 { h, _ := c(); return h }
func (c GenericFnc) Tail() Consumeable              { _, t := c(); return t }
func (c GenericFnc) Eval(args ...d.Native) d.Native { return c.Head().Eval(args...) }
func (c GenericFnc) TypeFnc() TyFnc                 { return Functor | c.Head().TypeFnc() }
func (c GenericFnc) TypeNat() d.TyNative {
	res, _ := c()
	return res.TypeNat() | d.Function
}
func (c GenericFnc) Call(args ...Callable) Callable { return c.Head().Call(args...) }

//// CURRY
func Curry(fnc Callable, arg Callable) Callable {
	return NaryFnc(func(args ...Callable) Callable {
		return fnc.Call(arg).Call(args...)
	})
}

// FUNCTOR
func NewFunctor(list Consumeable) FunctFnc {
	return func(args ...Callable) (Callable, FunctFnc) {
		var head, tail = list.DeCap()
		if head == nil {
			return nil, NewFunctor(tail)
		}
		return head.Call(args...), NewFunctor(tail)
	}
}

func (c FunctFnc) Call(args ...Callable) Callable {
	return FunctFnc(
		func(...Callable) (Callable, FunctFnc) {
			return c(args...)
		})
}

func (g FunctFnc) String() string                 { return "ϝ → Fₙ [F]" }
func (c FunctFnc) Ident() Callable                { return c }
func (c FunctFnc) DeCap() (Callable, Consumeable) { return c() }
func (c FunctFnc) Head() Callable                 { h, _ := c(); return h }
func (c FunctFnc) Tail() Consumeable              { _, t := c(); return t }
func (c FunctFnc) Eval(args ...d.Native) d.Native { return c.Head().Eval(args...) }
func (c FunctFnc) TypeFnc() TyFnc                 { return Functor | c.Head().TypeFnc() }
func (c FunctFnc) TypeNat() d.TyNative {
	res, _ := c()
	return res.TypeNat() | d.Function
}

// APPLY FUNCTION
// appliccable lists 'know' how to 'treat' the contained values, given the
// parameters that have been passed. the apply function, promotes applys the
// args to the current head element of the list and progresses to yield a new
// applicable, reducing the tail.
func NewAppliFnc(
	applf NaryFnc,
	list ListFnc,
) AppliFnc {

	var apply = applf

	return func(args ...Callable) (Callable, AppliFnc) {

		var head, tail = list()
		var call Callable

		if head != nil {

			if len(args) > 0 {

				call = apply(

					append(
						[]Callable{head},
						args...,
					)...,
				)
			} else {

				call = apply(head)
			}
		}

		return call, NewAppliFnc(apply, tail)
	}
}

func (c AppliFnc) String() string                 { return "Apply " + c.Head().TypeFnc().String() }
func (c AppliFnc) Ident() Callable                { return c }
func (c AppliFnc) DeCap() (Callable, Consumeable) { return c() }
func (c AppliFnc) Head() Callable                 { h, _ := c(); return h }
func (c AppliFnc) Tail() Consumeable              { _, t := c(); return t }
func (c AppliFnc) Call(args ...Callable) Callable { return c.Head().Call(args...) }
func (c AppliFnc) Eval(args ...d.Native) d.Native { return c.Head().Eval(args...) }
func (c AppliFnc) TypeFnc() TyFnc                 { return Applicaple | c.Head().TypeFnc() }
func (c AppliFnc) TypeNat() d.TyNative {
	res, _ := c()
	return res.TypeNat() | d.Function
}

// MONADIC
func (c MonaFnc) Call(args ...Callable) Callable {
	var result Callable
	return result
}
func (g MonaFnc) String() string                 { return "ϝ → Mₙₘ [M]" }
func (c MonaFnc) Ident() Callable                { return c }
func (c MonaFnc) DeCap() (Callable, Consumeable) { return c() }
func (c MonaFnc) Head() Callable                 { h, _ := c(); return h }
func (c MonaFnc) Tail() Consumeable              { _, t := c(); return t }
func (c MonaFnc) Eval(args ...d.Native) d.Native { return c.Head().Eval(args...) }
func (c MonaFnc) TypeFnc() TyFnc                 { return Monad | c.Head().TypeFnc() }
func (c MonaFnc) TypeNat() d.TyNative            { res, _ := c(); return res.TypeNat() | d.Function }

//////////////////////////////////////////////////////////////////////////////////////
// LEFT FOLD FUNCTOR
func Fold(list GenericFnc, fold BinaryFnc, ilem Callable) Callable {
	var head, tail = list()
	for head != nil {
		ilem = fold(ilem, head)
		head, tail = tail()
	}
	return ilem
}

/// EAGER MAP FUNCTOR
func Map(list GenericFnc, fmap UnaryFnc) VecFnc {
	var result = NewVector()
	var head, tail = list()
	for head != nil {
		result = ConVector(result, fmap(head))
		head, tail = tail()
	}
	return result
}

/// FILTER FUNCTOR EAGER
func Filter(list GenericFnc, filter TruthFnc) VecFnc {
	var result = NewVector()
	var head, tail = list()
	for head != nil {
		if filter(head) {
			result = ConVector(result, head)
		}
		head, tail = tail()
	}
	return result
}

//// LATE BINDING functor COMPOSITION
///
// FOLD FUNCTOR LATE BINDING
//
// returns a list of continuations, yielding accumulated result & list of
// follow-up continuations. when the list is depleted, return result only.
func FoldF(list GenericFnc, fold BinaryFnc, ilem Callable) GenericFnc {

	return GenericFnc(

		func(args ...Callable) (Callable, GenericFnc) {

			var head, tail = list()

			if head == nil {
				return list, nil
			}

			// update the accumulated result
			ilem = fold(
				ilem,
				head.Call(args...),
			)

			// return result & continuation
			return ilem,
				FoldF(
					tail,
					fold,
					ilem,
				)
		})
}

// MAP FUNCTOR LATE BINDING
func MapF(list GenericFnc, fmap UnaryFnc) GenericFnc {

	return GenericFnc(

		func(args ...Callable) (Callable, GenericFnc) {

			var head, tail = list()

			if head == nil {
				return nil, list
			}

			return fmap(
					head.Call(args...),
				),
				MapF(
					tail,
					fmap,
				)
		})
}

// FILTER FUNCTOR LATE BINDING
func FilterF(list GenericFnc, filter TruthFnc) GenericFnc {

	return GenericFnc(

		func(args ...Callable) (Callable, GenericFnc) {

			var head, tail = list()

			if head == nil {
				return nil, list
			}

			// applying args by calling the head element, yields
			// result to filter
			var result = head.Call(args...)

			// if result is filtered out‥.
			if !filter(result) {
				// progress by passing args, filter & tail on
				// recursively
				return head, FilterF(tail, filter)
			}

			// otherwise return result & continuation on remaining
			// elements, possibly taking new arguments into
			// consideration, when called
			return result,
				FilterF(
					tail,
					filter,
				)
		})
}
