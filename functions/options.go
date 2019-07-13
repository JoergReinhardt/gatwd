package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

type (
	//// NONE VALUE CONSTRUCTOR
	NoneVal func()
	//// TRUTH VALUE CONSTRUCTOR
	TestExpr func(...Expression) Typed

	//// CASE & SWITCH TYPE CONSTRUCTORS
	CaseExpr   func(...Expression) (Expression, bool)
	CaseSwitch func(...Expression) (Expression, CaseSwitch, bool)
)

//// NONE VALUE CONSTRUCTOR
///
// none represens the abscence of a value of any type. implements consumeable,
// key-, index & generic pair interface to be returneable as such.
func NewNone() NoneVal { return func() {} }

func (n NoneVal) Ident() Expression             { return n }
func (n NoneVal) Head() Expression              { return n }
func (n NoneVal) Tail() Consumeable             { return n }
func (n NoneVal) Len() d.IntVal                 { return 0 }
func (n NoneVal) String() string                { return "⊥" }
func (n NoneVal) Call(...Expression) Expression { return nil }
func (n NoneVal) Key() Expression               { return nil }
func (n NoneVal) Index() Expression             { return nil }
func (n NoneVal) Left() Expression              { return nil }
func (n NoneVal) Right() Expression             { return nil }
func (n NoneVal) Both() Expression              { return nil }
func (n NoneVal) Value() Expression             { return nil }
func (n NoneVal) Empty() d.BoolVal              { return true }
func (n NoneVal) Flag() d.BitFlag               { return d.BitFlag(None) }
func (n NoneVal) TypeFnc() TyFnc                { return None }
func (n NoneVal) TypeNat() d.TyNat              { return d.Nil }
func (n NoneVal) TypeElem() TyFnc               { return None }
func (n NoneVal) TypeName() string              { return n.String() }
func (n NoneVal) FlagType() d.Uint8Val          { return Flag_Function.U() }
func (n NoneVal) Type() Typed                   { return Define(n.TypeName(), None) }
func (n NoneVal) Consume() (Expression, Consumeable) {
	return NewNone(), NewNone()
}

//// TRUTH VALUE CONSTRUCTOR
func NewTestTruth(name string, test func(...Expression) d.BoolVal, paratypes ...Typed) TestExpr {

	if name == "" {
		name = "Truth"
	}
	var params = make([]Typed, 0, len(paratypes))
	if len(paratypes) == 0 {
		paratypes = append(paratypes, Type)
	} else {
		for _, param := range paratypes {
			params = append(params, param)
		}
	}

	return func(args ...Expression) Typed {
		if len(args) > 0 {
			if test(args...) {
				return True

			}
			return False
		}
		return Define(name, Truth, params...)
	}
}

func NewTestTrinary(name string, test func(...Expression) int, paratypes ...Typed) TestExpr {

	if name == "" {
		name = "Trinary"
	}
	var params = make([]Typed, 0, len(paratypes))
	if len(paratypes) == 0 {
		paratypes = append(paratypes, Type)
	} else {
		for _, param := range paratypes {
			params = append(params, param)
		}
	}

	return func(args ...Expression) Typed {
		if len(args) > 0 {
			if test(args...) > 0 {
				return True
			}
			if test(args...) < 0 {
				return False
			}
			return Undecided
		}
		return Define(name, Trinary, params...)
	}
}

func NewTestCMP(name string, test func(...Expression) d.IntVal, paratypes ...Typed) TestExpr {

	if name == "" {
		name = "CMP"
	}
	var params = make([]Typed, 0, len(paratypes))
	if len(paratypes) == 0 {
		paratypes = append(paratypes, Type)
	}
	for _, param := range paratypes {
		params = append(params, param)
	}

	return func(args ...Expression) Typed {
		if len(args) > 0 {
			if test(args...) > 0 {
				return GT
			}
			if test(args...) < 0 {
				return LT
			}
			return EQ
		}
		return Define(name, CMP, params...)
	}
}
func (t TestExpr) Type() Typed    { return t().(Typed) }
func (t TestExpr) TypeFnc() TyFnc { return t.Type().(TyDef).Return().(TyFnc) }
func (t TestExpr) TypeNat() d.TyNat {
	if t.TypeFnc() == CMP {
		return d.Int
	}
	return d.Bool
}

