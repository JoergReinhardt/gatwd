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
func Curry(f, g FuncDecl) FuncDecl {
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
//
func Map(
	con Continuation,
	mapf func(Expression) Expression,
) SeqVal {
	if con.Empty() {
		return Map(con, mapf)
	}
	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
		if len(args) > 0 {
			con = con.Call(args...).(Continuation)
		}
		var head, tail = con.Continue()
		return mapf(head), Map(tail, mapf)
	})
}

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
			return apply(head, args...),
				Apply(tail, apply)
		}
		return apply(head), Apply(tail, apply)
	})
}

func Fold(
	con Continuation,
	init Expression,
	fold func(init, head Expression) Expression,
) SeqVal {
	if con.Empty() {
		return Fold(con, init, fold)
	}
	var head, tail = con.Continue()
	if head.Type().Match(None) {
		if !tail.Empty() {
			return Fold(tail, init, fold)
		}
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
				return init
			}
			return init.(SeqVal).Cons(head)
		}
	)
	return Fold(con, init, fold)
}

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
			return init
		}
	)
	return Fold(con, init, fold)
}

func TakeN(con Continuation, n int) SeqVal {
	if con.Empty() {
		return TakeN(con, n)
	}
	var (
		init = NewPair(NewVector(), NewVector())
		fold = func(init, head Expression) Expression {
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
	return Fold(con, init, fold)
}

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
	if con.TypeElem().Match(Pair) {
		return Fold(con, init, split)
	}
	con = Map(
		TakeN(con, 2),
		func(arg Expression) Expression {
			var slice = arg.(VecVal)()
			return NewPair(slice[0], slice[1])
		},
	)
	return Fold(con, init, split)
}

func Bind(
	m, n Continuation,
	bind func(f, g Expression) Expression,
) SeqVal {
	if m.Empty() || n.Empty() {
		return Bind(m, n, bind)
	}
	var (
		mh, mt = m.Continue()
		nh, nt = n.Continue()
		head   = bind(nh, mh)
		bound  = Bind(mt, nt, bind)
	)
	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
		if len(args) > 0 {
			return head.Call(args...), bound
		}
		return head, bound
	})
}

func Zip(
	left, right Continuation,
	zip func(l, r Expression) Expression,
) SeqVal {
	return Bind(left, right, zip)
}

//func (s SeqVal) Map(mapf Expression) Sequential {
//	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
//		var (
//			head, tail = s()
//			lst        Expression
//		)
//		if len(args) > 0 {
//			lst = args[len(args)-1]
//			// cross product, if last argument is a functor
//			if lst.Type().Match(Functors) {
//				if arg, ok := lst.(Functorial); ok {
//					if len(args) > 1 {
//						return s.MapX(head.Call(args...),
//							mapf, arg), tail.Map(mapf).(SeqVal)
//					}
//					return s.MapX(head, mapf, arg), tail.Map(mapf).(SeqVal)
//				}
//			}
//			// dot product, since last argument is not a functor
//			return mapf.Call(head.Call(args...)), tail.Map(mapf).(SeqVal)
//		}
//		// no arguments given
//		return mapf.Call(head), tail.Map(mapf).(SeqVal)
//	})
//}
//
//func (s SeqVal) MapX(head, mapf Expression, arg Continuation) Sequential {
//	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
//		// check if current head of parent list is none
//		// yield step & next continuation from argument
//		var step, next = arg.Continue()
//		if len(args) > 0 { // if args have been passed
//			// call mapf with current parent lists head &
//			// arguments passed during call to get step.
//			// s-map tail of sequential argument
//			return mapf.Call(head, step.Call(args...)),
//				s.MapX(head, mapf, next).(SeqVal)
//		}
//		return mapf.Call(head, step), NewSequence()
//	})
//}

