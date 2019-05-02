package functions

import (
	"strings"

	d "github.com/joergreinhardt/gatwd/data"
)

type (
	NoneVal    func()
	TruthExpr  func(...Callable) bool
	CaseCheck  func(...Callable) (Callable, bool)
	CaseSwitch func(...Callable) (Callable, ListVal, bool)
)

///////////////////////////////////////////////////////////////////////////////
func NewNone() NoneVal {
	return func() {}
}
func (n NoneVal) Ident() Callable           { return n }
func (n NoneVal) Maybe() bool               { return false }
func (n NoneVal) Empty() bool               { return true }
func (n NoneVal) Eval(...d.Native) d.Native { return nil }
func (n NoneVal) Value() Callable           { return nil }
func (n NoneVal) Call(...Callable) Callable { return nil }
func (n NoneVal) String() string            { return "⊥" }
func (n NoneVal) Len() int                  { return 0 }
func (n NoneVal) TypeFnc() TyFnc            { return None }
func (n NoneVal) TypeNat() d.TyNat          { return d.Nil }
func (n NoneVal) Type() Typed               { return None }
func (n NoneVal) TypeName() string          { return n.String() }

///////////////////////////////////////////////////////////////////////////////
//// TRUTH MODE
///
// truth mode flag determines if truth function will retun true only if all
// arguments evaluate true, or if any is af the arguments evaluates true.
type TruthMode bool

func (t TruthMode) String() string {
	if t {
		return "All"
	}
	return "Any"
}

const (
	All TruthMode = true
	Any TruthMode = false
)

// mode parameter is optional, default will return true if all passed arguments
// evaluate to be true.
func NewTruth(
	truth func(Callable) bool,
	mode ...TruthMode,
) TruthExpr {
	var all = All
	if len(mode) > 0 {
		all = mode[0]
	}
	if all {
		// All
		return func(args ...Callable) bool {
			for _, arg := range args {
				if !truth(arg) {
					return false
				}
			}
			return true
		}
	}
	// Any
	return func(args ...Callable) bool {
		for _, arg := range args {
			if truth(arg) {
				return true
			}
		}
		return false
	}
}
func (t TruthExpr) Call(args ...Callable) Callable {
	return NewFromData(d.BoolVal(t(args...)))
}
func (t TruthExpr) Eval(args ...d.Native) d.Native {
	return d.BoolVal(t(NatToFnc(args...)...))
}
func (t TruthExpr) Ident() Callable  { return t }
func (t TruthExpr) String() string   { return "Truth" }
func (t TruthExpr) TypeName() string { return t.String() }
func (t TruthExpr) Type() Typed      { return Truth }
func (t TruthExpr) TypeFnc() TyFnc   { return Truth }
func (t TruthExpr) TypeNat() d.TyNat { return d.Expression | d.Bool }

///////////////////////////////////////////////////////////////////////////////
//// CASE SWITCH
///
// case function represents a single case in a case switch and returns true and
// it's expression, if the passed arguments evaluate true, or false and a none
// instance otherwise.
func NewCaseFnc(expr Callable, truth TruthExpr) CaseCheck {
	return func(args ...Callable) (Callable, bool) {
		if truth(args...) {
			return expr, true
		}
		return NewNone(), false
	}
}
func (c CaseCheck) Truth() Callable {
	var _, ok = c()
	return NewFromData(d.BoolVal(ok))
}
func (c CaseCheck) Expr() Callable {
	var expr, _ = c()
	return expr
}
func (c CaseCheck) Call(args ...Callable) Callable {
	var result, _ = c(args...)
	return result
}
func (c CaseCheck) Eval(args ...d.Native) d.Native {
	return c.Expr().Eval(args...)
}
func (c CaseCheck) String() string {
	return "Case " + c.Expr().String()
}
func (c CaseCheck) Ident() Callable  { return c }
func (c CaseCheck) Type() Typed      { return Switch }
func (c CaseCheck) TypeName() string { return c.Type().String() }
func (c CaseCheck) TypeFnc() TyFnc   { return Case }
func (c CaseCheck) TypeNat() d.TyNat { return d.Expression }

