/*
  FUNCTIONAL CONTAINERS

  containers implement enumeration of functional types, aka lists, vectors sets, pairs, tuples‥.
*/
package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

type (
	//// PREDICATE | CASE | CASE-SWITCH
	PredictExpr    func(...Callable) bool
	CaseExpr       func(...Callable) (Callable, bool)
	CaseSwitchExpr func(...Callable) (Callable, bool, Consumeable)

	//// MAYBE | JUST | NONE
	NoneVal       func()
	JustVal       func(...Callable) Callable
	MaybeVal      func(...Callable) Callable
	MaybeType     func(Callable) MaybeVal
	MaybeTypeCons func(...Callable) MaybeType

	//// TUPLE
	TupleElem     func(...Callable) (Callable, int)
	TupleValue    func(...Callable) TupleElem
	TupleType     func(...Callable) []TupleValue
	TupleTypeCons func(...Callable) TupleType

	//// STATIC EXPRESSIONS
	ConstantExpr func() Callable
	UnaryExpr    func(Callable) Callable
	BinaryExpr   func(a, b Callable) Callable
	NaryExpr     func(...Callable) Callable

	//// DATA VALUE
	DataVal func(args ...d.Native) d.Native
)

//// TUPLE
///
//
func TupleTypeConstructor(elems ...Callable) TupleType {
	return func(iniargs ...Callable) []TupleValue {
		var tuple = []TupleValue{}
		for i, ini := range iniargs {
			tuple = append(
				tuple,
				TupleValue(
					func(args ...Callable) TupleElem {
						return func(args ...Callable) (Callable, int) {
							return ini, i
						}
					}))
		}
		return tuple
	}
}

//// HELPER FUNCTIONS TO HANDLE ARGUMENTS
///
// since every callable also needs to implement the eval interface and data as
// such allways boils down to native values, conversion between callable-/ &
// native arguments is frequently needed. arguments may also need to be
// reversed when intendet to be passed to certain recursive expressions, or
// returned by those
//
/// REVERSE ARGUMENTS
func RevArgs(args ...Callable) []Callable {
	var rev = []Callable{}
	for i := len(args) - 1; i > 0; i-- {
		rev = append(rev, args[i])
	}
	return rev
}

/// CONVERT NATIVE TO FUNCTIONAL
func NatToFnc(args ...d.Native) []Callable {
	var result = []Callable{}
	for _, arg := range args {
		result = append(result, NewFromData(arg))
	}
	return result
}

/// CONVERT FUNCTIONAL TO NATIVE
func FncToNat(args ...Callable) []d.Native {
	var result = []d.Native{}
	for _, arg := range args {
		result = append(result, arg.Eval())
	}
	return result
}

/// GROUP ARGUMENTS PAIRWISE
//
// assumes the arguments to either implement paired, or be alternating pairs of
// key & value. in case the number of passed arguments that are not pairs is
// uneven, last field will be filled up with a value of type none
func ArgsToPaired(args ...Callable) []Paired {
	var pairs = []Paired{}
	var alen = len(args)
	for i, arg := range args {
		if arg.TypeFnc().Match(Pair) {
			pairs = append(pairs, arg.(Paired))
		}
		if i < alen-2 {
			i = i + 1
			pairs = append(pairs, NewPair(arg, args[i]))
		}
		pairs = append(pairs, NewPair(arg, NewNone()))
	}
	return pairs
}

//// CASE EXPRESSION
///
// case evaluates first argument by applying it to the predicate and either
// returns the argument, if predicate yields true, a none instance and false if
// it's not. if more than one argument is given, additional arguments will be
// evaluated recursively.
func NewCaseExpr(expr Callable, pred PredictExpr) CaseExpr {
	return func(args ...Callable) (Callable, bool) {
		var arg Callable
		if len(args) > 0 {
			arg = args[0]
			if pred(arg) {
				return arg, true
			}
			if len(args) > 1 {
				args = args[1:]
				return NewCaseExpr(expr, pred)(args...)
			}
		}
		return NewNone(), false
	}
}
func (c CaseExpr) Ident() Callable  { return c }
func (c CaseExpr) String() string   { return "Case" }
func (c CaseExpr) TypeFnc() TyFnc   { return Case }
func (c CaseExpr) TypeNat() d.TyNat { return d.Expression }
func (c CaseExpr) Call(args ...Callable) Callable {
	var val Callable
	var ok bool
	if len(args) > 0 {
		val, ok = c(args[0])
		if len(args) > 1 {
			val = val.Call(args[1:]...)
		}
	}
	if ok {
		return val
	}
	return NewNone()
}

