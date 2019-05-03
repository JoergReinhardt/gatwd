package functions

import (
	"strings"

	d "github.com/joergreinhardt/gatwd/data"
)

type (
	NoneVal    func()
	TruthExpr  func(...Callable) bool
	CaseExpr   func(...Callable) (Callable, bool)
	SwitchExpr func(...Callable) (Callable, ListVal, bool)
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
		return "All True"
	}
	return "Any True"
}

const (
	All TruthMode = true
	Any TruthMode = false
)

// truth function iterates over passed arguments and behaves in one of two
// modes, according to the mode parameter.  mode parameter is optional, default
// will return true if all passed arguments evaluate to be true. setting the
// parameter to 'Any', will change truth behaviour to yield false on first
// value that evaluates true
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
func NewCase(expr Callable, truth TruthExpr) CaseExpr {
	return func(args ...Callable) (Callable, bool) {
		if truth(args...) {
			return expr, true
		}
		return NewNone(), false
	}
}
func (c CaseExpr) Truth() Callable {
	var _, ok = c()
	return NewFromData(d.BoolVal(ok))
}
func (c CaseExpr) Expr() Callable {
	var expr, _ = c()
	return expr
}
func (c CaseExpr) Call(args ...Callable) Callable {
	var result, _ = c(args...)
	return result
}
func (c CaseExpr) Eval(args ...d.Native) d.Native {
	return c.Expr().Eval(args...)
}
func (c CaseExpr) String() string {
	return "Case " + c.Expr().String()
}
func (c CaseExpr) Ident() Callable  { return c }
func (c CaseExpr) Type() Typed      { return Switch }
func (c CaseExpr) TypeName() string { return c.Type().String() }
func (c CaseExpr) TypeFnc() TyFnc   { return Case }
func (c CaseExpr) TypeNat() d.TyNat { return d.Expression }

///////////////////////////////////////////////////////////////////////////////
//// CASE SWITCH
///
// case-switch encloses case functions passed to it and evaluates one after
// another recursively and either returns the yielded value, an empty list of
// cases and 'true', or an instance of none, the list of remaining cases and
// 'false'.
func NewSwitch(caseFncs ...CaseExpr) SwitchExpr {
	var args = []Callable{}
	for _, arg := range caseFncs {
		args = append(args, arg)
	}
	var list = NewList(args...)
	return ConsCaseSwitch(list)
}

// cons-case-switch takes a list of cases assumed to be case functions, pops
// the head applys the arguments and either returns the yielded value, an empty
// list of remaining cases and 'true', if the case evaluates true, or an
// instance of none the list of remaining cases and 'false'
func ConsCaseSwitch(cases ListVal) SwitchExpr {
	return func(args ...Callable) (Callable, ListVal, bool) {
		var head Callable
		if head, cases = cases(); head != nil {
			if check, ok := head.(CaseExpr); ok {
				if val, ok := check(args...); ok {
					return val, NewList(), true
				}
			}
		}
		return NewNone(), cases, false
	}
}
func (c SwitchExpr) Expr() Callable {
	var expr, _, _ = c()
	return expr
}
func (c SwitchExpr) Cases() Consumeable {
	var _, cases, _ = c()
	return cases
}
func (c SwitchExpr) Truth() Callable {
	var _, _, ok = c()
	return NewFromData(d.BoolVal(ok))
}
func (c SwitchExpr) Call(args ...Callable) Callable { return c.Expr().Call(args...) }
func (c SwitchExpr) Eval(args ...d.Native) d.Native { return c.Call(NatToFnc(args...)...) }
func (c SwitchExpr) String() string {
	return "Switch " + c.Expr().String()
}
func (c SwitchExpr) Ident() Callable  { return c }
func (c SwitchExpr) Type() Typed      { return Switch }
func (c SwitchExpr) TypeName() string { return c.Type().String() }
func (c SwitchExpr) TypeFnc() TyFnc   { return Switch }
func (c SwitchExpr) TypeNat() d.TyNat { return d.Expression }

