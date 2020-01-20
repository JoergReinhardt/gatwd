/*

PRODUCT TYPES
-------------
*/
package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

type (
	//// BOOLEAN ALGEBRA
	///
	// BOOL VALUE TYPE
	Bool bool

	// TEST & COMPARE
	Test    Def
	Compare Def

	// BOOL OPERATORS
	NOT Def
	AND Def
	XOR Def
	OR  Def

	// MAYBE (JUST | NONE)
	Maybe Def

	// CASE & SWITCH
	Switch ListVal

	// ALTERNATETIVES TYPE (EITHER | OR)
	EitherOr Switch
)

//// TRUTH VALUE
///
// truth value aliases the native bool type & returns its function type as
// either 'True', or 'False' depending on the aliased instance
func (b Bool) TypeFnc() TyFnc {
	if b {
		return True
	}
	return False
}
func (b Bool) Type() Decl {
	return Declare(b.TypeFnc())
}
func (b Bool) Call(...Functor) Functor {
	return Box(d.BoolVal(b))
}
func (b Bool) String() string {
	if b {
		return "True"
	}
	return "False"
}

//// TEST
///
// test takes a function that takes two functors to scrutinize and returns a
// boolean value to indicate test result.
func NewTest(
	atype d.Typed,
	test func(a, b Functor) bool,
) Test {
	return Test(Define(Lambda(func(args ...Functor) Functor {
		return Bool(test(args[0], args[1]))
	}), DecSym("Test"), Truth, Declare(atype, atype)))
}
func (t Test) Unbox() Functor { return Def(t).Unbox() }
func (t Test) TypeFnc() TyFnc {
	return Truth
}
func (t Test) Type() Decl {
	return Declare(
		DecSym("Test"),
		Def(t).TypeRet(),
		Def(t).TypeArgs())
}
func (t Test) String() string {
	return t.TypeFnc().TypeName()
}
func (t Test) Test(a, b Functor) bool {
	return bool(Def(t).Unbox().Call(a, b).(Bool))
}
func (t Test) Compare(a, b Functor) int {
	if t.Test(a, b) {
		return 0
	}
	return -1
}
func (t Test) Call(args ...Functor) Functor { return t(args...) }
func (t Test) Equal() Def {
	return Define(t.Unbox(), Equal, Truth, t.Type().TypeArgs())
}

//// COMPARATOR
///
// comparator takes two functors to compare and returns an integer to indicate
// the result.  if both functors are considered equal by the passed comparing
// expression, zero is retuned, a negative result, if the left argument is
// lesser and a positive result, if its greater than the right argument.
func NewComparator(
	argtype d.Typed,
	comp func(a, b Functor) int,
) Compare {
	return Compare(Define(Lambda(func(args ...Functor) Functor {
		if len(args) == 0 { // return argument type, when called empty
			if Kind_Decl.Match(argtype.Kind()) {
				return argtype.(Decl)
			}
			return Declare(Comparison,
				Declare(Lesser|Greater|Equal),
				Declare(argtype, argtype))
		}
		if comp(args[0], args[1]) < 0 {
			return Declare(Lesser)
		}
		if comp(args[0], args[1]) > 0 {
			return Declare(Greater)
		}
		return Declare(Equal)
	}), DecSym("Compare"),
		Declare(Lesser|Greater|Equal),
		Declare(argtype, argtype)))
}
func (t Compare) Unbox() Functor               { return Def(t).Unbox() }
func (t Compare) Type() Decl                   { return Def(t).Type() }
func (t Compare) TypeRet() Decl                { return Def(t).TypeRet() }
func (t Compare) TypeArgs() Decl               { return Def(t).TypeArgs() }
func (t Compare) String() string               { return Def(t).TypeName() }
func (t Compare) TypeFnc() TyFnc               { return Comparison }
func (t Compare) Call(args ...Functor) Functor { return t(args...) }
func (t Compare) Compare(a, b Functor) int {
	return int(t.Unbox().Call(a, b).(Atom)().(d.IntVal))
}
func (t Compare) Equal() Def {
	return Define(Lambda(func(args ...Functor) Functor {
		return Bool(t.Call(args...).(Atom)().(d.IntVal) == 0)
	}), Equal, Truth, t.TypeArgs())
}
func (t Compare) Lesser() Def {
	return Define(Lambda(func(args ...Functor) Functor {
		return Bool(t.Call(args...).(Atom)().(d.IntVal) < 0)
	}), Lesser, Truth, t.TypeArgs())
}
func (t Compare) Greater() Def {
	return Define(Lambda(func(args ...Functor) Functor {
		return Bool(t.Call(args...).(Atom)().(d.IntVal) > 0)
	}), Greater, Truth, t.TypeArgs())
}