func (c CaseExpr) Eval(args ...d.Native) d.Native {
	var val d.Native
	var ok bool
	if len(args) > 0 {
		val, ok = c(NewFromData(args[0]))
		if len(args) > 1 {
			val = val.Eval(args[1:]...)
		}
	}
	if ok {
		return val.Eval()
	}
	return d.NilVal{}
}

//// CASE SWITCH
// takes first argument to apply to case. if first argument is the only passed
// argument, it will be reused and applyed to all cases until one matches, or
// cases are depleted.
func NewCaseSwitch(cases ...CaseExpr) CaseSwitchExpr {

	// vectorive cases to be consumeable
	var cas CaseExpr
	var vec VecVal
	if len(cases) > 0 {
		cas = cases[0]
		if len(cases) > 1 {
			cases = cases[1:]
		}
		var args []Callable
		for _, arg := range cases {
			args = append(args, arg)
		}
		vec = NewVector(args...)
	}

	// case switch encloses and consumes passed cases & applys them
	// recursively to the passed argument(s) to return the resulting value,
	// or argument, depending on the boolean indicator and a consumeable
	// containing the remaining cases.
	return func(args ...Callable) (Callable, bool, Consumeable) {
		if len(args) > 0 {
			var val, ok = cas(args...)
			if ok {
				return val, ok, vec
			}
		}
		return NewNone(), false, NewList()
	}
}
func (s CaseSwitchExpr) String() string   { return "CaseSwitch" }
func (s CaseSwitchExpr) TypeFnc() TyFnc   { return CaseSwitch }
func (s CaseSwitchExpr) TypeNat() d.TyNat { return d.Expression }
func (s CaseSwitchExpr) Call(args ...Callable) Callable {
	var val, ok, _ = s(args...)
	if ok {
		return val
	}
	return NewNone()
}
func (s CaseSwitchExpr) Eval(args ...d.Native) d.Native {
	var val, ok, _ = s(NatToFnc(args...)...)
	if ok {
		return val.Eval()
	}
	return d.NilVal{}
}

//// NONE VALUE
func NewNone() NoneVal                             { return func() {} }
func (n NoneVal) Ident() Callable                  { return n }
func (n NoneVal) Len() int                         { return 0 }
func (n NoneVal) String() string                   { return "⊥" }
func (n NoneVal) Eval(...d.Native) d.Native        { return nil }
func (n NoneVal) Value() Callable                  { return nil }
func (n NoneVal) Call(...Callable) Callable        { return nil }
func (n NoneVal) Empty() bool                      { return true }
func (n NoneVal) TypeFnc() TyFnc                   { return None }
func (n NoneVal) TypeNat() d.TyNat                 { return d.Nil }
func (n NoneVal) TypeName() string                 { return n.String() }
func (n NoneVal) Head() Callable                   { return NewNone() }
func (n NoneVal) Tail() Consumeable                { return NewNone() }
func (n NoneVal) Consume() (Callable, Consumeable) { return NewNone(), NewNone() }

//// PREDICATE
func NewPredicate(pred func(...Callable) bool) PredictExpr { return pred }
func (p PredictExpr) String() string                       { return "Predicate" }
func (p PredictExpr) TypeNat() d.TyNat                     { return d.Expression }
func (p PredictExpr) TypeFnc() TyFnc                       { return Predicate }
func (p PredictExpr) Eval(args ...d.Native) d.Native {
	if len(args) > 0 {
		return p.Call(NewFromData(args[0]))
	}
	return d.NilVal{}
}
func (p PredictExpr) Call(args ...Callable) Callable {
	if len(args) > 0 {
		return p.Call(args[0])
	}
	return NewNone()
}

