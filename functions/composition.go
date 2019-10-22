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
	AppVal func(...Expression) (Expression, AppVal)

	//// MONAD
	MonVal func(...Expression) (Expression, MonVal)

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
func (g GenVal) TypeFnc() TyFnc { return Generator }
func (g GenVal) Type() TyComp   { return Def(Generator, g.Head().Type()) }
func (g GenVal) String() string { return g.Head().String() }

func (g GenVal) Empty() bool {
	if !g.Tail().Empty() || !g.Head().Type().Match(None) {
		return false
	}
	return true
}
func (g GenVal) Traverse() (Expression, Traversable) { return g() }
func (g GenVal) Head() Expression                    { return g.Expr() }
func (g GenVal) Tail() Traversable                   { return g.Generator() }

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
		g.Head().Type().TypeReturn(),
		g.Head().Type().TypeArguments(),
	)
}
func (g AccVal) String() string { return g.Head().String() }

func (a AccVal) Empty() bool {
	if !a.Tail().Empty() || !a.Head().Type().Match(None) {
		return false
	}
	return true
}
func (g AccVal) Head() Expression                    { return g.Result() }
func (g AccVal) Tail() Traversable                   { return g.Accumulator() }
func (g AccVal) Traverse() (Expression, Traversable) { return g() }

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
				vec = vec.ConsVec(e.Call(arg))
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
func NewSequence(seq Sequential) SeqVal {
	var (
		head Expression
		tail Sequential
	)
	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
		if len(args) > 0 {
			seq = seq.Cons(args...)
		}
		head, tail = seq.Consume()
		if head.Type().Match(None) {
			tail = NewSequence(seq)
		}
		return head, NewSequence(tail)
	})
}
func (s SeqVal) Head() Expression {
	var expr, _ = s()
	return expr
}
func (s SeqVal) Tail() Traversable {
	var _, seq = s()
	return seq
}
func (s SeqVal) TailSeq() SeqVal { return s.Tail().(SeqVal) }
func (s SeqVal) ConsumeSeq() (Expression, SeqVal) {
	return s.Head(), s.TailSeq()
}
func (s SeqVal) TypeElem() TyComp { return s.Head().Type() }
func (s SeqVal) TypeFnc() TyFnc   { return Sequence }
func (s SeqVal) Type() TyComp     { return Def(Sequence, s.TypeElem()) }
func (s SeqVal) Cons(elems ...Expression) Sequential {
	if len(elems) > 0 {
		if len(elems) == 1 {
			return SeqVal(func(args ...Expression) (Expression, SeqVal) {
				if len(args) > 0 {
					var head, tail = s.Cons(
						append(elems, args...)...,
					).(SeqVal)()
					return head, NewSequence(tail)
				}
				return elems[0], s
			})
		}
		return SeqVal(func(args ...Expression) (Expression, SeqVal) {
			if len(args) > 0 {
				var head, tail = s.Cons(
					append(elems, args...)...,
				).(SeqVal)()
				return head, NewSequence(tail)
			}
			return elems[0], NewSequence(s.Cons(elems[1:]...))
		})
	}
	return s
}
func (s SeqVal) Call(args ...Expression) Expression {
	var head, tail = s()
	return NewPair(head, tail)
}
func (s SeqVal) Empty() bool {
	if !s.Tail().Empty() || !s.Head().Type().Match(None) {
		return false
	}
	return true
}
func (s SeqVal) Traverse() (Expression, Traversable) { return s() }
func (s SeqVal) Consume() (Expression, Sequential)   { return s() }
func (s SeqVal) String() string {
	var (
		hstr, tstr string
		head, tail = s()
	)
	for !head.Type().Match(None) {
		hstr = hstr + "[ " + head.String() + " "
		tstr = tstr + "]"
		head, tail = tail()
	}
	return hstr + tstr
}

func (s SeqVal) Concat(right Sequential) Sequential {
	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
		if len(args) > 0 {
			s = s.Cons(args...).(SeqVal)
		}
		var head, tail = s.Consume()
		if tail.Empty() {
			tail = right
		}
		return head, NewSequence(tail)
	})
}

