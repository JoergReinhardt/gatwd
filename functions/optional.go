package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

type (
	//// FUNCTOR CONSTRUCTORS
	///
	// CONDITIONAL, CASE & OPTIONAL
	NoOp     func()
	PredExpr func(...Functional) bool
	MaybeVal func() Paired
	CaseVal  func() (PredExpr, Functional)
	CaseExpr func() []CaseVal
)

// PRAEDICATE
func NewPredicate(pred func(scrut ...Functional) bool) PredExpr { return PredExpr(pred) }
func (p PredExpr) TypeFnc() TyFnc                               { return Predicate }
func (p PredExpr) TypeNat() d.TyNative                          { return d.Bool }
func (p PredExpr) Ident() Functional                            { return p }
func (p PredExpr) String() string {
	return "T → λpredicate  → Bool"
}

func (p PredExpr) True(val Functional) bool { return p(val) }

func (p PredExpr) All(val ...Functional) bool {
	for _, v := range val {
		if !p(v) {
			return false
		}
	}
	return true
}

func (p PredExpr) Any(val ...Functional) bool {
	for _, v := range val {
		if p(v) {
			return true
		}
	}
	return false
}

func (p PredExpr) Call(v ...Functional) Functional {
	if len(v) == 1 {
		return New(p.True(v[0]))
	}
	return New(p.All(v...))
}

func (p PredExpr) Eval(dat ...d.Native) d.Native {
	var fncs = []Functional{}
	for _, nat := range dat {
		fncs = append(fncs, NewFromData(nat))
	}
	return p.Call(fncs...)
}

// NONE
func NewNoOp() NoOp                          { return NoOp(func() {}) }
func (n NoOp) Ident() Functional             { return n }
func (n NoOp) Call(...Functional) Functional { return n }
func (n NoOp) Eval(...d.Native) d.Native     { return d.NilVal{}.Eval() }
func (n NoOp) Maybe() bool                   { return false }
func (n NoOp) Value() Functional             { return NewNoOp() }
func (n NoOp) Nullable() d.Native            { return d.NilVal{} }
func (n NoOp) TypeFnc() TyFnc                { return Option | None }
func (n NoOp) TypeNat() d.TyNative           { return d.Nil }
func (n NoOp) String() string                { return "⊥" }

// OPTIONAL
func (o MaybeVal) TypeNat() d.TyNative {
	return d.Function | o.Value().TypeNat()
}
func (o MaybeVal) TypeFnc() TyFnc {
	return Option | o.Value().TypeFnc()
}
func (o MaybeVal) Eval(dat ...d.Native) d.Native     { return o.Value().Eval(dat...) }
func (o MaybeVal) Call(val ...Functional) Functional { return o.Value().Call(val...) }
func (o MaybeVal) String() string                    { return o.Value().String() }
func (o MaybeVal) Left() Functional                  { return o().Left() }
func (o MaybeVal) Right() Functional                 { return o().Right() }
func (o MaybeVal) Maybe() bool {
	if b, ok := o().Left().Eval().(d.BoolVal); ok && bool(b) {
		return true
	}
	return false
}
func (o MaybeVal) Value() Functional {
	if o.Maybe() {
		return o().Left()
	}
	return NewNoOp()
}
func NewMaybe(maybe bool, expr ...Functional) MaybeVal {
	if maybe && len(expr) > 0 {
		var exp Functional
		if len(expr) > 1 {
			exp = NewVector(expr...)
		} else {
			exp = expr[0]
		}
		return MaybeVal(func() Paired {
			return NewPair(
				NewFromData(d.BoolVal(true)),
				exp,
			)
		})
	}
	return MaybeVal(func() Paired {
		return NewPair(
			NewFromData(d.BoolVal(false)),
			NewNoOp(),
		)
	})
}

/// CASE VALUE
// returns a predicate function to apply a single argument to & an expression
// to be returned in case predicate application evaluates true.
func NewCaseVal(pred PredExpr, expr Functional) CaseVal {
	return CaseVal(func() (PredExpr, Functional) {
		return pred, expr
	})
}
func (c CaseVal) pred() PredExpr   { p, _ := c(); return p }
func (c CaseVal) expr() Functional { _, e := c(); return e }

// implements case interface and returns an optional
func (c CaseVal) Case(expr ...Functional) Functional {
	var opt = c.CaseMaybe(expr...)
	if opt.Maybe() {
		return opt.Value()
	}
	return NewNoOp()
}

// returns optional either containing the enclosed expression, or no-op based
// on the result of applying the scrutinee to the enclosed case expression.
func (c CaseVal) CaseMaybe(scruts ...Functional) MaybeVal {
	// range over all passed arguments
	for _, scrut := range scruts {
		// test against enclosed case
		if pred, ret := c(); pred(scrut) {
			// return optional containing enclosed epression, on
			// first matching case
			return NewMaybe(true, ret)
		}
	}
	// return optional containing no-op, of no case matches.
	return NewMaybe(false)
}
func (c CaseVal) TypeNat() d.TyNative {
	return d.Function | c.expr().TypeNat()
}
func (c CaseVal) TypeFnc() TyFnc {
	return Case | c.expr().TypeFnc()
}
func (c CaseVal) Eval(dat ...d.Native) d.Native     { return c.expr().Eval(dat...) }
func (c CaseVal) Call(val ...Functional) Functional { return c.expr().Call(val...) }
func (c CaseVal) String() string                    { return "case true: " + c.expr().String() }

/// CASE FUNCTION
func NewCaseExpr(cases ...CaseVal) CaseExpr {
	return CaseExpr(func() []CaseVal { return cases })
}
func (c CaseExpr) Case(expr ...Functional) Functional {
	// range over all cases
	for _, cas := range c() {
		// each case ranges over all expressions to scrutinize to
		// yield an optional
		var option = cas.CaseMaybe(expr...)
		// if optional contains value‥.
		if option.Maybe() {
			//‥.return first match
			return option.Value()
		}
	}
	// if nothing matched, return empty
	return NewNoOp()
}
func (c CaseExpr) TypeNat() d.TyNative {
	return d.Function
}
func (c CaseExpr) TypeFnc() TyFnc {
	return Case
}
func (c CaseExpr) Call(val ...Functional) Functional {
	return c.Case(val...)
}
func (c CaseExpr) Eval(dat ...d.Native) d.Native {
	var fncs = []Functional{}
	for _, nat := range dat {
		fncs = append(fncs, NewFromData(nat))
	}
	return c.Call(fncs...)
}
func (c CaseExpr) String() string {
	var str = "cases:\n"
	var cases = c()
	var l = len(cases)
	for i, cas := range cases {
		str = str + cas.String()
		if i < l-1 {
			str = str + "\n"

		}
	}
	return str
}
