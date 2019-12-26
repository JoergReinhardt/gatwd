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
	TupCon func(...Expression) TupVal
	TupVal []Expression

	// RECORD (PAIR(KEY, VAL)[0]...PAIR(KEY, VAL)[N])
	RecCon func(...Expression) RecVal
	RecVal []KeyPair

	//// ENUMERABLE
	EnumDef func(d.Numeral) EnumVal
	EnumVal func(...Expression) (Expression, d.Numeral, EnumDef)
)

//// NONE VALUE CONSTRUCTOR
///
// none represens the abscence of a value of any type. implements countable,
// sliceable, consumeable, testable, compareable, key-, index- and generic pair
// interfaces to be able to stand in as return value for such expressions.
func NewNone() NoneVal { return func() {} }

func (n NoneVal) Current() Expression                  { return n }
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
			return Def(append([]d.Typed{expr.Type().TypeId()}, types...)...)
		}
	}
	// ‥.otherwise define by expressions identity entirely in terms of the
	// passed expression type
	return Def(expr.Type().TypeId(),
		expr.Type().TypeRet(),
		expr.Type().TypeArgs())

}
func Define(
	expr Expression,
	types ...d.Typed,
) FuncDef {
	var (
		ct     = createFuncType(expr, types...)
		arglen = ct.TypeArgs().Len()
	)
	// return partialy applicable function
	return func(args ...Expression) Expression {
		var length = len(args)
		if length > 0 {
			if ct.TypeArgs().MatchArgs(args...) {
				switch {
				// NUMBER OF PASSED ARGUMENTS MATCHES EXACTLY →
				case length == arglen:
					return expr.Call(args...)

				// NUMBER OF PASSED ARGUMENTS IS INSUFFICIENT →
				case length < arglen:
					// safe types of arguments remaining to be filled
					var (
						remains = ct.TypeArgs().Types()[length:]
						newpat  = Def(
							ct.TypeId(),
							ct.TypeRet(),
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
func (e FuncDef) TypeId() TyComp                     { return e.Type().TypeId() }
func (e FuncDef) TypeArgs() TyComp                   { return e.Type().TypeArgs() }
func (e FuncDef) TypeRet() TyComp                    { return e.Type().TypeRet() }
func (e FuncDef) ArgCount() int                      { return e.Type().TypeArgs().Count() }
func (e FuncDef) String() string                     { return e().String() }
func (e FuncDef) Call(args ...Expression) Expression { return e(args...) }

//// TUPLE TYPE
///
// tuple type constructor expects a slice of field types and possibly a symbol
// type flag, to define the types name, otherwise 'tuple' is the type name and
// the sequence of field types is shown instead
func NewTupleType(types ...d.Typed) TupCon {
	return func(args ...Expression) TupVal {
		var tup = make(TupVal, 0, len(args))
		if Def(types...).MatchArgs(args...) {
			for _, arg := range args {
				tup = append(tup, arg)
			}
		}
		if len(tup) == 0 {
			for _, t := range types {
				if Kind_Comp.Match(t.Kind()) {
					tup = append(tup, t.(TyComp))
				}
				tup = append(tup, Def(t))
			}
		}
		return tup
	}
}

func (t TupCon) Call(args ...Expression) Expression { return t(args...) }
func (t TupCon) TypeFnc() TyFnc                     { return Tuple | Constructor }
func (t TupCon) String() string                     { return t.Type().String() }
func (t TupCon) Type() TyComp {
	var types = make([]d.Typed, 0, len(t()))
	for _, c := range t() {
		if Kind_Comp.Match(c.Type().Kind()) {
			types = append(types, c.(TyComp))
			continue
		}
		types = append(types, c.Type())
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
func NewRecordType(fields ...KeyPair) RecCon {
	return func(args ...Expression) RecVal {
		var rec = make(RecVal, 0, len(args))
		if len(args) > 0 {
			for n, arg := range args {
				if len(fields) > n && arg.Type().Match(Key|Pair) {
					if kp, ok := arg.(KeyPair); ok {
						if strings.Compare(
							string(kp.KeyStr()),
							string(fields[n].KeyStr()),
						) == 0 &&
							fields[n].Value().Type().Match(
								kp.Value().Type(),
							) {
							rec = append(rec, kp)
						}
					}
				}
			}
		}
		if len(rec) == 0 {
			return fields
		}
		return rec
	}
}

func (t RecCon) Call(args ...Expression) Expression { return t(args...) }
func (t RecCon) TypeFnc() TyFnc                     { return Record | Constructor }
func (t RecCon) Type() TyComp {
	var types = make([]d.Typed, 0, len(t()))
	for _, field := range t() {
		types = append(types, Def(
			DefSym(field.KeyStr()),
			Def(field.Value().Type()),
		))
	}
	return Def(Record, Def(types...))
}
func (t RecCon) String() string { return t.Type().String() }

/// RECORD VALUE
// tuple value is a slice of expressions, constructed by a tuple type
// constructor validated according to its type pattern.
func (t RecVal) TypeFnc() TyFnc { return Record }
func (t RecVal) Call(args ...Expression) Expression {
	var fields = make([]Expression, 0, len(t)+len(args))
	for _, field := range t {
		fields = append(fields, field)
	}
	for _, arg := range args {
		if arg.Type().Match(Pair | Key) {
			if kp, ok := arg.(KeyPair); ok {
				fields = append(fields, kp)
			}
		}
	}
	return NewVector(fields...)
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
