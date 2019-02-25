package functions

import (
	"fmt"

	d "github.com/JoergReinhardt/gatwd/data"
)

type (
	//// FUNCTOR CONSTRUCTORS
	///
	// CONDITIONAL, CASE & OPTIONAL
	OptionalFnc func(...Functional) Optional
	CaseFnc     func(...Functional) Optional
	PredFnc     func(...Functional) bool

	//// OPTIONAL TYPE DATA CONSTRUCTORS
	CaseExpr func(scrut ...Functional) Optional // Case | Just | None
	MaybeFnc func() Optional                    // Just | None

	//// DATA CONSTRUCTORS FOR OPTION TYPE PARTIALS
	///
	// TRUE/FALSE
	TrueVal  func() Boolean // boolean constants true & false
	FalseVal func() Boolean // implementing 'Boolen :: Bool() bool'

	// EITHER/OR
	OrVal     func() Functional // implementing 'Option :: Maybe() bool'
	EitherVal func() Functional // None and Just are a pair of optional types

	// NONE/JUST/CASE/ERROR
	NoneVal  func()                     // None and Just are a pair of optional types
	JustVal  func() Functional          // implementing 'Option :: Maybe() bool'
	ErrorVal func(err ...error) []error // stackable errors

	// IO RELATED SIDE EFFECTS
	IOSyncFnc  func() Optional         // Just | None, blocks until value yielded
	IOAsyncFnc func(update Functional) // updates a value by calling the update fnc
)

/////////////////////////////////////////////////////////////////////////////
/// PRAEDICATE
//
// encloses a test that expects some type of input value to test against and
// returns either true, or false.
func NewPredicate(pred func(scrut ...Functional) bool) PredFnc { return PredFnc(pred) }
func (p PredFnc) TypeFnc() TyFnc                               { return Predicate }
func (p PredFnc) TypeNat() d.TyNative                          { return d.Bool }
func (p PredFnc) Ident() Functional                            { return p }
func (p PredFnc) String() string {
	return "T → λpredicate  → Bool"
}
func (p PredFnc) Eval(dat ...d.Native) d.Native {
	if len(dat) > 0 {
		return d.BoolVal(p(New(dat[0])))
	}
	return d.NilVal{}
}
func (p PredFnc) Call(v ...Functional) Functional {
	if len(v) > 0 {
		p(v[0])
	}
	return NewNone()
}

/////////////////////////////////////////////////////////////////////////////
//// OPTIONAL FUNCTIONS
///
// generic optional
func NewOptional(predicate PredFnc, left, right Optional) OptionalFnc {
	return func(scrutinee ...Functional) Optional {
		if predicate(scrutinee...) {
			return left
		}
		return right
	}
}
func (o OptionalFnc) TypeFnc() TyFnc      { return Option }
func (o OptionalFnc) TypeNat() d.TyNative { return d.Function | d.Type }
func (o OptionalFnc) String() string      { return "Option" }
func (o OptionalFnc) Value() Functional   { return NewNone() }
func (o OptionalFnc) Maybe() bool {
	return false // since no argument has been applyed
}
func (o OptionalFnc) Call(vals ...Functional) Functional {
	return o(vals...)
}
func (o OptionalFnc) Eval(nats ...d.Native) d.Native {
	var fncs []Functional
	for _, nat := range nats {
		if nat.TypeNat().Flag().Match(d.Function) {
			fncs = append(fncs, nat.(Functional))
			continue
		}
		fncs = append(fncs, NewFromData(nat))
	}
	return o(fncs...)
}

// maybe takes a preadicate & an expression that's expected to evaluate to a
// value, or not. when called, the expressions evaluation result is then
// applyed to the praedicate and depending on this applications result, either
// the value is returned wrapped in a 'Just" instance.
func NewMaybe(pred PredFnc, expr Functional) OptionalFnc {
	return NewOptional(pred, NewJustVal(expr), NewNone())
}

