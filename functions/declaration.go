package functions

import (
	"strings"

	d "github.com/joergreinhardt/gatwd/data"
)

type (
	ValueType    func() Expression
	FunctionType func(...Expression) Expression

	ArgumentSet    func() []d.Typed
	ExpressionType func(...Expression) Expression
	ParametricType func(...Expression) Expression
	CurryedType    func(...Expression) Expression
	CollectionType func(...Expression) (Expression, Consumeable)
)

//// CONSTANT DECLARATION
///
// declares an expression from a constant function, that returns an expression
func DeclareConstant(fn func() Expression) ValueType { return fn }
func (c ValueType) TypeFnc() TyFnc                   { return Constant | c().TypeFnc() }
func (c ValueType) Type() TyPattern                  { return c().Type() }
func (c ValueType) String() string                   { return c().String() }
func (c ValueType) Call(...Expression) Expression    { return c() }

//// FUNCTION DECLARATION
///
// declares an expression from some generic functions, with a signature
// indicating that it takes expressions as arguments and returns an expression
func DeclareFunction(fn func(...Expression) Expression, reType TyPattern) FunctionType {
	return func(args ...Expression) Expression {
		if len(args) > 0 {
			return fn(args...)
		}
		return reType
	}
}
func (g FunctionType) TypeFnc() TyFnc                     { return Value }
func (g FunctionType) Type() TyPattern                    { return g().(TyPattern) }
func (g FunctionType) String() string                     { return g().String() }
func (g FunctionType) Call(args ...Expression) Expression { return g(args...) }

//// ARGUMENT SET
///
// define a set of arguments as a sequence of argument types.
func DefineArgumentSet(types ...d.Typed) ArgumentSet     { return func() []d.Typed { return types } }
func (a ArgumentSet) TypeFnc() TyFnc                     { return Argument }
func (a ArgumentSet) Type() TyPattern                    { return Def(a()...) }
func (a ArgumentSet) Len() int                           { return len(a()) }
func (a ArgumentSet) Head() Expression                   { return a.Type().Head() }
func (a ArgumentSet) Tail() Consumeable                  { return a.Type().Tail() }
func (a ArgumentSet) Consume() (Expression, Consumeable) { return a.Type().Consume() }
func (a ArgumentSet) String() string {
	var strs = make([]string, 0, a.Len())
	for _, t := range a() {
		strs = append(strs, t.String())
	}
	return strings.Join(strs, " → ")
}
func (a ArgumentSet) Call(args ...Expression) Expression {
	if len(args) > 0 {
		if a.MatchArgs(args...) {
			if len(args) > 1 {
				return NewVector(args...)
			}
			return args[0]
		}
	}
	return NewNone()
}
func (a ArgumentSet) MatchArg(arg Expression) (ArgumentSet, bool) {
	var (
		types   = a()
		current d.Typed
	)
	if len(types) > 0 {
		current = types[0]
	}
	if len(types) > 1 {
		types = types[1:]
	} else {
		types = []d.Typed{}
	}
	return DefineArgumentSet(types...),
		current.Match(arg.Type())
}
func (a ArgumentSet) MatchArgs(args ...Expression) bool {
	var (
		at      = a
		ok      bool
		current Expression
	)
	for len(args) > 0 {
		if len(args) > 0 {
			current = args[0]
		}
		if len(args) > 1 {
			args = args[1:]
		} else {
			args = []Expression{}
		}
		if at, ok = at.MatchArg(current); !ok {
			return ok // ← will be false
		}
	}
	return ok // ← will be true
}

//// TYPE SAFE EXPRESSION
///
// declare a type-safe expression. argument types will be matched with the
// types of passed arguments. declared expression can be applyed partialy. for
// multi parameter function, there a three possible sorts of legal calls:
//
// - a call can be undersatisfied by not passing all arguments. in that case a
//   new UeclaredExpr is returned, with an argument set reduced by the the
//   arguments passed and enclosing those.
//
// - a call can pass the exact right number and types of arguments, in which
//    case they will be applyed to the enclosed expression to yield the result.
//
// - a call can pass a sequence of multiple argument sets in which case a
//   vector of results, the last of which might be a partialy applyed
//   expression, will be returned,
func DeclareExpression(expr Expression, types ...d.Typed) ExpressionType {
	var tlen = len(types)
	return func(args ...Expression) Expression {
		var alen = len(args)
		if alen > 0 {
			switch {

			// satisfied
			case alen == tlen:
				var matcher = DefineArgumentSet(types...)
				if matcher.MatchArgs(args...) {
					return expr.Call(args...)
				}

			// undersatisfied
			case alen < tlen:
				var (
					currenTypes = types[:alen]
					remainTypes = types[alen:]
					matcher     = DefineArgumentSet(currenTypes...)
				)
				if matcher.MatchArgs(args...) {
					return DeclareExpression(DeclareFunction(
						func(lateargs ...Expression) Expression {
							return expr.Call(
								append(
									args,
									lateargs...,
								)...)
						}, expr.Type()), remainTypes...)
				}

			// oversatisfied
			case alen > tlen:
				var (
					currenArgs = args[:tlen]
					remainArgs = args[tlen:]
					matcher    = DefineArgumentSet(types...)
					vec        = NewVector()
				)
				if matcher.MatchArgs(currenArgs...) {
					for len(remainArgs) > 0 {
						if len(remainArgs) >= tlen {
							currenArgs = remainArgs[:tlen]
							remainArgs = remainArgs[tlen:]
						} else {
							currenArgs = remainArgs
							remainArgs = []Expression{}
						}
						vec = vec.Con(
							DeclareExpression(
								expr, types...,
							)(currenArgs...))
					}
					return vec
				}
			}
			return NewNone()
		}
		return NewPair(expr, DefineArgumentSet(types...))
	}
}
func (e ExpressionType) ArgType() ArgumentSet { return e().(PairVal).Right().(ArgumentSet) }
func (e ExpressionType) Unbox() Expression    { return e().(PairVal).Left() }
func (e ExpressionType) Type() TyPattern      { return e.ArgType().Type() }
func (e ExpressionType) TypeFnc() TyFnc       { return Value }
func (e ExpressionType) Call(args ...Expression) Expression {
	if len(args) > 0 {
		return e(args...)
	}
	return e.Unbox()
}
func (e ExpressionType) String() string {
	return strings.Join(append(
		make(
			[]string, 0,
			e.ArgType().Len(),
		),
		e.ArgType().String(),
		e.Unbox().Type().String(),
		e.Unbox().Type().String()),
		" → ",
	)
}

