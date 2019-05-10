/*
  FUNCTIONAL CONTAINERS

  containers implement enumeration of functional types, aka lists, vectors sets, pairs, tuples‥.
*/
package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

type (
	//// NONE
	NoneVal   func()
	JustVal   func() Callable
	MaybeVal  func() Callable
	MaybeCons func(Callable) MaybeVal
	CaseExpr  func(Callable) (Callable, bool)

	//// DATA
	NativeVal func(args ...interface{}) d.Native
	DataVal   func(args ...d.Native) d.Native

	//// EXPRESSION
	ConstantExpr func() Callable
	UnaryExpr    func(Callable) Callable
	BinaryExpr   func(a, b Callable) Callable
	NaryExpr     func(...Callable) Callable
)

// reverse arguments
func RevArgs(args ...Callable) []Callable {
	var rev = []Callable{}
	for i := len(args) - 1; i > 0; i-- {
		rev = append(rev, args[i])
	}
	return rev
}

// convert native to functional values
func NatToFnc(args ...d.Native) []Callable {
	var result = []Callable{}
	for _, arg := range args {
		result = append(result, NewFromData(arg))
	}
	return result
}

// convert functional to native values
func FncToNat(args ...Callable) []d.Native {
	var result = []d.Native{}
	for _, arg := range args {
		result = append(result, arg.Eval())
	}
	return result
}

// assumes the arguments to either implement paired, or be alternating pairs of
// keys & values. in case the number of passed arguments that are not pairs is
// uneven, last field will be filled with a value of type none
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

//// JUST VALUE
func NewJust(arg Callable) JustVal {
	return JustVal(func() Callable { return arg })
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
func NewMaybeVal(expr func() Callable) MaybeVal     { return expr }
func (m MaybeVal) String() string                   { return "Maybe " + m().String() }
func (m MaybeVal) TypeFnc() TyFnc                   { return Maybe }
func (m MaybeVal) TypeNat() d.TyNat                 { return d.Expression }
func (m MaybeVal) Consume() (Callable, Consumeable) { return m(), m }
func (m MaybeVal) Head() Callable                   { return m() }
func (m MaybeVal) Tail() Consumeable                { return m }
func (m MaybeVal) Call(args ...Callable) Callable {
	if len(args) > 0 {
		return m().Call(args...)
	}
	return m()
}
func (m MaybeVal) Eval(args ...d.Native) d.Native {
	return m().Eval(args...)
}

//// MAYBE CONSTRUCTOR
func NewMaybeConstructor(test CaseExpr) MaybeCons {
	return MaybeCons(func(value Callable) MaybeVal {
		if val, ok := test(value); ok {
			return MaybeVal(func() Callable { return NewJust(val) })
		}
		return MaybeVal(func() Callable { return NewNone() })
	})
}
func (c MaybeCons) String() string   { return "Maybe" }
func (c MaybeCons) TypeFnc() TyFnc   { return Maybe }
func (c MaybeCons) TypeNat() d.TyNat { return d.Expression }
func (c MaybeCons) Call(args ...Callable) Callable {
	if len(args) > 0 {
		return c(args[0])
	}
	return NewNone()
}

//// CASE EXPRESSION
func NewCaseExpr(expr func(arg Callable) (Callable, bool)) CaseExpr {
	return CaseExpr(expr)
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

//// STATIC EXPRESSIONS
///
// generic functional enclosures to functionalize every function that happens
// to implement the correct signature
// CONSTANT EXPRESSION
func NewConstant(
	fnc func() Callable,
) ConstantExpr {
	return fnc
}
func (c ConstantExpr) Ident() Callable           { return c() }
func (c ConstantExpr) TypeFnc() TyFnc            { return Expression }
func (c ConstantExpr) TypeNat() d.TyNat          { return c().TypeNat() }
func (c ConstantExpr) Call(...Callable) Callable { return c() }
func (c ConstantExpr) Eval(...d.Native) d.Native { return c().Eval() }

/// UNARY EXPRESSION
func NewUnaryExpr(
	fnc func(Callable) Callable,
) UnaryExpr {
	return fnc
}
func (u UnaryExpr) Ident() Callable               { return u }
func (u UnaryExpr) TypeFnc() TyFnc                { return Expression }
func (u UnaryExpr) TypeNat() d.TyNat              { return d.Expression.TypeNat() }
func (u UnaryExpr) Call(arg ...Callable) Callable { return u(arg[0]) }
func (u UnaryExpr) Eval(arg ...d.Native) d.Native { return u(NewFromData(arg...)) }

/// BINARY EXPRESSION
func NewBinaryExpr(
	fnc func(l, r Callable) Callable,
) BinaryExpr {
	return fnc
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
	fnc func(...Callable) Callable,
) NaryExpr {
	return fnc
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

//// DATA
///
// native val encloses golang literal values to implement the callable
// interface
func NewLiteral(init ...interface{}) NativeVal {
	return func(args ...interface{}) d.Native {
		var natives = []d.Native{}
		if len(args) > 0 {
			for _, arg := range args {
				natives = append(natives, d.New(arg))
			}
			return NewLiteral(init...).Eval(natives...)
		}
		if len(init) > 0 {
			if len(init) > 1 {
				return d.New(init...)
			}
			return d.New(init[0])
		}
		return d.NilVal{}
	}
}
func (n NativeVal) String() string   { return n().String() }
func (n NativeVal) TypeNat() d.TyNat { return n().TypeNat() }
func (n NativeVal) TypeFnc() TyFnc   { return Native }
func (n NativeVal) Call(args ...Callable) Callable {
	return NewFromData(n()).Call(args...)
}
func (n NativeVal) Eval(args ...d.Native) d.Native {
	return n().Eval(args...)
}

// data value is a callable implementation of an enclosure for values
// implementing data/Native
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
