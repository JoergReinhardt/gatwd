package functions

import ()

type (

	//// COLLECTIONS
	GroupVal func(...Expression) (Expression, Grouped) // interface type
	ListVal  func(...Expression) (Expression, ListVal) // instance type
)

///////////////////////////////////////////////////////////////////////////////
//// GENERIC SEQUENCE TYPE
///
// generic sequential type
func NewGroupFromContinuation(con Continued) Grouped {
	return GroupVal(func(args ...Expression) (Expression, Grouped) {
		var (
			head Expression
			tail Grouped
		)
		if len(args) > 0 {
			head, tail = con.Call(
				args...,
			).(Continued).Continue()
			return head, NewGroup(tail)
		}
		head, tail = con.Continue()
		return head, NewGroup(tail)
	})
}

func NewGroup(args ...Expression) Grouped {
	return GroupVal(func(args ...Expression) (Expression, Grouped) {
		var (
			head Expression
			tail Grouped
		)
		if len(args) > 0 {
			head, tail = NewList(args...).Continue()
			return head, NewGroup(tail)
		}
		head, tail = NewList(args...).Continue()
		return head, NewGroup(tail)
	})
}

func (g GroupVal) Continue() (Expression, Grouped) { return g() }
func (g GroupVal) Head() Expression                { var head, _ = g(); return head }
func (g GroupVal) Tail() Grouped                   { var _, tail = g(); return tail }
func (g GroupVal) Cons(args ...Expression) Grouped {
	if len(args) == 0 {
		return g
	}
	if len(args) == 1 {
		return GroupVal(func(late ...Expression) (Expression, Grouped) {
			if len(late) > 0 {
				if len(late) > 1 {
					return late[0],
						g.Cons(append(late[1:], args[0])...).(ListVal)
				}
				return late[0], g.Cons(args[0]).(GroupVal)
			}
			return args[0], g
		})
	}
	return GroupVal(func(late ...Expression) (Expression, Grouped) {
		if len(late) > 0 {
			if len(late) > 1 {
				return late[0],
					g.Cons(append(late[1:], args...)...).(GroupVal)
			}
			return late[0], g.Cons(args...).(GroupVal)
		}
		return args[0], g.Cons(args[1:]...).(GroupVal)
	})

}
func (s GroupVal) Concat(grp Continued) Grouped {
	if !s.Empty() {
		return GroupVal(func(args ...Expression) (Expression, Grouped) {
			if len(args) > 0 {
				var head, tail = s.Cons(args...).Continue()
				return head, tail.Concat(grp).(GroupVal)
			}
			var head, tail = s.Continue()
			return head, tail.Concat(grp).(GroupVal)
		})
	}
	return grp.(Grouped)
}
func (g GroupVal) Call(args ...Expression) Expression {
	var (
		head Expression
		tail Grouped
	)
	if len(args) > 0 {
		head, tail = g(args...)
		return NewPair(head, tail)
	}
	return NewPair(g())
}
func (g GroupVal) Empty() bool {
	var head, tail = g()
	return head.Type().Match(None) && tail == nil
}
func (g GroupVal) String() string {
	if g.Empty() {
		return "()"
	}
	var (
		hstr, tstr string
		head, tail = g()
	)
	for !tail.Empty() {
		hstr = hstr + "(" + head.String() + " "
		tstr = tstr + ")"
		head, tail = tail.Continue()
	}
	hstr = hstr + "(" + head.String()
	tstr = tstr + ")"
	return hstr + tstr
}
func (g GroupVal) TypeFnc() TyFnc   { return Group }
func (g GroupVal) TypeElem() TyComp { return g.Head().Type() }
func (g GroupVal) Type() TyComp     { return Def(Group, g.TypeElem(), g.TypeElem()) }

