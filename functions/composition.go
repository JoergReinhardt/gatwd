package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

type (
	//// GENERATOR | ACCUMULATOR
	GenVal func() (Expression, GenVal)
	AccVal func(...Expression) (Expression, AccVal)

	//// ENUMERABLE
	EnumDef func(d.Numeral) EnumVal
	EnumVal func(...Expression) (Expression, d.Numeral, EnumDef)

	//// SEQUENCE
	SeqVal func(...Expression) (Expression, SeqVal)

	//// STATE MONADS
	StateD func(args ...d.Native) (d.Native, StateD)
	StateE func(args ...Expression) (Expression, StateE)
)

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
func (g GenVal) Type() TyComp     { return Def(Generator, g.Step().Type()) }
func (g GenVal) TypeElem() TyComp { return g.Step().Type() }
func (g GenVal) String() string   { return g.Step().String() }
func (g GenVal) End() bool {
	if g.Step().Type().Match(None) {
		return true
	}
	return false
}
func (g GenVal) Continue() (Expression, Continuation) { return g() }
func (g GenVal) Step() Expression                     { return g.Expr() }
func (g GenVal) Next() Continuation                   { return g.Generator() }

//// ACCUMULATOR
///
// accumulator expects an expression as input, that returns itself unboxed,
// when called empty and returns a new accumulator accumulating its value and
// arguments to create a new accumulator, if arguments where passed.
func NewAccumulator(init, acc Expression) AccVal {
	return AccVal(func(args ...Expression) (Expression, AccVal) {
		if len(args) > 0 {
			init = acc.Call(append([]Expression{init}, args...)...)
			return init, NewAccumulator(init, acc)
		}
		return init, NewAccumulator(init, acc)
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
		g.Step().Type().TypeReturn(),
		g.Step().Type().TypeArguments(),
	)
}
func (g AccVal) String() string { return g.Step().String() }

func (a AccVal) End() bool {
	if a.Step().Type().Match(None) {
		return true
	}
	return false
}
func (g AccVal) Step() Expression                     { return g.Result() }
func (g AccVal) TypeElem() TyComp                     { return g.Step().Type() }
func (g AccVal) Next() Continuation                   { return g.Accumulator() }
func (g AccVal) Continue() (Expression, Continuation) { return g() }

//// ENUM TYPE
///
// declares an enumerable type returning instances from the set of enumerables
// defined by the passed function
func NewEnumType(fnc func(d.Numeral) Expression) EnumDef {
	return func(idx d.Numeral) EnumVal {
		return func(args ...Expression) (Expression, d.Numeral, EnumDef) {
			if len(args) > 0 {
				return fnc(idx).Call(args...), idx, NewEnumType(fnc)
			}
			return fnc(idx), idx, NewEnumType(fnc)
		}
	}
}
func (e EnumDef) Expr() Expression            { return e(d.IntVal(0)) }
func (e EnumDef) Alloc(idx d.Numeral) EnumVal { return e(idx) }
func (e EnumDef) Type() TyComp {
	return Def(Enum, e.Expr().Type().TypeReturn())
}
func (e EnumDef) TypeFnc() TyFnc { return Enum }
func (e EnumDef) String() string { return e.Type().TypeName() }
func (e EnumDef) Call(args ...Expression) Expression {
	if len(args) > 0 {
		if len(args) > 1 {
			var vec = NewVector()
			for _, arg := range args {
				vec = vec.Cons(e.Call(arg)).(VecVal)
			}
			return vec
		}
		var arg = args[0]
		if arg.Type().Match(Data) {
			if nat, ok := arg.(NatEval); ok {
				if i, ok := nat.Eval().(d.Numeral); ok {
					return e(i)
				}
			}
		}
	}
	return e
}

//// ENUM VALUE
///
//
func (e EnumVal) Expr() Expression {
	var expr, _, _ = e()
	return expr
}
func (e EnumVal) Index() d.Numeral {
	var _, idx, _ = e()
	return idx
}
func (e EnumVal) EnumType() EnumDef {
	var _, _, et = e()
	return et
}
func (e EnumVal) Alloc(idx d.Numeral) EnumVal { return e.EnumType().Alloc(idx) }
func (e EnumVal) Next() EnumVal {
	var result = e.EnumType()(e.Index().Int() + d.IntVal(1))
	return result
}
func (e EnumVal) Previous() EnumVal {
	var result = e.EnumType()(e.Index().Int() - d.IntVal(1))
	return result
}
func (e EnumVal) String() string { return e.Expr().String() }
func (e EnumVal) Type() TyComp {
	var (
		nat d.Native
		idx = e.Index()
	)
	if idx.Type().Match(d.BigInt) {
		nat = idx.BigInt()
	} else {
		nat = idx.Int()
	}
	return Def(Def(Enum, DefValNat(nat)), e.Expr().Type())
}
func (e EnumVal) TypeFnc() TyFnc { return Enum | e.Expr().TypeFnc() }
func (e EnumVal) Call(args ...Expression) Expression {
	var r, _, _ = e(args...)
	return r
}

//// SEQUENCE TYPE
///
// generic sequential type
func NewSequence(elems ...Expression) SeqVal {
	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
		if len(elems) > 0 {
			var head = elems[len(elems)-1]
			if len(elems) > 1 {
				elems = elems[:len(elems)-1]
			} else {
				elems = elems[:0]
			}
			if len(args) > 0 {
				return head.Call(args...), NewSequence(elems...)
			}
			return head, NewSequence(elems...)
		}
		return NewNone(), NewSequence()
	})
}

