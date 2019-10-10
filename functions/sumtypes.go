package functions

import (
	"strings"

	d "github.com/joergreinhardt/gatwd/data"
)

type (
	// GENERIC EXPRESSIONS
	NoneVal  func()
	ConstVal func() Expression
	FuncVal  func(...Expression) Expression

	//// DECLARED EXPRESSION
	ExprType func(...Expression) ExprVal
	ExprVal  Expression

	// TUPLE (TYPE[0]...TYPE[N])
	TupleType func(...Expression) TupleVal
	TupleVal  []Expression

	// RECORD (PAIR(KEY, VAL)[0]...PAIR(KEY, VAL)[N])
	RecordType func(...KeyPair) RecordVal
	RecordVal  []KeyPair
)

//// NONE VALUE CONSTRUCTOR
///
// none represens the abscence of a value of any type. implements countable,
// sliceable, consumeable, testable, compareable, key-, index- and generic pair
// interfaces to be able to stand in as return value for such expressions.
func NewNone() NoneVal { return func() {} }

func (n NoneVal) Head() Expression                  { return n }
func (n NoneVal) Tail() Sequential                  { return n }
func (n NoneVal) Cons(...Expression) Sequential     { return n }
func (n NoneVal) Append(...Expression) Sequential   { return n }
func (n NoneVal) Len() int                          { return 0 }
func (n NoneVal) Compare(...Expression) int         { return -1 }
func (n NoneVal) String() string                    { return "⊥" }
func (n NoneVal) Call(...Expression) Expression     { return nil }
func (n NoneVal) Key() Expression                   { return nil }
func (n NoneVal) Index() Expression                 { return nil }
func (n NoneVal) Left() Expression                  { return nil }
func (n NoneVal) Right() Expression                 { return nil }
func (n NoneVal) Both() Expression                  { return nil }
func (n NoneVal) Value() Expression                 { return nil }
func (n NoneVal) Empty() d.BoolVal                  { return true }
func (n NoneVal) Test(...Expression) bool           { return false }
func (n NoneVal) TypeFnc() TyFnc                    { return None }
func (n NoneVal) TypeNat() d.TyNat                  { return d.Nil }
func (n NoneVal) Type() TyPattern                   { return Def(None) }
func (n NoneVal) TypeElem() TyPattern               { return Def(None) }
func (n NoneVal) TypeName() string                  { return n.String() }
func (n NoneVal) Slice() []Expression               { return []Expression{} }
func (n NoneVal) Flag() d.BitFlag                   { return d.BitFlag(None) }
func (n NoneVal) FlagType() d.Uint8Val              { return Flag_Function.U() }
func (n NoneVal) Consume() (Expression, Sequential) { return NewNone(), NewNone() }

//// CONSTANT DECLARATION
///
// declares a constant value
func NewConstant(constant func() Expression) ConstVal { return constant }

func (c ConstVal) Type() TyPattern {
	return Def(Constant|c().Type().TypeFnc(), c().Type())
}
func (c ConstVal) TypeFnc() TyFnc                { return Constant }
func (c ConstVal) String() string                { return c().String() }
func (c ConstVal) Call(...Expression) Expression { return c() }

//// FUNCTION DECLARATION
///
// declares a constant value
func NewFunction(fnc func(...Expression) Expression) FuncVal { return fnc }

func (c FuncVal) Type() TyPattern {
	return Def(Value|c().Type().TypeFnc(), c().Type())
}
func (c FuncVal) TypeFnc() TyFnc                { return Value }
func (c FuncVal) String() string                { return c().String() }
func (c FuncVal) Call(...Expression) Expression { return c() }

