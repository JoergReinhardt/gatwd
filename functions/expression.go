package functions

import (
	"strings"

	d "github.com/joergreinhardt/gatwd/data"
)

type (
	// GENERIC EXPRESSIONS
	NoneVal  func()
	ConstVal func() Expression
	ExprVal  func(...Expression) Expression
	ExprType func(...Expression) Expression

	// TUPLE TYPE (FIELD | ‥. | FIELD)
	TupleVal  []Expression
	TupleType func(...Expression) TupleVal
)

//// NONE VALUE CONSTRUCTOR
///
// none represens the abscence of a value of any type. implements countable,
// sliceable, consumeable, testable, compareable, key-, index- and generic pair
// interfaces to be able to stand in as return value for such expressions.
func NewNone() NoneVal { return func() {} }

func (n NoneVal) Head() Expression                   { return n }
func (n NoneVal) Tail() Consumeable                  { return n }
func (n NoneVal) Append(...Expression) Consumeable   { return n }
func (n NoneVal) Len() int                           { return 0 }
func (n NoneVal) Compare(...Expression) int          { return -1 }
func (n NoneVal) String() string                     { return "⊥" }
func (n NoneVal) Call(...Expression) Expression      { return nil }
func (n NoneVal) Key() Expression                    { return nil }
func (n NoneVal) Index() Expression                  { return nil }
func (n NoneVal) Left() Expression                   { return nil }
func (n NoneVal) Right() Expression                  { return nil }
func (n NoneVal) Both() Expression                   { return nil }
func (n NoneVal) Value() Expression                  { return nil }
func (n NoneVal) Empty() d.BoolVal                   { return true }
func (n NoneVal) Test(...Expression) bool            { return false }
func (n NoneVal) TypeFnc() TyFnc                     { return None }
func (n NoneVal) TypeNat() d.TyNat                   { return d.Nil }
func (n NoneVal) Type() TyPattern                    { return Def(None) }
func (n NoneVal) TypeElem() TyPattern                { return Def(None) }
func (n NoneVal) TypeName() string                   { return n.String() }
func (n NoneVal) Slice() []Expression                { return []Expression{} }
func (n NoneVal) Flag() d.BitFlag                    { return d.BitFlag(None) }
func (n NoneVal) FlagType() d.Uint8Val               { return Flag_Function.U() }
func (n NoneVal) Consume() (Expression, Consumeable) { return NewNone(), NewNone() }

//// CONSTANT DECLARATION
///
// declares a constant value
func DecConstant(constant func() Expression) ConstVal { return constant }

func (c ConstVal) Type() TyPattern {
	return Def(None, Def(Constant, c().Type().TypeIdent()), c().Type())
}
func (c ConstVal) TypeFnc() TyFnc                { return Constant }
func (c ConstVal) String() string                { return c().String() }
func (c ConstVal) Call(...Expression) Expression { return c() }

//// EXPRESSION DECLARATION
///
// declares an expression with defined argument-, return- and an optional identity type
func DecFuntion(fn func(...Expression) Expression) ExprVal { return fn }

func (g ExprVal) TypeFnc() TyFnc                     { return Value }
func (g ExprVal) Type() TyPattern                    { return Def(Value) }
func (g ExprVal) String() string                     { return g.Type().TypeName() }
func (g ExprVal) Call(args ...Expression) Expression { return g(args...) }

/// PARTIAL APPLYABLE EXPRESSION VALUE
//
// element values yield a subelements of optional, tuple, or enumerable
// expressions with sub-type pattern as second return value
func ConstructExpressionType(
	expr Expression,
	argtype, retype d.Typed,
	identypes ...d.Typed,
) ExprType {

	var (
		arglen         int
		ident, pattern TyPattern
	)
	if len(identypes) == 0 {
		ident = Def(expr.TypeFnc())
	} else {
		ident = Def(identypes...)
	}
	if Flag_Pattern.Match(argtype.FlagType()) {
		arglen = len(argtype.(TyPattern).Patterns())
	} else {
		arglen = 1
	}
	pattern = Def(argtype, ident, retype)

	return func(args ...Expression) Expression {
		var length = len(args)
		if length > 0 {
			if pattern.TypeArguments().MatchArgs(args...) {
				var result Expression
				switch {
				case length == arglen:
					result = expr.Call(args...)
					return result

				case length < arglen:
					var argtypes = make(
						[]d.Typed, 0,
						len(pattern.TypeArguments()[length:]),
					)
					for _, atype := range pattern.TypeArguments()[length:] {
						argtypes = append(argtypes, atype)
					}
					var pattern = Def(Def(argtypes...), ident, retype)
					return ConstructExpressionType(ExprVal(
						func(lateargs ...Expression) Expression {
							if len(lateargs) > 0 {
								return expr.Call(append(
									args, lateargs...,
								)...)
							}
							return pattern
						}), Def(argtypes...), ident, retype)

				case length > arglen:
					var vector = NewVector()
					for len(args) > arglen {
						vector = vector.Con(
							expr.Call(args[:arglen]...))
						args = args[arglen:]
					}
					if length > 0 {
						vector = vector.Con(ConstructExpressionType(
							expr, argtype, retype, identypes...,
						).Call(args...))
					}
					return vector
				}
			}
			return NewNone()
		}
		return pattern
	}
}
func (e ExprType) TypeFnc() TyFnc                     { return Value }
func (e ExprType) Type() TyPattern                    { return e().(TyPattern) }
func (e ExprType) String() string                     { return e().String() }
func (e ExprType) Call(args ...Expression) Expression { return e(args...) }

