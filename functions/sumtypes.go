package functions

import (
	"strings"

	d "github.com/joergreinhardt/gatwd/data"
)

type (
	// GENERIC EXPRESSIONS
	NoneVal func()
	Const   func() Expression
	Lambda  func(...Expression) Expression

	//// DECLARED EXPRESSION
	FuncDef func(...Expression) Expression

	// TUPLE (TYPE[0]...TYPE[N])
	TupDef func(...Expression) TupVal
	TupVal []Expression

	// RECORD (PAIR(KEY, VAL)[0]...PAIR(KEY, VAL)[N])
	RecDef func(...Expression) RecVal
	RecVal []KeyPair
)

//// NONE VALUE CONSTRUCTOR
///
// none represens the abscence of a value of any type. implements countable,
// sliceable, consumeable, testable, compareable, key-, index- and generic pair
// interfaces to be able to stand in as return value for such expressions.
func NewNone() NoneVal { return func() {} }

func (n NoneVal) Step() Expression                     { return n }
func (n NoneVal) Next() Continuation                   { return n }
func (n NoneVal) Cons(...Expression) Sequential        { return n }
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
func (n NoneVal) End() bool                            { return true }
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
func (c Const) TypeIdent() TyComp             { return c().Type().TypeIdent() }
func (c Const) TypeReturn() TyComp            { return c().Type().TypeReturn() }
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
func (c Lambda) TypeIdent() TyComp     { return c().Type().TypeIdent() }
func (c Lambda) TypeReturn() TyComp    { return c().Type().TypeReturn() }
func (c Lambda) TypeArguments() TyComp { return c().Type().TypeArguments() }

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
	}
	// ‥.otherwise define by expressions identity entirely in terms of the
	// passed expression type
	return Def(expr.Type().TypeIdent(),
		expr.Type().TypeReturn(),
		expr.Type().TypeArguments())

}
func Define(
	expr Expression,
	types ...d.Typed,
) FuncDef {
	var (
		ct     = createFuncType(expr, types...)
		arglen = ct.TypeArguments().Len()
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
							ct.TypeIdent(),
							ct.TypeReturn(),
							Def(remains...))
					)
					// define new function from remaining
					// set of argument types, enclosing the
					// current arguments & appending its
					// own aruments to them, when called.
					return Define(Lambda(func(lateargs ...Expression) Expression {
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
					}), newpat.Types()...)

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
func (e FuncDef) TypeIdent() TyComp                  { return e.Type().TypeIdent() }
func (e FuncDef) TypeArguments() TyComp              { return e.Type().TypeArguments() }
func (e FuncDef) TypeReturn() TyComp                 { return e.Type().TypeReturn() }
func (e FuncDef) ArgCount() int                      { return e.Type().TypeArguments().Count() }
func (e FuncDef) String() string                     { return e().String() }
func (e FuncDef) Call(args ...Expression) Expression { return e(args...) }

//// TUPLE TYPE
///
// tuple type constructor expects a slice of field types and possibly a symbol
// type flag, to define the types name, otherwise 'tuple' is the type name and
// the sequence of field types is shown instead
func NewTuple(types ...d.Typed) TupDef {
	return func(args ...Expression) TupVal {
		var tup = make(TupVal, 0, len(args))
		if Def(types...).MatchArgs(args...) {
			for _, arg := range args {
				tup = append(tup, arg)
			}
		}
		return tup
	}
}

func (t TupDef) Call(args ...Expression) Expression { return t(args...) }
func (t TupDef) TypeFnc() TyFnc                     { return Tuple | Constructor }
func (t TupDef) String() string                     { return t.Type().String() }
func (t TupDef) Type() TyComp {
	var types = make([]d.Typed, 0, len(t()))
	for _, tup := range t() {
		types = append(types, tup.Type())
	}
	return Def(Tuple, Def(types...))
}

/// TUPLE VALUE
// tuple value is a slice of expressions, constructed by a tuple type
// constructor validated according to its type pattern.
func (t TupVal) Len() int { return len(t) }
func (t TupVal) String() string {
	var strs = make([]string, 0, t.Len())
	for _, val := range t {
		strs = append(strs, val.String())
	}
	return "[" + strings.Join(strs, ", ") + "]"
}
func (t TupVal) Get(idx int) Expression {
	if idx < t.Len() {
		return t[idx]
	}
	return NewNone()
}
func (t TupVal) TypeFnc() TyFnc                     { return Tuple }
func (t TupVal) Call(args ...Expression) Expression { return NewVector(append(t, args...)...) }
func (t TupVal) Type() TyComp {
	var types = make([]d.Typed, 0, len(t))
	for _, tup := range t {
		types = append(types, tup.Type())
	}
	return Def(Tuple, Def(types...))
}

//// RECORD TYPE
///
//
func NewRecord(types ...KeyPair) RecDef {
	return func(args ...Expression) RecVal {
		var rec = make(RecVal, 0, len(args))
		if len(args) > 0 {
			for n, arg := range args {
				if len(types) > n && arg.Type().Match(Key|Pair) {
					if kp, ok := arg.(KeyPair); ok {
						if strings.Compare(
							string(kp.KeyStr()),
							string(types[n].KeyStr()),
						) == 0 &&
							types[n].Value().Type().Match(
								kp.Value().Type(),
							) {
							rec = append(rec, kp)
						}
					}
				}
			}
		}
		return rec
	}
}

func (t RecDef) Call(args ...Expression) Expression { return t(args...) }
func (t RecDef) TypeFnc() TyFnc                     { return Record | Constructor }
func (t RecDef) Type() TyComp {
	var types = make([]d.Typed, 0, len(t()))
	for _, field := range t() {
		types = append(types, Def(
			DefSym(field.KeyStr()),
			Def(field.Value().Type()),
		))
	}
	return Def(Record, Def(types...))
}
func (t RecDef) String() string { return t.Type().String() }

/// RECORD VALUE
// tuple value is a slice of expressions, constructed by a tuple type
// constructor validated according to its type pattern.
func (t RecVal) TypeFnc() TyFnc { return Record }
func (t RecVal) Call(args ...Expression) Expression {
	var exprs = make([]Expression, 0, len(t)+len(args))
	for _, elem := range t {
		exprs = append(exprs, elem)
	}
	for _, arg := range args {
		if arg.Type().Match(Pair | Key) {
			if kp, ok := arg.(KeyPair); ok {
				exprs = append(exprs, kp)
			}
		}
	}
	return NewVector(exprs...)
}
func (t RecVal) Type() TyComp {
	var types = make([]d.Typed, 0, len(t))
	for _, tup := range t {
		types = append(types, tup.Type())
	}
	return Def(Record, Def(types...))
}
func (t RecVal) Len() int { return len(t) }
func (t RecVal) String() string {
	var strs = make([]string, 0, t.Len())
	for _, field := range t {
		strs = append(strs,
			`"`+field.Key().String()+`"`+" ∷ "+field.Value().String())
	}
	return "{" + strings.Join(strs, " ") + "}"
}