///////////////////////////////////////////////////////////////////////////////
//// LINKED LIST TYPE
///
// linked list type implementing sequential
func NewListFromGroup(grp Grouped) ListVal {
	return ListVal(func(args ...Expression) (Expression, ListVal) {
		if len(args) > 0 {
			grp = grp.Cons(args...)
		}
		var head, tail = grp.Continue()
		return head, NewListFromGroup(tail)
	})
}
func NewList(elems ...Expression) ListVal {

	// return empty list able to be extended by cons, when no initial
	// elements are given/left
	if len(elems) == 0 {
		return func(args ...Expression) (Expression, ListVal) {
			if len(args) > 0 {
				if len(args) > 1 {
					return args[0], NewList(args[1:]...)
				}
				return args[0], NewList()
			}
			// return instance of none as head and a nil pointer as
			// tail, if neither elements nor arguments where passed
			return NewNone(), nil
		}
	}

	// at least one of the initial elements is left‥.
	return func(args ...Expression) (Expression, ListVal) {

		// if arguments are passed, prepend those and return first
		// argument as head‥.
		if len(args) > 0 {
			// ‥.put arguments up front of preceeding elements
			if len(args) > 1 {
				return args[0], NewList(
					append(
						args,
						elems...,
					)...)
			}
			// use single argument as new head of sequence and
			// preceeding elements as tail
			return args[0],
				NewList(elems...)
		}

		// no arguments given, but more than one element left → return
		// first element as head, and remaining elements as tail of
		// sequence
		if len(elems) > 1 {
			return elems[0],
				NewList(elems[1:]...)
		}
		// return last element and empty sequence
		return elems[0], NewList()

	}
}
func (s ListVal) Head() Expression {
	var cur, _ = s()
	return cur
}
func (s ListVal) Tail() Grouped {
	var _, tail = s()
	return tail
}
func (s ListVal) Continue() (Expression, Grouped) {
	return s.Head(), s.Tail()
}
func (s ListVal) Cons(suffix ...Expression) Grouped {
	if len(suffix) == 0 {
		return s
	}
	if len(suffix) == 1 {
		return ListVal(func(late ...Expression) (Expression, ListVal) {
			if len(late) > 0 {
				if len(late) > 1 {
					return late[0],
						s.Cons(append(late[1:], suffix[0])...).(ListVal)
				}
				return late[0], s.Cons(suffix[0]).(ListVal)
			}
			return suffix[0], s
		})
	}
	return ListVal(func(late ...Expression) (Expression, ListVal) {
		if len(late) > 0 {
			if len(late) > 1 {
				return late[0],
					s.Cons(append(late[1:], suffix...)...).(ListVal)
			}
			return late[0], s.Cons(suffix...).(ListVal)
		}
		return suffix[0], s.Cons(suffix[1:]...).(ListVal)
	})

}

func (s ListVal) ConsGroup(suffix Grouped) Grouped {
	var head, tail = suffix.Continue()
	// if tail is empty‥.
	if tail.Empty() {
		// if head is none, return original s
		if head.Type().Match(None) {
			return s
		}
		// return a sequence starting with head yielded by prepended
		// seqval, followed by s as its tail
		return ListVal(func(args ...Expression) (Expression, ListVal) {
			if len(args) > 0 {
				if len(args) > 1 {
					head, tail = suffix.Call(args...,
					).(Continued).Continue()
					return head, s.ConsGroup(tail).(ListVal)
				}
			}
			return head, s
		})
	}
	// tail is not empty yet, return a sequence starting with yielded head
	// followed by remaining tail consed to s recursively
	return ListVal(func(args ...Expression) (Expression, ListVal) {
		if len(args) > 0 {
			head, tail = suffix.Call(args...,
			).(Continued).Continue()
			return head, s.ConsGroup(tail).(ListVal)
		}
		return head, s.ConsGroup(tail).(ListVal)
	})

}
func (s ListVal) ConsSeqVal(seq ListVal) Grouped { return s.ConsGroup(seq).(Grouped) }
func (s ListVal) Concat(grp Continued) Grouped {
	if !s.Empty() {
		return ListVal(func(args ...Expression) (Expression, ListVal) {
			if len(args) > 0 {
				var head, tail = s.Cons(args...).Continue()
				return head, tail.Concat(grp).(ListVal)
			}
			var head, tail = s.Continue()
			return head, tail.Concat(grp).(ListVal)
		})
	}
	return grp.(Grouped)
}
func (v ListVal) Push(arg Expression) Stack { return v.Cons(arg).(ListVal) }
func (v ListVal) Pop() (Expression, Stack)  { return v() }

