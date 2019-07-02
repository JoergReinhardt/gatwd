/*
  FUNCTIONAL CONTAINERS

  containers implement enumeration of functional types, aka lists, vectors
  sets, pairs, tuples‥.
*/
package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

type (
	//// EXPRESSION CONSTRUCTORS
	ConstantExpr func() Expression
	VariadicExpr func(...Expression) Expression

	//// NARY EXPRESSION TYPE CONSTRUCTOR
	NaryExpr func(...Expression) Expression

	//// NATIVE TYPE & VALUE CONSTRUCTORS
	Native     func(...d.Native) d.Native
	NativePair func(...d.Native) d.Paired
	NativeSet  func(...d.Native) d.Mapped
	NativeCol  func(...d.Native) d.Sliceable
)

/// CONSTANT VALUE CONSTRUCTOR
//
func NewConstant(constant func() Expression) ConstantExpr { return constant }

func (c ConstantExpr) Ident() Expression                  { return c }
func (c ConstantExpr) Call(args ...Expression) Expression { return c() }
func (c ConstantExpr) Arity() Arity                       { return Arity(0) }
func (c ConstantExpr) TypeFnc() TyFnc                     { return Constant }
func (c ConstantExpr) TypeNat() d.TyNat                   { return c().TypeNat() }
func (c ConstantExpr) String() string                     { return c().String() }
func (c ConstantExpr) Eval(args ...d.Native) d.Native     { return c().Eval(args...) }
func (c ConstantExpr) FlagType() d.Uint8Val               { return Flag_Functional.U() }
func (c ConstantExpr) TypeName() string {
	return "λ → " + c().TypeName()
}
func (c ConstantExpr) Type() Typed {
	return Define(c().TypeName(), c().TypeFnc())
}

func NewExpression(expr func(args ...Expression) Expression) VariadicExpr { return expr }

func (c VariadicExpr) Ident() Expression                  { return c }
func (c VariadicExpr) TypeFnc() TyFnc                     { return Function }
func (c VariadicExpr) TypeNat() d.TyNat                   { return d.Function }
func (c VariadicExpr) String() string                     { return c.TypeName() }
func (c VariadicExpr) Call(args ...Expression) Expression { return c().Call(args...) }
func (c VariadicExpr) Eval(args ...d.Native) d.Native     { return c().Eval(args...) }
func (c VariadicExpr) FlagType() d.Uint8Val               { return Flag_Functional.U() }
func (c VariadicExpr) TypeName() string {
	return "λ.[T] → " + c().TypeName()
}
func (c VariadicExpr) Type() Typed {
	return Define(c().TypeName(), c().TypeFnc())
}

//// NARY EXPRESSION TYPE CONSTRUCTOR
func NewNary(expr Expression, signat ...Expression) NaryExpr {
	// take number of signature expressions as arity
	var arity = len(signat)
	// if no signature is passed, return expression as constant
	if arity == 0 {
		return func(...Expression) Expression {
			return NewConstant(
				func() Expression { return expr.Call() })
		}
	}
	return func(args ...Expression) Expression {
		var arglen = len(args)
		if arglen > 0 {
			// argument number satifies expression arity exactly
			if arglen == arity {
				return expr.Call(args...)
			}
			// argument number undersatisfies expression arity
			if arglen < arity {
				return NewNary(VariadicExpr(
					func(later ...Expression) Expression {
						return expr.Call(append(args, later...)...)
					}), signat[:arglen]...)
			}
			// argument number oversatisfies expressions arity
			if arglen > arity {
				var remain []Expression
				args, remain = args[:arity], args[arity:]
				var vec = NewVector(NewNary(expr,
					signat...).Call(args...))
				for len(remain) >= arity {
					args, remain = remain[:arity], remain[arity:]
					vec = vec.Append(NewNary(expr,
						signat...).Call(args...))
				}
				if len(remain) == 0 {
					return vec
				}
				return vec.Append(NewNary(expr,
					signat...).Call(remain...))
			}
		}
		return NewPair(expr, NewVector(signat...))
	}
}

