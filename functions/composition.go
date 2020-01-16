package functions

import "fmt"

///////////////////////////////////////////////////////////////////////////////
//// COMPOSITION OF DEFINED FUNCTIONS
///
// define the curryed function, so that it accepts the argument types of the g
// function passed as second argument to the constructor and the return type of
// the g function passed as its second argument.
func Curry(f, g FuncVal) FuncVal {
	if f.TypeArgs().Match(g.TypeRet()) {
		return Define(Lambda(

			func(args ...Functor) Functor {

				// call f with the result of calling g applying
				// the arguments if any are given
				if len(args) > 0 {
					return f.Call(g.Call(args...))
				}
				return f.Call(g.Call())
			}),

			// define a function by composing both type ids with
			// the argument type of g passed as second argument and
			// the return type of f passed as first argument
			Def(g.TypeId(), Def(f.TypeId())),
			f.TypeRet(),  // ‥.return type of g &
			g.TypeArgs(), //‥.argument type of f
		)
	}
	return Define(NewNone(), None, None)
}

/// FLATTEN
// flattens sequences of sequences to one dimension
func Flatten(grp Sequential) ListVal {
	if grp.Empty() {
		return NewList()
	}
	if grp.Head().Type().Match(Additives) {
		return Flatten(
			grp.Head().(Sequential),
		).Concat(Flatten(grp.Tail())).(ListVal)
	}
	return ListVal(func(args ...Functor) (Functor, ListVal) {
		if len(args) > 0 {
			grp = grp.Call(args...).(Sequential)
		}
		var head, tail = grp.Continue()
		return head, Flatten(tail)
	})
}

/// MAP
// map returns a continuation calling the map function for every element
func Map(
	con Sequential,
	mapf func(Functor) Functor,
) ListVal {
	if con.Empty() {
		return ListVal(func(args ...Functor) (Functor, ListVal) {
			if len(args) > 0 {
				var head, tail = Map(
					NewList(args...), mapf,
				).Continue()
				return head, tail.(ListVal)
			}
			return NewNone(), nil
		})
	}
	return ListVal(func(args ...Functor) (Functor, ListVal) {
		if len(args) > 0 {
			con = con.Call(args...).(Sequential)
		}
		var head, tail = con.Continue()
		// skip none instances, when tail has further elements
		if IsNone(head) && !tail.Empty() {
			return Map(tail, mapf)()
		}
		return mapf(head), Map(tail, mapf)
	})
}

/// APPLY
// apply returns a continuation called on every element of the continuation.
// when continuation is called pssing arguments those, are passed to apply
// alongside the current element
func Apply(
	con Sequential,
	apply func(Functor, ...Functor) Functor,
) ListVal {
	if con.Empty() {
		return ListVal(func(args ...Functor) (Functor, ListVal) {
			if len(args) > 0 {
				var head, tail = Apply(
					NewList(args...), apply,
				).Continue()
				return head, tail.(ListVal)
			}
			return NewNone(), nil
		})
	}

	var head, tail = con.Continue()

	return ListVal(func(args ...Functor) (Functor, ListVal) {
		if len(args) > 0 {
			head, tail = apply(head, args...),
				Apply(tail, apply)
				// skip none
			if IsNone(head) && !tail.Empty() {
				return Apply(tail, apply)()
			}

		}
		var head, tail = apply(head), Apply(tail, apply)
		// skip none
		if IsNone(head) && !tail.Empty() {
			return Apply(tail, apply)()
		}
		return head, tail
	})
}

/// FOLD
// fold takes a continuation, an initial expression and a fold function.  the
// fold function is called for every element of the continuation and passed the
// current element and init expression and returns a possbly altered init
// element to pass to next call
func Fold(
	con Sequential,
	init Functor,
	fold func(init, head Functor) Functor,
) ListVal {

	if con.Empty() { // return accumulated result, when empty
		return NewList()
	}
	var (
		head, tail = con.Continue()   // pop current head & list
		result     = fold(init, head) // calculate temporary result
	)
	if IsNone(result) { // if computation yields none‥.
		// ‥.as long as list is not empty‥.
		for IsNone(result) && !tail.Empty() { // ‥.and results are none‥.
			head, tail = tail.Continue() // ‥.pop heads‥.
			result = fold(init, head)    // ‥.and calculate results‥.
		}
		if IsNone(result) { // result still none →  tail depleted‥.
			// cons accumulated result to empty tail
			return NewList(init)
		}
	}
	// result is not empty, list has further elements‥.
	return NewList(result).Concat(Fold(tail, result, fold)).(ListVal)
}