//// PARAMETRIC VALUE
///
// declare parametric value from set of declared expressions.
func ConstructParametricType(exprs ...ExprType) SwitchType {
	var cases = make([]CaseType, 0, len(exprs))
	for _, expr := range exprs {
		cases = append(cases, CaseType(expr))
	}
	return DecSwitch(cases...)
}

//// TUPLE VALUE CONSTRUCTOR
///
// returns a tuple type constructor, defined by tuples field types. the first
// typed argument passed may be the types symbolic name.
//
// tuple type constructor either expects a sequence of flat args, or pairs with
// an index assigned to the left fields corresponding to fields position, for
// cases where the constructor is wrapped as declared expression and arguments
// are not intendet to be passed in the correct order. non-pair arguments are
// assigned in the order they are passed in.  field types will be replaced with
// corresponding argument if it matches the fields type and an array with the
// tuple types name as first and its fields as succeeding elements will be
// returned.
//
// if partial application is intendet, constructor needs to be declared as
// expression by calling the declaration method.
func ConstructTupleType(types ...d.Typed) TupleType {
	var (
		symbol TySymbol
		ftypes []Expression
	)
	// return if no types where passed
	if len(types) == 0 {
		return nil
	}
	if len(types) > 0 {
		// set tuple typed name by first element, if its a symbol
		if Flag_Symbol.Match(types[0].FlagType()) {
			// return early if nothing but the name has been passed
			if len(types) == 1 {
				return nil
			}
			symbol = types[0].(TySymbol)
			// shift fields
			types = types[1:]
		} else {
			// generate tuple types name from type pattern
			symbol = DefSym(Tuple.String())
		}
	}
	// allocate field set with an extra element to hold the name
	ftypes = make([]Expression, 0, len(types))
	ftypes = append(ftypes, symbol)
	// expect all remaining type arguments, to be tuples field types
	for _, typ := range types {
		if Flag_Pattern.Match(typ.FlagType()) {
			ftypes = append(ftypes, typ.(TyPattern))
		}
	}
	return func(args ...Expression) TupleVal {
		var fields []Expression
		if fields == nil {
			fields = make([]Expression, 0, len(types))
			fields = append(fields, symbol)
		}
		if len(args) == len(ftypes)-1 {
			for n, arg := range args {
				if ftypes[n+1].Type().Match(arg.Type()) {
					fields = append(fields, arg)
				}
			}
			return fields
		}
		return ftypes
	}
}
func (t TupleType) TypeFnc() TyFnc                       { return Type | Tuple }
func (t TupleType) Call(args ...Expression) Expression   { return t(args...) }
func (t TupleType) Allocate(args ...Expression) TupleVal { return t(args...) }
func (t TupleType) Symbol() TySymbol                     { return t()[0].(TySymbol) }
func (t TupleType) Fields() []Expression                 { return t()[1:] }
func (t TupleType) Declare() ExprType {
	var (
		fields = t.Fields()
		types  = make([]d.Typed, 0, len(fields))
	)
	for _, field := range fields {
		types = append(types, field.Type())
	}
	return ConstructExpressionType(t, Def(types...),
		t()[0].(TySymbol), Def(types...))
}
func (t TupleType) Type() TyPattern {
	var (
		fields = t.Fields()
		types  = make([]d.Typed, 0, len(fields))
	)
	for _, field := range fields {
		types = append(types, field.Type())
	}
	return Def(t.Symbol(), Def(types...))
}
func (t TupleType) String() string {
	var (
		fields = t()
		strs   = make([]string, 0, len(fields))
	)
	for _, field := range fields {
		strs = append(strs, field.String())
	}
	return strings.Join(strs, ", ")
}

//// TUPLE VALUE
///
//
func (t TupleVal) TypeFnc() TyFnc                { return Tuple }
func (t TupleVal) Symbol() TySymbol              { return t[0].(TySymbol) }
func (t TupleVal) Name() string                  { return t.Symbol().String() }
func (t TupleVal) Fields() []Expression          { return t[1:] }
func (t TupleVal) Len() int                      { return len(t.Fields()) }
func (t TupleVal) Call(...Expression) Expression { return TupleVal(t.Fields()) }
func (t TupleVal) Type() TyPattern {
	var types = make([]d.Typed, 0, len(t))
	for _, field := range t.Fields() {
		types = append(types, field.Type())
	}
	return Def(t.Symbol(), Def(types...))
}
func (t TupleVal) String() string {
	var strs = make([]string, 0, len(t.Fields()))
	for _, field := range t.Fields() {
		strs = append(strs, field.String())
	}
	return "[" + strings.Join(strs, " ") + "]"
}