func (t TestExpr) TypeName() string {
	return t.Type().(TyDef).Name() + " → " + t.Type().(TyDef).Return().TypeName()
}
func (t TestExpr) String() string       { return t.TypeName() }
func (t TestExpr) FlagType() d.Uint8Val { return Flag_Function.U() }
func (t TestExpr) Call(args ...Expression) Expression {
	if t.TypeFnc() == CMP {
		return NewData(d.IntVal(t.CMP(args...)))
	}
	return NewData(d.BoolVal(t.Test(args...)))
}

func (t TestExpr) Test(args ...Expression) d.BoolVal {
	if t.TypeFnc() == CMP {
		if t(args...) == LT || t(args...) == GT {
			return false
		} else {
			return true
		}
	}
	if t.TypeFnc() == Trinary {
		if t(args...) == False || t(args...) == Undecided {
			return false
		} else {
			return true
		}
	}
	if t(args...) != True {
		return false
	}
	return true
}

func (t TestExpr) CMP(args ...Expression) d.IntVal {
	if t.TypeFnc() == CMP {
		switch t(args...) {
		case LT:
			return -1
		case EQ:
			return 0
		case GT:
			return 1
		}
	}
	if t.TypeFnc() == Trinary {
		switch t(args...) {
		case False:
			return -1
		case Undecided:
			return 0
		case True:
			return 1
		}
	}
	if t(args...) != True {
		return -1
	}
	return 0
}

func (t TestExpr) True(arg Expression) d.BoolVal {
	if t.TypeFnc() == Truth || t.TypeFnc() == Trinary {
		if t(arg) == True {
			return true
		}
	}
	return false
}

func (t TestExpr) False(arg Expression) d.BoolVal {
	if t.TypeFnc() == Truth || t.TypeFnc() == Trinary {
		if t(arg) == False {
			return true
		}
	}
	return false
}

func (t TestExpr) Undecided(arg Expression) d.BoolVal {
	if t.TypeFnc() == Trinary {
		if t(arg) == Undecided {
			return true
		}
	}
	return false
}

func (t TestExpr) LT(arg Expression) d.BoolVal {
	if t.TypeFnc() == CMP {
		if t(arg) == LT {
			return true
		}
	}
	return false
}

func (t TestExpr) GT(arg Expression) d.BoolVal {
	if t.TypeFnc() == CMP {
		if t(arg) == GT {
			return true
		}
	}
	return false
}

func (t TestExpr) EQ(arg Expression) d.BoolVal {
	if t.TypeFnc() == CMP {
		if t(arg) == EQ {
			return true
		}
	}
	return false
}

//// CASE EXPRESSION CONSTRUCTOR
///
// takes a test expression and an expression to apply arguments to and return
// result from, if arguments applyed to the test expression returned true.
func NewCase(test TestExpr, expr Expression) CaseExpr {

	// generate expression to return arguments, when none has been passed
	if expr == nil {
		expr = NewGeneric(func(args ...Expression) Expression {
			switch len(args) {
			case 1:
				return args[0]
			case 2:
				return NewPair(
					args[0],
					args[1])
			}
			return NewVector(args...)
		}, "Return", Type)
	}

	// construct case type definition
	var pattern = expr.Type().(TyDef).Arguments()
	if len(pattern) == 0 {
		pattern = []Typed{Type}
	}
	var typed = Define(test.Type().(TyDef).Name(),
		expr.Type().(TyDef).Return(), pattern...)

	// construct case type name
	var ldel, rdel = "(", ")"
	var name = NewData(d.StrVal(
		ldel + expr.Type().(TyDef).Signature() + " → " +
			typed.Name() + " ⇒ " +
			expr.Type().(TyDef).Name() + " → " +
			expr.Type().(TyDef).ReturnName() + rdel,
	))

	// return constructed case expression
	return func(args ...Expression) (Expression, bool) {

		if len(args) > 0 {
			if test.Test(args...) {
				return expr.Call(args...), true
			}
			return NewVector(args...), false
		}
		// return test, expression, type & name
		return NewVector(test, expr, typed, name), false
	}
}
func (s CaseExpr) Test() TestExpr {
	var vec, _ = s()
	return vec.(VecCol)()[0].(TestExpr)
}
func (s CaseExpr) Expr() Expression {
	var vec, _ = s()
	return vec.(VecCol)()[1]
}
func (s CaseExpr) Type() Typed {
	var vec, _ = s()
	return vec.(VecCol)()[2].(Typed)
}
func (s CaseExpr) TypeName() string {
	var vec, _ = s()
	return vec.(VecCol)()[3].String()
}
func (s CaseExpr) String() string       { return s.TypeName() }
func (s CaseExpr) TypeFnc() TyFnc       { return Case }
func (s CaseExpr) TypeNat() d.TyNat     { return d.Function }
func (s CaseExpr) FlagType() d.Uint8Val { return Flag_Function.U() }
func (s CaseExpr) Call(args ...Expression) Expression {
	if result, ok := s(args...); ok {
		return result
	}
	return NewNone()
}

