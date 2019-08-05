package functions

import (
	"strings"

	d "github.com/joergreinhardt/gatwd/data"
)

type (
	ConstantType    func() Expression
	GeneratorType   func(...Expression) Expression
	FunctionType    func(...Expression) Expression
	AccumulatorType func(...Expression) (Expression, AccumulatorType)

	ArgumentSet    func() []d.Typed
	ExpressionType func(...Expression) Expression
	ParametricType func(...Expression) Expression
	CurryedType    func(...Expression) Expression
	CollectionType func(...Expression) (Expression, Consumeable)
)

//// CONSTANT DECLARATION
///
// declares an expression from a constant function, that returns an expression
func DeclareConstant(fn func() Expression) ConstantType { return fn }
func (c ConstantType) TypeFnc() TyFnc                   { return Constant | c().TypeFnc() }
func (c ConstantType) Type() TyPattern                  { return c().Type() }
func (c ConstantType) String() string                   { return c().String() }
func (c ConstantType) Call(...Expression) Expression    { return c() }

//// GENERATOR DECLARATION
///
// declares an expression generating a series of values of the same type,
// yielding the next element in sequence with every call.
func DeclareGenerator(constructor func(...Expression) func() Expression, retype TyPattern) GeneratorType {

	var generator = constructor()

	return func(args ...Expression) Expression {
		if len(args) > 0 {
			if args[0].Type().Match(Type) {
				return retype
			}
			generator = constructor(args...)
		}
		var result = generator()
		generator = constructor(result)
		return result
	}
}
func (c GeneratorType) TypeFnc() TyFnc                { return Generator | c(Type).TypeFnc() }
func (c GeneratorType) Type() TyPattern               { return c(Type).(TyPattern) }
func (c GeneratorType) String() string                { return c().String() }
func (c GeneratorType) Call(...Expression) Expression { return c() }

//// FUNCTION DECLARATION
///
// declares an expression from some generic functions, with a signature
// indicating that it takes expressions as arguments and returns an expression
func DeclareFunction(fn func(...Expression) Expression, retype TyPattern) FunctionType {
	return func(args ...Expression) Expression {
		if len(args) > 0 {
			return fn(args...)
		}
		return retype
	}
}
func (g FunctionType) TypeFnc() TyFnc                     { return Value | g().TypeFnc() }
func (g FunctionType) Type() TyPattern                    { return g().(TyPattern) }
func (g FunctionType) String() string                     { return g().String() }
func (g FunctionType) Call(args ...Expression) Expression { return g(args...) }

//// ACCUMULATOR DECLARATION
///
// declares an expression accumulating results in a value intendet to be
// reassigned to itself with every accumulation.
func DeclareAccumulator(acc func(...Expression) (Expression, AccumulatorType)) AccumulatorType {
	return func(args ...Expression) (Expression, AccumulatorType) {
		var head Expression
		if len(args) > 0 {
			head, acc = acc(args...)
			return head, acc
		}
		return acc()
	}
}
func (g AccumulatorType) Expr() Expression {
	var expr, _ = g()
	return expr
}
func (g AccumulatorType) Accumulator() AccumulatorType {
	var _, acc = g()
	return acc
}
func (g AccumulatorType) TypeFnc() TyFnc {
	return Accumulator | g.Expr().TypeFnc()
}
func (g AccumulatorType) Type() TyPattern                    { return g.Expr().Type() }
func (g AccumulatorType) String() string                     { return g.Expr().String() }
func (g AccumulatorType) Call(args ...Expression) Expression { return g.Expr().Call(args...) }

//// ARGUMENT SET
///
// define a set of arguments as a sequence of argument types.
func DefineArgumentSet(types ...d.Typed) ArgumentSet     { return func() []d.Typed { return types } }
func (a ArgumentSet) TypeFnc() TyFnc                     { return Argument }
func (a ArgumentSet) Type() TyPattern                    { return Def(a()...) }
func (a ArgumentSet) Head() Expression                   { return a.Type().Head() }
func (a ArgumentSet) Tail() Consumeable                  { return a.Type().Tail() }
func (a ArgumentSet) Consume() (Expression, Consumeable) { return a.Type().Consume() }
func (a ArgumentSet) Len() int                           { return len(a()) }
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
	return DeclareNone()
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
			return DeclareNone()
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
			return DeclareSwitch(cases...).Call(args...)
		}
		return DeclareSwitch(cases...)
	})
}

func (p ParametricType) TypeFnc() TyFnc { return Parametric }

func (p ParametricType) Unbox() Expression { return p() } // ← switch-type
func (p ParametricType) Cases() []CaseType { return p().(SwitchType).Cases() }
func (p ParametricType) String() string    { return p().(SwitchType).String() }
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
	switch len(fns) {
	case 0:
		return DeclareExpression(DeclareNone(), None)
	case 1:
		return fns[0]
	case 2:
		return DeclareExpression(DeclareFunction(
			func(args ...Expression) Expression {
				if len(args) > 0 {
					return fns[0].Call(
						fns[1]).Call(
						args...)
				}
				return fns[0].Call(fns[1])
			}, fns[1].Type()),
			Def(fns[1].Type(), fns[0].Type()))
	}
	var pattern TyPattern
	for _, fn := range fns[:len(fns)-1] {
		if pattern == nil {
			pattern = fn.Type()
			continue
		}
		pattern = Def(pattern, fn.Type())
	}
	return DeclareExpression(DeclareFunction(
		func(args ...Expression) Expression {
			var expr = Curry(append(
				[]ExpressionType{
					Curry(fns[0], fns[1]),
				}, fns[2:]...)...)
			if len(args) > 0 {
				return expr.Call(args...)
			}
			return expr.Call()
		}, fns[len(fns)-1].Type()), pattern)
}

