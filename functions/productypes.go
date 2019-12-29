/*

SUM TYPES
---------

sumtypes are parametric types, that generate a subtype for every type of
argument. all elements of a sum type are of the same type.

examples for sumtypes are would be all collection types, enumerables, the set
of integers‥.

*/
package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

type (
	//// GENERIC EXPRESSIONS
	NoneVal func()
	Const   func() Expression
	Lambda  func(...Expression) Expression

	//// GENERATOR | ACCUMULATOR
	GenVal func() (Expression, GenVal)
	AccVal func(...Expression) (Expression, AccVal)

	//// ENUMERABLE
	///
	// enumerable value returns its index position within set of
	// enumerables and its data constructor to enable previous, next
	// methods
	EnumCon func(d.Numeral) EnumVal
	EnumVal func(...Expression) (Expression, d.Numeral, EnumCon)
)

//// NONE VALUE CONSTRUCTOR
///
// none represens the abscence of a value of any type.  implements countable,
// sliceable, consumeable, testable, compareable, key-, index- and generic pair
// interfaces to be able to stand in as return value for such expressions.
func NewNone() NoneVal { return func() {} }

func (n NoneVal) Head() Expression                     { return n }
func (n NoneVal) Tail() Continuation                   { return n }
func (n NoneVal) Cons(...Expression) Sequential        { return n }
func (n NoneVal) ConsContinue(Continuation) Sequential { return n }
func (n NoneVal) Concat(...Expression) Sequential      { return n }
func (n NoneVal) Prepend(...Expression) Sequential     { return n }
func (n NoneVal) Append(...Expression) Sequential      { return n }
func (n NoneVal) Len() int                             { return 0 }
func (n NoneVal) Compare(...Expression) int            { return -1 }
func (n NoneVal) String() string                       { return "⊥" }
func (n NoneVal) Call(...Expression) Expression        { return nil }
func (n NoneVal) Key() Expression                      { return nil }
func (n NoneVal) Index() Expression                    { return nil }
func (n NoneVal) Left() Expression                     { return nil }
func (n NoneVal) Right() Expression                    { return nil }
func (n NoneVal) Both() Expression                     { return nil }
func (n NoneVal) Value() Expression                    { return nil }
func (n NoneVal) Empty() bool                          { return true }
func (n NoneVal) Test(...Expression) bool              { return false }
func (n NoneVal) TypeFnc() TyFnc                       { return None }
func (n NoneVal) TypeNat() d.TyNat                     { return d.Nil }
func (n NoneVal) Type() TyComp                         { return Def(None) }
func (n NoneVal) TypeElem() TyComp                     { return Def(None) }
func (n NoneVal) TypeName() string                     { return n.String() }
func (n NoneVal) Slice() []Expression                  { return []Expression{} }
func (n NoneVal) Flag() d.BitFlag                      { return d.BitFlag(None) }
func (n NoneVal) FlagType() d.Uint8Val                 { return Kind_Fnc.U() }
func (n NoneVal) Continue() (Expression, Continuation) { return NewNone(), NewNone() }
func (n NoneVal) Consume() (Expression, Sequential)    { return NewNone(), NewNone() }

//// GENERIC CONSTANT DEFINITION
///
// declares a constant value
func NewConstant(constant func() Expression) Const { return constant }

func (c Const) Type() TyComp                  { return Def(Constant, c().Type(), None) }
func (c Const) TypeIdent() TyComp             { return c().Type().TypeId() }
func (c Const) TypeReturn() TyComp            { return c().Type().TypeRet() }
func (c Const) TypeArguments() TyComp         { return Def(None) }
func (c Const) TypeFnc() TyFnc                { return Constant }
func (c Const) String() string                { return c().String() }
func (c Const) Call(...Expression) Expression { return c() }

//// GENERIC FUNCTION DEFINITION
///
// declares a constant value
func NewLambda(fnc func(...Expression) Expression) Lambda {
	return func(args ...Expression) Expression {
		if len(args) > 0 {
			return fnc(args...)
		}
		return fnc()
	}
}

func (c Lambda) Call(args ...Expression) Expression {
	if len(args) > 0 {
		return c(args...)
	}
	return c()
}
func (c Lambda) String() string        { return c().String() }
func (c Lambda) TypeFnc() TyFnc        { return c().TypeFnc() }
func (c Lambda) Type() TyComp          { return c().Type() }
func (c Lambda) TypeIdent() TyComp     { return c().Type().TypeId() }
func (c Lambda) TypeReturn() TyComp    { return c().Type().TypeRet() }
func (c Lambda) TypeArguments() TyComp { return c().Type().TypeArgs() }

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

//// ENUM TYPE
///
// declares an enumerable type returning instances from the set of enumerables
// defined by the passed function
func NewEnumType(fnc func(d.Numeral) Expression) EnumCon {
	return func(idx d.Numeral) EnumVal {
		return func(args ...Expression) (Expression, d.Numeral, EnumCon) {
			if len(args) > 0 {
				return fnc(idx).Call(args...), idx, NewEnumType(fnc)
			}
			return fnc(idx), idx, NewEnumType(fnc)
		}
	}
}
func (e EnumCon) Expr() Expression            { return e(d.IntVal(0)) }
func (e EnumCon) Alloc(idx d.Numeral) EnumVal { return e(idx) }
func (e EnumCon) Type() TyComp {
	return Def(Enum, e.Expr().Type().TypeRet())
}
func (e EnumCon) TypeFnc() TyFnc { return Enum }
func (e EnumCon) String() string { return e.Type().TypeName() }
func (e EnumCon) Call(args ...Expression) Expression {
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
func (e EnumVal) EnumType() EnumCon {
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
