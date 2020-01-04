package functions

///////////////////////////////////////////////////////////////////////////////
//// COMPOSITION OF DEFINED FUNCTIONS
///
// define the curryed function, so that it accepts the argument types of the g
// function passed as second argument to the constructor and the return type of
// the g function passed as its second argument.
func Curry(f, g FuncDef) FuncDef {
	if f.TypeArgs().Match(g.TypeRet()) {
		return Define(Lambda(

			func(args ...Expression) Expression {

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

///////////////////////////////////////////////////////////////////////////////
//// CONTINUATION COMPOSITION
///
// flatten flattens sequences of sequences to one dimension
func Flatten(grp Continued) Continued {
	if grp.Empty() {
		return NewList()
	}
	if grp.Head().Type().Match(Collections) {
		return Flatten(
			grp.Head().(Continued),
		).Concat(
			Flatten(grp.Tail()))
	}
	return ListVal(func(args ...Expression) (Expression, ListVal) {
		if len(args) > 0 {
			grp = grp.Call(args...).(Continued)
		}
		var head, tail = grp.Continue()
		return head, Flatten(tail).(ListVal)
	})
}

// map returns a continuation calling the map function for every element
func Map(
	con Continued,
	mapf func(Expression) Expression,
) ListVal {
	if con.Empty() {
		return ListVal(func(args ...Expression) (Expression, ListVal) {
			if len(args) > 0 {
				var head, tail = Map(
					NewList(args...), mapf,
				).Continue()
				return head, tail.(ListVal)
			}
			return NewNone(), nil
		})
	}
	return ListVal(func(args ...Expression) (Expression, ListVal) {
		if len(args) > 0 {
			con = con.Call(args...).(Continued)
		}
		var head, tail = con.Continue()
		// skip none instances, when tail has further elements
		if head.Type().Match(None) && !tail.Empty() {
			return Map(tail, mapf)()
		}
		return mapf(head), Map(tail, mapf)
	})
}

// apply returns a continuation called on every element of the continuation.
// when continuation is called pssing arguments those, are passed to apply
// alongside the current element
func Apply(
	con Continued,
	apply func(Expression, ...Expression) Expression,
) ListVal {
	if con.Empty() {
		return ListVal(func(args ...Expression) (Expression, ListVal) {
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
	return ListVal(func(args ...Expression) (Expression, ListVal) {
		if len(args) > 0 {
			var head, tail = apply(head, args...),
				Apply(tail, apply)
				// skip none
			if head.Type().Match(None) && !tail.Empty() {
				return Apply(tail, apply)()
			}

		}
		var head, tail = apply(head), Apply(tail, apply)
		// skip none
		if head.Type().Match(None) && !tail.Empty() {
			return Apply(tail, apply)()
		}
		return head, tail
	})
}

// fold takes a continuation, an initial expression and a fold function.  the
// fold function is called for every element of the continuation and passed the
// current element and init expression and returns a possbly altered init
// element to pass to next call
func Fold(
	con Continued,
	init Expression,
	fold func(init, head Expression) Expression,
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

// continuation of elements not matched by test
func Filter(
	con Continued,
	filter func(Expression) bool,
) Grouped {
	if con.Empty() {
		return ListVal(func(args ...Expression) (Expression, ListVal) {
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
		fold = func(init, head Expression) Expression {
			if filter(head) {
				return NewNone()
			}
			return init.(VecVal).Cons(head)
		}
	)
	return Fold(con, init, fold)
}

// continuation of elements matched by test
func Pass(
	con Continued,
	pass func(Expression) bool,
) Grouped {
	if con.Empty() {
		return ListVal(func(args ...Expression) (Expression, ListVal) {
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
		fold = func(init, head Expression) Expression {
			if pass(head) {
				return init.(VecVal).Cons(head)
			}
			return NewNone()
		}
	)
	return Fold(con, init, fold)
}

// take-n is a variation of fold that takes an initial continuation cuts and
// returns it as continuation of vector instances of length n
func TakeN(grp Continued, n int) Continued {
	if grp.Empty() {
		return grp
	}
	var (
		vec   = NewVector()
		takeN = func(init Expression, arg Expression) Expression {
			var vector = init.(VecVal)
			if vector.Len() == n {
				return NewVector(arg)
			}
			return vector.Cons(arg).(VecVal)
		}
	)
	return Filter(Fold(grp, vec, takeN),
		func(arg Expression) bool {
			return arg.(VecVal).Len() < n
		})
}

// zip expects two continuations and a function to create a list of resulting
// elements each created from the two current continuation heads, using the
// passed zip function.  if arguments are passed calling the lists call method,
// the results call method is called after heads have been zipped, passing on
// those arguemnts.
func Zip(
	left, right Continued,
	zip func(l, r Expression) Expression,
) ListVal {
	if left.Empty() && right.Empty() {
		return ListVal(func(args ...Expression) (Expression, ListVal) {
			return NewNone(), nil
		})
	}
	return ListVal(func(args ...Expression) (Expression, ListVal) {
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

// split is a variation of fold that splits either a continuation of pairs, or
// takes two arguments at a time and splits those into continuation of left and
// right values and returns those as elements of a pair
func Split(
	con Continued,
	split func(Expression) Paired,
) ListVal {
	var pair = NewPair(NewVector(), NewVector())
	return Fold(con, pair, func(init, head Expression) Expression {
		var (
			pair  = split(head)
			left  = init.(Paired).Left().(VecVal).Cons(pair.Left())
			right = init.(Paired).Right().(VecVal).Cons(pair.Right())
		)
		return NewPair(left, right)
	})
}

// bind works similar to zip, but the bind function takes additional arguments
// during runtime and passes them with the heads of both lists passed to the
// call method (instead of passing on to results call method, analog to
// map/apply).  when both functions are curryed and the arguments are passed to
// the resulting function, bind behaves like the '.' operator in haskell.
func Bind(
	f, g Continued,
	bind func(l, r Expression, args ...Expression) Expression,
) ListVal {
	if f.Empty() || g.Empty() {
		return ListVal(func(args ...Expression) (Expression, ListVal) {
			return NewNone(), nil
		})
	}
	return ListVal(func(args ...Expression) (Expression, ListVal) {
		var (
			gh, gt = g.Continue()
			fh, ft = f.Continue()
		)
		if len(args) > 0 {
			// pass arguments on to bind
			return bind(gh, fh, args...),
				Bind(gt, ft, bind)
		}
		return bind(gh, fh), Bind(gt, ft, bind)
	})
}

// lazy quick sort implementation expects a continuation and a less function
// that returns true, if its right element is lesser (not equal!) compared to
// the left one, as its arguments.
//
// sort picks continuations current head as pivot, and defines a split function
// enclosing it, returning elements divided by the less function with pivot as
// first and current element as second argument to compare against.
//
// should the list of lesser elements turn out to be empty, current pivot is
// returned as head of sorted list and the sorted list of greater elements as
// its tail.
//
// otherwise the list resulting from concatenating the sorted list of greater
// and equal elements (will include pivot) to the sorted list of lesser
// elements, is returned.
//
// sort calls itself recursively on progressively shorter subsets of the list.
// computations are evaluated *until* the smallest element of current calls
// subset is found →
//
//  - lazy quicksort type divide & conquer algorithm
//  - only part of the list lesser current element needs to be sorted
//  - infinite lists can be sorted (returns a sorted list of all results until
//    current computation, with every call)
//
func Sort(
	grp Grouped, less func(l, r Expression) bool,
) Grouped {
	if grp.Empty() {
		return Sort(grp, less)
	}
	var (
		pivot, tail = grp.Continue()
		lt          = func(arg Expression) bool { return less(arg, pivot) }
		gteq        = func(arg Expression) bool { return less(pivot, arg) }
		lesser      = Filter(tail, lt)
		greater     = Filter(tail, gteq)
	)
	if lesser.Empty() {
		return Sort(greater, less).Cons(pivot)
	}
	return Sort(lesser, less).Concat(Sort(greater, less).Cons(pivot))
}
