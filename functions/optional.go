package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

type (
	//// FUNCTOR CONSTRUCTORS
	///
	// CONDITIONAL, CASE & OPTIONAL
	NoOp     func()
	PredExpr func(...Parametric) bool
	JustNone func() Paired
	EitherOr func() Paired
	CaseVal  func() (PredExpr, Parametric)
	CaseExpr func() []CaseVal
	// RESORCFUL FUNCTIONS (depend on free vars besides argset
	GeneratorFnc  func() Parametric
	AggregatorFnc func(args ...Parametric) (Parametric, NaryFnc)
)

/// GENERATOR
func NewGenerator(fnc func() Parametric) GeneratorFnc {
	return GeneratorFnc(func() Parametric { return fnc() })
}

func (g GeneratorFnc) Ident() Parametric               { return g() }
func (g GeneratorFnc) TypeFnc() TyFnc                  { return Generator }
func (g GeneratorFnc) TypeNat() d.TyNative             { return g().TypeNat() }
func (g GeneratorFnc) Eval(p ...d.Native) d.Native     { return g().Eval() }
func (g GeneratorFnc) Call(d ...Parametric) Parametric { return g() }
func (g GeneratorFnc) Next() JustNone {
	var elem = g()
	if !ElemEmpty(elem) {
		return NewJustNone(true, elem)
	}
	return NewJustNone(false)
}

/// AGGREGATOR
//  applies old parameter and new args to nary function to yield aggregated
//  result
func NewAggregator(aggr NaryFnc) AggregatorFnc { return conAggregator(aggr) }

// construct aggregator, optionaly pass arguments to aggregate with arguments
// at call site, to compute aggregated result.
func conAggregator(aggregator NaryFnc, passed ...Parametric) AggregatorFnc {
	return AggregatorFnc(func(args ...Parametric) (Parametric, NaryFnc) {
		if len(passed) > 0 {
			return aggregator(append(passed, args...)...), aggregator
		}
		return aggregator(args...), aggregator
	})
}
func (g AggregatorFnc) Ident() Parametric               { return g }
func (g AggregatorFnc) TypeFnc() TyFnc                  { return Aggregator }
func (g AggregatorFnc) TypeNat() d.TyNative             { return g.Result().TypeNat() }
func (g AggregatorFnc) String() string                  { return g.Result().String() }
func (g AggregatorFnc) Eval(p ...d.Native) d.Native     { return g.Aggregator().Eval(p...) }
func (g AggregatorFnc) Call(d ...Parametric) Parametric { return g.Aggregator()(d...) }
func (g AggregatorFnc) Result() Parametric              { parm, _ := g(); return parm }
func (g AggregatorFnc) Aggregator() NaryFnc             { _, aggr := g(); return aggr }
func (g AggregatorFnc) Aggregate(args ...Parametric) Parametric {
	return g.Aggregator()(append([]Parametric{g.Result()}, args...)...)
}

// PRAEDICATE
func NewPredicate(pred func(scrut ...Parametric) bool) PredExpr { return PredExpr(pred) }
func (p PredExpr) TypeFnc() TyFnc                               { return Predicate }
func (p PredExpr) TypeNat() d.TyNative                          { return d.Bool }
func (p PredExpr) Ident() Parametric                            { return p }
func (p PredExpr) String() string {
	return "T → λpredicate  → Bool"
}

func (p PredExpr) True(val Parametric) bool { return p(val) }

func (p PredExpr) All(val ...Parametric) bool {
	for _, v := range val {
		if !p(v) {
			return false
		}
	}
	return true
}

func (p PredExpr) Any(val ...Parametric) bool {
	for _, v := range val {
		if p(v) {
			return true
		}
	}
	return false
}

func (p PredExpr) Call(v ...Parametric) Parametric {
	if len(v) == 1 {
		return New(p.True(v[0]))
	}
	return New(p.All(v...))
}

func (p PredExpr) Eval(dat ...d.Native) d.Native {
	var fncs = []Parametric{}
	for _, nat := range dat {
		fncs = append(fncs, NewFromData(nat))
	}
	return p.Call(fncs...)
}