/// FILTER
// continuation of elements not matched by test
func Filter(
	con Sequential,
	filter func(Functor) bool,
) Applicative {
	if con.Empty() {
		return ListVal(func(args ...Functor) (Functor, ListVal) {
			if len(args) > 0 {
				var head, tail = Pass(
					NewList(args...), filter,
				).Continue()
				return head, tail.(ListVal)
			}
			return NewNone(), nil
		})
	}
	var (
		init = NewVector()
		fold = func(init, head Functor) Functor {
			if filter(head) {
				return NewNone()
			}
			return init.(VecVal).Cons(head)
		}
	)
	return Fold(con, init, fold)
}

/// PASS
// continuation of elements matched by test
func Pass(
	con Sequential,
	pass func(Functor) bool,
) Applicative {
	if con.Empty() {
		return ListVal(func(args ...Functor) (Functor, ListVal) {
			if len(args) > 0 {
				var head, tail = Pass(
					NewList(args...), pass,
				).Continue()
				return head, tail.(ListVal)
			}
			return NewNone(), nil
		})
	}
	var (
		init = NewVector()
		fold = func(init, head Functor) Functor {
			if pass(head) {
				return init.(VecVal).Cons(head)
			}
			return NewNone()
		}
	)
	return Fold(con, init, fold)
}

/// TAKE-N
// take-n is a variation of fold that takes an initial continuation cuts and
// returns it as continuation of vector instances of length n
func TakeN(grp Sequential, n int) Sequential {
	if grp.Empty() {
		return grp
	}
	var (
		vec   = NewVector()
		takeN = func(init Functor, arg Functor) Functor {
			var vector = init.(VecVal)
			if vector.Len() == n {
				return NewVector(arg)
			}
			return vector.Cons(arg).(VecVal)
		}
	)
	return Pass(Fold(grp, vec, takeN),
		func(arg Functor) bool {
			return arg.(VecVal).Len() == n
		})
}

/// ZIP
// zip expects two continuations and a function to create a list of resulting
// elements each created from the two current continuation heads, using the
// passed zip function.  if arguments are passed calling the lists call method,
// the results call method is called after heads have been zipped, passing on
// those arguemnts.
func Zip(
	left, right Sequential,
	zip func(l, r Functor) Functor,
) ListVal {
	if left.Empty() && right.Empty() {
		return ListVal(func(args ...Functor) (Functor, ListVal) {
			return NewNone(), nil
		})
	}
	return ListVal(func(args ...Functor) (Functor, ListVal) {
		var (
			gh, gt = left.Continue()
			fh, ft = right.Continue()
		)
		if len(args) > 0 {
			// pass arguments to results call method
			return zip(gh, fh).Call(args...),
				Zip(gt, ft, zip)
		}
		return zip(gh, fh), Zip(gt, ft, zip)
	})
}

/// SPLIT
// split is a variation of fold that splits either a continuation of pairs, or
// takes two arguments at a time and splits those into continuation of left and
// right values and returns those as elements of a pair
func Split(
	con Sequential,
	split func(Functor) Paired,
) ListVal {
	var pair = NewPair(NewVector(), NewVector())
	return Fold(con, pair, func(init, head Functor) Functor {
		var (
			pair  = split(head)
			left  = init.(Paired).Left().(VecVal).Cons(pair.Left())
			right = init.(Paired).Right().(VecVal).Cons(pair.Right())
		)
		return NewPair(left, right)
	})
}

/// BIND
func Bind(
	f, g Applicative, bind func(f, g Applicative, args ...Functor) (
		Functor, Applicative, Applicative),
) ListVal {
	return ListVal(func(args ...Functor) (Functor, ListVal) {

		var result Functor
		if len(args) > 0 {
			result, f, g = bind(f, g, args...)
		} else {
			result, f, g = bind(f, g)
		}
		// ‥.as long as continuations are not depleted‥.
		for IsNone(result) { // ‥.and no result is yielded‥.
			result, f, g = bind(f, g) // ‥.re-calculate the result‥.
			if f.Empty() && g.Empty() {
				return result, NewList() // RETURN FINAL RESULT
			}
		}
		fmt.Printf("result: %s\tnone?: %t\tf-empty?: %t\tg-empty?: %t\n\n",
			result, IsNone(result), f.Empty(), g.Empty())

		// return result yielded by bind operation
		return result, Bind(f, g, bind)
	})
}
