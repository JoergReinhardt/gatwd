/*
  FUNCTIONAL CONTAINERS

  containers implement enumeration of functional types, aka lists, vectors sets, pairs, tuples‥.
*/
package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

type (
	//// PREDICATE
	PredictArg  func(Callable) bool
	PredictAll  func(...Callable) bool
	PredictAny  func(...Callable) bool
	PredictNarg func(...Callable) bool

	//// CASE-EXPRESSION | CASE-SWITCH
	CaseExpr   func(...Callable) (Callable, bool)
	CaseSwitch func(...Callable) (Callable, bool, []CaseExpr)

	////  NONE | JUST | MAYBE
	NoneVal   func()
	JustVal   func(...Callable) Callable
	MaybeVal  func(...Callable) Callable
	MaybeType func(...Callable) MaybeVal

	//// STATIC EXPRESSIONS
	ConstantExpr func() Callable
	UnaryExpr    func(Callable) Callable
	BinaryExpr   func(a, b Callable) Callable
	NaryExpr     func(...Callable) Callable
	VariadicExpr func(...Callable) Callable

	//// DATA VALUE
	AtomVal func(args ...d.Native) d.Native
)

//// PREDICATE
///
// predict one is an expression that returns either true, or false depending on
// first passed arguement passed. succeeding arguements are ignored
func NewPredictArg(pred func(Callable) bool) PredictArg {
	return func(arg Callable) bool { return pred(arg) }
}
func (p PredictArg) String() string   { return "Predicate" }
func (p PredictArg) TypeNat() d.TyNat { return d.Functor }
func (p PredictArg) TypeFnc() TyFnc   { return Predicate }
func (p PredictArg) Eval(args ...d.Native) d.Native {
	if len(args) > 0 {
		return d.BoolVal(p(NewAtom(args[0])))
	}
	return d.BoolVal(false)
}
func (p PredictArg) Call(args ...Callable) Callable {
	if len(args) > 0 {
		return NewAtom(d.BoolVal(p(args[0])))
	}
	return NewAtom(d.BoolVal(false))
}
func (p PredictArg) ToPredictNarg() PredictNarg {
	return func(args ...Callable) bool {
		if len(args) > 0 {
			return p(args[0])
		}
		return false
	}
}

///////////////////////////////////////////////////////////////////////////////
// predict many returns true, or false depending on all arguments that have
// been passed calling it
func NewPredictNarg(pred func(...Callable) bool) PredictNarg {
	return func(args ...Callable) bool {
		return pred(args...)
	}
}
func (p PredictNarg) String() string   { return "Nary Predicate" }
func (p PredictNarg) TypeNat() d.TyNat { return d.Functor }
func (p PredictNarg) TypeFnc() TyFnc   { return Predicate }
func (p PredictNarg) Eval(args ...d.Native) d.Native {
	if len(args) > 0 {
		var exprs = []Callable{}
		for _, arg := range args {
			exprs = append(exprs, NewAtom(arg))
			return d.BoolVal(p(exprs...))
		}
	}
	return d.BoolVal(false)
}
func (p PredictNarg) Call(args ...Callable) Callable {
	if len(args) > 0 {
		return NewAtom(d.BoolVal(p(args...)))
	}
	return NewAtom(d.BoolVal(false))
}

///////////////////////////////////////////////////////////////////////////////
// all-predicate returns true, if all arguments passed yield true, when applyed
// to predicate one after another
func NewPredictAll(pred func(Callable) bool) PredictNarg {
	return func(args ...Callable) bool {
		var result = true
		for _, arg := range args {
			if !pred(arg) {
				return false
			}
		}
		return result
	}
}

// eval converts all it's arguments to a single atomic expression, applys it to
// the predicate and returns the resulting boolean as instance of an atomic
// expression. if no arguemnts are passed, the atomic result yields 'false' as
// its default return value.
func (p PredictAll) Eval(args ...d.Native) d.Native {
	if len(args) > 0 {
		var exprs = []Callable{}
		for _, arg := range args {
			exprs = append(exprs, NewAtom(arg))
			return d.BoolVal(p(exprs...))
		}
	}
	return d.BoolVal(false)
}