/// PARTIAL APPLYABLE EXPRESSION VALUE
//
// element values yield a subelements of optional, tuple, or enumerable
// expressions with sub-type pattern as second return value
func tTp(typ []d.Typed) []TyPattern {
	var pat = make([]TyPattern, 0, len(typ))
	for _, t := range typ {
		if Flag_Pattern.Match(t.FlagType()) {
			pat = append(pat, t.(TyPattern))
			continue
		}
		pat = append(pat, Def(t))
	}
	return pat
}
func pTt(pat []TyPattern) []d.Typed {
	var typ = make([]d.Typed, 0, len(pat))
	for _, p := range pat {
		typ = append(typ, p)
	}
	return typ
}
func Define(
	expr Expression,
	argtype, retype TyPattern,
	propertys ...d.Typed,
) ExprType {

	if !Flag_Pattern.Match(argtype.FlagType()) {
		argtype = Def(argtype)
	}

	var (
		ident         d.Typed
		pattern       TyPattern
		props, idents []d.Typed
		arglen        = argtype.Len()
		argtypes      = argtype.Pattern()
	)

	// if no propertys are passed, function type is the types ident
	if len(propertys) == 0 {
		ident = expr.TypeFnc()
	} else {
		// scan passed propertys for symbols
		props, idents = []d.Typed{}, []d.Typed{}
		for _, typ := range propertys {
			if Flag_Symbol.Match(typ.FlagType()) {
				idents = append(idents, typ)
				continue
			}
			props = append(props, typ)
		}
		// define type identity from fetched symbol(s)
		ident = Def(idents...)
	}

	if len(props) > 0 {
		pattern = Def(argtype, ident, retype, Def(props...))
	} else {
		pattern = Def(argtype, ident, retype)
	}

	return func(args ...Expression) ExprVal {
		var length = len(args)
		if length > 0 {
			if pattern.TypeArguments().MatchArgs(args...) {
				switch {
				case length == arglen:
					return expr.Call(args...)

				case length < arglen:
					var remains = argtypes[length:]
					return Define(ExprType(
						func(lateargs ...Expression) ExprVal {
							if len(lateargs) > 0 {
								return expr.Call(append(
									args, lateargs...,
								)...)
							}
							return Def(Def(pTt(remains)...), ident, retype)
						}), Def(pTt(remains)...), Def(ident), retype)

				case length > arglen:
					var vector = NewVector()
					for len(args) > arglen {
						vector = vector.Cons(
							expr.Call(args[:arglen]...)).(VecVal)
						args = args[arglen:]
					}
					if length > 0 {
						vector = vector.Cons(Define(
							expr, argtype, retype, propertys...,
						).Call(args...)).(VecVal)
					}
					return vector
				}
			}
			return None
		}
		return pattern
	}
}
func (e ExprType) TypeFnc() TyFnc                     { return Constructor | Value }
func (e ExprType) Type() TyPattern                    { return e().Call().(TyPattern) }
func (e ExprType) ArgCount() int                      { return e.Type().TypeArguments().Count() }
func (e ExprType) String() string                     { return e().String() }
func (e ExprType) Call(args ...Expression) Expression { return e(args...) }

//// TUPLE TYPE
///
// tuple type constructor expects a slice of field types and possibly a symbol
// type flag, to define the types name, otherwise 'tuple' is the type name and
// the sequence of field types is shown instead
func NewTuple(types ...d.Typed) ExprType {
	var (
		pattern = make([]Expression, 0, len(types))
		symbol  d.Typed
	)
	if len(types) > 1 {
		if Flag_Symbol.Match(types[0].FlagType()) {
			symbol = types[0]
			types = types[1:]
		}
	}
	if symbol == nil {
		symbol = Tuple
	}
	for _, typ := range types {
		if Flag_Pattern.Match(typ.FlagType()) {
			pattern = append(pattern, typ.(TyPattern))
		} else {
			pattern = append(pattern, Def(typ))
		}
	}
	return Define(
		TupleType(func(args ...Expression) TupleVal {
			if len(args) > 0 {
				return args
			}
			return pattern
		}),
		Def(types...), Def(symbol, Def(types...)))
}
func (t TupleType) Call(args ...Expression) Expression { return t(args...) }
func (t TupleType) TypeFnc() TyFnc                     { return Tuple | Constructor }
func (t TupleType) String() string                     { return t.Type().String() }
func (t TupleType) Type() TyPattern {
	var (
		elems = t()
		count = len(elems)
		types = make([]d.Typed, 0, count)
	)
	for _, elem := range elems {
		types = append(types, elem.Type())
	}
	return Def(Tuple, Def(types...))
}