//func (s SeqVal) Flatten() SeqVal {
//	var head, tail = s()
//	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
//		if head.Type().Match(Sequences) {
//			if seq, ok := head.(Sequential); ok {
//				seq = NewSeqCont(seq).Flatten().ConcatVal(tail.Flatten())
//				return seq.Current(), NewSeqCont(seq.Next())
//			}
//		}
//		return head, tail
//	})
//}
//
//func (s SeqVal) Fold(
//	acc Expression,
//	fold func(acc, head Expression) Expression,
//) SeqVal {
//	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
//		var (
//			result     Expression
//			head, tail = s()
//		)
//		if head.TypeFnc().Match(None) {
//			return acc, tail
//		}
//		result = fold(acc, head)
//		if len(args) > 0 {
//			return result.Call(args...), tail.Fold(result, fold)
//		}
//		return result, tail.Fold(result, fold)
//	})
//}
//
//func (s SeqVal) Filter(test Testable) Sequential {
//	var (
//		seq        = NewSequence()
//		head, tail = s()
//	)
//	if head.TypeFnc().Match(None) {
//		return NewSequence()
//	}
//	if !test.Test(head) {
//		return seq.Concat(head).(SeqVal).ConcatVal(tail.Filter(test).(SeqVal))
//	}
//	return seq.ConcatVal(tail.Filter(test).(SeqVal))
//}
//
//func (s SeqVal) Pass(test Testable) Sequential {
//	var (
//		seq        = NewSequence()
//		head, tail = s()
//	)
//	if head.TypeFnc().Match(None) {
//		return NewSequence()
//	}
//	if test.Test(head) {
//		return seq.Cons(head).Cons(tail.Filter(test))
//	}
//	return seq.Concat(tail.Filter(test))
//}
//
//// application of boxed arguments to boxed functions
//func (s SeqVal) Apply(
//	apply func(
//		seq Sequential,
//		args ...Expression,
//	) (
//		Expression,
//		Continuation,
//	)) Sequential {
//	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
//		if len(args) > 0 {
//			var result, seq = apply(s, args...)
//			return result, NewSeqCont(seq)
//		}
//		var result, seq = apply(s)
//		return result, NewSeqCont(seq)
//	})
//}
//
//// sequential composition of function application
//func (s SeqVal) Bind(bind Expression, cont Continuation) Sequential {
//	var step, next = s()
//	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
//		if len(args) > 0 {
//			return step.Call(cont.Current().Call(args...)),
//				next.Bind(bind, cont.Next()).(SeqVal)
//		}
//		return step.Call(cont.Current()),
//			next.Bind(bind, cont.Next()).(SeqVal)
//	})
//}
//
//func (s SeqVal) ZipWith(
//	zipf func(l, r Continuation) Sequential,
//	cont Continuation,
//) SeqVal {
//	var (
//		leftStep, left   = s()
//		rightStep, right = cont.Continue()
//	)
//	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
//		if leftStep.Type().Match(None) || rightStep.Type().Match(None) {
//			return NewNone(), NewSequence()
//		}
//		if len(args) > 0 {
//			return NewPair(leftStep, rightStep).Call(args...),
//				left.ZipWith(zipf, right)
//		}
//		return NewPair(leftStep, rightStep),
//			left.ZipWith(zipf, right)
//	})
//}
//
//func (s SeqVal) Split() (Sequential, Sequential) {
//	var (
//		head, tail  = s.Continue()
//		left, right = tail.(Zipped).Split()
//	)
//	if head.Type().Match(Pair) { // list of pairs gets zipped into keys & values
//		if pair, ok := head.(Paired); ok {
//			return SeqVal(func(args ...Expression) (Expression, SeqVal) {
//					if len(args) > 0 {
//						return pair.Left().Call(args...), left.(SeqVal)
//					}
//					return pair.Left(), left.(SeqVal)
//				}),
//				SeqVal(func(args ...Expression) (Expression, SeqVal) {
//					if len(args) > 0 {
//						return pair.Right().Call(args...), right.(SeqVal)
//					}
//					return pair.Right(), right.(SeqVal)
//				})
//		}
//	}
//	if !head.Type().Match(None) { // flat lists are split two elements at a step
//		var resl, resr Sequential
//		if !head.Type().Match(None) {
//			resl = SeqVal(func(args ...Expression) (Expression, SeqVal) {
//				if len(args) > 0 {
//					return head.Call(args...), left.(SeqVal)
//				}
//				return head, left.(SeqVal)
//			})
//		} else {
//			resl = NewSequence()
//		}
//		head, tail = tail.Continue()
//		if !head.Type().Match(None) {
//			resr = SeqVal(func(args ...Expression) (Expression, SeqVal) {
//				if len(args) > 0 {
//					return head.Call(args...), right.(SeqVal)
//				}
//				return head, right.(SeqVal)
//			})
//		} else {
//			resr = NewSequence()
//		}
//		return resl, resr
//	}
//	// head is a none value
//	return NewSequence(), NewSequence()
//}
//
