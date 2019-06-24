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
	NoneVal func()

	//// TRUTH VALUES
	TruthVal   func() TyFnc
	TrinaTruth func() TyFnc

	//// DATA VALUES
	Native     func() d.Native
	NativePair func() d.Paired
	NativeSet  func() d.Mapped
	NativeCol  func() d.Sliceable
	NatUnboxed func() d.Sliceable

	//// STATIC EXPRESSIONS
	ConstEq    func() Callable
	UnaryEq    func(Callable) Callable
	BinaryEq   func(l, r Callable) Callable
	VariadicEq func(...Callable) Callable
	NaryEq     func(...Callable) Callable
)

//// NONE VALUE
///
// none representing the abscence of a value of any type.  implements
// consumeable, key-, index & generic pair interface to be returneable as such.
func NewNone() NoneVal { return func() {} }

func (n NoneVal) Ident() Callable                  { return n }
func (n NoneVal) Len() int                         { return 0 }
func (n NoneVal) String() string                   { return "⊥" }
func (n NoneVal) Eval(args ...d.Native) d.Native   { return nil }
func (n NoneVal) Call(...Callable) Callable        { return nil }
func (n NoneVal) Key() Callable                    { return nil }
func (n NoneVal) Index() Callable                  { return nil }
func (n NoneVal) Left() Callable                   { return nil }
func (n NoneVal) Right() Callable                  { return nil }
func (n NoneVal) Both() Callable                   { return nil }
func (n NoneVal) Value() Callable                  { return nil }
func (n NoneVal) Empty() bool                      { return true }
func (n NoneVal) TypeFnc() TyFnc                   { return None }
func (n NoneVal) TypeNat() d.TyNat                 { return d.Nil }
func (n NoneVal) TypeName() string                 { return n.String() }
func (n NoneVal) Head() Callable                   { return NewNone() }
func (n NoneVal) Tail() Consumeable                { return NewNone() }
func (n NoneVal) Consume() (Callable, Consumeable) { return NewNone(), NewNone() }

//// TRUTH VALUE
func NewTruth(truth ...bool) TruthVal {
	return func() TyFnc {
		if truth[0] {
			return True
		}
		return False
	}
}
func (t TruthVal) Call(...Callable) Callable      { return t }
func (t TruthVal) TypeFnc() TyFnc                 { return t() }
func (t TruthVal) TypeNat() d.TyNat               { return d.Function }
func (t TruthVal) TypeName() string               { return t().TypeName() }
func (t TruthVal) String() string                 { return t().TypeName() }
func (t TruthVal) Trinary() TrinaTruth            { return NewTrinaryTruth(t.Int()) }
func (t TruthVal) Eval(args ...d.Native) d.Native { return d.BoolVal(t.Bool()) }
func (t TruthVal) True() d.Native                 { return t.Eval() }
func (t TruthVal) Int() int {
	if t().Match(True) {
		return 1
	}
	return -1
}
func (t TruthVal) Bool() bool {
	if t().Match(Truth) {
		return true
	}
	return false
}

//// TRINARY TRUTH VALUE
func NewTrinaryTruth(truth int) TrinaTruth {
	return func() TyFnc {
		if truth > 0 {
			return True
		}
		if truth < 0 {
			return False
		}
		return Undecided
	}
}
func (t TrinaTruth) Call(...Callable) Callable { return t }
func (t TrinaTruth) TypeFnc() TyFnc            { return t() }
func (t TrinaTruth) TypeNat() d.TyNat          { return d.Function }
func (t TrinaTruth) TypeName() string          { return t().TypeName() }
func (t TrinaTruth) String() string            { return t().TypeName() }
func (t TrinaTruth) Truth() TruthVal           { return NewTruth(t.Bool()) }
func (t TrinaTruth) Eval(args ...d.Native) d.Native {
	if t().Match(Truth) {
		return d.IntVal(1)
	}
	if t().Match(False) {
		return d.IntVal(-1)
	}
	return d.IntVal(0)
}
func (t TrinaTruth) True() d.Native {
	if t().Match(True) {
		return d.BoolVal(true)
	}
	if t().Match(False) {
		return d.BoolVal(true)
	}
	return d.NewNil()
}
func (t TrinaTruth) Int() int {
	if t().Match(True) {
		return 1
	}
	if t().Match(False) {
		return 1
	}
	return 0
}
func (t TrinaTruth) Bool() bool {
	if t().Match(False) {
		return false
	}
	return true
}