func NewSeqCont(cont Continuation) SeqVal {
	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
		var head, tail = cont.Continue()
		if len(args) > 0 {
			return head.Call(args...), NewSeqCont(tail)
		}
		return head, NewSeqCont(tail)
	})
}

func (s SeqVal) Call(args ...Expression) Expression {
	if len(args) > 0 {
		return NewPair(s.Step().Call(args...), s.Next())
	}
	return NewPair(s.Step(), s.Next())
}
func (s SeqVal) Continue() (Expression, Continuation) { return s() }
func (s SeqVal) Step() Expression {
	var expr, _ = s()
	return expr
}
func (s SeqVal) NextSeq() SeqVal { return s.Next().(SeqVal) }
func (s SeqVal) Next() Continuation {
	var _, seq = s()
	return seq
}
func (s SeqVal) TypeElem() TyComp { return s.Step().Type() }
func (s SeqVal) TypeFnc() TyFnc   { return Sequence }
func (s SeqVal) Type() TyComp     { return Def(Sequence, s.TypeElem()) }
func (s SeqVal) End() bool {
	if s.Step().Type().Match(None) {
		return true
	}
	return false
}

func (s SeqVal) String() string {
	var (
		hstr, tstr string
		head, tail = s()
	)
	for !head.Type().Match(None) {
		hstr = hstr + "( " + head.String() + " "
		tstr = tstr + ")"
		head, tail = tail()
	}
	return hstr + tstr
}

func (s SeqVal) Concat(elems ...Expression) Sequential {
	return s.ConcatSeq(NewSequence(elems...))
}

func (s SeqVal) ConcatSeq(seq Sequential) SeqVal {
	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
		var head, tail = s()
		if tail.End() {
			if len(args) > 0 {
				return head.Call(args...), NewSeqCont(seq)
			}
			return head, NewSeqCont(seq)
		}
		if len(args) > 0 {
			return head.Call(args...), tail.ConcatSeq(seq)
		}
		return head, tail.ConcatSeq(seq)
	})
}

func (s SeqVal) Cons(elems ...Expression) Sequential { return s.ConsSeq(NewSequence(elems...)) }
func (s SeqVal) ConsSeq(elems ...Expression) SeqVal {
	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
		if len(elems) > 0 {
			var head = elems[len(elems)-1]
			if len(elems) > 1 {
				var tail = elems[1:]
				if len(args) > 0 {
					return head.Call(args...), s.Cons(tail...).(SeqVal)
				}
				return head, s.Cons(tail...).(SeqVal)
			}
			if len(args) > 0 {
				return head.Call(args...), s
			}
			return head, s
		}
		return s()
	})
}

func (s SeqVal) MapF(mapf Expression) Sequential {
	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
		var (
			head, tail = s()
			lst        Expression
		)
		if len(args) > 0 {
			lst = args[len(args)-1]
			// cross product, if last argument is a functor
			if lst.Type().Match(Functors) {
				if arg, ok := lst.(Functoric); ok {
					if len(args) > 1 {
						return s.MapS(head.Call(args...),
							mapf, arg), tail.MapF(mapf).(SeqVal)
					}
					return s.MapS(head, mapf, arg), tail.MapF(mapf).(SeqVal)
				}
			}
			// dot product, since last argument is not a functor
			return mapf.Call(head.Call(args...)), tail.MapF(mapf).(SeqVal)
		}
		// no arguments given
		return mapf.Call(head), tail.MapF(mapf).(SeqVal)
	})
}

