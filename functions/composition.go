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
)

///////////////////////////////////////////////////////////////////////////////
//// COMPOSITION PRIMITIVES
///
// define the curryed function
func Curry(f, g FuncDef) FuncDef {
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
func (g GenVal) Type() TyComp     { return Def(Generator, g.Current().Type()) }
func (g GenVal) TypeElem() TyComp { return g.Current().Type() }
func (g GenVal) String() string   { return g.Current().String() }
func (g GenVal) End() bool {
	if g.Current().Type().Match(None) {
		return true
	}
	return false
}
func (g GenVal) Continue() (Expression, Continuation) { return g() }
func (g GenVal) Current() Expression                  { return g.Expr() }
func (g GenVal) Next() Continuation                   { return g.Generator() }

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
		g.Current().Type().TypeRet(),
		g.Current().Type().TypeArgs(),
	)
}
func (g AccVal) String() string { return g.Current().String() }

func (a AccVal) End() bool {
	if a.Current().Type().Match(None) {
		return true
	}
	return false
}
func (g AccVal) Current() Expression                  { return g.Result() }
func (g AccVal) TypeElem() TyComp                     { return g.Current().Type() }
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
	return Def(Enum, e.Expr().Type().TypeRet())
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
		return NewPair(s.Current().Call(args...), s.Next())
	}
	return NewPair(s.Current(), s.Next())
}
func (s SeqVal) Continue() (Expression, Continuation) { return s() }
func (s SeqVal) Current() Expression {
	var expr, _ = s()
	return expr
}
func (s SeqVal) NextSeq() SeqVal { return s.Next().(SeqVal) }
func (s SeqVal) Next() Continuation {
	var _, seq = s()
	return seq
}
func (s SeqVal) TypeElem() TyComp { return s.Current().Type() }
func (s SeqVal) TypeFnc() TyFnc   { return Sequence }
func (s SeqVal) Type() TyComp     { return Def(Sequence, s.TypeElem()) }
func (s SeqVal) End() bool {
	if s.Current().Type().Match(None) {
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

func (s SeqVal) Map(mapf Expression) Sequential {
	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
		var (
			head, tail = s()
			lst        Expression
		)
		if len(args) > 0 {
			lst = args[len(args)-1]
			// cross product, if last argument is a functor
			if lst.Type().Match(Functors) {
				if arg, ok := lst.(Functorial); ok {
					if len(args) > 1 {
						return s.MapX(head.Call(args...),
							mapf, arg), tail.Map(mapf).(SeqVal)
					}
					return s.MapX(head, mapf, arg), tail.Map(mapf).(SeqVal)
				}
			}
			// dot product, since last argument is not a functor
			return mapf.Call(head.Call(args...)), tail.Map(mapf).(SeqVal)
		}
		// no arguments given
		return mapf.Call(head), tail.Map(mapf).(SeqVal)
	})
}

func (s SeqVal) MapX(head, mapf Expression, arg Continuation) Sequential {
	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
		// check if current head of parent list is none
		// yield step & next continuation from argument
		var step, next = arg.Continue()
		if len(args) > 0 { // if args have been passed
			// call mapf with current parent lists head &
			// arguments passed during call to get step.
			// s-map tail of sequential argument
			return mapf.Call(head, step.Call(args...)),
				s.MapX(head, mapf, next).(SeqVal)
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
				return seq.Current(), NewSeqCont(seq.Next())
			}
		}
		return head, tail
	})
}

func (s SeqVal) Fold(
	acc Expression,
	fold func(acc, head Expression) Expression,
) SeqVal {
	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
		var (
			result     Expression
			head, tail = s()
		)
		if head.TypeFnc().Match(None) {
			return acc, tail
		}
		result = fold(acc, head)
		if len(args) > 0 {
			return result.Call(args...), tail.Fold(result, fold)
		}
		return result, tail.Fold(result, fold)
	})
}

func (s SeqVal) Filter(test Testable) Sequential {
	var (
		seq        = NewSequence()
		head, tail = s()
	)
	if head.TypeFnc().Match(None) {
		return NewSequence()
	}
	if !test.Test(head) {
		return seq.Concat(head).(SeqVal).ConcatSeq(tail.Filter(test))
	}
	return seq.ConcatSeq(tail.Filter(test))
}

func (s SeqVal) Pass(test Testable) Sequential {
	var (
		seq        = NewSequence()
		head, tail = s()
	)
	if head.TypeFnc().Match(None) {
		return NewSequence()
	}
	if test.Test(head) {
		return seq.Concat(head).Concat(tail.Filter(test))
	}
	return seq.Concat(tail.Filter(test))
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
func (s SeqVal) Bind(bind Expression, cont Continuation) Sequential {
	var step, next = s()
	return SeqVal(func(args ...Expression) (Expression, SeqVal) {
		if len(args) > 0 {
			return step.Call(cont.Current().Call(args...)),
				next.Bind(bind, cont.Next()).(SeqVal)
		}
		return step.Call(cont.Current()),
			next.Bind(bind, cont.Next()).(SeqVal)
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
