package functions

//// CURRY
///
//
func Curry(fnc Callable, arg Callable) Callable {
	return NaryFnc(func(args ...Callable) Callable {
		return fnc.Call(arg).Call(args...)
	})
}

///////////////////////////////////////////////////////////////////////////
//// EAGER LIST COMPOSITION
///
// LEFT FOLD LIST
func Fold(resource FunctorFnc, fold BinaryFnc, ilem Callable) Callable {
	var head, tail = resource()
	for head != nil {
		ilem = fold(ilem, head)
		head, tail = tail()
	}
	return ilem
}

/// EAGER MAP LIST
func Map(resource FunctorFnc, fmap UnaryFnc) FunctorFnc {
	var result = NewVector()
	var head, tail = resource()
	for head != nil {
		result = ConVector(result, fmap(head))
		head, tail = tail()
	}
	return NewFunctor(result)
}

/// FILTER LIST EAGER
func Filter(resource FunctorFnc, filter TruthFnc) FunctorFnc {
	var result = NewVector()
	var head, tail = resource()
	for head != nil {
		if filter(head) {
			result = ConVector(result, head)
		}
		head, tail = tail()
	}
	return NewFunctor(result)
}

//// LATE BINDING LIST COMPOSITION
///
// FOLD LIST LATE BINDING
//
// returns a list of continuations, yielding accumulated result & list of
// follow-up continuations. when the list is depleted, return result only.
func FoldF(resource FunctorFnc, fold BinaryFnc, ilem Callable) FunctorFnc {

	return FunctorFnc(

		func(args ...Callable) (Callable, FunctorFnc) {

			var head, tail = resource()

			if head == nil {
				return resource, nil
			}

			// update the accumulated result
			ilem = fold(
				ilem,
				head.Call(args...),
			)

			// return result & continuation
			return ilem,
				FoldF(
					NewFunctor(tail),
					fold,
					ilem,
				)
		})
}

// MAP LIST LATE BINDING
func MapF(resource FunctorFnc, fmap UnaryFnc) FunctorFnc {

	return FunctorFnc(

		func(args ...Callable) (Callable, FunctorFnc) {

			var head, tail = resource()

			if head == nil {
				return nil, resource
			}

			return fmap(
					head.Call(args...),
				),
				MapF(
					NewFunctor(tail),
					fmap,
				)
		})
}

// FILTER LIST LATE BINDING
func FilterF(resource FunctorFnc, filter TruthFnc) FunctorFnc {

	return FunctorFnc(

		func(args ...Callable) (Callable, FunctorFnc) {

			var head, tail = resource()

			if head == nil {
				return nil, resource
			}

			// applying args by calling the head element, yields
			// result to filter
			var result = head.Call(args...)

			// if result is filtered out‥.
			if !filter(result) {
				// progress by passing args, filter & tail on
				// recursively
				return FilterF(
					NewFunctor(tail),
					filter,
				)(
					args...,
				)
			}

			// otherwise return result & continuation on remaining
			// elements, possibly taking new arguments into
			// consideration, when called
			return result,
				FilterF(
					NewFunctor(tail),
					filter,
				)
		})
}