// returns the value returned when calling itself directly, passing arguments
func (n NaryExpr) Ident() Expression    { return n }
func (n NaryExpr) Expr() Expression     { return n().(Paired).Left() }
func (n NaryExpr) Signature() VecCol    { return n().(Paired).Right().(VecCol) }
func (n NaryExpr) Arity() Arity         { return Arity(n.Signature().Len()) }
func (n NaryExpr) FlagType() d.Uint8Val { return Flag_Functional.U() }
func (n NaryExpr) TypeFnc() TyFnc       { return n.Expr().TypeFnc() }
func (n NaryExpr) TypeNat() d.TyNat     { return n.Expr().TypeNat() }
func (n NaryExpr) String() string       { return n.TypeName() }
func (n NaryExpr) Eval(args ...d.Native) d.Native {
	if len(args) > 0 {
		return n.Expr().Eval(args...)
	}
	return n.Expr().Eval()
}
func (n NaryExpr) Call(args ...Expression) Expression {
	if len(args) > 0 {
		return n.Expr().Call(args...)
	}
	return n.Expr().Call()
}
func (n NaryExpr) Type() Typed {
	return Define(
		n.Expr().TypeName(),
		Define("", NewPair(Signature, NewPair(
			Define("", NewPair(
				Argument,
				n.Signature(),
			)),
			Define("", NewPair(
				Return,
				n.Expr(),
			)),
		))),
	)
}
func (n NaryExpr) TypeName() string {
	var name string
	return name
}

//// NATIVE EXPRESSION CONSTRUCTOR
///
// returns an expression with native return type implementing the callable
// interface
func New(inf ...interface{}) Expression {
	return NewNative(d.New(inf...))
}

func NewNative(args ...d.Native) Expression {
	// if any initial arguments have been passed

	var nat = d.NewFromData(args...)
	var tnat = nat.TypeNat()

	switch {
	case tnat.Match(d.Slice):
		if slice, ok := nat.(d.Sliceable); ok {
			return NativeCol(func(nats ...d.Native) d.Sliceable {
				if len(nats) > 0 {
					return slice.Eval(nats...).(d.Sliceable)
				}
				return slice
			})
		}
	case tnat.Match(d.Unboxed):
		if slice, ok := nat.(d.Sliceable); ok {
			return NativeCol(func(nats ...d.Native) d.Sliceable {
				if len(nats) > 0 {
					return slice.Eval(nats...).(d.Sliceable)
				}
				return slice
			})
		}
	case tnat.Match(d.Pair):
		if pair, ok := nat.(d.Paired); ok {
			return NativePair(func(nats ...d.Native) d.Paired {
				if len(nats) > 0 {
					return pair.Eval(nats...).(d.Paired)
				}
				return pair
			})
		}
	case tnat.Match(d.Map):
		if set, ok := nat.(d.Mapped); ok {
			return NativeSet(func(nats ...d.Native) d.Mapped {
				if len(nats) > 0 {
					return set.Eval(nats...).(d.Mapped)
				}
				return set
			})
		}
	default:
		return Native(func(nats ...d.Native) d.Native {
			if len(nats) > 0 {
				return nat.Eval(nats...)
			}
			return nat
		})
	}
	return Native(func(...d.Native) d.Native { return d.NewNil() })
}

// ATOMIC NATIVE VALUE CONSTRUCTOR
func (n Native) Call(...Expression) Expression  { return n }
func (n Native) TypeFnc() TyFnc                 { return Data }
func (n Native) Eval(args ...d.Native) d.Native { return n(args...) }
func (n Native) TypeNat() d.TyNat               { return n().TypeNat() }
func (n Native) FlagType() d.Uint8Val           { return Flag_Functional.U() }
func (n Native) String() string                 { return n().String() }
func (n Native) TypeName() string               { return n().TypeName() }
func (n Native) Type() Typed {
	return TyDef(func() (string, Expression) {
		return n().TypeName(), New(n().TypeNat())
	})
}

