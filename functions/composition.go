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
func FoldE(resource ResourceFnc, fold BinaryFnc, ilem Parametric) Parametric {
	var head, tail = resource()
	for head != nil {
		ilem = fold(ilem, head)
		head, tail = tail()
	}
	return ilem
}

/// EAGER MAP LIST
func MapE(resource ResourceFnc, fmap UnaryFnc) ResourceFnc {
	var result = NewVector()
	var head, tail = resource()
	for head != nil {
		result = conVec(result, fmap(head))
		head, tail = tail()
	}
	return NewResource(result)
}

/// FILTER LIST EAGER
func FilterE(resource ResourceFnc, filter TruthFnc) ResourceFnc {
	var result = NewVector()
	var head, tail = resource()
	for head != nil {
		if filter(head)() {
			result = conVec(result, head)
			head, tail = tail()
		}
	}
	return NewResource(result)
}

//// LATE BINDING LIST COMPOSITION
///
// FOLD LIST LATE BINDING
func FoldL(resource ResourceFnc, fold BinaryFnc, ilem Parametric) ResourceFnc {

	return ResourceFnc(

		func(args ...Parametric) (Parametric, ResourceFnc) {

			var head, tail = resource()

			if head == nil {
				return nil, resource
			}

			return fold(
					ilem,
					head.Call(args...),
				),
				FoldL(
					NewResource(tail),
					fold,
					ilem,
				)
		})
}

// MAP LIST LATE BINDING
func MapL(resource ResourceFnc, fmap UnaryFnc) ResourceFnc {

	return ResourceFnc(

		func(args ...Parametric) (Parametric, ResourceFnc) {

			var head, tail = resource()

			if head == nil {
				return nil, resource
			}

			return fmap(
					head.Call(args...),
				),
				MapL(
					NewResource(tail),
					fmap,
				)
		})
}

// FILTER LIST LATE BINDING
func FilterL(resource ResourceFnc, filter TruthFnc) ResourceFnc {

	return ResourceFnc(

		func(args ...Parametric) (Parametric, ResourceFnc) {

			var head, tail = resource()

			if head == nil {
				return nil, resource
			}

			var result = head.Call(args...)

			if filter(result)() {
				return result,
					FilterL(
						NewResource(tail),
						filter,
					)
			}

			return FilterL(
				NewResource(tail),
				filter,
			)(
				args...,
			)
		})
}
