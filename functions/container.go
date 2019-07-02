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
func (c ConstantExpr) TypeName() string                   { return c().TypeName() }
func (c ConstantExpr) Type() Typed {
	return Define("ϝ → "+c().TypeName(), NewPair(Constant, c().TypeFnc()))
}

func NewExpression(expr func(args ...Expression) Expression) VariadicExpr { return expr }

func (c VariadicExpr) Ident() Expression                  { return c }
func (c VariadicExpr) TypeFnc() TyFnc                     { return Function }
func (c VariadicExpr) TypeNat() d.TyNat                   { return d.Function }
func (c VariadicExpr) String() string                     { return c.TypeName() }
func (c VariadicExpr) Call(args ...Expression) Expression { return c(args...) }
func (c VariadicExpr) Eval(args ...d.Native) d.Native {
	var exprs = make([]Expression, 0, len(args))
	for _, arg := range args {
		exprs = append(exprs, NewNative(arg))
	}
	return c(exprs...)
}
func (c VariadicExpr) FlagType() d.Uint8Val { return Flag_Functional.U() }
func (c VariadicExpr) TypeName() string {
	return "[T] → ϝ → T" + c().Type().TypeName()
}
func (c VariadicExpr) Type() Typed {
	return Define(c().TypeName(), c().TypeFnc())
}

//// NARY EXPRESSION TYPE CONSTRUCTOR
///
// TODO: make nary type safe by deriving type switch from signature
func NewNary(expr Expression, signat ...Expression) NaryExpr {
	// allocate argument-/ and return expression
	var argtype, returntype Expression
	// no signature expression passed, return none definition
	if len(signat) == 0 {
		var none = NewNone()
		argtype = Define(none.TypeName(), none)
		returntype = expr.Type().(TyDef)
	}
	// one signature expression passed as return type, argument type none
	if len(signat) == 1 {
		var none = NewNone()
		argtype = Define(none.TypeName(), none)
		returntype = Define(signat[0].TypeName(), signat[0])
	}
	// number of signature expressions minus return expression equals arity
	var arity = len(signat) - 1
	var argslice = make([]Expression, 0, arity)
	// split off argument expression slice and define return expression type
	argslice, returntype = signat[:arity], Define(signat[arity].TypeName(), signat[arity])
	// allocate vector to hold argument definitions
	var argvec = NewVector()
	// range over argument expression slice
	for _, expr := range argslice {
		// it argument expression is already a type definition, append
		// it to argument definition vector and continue loop
		if expr.FlagType() == Flag_Def.U() {
			if def, ok := expr.(TyDef); ok {
				argvec = argvec.Append(def)
				continue
			}
		}
		// otherwise define type from plain expression and append to
		// argument definition slice
		argvec = argvec.Append(Define(expr.TypeName(), expr))
	}
	// define argument type from vector af argument definitions
	argtype = Define("", argvec)

	// create and return nary expression
	return func(args ...Expression) Expression {
		var arglen = len(args) // count arguments
		if arglen > 0 {        // if arguments got passed
			// argument number satisfies expression arity exactly
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
				// allocate slice for remaining arguments
				var remain []Expression
				// split arguments by arity
				args, remain = args[:arity], args[arity:]
				// allocate results vector and assign result of
				// fully satisfied expression as first element.
				var vec = NewVector(NewNary(expr,
					signat...).Call(args...))
				// as long as remaining arguments satisfy, or
				// oversatisfy arity, assign further elements
				for len(remain) >= arity {
					args, remain = remain[:arity], remain[arity:]
					vec = vec.Append(NewNary(expr,
						signat...).Call(args...))
				}
				// if arguments are depleted return vector
				// containing all results
				if len(remain) == 0 {
					return vec
				}
				// if remaining arguments undersatisfy arity,
				// assign partialy applyed nary as last element
				return vec.Append(NewNary(expr,
					signat...).Call(remain...))
			}
		}
		// no arguments passed, return expression, argument-/ and
		// returntype of nary expression definition
		return NewVector(expr, argtype, returntype)
	}
}

