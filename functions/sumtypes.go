package functions

import (
	"strings"

	d "github.com/joergreinhardt/gatwd/data"
)

type (
	// GENERIC EXPRESSIONS
	NoneVal      func()
	GenericConst func() Expression
	GenericFunc  func(...Expression) Expression

	//// NAMED EXPRESSION
	NameDef func(...Expression) (Expression, TyComp)

	//// DECLARED EXPRESSION
	FuncDef func(...Expression) Expression

	// TUPLE (TYPE[0]...TYPE[N])
	TupleDef func(...Expression) Expression
	TupleVal func() ([]Expression, TyComp)

	// RECORD (PAIR(KEY, VAL)[0]...PAIR(KEY, VAL)[N])
	RecordDef func(...KeyPair) RecordVal
	RecordVal func() ([]KeyPair, TyComp)
)

//// NONE VALUE CONSTRUCTOR
///
// none represens the abscence of a value of any type. implements countable,
// sliceable, consumeable, testable, compareable, key-, index- and generic pair
// interfaces to be able to stand in as return value for such expressions.
func NewNone() NoneVal { return func() {} }

func (n NoneVal) Head() Expression                   { return n }
func (n NoneVal) Tail() Consumeable                  { return n }
func (n NoneVal) Cons(...Expression) Sequential      { return n }
func (n NoneVal) Prepend(...Expression) Sequential   { return n }
func (n NoneVal) Append(...Expression) Sequential    { return n }
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
func (n NoneVal) Type() TyComp                       { return Def(None) }
func (n NoneVal) TypeElem() TyComp                   { return Def(None) }
func (n NoneVal) TypeName() string                   { return n.String() }
func (n NoneVal) Slice() []Expression                { return []Expression{} }
func (n NoneVal) Flag() d.BitFlag                    { return d.BitFlag(None) }
func (n NoneVal) FlagType() d.Uint8Val               { return Kind_Fnc.U() }
func (n NoneVal) Consume() (Expression, Consumeable) { return NewNone(), NewNone() }

//// GENERIC CONSTANT DEFINITION
///
// declares a constant value
func NewConstant(constant func() Expression) GenericConst { return constant }

func (c GenericConst) Type() TyComp                  { return Def(Constant, c().Type()) }
func (c GenericConst) TypeFnc() TyFnc                { return Constant }
func (c GenericConst) String() string                { return c().String() }
func (c GenericConst) Call(...Expression) Expression { return c() }

//// GENERIC FUNCTION DEFINITION
///
// declares a constant value
func NewFunction(fnc func(...Expression) Expression) GenericFunc {
	return func(args ...Expression) Expression {
		if len(args) > 0 {
			return fnc(args...)
		}
		return fnc()
	}
}

func (c GenericFunc) Call(args ...Expression) Expression {
	if len(args) > 0 {
		return c(args...)
	}
	return c()
}
func (c GenericFunc) String() string { return c().String() }
func (c GenericFunc) Type() TyComp   { return c().Type() }
func (c GenericFunc) TypeFnc() TyFnc { return c().TypeFnc() }

//// NAMED EXPRESSION DEFINITION
///
// declares a constant value
func NewNamedDefinition(
	fnc func(...Expression) Expression,
	name string,
	retype d.Typed,
	argtypes ...d.Typed,
) NameDef {
	var tc = Def(DefSym(name), retype, Def(argtypes...))
	return NameDef(func(args ...Expression) (Expression, TyComp) {
		if len(args) > 0 {
			return fnc(args...), tc
		}
		return fnc(), tc
	})
}

// define named alias from defined expression
func NewAliasDefinition(
	expr Expression,
	name string,
) NameDef {
	var tc = Def(
		DefSym(name),
		expr.Type().TypeReturn(),
		expr.Type().TypeArguments(),
	)
	return NameDef(func(args ...Expression) (Expression, TyComp) {
		if len(args) > 0 {
			return expr.Call(args...), tc
		}
		return expr, tc
	})
}
func (c NameDef) Expr() Expression { var e, _ = c(); return e }
func (c NameDef) Type() TyComp     { var _, t = c(); return t }
func (c NameDef) TypeFnc() TyFnc   { return c.Expr().TypeFnc() }
func (c NameDef) String() string   { return c.Expr().String() }
func (c NameDef) Call(args ...Expression) Expression {
	if len(args) > 0 {
		return c.Expr().Call(args...)
	}
	return c.Expr()
}

//// TUPLE TYPE
///
// tuple type constructor expects a slice of field types and possibly a symbol
// type flag, to define the types name, otherwise 'tuple' is the type name and
// the sequence of field types is shown instead
//
// tuple definition helper either|or type definition.
// f(args VecVal, types TyComp) (TupleDef|TupleVal)
func createTupleType(types ...d.Typed) TyComp {
	if len(types) > 0 {
		if Kind_Sym.Match(types[0].Kind()) {
			return Def(types...)
		}
		return Def(append([]d.Typed{Tuple}, types...)...)
	}
	return Def(None)
}
func NewTuple(types ...d.Typed) TupleDef {
	var (
		tt  = createTupleType(types...)
		def = Define(TupleDef(func(args ...Expression) Expression {
			return TupleVal(func() ([]Expression, TyComp) { return args, tt })
		}), tt.Types()...)
	)
	return TupleDef(func(args ...Expression) Expression {
		if len(args) > 0 {
			return def(args...)
		}
		return tt
	})
}