func (s SeqVal) MapS(head, mapf Expression, arg Continuation) Sequential {
	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
		// check if current head of parent list is none
		// yield step & next continuation from argument
		var step, next = arg.Continue()
		if len(args) > 0 { // if args have been passed
			// call mapf with current parent lists head &
			// arguments passed during call to get step.
			// s-map tail of sequential argument
			return mapf.Call(head, step.Call(args...)),
				s.MapS(head, mapf, next).(SeqVal)
		}
		return mapf.Call(head, step), NewSequence()
	})
}

func (s SeqVal) Flatten() SeqVal {
	var head, tail = s()
	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
		if head.Type().Match(Sequences) {
			if seq, ok := head.(Sequential); ok {
				seq = NewSeqCont(seq).Flatten().ConcatSeq(tail.Flatten())
				return seq.Step(), NewSeqCont(seq.Next())
			}
		}
		return head, tail
	})
}

func (s SeqVal) FoldL(acc Expression, fold func(...Expression) Expression) Sequential {
	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
		var (
			result     Expression
			head, tail = s()
		)
		// try to yield accumulate from applying head
		// to arguments and folding the result‥.
		if len(args) > 0 {
			result = fold(append(
				[]Expression{acc, head},
				args...)...)
		} else {
			result = fold(acc, head)
		}
		// skip result, if it is none
		for result.Type().Match(None) {
			// when parent list ist empty → return
			// final result and mapped list
			if head.Type().Match(None) {
				return acc, s.FoldL(acc, fold).(SeqVal)
			}
			// yield next head/tail
			head, tail = tail()
			// compute next result
			result = fold(append([]Expression{acc, head}, args...)...)
		}
		// result is not none → return result and folded tail
		return result, tail.FoldL(result, fold).(SeqVal)
	})
}

func (s SeqVal) Filter(test Testable) Sequential {
	var (
		list   = NewSequence()
		filter = CaseDef(func(args ...Expression) Expression {
			if test.Test(args...) {
				return NewNone()
			}
			if len(args) > 1 {
				return NewVector(args...)
			}
			return args[0]
		})
	)
	return s.FoldL(list, filter)
}

func (s SeqVal) Pass(test Testable) Sequential {
	var (
		list   = NewSequence()
		filter = CaseDef(func(args ...Expression) Expression {
			if !test.Test(args...) {
				return NewNone()
			}
			if len(args) > 1 {
				return NewVector(args...)
			}
			return args[0]

		})
	)
	return s.FoldL(list, filter)
}

// application of boxed arguments to boxed functions
func (s SeqVal) Apply(
	apply func(
		seq Sequential,
		args ...Expression,
	) (
		Expression,
		Continuation,
	)) Sequential {
	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
		if len(args) > 0 {
			var result, seq = apply(s, args...)
			return result, NewSeqCont(seq)
		}
		var result, seq = apply(s)
		return result, NewSeqCont(seq)
	})
}

// sequential composition of function application
func (s SeqVal) BindM(bindf Expression, cont Continuation) Sequential {
	var step, next = s()
	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
		if len(args) > 0 {
			return step.Call(cont.Step().Call(args...)),
				next.BindM(bindf, cont.Next()).(SeqVal)
		}
		return step.Call(cont.Step()),
			next.BindM(bindf, cont.Next()).(SeqVal)
	})
}

func (s SeqVal) ZipWith(
	zipf func(l, r Continuation) Sequential,
	cont Continuation,
) SeqVal {
	var (
		leftStep, left   = s()
		rightStep, right = cont.Continue()
	)
	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
		if leftStep.Type().Match(None) || rightStep.Type().Match(None) {
			return NewNone(), NewSequence()
		}
		if len(args) > 0 {
			return NewPair(leftStep, rightStep).Call(args...),
				left.ZipWith(zipf, right)
		}
		return NewPair(leftStep, rightStep),
			left.ZipWith(zipf, right)
	})
}

func (s SeqVal) Split() (Sequential, Sequential) {
	var (
		head, tail  = s.Continue()
		left, right = tail.(Zipped).Split()
	)
	if head.Type().Match(Pair) { // list of pairs gets zipped into keys & values
		if pair, ok := head.(Paired); ok {
			return SeqVal(func(args ...Expression) (Expression, SeqVal) {
					if len(args) > 0 {
						return pair.Left().Call(args...), left.(SeqVal)
					}
					return pair.Left(), left.(SeqVal)
				}),
				SeqVal(func(args ...Expression) (Expression, SeqVal) {
					if len(args) > 0 {
						return pair.Right().Call(args...), right.(SeqVal)
					}
					return pair.Right(), right.(SeqVal)
				})
		}
	}
	if !head.Type().Match(None) { // flat lists are split two elements at a step
		var resl, resr Sequential
		if !head.Type().Match(None) {
			resl = SeqVal(func(args ...Expression) (Expression, SeqVal) {
				if len(args) > 0 {
					return head.Call(args...), left.(SeqVal)
				}
				return head, left.(SeqVal)
			})
		} else {
			resl = NewSequence()
		}
		head, tail = tail.Continue()
		if !head.Type().Match(None) {
			resr = SeqVal(func(args ...Expression) (Expression, SeqVal) {
				if len(args) > 0 {
					return head.Call(args...), right.(SeqVal)
				}
				return head, right.(SeqVal)
			})
		} else {
			resr = NewSequence()
		}
		return resl, resr
	}
	// head is a none value
	return NewSequence(), NewSequence()
}

