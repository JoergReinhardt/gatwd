package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

type (
	//// GENERATOR | ACCUMULATOR
	GenVal func(...Expression) Expression
	AccVal func(...Expression) Expression

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
func NewGenerator(gen Expression) GenVal {
	return func(args ...Expression) Expression {
		if len(args) > 0 {
			NewGenerator(gen.Call(args...))
		}
		return gen.Call()
	}
}
func (g GenVal) Call(...Expression) Expression {
	return NewPair(g(), NewGenerator(g()))
}
func (g GenVal) TypeFnc() TyFnc { return Generator }
func (g GenVal) Type() TyComp   { return Def(Generator, g.Head().Type()) }
func (g GenVal) String() string { return g.Head().String() }

func (g GenVal) Consume() (Expression, Consumeable) { return g(), NewGenerator(g()) }
func (g GenVal) Head() Expression                   { return g() }
func (g GenVal) Tail() Consumeable {
	var _, tail = g.Consume()
	return tail
}

//// ACCUMULATOR
///
// accumulator expects an expression as input, that returns itself unboxed,
// when called empty and returns a new accumulator accumulating its value and
// arguments to create a new accumulator, if arguments where passed.
func NewAccumulator(acc Expression) AccVal {
	return AccVal(func(args ...Expression) Expression {
		if len(args) > 0 {
			return NewAccumulator(acc.Call(args...))
		}
		return acc
	})
}

func (g AccVal) Call(args ...Expression) Expression {
	if len(args) > 0 {
		return g(args...)
	}
	return g()
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

func (g AccVal) Head() Expression                   { return g() }
func (g AccVal) Tail() Consumeable                  { return NewAccumulator(g) }
func (g AccVal) Consume() (Expression, Consumeable) { return g(), NewAccumulator(g) }

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
				vec = vec.AppendVec(e.Call(arg))
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
func NewSequence(seq Consumeable) SeqVal {
	var (
		head Expression
		tail Consumeable
	)
	return func(args ...Expression) (Expression, SeqVal) {
		if len(args) > 0 {
			seq = seq.Call(args...).(Consumeable)
		}
		head, tail = seq.Consume()
		if head.Type().Match(None) {
			tail = NewSequence(seq)
		}
		return head, NewSequence(tail)
	}
}
func (s SeqVal) TypeFnc() TyFnc   { return s.Tail().TypeFnc() }
func (s SeqVal) Type() TyComp     { return s.Tail().Type() }
func (s SeqVal) TypeElem() TyComp { return s.Head().Type() }
func (s SeqVal) Cons(elems ...Expression) Sequential {
	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
		var (
			head Expression
			tail Consumeable
		)
		if len(args) > 0 {
			_, s = s(args...)
		}
		if len(elems) > 0 {
			head, tail = s(elems...)
			return head, NewSequence(tail)
		}
		head, tail = s()
		return head, NewSequence(tail)
	})
}
func (s SeqVal) Call(args ...Expression) Expression {
	var head Expression
	if len(args) > 0 {
		head, _ = s(args...)
		return head
	}
	head, _ = s()
	return head
}
func (s SeqVal) Consume() (Expression, Consumeable) { return s() }
func (s SeqVal) Head() Expression {
	var expr, _ = s()
	return expr
}
func (s SeqVal) Tail() Consumeable {
	var _, seq = s()
	return seq
}
func (s SeqVal) TailSeq() SeqVal { return s.Tail().(SeqVal) }
func (s SeqVal) ConsumeSeq() (Expression, SeqVal) {
	return s.Head(), s.TailSeq()
}
func (s SeqVal) String() string {
	var head, tail = s()
	return tail.Cons(head).String()
}
func (s SeqVal) Apply(apply func(...Expression) Expression) AppVal {
	return func(args ...Expression) (Expression, AppVal) {
		if len(args) > 0 {
			return apply(append([]Expression{
					s.Head()},
					args...)...),
				s.TailSeq().Apply(apply)
		}
		return apply(s.Head()), s.TailSeq().Apply(apply)
	}
}
func (s AppVal) TypeFnc() TyFnc   { return s.Tail().TypeFnc() }
func (s AppVal) Type() TyComp     { return s.Tail().Type() }
func (s AppVal) TypeElem() TyComp { return s.Head().Type() }
func (s AppVal) Cons(elems ...Expression) Sequential {
	return AppVal(func(args ...Expression) (Expression, AppVal) {
		var (
			head Expression
			tail AppVal
		)
		if len(args) > 0 {
			_, s = s(args...)
		}
		if len(elems) > 0 {
			head, tail = s(elems...)
			return head, tail
		}
		head, tail = s()
		return head, AppVal(tail)
	})
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
func (s AppVal) Consume() (Expression, Consumeable) { return s() }
func (s AppVal) Head() Expression {
	var expr, _ = s()
	return expr
}
func (s AppVal) Tail() Consumeable {
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
//// COMPOSITION
///
// function composition primitives
func Map(s Sequential, mapf func(Expression) Expression) Sequential {
	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
		if len(args) > 0 {
			s = s.Cons(args...)
		}
		var head, tail = s.Consume()
		if !head.Type().Match(None) {
			var result = mapf(head)
			// skip function applications yielding none
			for result.Type().Match(None) {
				head, tail = s.Consume()
				result = mapf(head)
			}
			head = result
		}
		var seq = tail.(SeqVal)
		return head, Map(seq, mapf).(SeqVal)
	})
}

func Fold(s Sequential, mapf, acc Expression) Sequential {
	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
		if len(args) > 0 {
			s = s.Cons(args...)
		}
		var head, tail = s.Consume()
		if !head.Type().Match(None) {
			var result = acc.Call(mapf.Call(head))
			for result.Type().Match(None) {
				head, tail = s.Consume()
				result = acc.Call(mapf.Call(head))
			}
			return result, Fold(tail.(SeqVal), mapf, result).(SeqVal)
		}
		return acc, NewSequence(NewList())
	})
}

func Filter(s Sequential, test Testable) Sequential {
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
	return Fold(s, acc, filter)
}

func Pass(s Sequential, test Testable) Sequential {
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
	return Fold(s, acc, filter)
}