//// MAYBE VALUE
///
// Definitions may have an implicit return type T|⊥  when there is a case where
// the passed arguments do not match the declared argument types.  the maybe
// constructor boxes definitions by redefining them with a return type
// expressing that optionality (maybe → T|⊥) of the return type explicitly by
// returning a boxed value with return type 'Just T|⊥' as declared result type.
func NewMaybe(expr Def) Maybe {

	return Maybe(Define(Lambda(func(args ...Functor) Functor {
		if len(args) > 0 {
			var result = expr.Call(args...)
			if result.Type().Match(None) {
				return result
			}
			return Define(result,
				Declare(Just, result.Type().TypeRet()),
				result.Type().TypeRet(), expr.Type().TypeArgs())
		}
		return expr
	}),
		Declare(DecSym("Maybe"), expr.Type()),
		Declare(Option, Declare(Just, expr.TypeRet()), None),
		expr.TypeArgs()))
}

func (t Maybe) Call(args ...Functor) Functor { return t(args...) }
func (t Maybe) Unbox() Functor               { return Def(t).Unbox() }
func (t Maybe) String() string               { return Def(t).String() }
func (t Maybe) TypeArgs() Decl               { return Def(t).TypeArgs() }
func (t Maybe) TypeRet() Decl                { return Def(t).TypeRet() }
func (t Maybe) Type() Decl                   { return Def(t).Type() }
func (t Maybe) TypeFnc() TyFnc               { return Option }

/// SWITCH
//
// switch takes a slice of cases and evaluates them against its arguments to
// yield either a none value, or the result of the case application and a
// switch enclosing the remaining cases.  id all cases are depleted, a none
// instance will be returned as result and nil will be yielded instead of the
// switch value
//
// when called, a switch evaluates all it's cases until it yields either
// results from applying the first case that matched the arguments, or none.
//func NewSwitch(cases ...Def) Switch {
//	var types = make([]d.Typed, 0, len(cases))
//	for _, c := range cases {
//		types = append(types, c.Type())
//	}
//	var (
//		current Def
//		remains = cases
//		pattern = Declare(Choice, Declare(types...))
//	)
//	return func(args ...Functor) (Functor, []Def) {
//		if len(args) > 0 {
//			if remains != nil {
//				current = remains[0]
//				if len(remains) > 1 {
//					remains = remains[1:]
//				} else {
//					remains = remains[:0]
//				}
//				var result = current(args...)
//				if result.Type().Match(None) {
//					return result, remains
//				}
//				remains = cases
//				return result, cases
//			}
//			remains = cases
//			return NewNone(), cases
//		}
//		return pattern, cases
//	}
//}
//func (t Switch) Cases() []Def {
//	var _, cases = t()
//	return cases
//}
//func (t Switch) Type() Decl {
//	var pat, _ = t()
//	return pat.(Decl)
//}
//func (t Switch) String() string {
//	return t.Type().TypeName()
//}
//func (t Switch) TypeFnc() TyFnc {
//	return Choice
//}
//func (t Switch) Call(args ...Functor) Functor {
//	var (
//		cases   = t.Cases()
//		current Def
//		result  Functor
//	)
//	for len(cases) > 0 {
//		current = cases[0]
//		if len(cases) > 1 {
//			cases = cases[1:]
//		} else {
//			cases = cases[:0]
//		}
//		result = current(args...)
//		if !result.TypeFnc().Match(None) {
//			return result
//		}
//	}
//	return NewNone()
//}