///////////////////////////////////////////////////////////////////////////////
type (
	TypeFlag  func() ([]string, d.TyNat, TyFnc)
	TypeCheck func(args ...Callable) (Callable, bool)
	TypeSig   func() (TypeFlag, []ListVal)
	TypeCons  func() (Callable, TypeSig)
	ParamType func() (TypeFlag, SwitchExpr)
)

///////////////////////////////////////////////////////////////////////////////
//// TYPE CHECK
func NewTypeCheck(expr Callable, check TruthExpr) TypeCheck {
	return TypeCheck(NewCase(expr, check))
}
func (c TypeCheck) Expression() Callable {
	var expr, _ = c()
	return expr
}
func (c TypeCheck) Ok() bool {
	var _, ok = c()
	return ok
}
func (c TypeCheck) String() string                 { return c.Expression().String() }
func (c TypeCheck) TypeNat() d.TyNat               { return c.Expression().TypeNat() }
func (c TypeCheck) TypeFnc() TyFnc                 { return c.Expression().TypeFnc() }
func (c TypeCheck) Call(args ...Callable) Callable { return c.Expression().Call(args...) }
func (c TypeCheck) Eval(args ...d.Native) d.Native { return c.Expression().Eval(args...) }

///////////////////////////////////////////////////////////////////////////////
func NewTypeFlag(nat d.TyNat, fnc TyFnc, names ...string) TypeFlag {

	return func() ([]string, d.TyNat, TyFnc) {
		return names, nat, fnc
	}
}

func NewFlagFromExpr(expr Callable, name ...string) TypeFlag {
	return NewTypeFlag(expr.TypeNat(), expr.TypeFnc(), name...)
}

func (f TypeFlag) Ident() Callable { return f }
func (f TypeFlag) Type() (d.TyNat, TyFnc) {
	return d.Flag | f.TypeNat(), HigherOrder | f.TypeFnc()
}
func (f TypeFlag) String() string {
	var names = f.TypeName()
	for _, flag := range f.TypeNat().Flag().Decompose() {
		names = append(names, d.TyNat(flag.Flag()).String())
	}

	for _, flag := range f.TypeFnc().Flag().Decompose() {
		names = append(names, TyFnc(flag.Flag()).String())
	}
	return strings.Join(names, "·")
}
func (f TypeFlag) TypeName() []string {
	var names, _, _ = f()
	return names
}
func (f TypeFlag) AtomicName() bool {
	// name has at least type-native and type-functional elements. anything
	// greater than this, is not an atomic types name.
	if len(f.TypeName()) > 0 {
		return false
	}
	return true
}
func (f TypeFlag) TypeNat() d.TyNat {
	var _, nat, _ = f()
	return nat
}
func (f TypeFlag) AtomicNat() bool {
	if len(f.TypeNat().Flag().Decompose()) > 1 {
		return false
	}
	return true
}
func (f TypeFlag) NativeElements() []Typed {
	var flags = []Typed{}
	for _, flag := range f.TypeNat().Flag().Decompose() {
		flags = append(flags, flag)
	}
	return flags
}
func (f TypeFlag) TypeFnc() TyFnc {
	var _, _, fnc = f()
	return fnc
}
func (f TypeFlag) AtomicFnc() bool {
	if len(f.TypeFnc().Flag().Decompose()) > 1 {
		return false
	}
	return true
}
func (f TypeFlag) FunctionalElements() []Typed {
	var flags = []Typed{}
	for _, flag := range f.TypeFnc().Flag().Decompose() {
		flags = append(flags, flag)
	}
	return flags
}
func (f TypeFlag) Atomic() bool {
	return f.AtomicName() && f.AtomicNat() && f.AtomicFnc()
}
func (f TypeFlag) Eval(args ...d.Native) d.Native {
	return d.StrVal(f.String())
}
func (f TypeFlag) Call(args ...Callable) Callable {
	return NewFromData(f.Eval())
}