// call passes arguments on to the enclosed all-predicate
func (p PredictAll) Call(args ...Callable) Callable {
	if len(args) > 0 {
		return NewAtom(d.BoolVal(p(args...)))
	}
	return NewAtom(d.BoolVal(false))
}

func (p PredictAll) String() string   { return "All Predicate" }
func (p PredictAll) TypeFnc() TyFnc   { return Predicate }
func (p PredictAll) TypeNat() d.TyNat { return d.Functor }

///////////////////////////////////////////////////////////////////////////////
// will return true, if any of the passed arguments yield true, when applyed to
// predicate one after another
func NewPredictAny(pred func(Callable) bool) PredictNarg {
	return func(args ...Callable) bool {
		var result = false
		for _, arg := range args {
			if pred(arg) {
				return true
			}
		}
		return result
	}
}
func (p PredictAny) String() string   { return "Any Predicate" }
func (p PredictAny) TypeNat() d.TyNat { return d.Functor }
func (p PredictAny) TypeFnc() TyFnc   { return Predicate }
func (p PredictAny) Eval(args ...d.Native) d.Native {
	if len(args) > 0 {
		var exprs = []Callable{}
		for _, arg := range args {
			exprs = append(exprs, NewAtom(arg))
			return d.BoolVal(p(exprs...))
		}
	}
	return d.BoolVal(false)
}
func (p PredictAny) Call(args ...Callable) Callable {
	if len(args) > 0 {
		return NewAtom(d.BoolVal(p(args...)))
	}
	return NewAtom(d.BoolVal(false))
}

///////////////////////////////////////////////////////////////////////////////
//// CASE EXPRESSION
///
func NewCase(pred PredictNarg) CaseExpr {
	return func(args ...Callable) (Callable, bool) {
		if len(args) > 0 {
			if len(args) > 1 {
				if pred(args...) {
					return NewVector(args...), true
				}
				return NewVector(args...), false
			}
			if pred(args[0]) {
				return args[0], true
			}
			return NewVector(args...), false
		}
		return NewNone(), false
	}
}
func (c CaseExpr) Ident() Callable  { return c }
func (c CaseExpr) String() string   { return "Case" }
func (c CaseExpr) TypeFnc() TyFnc   { return Case }
func (c CaseExpr) TypeNat() d.TyNat { return d.Functor }
func (c CaseExpr) Call(args ...Callable) Callable {
	if len(args) > 0 {
		if result, ok := c(args...); ok {
			return result
		}
	}
	return NewNone()
}

func (c CaseExpr) Eval(args ...d.Native) d.Native {
	if len(args) > 0 {
		if expr, ok := c(NewAtom(args...)); ok {
			return expr
		}
	}
	return d.NilVal{}
}

//// CASE SWITCH
///
// applys passed arguments to all enclosed cases in the order passed to the
// switch constructor
func NewSwitch(cases ...CaseExpr) CaseSwitch {
	return func(args ...Callable) (Callable, bool, []CaseExpr) {
		if len(cases) > 0 {
			var expr, ok = cases[0](args...)
			if len(cases) > 1 {
				return expr, ok, cases[1:]
			}
			return expr, ok, nil
		}
		return NewNone(), false, nil
	}
}
func (s CaseSwitch) String() string   { return "Switch" }
func (s CaseSwitch) TypeFnc() TyFnc   { return Switch }
func (s CaseSwitch) TypeNat() d.TyNat { return d.Functor }

// switches call method iterates over cases until either boolean true is
// yielded and returns result, or all cases are depleted
func (s CaseSwitch) Call(args ...Callable) Callable {
	if len(args) > 0 {
		var result, ok, cases = s(args...)
		for len(cases) > 0 {
			result, ok = cases[0](args...)
			if ok {
				return result
			}
			if len(cases) > 1 {
				cases = cases[1:]
			} else {
				return NewNone()
			}
		}
	}
	return NewNone()
}

// eval converts its arguments to callable and evaluates the result to yield a
// return value of native type
func (s CaseSwitch) Eval(args ...d.Native) d.Native {
	if len(args) > 0 {
		var exprs = []Callable{}
		for _, arg := range args {
			exprs = append(exprs, NewAtom(arg))
		}
		return s.Call(exprs...).Eval()
	}
	return d.NilVal{}
}