//// JUST VALUE
///
// new maybe-type constructor returns a constructor of values of a distinct
// maybe type, as defined by the predicate passed to it and thereby effectively
// declares a new higher order type at runtime.
//
// apart from the predicate, a type signature can be passed, to be returned by
// the defined maybe data constructor, when called without arguments, to be
// returned by typeFnc, typeNat, string‥.
func NewJust(val Callable) JustVal {
	var just JustVal
	just = JustVal(
		func(args ...Callable) Callable {
			if len(args) > 0 {
				return val.Call(args...)
			}
			return val
		})
	return just
}
func (n JustVal) Ident() Callable   { return n }
func (n JustVal) Value() Callable   { return n() }
func (n JustVal) Head() Callable    { return n() }
func (n JustVal) Tail() Consumeable { return n }
func (n JustVal) Consume() (Callable, Consumeable) {
	return n(), NewNone()
}
func (n JustVal) String() string {
	return "Just·" + n().TypeNat().String() + " " + n().String()
}
func (n JustVal) Call(args ...Callable) Callable {
	return n().Call(args...)
}
func (n JustVal) Eval(args ...d.Native) d.Native {
	return n().Eval(args...)
}
func (n JustVal) Empty() bool {
	if n() != nil {
		if n().TypeFnc().Match(None) ||
			n().TypeNat().Match(d.Nil) {
			return false
		}
	}
	return true
}
func (n JustVal) TypeFnc() TyFnc {
	return Just | n().TypeFnc()
}
func (n JustVal) TypeNat() d.TyNat {
	return n().TypeNat()
}
func (n JustVal) TypeName() string {
	return "JustVal·" + n().TypeFnc().String()
}

//// MAYBE VALUE
func (m MaybeVal) String() string                 { return m().String() }
func (m MaybeVal) TypeFnc() TyFnc                 { return m().TypeFnc() }
func (m MaybeVal) TypeNat() d.TyNat               { return m().TypeNat() }
func (m MaybeVal) Call(args ...Callable) Callable { return m().Call(args...) }
func (m MaybeVal) Eval(args ...d.Native) d.Native { return m().Eval(args...) }
func NewMaybeValue(iniargs ...Callable) MaybeVal {
	return func(args ...Callable) Callable {
		if len(iniargs) > 0 {
			if len(iniargs) > 1 {
				return NewJust(CurryN(iniargs...))
			}
			return NewJust(iniargs[0])
		}
		return NewNone()
	}
}

//// MAYBE TYPE CONSTRUCTOR
func NewMaybeTypeConstructor(pred PredictExpr) MaybeType {
	return func(arg Callable) MaybeVal {
		if pred(arg) {
			return NewMaybeValue(MaybeVal(
				func(args ...Callable) Callable {
					if len(args) > 0 {
						return arg.Call(args...)
					}
					return arg
				}))
		}
		return NewMaybeValue(NewNone())
	}
}

func (c MaybeTypeCons) String() string   { return "Maybe Type Constructor" }
func (c MaybeTypeCons) TypeFnc() TyFnc   { return Constructor | Maybe }
func (c MaybeTypeCons) TypeNat() d.TyNat { return d.Expression }
func (c MaybeTypeCons) Call(args ...Callable) Callable {
	if len(args) > 0 {
		if args[0].TypeFnc().Match(Predicate) {
			return c(args[0])
		}
	}
	return NewNone()
}

func (m MaybeType) String() string   { return "Maybe Type" }
func (m MaybeType) TypeFnc() TyFnc   { return Maybe }
func (m MaybeType) TypeNat() d.TyNat { return d.Expression }
func (m MaybeType) Call(args ...Callable) Callable {
	if len(args) > 0 {
		if len(args) > 1 {
			var vals = []Callable{}
			for _, arg := range args {
				vals = append(
					vals,
					m(arg),
				)
			}
		}
		var arg = args[0]
		return m(arg)
	}
	return NewNone()
}
func (m MaybeType) Eval(args ...d.Native) d.Native {
	if len(args) > 0 {
		if len(args) > 1 {
			var vals = []Callable{}
			for _, arg := range args {
				vals = append(
					vals,
					m(DataVal(arg.Eval)),
				)
			}
		}
		var arg = args[0]
		return m(DataVal(arg.Eval))
	}
	return NewNone()
}

