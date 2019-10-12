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

	//// MONAD
	MonVal func(...Expression) (Expression, MonVal)

	//// STATE MONADS
	DataState func(args ...d.Native) (d.Native, DataState)
	ExprState func(args ...Expression) (Expression, ExprState)
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
func (g GenVal) Type() TyPat    { return Def(Generator, g.Head().Type()) }
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
func (g AccVal) Type() TyPat {
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
func (e EnumDef) Type() TyPat {
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
func (e EnumVal) Type() TyPat {
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
		tail Consumeable
	)
	return func(args ...Expression) (Expression, SeqVal) {
		if len(args) > 0 {
			head, tail = seq.Cons(args...).Consume()
		}
		head, tail = seq.Consume()
		if head.Type().Match(None) {
			tail = NewSequence(seq)
		}
		return head, tail.(SeqVal)
	}
}
func (s SeqVal) TypeFnc() TyFnc  { return s.Tail().TypeFnc() }
func (s SeqVal) Type() TyPat     { return s.Tail().Type() }
func (s SeqVal) TypeElem() TyPat { return s.Head().Type() }
func (s SeqVal) Cons(elems ...Expression) Sequential {
	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
		if len(args) > 0 {
			_, s = s(args...)
		}
		if len(elems) > 0 {
			return s(elems...)
		}
		return s()
	})
}
func (s SeqVal) Prepend(elems ...Expression) Sequential {
	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
		if len(args) > 0 {
			_, s = s(args...)
		}
		if len(elems) > 0 {
			return s(elems...)
		}
		return s()
	})
}
func (s SeqVal) Append(elems ...Expression) Sequential {
	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
		if len(args) > 0 {
			_, s = s(args...)
		}
		if len(elems) > 0 {
			return s(elems...)
		}
		return s()
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
func (s SeqVal) String() string {
	var head, tail = s()
	return tail.Cons(head).String()
}

// apply takes another sequence of elements as arguments to apply to its
// collection of expressions elements.
func (s SeqVal) Apply(
	apply func(Expression) Expression) SeqVal {
	var (
		head Expression
		tail SeqVal
	)
	return func(args ...Expression) (Expression, SeqVal) {
		head, tail = s()
		if len(args) > 0 {
			head, tail = tail(args...)
		}
		return apply(head), tail.Apply(apply)
	}
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

func (s MonVal) Type() TyPat     { return Def(Monad, s.Current().Type()) }
func (s MonVal) TypeElem() TyPat { return s.Current().Type() }
func (s MonVal) TypeFnc() TyFnc  { return s.Current().TypeFnc() }
func (s MonVal) String() string  { return s.Current().String() }

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
) DataState {
	return func(args ...d.Native) (d.Native, DataState) {
		if len(args) > 0 {
			return trans(state, args...), NewStatefulData(state, trans)
		}
		return state, NewStatefulData(state, trans)
	}
}
func (s DataState) Monad() Monadic { var _, m = s(); return m }
func (s DataState) Step(args ...Expression) (Expression, Monadic) {
	var r = s.Call(args...)
	s = s.Call().(DataState)
	return r, s
}
func (s DataState) Current() Expression {
	var c, _ = s()
	return Box(c)
}
func (s DataState) Sequence() Sequential {
	var (
		data  d.Native
		state DataState
		pair  Paired
	)
	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
		if len(args) > 0 {
			pair = s.Call(args...).(Paired)
			return pair.Left(), pair.Right().(DataState).Sequence().(SeqVal)
		}
		data, state = s()
		return Box(data), state.Sequence().(SeqVal)
	})
}
func (s DataState) String() string { return s.Monad().String() }
func (s DataState) Type() TyPat    { return Def(Monad, s.Current().Type()) }
func (s DataState) TypeFnc() TyFnc { return Data | State }
func (s DataState) Call(args ...Expression) Expression {
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
	state Expression, trans func(state Expression, args ...Expression) Expression) ExprState {
	return func(args ...Expression) (Expression, ExprState) {
		if len(args) > 0 {
			return trans(state, args...), NewStatefulExpression(state, trans)
		}
		return state, NewStatefulExpression(state, trans)
	}
}
func (s ExprState) Step(args ...Expression) (Expression, Monadic) {
	return s(args...)
}
func (s ExprState) Current() Expression { var c, _ = s(); return c }
func (s ExprState) Monad() Monadic      { var _, m = s(); return m }
func (s ExprState) String() string      { return s.Monad().String() }
func (s ExprState) Type() TyPat         { return Def(Monad, s.Current().Type()) }
func (s ExprState) TypeFnc() TyFnc      { return State | s.Monad().TypeFnc() }
func (s ExprState) Sequence() Sequential {
	var (
		expr  Expression
		state ExprState
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
func (s ExprState) Call(args ...Expression) Expression {
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
func MapF(s Sequential, mapf func(Expression) Expression) Sequential {
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
		return head, MapF(seq, mapf).(SeqVal)
	})
}

func FoldL(s Sequential, mapf, acc Expression) Sequential {
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
			return result, FoldL(tail.(SeqVal), mapf, result).(SeqVal)
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
	return FoldL(s, acc, filter)
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
	return FoldL(s, acc, filter)
}

//func TakeN(s Sequential, num int) Sequential {
//	var ()
//	return FoldL(s, acc, taker)
//}