// either takes a predicate and two expressions, one of which is expected to
// evaluate true, when applyed to the predicate.
func NewEither(pred PredFnc, lexpr, rexpr Functional) OptionalFnc {
	return NewOptional(pred, NewEitherVal(lexpr), NewOrVal(rexpr))
}

// truth function takes a predicate and returns either true, or false
func NewTruth(pred PredFnc) OptionalFnc {
	return NewOptional(pred, NewTrue(), NewFalse())
}

/////////////////////////////////////////////////////////////////////////////
/// CASE
//
func NewCaseExpr(pred PredFnc, expr Functional) CaseExpr {
	return CaseExpr(func(scrut ...Functional) Optional {
		if pred(scrut...) {
			return NewJustVal(expr)
		}
		return NewNone()
	})
}
func NewCaseFnc(cases ...CaseExpr) CaseFnc {
	// declare case function constructor expects a list of cases to apply
	// to the passed scrutinee & the expression to return when a case
	// statement matches.
	var conCases func(...CaseExpr) CaseFnc
	// define case function constructor
	conCases = func(cases ...CaseExpr) CaseFnc {
		return CaseFnc(func(scrut ...Functional) Optional {
			// as long as there are cases to evaluate‥.
			if len(cases) > 0 {
				// evaluate first passed case
				var c = cases[0](scrut...)
				if c.Maybe() {
					// maybe a value? → return 'Just' instance
					return NewJustVal(c.Value())
				}
				// if there are additional cases to evaluate‥.
				if len(cases) > 1 {
					// pass on remaining cases to the case
					// function constructor, and apply
					// scrutinee recursively
					cases = cases[1:]
					return conCases(cases...)(scrut...)
				}
			}
			// all cases are depleted without one matching → return
			// final 'None' instance
			return NewNone()
		})
	}
	return conCases(cases...)
}
func (c CaseFnc) Ident() Functional                  { return c }
func (c CaseFnc) Maybe() bool                        { return c().Maybe() }
func (c CaseFnc) Value() Functional                  { return c().Value() }
func (c CaseFnc) Call(vals ...Functional) Functional { return c(vals...) }
func (c CaseFnc) TypeFnc() TyFnc                     { return Case | Option }
func (c CaseFnc) TypeNat() d.TyNative                { return c().TypeNat() }
func (c CaseFnc) String() string                     { return c().String() }
func (c CaseFnc) Eval(nats ...d.Native) d.Native {
	var vals = []Functional{}
	for _, val := range nats {
		vals = append(vals, NewFromData(val.Eval()))
	}
	return c.Call(vals...).Eval()
}

/////////////////////////////////////////////////////////////////////////////
//// RETURN TYPE PAIRS OF THE OPTIONAL TYPE
///
// EITHER
func NewEitherVal(v Functional) EitherVal {
	return EitherVal(func() Functional { return v })
}
func (e EitherVal) Ident() Functional               { return e }
func (e EitherVal) Call(v ...Functional) Functional { return e.Call(v...) }
func (e EitherVal) Eval(p ...d.Native) d.Native     { return e().Eval(p...) }
func (e EitherVal) Maybe() bool                     { return true }
func (e EitherVal) Value() Functional               { return e() }
func (e EitherVal) Nullable() d.Native              { return e.Eval() }
func (e EitherVal) TypeFnc() TyFnc                  { return EitherOr | Either }
func (e EitherVal) TypeNat() d.TyNative             { return e().TypeNat() }
func (e EitherVal) String() string                  { return e().String() }

// OR
func NewOrVal(v Functional) OrVal {
	return OrVal(func() Functional { return v })
}
func (o OrVal) Ident() Functional               { return o }
func (o OrVal) Call(v ...Functional) Functional { return o.Call(v...) }
func (o OrVal) Eval(p ...d.Native) d.Native     { return o().Eval(p...) }
func (o OrVal) Maybe() bool                     { return false }
func (o OrVal) Value() Functional               { return o() }
func (o OrVal) Nullable() d.Native              { return o.Eval() }
func (o OrVal) TypeFnc() TyFnc                  { return EitherOr | Or }
func (o OrVal) TypeNat() d.TyNative             { return o().TypeNat() }
func (o OrVal) String() string                  { return o().String() }