///////////////////////////////////////////////////////////////////////////////
//// CASE SWITCH
///
// case-switch encloses case functions passed to it and evaluates one after
// another recursively and either returns the yielded value, an empty list of
// cases and 'true', or an instance of none, the list of remaining cases and
// 'false'.
func NewCaseSwitch(caseFncs ...CaseCheck) CaseSwitch {
	var list = NewList()
	for _, cf := range caseFncs {
		list = list.Cons(cf)
	}
	return ConsCaseSwitch(list)
}

// cons-case-switch takes a list of cases assumed to be case functions, pops
// the head applys the arguments and either returns the yielded value, an empty
// list of remaining cases and 'true', if the case evaluates true, or an
// instance of none the list of remaining cases and 'false'
func ConsCaseSwitch(cases ListVal) CaseSwitch {
	return func(args ...Callable) (Callable, ListVal, bool) {
		var head Callable
		if head, cases = cases(); head != nil {
			if check, ok := head.(CaseCheck); ok {
				if val, ok := check(args...); ok {
					return val, NewList(), true
				}
			}
		}
		return NewNone(), cases, false
	}
}
func (c CaseSwitch) Expr() Callable {
	var expr, _, _ = c()
	return expr
}
func (c CaseSwitch) Cases() Consumeable {
	var _, cases, _ = c()
	return cases
}
func (c CaseSwitch) Truth() Callable {
	var _, _, ok = c()
	return NewFromData(d.BoolVal(ok))
}
func (c CaseSwitch) Call(args ...Callable) Callable { return c.Expr().Call(args...) }
func (c CaseSwitch) Eval(args ...d.Native) d.Native { return c.Call(NatToFnc(args...)...) }
func (c CaseSwitch) String() string {
	return "Switch " + c.Expr().String()
}
func (c CaseSwitch) Ident() Callable  { return c }
func (c CaseSwitch) Type() Typed      { return Switch }
func (c CaseSwitch) TypeName() string { return c.Type().String() }
func (c CaseSwitch) TypeFnc() TyFnc   { return Switch }
func (c CaseSwitch) TypeNat() d.TyNat { return d.Expression }

///////////////////////////////////////////////////////////////////////////////
type (
	ParamFlag  func() ([]string, d.TyNat, TyFnc)
	TypeSig    func() (ParamFlag, []ListVal)
	ParamValue func(...Callable) (Callable, TypeSig)
	TypeCheck  func(...ParamFlag) (ParamFlag, bool)
)

///////////////////////////////////////////////////////////////////////////////
//// TYPE CHECK
///
// type check constructor expects a case-check expression as argument, that
// takes intances of type param-flags as it's arguments and also returns a
// value of that type.
// CaseCheck  func(...Callable) (Callable, bool)

///////////////////////////////////////////////////////////////////////////////
func NewParamFlag(nat d.TyNat, fnc TyFnc, names ...string) ParamFlag {
	var str = []string{}

	if len(names) > 0 {
		for _, name := range names {
			str = append(str, name)
		}
	}

	for _, n := range nat.Flag().Decompose() {
		str = append(str, d.TyNat(n.Flag()).String())
	}

	for _, f := range fnc.Flag().Decompose() {
		str = append(str, TyFnc(f.Flag()).String())
	}

	return func() ([]string, d.TyNat, TyFnc) {
		return str, nat, fnc
	}
}

func NewParamFlagFromExpression(expr Callable, name ...string) ParamFlag {
	return NewParamFlag(expr.TypeNat(), expr.TypeFnc(), name...)
}