//// NONE VALUE
func NewNone() NoneVal                             { return func() {} }
func (n NoneVal) Ident() Callable                  { return n }
func (n NoneVal) Len() int                         { return 0 }
func (n NoneVal) String() string                   { return "⊥" }
func (n NoneVal) Eval(...d.Native) d.Native        { return nil }
func (n NoneVal) Value() Callable                  { return nil }
func (n NoneVal) Call(...Callable) Callable        { return nil }
func (n NoneVal) Empty() bool                      { return true }
func (n NoneVal) TypeFnc() TyFnc                   { return None }
func (n NoneVal) TypeNat() d.TyNat                 { return d.Nil }
func (n NoneVal) TypeName() string                 { return n.String() }
func (n NoneVal) Head() Callable                   { return NewNone() }
func (n NoneVal) Tail() Consumeable                { return NewNone() }
func (n NoneVal) Consume() (Callable, Consumeable) { return NewNone(), NewNone() }

//// MAYVE | JUST | NONE
///
/// MAYBE TYPE CONSTRUCTOR
//
// maybe type constructor returns a data constructor of the maybe type defined
// by it's predicate expression when called. the data constructor in turn,
// either returns an instance of the just-, or none type depending on the
// enclosed predicate.
func DefineMaybeType(predi PredictNarg) MaybeType {
	return func(exprs ...Callable) MaybeVal {
		if len(exprs) > 0 {
			if len(exprs) > 1 {
				var maybes = []Callable{}
				for _, expr := range exprs {
					maybes = append(maybes, consMaybe(predi, expr.Call))
				}
				return NewVector(maybes...).Call
			}
			return consMaybe(predi, exprs[0].Call)
		}
		return consMaybe(predi, NewNone().Call)
	}
}

//// maybe type constructors callable interface implementation
func (m MaybeType) String() string   { return "Maybe·Type" }
func (m MaybeType) TypeFnc() TyFnc   { return Constructor | Maybe }
func (m MaybeType) TypeNat() d.TyNat { return d.Functor }

// returns a maybe data constructor based on its type defining predicate
func (m MaybeType) Call(args ...Callable) Callable {
	if len(args) > 0 {
		return m(args...)
	}
	return m()
}

// returns a maybe data constructor by converting native arguments to atomic
// instances and calling itself passing those
func (m MaybeType) Eval(args ...d.Native) d.Native {
	if len(args) > 0 {
		var exprs = []Callable{}
		for _, arg := range args {
			exprs = append(exprs, NewAtom(arg))
		}
		return m(exprs...).Eval()
	}
	return m().Eval()
}

//// MAYBE VALUE
///
// maybe values are created by a data constructor which is defined and created
// by the maybe-type constructor & uses its enclosed predicate to choose
// between and create either just or none instances from passed arguments.
func consMaybe(predi PredictNarg, expr VariadicExpr) MaybeVal {
	return func(args ...Callable) Callable {
		var result = expr(args...)
		if predi(result) {
			return consJust(result)
		}
		return NewNone()
	}
}
func (m MaybeVal) TypeName() string {
	return TyFnc(m().TypeFnc().Flag().Mask(Atom|Vector).Flag()).TypeName() + "·" + m().TypeNat().TypeName()
}
func (m MaybeVal) String() string                 { return m().String() }
func (m MaybeVal) TypeFnc() TyFnc                 { return m().TypeFnc() }
func (m MaybeVal) TypeNat() d.TyNat               { return m().TypeNat() }
func (m MaybeVal) Call(args ...Callable) Callable { return m().Call(args...) }
func (m MaybeVal) Eval(args ...d.Native) d.Native { return m().Eval(args...) }

//// JUST VALUE
///
// instances of the just type are constructed by maybe data constructors
// exlusively, when arguments passed to them yield true, when applyed to the
// particular maybe types enclosed predicate
func consJust(expr Callable) JustVal {
	return func(args ...Callable) Callable {
		if len(args) > 0 {
			return expr.Call(args...)
		}
		return expr.Call()
	}
}
func (n JustVal) Ident() Callable                { return n }
func (n JustVal) Value() Callable                { return n() }
func (n JustVal) Call(args ...Callable) Callable { return n().Call(args...) }
func (n JustVal) Eval(args ...d.Native) d.Native { return n().Eval(args...) }
func (n JustVal) String() string                 { return n().String() }
func (n JustVal) TypeNat() d.TyNat               { return n().TypeNat() }
func (n JustVal) TypeFnc() TyFnc                 { return Just | n().TypeFnc() }
func (n JustVal) TypeName() string               { return "Just·" + n().TypeFnc().String() }