//// DATA STATE
///
// monad enclosing over stateful native data
func NewStatefulData(
	state d.Native,
	trans func(state d.Native, args ...d.Native) d.Native,
) StateD {
	return func(args ...d.Native) (d.Native, StateD) {
		if len(args) > 0 {
			return trans(state, args...), NewStatefulData(state, trans)
		}
		return state, NewStatefulData(state, trans)
	}
}
func (s StateD) Monad() Monadic { var _, m = s(); return m }
func (s StateD) Step(args ...Expression) (Expression, Monadic) {
	var r = s.Call(args...)
	s = s.Call().(StateD)
	return r, s
}
func (s StateD) Current() Expression {
	var c, _ = s()
	return Box(c)
}
func (s StateD) Sequence() Sequential {
	var (
		data  d.Native
		state StateD
		pair  Paired
	)
	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
		if len(args) > 0 {
			pair = s.Call(args...).(Paired)
			return pair.Left(), pair.Right().(StateD).Sequence().(SeqVal)
		}
		data, state = s()
		return Box(data), state.Sequence().(SeqVal)
	})
}
func (s StateD) String() string { return s.Monad().String() }
func (s StateD) Type() TyComp   { return Def(Monad, s.Current().Type()) }
func (s StateD) TypeFnc() TyFnc { return Data | State }
func (s StateD) Call(args ...Expression) Expression {
	var (
		n d.Native
		m Monadic
	)
	if len(args) > 0 {
		var nats = make([]d.Native, 0, len(args))
		for _, arg := range args {
			if arg.Type().Match(Data) {
				if dat, ok := arg.(NatEval); ok {
					nats = append(nats, dat.Eval())
				}
			}
		}
		n, m = s(nats...)
		return NewPair(Box(n), m)
	}
	n, m = s()
	return NewPair(Box(n), m)
}

//// EXPRESSION STATE
///
// monad enclosing over stateful expressions
func NewStatefulExpression(
	state Expression, trans func(state Expression, args ...Expression) Expression) StateE {
	return func(args ...Expression) (Expression, StateE) {
		if len(args) > 0 {
			return trans(state, args...), NewStatefulExpression(state, trans)
		}
		return state, NewStatefulExpression(state, trans)
	}
}
func (s StateE) Step(args ...Expression) (Expression, Monadic) { return s(args...) }
func (s StateE) Current() Expression                           { var c, _ = s(); return c }
func (s StateE) Monad() Monadic                                { var _, m = s(); return m }
func (s StateE) String() string                                { return s.Monad().String() }
func (s StateE) Type() TyComp                                  { return Def(Monad, s.Current().Type()) }
func (s StateE) TypeFnc() TyFnc                                { return State | s.Monad().TypeFnc() }
func (s StateE) Sequence() Sequential {
	var (
		expr  Expression
		state StateE
	)

	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
		if len(args) > 0 {
			expr, state = s(args...)
			return expr, state.Sequence().(SeqVal)
		}
		expr, state = s()
		return expr, state.Sequence().(SeqVal)
	})
}
func (s StateE) Call(args ...Expression) Expression {
	if len(args) > 0 {
		var e, m = s(args...)
		return NewPair(e, m)
	}
	var e, m = s()
	return NewPair(e, m)
}

///////////////////////////////////////////////////////////////////////////////
//// COMPOSITION PRIMITIVES
///
// define the curryed function
func Curry(f, g FuncDef) FuncDef {
	if f.TypeArguments().Match(g.TypeReturn()) {
		return Define(Lambda(
			func(args ...Expression) Expression {
				if len(args) > 0 {
					return f.Call(g.Call(args...))
				}
				return f.Call(g.Call())
			}),
			Def(
				f.TypeIdent(),
				g.TypeIdent()),
			f.TypeReturn(),
			f.TypeArguments(),
		)
	}
	return Define(NewNone(), None, None)
}