// JUST
func NewJustVal(v Functional) JustVal {
	return JustVal(func() Functional { return v })
}
func (j JustVal) Ident() Functional               { return j }
func (j JustVal) Call(v ...Functional) Functional { return j.Call(v...) }
func (j JustVal) Eval(p ...d.Native) d.Native     { return j().Eval(p...) }
func (j JustVal) Maybe() bool                     { return true }
func (j JustVal) Value() Functional               { return j() }
func (j JustVal) Nullable() d.Native              { return j.Eval() }
func (j JustVal) TypeFnc() TyFnc                  { return Option | Just }
func (j JustVal) TypeNat() d.TyNative             { return j().TypeNat() }
func (j JustVal) String() string                  { return j().String() }

// NONE
func NewNone() NoneVal                          { return NoneVal(func() {}) }
func (n NoneVal) Ident() Functional             { return n }
func (n NoneVal) Call(...Functional) Functional { return n }
func (n NoneVal) Eval(...d.Native) d.Native     { return d.NilVal{}.Eval() }
func (n NoneVal) Maybe() bool                   { return false }
func (n NoneVal) Value() Functional             { return NewNone() }
func (n NoneVal) Nullable() d.Native            { return d.NilVal{} }
func (n NoneVal) TypeFnc() TyFnc                { return Option | None }
func (n NoneVal) TypeNat() d.TyNative           { return d.Nil }
func (n NoneVal) String() string                { return "⊥" }

//// FUNCTIONAL TRUTH VALUES
///
// TRUE
func (t TrueVal) Call(...Functional) Functional {
	return New(d.BoolVal(true))
}
func (t TrueVal) Ident() Functional         { return t }
func (t TrueVal) Eval(...d.Native) d.Native { return t }
func (t TrueVal) Bool() bool                { return true }
func (t TrueVal) Maybe() bool               { return true }
func (t TrueVal) Value() Functional         { return t }
func (t TrueVal) TypeFnc() TyFnc            { return Truth | True }
func (t TrueVal) TypeNat() d.TyNative       { return d.Bool }
func (t TrueVal) String() string            { return "True" }
func NewTrue() TrueVal {
	return TrueVal(func() Boolean { return NewTrue() })
}

// FALSE
func (f FalseVal) Call(...Functional) Functional {
	return New(d.BoolVal(false))
}
func (f FalseVal) Iwdent() Functional        { return f }
func (f FalseVal) Eval(...d.Native) d.Native { return f }
func (f FalseVal) Bool() bool                { return false }
func (f FalseVal) Maybe() bool               { return false }
func (f FalseVal) Value() Functional         { return f }
func (f FalseVal) TypeFnc() TyFnc            { return Truth | False }
func (f FalseVal) TypeNat() d.TyNative       { return d.Bool | d.Function }
func (f FalseVal) String() string            { return "False" }
func NewFalse() TrueVal {
	return func() Boolean { return NewFalse() }
}

// ERROR
func (f ErrorVal) Call(fncs ...Functional) Functional { return f }
func (f ErrorVal) Ident() Functional                  { return f }
func (f ErrorVal) Value() Functional                  { return f }
func (f ErrorVal) Eval(...d.Native) d.Native          { return f }
func (f ErrorVal) Bool() bool                         { return false }
func (f ErrorVal) TypeFnc() TyFnc                     { return Error }
func (f ErrorVal) TypeNat() d.TyNative                { return d.Function }
func (f ErrorVal) String() string                     { return f.String() }
func (f ErrorVal) Error() string {
	var str string
	for i, err := range f() {
		str = fmt.Sprintf("Error %d:\t", i)
		str = str + fmt.Sprintf("%s\n\n", err.Error())
	}
	return str
}
func NewError(err ...error) ErrorVal {
	return ErrorVal(func(err ...error) []error { return err })
}