/// TUPLE VALUE
//
// tuple value is a slice of expressions, constructed by a tuple type
// constructor validated according to its type pattern.
func (t TupleVal) Call(...Expression) Expression { return t }
func (t TupleVal) Count() int                    { return len(t) }
func (t TupleVal) TypeFnc() TyFnc                { return Tuple }
func (t TupleVal) Type() TyPattern {
	var types = make([]d.Typed, 0, t.Count())
	for _, elem := range t {
		types = append(types, elem.Type())
	}
	return Def(Tuple, Def(types...))
}
func (t TupleVal) Get(idx int) Expression {
	if idx < t.Count() {
		return t[idx]
	}
	return NewNone()
}
func (t TupleVal) String() string {
	var strs = make([]string, 0, t.Count())
	for _, elem := range t {
		strs = append(strs, elem.String())
	}
	return "[" + strings.Join(strs, ", ") + "]"
}

//// RECORD TYPE
func NewRecordType(types ...KeyPair) RecordType {
	return func(args ...KeyPair) RecordVal {
		if len(args) > 0 {
			for idx, arg := range args {
				if idx < len(types) {
					if types[idx].Key() == arg.Key() {
						if !types[idx].Value().Type().Match(arg.Type()) {
							return []KeyPair{}
						}
					}
				}
			}
			return args
		}
		return types
	}
}

func (r RecordType) TypeFnc() TyFnc  { return Record }
func (r RecordType) Type() TyPattern { return Def(Record, Def(r.Types()...)) }
func (r RecordType) String() string {
	var strs = make([]string, 0, len(r()))
	for _, record := range r() {
		strs = append(
			strs,
			record.Key().String()+"∷ "+
				record.Value().Type().TypeName(),
		)
	}
	return "{" + strings.Join(strs, ", ") + "}"
}
func (r RecordType) Types() []d.Typed {
	var (
		pairs = r()
		types = make([]d.Typed, 0, len(pairs))
	)
	for _, pair := range pairs {
		types = append(types, pair.Value().Type())
	}
	return types
}
func (r RecordType) Call(args ...Expression) Expression {
	var pairs = make([]KeyPair, 0, len(args))
	for _, arg := range args {
		if arg.TypeFnc().Match(Key | Pair) {
			pairs = append(pairs, arg.(KeyPair))
		}
	}
	return r(pairs...)
}

//// RECORD VALUE
///
//
func (r RecordVal) Count() int      { return len(r) }
func (r RecordVal) Type() TyPattern { return Def(Record, Def(r.Types()...)) }
func (r RecordVal) TypeFnc() TyFnc  { return Record }
func (r RecordVal) Keys() []string {
	var keys = make([]string, 0, len(r))
	for _, record := range r {
		keys = append(keys, record.Key().String())
	}
	return keys
}
func (r RecordVal) Types() []d.Typed {
	var types = make([]d.Typed, 0, len(r))
	for _, pair := range r {
		types = append(types, pair.Value().Type())
	}
	return types
}
func (r RecordVal) Call(args ...Expression) Expression {
	if len(args) > 0 {
		var result = make([]KeyPair, 0, len(r))
		for _, pair := range r {
			result = append(result,
				NewKeyPair(
					pair.Key().String(),
					pair.Value().Call(args...)))
		}
	}
	return r
}
func (r RecordVal) Get(key string) Expression {
	for _, record := range r {
		if record.Key().String() == key {
			return record
		}
	}
	return NewNone()
}
func (r RecordVal) String() string {
	var strs = make([]string, 0, len(r))
	for _, record := range r {
		strs = append(
			strs,
			record.Key().String()+"∷ "+
				record.Value().String(),
		)
	}
	return "{" + strings.Join(strs, ", ") + "}"
}
