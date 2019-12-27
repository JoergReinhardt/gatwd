package functions

type (
	//// GENERATOR | ACCUMULATOR
	GenVal func() (Expression, GenVal)
	AccVal func(...Expression) (Expression, AccVal)
)

///////////////////////////////////////////////////////////////////////////////
//// COMPOSITION PRIMITIVES
///
// define the curryed function
func Curry(f, g FuncVal) FuncVal {
	if f.TypeArgs().Match(g.TypeRet()) {
		return Define(Lambda(
			func(args ...Expression) Expression {
				if len(args) > 0 {
					return f.Call(g.Call(args...))
				}
				return f.Call(g.Call())
			}),
			Def(
				f.TypeId(),
				g.TypeId()),
			f.TypeRet(),
			f.TypeArgs(),
		)
	}
	return Define(NewNone(), None, None)
}

//// GENERATOR
///
// expects an expression that returns an unboxed value, when called empty and
// some notion of 'next' value, relative to its arguments, if arguments where
// passed.
func NewGenerator(init, generate Expression) GenVal {
	return func() (Expression, GenVal) {
		var next = generate.Call(init)
		return init, NewGenerator(next, generate)
	}
}
func (g GenVal) Expr() Expression {
	var expr, _ = g()
	return expr
}
func (g GenVal) Generator() GenVal {
	var _, gen = g()
	return gen
}
func (g GenVal) Call(args ...Expression) Expression {
	if len(args) > 0 {
		return NewPair(g.Expr().Call(args...), g.Generator())
	}
	return NewPair(g.Expr(), g.Generator())
}
func (g GenVal) TypeFnc() TyFnc   { return Generator }
func (g GenVal) Type() TyComp     { return Def(Generator, g.Head().Type()) }
func (g GenVal) TypeElem() TyComp { return g.Head().Type() }
func (g GenVal) String() string   { return g.Head().String() }
func (g GenVal) Empty() bool {
	if g.Head().Type().Match(None) {
		return true
	}
	return false
}
func (g GenVal) Continue() (Expression, Continuation) { return g() }
func (g GenVal) Head() Expression                     { return g.Expr() }
func (g GenVal) Tail() Continuation                   { return g.Generator() }

//// ACCUMULATOR
///
// accumulator expects an expression as input, that returns itself unboxed,
// when called empty and returns a new accumulator accumulating its value and
// arguments to create a new accumulator, if arguments where passed.
func NewAccumulator(acc, fnc Expression) AccVal {
	return AccVal(func(args ...Expression) (Expression, AccVal) {
		if len(args) > 0 {
			acc = fnc.Call(append([]Expression{acc}, args...)...)
			return acc, NewAccumulator(acc, fnc)
		}
		return acc, NewAccumulator(acc, fnc)
	})
}

func (g AccVal) Result() Expression {
	var res, _ = g()
	return res
}
func (g AccVal) Accumulator() AccVal {
	var _, acc = g()
	return acc
}
func (g AccVal) Call(args ...Expression) Expression {
	if len(args) > 0 {
		var res, acc = g(args...)
		return NewPair(res, acc)
	}
	return g.Result()
}
func (g AccVal) TypeFnc() TyFnc { return Accumulator }
func (g AccVal) Type() TyComp {
	return Def(
		Accumulator,
		g.Head().Type().TypeRet(),
		g.Head().Type().TypeArgs(),
	)
}
func (g AccVal) String() string { return g.Head().String() }

func (a AccVal) Empty() bool {
	if a.Head().Type().Match(None) {
		return true
	}
	return false
}
func (g AccVal) Head() Expression                     { return g.Result() }
func (g AccVal) TypeElem() TyComp                     { return g.Head().Type() }
func (g AccVal) Tail() Continuation                   { return g.Accumulator() }
func (g AccVal) Continue() (Expression, Continuation) { return g() }

///////////////////////////////////////////////////////////////////////////////
//// CONTINUATION COMPOSITION
///
// flatten flattens sequences of sequences to one dimension
//func Flatten(con Continuation) SeqVal {
//	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
//		if len(args) > 0 {
//			con = con.Call(args...).(Continuation)
//		}
//		var head, tail = con.Continue()
//		if head.Type().Match(Sequences) {
//			head, tail = head.(Sequential).ConsSeq(tail).Continue()
//			return head, tail.(SeqVal)
//		}
//	})
//}

