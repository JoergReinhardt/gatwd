package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

type (
	//// FUNCTOR CONSTRUCTORS
	///
	// CONDITIONAL, CASE & OPTIONAL
	NoOp       func()
	OptVal     func() PairFnc
	PredExpr   func(...Parametric) bool
	JustNone   func(...Parametric) OptVal
	EitherOr   func(...Parametric) OptVal
	SwitchCase func(...Parametric) OptVal
	CaseSwitch func(...Parametric) (SwitchCase, []SwitchCase)
	// RESORCFUL FUNCTIONS (depend on free vars besides argset
	GeneratorFnc  func() (Parametric, GeneratorFnc)
	AggregatorFnc func(args ...Parametric) (Parametric, AggregatorFnc)
)

/// OPTIONAL VALUES
// based on a paired with extra bells'n whistles
func NewOptVal(left, right Parametric) OptVal {
	return OptVal(func() PairFnc {
		return NewPair(left, right)
	})
}
func (o OptVal) Ident() Parametric   { return o }
func (o OptVal) Left() Parametric    { return o().Left() }
func (o OptVal) Right() Parametric   { return o().Right() }
func (o OptVal) TypeNat() d.TyNative { return o.Right().TypeNat() }
func (o OptVal) TypeFnc() TyFnc      { return o.Right().TypeFnc() | Option }
func (o OptVal) String() string {
	return "optional: " + o.Left().String() + "|" + o.Right().String()
}
func (o OptVal) Call(args ...Parametric) Parametric { return o.Right().Call(args...) }
func (o OptVal) Eval(vars ...d.Native) d.Native {
	var args = []Parametric{}
	for _, arg := range vars {
		args = append(args, NewFromData(arg))
	}
	return o.Call(args...)
}

//// JUST NONE
///
// expression is applyed to arguments passed at runtime. result of calling the
// expression is applyed to predex. if the predicate matches, result is
// returned as 'just' value, otherwise NoOp is returned
func NewJustNone(predex PredExpr, expr Parametric) JustNone {
	return JustNone(func(args ...Parametric) OptVal {
		var result = expr.Call(args...)
		if predex(result) {
			return NewOptVal(New(true), result)
		}
		return NewOptVal(New(false), NewNoOp())
	})
}
func (j JustNone) Ident() Parametric                  { return j }
func (j JustNone) String() string                     { return "just-none" }
func (j JustNone) TypeFnc() TyFnc                     { return Option | Just | None }
func (j JustNone) TypeNat() d.TyNative                { return d.Function }
func (j JustNone) Return(args ...Parametric) OptVal   { return j(args...) }
func (j JustNone) Call(args ...Parametric) Parametric { return j(args...) }
func (j JustNone) Eval(vars ...d.Native) d.Native {
	var args = []Parametric{}
	for _, v := range vars {
		args = append(args, NewFromData(v))
	}
	return j(args...)
}

//// EITHER OR
///
// left pair value indicates: 0 = 'either', 1 = 'or', -1 = 'no value yielded'
func NewEitherOr(predex PredExpr, either, or Parametric) EitherOr {
	return EitherOr(func(args ...Parametric) OptVal {
		var val Parametric
		val = either.Call(args...)
		if predex(val) {
			return NewOptVal(New(0), val)
		}
		val = or.Call(args...)
		if predex(val) {
			return NewOptVal(New(1), val)
		}
		return NewOptVal(New(-1), NewNoOp())
	})
}
func (e EitherOr) Ident() Parametric                  { return e }
func (e EitherOr) String() string                     { return "either-or" }
func (e EitherOr) TypeFnc() TyFnc                     { return Option | Either | Or }
func (e EitherOr) TypeNat() d.TyNative                { return d.Function }
func (e EitherOr) Return(args ...Parametric) OptVal   { return e(args...) }
func (e EitherOr) Call(args ...Parametric) Parametric { return e(args...) }
func (e EitherOr) Eval(vars ...d.Native) d.Native {
	var args = []Parametric{}
	for _, v := range vars {
		args = append(args, NewFromData(v))
	}
	return e.Call(args...)
}

