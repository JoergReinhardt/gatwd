package functions

type (

	//// COLLECTIONS
	SeqVal func(...Expression) (Expression, SeqVal)
)

///////////////////////////////////////////////////////////////////////////////
//// SEQUENCE TYPE
///
// generic sequential type
func NewSeqFromGroup(grp Group) SeqVal {
	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
		if len(args) > 0 {
			grp = grp.Cons(args...)
		}
		var head, tail = grp.Continue()
		return head, NewSeqFromGroup(tail)
	})
}
func NewSequence(elems ...Expression) SeqVal {

	// return empty list able to be extended by cons, when no initial
	// elements are given/left
	if len(elems) == 0 {
		return func(args ...Expression) (Expression, SeqVal) {
			if len(args) > 0 {
				if len(args) > 1 {
					return args[0], NewSequence(args[1:]...)
				}
				return args[0], NewSequence()
			}
			// return instance of none as head and a nil pointer as
			// tail, if neither elements nor arguments where passed
			return NewNone(), nil
		}
	}

	// at least one of the initial elements is left‥.
	return func(args ...Expression) (Expression, SeqVal) {

		// if arguments are passed, prepend those and return first
		// argument as head‥.
		if len(args) > 0 {
			// ‥.put arguments up front of preceeding elements
			if len(args) > 1 {
				return args[0], NewSequence(
					append(
						args,
						elems...,
					)...)
			}
			// use single argument as new head of sequence and
			// preceeding elements as tail
			return args[0],
				NewSequence(elems...)
		}

		// no arguments given, but more than one element left → return
		// first element as head, and remaining elements as tail of
		// sequence
		if len(elems) > 1 {
			return elems[0],
				NewSequence(elems[1:]...)
		}
		// return last element and empty sequence
		return elems[0], NewSequence()

	}
}
func (s SeqVal) Head() Expression {
	var cur, _ = s()
	return cur
}
func (s SeqVal) Tail() Group {
	var _, tail = s()
	return tail
}
func (s SeqVal) Continue() (Expression, Group) {
	return s.Head(), s.Tail()
}
func (s SeqVal) Cons(suffix ...Expression) Group {
	if len(suffix) == 0 {
		return s
	}
	if len(suffix) == 1 {
		return SeqVal(func(late ...Expression) (Expression, SeqVal) {
			if len(late) > 0 {
				if len(late) > 1 {
					return late[0],
						s.Cons(append(late[1:], suffix[0])...).(SeqVal)
				}
				return late[0], s.Cons(suffix[0]).(SeqVal)
			}
			return suffix[0], s
		})
	}
	return SeqVal(func(late ...Expression) (Expression, SeqVal) {
		if len(late) > 0 {
			if len(late) > 1 {
				return late[0],
					s.Cons(append(late[1:], suffix...)...).(SeqVal)
			}
			return late[0], s.Cons(suffix...).(SeqVal)
		}
		return suffix[0], s.Cons(suffix[1:]...).(SeqVal)
	})

}

func (s SeqVal) ConsGroup(suffix Group) Group {
	var head, tail = suffix.Continue()
	// if tail is empty‥.
	if tail.Empty() {
		// if head is none, return original s
		if head.Type().Match(None) {
			return s
		}
		// return a sequence starting with head yielded by prepended
		// seqval, followed by s as its tail
		return SeqVal(func(args ...Expression) (Expression, SeqVal) {
			if len(args) > 0 {
				if len(args) > 1 {
					head, tail = suffix.Call(args...,
					).(Continuation).Continue()
					return head, s.ConsGroup(tail).(SeqVal)
				}
			}
			return head, s
		})
	}
	// tail is not empty yet, return a sequence starting with yielded head
	// followed by remaining tail consed to s recursively
	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
		if len(args) > 0 {
			head, tail = suffix.Call(args...,
			).(Continuation).Continue()
			return head, s.ConsGroup(tail).(SeqVal)
		}
		return head, s.ConsGroup(tail).(SeqVal)
	})

}
func (s SeqVal) ConsSeqVal(seq SeqVal) Group { return s.ConsGroup(seq).(Group) }
func (s SeqVal) Concat(grp Continuation) Group {
	if !s.Empty() {
		return SeqVal(func(args ...Expression) (Expression, SeqVal) {
			if len(args) > 0 {
				var head, tail = s.Cons(args...).Continue()
				return head, tail.Concat(grp).(SeqVal)
			}
			var head, tail = s.Continue()
			return head, tail.Concat(grp).(SeqVal)
		})
	}
	return grp.(Group)
}
func (v SeqVal) Push(arg Expression) Stack { return v.Cons(arg).(SeqVal) }
func (v SeqVal) Pop() (Expression, Stack)  { return v() }