//// NATIVE VALUE
///
// data value implements the callable interface but returns an instance of
// data/Value. the eval method of every native can be passed as argument
// instead of the value itself, as in 'DataVal(native.Eval)', to delay, or even
// possibly ommit evaluation of the underlying data value for cases where
// lazynes is paramount.
func New(inf ...interface{}) Callable { return NewNative(d.New(inf...)) }

func NewNative(args ...d.Native) Callable {
	// if any initial arguments have been passed
	if len(args) > 0 {

		var nat = d.NewFromData(args...)
		var tnat = nat.TypeNat()

		switch {
		case d.Slice.Match(tnat):
			if slice, ok := nat.(d.Sliceable); ok {
				return NativeCol(func() d.Sliceable {
					return slice
				})
			}
		case d.Unboxed.Match(tnat):
			if slice, ok := nat.(d.Sliceable); ok {
				return NativeCol(func() d.Sliceable {
					return slice
				})
			}
		case d.Pair.Match(tnat):
			if pair, ok := nat.(d.Paired); ok {
				return NativePair(func() d.Paired {
					return pair
				})
			}
		case d.Map.Match(tnat):
			if set, ok := nat.(d.Mapped); ok {
				return NativeSet(func() d.Mapped {
					return set
				})
			}
		default:
			return Native(func() d.Native { return nat })
		}
	}
	return Native(func() d.Native { return d.NewNil() })
}

///////////////////////////////////////////////////////////////////////////////
//// NATIVE TYPES
///
// call method ignores arguments, natives are immutable
func (n Native) Call(...Callable) Callable      { return n }
func (n Native) Eval(args ...d.Native) d.Native { return n().Eval(args...) }
func (n Native) TypeNat() d.TyNat               { return n().TypeNat() }
func (n Native) TypeFnc() TyFnc                 { return Data }
func (n Native) String() string                 { return n().String() }
func (n Native) TypeName() string               { return n().TypeName() }

func (n NativeCol) Call(...Callable) Callable      { return n }
func (n NativeCol) Eval(args ...d.Native) d.Native { return n().Eval(args...) }
func (n NativeCol) Len() int                       { return n().Len() }
func (n NativeCol) SliceNat() []d.Native           { return n().Slice() }
func (n NativeCol) Get(key d.Native) d.Native      { return n().Get(key) }
func (n NativeCol) GetInt(idx int) d.Native        { return n().GetInt(idx) }
func (n NativeCol) Range(s, e int) d.Native        { return n().Range(s, e) }
func (n NativeCol) Copy() d.Native                 { return n().Copy() }
func (n NativeCol) TypeNat() d.TyNat               { return n().TypeNat() }
func (n NativeCol) TypeName() string               { return n().TypeName() }
func (n NativeCol) String() string                 { return n().String() }
func (n NativeCol) TypeFnc() TyFnc                 { return Data }
func (n NativeCol) Vector() VecCol                 { return NewVector(n.Slice()...) }
func (n NativeCol) Slice() []Callable {
	var slice = []Callable{}
	for _, val := range n.SliceNat() {
		slice = append(slice, NewNative(val))
	}
	return slice
}

func (n NativePair) Call(...Callable) Callable      { return n }
func (n NativePair) Eval(args ...d.Native) d.Native { return n().Eval(args...) }
func (n NativePair) LeftNat() d.Native              { return n().Left() }
func (n NativePair) RightNat() d.Native             { return n().Right() }
func (n NativePair) BothNat() (l, r d.Native)       { return n().Both() }
func (n NativePair) Left() Callable                 { return NewNative(n().Left()) }
func (n NativePair) Right() Callable                { return NewNative(n().Right()) }
func (n NativePair) LeftTypeNat() d.TyNat           { return n().LeftType() }
func (n NativePair) RightTypeNat() d.TyNat          { return n().RightType() }
func (n NativePair) TypeNat() d.TyNat               { return n().TypeNat() }
func (n NativePair) TypeName() string               { return n().TypeName() }
func (n NativePair) String() string                 { return n().String() }
func (n NativePair) TypeFnc() TyFnc                 { return Data }
func (n NativePair) Pair() Paired {
	return NewPair(
		NewNative(n().Left()),
		NewNative(n().Right()))
}
func (n NativePair) Both() (l, r Callable) {
	return NewNative(n().Left()),
		NewNative(n().Right())
}