//// STATIC FUNCTION EXPRESSIONS OF PREDETERMINED ARITY
///
// static arity functions ignore abundant arguments and return none, if the
// number of arguments on calling Call, or Eval does not match the functions
// arity

/// CONSTANT EXPRESSION
//
// returns a constant expression ignoring all arguments on eval & call
func NewConstant(
	expr Callable,
) ConstantExpr {
	return func() Callable { return expr }
}
func (c ConstantExpr) Ident() Callable           { return c() }
func (c ConstantExpr) Arity() Arity              { return Arity(0) }
func (c ConstantExpr) TypeFnc() TyFnc            { return c().TypeFnc() }
func (c ConstantExpr) TypeNat() d.TyNat          { return c().TypeNat() }
func (c ConstantExpr) Eval(...d.Native) d.Native { return c().Eval() }
func (c ConstantExpr) Call(...Callable) Callable { return c() }

/// UNARY EXPRESSION
//
// expects one argument, ignores further arguments on eval and call and returns
// nil/none, when no arguments are passed.
func NewUnary(
	expr Callable,
) UnaryExpr {
	return func(arg Callable) Callable { return expr.Call(arg) }
}
func (u UnaryExpr) Ident() Callable  { return u }
func (u UnaryExpr) Arity() Arity     { return Arity(1) }
func (u UnaryExpr) TypeFnc() TyFnc   { return u(NewNone()).TypeFnc() }
func (u UnaryExpr) TypeNat() d.TyNat { return d.Functor.TypeNat() }
func (u UnaryExpr) Call(args ...Callable) Callable {
	if len(args) > 1 {
		return u(args[0]).Call()
	}
	return NewNone()
}
func (u UnaryExpr) Eval(arg ...d.Native) d.Native {
	if len(arg) > 0 {
		return u(NewAtom(arg[0]))
	}
	return d.NilVal{}
}

/// BINARY EXPRESSION
//
// expects two arguments, ignores further arguments on eval and call and
// returns nil/none, when less than two arguments are passed.
func NewBinary(
	expr Callable,
) BinaryExpr {
	return func(a, b Callable) Callable {
		return expr.Call(a, b)
	}
}

func (b BinaryExpr) Ident() Callable  { return b }
func (b BinaryExpr) Arity() Arity     { return Arity(2) }
func (b BinaryExpr) TypeFnc() TyFnc   { return b(NewNone(), NewNone()).TypeFnc() }
func (b BinaryExpr) TypeNat() d.TyNat { return d.Functor.TypeNat() }
func (b BinaryExpr) Call(args ...Callable) Callable {
	if len(args) > 1 {
		return b(args[0], args[1])
	}
	return NewNone()
}
func (b BinaryExpr) Eval(args ...d.Native) d.Native {
	if len(args) > 1 {
		return b(NewAtom(args[0]), NewAtom(args[1]))
	}
	return d.NilVal{}
}

