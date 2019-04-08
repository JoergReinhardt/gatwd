package functions

//// CURRY
///
//
func Curry(fnc Parametric, arg Parametric) Parametric {
	return NaryFnc(func(args ...Parametric) Parametric {
		return fnc.Call(arg).Call(args...)
	})
}

///////////////////////////////////////////////////////////////////////////
//// EAGER LIST COMPOSITION
///
// LEFT FOLD LIST
func Fold(resource FunctorFnc, fold BinaryFnc, ilem Parametric) Parametric {
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
		result = conVec(result, fmap(head))
		head, tail = tail()
	}
	return NewFunctor(result)
}

/// FILTER LIST EAGER
func Filter(resource FunctorFnc, filter TruthFnc) FunctorFnc {
	var result = NewVector()
	var head, tail = resource()
	for head != nil {
		if filter(head)() {
			result = conVec(result, head)
			head, tail = tail()
		}
	}
	return NewFunctor(result)
}

//// LATE BINDING LIST COMPOSITION
///
// FOLD LIST LATE BINDING
func FoldF(resource FunctorFnc, fold BinaryFnc, ilem Parametric) FunctorFnc {

	return FunctorFnc(

		func(args ...Parametric) (Parametric, FunctorFnc) {

			var head, tail = resource()

			if head == nil {
				return nil, resource
			}

			return fold(
					ilem,
					head.Call(args...),
				),
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

		func(args ...Parametric) (Parametric, FunctorFnc) {

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

		func(args ...Parametric) (Parametric, FunctorFnc) {

			var head, tail = resource()

			if head == nil {
				return nil, resource
			}

			// applying args by calling the head element, yields
			// result to filter
			var result = head.Call(args...)

			// if result is filtered out‥.
			if !filter(result)() {
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