//// SWITCH CASE
///
// switch case applies predicate to arguments passed at runtime & either
// returns either the enclosed expression, the next case, or no-op in case an
// error occured.
func NewSwitchCase(predex PredExpr, value Parametric, nextcase ...SwitchCase) SwitchCase {
	// if runtime arguments applyed to predicate expression yields true, value
	// will be returned, or otherwise the next case will be the return value.
	return SwitchCase(func(args ...Parametric) OptVal {
		if predex(args...) { // return value if runtime args match predicate
			return OptVal(func() PairFnc {
				return NewPair(New(0), value)
			})
		} // return next switch case to test against, if at least one
		// more case was passed
		if len(nextcase) > 0 {
			return OptVal(func() PairFnc {
				return NewPair(New(1), nextcase[0])
			})
		}
		return OptVal(func() PairFnc {
			return NewPair(New(-1), NewNoOp())
		})
	})
}
func (s SwitchCase) Ident() Parametric                  { return s }
func (s SwitchCase) String() string                     { return "switch-case" }
func (s SwitchCase) TypeFnc() TyFnc                     { return Option | Case | Switch }
func (s SwitchCase) TypeNat() d.TyNative                { return d.Function }
func (s SwitchCase) Return(args ...Parametric) OptVal   { return s(args...) }
func (s SwitchCase) Call(args ...Parametric) Parametric { return s(args...) }
func (s SwitchCase) Eval(vars ...d.Native) d.Native {
	var args = []Parametric{}
	for _, v := range vars {
		args = append(args, NewFromData(v))
	}
	return s.Call(args...)
}

/// GENERATOR
// initialize generator function with any Parametric function. will be called
// without arguments & is expexted tp return a new value per call
func NewGenerator(fnc Parametric) GeneratorFnc {
	var next = fnc.Call()
	var genfnc GeneratorFnc
	genfnc = GeneratorFnc(func() (Parametric, GeneratorFnc) {
		return next, conGenerator(genfnc)
	})
	return genfnc
}
func conGenerator(genfnc GeneratorFnc) GeneratorFnc {
	var value Parametric
	value, genfnc = genfnc()
	return GeneratorFnc(func() (Parametric, GeneratorFnc) {
		return value, conGenerator(genfnc)
	})
}

func (g GeneratorFnc) Ident() Parametric                  { return g }
func (g GeneratorFnc) String() string                     { return "generator" }
func (g GeneratorFnc) Current() Parametric                { p, _ := g(); return p }
func (g GeneratorFnc) Generator() GeneratorFnc            { _, gen := g(); return gen }
func (g GeneratorFnc) TypeFnc() TyFnc                     { return Generator }
func (g GeneratorFnc) TypeNat() d.TyNative                { return g.Current().TypeNat() }
func (g GeneratorFnc) Eval(vars ...d.Native) d.Native     { return g.Current().Eval(vars...) }
func (g GeneratorFnc) Call(args ...Parametric) Parametric { return g.Current().Call(args...) }

/// AGGREGATOR
//  applies old parameter and new args to nary function to yield aggregated
//  result
func NewAggregator(nary NaryFnc) AggregatorFnc {
	var aggrFnc AggregatorFnc
	aggrFnc = AggregatorFnc(func(args ...Parametric) (Parametric, AggregatorFnc) {
		var aggr = nary(args...)
		var aggregator = conAggregator(aggrFnc, aggr)
		return aggr, aggregator
	})
	return aggrFnc
}

// construct aggregator, optionaly pass arguments to aggregate with arguments
// at call site, to compute aggregated result.
func conAggregator(aggregator AggregatorFnc, passed ...Parametric) AggregatorFnc {

	return AggregatorFnc(func(args ...Parametric) (Parametric, AggregatorFnc) {
		var aggr Parametric
		aggr, aggregator = aggregator(args...)
		return aggr, conAggregator(aggregator, append(passed, args...)...)
	})
}
func (g AggregatorFnc) Ident() Parametric                          { return g }
func (g AggregatorFnc) TypeFnc() TyFnc                             { return Aggregator }
func (g AggregatorFnc) TypeNat() d.TyNative                        { return d.Function }
func (g AggregatorFnc) Aggregator() AggregatorFnc                  { _, aggr := g(); return aggr }
func (g AggregatorFnc) Current() Parametric                        { cur, _ := g(); return cur }
func (g AggregatorFnc) String() string                             { return "aggregator" }
func (g AggregatorFnc) Eval(p ...d.Native) d.Native                { return g.Aggregator().Eval(p...) }
func (g AggregatorFnc) Call(d ...Parametric) Parametric            { return g.Aggregator().Call(d...) }
func (g AggregatorFnc) Aggregate(args ...Parametric) AggregatorFnc { return conAggregator(g, args...) }

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
func (n NoOp) Maybe() bool                   { return false }
func (n NoOp) Empty() bool                   { return true }
func (n NoOp) String() string                { return "⊥" }
func (n NoOp) Len() int                      { return -1 }
func (n NoOp) Value() Parametric             { return n }
func (n NoOp) Ident() Parametric             { return n }
func (n NoOp) Call(...Parametric) Parametric { return n }
func (n NoOp) Eval(...d.Native) d.Native     { return d.NilVal{} }
func (n NoOp) TypeNat() d.TyNative           { return d.Nil }
func (n NoOp) TypeFnc() TyFnc                { return None }