/// NARY EXPRESSION
//
// nary expression knows it's arity and returns an expression by applying
// arguments to the enclosed expression, handling partial-, exact-, and
// oversatisfied calls, by returning either
//
// - a partialy applied function and an altered arity reduced by the number of
//   arguments passed allready,
//
// - the result of applying the exact number of arguments to the expression and
//   a zero arity,
//
// - or a pair instance returning the result of applying the exact number of
//   arguments matching the arity as it's left field and a continuation
//   returning the result of creating a new nary instance from the initial
//   expression & arity and calling it with the remaining arguments as it's
//   right field and whatever arity was returned when creating that instance.
func NewNary(
	expr Callable,
	arity int,
	typ TyComp,
) NaryExpr {
	return func(args ...Callable) Callable {

		if len(args) > 0 {

			// argument number satify expression arity exactly
			if len(args) == arity {
				// expression expects one or more arguments
				if arity > 0 {
					// return fully applyed expression with
					// remaining arity set to be zero
					return expr.Call(args...)
				}
			}

			// argument number undersatisfies expression arity
			if len(args) < arity {
				// return a parially applyed expression with reduced
				// arity wrapping a variadic expression that can take
				// succeeding arguments to concatenate to arguments
				// passed in  prior calls.
				return NewNary(VariadicExpr(
					func(succs ...Callable) Callable {
						// return result of calling the
						// nary, passing arguments
						// concatenated to those passed
						// in preceeding calls
						return NewNary(expr, arity, typ).Call(
							append(args, succs...)...)
					}), arity-len(args), typ)
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
						NewNary(expr, arity, typ)(
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
						NewNary(expr, arity, typ)(
							args...))
				}
				// return results slice
				return NewVector(results...)
			}
		}
		return NewPair(Arity(arity), typ)
	}
}
func (n NaryExpr) Ident() Callable  { return n }
func (n NaryExpr) Arity() Arity     { val := n(); return val.(Paired).Left().(Arity) }
func (n NaryExpr) CompType() TyComp { val := n(); return val.(Paired).Right().(TyComp) }
func (n NaryExpr) TypeFnc() TyFnc   { return n.CompType().TypeFnc() }
func (n NaryExpr) TypeNat() d.TyNat { return n.CompType().TypeNat() }
func (n NaryExpr) TypeName() string { return n.CompType().TypeName() }

// returns the value returned when calling itself directly, passing along any
// given argument.
func (n NaryExpr) Call(args ...Callable) Callable {
	if len(args) > 0 {
		return n(args...)
	}
	return n()
}

// eval method converts it's arguments to implement callable to pass on to the
// call method and returns the result as native by evaluating it
func (n NaryExpr) Eval(args ...d.Native) d.Native {
	var exprs = []Callable{}
	if len(args) > 0 {
		for _, arg := range args {
			exprs = append(exprs, NewAtom(arg))
		}
	}
	return n(exprs...).Eval()
}

/// VARIADIC EXPRESSION
//
// variadic expression has an unknown arity and can take a varying number of
// arguments passed calling it
func NewVariadic(
	expr Callable,
) VariadicExpr {
	return func(args ...Callable) Callable {
		return expr.Call(args...)
	}
}
func (n VariadicExpr) Ident() Callable             { return n }
func (n VariadicExpr) TypeFnc() TyFnc              { return n().TypeFnc() }
func (n VariadicExpr) TypeNat() d.TyNat            { return d.Functor.TypeNat() }
func (n VariadicExpr) Call(d ...Callable) Callable { return n(d...) }
func (n VariadicExpr) Eval(args ...d.Native) d.Native {
	var params = []Callable{}
	if len(args) > 0 {
		for _, arg := range args {
			params = append(params, NewAtom(arg))
		}
		return n(params...)
	}
	return n()
}

//// DATA VALUE
///
// data value implements the callable interface but returns an instance of
// data/Value. the eval method of every native can be passed as argument
// instead of the value itself, as in 'DataVal(native.Eval)', to delay, or even
// possibly ommit evaluation of the underlying data value for cases where
// lazynes is paramount.
func New(inf ...interface{}) Callable { return NewAtom(d.New(inf...)) }