//// STATIC FUNCTION EXPRESSIONS OF PREDETERMINED ARITY
///
// used to guard expression arity, or whenever a type is needed to have a non
// variadic argument signature.
//
/// CONSTANT EXPRESSION
func NewConstant(
	expr Callable,
) ConstantExpr {
	return func() Callable { return expr }
}
func (c ConstantExpr) Ident() Callable           { return c() }
func (c ConstantExpr) TypeFnc() TyFnc            { return Expression }
func (c ConstantExpr) TypeNat() d.TyNat          { return c().TypeNat() }
func (c ConstantExpr) Call(...Callable) Callable { return c() }
func (c ConstantExpr) Eval(...d.Native) d.Native { return c().Eval() }

/// UNARY EXPRESSION
func NewUnaryExpr(
	expr Callable,
) UnaryExpr {
	return func(arg Callable) Callable { return expr.Call(arg) }
}
func (u UnaryExpr) Ident() Callable               { return u }
func (u UnaryExpr) TypeFnc() TyFnc                { return Expression }
func (u UnaryExpr) TypeNat() d.TyNat              { return d.Expression.TypeNat() }
func (u UnaryExpr) Call(arg ...Callable) Callable { return u(arg[0]) }
func (u UnaryExpr) Eval(arg ...d.Native) d.Native { return u(NewFromData(arg...)) }

/// BINARY EXPRESSION
func NewBinaryExpr(
	expr Callable,
) BinaryExpr {
	return func(a, b Callable) Callable { return expr.Call(a, b) }
}

func (b BinaryExpr) Ident() Callable                { return b }
func (b BinaryExpr) TypeFnc() TyFnc                 { return Expression }
func (b BinaryExpr) TypeNat() d.TyNat               { return d.Expression.TypeNat() }
func (b BinaryExpr) Call(args ...Callable) Callable { return b(args[0], args[1]) }
func (b BinaryExpr) Eval(args ...d.Native) d.Native {
	return b(NewFromData(args[0]), NewFromData(args[1]))
}

/// NARY EXPRESSION
func NewNaryExpr(
	expr Callable,
) NaryExpr {
	return func(args ...Callable) Callable {
		return expr.Call(args...)
	}
}
func (n NaryExpr) Ident() Callable             { return n }
func (n NaryExpr) TypeFnc() TyFnc              { return Expression }
func (n NaryExpr) TypeNat() d.TyNat            { return d.Expression.TypeNat() }
func (n NaryExpr) Call(d ...Callable) Callable { return n(d...) }
func (n NaryExpr) Eval(args ...d.Native) d.Native {
	var params = []Callable{}
	for _, arg := range args {
		params = append(params, NewFromData(arg))
	}
	return n(params...)
}

//// DATA VALUE
///
// data value implements the callable interface but returns an instance of
// data/Value. the eval method of every native can be passed as argument
// instead of the value itself, as in 'DataVal(native.Eval)', to delay, or even
// possibly ommit evaluation of the underlying data value for cases where
// lazynes is paramount.
func New(inf ...interface{}) Callable { return NewFromData(d.New(inf...)) }

func NewDataVal() DataVal {
	var value = d.NilVal{}
	return DataVal(func(args ...d.Native) d.Native {
		if len(args) > 1 {
			return d.NewSlice(args...)
		}
		if len(args) > 0 {
			return args[0]
		}
		return value
	})
}

func NewFromData(data ...d.Native) DataVal {
	var eval func(...d.Native) d.Native
	for _, val := range data {
		eval = val.Eval
	}
	return func(args ...d.Native) d.Native { return eval(args...) }
}

func (n DataVal) Eval(args ...d.Native) d.Native {
	if len(args) > 0 {
		if len(args) > 1 {
			return n().Eval(args...)
		}
		return n().Eval(args[0])
	}
	return n().Eval()
}

func (n DataVal) Call(vals ...Callable) Callable {
	if len(vals) > 0 {
		if len(vals) > 1 {
			return NewFromData(n(FncToNat(vals...)...))
		}
		return NewFromData(n.Eval(vals[0].Eval()))
	}
	return NewFromData(n.Eval())
}

func (n DataVal) TypeFnc() TyFnc   { return Data }
func (n DataVal) TypeNat() d.TyNat { return n().TypeNat() }
func (n DataVal) String() string   { return n().String() }
