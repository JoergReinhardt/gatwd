package functions

///////////////////////////////////////////////////////////////////////////////
//// COMPOSITION OF DEFINED FUNCTIONS
///
// define the curryed function, so that it accepts the argument types of the g
// function passed as second argument to the constructor and the return type of
// the g function passed as its second argument.
func Curry(f, g Def) Def {
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
			Declare(g.TypeId(), Declare(f.TypeId())),
			f.TypeRet(),  // ‥.return type of g &
			g.TypeArgs(), //‥.argument type of f
		)
	}
	return Define(NewNone(), None, None)
}

/// MAP
// map returns a continuation calling the map function for every element
func Map(
	seq Sequential,
	mapf func(Functor) Functor,
) ListVal {

	// return when list is depleted
	if seq.Empty() {
		// preserve map function in returned empty list
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

	// return next continuation
	return ListVal(func(args ...Functor) (Functor, ListVal) {
		if len(args) > 0 {
			seq = seq.Call(args...).(Sequential)
		}
		var head, tail = seq.Continue()
		// skip none instances, when tail has further elements
		if IsNone(head) && !tail.Empty() {
			return Map(tail, mapf)()
		}
		return mapf(head), Map(tail, mapf)
	})
}

/// FLATTEN
// flattens sequences of sequences to one dimension recursively (cps)
func Flatten(seq Sequential) ListVal {
	if seq.Empty() {
		return NewList()
	}
	if seq.Head().Type().Match(Additives) {
		return Flatten(
			seq.Head().(Sequential),
		).Concat(Flatten(seq.Tail())).(ListVal)
	}
	return ListVal(func(args ...Functor) (Functor, ListVal) {
		if len(args) > 0 {
			seq = seq.Call(args...).(Sequential)
		}
		var head, tail = seq.Continue()
		return head, Flatten(tail)
	})
}

/// APPLY
// apply returns a continuation called on every element of the continuation.
// when continuation is called pssing arguments those, are passed to apply
// alongside the current element
func Apply(
	seq Sequential,
	apply func(Functor, ...Functor) Functor,
) Applicative {
	return Map(seq, func(head Functor) Functor {
		return Lambda(func(args ...Functor) Functor {
			return apply(head, args...)
		})
	})
}

//// FOLD
///
// fold takes a continuation, an initial expression and a fold function.  the
// fold function is called for every element of the continuation and passed the
// current element and init expression as arguments.  it returns an instance of
// the init elements type (usually some sort of aggregation over all elements
// seen so far), to pass on as argument for the next call to fold.  fold
// reduces the returned list, by only returning values that aren't instances of
// none. none instances will be skipped in a loop.  that way sequences can be
// folded over lists, that take a variadic number of arguments before they
// return any value,
func Fold(
	seq Sequential,
	init Functor,
	fold func(init, head Functor) Functor,
) ListVal {

	// return initial element wrapped in a list
	if seq.Empty() {
		return ListVal(func(args ...Functor) (Functor, ListVal) {
			if len(args) > 0 {
				return Fold(NewList(args...), init, fold)()
			}
			return NewNone(), nil
		})
	}
	var ( // yield result of current step
		head, tail = seq.Continue()   // pop current head & list
		result     = fold(init, head) // calculate temporary result
	)

	// filter out none instances from returned list
	if IsNone(result) { // if computation yields none‥.
		// ‥.as long as list is not empty‥.
		for IsNone(result) && !tail.Empty() { // ‥.and results are none‥.
			head, tail = tail.Continue() // ‥.pop heads‥.
			result = fold(init, head)    // ‥.and calculate results‥.
		}
	}
	// result is not empty, list has further elements‥.
	// return NewList(result).Concat(Fold(tail, result, fold)).(ListVal)
	return ListVal(func(args ...Functor) (Functor, ListVal) {
		if len(args) > 0 {
			return tail.Cons(result).(ListVal)(args...)
		}
		return result, Fold(tail, result, fold)
	})
}

/// FILTER
// continuation of elements not matched by test
func Filter(
	con Sequential,
	filter func(Functor) bool,
) ListVal {
	if con.Empty() {
		return ListVal(func(args ...Functor) (Functor, ListVal) {
			if len(args) > 0 {
				return Filter(NewList(args...), filter)()
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

// predefined filter to strip all instances of none from a sequence
func StripNone(seq Sequential) ListVal {
	return Filter(seq, func(arg Functor) bool {
		return IsNone(arg)
	})
}

// predefined filter to strip all instances of partials from a sequence
func StripPartial(seq Sequential) ListVal {
	return Filter(seq, func(arg Functor) bool {
		return IsPartial(arg)
	})
}

// reduce a list composition to its normalform (final resulting state, after
// all elements & operations have been applyed)
func Reduce(seq Sequential) ListVal { return StripNone(StripPartial(seq)) }

/// PASS
// continuation of elements matched by test
func Pass(
	con Sequential,
	pass func(Functor) bool,
) ListVal {
	if con.Empty() {
		return ListVal(func(args ...Functor) (Functor, ListVal) {
			if len(args) > 0 {
				return Pass(NewList(args...), pass)()
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
func TakeN(seq Sequential, n int) ListVal {
	if seq.Empty() {
		return seq.(ListVal)
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
	return Pass(Fold(seq, vec, takeN),
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

//// BIND
///
// applys a sequence of arguments to f, and returns the sequence of results
// from applying its results to g.  to deal with expressions that take a fixed,
// or variadic number of arguments equal, or greater than zero, partial return
// values will be stripped from temporary, as well as final results.  the
// stripping is implemented as filter, which internaly folds, so that partial
// result will be progressed by a for loop through all invocations neccesary to
// yield either none, or some instance of the result type.
//
// EXAMPLE: the function 'ascending' takes a sequence of ascending elements
// from a randomized initial list.  the number of arguments it needs to take in
// order to return an ascending sequence and start the next one, varies from
// list to list.  'ascending' needs to be called once per element in the
// initial list and it needs to get the initial element returned from and
// modified at the last call passed, together with the current element. in
// order to extract the complete sequence of ascending subsequences from the
// initial list.  if 'ascending' is the 'f' function in a bind composal, 'g' is
// only supposed to be called once per return value of 'f', and not for every
// temporal result returned by an partialy applyed 'f'.
//
// this works both ways of course, when 'f' returns multiple results per
// element of the initial list, 'g' will be called and passed every single
// return value once at a time.  that way bind can fan out, permutate over
// products of sets, map/reduce‥.
func Bind(
	seq Sequential,
	f func(...Functor) Functor,
	g func(...Functor) Functor,
) ListVal {
	var fnc = Fold(seq, NewVector(), func(head, arg Functor) Functor {
		return f(head, arg)
	})
	return Fold(fnc, NewVector(), func(head, arg Functor) Functor {
		return g(head, arg)
	})
}

//func Sort(
//	seq Sequential,
//	comp Compare,
//) ListVal {
//	var (
//		init      = NewVector()
//		ascending = Define(func(init, arg Functor) Functor {
//		}, DecSym("ascending"), Declare(Vector, T), T)
//	)
//}