//// TYPE SAFE COLLECTIONS
///
//
func DeclareCollection(col Consumeable, elemtype TyPattern) CollectionType {
	return func(args ...Expression) (Expression, Consumeable) {
		if len(args) > 0 {
			if len(args) == 1 &&
				args[0].Type().Match(Type) {
				return elemtype, col
			}
			var fas = make([]Expression, 0, len(args))
			for _, arg := range args {
				if elemtype.Match(arg.Type()) {
					fas = append(fas, arg)
				}
			}
			col = col.Append(fas...)
		}
		var head, tail = col.Consume()
		return head, DeclareCollection(tail, elemtype)
	}
}
func (c CollectionType) Type() TyPattern                    { return c.Unbox().Type() }
func (c CollectionType) TypeFnc() TyFnc                     { return c.Unbox().TypeFnc() }
func (c CollectionType) Len() int                           { return c.Unbox().Len() }
func (c CollectionType) String() string                     { return c.Unbox().String() }
func (c CollectionType) Consume() (Expression, Consumeable) { return c() }
func (c CollectionType) Head() Expression {
	var head, _ = c()
	return head
}
func (c CollectionType) Tail() Consumeable {
	var _, tail = c()
	return tail
}
func (c CollectionType) TypeElem() d.Typed {
	var elemtype, _ = c(Type)
	return elemtype.(TyPattern)
}
func (c CollectionType) Unbox() Consumeable {
	var _, col = c(Type)
	return col
}
func (c CollectionType) Call(args ...Expression) Expression {
	if len(args) > 0 {
		var _, col = c(args...)
		return col
	}
	var head, _ = c()
	return head
}

func (c CollectionType) Append(args ...Expression) Consumeable {
	var (
		cons  = c.Unbox()
		etype TyPattern
	)
	if Flag_Pattern.Match(cons.TypeElem().FlagType()) {
		etype = cons.TypeElem().(TyPattern)
	} else {
		etype = Def(cons.TypeElem())
	}
	return DeclareCollection(cons.Append(args...), etype)
}

func (c CollectionType) Concat(colls ...Consumeable) CollectionType {
	return func(args ...Expression) (Expression, Consumeable) {
		var (
			cons = c.Unbox()
			head Expression
			coll CollectionType
		)
		if len(colls) > 0 {
			if len(args) > 0 {
				cons = cons.Append(args...)
			}
			head, cons = cons.Consume()
			if head != nil {
				coll = cons.(CollectionType)
				return head, coll
			}
			coll = colls[0].(CollectionType)
			if len(colls) > 1 {
				colls = colls[1:]
			} else {
				colls = []Consumeable{}
			}
		}
		return coll.Consume()
	}
}

func (c CollectionType) Map(fn Expression) CollectionType {
	return func(args ...Expression) (Expression, Consumeable) {
		var (
			cons = c.Unbox()
			head Expression
			col  CollectionType
		)
		if len(args) > 0 {
			cons = cons.Append(args...)
		}
		head, cons = cons.Consume()
		col = cons.(CollectionType).Map(fn)
		if head != nil {
			return fn.Call(head), col
		}
		return nil, col
	}
}

func (c CollectionType) Apply(cons Consumeable) CollectionType {

	var pattern TyPattern
	if Flag_Pattern.Match(cons.TypeElem().FlagType()) {
		pattern = cons.TypeElem().(TyPattern)
	} else {
		pattern = Def(cons.TypeElem())
	}
	var fns = DeclareCollection(cons, pattern)

	return func(args ...Expression) (Expression, Consumeable) {
		var fn, fns = fns.Consume()
		if fn != nil {
			return c.Map(fn).Concat(c.Apply(fns)).Consume()
		}
		return c.Consume()
	}
}
func (c CollectionType) FoldL(acc AccumulatorType) CollectionType {
	return func(args ...Expression) (Expression, Consumeable) {
		var (
			cons   = c.Unbox()
			result Expression
			head   Expression
			col    CollectionType
		)
		if len(args) > 0 {
			cons = cons.Append(args...)
		}
		head, cons = cons.Consume()
		col = cons.(CollectionType).FoldL(acc)
		result, acc = acc(head)
		return result, col
	}
}

func (c CollectionType) Filter(filter TestType) CollectionType {
	return func(args ...Expression) (Expression, Consumeable) {
		var (
			cons = c.Unbox()
			col  CollectionType
			head Expression
		)
		if len(args) > 0 {
			cons = cons.Append(args...)
		}
		head, cons = cons.Consume()
		col = cons.(CollectionType).Filter(filter)
		if head != nil {
			if filter(head) {
				return head, col
			}
		}
		return nil, col
	}
}