// NATIVE SLICE VALUE CONSTRUCTOR
func (n NativeCol) Call(...Expression) Expression  { return n }
func (n NativeCol) TypeFnc() TyFnc                 { return Data }
func (n NativeCol) Eval(args ...d.Native) d.Native { return n(args...) }
func (n NativeCol) Len() int                       { return n().Len() }
func (n NativeCol) SliceNat() []d.Native           { return n().Slice() }
func (n NativeCol) Get(key d.Native) d.Native      { return n().Get(key) }
func (n NativeCol) GetInt(idx int) d.Native        { return n().GetInt(idx) }
func (n NativeCol) Range(s, e int) d.Native        { return n().Range(s, e) }
func (n NativeCol) Copy() d.Native                 { return n().Copy() }
func (n NativeCol) TypeNat() d.TyNat               { return n().TypeNat() }
func (n NativeCol) String() string                 { return n().String() }
func (n NativeCol) TypeName() string               { return n().TypeName() }
func (n NativeCol) Vector() VecCol                 { return NewVector(n.Slice()...) }
func (n NativeCol) FlagType() d.Uint8Val           { return Flag_Functional.U() }
func (n NativeCol) Type() Typed {
	return TyDef(func() (string, Expression) {
		return n().TypeName(), New(n().TypeNat())
	})
}
func (n NativeCol) Slice() []Expression {
	var slice = []Expression{}
	for _, val := range n.SliceNat() {
		slice = append(slice, NewNative(val))
	}
	return slice
}

// NATIVE PAIR VALUE CONSTRUCTOR
func (n NativePair) Call(...Expression) Expression  { return n }
func (n NativePair) TypeFnc() TyFnc                 { return Data }
func (n NativePair) TypeNat() d.TyNat               { return n().TypeNat() }
func (n NativePair) Eval(args ...d.Native) d.Native { return n(args...) }
func (n NativePair) LeftNat() d.Native              { return n().Left() }
func (n NativePair) RightNat() d.Native             { return n().Right() }
func (n NativePair) BothNat() (l, r d.Native)       { return n().Both() }
func (n NativePair) Left() Expression               { return NewNative(n().Left()) }
func (n NativePair) Right() Expression              { return NewNative(n().Right()) }
func (n NativePair) KeyType() d.TyNat               { return n().LeftType() }
func (n NativePair) ValType() d.TyNat               { return n().RightType() }
func (n NativePair) SubType() d.Typed               { return n().TypeNat() }
func (n NativePair) TypeName() string               { return n().TypeName() }
func (n NativePair) FlagType() d.Uint8Val           { return Flag_Functional.U() }
func (n NativePair) String() string                 { return n().String() }
func (n NativePair) Type() Typed {
	return TyDef(func() (string, Expression) {
		return n().TypeName(), New(n().TypeNat())
	})
}
func (n NativePair) Pair() Paired {
	return NewPair(
		NewNative(n().Left()),
		NewNative(n().Right()))
}
func (n NativePair) Both() (l, r Expression) {
	return NewNative(n().Left()),
		NewNative(n().Right())
}

func (n NativeSet) Call(args ...Expression) Expression {
	for _, arg := range args {
		n().Eval(arg.Eval())
	}
	return n
}

// NATIVE SET VALUE CONSTRUCTOR
func (n NativeSet) Ident() Expression                    { return n }
func (n NativeSet) TypeFnc() TyFnc                       { return Data }
func (n NativeSet) TypeNat() d.TyNat                     { return n().TypeNat() }
func (n NativeSet) Eval(args ...d.Native) d.Native       { return n(args...) }
func (n NativeSet) GetNat(acc d.Native) (d.Native, bool) { return n().Get(acc) }
func (n NativeSet) SetNat(acc, val d.Native) d.Mapped    { return n().Set(acc, val) }
func (n NativeSet) Delete(acc d.Native) bool             { return n().Delete(acc) }
func (n NativeSet) KeysNat() []d.Native                  { return n().Keys() }
func (n NativeSet) DataNat() []d.Native                  { return n().Data() }
func (n NativeSet) Fields() []d.Paired                   { return n().Fields() }
func (n NativeSet) KeyTypeNat() d.TyNat                  { return n().KeyType() }
func (n NativeSet) ValTypeNat() d.TyNat                  { return n().ValType() }
func (n NativeSet) SubType() d.Typed                     { return n().TypeNat() }
func (n NativeSet) TypeName() string                     { return n().TypeName() }
func (n NativeSet) FlagType() d.Uint8Val                 { return Flag_Functional.U() }
func (n NativeSet) String() string                       { return n().String() }
func (n NativeSet) Set() SetCol                          { return NewSet(n.Pairs()...) }
func (n NativeSet) Type() Typed {
	return TyDef(func() (string, Expression) {
		return n().TypeName(), New(n().TypeNat())
	})
}
func (n NativeSet) Pairs() []Paired {
	var pairs = []Paired{}
	for _, field := range n.Fields() {
		pairs = append(
			pairs, NewPair(
				NewNative(field.Left()),
				NewNative(field.Right())))
	}
	return pairs
}