func (s ListVal) First() Expression { return s.Head() }

func (s ListVal) Empty() bool {
	var head, tail = s()
	return head.Type().Match(None) && tail == nil
}

func (s ListVal) Null() ListVal    { return NewList() }
func (s ListVal) TypeElem() TyComp { return s.Head().Type() }
func (s ListVal) TypeFnc() TyFnc   { return Group }
func (s ListVal) Type() TyComp     { return Def(Group, s.TypeElem()) }
func (s ListVal) Call(args ...Expression) Expression {
	if len(args) > 0 {
		var head, tail = s(args...)
		return NewPair(head, tail)
	}
	var head, tail = s()
	return NewPair(head, tail)
}
func (s ListVal) Slice() []Expression {
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
func (s ListVal) Vector() VecVal { return NewVector(s.Slice()...) }

func (s ListVal) String() string {
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
func Flatten(grp Continued) Continued {
	if grp.Empty() {
		return NewList()
	}
	if grp.Head().Type().Match(Sequences) {
		return Flatten(
			grp.Head().(Continued),
		).Concat(
			Flatten(grp.Tail()))
	}
	var head, tail = grp.Continue()
	return ListVal(func(args ...Expression) (Expression, ListVal) {
		if len(args) > 0 {
			return head, Flatten(tail.Cons(args...)).(ListVal)
		}
		return head, Flatten(tail).(ListVal)
	})
}

// map returns a continuation calling the map function for every element
func Map(
	con Continued,
	mapf func(Expression) Expression,
) Grouped {
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
			return Map(tail, mapf).(ListVal)()
		}
		return mapf(head), Map(tail, mapf).(ListVal)
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
) Expression {
	if con.Empty() {
		return init
	}
	//	if con.Empty() {
	//		return SeqVal(func(args ...Expression) (Expression, SeqVal) {
	//			if len(args) > 0 {
	//				var head, tail = Fold(
	//					NewSequence(args...), init, fold,
	//				).Continue()
	//				return head, tail.(SeqVal)
	//			}
	//			return NewNone(), nil
	//		})
	//	}
	var (
		head, tail = con.Continue()
		result     = fold(init, head)
	)
	// skip none instances, when tail has further elements
	if result.Type().Match(None) || !tail.Empty() {
		for result.Type().Match(None) && !tail.Empty() {
			head, tail = tail.Continue()
			result = fold(init, head)
		}
	}
	return ListVal(func(args ...Expression) (Expression, ListVal) {
		if len(args) > 0 {
			return result.Call(args...), Fold(tail, result, fold).(ListVal)
		}
		return result, Fold(tail, result, fold).(ListVal)
	})
}

// continuation of elements not matched by test
func Filter(
	con Continued,
	filter func(Expression) bool,
) ListVal {
	//	if con.Empty() {
	//		return SeqVal(func(args ...Expression) (Expression, SeqVal) {
	//			if len(args) > 0 {
	//				var head, tail = Filter(
	//					NewSequence(args...), filter,
	//				).Continue()
	//				return head, tail.(SeqVal)
	//			}
	//			return NewNone(), nil
	//		})
	//	}
	var (
		init = NewList()
		fold = func(init, head Expression) Expression {
			if filter(head) {
				return NewNone()
			}
			return init.(ListVal).Cons(head)
		}
	)
	return Fold(con, init, fold).(ListVal)
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
		init = NewList()
		fold = func(init, head Expression) Expression {
			if pass(head) {
				return init.(ListVal).Cons(head)
			}
			return NewNone()
		}
	)
	return Fold(con, init, fold).(ListVal)
}

// take-n is a variation of fold that takes an initial continuation cuts and
// returns it as continuation of vector instances of length n
func TakeN(grp Continued, n int) Continued {
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
	return Filter(Fold(grp, init, takeN).(Continued),
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
	pair Paired,
	split func(Paired, Expression) Paired,
) ListVal {
	return Fold(con, pair, func(init, head Expression) Expression {
		var pair = init.(Paired)
		return split(pair, head)
	}).(ListVal)
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