func (f ParamFlag) Ident() Callable { return f }
func (f ParamFlag) Type() (d.TyNat, TyFnc) {
	return d.Flag | f.TypeNat(), HigherOrder | f.TypeFnc()
}
func (f ParamFlag) String() string {
	var name = f.TypeName()
	return name
}
func (f ParamFlag) TypeName() string {
	var names, _, _ = f()
	return strings.Join(names, "·")
}
func (f ParamFlag) AtomicName() bool {
	// name has at least type-native and type-functional elements. anything
	// greater than this, is not an atomic types name.
	if len(strings.Split(f.TypeName(), "·")) > 2 {
		return false
	}
	return true
}
func (f ParamFlag) TypeNat() d.TyNat {
	var _, nat, _ = f()
	return nat
}
func (f ParamFlag) AtomicNat() bool {
	if len(f.TypeNat().Flag().Decompose()) > 1 {
		return false
	}
	return true
}
func (f ParamFlag) NativeElements() []Typed {
	var flags = []Typed{}
	for _, flag := range f.TypeNat().Flag().Decompose() {
		flags = append(flags, flag)
	}
	return flags
}
func (f ParamFlag) TypeFnc() TyFnc {
	var _, _, fnc = f()
	return fnc
}
func (f ParamFlag) AtomicFnc() bool {
	if len(f.TypeFnc().Flag().Decompose()) > 1 {
		return false
	}
	return true
}
func (f ParamFlag) FunctionalElements() []Typed {
	var flags = []Typed{}
	for _, flag := range f.TypeFnc().Flag().Decompose() {
		flags = append(flags, flag)
	}
	return flags
}
func (f ParamFlag) Atomic() bool {
	return f.AtomicName() && f.AtomicNat() && f.AtomicFnc()
}
func (f ParamFlag) Eval(args ...d.Native) d.Native {
	return d.StrVal(f.String())
}
func (f ParamFlag) Call(args ...Callable) Callable {
	return NewFromData(f.Eval())
}

///////////////////////////////////////////////////////////////////////////////
//// parse-nat
///
// parses a native type flag to yield all it's sub types signatures. takes an
// optional list of signatures as second argument and returns a type signature
// that has been composed to represent the corresponding product-, and/or
// sum-type.
func parseNat(flag d.TyNat) TypeSig {
	var elems = flag.Flag().Decompose()
	if len(elems) > 1 {
		switch {
		case d.Pair.Flag().Match(flag.Flag()):
			flag = d.TyNat(flag.Flag().Mask(d.Pair).Flag())
			elems = flag.Flag().Decompose()
			if len(elems) == 2 {
				return NewSumTypeSignature(
					NewParamFlag(d.Pair, Native),
					parseNat(d.TyNat(elems[0].Flag())),
					parseNat(d.TyNat(elems[1].Flag())),
				)
			}
		case d.Slice.Flag().Match(flag.Flag()):
			flag = d.TyNat(flag.Flag().Mask(d.Slice).Flag())
			return NewProductTypeSignature(
				NewFlatTypeSignature(
					NewParamFlag(d.Slice, Native),
				),
				parseNat(flag),
			)
		case d.Map.Flag().Match(flag.Flag()):
			flag = d.TyNat(flag.Flag().Mask(d.Map).Flag())
			return NewProductTypeSignature(
				NewFlatTypeSignature(
					NewParamFlag(d.Map, Native),
				),
				parseNat(flag),
			)
		default:
			// masks biggest element from flag, since output of
			// flag decomposition is sorted
			flag = d.TyNat(flag.Flag().Mask(elems[len(elems)-1].Flag()).Flag())
			return NewProductTypeSignature(
				NewFlatTypeSignature(
					NewParamFlag(d.Map, Native),
				),
				parseNat(flag),
			)
		}
	}
	return NewFlatTypeSignature(NewParamFlag(d.TyNat(elems[0].Flag()), Native))
}