///////////////////////////////////////////////////////////////////////////////
//// parse-nat
///
func flagDecap(flag d.BitFlag) (head, tail d.BitFlag) {
	if flag.Count() > 0 {
		if flag.Count() > 1 {
			var flags = flag.Decompose()
			var head = flags[len(flags)-1].(d.BitFlag)
			flag = flag.Mask(head).Flag()
			return head, flag
		}
		return head, None.Flag()
	}
	return None.Flag(), None.Flag()
}

func parseNat(flag d.BitFlag) d.Native {
	if flag.Flag().Count() > 1 {
		var head d.BitFlag
		head, flag = flagDecap(flag)
		switch {
		case head.Match(d.Multiples):
			return d.NewPair(d.TyNat(head.Flag()), parseNat(flag))
		case head.Match(d.SideEffects):
			return d.NewPair(d.TyNat(head.Flag()), parseNat(flag))
		case head.Match(d.Pair):
			var left, right d.BitFlag = d.Nil.Flag(), d.Nil.Flag()
			if flag.Flag().Count() > 0 {
				var flags = flag.Decompose()
				right = flags[0].(d.BitFlag)
				flag = flag.Mask(right).Flag()
				if flag.Flag().Count() > 0 {
					flags = flag.Decompose()
					left = flags[0].(d.BitFlag)
					flag = flag.Mask(left).Flag()
				}
			}
			return d.NewPair(
				d.Pair,
				d.NewPair(
					parseNat(left),
					parseNat(right),
				))
		}
	}
	return d.ConEmptyFromFlag(flag.Flag())
}

func parseOptions(flag d.BitFlag) TypeCons {
	if flag.Count() > 0 {
		var head d.BitFlag
		head, flag = flagDecap(flag)
		switch {
		case head.Match(Truth):
		case head.Match(Ordered):
		case head.Match(Maybe):
		case head.Match(CaseSwitch):
		case head.Match(Alternatives):
		case head.Match(Branch):
		case head.Match(Continue):
		}
	}
	return parseFnc(flag)
}

func parseParameters(flag d.BitFlag) TypeCons {
	if flag.Count() > 1 {
		var head d.BitFlag
		head, flag = flagDecap(flag)
		switch {
		case head.Match(CallArity):
		case head.Match(CallPropertys):
		}
	}
	// return single unparsable flag
	return parseFnc(flag)
}

func parseFunctors(flag d.BitFlag) TypeCons {
	if flag.Count() > 1 {
		var head d.BitFlag
		head, flag = flagDecap(flag)
		switch {
		case head.Match(Functor):
		case head.Match(Applicable):
		case head.Match(Operator):
		case head.Match(Monad):
		}
	}
	// return single unparsable flag
	return parseFnc(flag)
}

//func parseProdType(head, flag d.BitFlag) TypeCons {
//}
//func parseSumType(option, flag d.BitFlag) TypeCons {
//}