// create an atomic instance returning a single instance of native, that may
// turn out to be an unboxed vector of native type, in case all passed
// arguments yield the same type, a slice of native instances in case arguments
// are of mixed type, or the result of evaluating the first argument, either
// passing on succeeding arguments, or not when a single initial argument has
// been passed.
func NewAtom(args ...d.Native) AtomVal {
	// if any initial arguments have been passed
	if len(args) > 0 {
		// if more than a single initial argument has been passed
		if len(args) > 1 {
			return func(succs ...d.Native) d.Native {
				// if succeeding arguments have been passed
				if len(succs) > 0 {
					// try to convert to slice of unboxed
					// natives. falls back automaticly, if
					// arguments are of multiple type
					return d.SliceToNatives(
						// append succeeding arguments
						// to initial arguments
						d.NewSlice(append(args, succs...)...),
					)
				}
				// no succeeding arguments try to convert
				// initial arguments to slice of unboxed
				// natives if possible. falls back automaticly
				// to return slice of natives with multiple
				// types, if that's the case
				return d.SliceToNatives(d.NewSlice(args...))
			}
		}

		// special case, where only a single argument has been passed
		// initialy → return atomic expression to return the argument,
		// appending succeeding arguments in case such have been passed
		// in succeeding call. yields a single native value when no
		// succeeding arguments are passed, an unboxed vector, if all
		// succeeding arguments happen to be of the same type as the
		// initial one, or a slice of multiple typed native instances
		// it thats not the case
		return func(succs ...d.Native) d.Native {
			// if succeeding arguments are passed‥.
			if len(succs) > 0 {
				// append succeding arguments to initial
				// argument and call new atom recursively
				return NewAtom(append(args, succs...)...)
			}
			// return first argument unchanged
			return args[0]
		}
	}

	// no initial arguments have been passed. return atomic instance
	// returning a nil instance, when no succeding arguments are passed or
	// the result of creating an atomic instance from succeeding arguments
	// and evaluating it with an empty set of arguments
	return func(succs ...d.Native) d.Native {
		// if arguments have been passed at succsseeding call
		if len(succs) > 0 {
			// return native instance enclosed by atomic val
			// created by passing succsseeding arguments on to call
			// new-atom
			return NewAtom(succs...)()
		}
		// return instance of nil-value if neither initial, nor
		// succsseeding arguments have been passed
		return d.NilVal{}
	}
}

// evaluate passes arguemnts on to call new-atom and returns the native
// instance enclosed by resulting atomic expression
func (n AtomVal) Eval(args ...d.Native) d.Native {
	if len(args) > 0 {
		return NewAtom(args...)()
	}
	// returns the enclosed expression
	return n()
}

// call evaluates passed arguments to create a slice of native instances as
// arguements to pass on when calling new-atom to yield a Callable instance of
// atomic type. if no arguments are passed, expressions returns it's identity
// instead
func (n AtomVal) Call(args ...Callable) Callable {
	if len(args) > 0 {
		var nats = []d.Native{}
		for _, arg := range args {
			nats = append(nats, arg.Eval())
		}
		return NewAtom(nats...)
	}
	return n
}

// return atom type to indicate that this expression is limited to return an
// instance of native type
func (n AtomVal) TypeFnc() TyFnc { return Atom }

// return the enclosed expressions native type
func (n AtomVal) TypeNat() d.TyNat { return n().TypeNat() }

// return the string returned by stringger of enclosed type
func (n AtomVal) String() string   { return n().String() }
func (n AtomVal) TypeName() string { return n().TypeNat().TypeName() }

//// HELPER FUNCTIONS TO HANDLE ARGUMENTS
///
// since every callable also needs to implement the eval interface and data as
// such allways boils down to native values, conversion between callable-/ &
// native arguments is frequently needed. arguments may also need to be
// reversed when intendet to be passed to certain recursive expressions, or
// returned by those
//
/// REVERSE ARGUMENTS
func revArgs(args ...Callable) []Callable {
	var rev = []Callable{}
	for i := len(args) - 1; i > 0; i-- {
		rev = append(rev, args[i])
	}
	return rev
}

/// CONVERT NATIVE TO FUNCTIONAL
func natToFnc(args ...d.Native) []Callable {
	var result = []Callable{}
	for _, arg := range args {
		result = append(result, NewAtom(arg))
	}
	return result
}

/// CONVERT FUNCTIONAL TO NATIVE
func fncToNat(args ...Callable) []d.Native {
	var result = []d.Native{}
	for _, arg := range args {
		result = append(result, arg.Eval())
	}
	return result
}

/// GROUP ARGUMENTS PAIRWISE
//
// assumes the arguments to either implement paired, or be alternating pairs of
// key & value. in case the number of passed arguments that are not pairs is
// uneven, last field will be filled up with a value of type none
func argsToPaired(args ...Callable) []Paired {
	var pairs = []Paired{}
	var alen = len(args)
	for i, arg := range args {
		if arg.TypeFnc().Match(Pair) {
			pairs = append(pairs, arg.(Paired))
		}
		if i < alen-2 {
			i = i + 1
			pairs = append(pairs, NewPair(arg, args[i]))
		}
		pairs = append(pairs, NewPair(arg, NewNone()))
	}
	return pairs
}