func (s SeqVal) Map(mapf Expression) Monoidal {
	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
		if len(args) > 0 {
			s = s.Cons(args...).(SeqVal)
		}
		var head, tail = s.Consume()
		if !head.Type().Match(None) {
			var result = mapf.Call(head)
			// skip function applications yielding none
			for result.Type().Match(None) {
				head, tail = tail.Consume()
				result = mapf.Call(head)
			}
			return result, NewSequence(tail).Map(mapf).(SeqVal)
		}
		return head, NewSequence(tail).Map(mapf).(SeqVal)
	})
}

func (s SeqVal) MapX(mapx Expression) Monoidal {
	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
		if len(args) > 0 {
			s = s.Cons(args...).(SeqVal)
		}
		var head, tail = s.Consume()
		if !head.Type().Match(None) {
			if head.Type().Match(Traversables) {
				head, tail = head.(Monoidal).Map(mapx).Concat(tail).Consume()
			}
			var result = mapx.Call(head)
			// skip function applications yielding none
			for result.Type().Match(None) {
				head, tail = tail.Consume()
				result = mapx.Call(head)
			}
			return result, NewSequence(tail).MapX(mapx).(SeqVal)
		}
		return head, NewSequence(tail).MapX(mapx).(SeqVal)
	})
}

func (s SeqVal) Fold(acc Expression, fold func(...Expression) Expression) Monoidal {
	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
		if len(args) > 0 {
			s = s.Cons(args...).(SeqVal)
		}
		var head, tail = s.Traverse()
		if !head.Type().Match(None) {
			var result = acc.Call(fold(head))
			for result.Type().Match(None) {
				head, tail = tail.Traverse()
				result = acc.Call(fold(head))
			}
			return result, tail.(SeqVal).Fold(result, fold).(SeqVal)
		}
		return acc, NewSequence(NewList())
	})
}

func (s SeqVal) Apply(apply func(...Expression) Expression) AppVal { return NewApplicative(s, apply) }