// collections are product types containing elements of a specific subtype
func parseCollections(flag d.BitFlag) TypeCons {
	if flag.Count() > 0 {
		var expr Callable
		var head d.BitFlag
		head, flag = flagDecap(flag)
		switch {
		case head.Match(Pair):
			var left, right d.BitFlag
			if flag.Flag().Count() > 0 {
				var flags = flag.Decompose()
				right = flags[0].(d.BitFlag)
				flag = flag.Mask(right).Flag()
				if flag.Flag().Count() > 0 {
					flags = flag.Decompose()
					left = flags[0].(d.BitFlag)
					flag = flag.Mask(left).Flag()
				}
			}
			expr = NewPair(
				parseFnc(left),
				parseFnc(right),
			)
			return NewTypeCons(
				expr,
				NewProductTypeSignature(
					NewSumTypeSignature(
						NewTypeFlag(d.Pair, Pair|TyFnc(left|right)),
						parseFnc(left),
						parseFnc(right),
					),
					parseFnc(flag),
				),
			)
		case head.Match(List):
			expr = NewNaryExpr(func(args ...Callable) Callable {
				return NewList(args...)
			})
			return NewTypeCons(
				expr,
				NewProductTypeSignature(
					NewTypeSignature(expr.TypeNat(), List),
					parseFnc(flag),
				),
			)
		case head.Match(Vector):
			expr = NewNaryExpr(func(args ...Callable) Callable {
				return NewVector(args...)
			})
			return NewTypeCons(
				expr,
				NewProductTypeSignature(
					NewTypeSignature(expr.TypeNat(), Vector),
					parseFnc(flag),
				),
			)
		case head.Match(Tuple):
			expr = NewNaryExpr(func(args ...Callable) Callable {
				return NewTuple(args...)
			})
			return NewTypeCons(
				expr,
				NewProductTypeSignature(
					NewTypeSignature(expr.TypeNat(), Tuple),
					parseFnc(flag),
				),
			)
		case head.Match(Set):
			var pair Paired
			var ok bool
			if pair, ok = parseFnc(flag).Expression().(Paired); ok {
				expr = NewNaryExpr(func(args ...Callable) Callable {
					if len(args) > 0 {
						if len(args) > 1 {
							return NewPair(args[0], args[1])
						}
						return NewPair(args[0], NewNone())
					}
					return NewPair(NewNone(), NewNone())
				})
			}
			expr = NewAssocSet(pair)
			return NewTypeCons(
				expr,
				NewProductTypeSignature(
					NewTypeSignature(expr.TypeNat(), Set),
					parseFnc(flag),
				),
			)
		case head.Match(Enum):
			var pair Paired
			var ok bool
			if pair, ok = parseFnc(flag).Expression().(Paired); ok {
				expr = NewNaryExpr(func(args ...Callable) Callable {
					if len(args) > 0 {
						if len(args) > 1 {
							return NewPair(args[0], args[1])
						}
						return NewPair(args[0], NewNone())
					}
					return NewPair(NewNone(), NewNone())
				})
			}
			expr = NewAssocSet(pair)
			return NewTypeCons(
				expr,
				NewProductTypeSignature(
					NewTypeSignature(expr.TypeNat(), Enum),
					parseFnc(flag),
				),
			)
		case head.Match(Record):
			var pair Paired
			var ok bool
			if pair, ok = parseFnc(flag).Expression().(Paired); ok {
				expr = NewNaryExpr(func(args ...Callable) Callable {
					if len(args) > 0 {
						if len(args) > 1 {
							return NewPair(args[0], args[1])
						}
						return NewPair(args[0], NewNone())
					}
					return NewPair(NewNone(), NewNone())
				})
			}
			expr = NewAssocSet(pair)
			return NewTypeCons(
				expr,
				NewProductTypeSignature(
					NewTypeSignature(expr.TypeNat(), Record),
					parseFnc(flag),
				),
			)
		}
	}
	return parseFnc(flag)
}

// kinds returns native, data and functional expression constructor instances.
func parseKinds(flag d.BitFlag) TypeCons {
	if flag.Count() > 0 {
		var expr Callable
		var head d.BitFlag
		head, flag = flagDecap(flag)
		switch {
		case head.Match(Type):
			expr = parseFnc(head).Expression()
			return NewTypeCons(
				expr,
				NewProductTypeSignature(
					NewTypeSignature(d.Expression, Expression),
					parseFnc(flag),
				))
		case head.Match(Native):
			expr = NativeVal(New)
			return NewTypeCons(
				expr,
				NewProductTypeSignature(
					NewTypeSignature(d.Literal, Native),
					parseFnc(flag),
				))
		case head.Match(Data):
			expr = DataVal(d.NewFromData)
			return NewTypeCons(
				expr,
				NewProductTypeSignature(
					NewTypeSignature(d.Data, Data),
					parseFnc(flag),
				),
			)
		case head.Match(Expression):
			expr = NewNaryExpr(func(exprs ...Callable) Callable {
				if len(exprs) > 0 {
					if len(exprs) > 1 {
						return exprs[0].Call(exprs[1:]...)
					}
				}
				return exprs[0]
			})
			return NewTypeCons(
				expr,
				NewProductTypeSignature(
					NewTypeSignature(d.Expression, Expression),
					parseFnc(flag),
				),
			)
		}
	}
	// return single unparsable flag
	return parseFnc(flag)
}