func (s SeqVal) First() Expression { return s.Head() }

func (s SeqVal) Empty() bool {
	if head, tail := s(); tail == nil && head.Type().Match(None) {
		return true
	}
	return false
}

// sort takes a compare function to return a negative value, when a is lesser
// then b, zero when they are equal and a positive value, when a is greater
// then b and generates a sequence of sorted elements implementing quick sort
// by lazy list comprehension (split elements lesser & greater pivot).
func (s SeqVal) Sort(compare func(a, b Expression) int) Group {
	// if list is empty, or has a single element, just return it
	if s.Empty() || s.Tail().Empty() {
		return s
	}
	var ( // use head as pivot & tail to filter lesser & greater from
		pivot, list = s()
		lesser      = NewTest(
			func(arg Expression) bool {
				if compare(pivot, arg) < 0 {
					return true
				}
				return false
			})
		greater = NewTest(
			func(arg Expression) bool {
				if compare(pivot, arg) >= 0 {
					return true
				}
				return false
			})
	)
	// lazy list comprehension quick sort
	return Pass(list, lesser).(SeqVal).Sort(compare).
		Cons(pivot).ConsGroup(
		Pass(list, greater).(SeqVal).Sort(compare)).(Group)
}

func (s SeqVal) Null() SeqVal     { return NewSequence() }
func (s SeqVal) TypeElem() TyComp { return s.Head().Type() }
func (s SeqVal) TypeFnc() TyFnc   { return Sequence }
func (s SeqVal) Type() TyComp     { return Def(Sequence, s.TypeElem()) }
func (s SeqVal) Call(args ...Expression) Expression {
	if len(args) > 0 {
		var head, tail = s(args...)
		return NewPair(head, tail)
	}
	var head, tail = s()
	return NewPair(head, tail)
}
func (s SeqVal) Slice() []Expression {
	var (
		slice      []Expression
		head, tail = s()
	)
	for !head.Type().Match(None) && !tail.Empty() {
		slice = append(slice, head)
		head, tail = tail()
	}
	return slice
}
func (s SeqVal) Vector() VecVal { return NewVector(s.Slice()...) }

func (s SeqVal) String() string {
	if s.Empty() {
		return "()"
	}
	var (
		hstr, tstr string
		head, tail = s()
	)
	for !tail.Empty() {
		hstr = hstr + "(" + head.String() + " "
		tstr = tstr + ")"
		head, tail = tail()
	}
	hstr = hstr + "(" + head.String()
	tstr = tstr + ")"
	return hstr + tstr
}

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
func Flatten(grp Continuation) Continuation {
	if grp.Empty() {
		return NewSequence()
	}
	if grp.Head().Type().Match(Sequences) {
		return Flatten(
			grp.Head().(Continuation),
		).Concat(
			Flatten(grp.Tail()))
	}
	var head, tail = grp.Continue()
	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
		if len(args) > 0 {
			return head, Flatten(tail.Cons(args...)).(SeqVal)
		}
		return head, Flatten(tail).(SeqVal)
	})
}

// map returns a continuation calling the map function for every element
func Map(
	con Continuation,
	mapf func(Expression) Expression,
) Group {
	if con.Empty() {
		return SeqVal(func(args ...Expression) (Expression, SeqVal) {
			if len(args) > 0 {
				var head, tail = Map(
					NewSequence(args...), mapf,
				).Continue()
				return head, tail.(SeqVal)
			}
			return NewNone(), nil
		})
	}
	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
		if len(args) > 0 {
			con = con.Call(args...).(Continuation)
		}
		var head, tail = con.Continue()
		// skip none instances, when tail has further elements
		if head.Type().Match(None) && !tail.Empty() {
			return Map(tail, mapf).(SeqVal)()
		}
		return mapf(head), Map(tail, mapf).(SeqVal)
	})
}