// map returns a continuation calling the map function for every element
func Map(
	con Continuation,
	mapf func(Expression) Expression,
) SeqVal {
	if con.Empty() {
		return SeqVal(func(args ...Expression) (Expression, SeqVal) {
			if len(args) > 0 {
				con = con.Call(args...).(Sequential)
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
			return Map(tail, mapf)()
		}
		return mapf(head), Map(tail, mapf)
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
		return Apply(con, apply)
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

// fold takes a continuation, an initial expression and a fold function. the
// fold function is called for every element of the continuation and passed the
// current element and init expression and returns a possbly altered init
// element to pass to next call
func Fold(
	con Continuation,
	init Expression,
	fold func(init, head Expression) Expression,
) SeqVal {
	if con.Empty() {
		return Fold(con, init, fold)
	}
	var head, tail = con.Continue()
	// skip none instances, when tail has further elements
	if head.Type().Match(None) && !tail.Empty() {
		return Fold(tail, init, fold)
	}
	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
		init = fold(init, head)
		if len(args) > 0 {
			init = fold(init, head).Call(args...)
			return init, Fold(tail, init, fold)
		}
		return init, Fold(tail, init, fold)
	})
}

// continuation of elements not matched by test
func Filter(
	con Continuation,
	filter Testable,
) SeqVal {
	if con.Empty() {
		return Filter(con, filter)
	}
	var (
		init = NewSequence()
		fold = func(init, head Expression) Expression {
			if filter.Test(head) {
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
	pass Testable,
) SeqVal {
	if con.Empty() {
		return Pass(con, pass)
	}
	var (
		init = NewSequence()
		fold = func(init, head Expression) Expression {
			if pass.Test(head) {
				return init.(SeqVal).Cons(head)
			}
			return NewNone()
		}
	)
	return Fold(con, init, fold)
}

// take-n is a variation of fold that takes an initial continuation cuts and
// returns it as continuation of vector instances of length n
func TakeN(con Continuation, n int) SeqVal {
	if con.Empty() {
		return TakeN(con, n)
	}
	var (
		init = NewPair(NewVector(), NewVector())
		take = func(init, head Expression) Expression {
			var (
				pair = init.(Paired)
				vec  = pair.Left().(VecVal)
			)
			if vec.Len() < n {
				return NewPair(
					vec.Cons(head),
					pair.Right(),
				)
			}
			return NewPair(
				NewVector(head),
				pair.Right().(VecVal).Cons(pair.Left()),
			)
		}
	)
	// takeN returns a pair with the current accumulator as left and the
	// continuation of completed accumulations as right element for each
	// element of the initial continuation. to only return complete tokens
	// as left elements, initial output needs to be filtered by testing for
	// correct length of the left element
	return Filter( // filter out incomplete tokens
		Fold(con, init, take), // returns possibly incomplete element per call
		NewTest(func(arg Expression) bool { // tests if current element is complete
			var vec = arg.(Paired).Left().(VecVal)
			if vec.Len() < n {
				return true
			}
			return false
		}))
}

// split is a variation of fold that splits either a continuation of pairs, or
// takes two arguments at a time and splits those into continuation of left and
// right values and returns those as elements of a pair
func Split(con Continuation) SeqVal {
	var (
		init  = NewPair(NewVector(), NewVector())
		split = func(init, head Expression) Expression {
			var (
				pair = head.(Paired)
				pl   = init.(Paired)
				vl   = pl.Left().(VecVal)
				vr   = pl.Right().(VecVal)
			)
			return NewPair(
				vl.Cons(pair.Left()),
				vr.Cons(pair.Right()),
			)
		}
	)
	// split function expects a list of pairs‥.
	if con.TypeElem().Match(Pair) {
		return Fold(con, init, split)
	}
	// ‥.which is created by mapping take2 to a function that converts the
	// resulting slices of length two into pairs
	return Fold(
		Map(TakeN(con, 2), func(arg Expression) Expression {
			var vec = arg.(VecVal)()
			return NewPair(vec[0], vec[1])
		}), init, split)
}

// bind creates a list of results from calling the bind function and passing
// the head elements of both lists.
func Bind(
	left, right Continuation,
	bind func(f, g Expression) Expression,
) SeqVal {
	if left.Empty() || right.Empty() {
		return Bind(left, right, bind)
	}
	var (
		lhead, ltail = left.Continue()
		rhead, rtail = right.Continue()
		current      = bind(rhead, lhead)
		next         = Bind(ltail, rtail, bind)
	)
	// skip none heads, when both continuations still have elements
	if (lhead.Type().Match(None) || rhead.Type().Match(None)) &&
		(!ltail.Empty() && !rtail.Empty()) {
		return Bind(ltail, rtail, bind)
	}
	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
		if len(args) > 0 {
			return current.Call(args...), next
		}
		return current, next
	})
}

// zip is a variation of bind, that creates a list of pairs with left and right
// element taken from the respective lists.
func Zip(
	left, right Continuation,
	zip func(l, r Expression) Expression,
) SeqVal {
	return Bind(left, right, zip)
}