func parseFnc(flag d.BitFlag) TypeCons {
	var head d.BitFlag
	head, flag = flagDecap(flag)
	if flag.Count() > 1 {
		switch {
		case head.Match(Kinds):
			return parseKinds(flag)
		case head.Match(Parameters):
			return parseParameters(flag)
		case head.Match(Collections):
			return parseCollections(flag)
		case head.Match(Options):
			return parseOptions(flag)
		case head.Match(Functors):
			return parseFunctors(flag)

		}
	}
	// flag could not be parsed‥. return flag
	var expr = NewFromData(TyFnc(flag))
	return NewTypeCons(
		expr,
		NewTypeSignatureFromExpr(expr))
}

func parseExpr(expr Callable) TypeCons {
	var flag = expr.TypeFnc().Flag()
	if flag.Flag().Count() > 1 {
		expr = parseFnc(flag)
	}
	// type flag is atomic
	return NewTypeCons(
		expr,
		NewTypeSignatureFromExpr(expr))
}
func NewTypeConsAlias(expr Callable, name string) TypeCons {
	return func() (Callable, TypeSig) {
		return expr, NewFlatTypeSignature(NewTypeFlag(
			expr.TypeNat(), HigherOrder, name))
	}
}
func NewTypeCons(expr Callable, sig TypeSig) TypeCons {
	return func() (Callable, TypeSig) { return expr, sig }
}
func (c TypeCons) Expression() Callable {
	var expr, _ = c()
	return expr
}
func (c TypeCons) Signature() TypeSig {
	var _, sig = c()
	return sig
}
func (c TypeCons) String() string                 { return c.Signature().String() }
func (c TypeCons) TypeNat() d.TyNat               { return c.Signature().TypeNat() }
func (c TypeCons) TypeFnc() TyFnc                 { return c.Signature().TypeFnc() }
func (c TypeCons) Call(args ...Callable) Callable { return c.Expression().Call(args...) }
func (c TypeCons) Eval(args ...d.Native) d.Native { return c.Expression().Eval(args...) }

func NewTypeSignatureFromExpr(expr Callable) TypeSig {
	var sum, product ListVal
	return func() (TypeFlag, []ListVal) {
		return NewTypeFlag(expr.TypeNat(), expr.TypeFnc()), []ListVal{sum, product}
	}
}
func NewTypeSignature(nat d.TyNat, fnc TyFnc) TypeSig {
	var sum, product ListVal
	return func() (TypeFlag, []ListVal) {
		return NewTypeFlag(nat, fnc), []ListVal{sum, product}
	}
}
func NewFlatTypeSignature(flag TypeFlag) TypeSig {
	return func() (TypeFlag, []ListVal) {
		return flag, []ListVal{NewList()}
	}
}

func NewSumTypeSignature(flag TypeFlag, sigs ...TypeCons) TypeSig {
	var list = NewList()
	for _, sig := range sigs {
		list = list.Cons(sig)
	}
	return func() (TypeFlag, []ListVal) {
		return flag, []ListVal{list}
	}
}

func NewSumTypeFromSignature(flag TypeSig, sigs ...TypeCons) TypeSig {
	var list ListVal
	if !flag.Atomic() {
		list = flag.Sum()
	} else {
		list = NewList()
	}
	for _, sig := range sigs {
		list = list.Cons(sig)
	}
	return func() (TypeFlag, []ListVal) {
		return flag.Flag(), []ListVal{list, flag.Product()}
	}
}

func NewProductTypeSignature(flag TypeSig, sigs ...TypeCons) TypeSig {
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
	return func() (TypeFlag, []ListVal) {
		return flag.Flag(), []ListVal{elems, subs}
	}
}

func (f TypeSig) Ident() Callable { return f }
func (f TypeSig) Flag() TypeFlag {
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
func (f TypeSig) TypeName() []string { return f.Flag().TypeName() }
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
	return f.Flag().String()
}