// NONE
func NewNoOp() NoOp                          { return NoOp(func() {}) }
func (n NoOp) Ident() Parametric             { return n }
func (n NoOp) Call(...Parametric) Parametric { return n }
func (n NoOp) Eval(...d.Native) d.Native     { return d.NilVal{}.Eval() }
func (n NoOp) Maybe() bool                   { return false }
func (n NoOp) Value() Parametric             { return NewNoOp() }
func (n NoOp) Nullable() d.Native            { return d.NilVal{} }
func (n NoOp) TypeFnc() TyFnc                { return Option | None }
func (n NoOp) TypeNat() d.TyNative           { return d.Nil }
func (n NoOp) String() string                { return "⊥" }

// OPTIONAL
func (o JustNone) TypeNat() d.TyNative {
	return d.Function | o.Value().TypeNat()
}
func (o JustNone) TypeFnc() TyFnc {
	return Option | o.Value().TypeFnc()
}
func (o JustNone) Eval(dat ...d.Native) d.Native     { return o.Value().Eval(dat...) }
func (o JustNone) Call(val ...Parametric) Parametric { return o.Value().Call(val...) }
func (o JustNone) String() string                    { return o.Value().String() }
func (o JustNone) Left() Parametric                  { return o().Left() }
func (o JustNone) Right() Parametric                 { return o().Right() }
func (o JustNone) Maybe() bool {
	if b, ok := o().Left().Eval().(d.BoolVal); ok && bool(b) {
		return true
	}
	return false
}
func (o JustNone) Value() Parametric {
	if o.Maybe() {
		return o().Left()
	}
	return NewNoOp()
}
func NewJustNone(maybe bool, expr ...Parametric) JustNone {
	if maybe && len(expr) > 0 {
		var exp Parametric
		if len(expr) > 1 {
			exp = NewVector(expr...)
		} else {
			exp = expr[0]
		}
		return JustNone(func() Paired {
			return NewPair(
				NewFromData(d.BoolVal(true)),
				exp,
			)
		})
	}
	return JustNone(func() Paired {
		return NewPair(
			NewFromData(d.BoolVal(false)),
			NewNoOp(),
		)
	})
}

/// CASE VALUE
// returns a predicate function to apply a single argument to & an expression
// to be returned in case predicate application evaluates true.
func NewCaseVal(pred PredExpr, expr Parametric) CaseVal {
	return CaseVal(func() (PredExpr, Parametric) {
		return pred, expr
	})
}
func (c CaseVal) Predicate() PredExpr    { p, _ := c(); return p }
func (c CaseVal) Expression() Parametric { _, e := c(); return e }

// returns optional either containing the enclosed expression, or no-op based
// on the result of applying the scrutinee to the enclosed case expression.
func (c CaseVal) CaseMaybe(scruts ...Parametric) JustNone {
	// range over all passed arguments
	for _, scrut := range scruts {
		// test against enclosed case
		if pred, ret := c(); pred(scrut) {
			// return optional containing enclosed epression, on
			// first matching case
			return NewJustNone(true, ret)
		}
	}
	// return optional containing no-op, of no case matches.
	return NewJustNone(false)
}

// implements case interface and returns an optional
func (c CaseVal) Case(expr ...Parametric) Parametric {
	var opt = c.CaseMaybe(expr...)
	if opt.Maybe() {
		return opt.Value()
	}
	return NewNoOp()
}
func (c CaseVal) TypeNat() d.TyNative {
	return d.Function | c.Expression().TypeNat()
}
func (c CaseVal) TypeFnc() TyFnc {
	return Case | c.Expression().TypeFnc()
}
func (c CaseVal) Eval(dat ...d.Native) d.Native     { return c.Expression().Eval(dat...) }
func (c CaseVal) Call(val ...Parametric) Parametric { return c.Expression().Call(val...) }
func (c CaseVal) String() string                    { return "case true: " + c.Expression().String() }

/// CASE FUNCTION
func NewCaseExpr(cases ...CaseVal) CaseExpr {
	return CaseExpr(func() []CaseVal { return cases })
}
func (c CaseExpr) Case(expr ...Parametric) Parametric {
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
func (c CaseExpr) Call(val ...Parametric) Parametric {
	return c.Case(val...)
}
func (c CaseExpr) Eval(dat ...d.Native) d.Native {
	var fncs = []Parametric{}
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
