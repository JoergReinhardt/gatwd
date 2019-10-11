package functions

import d "github.com/joergreinhardt/gatwd/data"

type (
	//// SEQUENCE
	SequenceType func(...Expression) (Expression, Sequential)

	//// ENUMERABLE
	EnumType func(d.Integer) (EnumVal, d.Typed, d.Typed)
	EnumVal  func(...Expression) (Expression, d.Integer, EnumType)
)

//// SEQUENCE TYPE
///
//
func NewSequence(coll Sequential) SequenceType {
	return func(args ...Expression) (Expression, Sequential) {
		if len(args) > 0 {
			return coll.Cons(args...).Consume()
		}
		return coll.Consume()
	}
}
func (s SequenceType) TypeFnc() TyFnc      { return Sequence }
func (s SequenceType) Type() TyPattern     { return s.Tail().Type() }
func (s SequenceType) TypeElem() TyPattern { return s.Head().Type() }
func (s SequenceType) Cons(elems ...Expression) Sequential {
	return SequenceType(func(args ...Expression) (Expression, Sequential) {
		if len(args) > 0 {
			return s.Cons(elems...).Cons(args...).Consume()
		}
		return s.Cons(elems...).Consume()
	})
}
func (s SequenceType) Append(elems ...Expression) Sequential {
	return SequenceType(func(args ...Expression) (Expression, Sequential) {
		if len(args) > 0 {
			return s.Append(elems...).Cons(args...).Consume()
		}
		return s.Append(elems...).Consume()
	})
}
func (s SequenceType) Call(args ...Expression) Expression {
	var head Expression
	if len(args) > 0 {
		head, _ = s(args...)
		return head
	}
	head, _ = s()
	return head
}
func (s SequenceType) Consume() (Expression, Sequential) { return s() }
func (s SequenceType) Head() Expression {
	var expr, _ = s()
	return expr
}
func (s SequenceType) Tail() Sequential {
	var _, coll = s()
	return coll
}
func (s SequenceType) String() string {
	var head, tail = s()
	return tail.Cons(head).String()
}

//// ENUM TYPE
///
//
// check argument expression to implement data.integers interface
var isInt = NewTest(func(args ...Expression) bool {
	for _, arg := range args {
		if arg.Type().Match(Data) {
			if nat, ok := args[0].(Native); ok {
				if nat.Eval().Type().Match(d.Integers) {
					continue
				}
			}
		}
		return false
	}
	return true
})

/// TODO: refactor using maybe/eitherOr types
//
// creates low/high bound type argument and lesser/greater bounds
// checks, if no bound arguments where given, they will be set to minus
// infinity to infinity and always check out true.
func createBounds(bounds ...d.Integer) (low, high d.Typed, lesser, greater func(idx d.Integer) bool) {
	if len(bounds) == 0 {
		low, high = Def(Lex_Negative, Lex_Infinite), Lex_Infinite
		lesser = func(idx d.Integer) bool { return true }
		greater = func(idx d.Integer) bool { return true }
	}
	if len(bounds) > 0 {
		var minBound = bounds[0]
		// bound argument could be instance of type big int
		if minBound.(d.Native).Type().Match(d.BigInt) {
			lesser = func(arg d.Integer) bool {
				if minBound.(*d.BigIntVal).GoBigInt().Cmp(
					arg.(*d.BigIntVal).GoBigInt()) < 0 {
					low = DefValNative(minBound.(*d.BigIntVal))
					return true
				}
				return false
			}
		} else {
			lesser = func(arg d.Integer) bool {
				if minBound.(d.Integer).Int() >
					arg.(Native).Eval().(d.Integer).Int() {
					low = DefValNative(minBound.Int())
					return true
				}
				return false
			}
		}
	}

	if len(bounds) > 1 {
		var maxBound = bounds[1].(d.Native)
		high = DefValNative(maxBound)
		if maxBound.Type().Match(d.BigInt) {
			greater = func(arg d.Integer) bool {
				if maxBound.(*d.BigIntVal).GoBigInt().Cmp(
					arg.(*d.BigIntVal).GoBigInt()) > 0 {
					return true
				}
				return false
			}
		} else {
			greater = func(arg d.Integer) bool {
				if arg.(d.Integer).Int() >
					maxBound.(d.Integer).Int() {
					return true
				}
				return false
			}
		}
	}
	return low, high, lesser, greater
}

func inBound(lesser, greater func(d.Integer) bool, ints ...d.Integer) bool {
	for _, i := range ints {
		if !lesser(i) && !greater(i) {
			return true
		}
	}
	return false
}

func NewEnumType(fnc func(...d.Integer) Expression, limits ...d.Integer) EnumType {
	var low, high, lesser, greater = createBounds(limits...)
	return func(idx d.Integer) (EnumVal, d.Typed, d.Typed) {
		return func(args ...Expression) (Expression, d.Integer, EnumType) {
			if inBound(lesser, greater, idx) {
				if len(args) > 0 {
					return fnc(idx).Call(args...), idx, NewEnumType(fnc, limits...)
				}
				return fnc(idx), idx, NewEnumType(fnc, limits...)
			}
			return NewNone(), idx, NewEnumType(fnc, limits...)
		}, low, high
	}
}
func (e EnumType) Expr() Expression {
	var expr, _, _ = e(d.IntVal(0))
	return expr
}
func (e EnumType) Limits() (min, max d.Typed) {
	_, min, max = e(d.IntVal(0))
	return min, max
}
func (e EnumType) Low() d.Typed {
	var min, _ = e.Limits()
	return min
}
func (e EnumType) High() d.Typed {
	var _, max = e.Limits()
	return max
}
func (e EnumType) InBound(ints ...d.Integer) bool {
	var _, _, lesser, greater = createBounds(
		e.Low().(d.Integer),
		e.High().(d.Integer),
	)
	return inBound(lesser, greater, ints...)
}
func (e EnumType) Null() Expression {
	var result, _, _ = e(d.IntVal(0))
	return result
}
func (e EnumType) Unit() Expression {
	var result, _, _ = e(d.IntVal(1))
	return result
}
func (e EnumType) Type() TyPattern { return Def(Enum, e.Unit().Type()) }
func (e EnumType) TypeFnc() TyFnc  { return Enum | e.Unit().TypeFnc() }
func (e EnumType) String() string  { return e.Type().TypeName() }
func (e EnumType) Call(args ...Expression) Expression {
	if len(args) > 0 {
		return e.Expr().Call(args...)
	}
	return e.Expr().Call()
}

//// ENUM VALUE
///
//
func (e EnumVal) Expr() Expression {
	var expr, _, _ = e()
	return expr
}
func (e EnumVal) Index() d.Integer {
	var _, idx, _ = e()
	return idx
}
func (e EnumVal) EnumType() EnumType {
	var _, _, et = e()
	return et
}
func (e EnumVal) Next() EnumVal {
	var result, _, _ = e.EnumType()(e.Index().Int() + d.IntVal(1))
	return result
}
func (e EnumVal) Previous() EnumVal {
	var result, _, _ = e.EnumType()(e.Index().Int() - d.IntVal(1))
	return result
}
func (e EnumVal) String() string                     { return e.Expr().String() }
func (e EnumVal) Type() TyPattern                    { return e.EnumType().Type() }
func (e EnumVal) TypeFnc() TyFnc                     { return e.EnumType().TypeFnc() }
func (e EnumVal) Call(args ...Expression) Expression { return e.Expr().Call(args...) }