func (t TupleDef) Call(args ...Expression) Expression { return t(args...) }
func (t TupleDef) TypeFnc() TyFnc                     { return Tuple | Constructor }
func (t TupleDef) String() string                     { return t.Type().String() }
func (t TupleDef) Type() TyComp                       { return Def() }

/// TUPLE VALUE
// tuple value is a slice of expressions, constructed by a tuple type
// constructor validated according to its type pattern.
func (t TupleVal) TypeFnc() TyFnc                { return Tuple }
func (t TupleVal) Type() TyComp                  { var _, typ = t(); return typ }
func (t TupleVal) Value() []Expression           { var v, _ = t(); return v }
func (t TupleVal) Count() int                    { return len(t.Value()) }
func (t TupleVal) Call(...Expression) Expression { return NewVector(t.Value()...) }
func (t TupleVal) Get(idx int) Expression {
	if idx < t.Count() {
		return t.Value()[idx]
	}
	return NewNone()
}
func (t TupleVal) String() string {
	var strs = make([]string, 0, t.Count())
	for _, val := range t.Value() {
		strs = append(strs, val.String())
	}
	return "[" + strings.Join(strs, ", ") + "]"
}

//// RECORD TYPE

/// PARTIAL APPLYABLE EXPRESSION VALUE
//
// defines typesafe partialy applicable expression. if the set of optional type
// argument(s) starts with a symbol, that will be assumed to be the types
// identity. otherwise the identity is derived from the passed expression,
// types first field will be the return type, its second field the (set of)
// argument type(s), additional arguments are considered propertys.
func createFuncType(expr Expression, types ...d.Typed) TyComp {
	// if type arguments have been passed, build the type based on them‥.
	if len(types) > 0 {
		// if the first element in pattern is a symbol to be used as
		// ident, just define type from type arguments‥.
		if Kind_Sym.Match(types[0].Kind()) {
			return Def(types...)
		} else { // ‥.otherwise use the expressions ident type
			return Def(append([]d.Typed{expr.Type().TypeIdent()}, types...)...)
		}
	} // ‥.otherwise define by expressions identity, return-/ &
	// argument types
	return Def(expr.Type().TypeIdent(),
		expr.Type().TypeReturn(),
		expr.Type().TypeArguments())

}
func Define(
	expr Expression, types ...d.Typed,
) FuncDef {
	var (
		ct     = createFuncType(expr, types...)
		arglen = len(ct.TypeArguments())
	)
	// return partialy applicable function
	return func(args ...Expression) Expression {
		var length = len(args)
		if length > 0 {
			if ct.TypeArguments().MatchArgs(args...) {
				switch {
				// NUMBER OF PASSED ARGUMENTS MATCHES EXACTLY →
				case length == arglen:
					return expr.Call(args...)

				// NUMBER OF PASSED ARGUMENTS IS INSUFFICIENT →
				case length < arglen:
					// safe types of arguments remaining to be filled
					var (
						remains = ct.TypeArguments().Types()[length:]
						newpat  = Def(
							ct.Type().TypeIdent(),
							ct.Type().TypeReturn(),
							Def(remains...))
					)
					// define new function from remaining
					// set of argument types, enclosing the
					// current arguments & appending its
					// own aruments to them, when called.
					return Define(GenericFunc(func(lateargs ...Expression) Expression {
						// will return result, or
						// another partial, when called
						// with arguments
						if len(lateargs) > 0 {
							return expr.Call(append(
								args, lateargs...,
							)...)
						}
						// if no arguments where
						// passed, return the reduced
						// type ct
						return newpat
					}), newpat.TypeIdent(), newpat.TypeReturn(), newpat.TypeArguments())

				// NUMBER OF PASSED ARGUMENTS OVERSATISFYING →
				case length > arglen:
					// allocate vector to hold multiple instances
					var vector = NewVector()
					// iterate over arguments, allocate an instance per satisfying set
					for len(args) > arglen {
						vector = vector.Cons(
							expr.Call(args[:arglen]...)).(VecVal)
						args = args[arglen:]
					}
					if length > 0 { // number of leftover arguments is insufficient
						// add a partial expression as vectors last element
						vector = vector.Cons(Define(
							expr, ct.Types()...,
						).Call(args...)).(VecVal)
					}
					// return vector of instances
					return vector
				}
			}
			// passed argument(s) didn't match the expected type(s)
			return None
		}
		// no arguments where passed, return the expression type
		return ct
	}
}
func (e FuncDef) TypeFnc() TyFnc                     { return Constructor | Value }
func (e FuncDef) Type() TyComp                       { return e().(TyComp) }
func (e FuncDef) TypeArguments() TyComp              { return e.Type().TypeArguments() }
func (e FuncDef) TypeReturn() TyComp                 { return e.Type().TypeReturn() }
func (e FuncDef) ArgCount() int                      { return e.Type().TypeArguments().Count() }
func (e FuncDef) String() string                     { return e().String() }
func (e FuncDef) Call(args ...Expression) Expression { return e(args...) }
