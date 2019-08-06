package functions

type (
	ConstantType func() Expression
	FunctionType func(...Expression) Expression
)

//// CONSTANT DECLARATION
///
// declares an expression from a constant function, that returns an expression
func DeclareConstant(fn func() Expression) ConstantType { return fn }

func (c ConstantType) Call(...Expression) Expression { return c() }
func (c ConstantType) TypeFnc() TyFnc                { return Constant }
func (c ConstantType) String() string                { return c().String() }
func (c ConstantType) Type() TyPattern {
	return Def(None, Constant, c().Type())
}

//// FUNCTION DECLARATION
///
// declares an expression from some generic functions, with a signature
// indicating that it takes expressions as arguments and returns an expression
func DeclareFunction(
	fn func(...Expression) Expression,
	identype, argtype, retype TyPattern,
) FunctionType {
	return func(args ...Expression) Expression {
		if len(args) > 0 {
			return fn(args...)
		}
		return NewVector(argtype, identype, retype)
	}
}
func (g FunctionType) TypeFnc() TyFnc                     { return Value }
func (g FunctionType) String() string                     { return g().String() }
func (g FunctionType) Call(args ...Expression) Expression { return g(args...) }
func (g FunctionType) Type() TyPattern {
	var vec = g().(VecType)()
	return Def(vec[0].(TyPattern), vec[1].(TyPattern), vec[2].(TyPattern))
}