func (n NativeSet) Call(args ...Callable) Callable {
	for _, arg := range args {
		n().Eval(arg.Eval())
	}
	return n
}
func (n NativeSet) TypeFnc() TyFnc                       { return Data }
func (n NativeSet) Eval(args ...d.Native) d.Native       { return n().Eval(args...) }
func (n NativeSet) GetNat(acc d.Native) (d.Native, bool) { return n().Get(acc) }
func (n NativeSet) SetNat(acc, val d.Native) d.Mapped    { return n().Set(acc, val) }
func (n NativeSet) Delete(acc d.Native) bool             { return n().Delete(acc) }
func (n NativeSet) KeysNat() []d.Native                  { return n().Keys() }
func (n NativeSet) DataNat() []d.Native                  { return n().Data() }
func (n NativeSet) Fields() []d.Paired                   { return n().Fields() }
func (n NativeSet) KeyTypeNat() d.TyNat                  { return n().KeyType() }
func (n NativeSet) ValTypeNat() d.TyNat                  { return n().ValType() }
func (n NativeSet) TypeNat() d.TyNat                     { return n().TypeNat() }
func (n NativeSet) TypeName() string                     { return n().TypeName() }
func (n NativeSet) String() string                       { return n().String() }
func (n NativeSet) Set() SetCol                          { return NewSet(n.Pairs()...) }
func (n NativeSet) Pairs() []Paired {
	var pairs = []Paired{}
	for _, nat := range n.Fields() {
		pairs = append(
			pairs, NewPair(
				NewNative(nat.Left()),
				NewNative(nat.Right())))
	}
	return pairs
}

//// STATIC FUNCTION EXPRESSIONS OF PREDETERMINED ARITY
///
// static arity functions ignore abundant arguments and return none/nil when
// arguments are passed to call, or eval method
//
/// CONSTANT EXPRESSION
//
// does not expect any arguments and ignores all passed arguments
func NewConstant(constant func() Callable) ConstEq { return constant }

func (c ConstEq) Ident() Callable                { return c() }
func (c ConstEq) Arity() Arity                   { return Arity(0) }
func (c ConstEq) TypeFnc() TyFnc                 { return c().TypeFnc() }
func (c ConstEq) TypeNat() d.TyNat               { return c().TypeNat() }
func (c ConstEq) String() string                 { return c().String() }
func (c ConstEq) TypeName() string               { return c().TypeName() }
func (c ConstEq) Eval(args ...d.Native) d.Native { return c().Eval(args...) }
func (c ConstEq) Call(args ...Callable) Callable { return c() }

/// UNARY EXPRESSION
//
func NewUnary(unary func(arg Callable) Callable) UnaryEq { return unary }

func (u UnaryEq) Ident() Callable                { return u }
func (u UnaryEq) Arity() Arity                   { return Arity(1) }
func (u UnaryEq) TypeFnc() TyFnc                 { return Function }
func (u UnaryEq) TypeNat() d.TyNat               { return d.Function }
func (u UnaryEq) String() string                 { return "T → T" }
func (u UnaryEq) TypeName() string               { return u.String() }
func (u UnaryEq) Eval(args ...d.Native) d.Native { return d.NewNil() }
func (u UnaryEq) Call(args ...Callable) Callable {
	if len(args) > 0 {
		return u(args[0])
	}
	return u
}

/// BINARY EXPRESSION
//
func NewBinary(binary func(l, r Callable) Callable) BinaryEq { return binary }

func (b BinaryEq) Ident() Callable                { return b }
func (b BinaryEq) Arity() Arity                   { return Arity(2) }
func (b BinaryEq) TypeFnc() TyFnc                 { return Function }
func (b BinaryEq) TypeNat() d.TyNat               { return d.Function }
func (b BinaryEq) String() string                 { return "T → T → T" }
func (b BinaryEq) TypeName() string               { return b.String() }
func (b BinaryEq) Eval(args ...d.Native) d.Native { return d.NewNil() }
func (b BinaryEq) Call(args ...Callable) Callable {
	if len(args) > 0 {
		if len(args) > 1 {
			return b(args[0], args[1])
		}
		// return partialy applyed unary
		return NewUnary(func(arg Callable) Callable {
			return b(arg, args[0])
		})
	}
	return b
}