func Curry(f, g FuncDef) FuncDef {
	if f.TypeArguments().Match(g.TypeReturn()) {
		return Define(GenericFunc(
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

//// APPLICATIVE MONOID
func NewApplicative(s Sequential, apply func(...Expression) Expression) AppVal {
	return func(args ...Expression) (Expression, AppVal) {
		if len(args) > 0 {
			var head, tail = s.Cons(args...).Consume()
			return head, NewApplicative(tail, apply)
		}
		var head, tail = s.Consume()
		return head, NewApplicative(tail, apply)
	}
}
func (s AppVal) TypeElem() TyComp { return s.Head().Type() }
func (s AppVal) TypeFnc() TyFnc   { return Applicative }
func (s AppVal) Type() TyComp     { return Def(Applicative, s.TypeElem()) }
func (a AppVal) Empty() bool {
	if !a.Tail().Empty() || !a.Head().Type().Match(None) {
		return false
	}
	return true
}
func (s AppVal) Cons(elems ...Expression) Sequential {
	if len(elems) > 0 {
		if len(elems) == 1 {
			return AppVal(func(args ...Expression) (Expression, AppVal) {
				if len(args) > 0 {
					var head, tail = s(append(elems, args...)...)
					return head, tail
				}
				return elems[0], s
			})
		}
		return AppVal(func(args ...Expression) (Expression, AppVal) {
			if len(args) > 0 {
				var head, tail = s(append(elems, args...)...)
				return head, tail
			}
			return s(elems...)
		})
	}
	return s
}
func (s AppVal) Call(args ...Expression) Expression {
	var head Expression
	if len(args) > 0 {
		head, _ = s(args...)
		return head
	}
	head, _ = s()
	return head
}
func (s AppVal) Traverse() (Expression, Traversable) { return s() }
func (s AppVal) Consume() (Expression, Sequential)   { return s() }
func (s AppVal) Head() Expression {
	var expr, _ = s()
	return expr
}
func (s AppVal) Tail() Traversable {
	var _, seq = s()
	return seq
}
func (s AppVal) TailApp() AppVal { return s.Tail().(AppVal) }
func (s AppVal) ConsumeApp() (Expression, AppVal) {
	return s.Head(), s.TailApp()
}
func (s AppVal) String() string {
	var head, tail = s()
	return tail.Cons(head).String()
}
func (s AppVal) Concat(right Sequential) Sequential {
	return AppVal(func(args ...Expression) (Expression, AppVal) {
		if len(args) > 0 {
			s = s.Cons(args...).(AppVal)
		}
		var head, tail = s.Consume()
		if tail.Empty() {
			tail = right
		}
		return head, tail.(AppVal)
	})
}

func (s AppVal) Map(mapf Expression) Monoidal {
	return AppVal(func(args ...Expression) (Expression, AppVal) {
		if len(args) > 0 {
			s = s.Cons(args...).(AppVal)
		}
		var head, tail = s()
		if !head.Type().Match(None) {
		}
		return head, tail.Map(mapf).(AppVal)
	})
}

func (s AppVal) MapX(mapx Expression) Monoidal {
	return AppVal(func(args ...Expression) (Expression, AppVal) {
		if len(args) > 0 {
			s = s.Cons(args...).(AppVal)
		}
		var head, tail = s.Consume()
		if !head.Type().Match(None) {
			if head.Type().Match(Traversables) {
				head, tail = head.(Monoidal).Map(mapx).Concat(tail).Consume()
			}
			var result = mapx.Call(head)
			// skip function applications yielding none
			for result.Type().Match(None) {
				head, tail = tail.Consume()
				result = mapx.Call(head)
			}
			return result, tail.(Monoidal).MapX(mapx).(AppVal)
		}
		return head, tail.(Monoidal).MapX(mapx).(AppVal)
	})
}

func (s AppVal) Fold(acc Expression, fold func(...Expression) Expression) Monoidal {
	return AppVal(func(args ...Expression) (Expression, AppVal) {
		if len(args) > 0 {
			s = s.Cons(args...).(AppVal)
		}
		var head, tail = s.Traverse()
		if !head.Type().Match(None) {
			var result = acc.Call(fold(head))
			for result.Type().Match(None) {
				head, tail = tail.Traverse()
				result = acc.Call(fold(head))
			}
			return result, tail.(Monoidal).Fold(result, fold).(AppVal)
		}
		return acc, tail.(AppVal)
	})
}

//// MONAD TYPE
///
// sequence of computations
func NewMonad(expr Expression) MonVal {
	return MonVal(func(args ...Expression) (Expression, MonVal) {
		if len(args) > 0 {
			expr = expr.Call(args...)
			return expr, NewMonad(expr)
		}
		expr = expr.Call()
		return expr, NewMonad(expr)
	})
}

func (s MonVal) Type() TyComp     { return Def(Monad, s.Current().Type()) }
func (s MonVal) TypeElem() TyComp { return s.Current().Type() }
func (s MonVal) TypeFnc() TyFnc   { return s.Current().TypeFnc() }
func (s MonVal) String() string   { return s.Current().String() }

func (s MonVal) Step(args ...Expression) (Expression, Monadic) { return s(args...) }
func (s MonVal) Call(args ...Expression) Expression {
	var (
		expr Expression
		mon  MonVal
	)
	if len(args) > 0 {
		expr, mon = s(args...)
		return NewPair(expr, mon)
	}
	expr, mon = s()
	return NewPair(expr, mon)
}
func (s MonVal) Current() Expression {
	var cur, _ = s()
	return cur
}
func (s MonVal) Monad() Monadic {
	var _, mon = s()
	return mon
}
func (s MonVal) Sequence() Sequential {
	var (
		expr Expression
		mon  Monadic
	)
	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
		if len(args) > 0 {
			expr, mon = s(args...)
			return expr, mon.Sequence().(SeqVal)
		}
		expr, mon = s()
		return expr, mon.Sequence().(SeqVal)
	})
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

func Filter(s Monoidal, test Testable) Sequential {
	var (
		acc    = NewVector()
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
	return s.Fold(acc, filter)
}

func Pass(s Monoidal, test Testable) Sequential {
	var (
		acc    = NewVector()
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
	return s.Fold(acc, filter)
}