//// PARAMETRIC EXPRESSION
///
// the parametric expression constructor returns a parametric type by
// constructing a switch from a sequence of expression-type arguments, by
// declaring a case per expression type that tests, if the arguments passed
// during runtime match the expression type, or return none instead. first
// result from applying argument-set successfull to an expression constructor
// will be returned
func DeclareParametricExpression(exprs ...ExpressionType) ParametricType {
	var cases = make([]CaseType, 0, len(exprs))
	for _, expr := range exprs {
		cases = append(cases, DeclareCase(
			DeclareTest(func(args ...Expression) bool {
				return !expr.Call(args...).TypeFnc().Match(None)
			}), expr))
	}
	return ParametricType(func(args ...Expression) Expression {
		if len(args) > 0 {
			return NewSwitch(cases...).Call(args...)
		}
		return NewSwitch(cases...)
	})
}

func (p ParametricType) TypeFnc() TyFnc { return Parametric }

func (p ParametricType) Unbox() Expression { return p() } // ← switch-type
func (p ParametricType) String() string    { return p().(SwitchType).String() }
func (p ParametricType) Cases() []CaseType { return p().(SwitchType).Cases() }
func (p ParametricType) Len() int          { return len(p.Cases()) }

// yield slice of expressions enclosed by cases
func (p ParametricType) Slice() []Expression {
	var exprs = make([]Expression, 0, p.Len())
	for _, c := range p.Cases() {
		// discard testable
		var _, expr = c.Unbox()
		exprs = append(exprs, expr)
	}
	return exprs
}

// yield types of expressions enclosed by cases
func (p ParametricType) Type() TyPattern {
	var length = p.Len()
	var types = []d.Typed{}
	for n, expr := range p.Slice() {
		types = append(types, expr.Type())
		if n < length-1 {
			types = append(types, Lex_Pipe)
		}
	}
	return Def(types...)
}

// call method calls the enclosed switch to yield either none, or result of
// applying the arguments to the first matching case.
func (p ParametricType) Call(args ...Expression) Expression {
	if len(args) > 0 {
		return p(args...)
	}
	return p()
}

//// CURRY
///
//
func Curry(fns ...ExpressionType) ExpressionType {

	if len(fns) > 0 {

		var (
			expr ExpressionType
		)

		if len(fns) == 1 {
			expr = fns[0]
		}
		if len(fns) == 2 {
			expr = Curry(fns[0], fns[1])
		}
		if len(fns) > 2 {
			expr = Curry(append([]ExpressionType{
				Curry(fns[0], fns[1]),
			}, fns[2:]...)...)
		}

		if expr != nil {
			if !expr.TypeFnc().Match(None) {
				return DeclareExpression(DeclareFunction(
					func(args ...Expression) Expression {
						if len(args) > 0 {
							return expr.Call(args...)
						}
						return expr.Call()
					}, fns[0].Unbox().Type()), fns[0].Type())
			}
		}
	}

	return DeclareExpression(NewNone(), None)
}

//// TYPE SAFE COLLECTIONS
///
//
func DeclareCollection(col Consumeable, argtype d.Typed) CollectionType {
	return func(args ...Expression) (Expression, Consumeable) {
		if len(args) == 1 {
			var arg = args[0]
			if arg.Type().Match(argtype) {
				col = col.Append(arg)
			}
			if arg.TypeFnc().Match(Type) {
				if t, ok := arg.(TyFnc); ok {
					if t.Match(Type) {
						return DeclareNative(true), col
					}
				}
			}
		}
		if len(args) > 1 {
			var fas = make([]Expression, 0, len(args))
			for _, arg := range args {
				if argtype.Match(arg.Type()) {
					fas = append(fas, arg)
				}
			}
			col = col.Append(fas...)
		}
		var head, tail = col.Consume()
		return head, DeclareCollection(tail, argtype)
	}
}
func (c CollectionType) Consume() (Expression, Consumeable) { return c() }
func (c CollectionType) Head() Expression {
	var head, _ = c()
	return head
}
func (c CollectionType) Tail() Consumeable {
	var _, tail = c()
	return tail
}
func (c CollectionType) Unbox() Consumeable {
	var _, unboxed = c(Type)
	return unboxed
}
func (c CollectionType) TypeFnc() TyFnc    { return c.Unbox().TypeFnc() }
func (c CollectionType) Type() TyPattern   { return c.Unbox().Type() }
func (c CollectionType) TypeElem() d.Typed { return c.Unbox().TypeElem() }
func (c CollectionType) Len() int          { return c.Unbox().Len() }
func (c CollectionType) String() string    { return c.Unbox().String() }
func (c CollectionType) Append(args ...Expression) Consumeable {
	var _, col = c(args...)
	return col
}
func (c CollectionType) Call(args ...Expression) Expression {
	if len(args) > 0 {
		var head, tail = c(args...)
		return NewPair(head, tail)
	}
	var head, tail = c()
	return NewPair(head, tail)
}