// apply returns a continuation called on every element of the continuation.
// when continuation is called pssing arguments those, are passed to apply
// alongside the current element
func Apply(
	con Continuation,
	apply func(Expression, ...Expression) Expression,
) SeqVal {
	if con.Empty() {
		return SeqVal(func(args ...Expression) (Expression, SeqVal) {
			if len(args) > 0 {
				var head, tail = Apply(
					NewSequence(args...), apply,
				).Continue()
				return head, tail.(SeqVal)
			}
			return NewNone(), nil
		})
	}
	var head, tail = con.Continue()
	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
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
	con Continuation,
	init Expression,
	fold func(init, head Expression) Expression,
) SeqVal {
	if con.Empty() {
		return SeqVal(func(args ...Expression) (Expression, SeqVal) {
			if len(args) > 0 {
				var head, tail = Fold(
					NewSequence(args...), init, fold,
				).Continue()
				return head, tail.(SeqVal)
			}
			return NewNone(), nil
		})
	}
	var (
		head, tail = con.Continue()
		result     = fold(init, head)
	)
	// skip none instances, when tail has further elements
	if result.Type().Match(None) && !tail.Empty() {
		for result.Type().Match(None) && !tail.Empty() {
			head, tail = tail.Continue()
			result = fold(init, head)
		}
	}
	init = result
	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
		if len(args) > 0 {
			return init.Call(args...), Fold(tail, init, fold)
		}
		return init, Fold(tail, init, fold)
	})
}

// continuation of elements not matched by test
func Filter(
	con Continuation,
	filter func(Expression) bool,
) SeqVal {
	if con.Empty() {
		return SeqVal(func(args ...Expression) (Expression, SeqVal) {
			if len(args) > 0 {
				var head, tail = Filter(
					NewSequence(args...), filter,
				).Continue()
				return head, tail.(SeqVal)
			}
			return NewNone(), nil
		})
	}
	var (
		init = NewSequence()
		fold = func(init, head Expression) Expression {
			if filter(head) {
				return NewNone()
			}
			return init.(SeqVal).Cons(head)
		}
	)
	return Fold(con, init, fold)
}

// continuation of elements matched by test
func Pass(
	con Continuation,
	pass func(Expression) bool,
) Group {
	if con.Empty() {
		return SeqVal(func(args ...Expression) (Expression, SeqVal) {
			if len(args) > 0 {
				var head, tail = Pass(
					NewSequence(args...), pass,
				).Continue()
				return head, tail.(SeqVal)
			}
			return NewNone(), nil
		})
	}
	var (
		init = NewSequence()
		fold = func(init, head Expression) Expression {
			if pass(head) {
				return init.(SeqVal).Cons(head)
			}
			return NewNone()
		}
	)
	return Fold(con, init, fold)
}

// take-n is a variation of fold that takes an initial continuation cuts and
// returns it as continuation of vector instances of length n
func TakeN(grp Continuation, n int) Continuation {
	if grp.Empty() {
		return grp
	}
	var (
		init  = NewVector()
		takeN = func(init Expression, arg Expression) Expression {
			var (
				vector = init.(VecVal)
			)
			if vector.Len() == n {
				return NewVector(arg)
			}
			return vector.Cons(arg)
		}
	)
	return Filter(Fold(grp, init, takeN),
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
	left, right Continuation,
	zip func(l, r Expression) Expression,
) SeqVal {
	if left.Empty() && right.Empty() {
		return SeqVal(func(args ...Expression) (Expression, SeqVal) {
			return NewNone(), nil
		})
	}
	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
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
	con Continuation,
	pair Paired,
	split func(Paired, Expression) Paired,
) SeqVal {
	return Fold(con, pair, func(init, head Expression) Expression {
		var pair = init.(Paired)
		return split(pair, head)
	})
}

// bind works similar to zip, but the bind function takes additional arguments
// during runtime and passes them with the heads of both lists passed to the
// call method (instead of passing on to results call method, analog to
// map/apply).  when both functions are curryed and the arguments are passed to
// the resulting function, bind behaves like the '.' operator in haskell.
func Bind(
	f, g Continuation,
	bind func(l, r Expression, args ...Expression) Expression,
) SeqVal {
	if f.Empty() || g.Empty() {
		return SeqVal(func(args ...Expression) (Expression, SeqVal) {
			return NewNone(), nil
		})
	}
	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
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
//func Sort(
//	grp Group,
//	less func(l, r Expression) bool,
//) SeqVal {
//	if grp.Empty() {
//		return NewSequence()
//	}
//	var (
//		pivot = grp.Head()
//	)
//}