//// SWITCH CONSTRUCTOR ////
///
// type safe constructor wraps switch constructor that takes case arguments of
// the expression type, loops over case expression arguments once, reallocates
// as expresssion instances to pass on to the untyped private constructor.
func NewSwitch(cases ...CaseExpr) CaseSwitch {
	var exprs = make([]Expression, 0, len(cases))
	for _, cas := range cases {
		exprs = append(exprs, cas)
	}
	return conSwitch(exprs...)
}

// arbitrary typed switch constructor, to eliminate looping and reallocation of
// case expressions intendet to be stored as consumeable vector
func conSwitch(caseset ...Expression) CaseSwitch {

	return func(args ...Expression) (Expression, CaseSwitch, bool) {

		var cases = NewVector(caseset...)
		var head Expression
		var current CaseExpr
		var arguments VecCol
		var completed = NewVector()

		//// CALLED WITH ARGUMENTS ////
		///
		if len(args) > 0 {
			// safe passed arguments
			arguments = NewVector(args...)
			// if cases remain to be tested‥.
			if cases.Len() > 0 {
				// consume head & reassign caselist
				head, cases = cases.ConsumeVec()
				// cast head as case expression
				current = head.(CaseExpr)
				if result, ok := current(args...); ok {
					//// SUCCESSFUL CASE EVALUATION ////
					///
					// replenish case list for switch
					// instance reusal (in case switch gets
					// called empty to retrieve case list)
					cases = NewVector(caseset...)
					// return result, current case and
					// arguments that where passed.
					return result, conSwitch(caseset...), true
				}
				//// FAILED CASE EVALUATION ///
				///
				// add failed case and evaluated arguments to
				// the list of completed cases
				completed = completed.Append(
					NewPair(current, arguments))

				return completed,
					conSwitch(cases()...),
					false
			}
			//// CASES DEPLETION ///
			///
			// replenish cases for switch reusal and return
			// replenished switch instance for eventual reuse.
			cases = NewVector(caseset...)
			return nil, conSwitch(cases()...), false
		}
		//// CALLED WITH EMPTY ARGUMENT SET ////
		///
		// when called without arguments, return list of all defined
		// cases and log of cases completed so far, including arguments
		// that where passed to those cases for evaluation.
		return cases, conSwitch(caseset...), false
	}
}
func (s CaseSwitch) Cases() VecCol {
	var cases, _, _ = s()
	return cases.(VecCol)
}
func (s CaseSwitch) Type() Typed {
	return Define(s.TypeName(), s.TypeFnc())
}
func (s CaseSwitch) TypeName() string {
	return "[T] → (Case Switch) → (T, [T]) "
}
func (s CaseSwitch) String() string       { return s.TypeName() }
func (s CaseSwitch) TypeFnc() TyFnc       { return Switch }
func (s CaseSwitch) TypeNat() d.TyNat     { return d.Function }
func (s CaseSwitch) FlagType() d.Uint8Val { return Flag_Function.U() }

// test one set of arguments against all cases until either successful result
// is yielded, or all cases are depleted. gets called by eval & call methods.
// when called without arguments, list of all cases and list of completed
// cases, including former call arguments will be returned.
func (s CaseSwitch) TestAllCases(args ...Expression) (Expression, Expression) {
	var ok bool
	var result, caseargs Expression
	if len(args) > 0 {
		result, caseargs, ok = s(args...)
		for result != nil {
			if ok {
				return result, caseargs
			}
			result, caseargs, ok = caseargs.(CaseSwitch)(args...)
		}
		return nil, caseargs
	}
	return result, caseargs
}
func (s CaseSwitch) Call(args ...Expression) Expression {
	if len(args) > 0 {
		var result, _ = s.TestAllCases(args...)
		if result != nil {
			return result
		}
	}
	return NewNone()
}