/// VARIADIC EXPRESSION
//
func NewVariadic(expr Callable) VariadicEq {
	return func(args ...Callable) Callable {
		return expr.Call(args...)
	}
}
func (n VariadicEq) Ident() Callable                { return n }
func (n VariadicEq) Call(d ...Callable) Callable    { return n(d...) }
func (n VariadicEq) Eval(args ...d.Native) d.Native { return n().Eval(args...) }
func (n VariadicEq) TypeFnc() TyFnc                 { return Function }
func (n VariadicEq) TypeNat() d.TyNat               { return d.Function }
func (n VariadicEq) String() string                 { return n().String() }
func (n VariadicEq) TypeName() string               { return n().TypeName() }

/// NARY EXPRESSION
//
// nary expression knows it's arity and returns an expression by applying
// passed arguments to the enclosed expression. returns partialy applied
// expression on undersatisfied calls, result of computation on exact
// application of arguments & result followed by abundant arguments on
// oversatisfied calls.
func NewNary(
	expr func(...Callable) Callable,
	arity int,
) NaryEq {
	return func(args ...Callable) Callable {
		if len(args) > 0 {
			// argument number satify expression arity exactly
			if len(args) == arity {
				// expression expects one or more arguments
				if arity > 0 {
					// return fully applyed expression with
					// remaining arity set to be zero
					return expr(args...)
				}
			}

			// argument number undersatisfies expression arity
			if len(args) < arity {
				// return a parially applyed expression with
				// reduced arity wrapping a variadic expression
				// that can take succeeding arguments to
				// concatenate to arguments passed in  prior
				// calls.
				return NewNary(VariadicEq(
					func(succs ...Callable) Callable {
						// return result of calling the
						// nary, passing arguments
						// concatenated to those passed
						// in preceeding calls
						return NewNary(expr, arity).Call(
							append(args, succs...)...)
					}), arity-len(args))
			}

			// argument number oversatisfies expressions arity
			if len(args) > arity {
				// allocate slice of results
				var results = []Callable{}

				// iterate aver arguments & create fully satisfied
				// expressions, wile argument number is higher than
				// expression arity
				for len(args) > arity {
					// apped result of fully satisfiedying
					// expression to results slice
					results = append(
						results,
						NewNary(expr, arity)(
							args[:arity]...),
					)
					// reassign remaining arguments
					args = args[arity:]
				}

				// if any arguments remain, append result of
				// partial application
				if len(args) <= arity && arity > 0 {
					results = append(
						results,
						NewNary(expr, arity)(
							args...))
				}
				// return results slice
				return NewVector(results...)
			}
		}
		// no arguments passed by call, return arity and nested type
		// instead
		// return NewPair(NewNary(expr, arity-len(args)), Arity(arity))
		var remain = arity - len(args)
		return NewVector(NewNary(expr, remain), Arity(remain), Arity(arity))
	}
}

// returns the value returned when calling itself directly, passing along any
// given argument.
func (n NaryEq) Call(args ...Callable) Callable {
	if len(args) > 0 {
		return n(args...)
	}
	return n().(VecCol)()[0]
}
func (n NaryEq) Ident() Callable                { return n }
func (n NaryEq) Eval(args ...d.Native) d.Native { return n.Partial().Eval(args...) }
func (n NaryEq) Expression() Callable           { return n().(VecCol)()[0] }
func (n NaryEq) Remain() Arity                  { return n().(VecCol)()[1].(Arity) }
func (n NaryEq) Arity() Arity                   { return n().(VecCol)()[2].(Arity) }
func (n NaryEq) Partial() NaryEq {
	return NewNary(n.Expression().Call, int(n.Remain()))
}
func (n NaryEq) TypeFnc() TyFnc   { return Function }
func (n NaryEq) TypeNat() d.TyNat { return d.Function }
func (n NaryEq) String() string   { return n.Partial().TypeName() }
func (n NaryEq) TypeName() string {
	var num = int(n.Remain())
	var str string
	for i := 0; i < num; i++ {
		str = str + "T"
		if i < num-1 {
			str = str + " → "
		}
	}
	return str + " → T"
}
