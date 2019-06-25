/*
  FUNCTIONAL CONTAINERS

  containers implement enumeration of functional types, aka lists, vectors
  sets, pairs, tuples‥.
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
	Native     func(...d.Native) d.Native
	NativePair func(...d.Native) d.Paired
	NativeSet  func(...d.Native) d.Mapped
	NativeCol  func(...d.Native) d.Sliceable

	//// STATIC EXPRESSIONS
	ConstLambda  func() Callable
	UnaryLambda  func(Callable) Callable
	BinaryLambda func(l, r Callable) Callable
	VariadLambda func(...Callable) Callable
	NaryLambda   func(...Callable) Callable
)

//// NONE VALUE CONSTRUCTOR
///
// none represens the abscence of a value of any type. implements consumeable,
// key-, index & generic pair interface to be returneable as such.
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

//// TRUTH VALUE CONSTRUCTOR
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

//// TRINARY TRUTH VALUE CONSTRUCTOR
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

//// NATIVE EXPRESSION CONSTRUCTOR
///
// returns an expression with native return type implementing the callable
// interface
func New(inf ...interface{}) Callable {
	return NewNative(d.New(inf...))
}

func NewNative(args ...d.Native) Callable {
	// if any initial arguments have been passed

	var nat = d.NewFromData(args...)
	var tnat = nat.TypeNat()

	switch {
	case tnat.Match(d.Slice):
		if slice, ok := nat.(d.Sliceable); ok {
			return NativeCol(func(nats ...d.Native) d.Sliceable {
				if len(nats) > 0 {
					return slice.Eval(nats...).(d.Sliceable)
				}
				return slice
			})
		}
	case tnat.Match(d.Unboxed):
		if slice, ok := nat.(d.Sliceable); ok {
			return NativeCol(func(nats ...d.Native) d.Sliceable {
				if len(nats) > 0 {
					return slice.Eval(nats...).(d.Sliceable)
				}
				return slice
			})
		}
	case tnat.Match(d.Pair):
		if pair, ok := nat.(d.Paired); ok {
			return NativePair(func(nats ...d.Native) d.Paired {
				if len(nats) > 0 {
					return pair.Eval(nats...).(d.Paired)
				}
				return pair
			})
		}
	case tnat.Match(d.Map):
		if set, ok := nat.(d.Mapped); ok {
			return NativeSet(func(nats ...d.Native) d.Mapped {
				if len(nats) > 0 {
					return set.Eval(nats...).(d.Mapped)
				}
				return set
			})
		}
	default:
		return Native(func(nats ...d.Native) d.Native {
			if len(nats) > 0 {
				return nat.Eval(nats...)
			}
			return nat
		})
	}

	return Native(func(...d.Native) d.Native { return d.NewNil() })
}

//// NATIVE EXPRESSIONS
///
// expression with flat native return type
func (n Native) Call(...Callable) Callable      { return n }
func (n Native) Eval(args ...d.Native) d.Native { return n(args...) }
func (n Native) String() string                 { return n().String() }
func (n Native) TypeName() string               { return n().TypeName() }
func (n Native) TypeNat() d.TyNat               { return n().TypeNat() }
func (n Native) TypeFnc() TyFnc                 { return Data }

// expression which returns a native slice and implements consumeable
func (n NativeCol) Call(...Callable) Callable      { return n }
func (n NativeCol) TypeFnc() TyFnc                 { return Data }
func (n NativeCol) Eval(args ...d.Native) d.Native { return n(args...) }
func (n NativeCol) Len() int                       { return n().Len() }
func (n NativeCol) SliceNat() []d.Native           { return n().Slice() }
func (n NativeCol) Get(key d.Native) d.Native      { return n().Get(key) }
func (n NativeCol) GetInt(idx int) d.Native        { return n().GetInt(idx) }
func (n NativeCol) Range(s, e int) d.Native        { return n().Range(s, e) }
func (n NativeCol) Copy() d.Native                 { return n().Copy() }
func (n NativeCol) TypeNat() d.TyNat               { return n().TypeNat() }
func (n NativeCol) TypeName() string               { return n().TypeName() }
func (n NativeCol) String() string                 { return n().String() }
func (n NativeCol) Vector() VecCol                 { return NewVector(n.Slice()...) }
func (n NativeCol) Slice() []Callable {
	var slice = []Callable{}
	for _, val := range n.SliceNat() {
		slice = append(slice, NewNative(val))
	}
	return slice
}

// expression which returns a native pair and implements paired
func (n NativePair) Call(...Callable) Callable      { return n }
func (n NativePair) TypeFnc() TyFnc                 { return Data }
func (n NativePair) Eval(args ...d.Native) d.Native { return n(args...) }
func (n NativePair) LeftNat() d.Native              { return n().Left() }
func (n NativePair) RightNat() d.Native             { return n().Right() }
func (n NativePair) BothNat() (l, r d.Native)       { return n().Both() }
func (n NativePair) Left() Callable                 { return NewNative(n().Left()) }
func (n NativePair) Right() Callable                { return NewNative(n().Right()) }
func (n NativePair) KeyType() d.TyNat               { return n().LeftType() }
func (n NativePair) ValType() d.TyNat               { return n().RightType() }
func (n NativePair) TypeNat() d.TyNat               { return n().TypeNat() }
func (n NativePair) TypeName() string               { return n().TypeName() }
func (n NativePair) String() string                 { return n().String() }
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

// expression which returns a native set and implements mapped
func (n NativeSet) Ident() Callable                      { return n }
func (n NativeSet) TypeFnc() TyFnc                       { return Data }
func (n NativeSet) Eval(args ...d.Native) d.Native       { return n(args...) }
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

//// LAMBDA CONSTRUCTORS
///
//
/// CONSTANT EXPRESSION
//
// constant expression constructor creates an expression that does not expect
// any arguments
func NewConstant(constant func() Callable) ConstLambda { return constant }

func (c ConstLambda) Ident() Callable                { return c() }
func (c ConstLambda) Arity() Arity                   { return Arity(0) }
func (c ConstLambda) TypeFnc() TyFnc                 { return c().TypeFnc() }
func (c ConstLambda) TypeNat() d.TyNat               { return c().TypeNat() }
func (c ConstLambda) String() string                 { return c().String() }
func (c ConstLambda) TypeName() string               { return c().TypeName() }
func (c ConstLambda) Eval(args ...d.Native) d.Native { return c().Eval(args...) }
func (c ConstLambda) Call(args ...Callable) Callable { return c() }

//// UNARY EXPRESSION
///
// unary expression constructor
func NewUnary(unary func(arg Callable) Callable) UnaryLambda { return unary }

func (u UnaryLambda) Ident() Callable                { return u }
func (u UnaryLambda) Arity() Arity                   { return Arity(1) }
func (u UnaryLambda) TypeFnc() TyFnc                 { return Function }
func (u UnaryLambda) TypeNat() d.TyNat               { return d.Function }
func (u UnaryLambda) String() string                 { return "T → T" }
func (u UnaryLambda) TypeName() string               { return u.String() }
func (u UnaryLambda) Eval(args ...d.Native) d.Native { return d.NewNil() }
func (u UnaryLambda) Call(args ...Callable) Callable {
	if len(args) > 0 {
		return u(args[0])
	}
	return u
}

//// BINARY EXPRESSION
///
// binary expression constructor
func NewBinary(binary func(l, r Callable) Callable) BinaryLambda { return binary }

func (b BinaryLambda) Ident() Callable                { return b }
func (b BinaryLambda) Arity() Arity                   { return Arity(2) }
func (b BinaryLambda) TypeFnc() TyFnc                 { return Function }
func (b BinaryLambda) TypeNat() d.TyNat               { return d.Function }
func (b BinaryLambda) String() string                 { return "T → T → T" }
func (b BinaryLambda) TypeName() string               { return b.String() }
func (b BinaryLambda) Eval(args ...d.Native) d.Native { return d.NewNil() }
func (b BinaryLambda) Call(args ...Callable) Callable {
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

//// VARIADIC EXPRESSION
///
// variadic expression constructor creates expression to evaluate arbitrary
// number of arguments
func NewVariadic(expr Callable) VariadLambda {
	return func(args ...Callable) Callable {
		return expr.Call(args...)
	}
}
func (n VariadLambda) Ident() Callable                { return n }
func (n VariadLambda) Call(d ...Callable) Callable    { return n(d...) }
func (n VariadLambda) Eval(args ...d.Native) d.Native { return n().Eval(args...) }
func (n VariadLambda) TypeFnc() TyFnc                 { return Function }
func (n VariadLambda) TypeNat() d.TyNat               { return d.Function }
func (n VariadLambda) String() string                 { return n().String() }
func (n VariadLambda) TypeName() string               { return n().TypeName() }

//// NARY EXPRESSION
///
// nary expression constructor creates an expression that knows it's arity and
// returns a result by applying passed arguments to the enclosed expression.
// returns partialy applied expression on undersatisfied call, result of
// computation on exact application of arguments & a slice of results from
// applying abundant recursively on oversatisfied calls. last result may be
// partialy applied
func NewNary(
	expr func(...Callable) Callable,
	arity int,
) NaryLambda {
	return func(args ...Callable) Callable {

		if len(args) > 0 {

			// argument number satifies expression arity exactly
			if len(args) == arity {
				// expression expects one or more arguments
				if arity > 0 {
					// return fully applied expression with
					// remaining arity set to be zero
					return expr(args...)
				}
				// expression is a constant (don't do that,
				// it's what buildtin const expressions are
				// for)
				return expr()
			}

			// argument number undersatisfies expression arity
			if len(args) < arity {
				// return a parially applyed expression with
				// reduced arity wrapping a variadic expression
				// that can take succeeding arguments to
				// concatenate to arguments passed in  prior
				// calls.
				return NewNary(VariadLambda(
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

				// iterate aver arguments & create fully
				// satisfied expressions, while argument number
				// is higher than expression arity
				for len(args) > arity {
					// apped result of fully satisfied
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
				// partial application to slice of results
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

		// no arguments passed, return partial, remaining arity &
		// initial arity instead
		var remain = arity - len(args)
		return NewVector(
			NewNary(expr, remain),
			Arity(remain),
			Arity(arity),
		)
	}
}

// returns the value returned when calling itself directly, passing along any
// given argument.
func (n NaryLambda) Ident() Callable                { return n }
func (n NaryLambda) Expression() Callable           { return n().(VecCol)()[0] }
func (n NaryLambda) Remain() Arity                  { return n().(VecCol)()[1].(Arity) }
func (n NaryLambda) Arity() Arity                   { return n().(VecCol)()[2].(Arity) }
func (n NaryLambda) TypeFnc() TyFnc                 { return n.Expression().TypeFnc() }
func (n NaryLambda) TypeNat() d.TyNat               { return n.Expression().TypeNat() }
func (n NaryLambda) String() string                 { return n.Partial().TypeName() }
func (n NaryLambda) Eval(args ...d.Native) d.Native { return n.Partial().Eval(args...) }
func (n NaryLambda) Partial() NaryLambda {
	return NewNary(n.Expression().Call, int(n.Remain()))
}
func (n NaryLambda) Call(args ...Callable) Callable {
	if len(args) > 0 {
		return n(args...)
	}
	return n().(VecCol)()[0]
}
func (n NaryLambda) TypeName() string {
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