func parseFnc(flag TyFnc, sigs ...TypeSig) TypeSig {
	var elems = flag.Flag().Decompose()
	if len(elems) > 1 {
		switch {
		case Pair.Flag().Match(flag.Flag()):
			flag = TyFnc(flag.Flag().Mask(Pair).Flag())
			elems = flag.Flag().Decompose()
			if len(elems) == 2 {
				return NewSumTypeSignature(
					NewParamFlag(d.Data, Pair),
					parseFnc(TyFnc(elems[0].Flag())),
					parseFnc(TyFnc(elems[1].Flag())),
				)
			}
		case Tuple.Flag().Match(flag.Flag()):
			flag = TyFnc(flag.Flag().Mask(Tuple).Flag())
			elems = flag.Flag().Decompose()
			for _, elem := range elems {
				sigs = append(sigs, parseFnc(TyFnc(elem.Flag())))
			}
			return NewSumTypeSignature(NewParamFlag(d.Data, Tuple), sigs...)
		case Record.Flag().Match(flag.Flag()):
			flag = TyFnc(flag.Flag().Mask(Record).Flag())
			elems = flag.Flag().Decompose()
			for _, elem := range elems {
				sigs = append(sigs, parseFnc(TyFnc(elem.Flag())))
			}
			return NewSumTypeSignature(NewParamFlag(d.Data, Record), sigs...)
		case Enum.Flag().Match(flag.Flag()):
			flag = TyFnc(flag.Flag().Mask(Enum).Flag())
			elems = flag.Flag().Decompose()
			for _, elem := range elems {
				sigs = append(sigs, parseFnc(TyFnc(elem.Flag())))
			}
			return NewSumTypeSignature(NewParamFlag(d.Data, Enum), sigs...)
		case Set.Flag().Match(flag.Flag()):
			flag = TyFnc(flag.Flag().Mask(Set).Flag())
			elems = flag.Flag().Decompose()
			for _, elem := range elems {
				sigs = append(sigs, parseFnc(TyFnc(elem.Flag())))
			}
			return NewProductTypeSignature(
				NewFlatTypeSignature(
					NewParamFlag(d.Data, Set),
				),
				sigs...,
			)
		case List.Flag().Match(flag.Flag()):
			flag = TyFnc(flag.Flag().Mask(List).Flag())
			elems = flag.Flag().Decompose()
			for _, elem := range elems {
				sigs = append(sigs, parseFnc(TyFnc(elem.Flag())))
			}
			return NewProductTypeSignature(
				NewFlatTypeSignature(
					NewParamFlag(d.Data, List),
				),
				sigs...,
			)
		case Vector.Flag().Match(flag.Flag()):
			flag = TyFnc(flag.Flag().Mask(Vector).Flag())
			elems = flag.Flag().Decompose()
			for _, elem := range elems {
				sigs = append(sigs, parseFnc(TyFnc(elem.Flag())))
			}
			return NewProductTypeSignature(
				NewFlatTypeSignature(
					NewParamFlag(d.Data, Vector),
				),
				sigs...,
			)
		default:
			flag = TyFnc(flag.Flag().Mask(elems[len(elems)-1].Flag()).Flag())
			elems = flag.Flag().Decompose()
			for _, elem := range elems {
				sigs = append(sigs, parseFnc(TyFnc(elem.Flag())))
			}
			return NewSumTypeSignature(
				NewParamFlag(d.Expression, Expression),
				sigs...,
			)
		}
	}
	return NewSumTypeSignature(NewParamFlag(d.Data, TyFnc(elems[0].Flag())), sigs...)
}

func NewFlatTypeSignature(flag ParamFlag) TypeSig {
	return func() (ParamFlag, []ListVal) {
		return flag, []ListVal{NewList()}
	}
}

func NewSumTypeSignature(flag ParamFlag, sigs ...TypeSig) TypeSig {
	var list = NewList()
	for _, sig := range sigs {
		list = list.Cons(sig)
	}
	return func() (ParamFlag, []ListVal) {
		return flag, []ListVal{list}
	}
}

func NewSumTyoeFromSignature(flag TypeSig, sigs ...TypeSig) TypeSig {
	var list ListVal
	if !flag.Atomic() {
		list = flag.Sum()
	} else {
		list = NewList()
	}
	for _, sig := range sigs {
		list = list.Cons(sig)
	}
	return func() (ParamFlag, []ListVal) {
		return flag.Flag(), []ListVal{list, flag.Product()}
	}
}