// returns the value returned when calling itself directly, passing arguments
func (n NaryExpr) Ident() Expression    { return n }
func (n NaryExpr) String() string       { return n.TypeName() }
func (n NaryExpr) FlagType() d.Uint8Val { return Flag_Functional.U() }
func (n NaryExpr) TypeFnc() TyFnc       { return n.Expr().TypeFnc() }
func (n NaryExpr) TypeNat() d.TyNat     { return n.Expr().TypeNat() }
func (n NaryExpr) Expr() Expression     { return n().(VecCol)()[0] }
func (n NaryExpr) TypeArgs() Typed      { return n().(VecCol)()[1].(Typed) }
func (n NaryExpr) TypeReturn() Typed    { return n().(VecCol)()[2].(Typed) }
func (n NaryExpr) TypeExpr() Typed      { return n.Expr().Type() }
func (n NaryExpr) TypeName() string     { return n.Type().TypeName() }
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
func (n NaryExpr) Arity() Arity { return Arity(0) }
func (n NaryExpr) Type() Typed  { return n.TypeReturn() }

//// NATIVE EXPRESSION CONSTRUCTOR
///
// returns an expression with native return type implementing the callable
// interface
func New(inf ...interface{}) Expression {
	return NewNative(d.New(inf...))
}

// TODO replace expression wrappers by d.Native.Eval
func NewNative(args ...d.Native) Expression {
	// if any initial arguments have been passed

	var nat = d.NewFromData(args...)
	var tnat = nat.TypeNat()

	switch {
	case tnat.Match(d.Slice):
		if slice, ok := nat.(d.Sliceable); ok {
			return Native(slice.Eval)
		}
	case tnat.Match(d.Unboxed):
		if slice, ok := nat.(d.Sliceable); ok {
			return Native(slice.Eval)
		}
	case tnat.Match(d.Pair):
		if pair, ok := nat.(d.Paired); ok {
			return Native(pair.Eval)
		}
	case tnat.Match(d.Map):
		if set, ok := nat.(d.Mapped); ok {
			return Native(set.Eval)
		}
	default:
		return Native(nat.Eval)
	}
	return Native(func(...d.Native) d.Native { return d.NewNil() })
}

// ATOMIC NATIVE VALUE CONSTRUCTOR
func (n Native) Call(args ...Expression) Expression {
	var nats = make([]d.Native, 0, len(args))
	for _, arg := range args {
		nats = append(nats, arg.Eval())
	}
	return NewNative(n(nats...))
}
func (n Native) TypeFnc() TyFnc                 { return Data }
func (n Native) Eval(args ...d.Native) d.Native { return n(args...) }
func (n Native) TypeNat() d.TyNat               { return n().TypeNat() }
func (n Native) FlagType() d.Uint8Val           { return Flag_Functional.U() }
func (n Native) String() string                 { return n().String() }
func (n Native) TypeName() string               { return n().TypeName() }
func (n Native) Type() Typed {
	return Define(n().TypeName(), New(n().TypeNat()))
}

// NATIVE SLICE VALUE CONSTRUCTOR
func (n NativeCol) Call(args ...Expression) Expression {
	if len(args) > 0 {
		var nats = make([]d.Native, 0, len(args))
		for _, arg := range args {
			nats = append(nats, arg.Eval())
		}
		return NewNative(n(nats...))
	}
	return NewNative(n())
}
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
	return Define(n().TypeName(), New(n().TypeNat()))
}
func (n NativeCol) Slice() []Expression {
	var slice = []Expression{}
	for _, val := range n.SliceNat() {
		slice = append(slice, NewNative(val))
	}
	return slice
}

// NATIVE PAIR VALUE CONSTRUCTOR
func (n NativePair) Call(args ...Expression) Expression {
	var nats = make([]d.Native, 0, len(args))
	for _, arg := range args {
		nats = append(nats, arg.Eval())
	}
	return NewNative(n(nats...))
}
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
	return Define(n().TypeName(), New(n().TypeNat()))
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

// NATIVE SET VALUE CONSTRUCTOR

func (n NativeSet) Call(args ...Expression) Expression {
	var nats = make([]d.Native, 0, len(args))
	for _, arg := range args {
		nats = append(nats, arg.Eval())
	}
	return NewNative(n(nats...))
}
func (n NativeSet) Ident() Expression                    { return n }
func (n NativeSet) Eval(args ...d.Native) d.Native       { return n(args...) }
func (n NativeSet) TypeFnc() TyFnc                       { return Data }
func (n NativeSet) TypeNat() d.TyNat                     { return n().TypeNat() }
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
	return Define(n().TypeName(), New(n().TypeNat()))
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