func NewProductTypeSignature(flag TypeSig, sigs ...TypeSig) TypeSig {
	var elems, subs ListVal
	if !flag.Atomic() {
		elems = flag.Sum()
	} else {
		elems = NewList()
	}
	if !flag.Flat() {
		subs = flag.Product()
	} else {
		subs = NewList()
	}
	for _, sig := range sigs {
		subs = subs.Cons(sig)
	}
	return func() (ParamFlag, []ListVal) {
		return flag.Flag(), []ListVal{elems, subs}
	}
}

func (f TypeSig) Ident() Callable { return f }
func (f TypeSig) Flag() ParamFlag {
	var flag, _ = f()
	return flag
}
func (f TypeSig) Sum() ListVal {
	var _, lists = f()
	return lists[0]
}
func (f TypeSig) Atomic() bool {
	return f.Sum().Empty()
}
func (f TypeSig) Product() ListVal {
	var _, lists = f()
	if len(lists) > 1 {
		return lists[1]
	}
	return NewList()
}
func (f TypeSig) Flat() bool {
	var _, lists = f()
	if len(lists) > 1 {
		return false
	}
	return true
}
func (f TypeSig) Len() int {
	return f.Sum().Len()
}
func (f TypeSig) Type() (d.TyNat, TyFnc) {
	return f.Flag().Type()
}
func (f TypeSig) TypeName() string { return f.Flag().TypeName() }
func (f TypeSig) TypeNat() d.TyNat {
	return f.Flag().TypeNat()
}
func (f TypeSig) TypeFnc() TyFnc {
	return f.Flag().TypeFnc()
}
func (f TypeSig) Eval(args ...d.Native) d.Native {
	return d.StrVal(f.String())
}
func (f TypeSig) Call(args ...Callable) Callable {
	return NewFromData(f.Eval())
}
func (f TypeSig) String() string {
	var str = f.Flag().TypeName()
	return str
}

///////////////////////////////////////////////////////////////////////////////
//func NewParamValue(
//	expr func(...Callable) Callable,
//	nat d.TyNat,
//	fnc TyFnc,
//	name string,
//	sigs ...TypeSig,
//) ParamValue {
//	return func(args ...Callable) (Callable, TypeSig) {
//		return expr(args...), NewTypeSignature(NewParamFlag(nat, fnc, name), sigs...)
//	}
//}
//func NewAtomicParamValue(
//	expr func(...Callable) Callable,
//	nat d.TyNat,
//	fnc TyFnc,
//	name ...string,
//) ParamValue {
//	return func(args ...Callable) (Callable, TypeSig) {
//		return expr(args...), NewTypeSignature(NewParamFlag(nat, fnc, name...))
//	}
//}
//func NewAnonymousParamValue(
//	expr func(...Callable) Callable,
//	nat d.TyNat,
//	fnc TyFnc,
//	sigs ...TypeSig,
//) ParamValue {
//	return func(args ...Callable) (Callable, TypeSig) {
//		return expr(args...), NewTypeSignature(NewParamFlag(nat, fnc), sigs...)
//	}
//}
//func (v ParamValue) Ident() Callable { return v }
//func (v ParamValue) Expr() Callable {
//	var expr, _ = v()
//	return expr
//}
//func (v ParamValue) TypeNat() d.TyNat {
//	var _, typ = v()
//	return typ.TypeNat()
//}
//func (v ParamValue) TypeFnc() TyFnc {
//	var _, typ = v()
//	return typ.TypeFnc()
//}
//func (v ParamValue) TypeName() string {
//	var _, typ = v()
//	return typ.TypeName()
//}
//func (v ParamValue) Type() (d.TyNat, TyFnc) {
//	return v.TypeNat(), v.TypeFnc()
//}
//func (v ParamValue) String() string {
//	return v.Expr().String()
//}
//func (v ParamValue) Eval(args ...d.Native) d.Native {
//	return v.Expr().Eval(args...)
//}
//func (v ParamValue) Call(args ...Callable) Callable {
//	return v.Expr().Call(args...)
//}
