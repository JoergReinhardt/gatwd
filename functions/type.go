package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

type Arity d.Uint8Val

//go:generate stringer -type Arity
const (
	Nullary Arity = 0 + iota
	Unary
	Binary
	Ternary
	Quaternary
	Quinary
	Senary
	Septenary
	Octonary
	Novenary
	Denary
)

func (a Arity) Eval(v ...d.Native) d.Native { return a }

func (a Arity) Int() int { return int(a) }

func (a Arity) Flag() d.BitFlag { return d.BitFlag(a) }

func (a Arity) TypeNat() d.TyNative { return d.Flag }

func (a Arity) TypeFnc() TyFnc { return HigherOrder }

func (a Arity) Match(arg Arity) bool { return a == arg }

// properys relevant for application
type Propertys d.Uint8Val

//go:generate stringer -type Propertys
const (
	Default Propertys = 0
	PostFix Propertys = 1
	InFix   Propertys = 1 + iota
	// ⌐: PreFix
	Atomic
	// ⌐: Thunk
	Eager
	// ⌐: Lazy
	Right
	// ⌐: Left_Bound
	Mutable
	// ⌐: Imutable
	SideEffect
	// ⌐: Pure
	Primitive
	// ⌐: Parametric
)

func (p Propertys) TypePrime() d.TyNative { return d.Flag }

func (p Propertys) TypeFnc() TyFnc { return HigherOrder }

func (p Propertys) Flag() d.BitFlag { return p.TypeFnc().Flag() }

func (p Propertys) Eval(a ...d.Native) d.Native { return p.Flag() }

func (p Propertys) Match(arg Propertys) bool {
	if p&arg != 0 {
		return true
	}
	return false
}

// type TyFnc d.BitFlag
// encodes the kind of functional data as bitflag
type TyFnc d.UintVal

func (t TyFnc) Eval(...d.Native) d.Native { return t }

func (t TyFnc) TypeHO() TyFnc { return t }

func (t TyFnc) TypeNat() d.TyNative { return d.Flag }

func (t TyFnc) Flag() d.BitFlag { return d.BitFlag(t) }

func (t TyFnc) Uint() uint { return d.BitFlag(t).Uint() }

//go:generate stringer -type=TyFnc
const (
	Type TyFnc = 1 << iota
	Data
	///////////
	Definition
	Expression
	///////////
	Variable
	Function
	Closure
	///////////
	Argument
	Parameter
	Accessor
	Attribut
	Predicate
	Resource
	Aggregator
	Generator
	Constructor
	Functor
	Application
	Monad
	///////////
	Condition
	Equality
	False
	True
	Just
	None
	Case
	Switch
	Either
	Or
	If
	Else
	Error
	///////////
	Pair
	List
	ListR
	Tuple
	UniSet
	MuliSet
	AssocVec
	Record
	Vector
	DLink
	Link
	Node
	Tree
	IO
	///////////
	HigherOrder

	Truth      = True | False
	Option     = Just | None
	Chain      = Vector | Tuple | Record
	AccIndex   = Vector | Chain
	AccSymbol  = Tuple | AssocVec | Record
	AccCollect = AccIndex | AccSymbol
	Nests      = Tuple | List | ListR
	Sets       = UniSet | MuliSet | AssocVec | Record
	Links      = Link | DLink | Node | Tree
)

//// (RE-) INSTANCIATE PRIMARY DATA TO IMPLEMENT FUNCTIONS VALUE INTERFACE
///
//
func NewNative(infs ...interface{}) Native {
	return func() d.Native { return d.New(infs...) }
}

func (n Native) String() string { return n().String() }

func (n Native) Eval(args ...d.Native) d.Native { return n().Eval(args...) }

func (n Native) TypeNat() d.TyNative { return n().TypeNat() }

func (n Native) TypeFnc() TyFnc { return Data }

func (n Native) Empty() bool {
	if n != nil {
		if !d.Nil.Flag().Match(n.TypeNat()) {
			if !None.Flag().Match(n.TypeFnc()) {
				return false
			}
		}
	}
	return true
}

func (n Native) Call(vals ...Parametric) Parametric { return n }

func New(inf ...interface{}) Parametric { return NewFromData(d.New(inf...)) }

func NewFromData(data ...d.Native) Native {
	var result d.Native
	if len(data) == 1 {
		result = data[0]
	} else {
		result = d.DataSlice(data)
	}
	return func() d.Native { return result }
}

// type system implementation
type (
	/// FUNCTOR, APPLICAPLE & MONOID
	FunctorFnc    func(args ...Parametric) (Parametric, FunctorFnc)
	ApplicapleFnc func(args ...Parametric) (Parametric, ApplicapleFnc)
	MonadicFnc    func(args ...Parametric) (Parametric, MonadicFnc)
	// NATIVE DATA (aliased natives implementing parametric)
	Native  func() d.Native
	DataCon func(...d.Native) Native
	/// PURE FUNCTIONS (sole dependece on argset)
	ConstFnc  func() Parametric
	UnaryFnc  func(Parametric) Parametric
	BinaryFnc func(a, b Parametric) Parametric
	NaryFnc   func(...Parametric) Parametric
)

// FUNCTOR
func NewFunctor(resource Consumeable) FunctorFnc {
	return func(args ...Parametric) (Parametric, FunctorFnc) {
		var head, tail = resource.DeCap()
		if head == nil {
			return nil, NewFunctor(tail)
		}
		return head.Call(args...), NewFunctor(tail)
	}
}

func (c FunctorFnc) Call(args ...Parametric) Parametric {
	return FunctorFnc(
		func(...Parametric) (Parametric, FunctorFnc) {
			return c(args...)
		})
}

func (c FunctorFnc) DeCap() (Parametric, Consumeable) { return c() }

func (c FunctorFnc) Head() Parametric { h, _ := c(); return h }

func (c FunctorFnc) Tail() Consumeable { _, t := c(); return t }

func (c FunctorFnc) Eval(args ...d.Native) d.Native { return c.Head().Eval(args...) }

func (c FunctorFnc) Ident() Parametric { return c }

func (c FunctorFnc) TypeFnc() TyFnc { return Resource }

func (c FunctorFnc) TypeNat() d.TyNative { res, _ := c(); return res.TypeNat() | d.Function }

// APPLICAPLE
func NewApplicaple(resource Consumeable) ApplicapleFnc {
	return func(args ...Parametric) (Parametric, ApplicapleFnc) {
		var head, tail = resource.DeCap()
		if head == nil {
			return nil, NewApplicaple(tail)
		}
		return head.Call(args...), NewApplicaple(tail)
	}
}

func (c ApplicapleFnc) Call(args ...Parametric) Parametric {
	return ApplicapleFnc(
		func(...Parametric) (Parametric, ApplicapleFnc) {
			return c(args...)
		})
}

func (c ApplicapleFnc) DeCap() (Parametric, Consumeable) { return c() }

func (c ApplicapleFnc) Head() Parametric { h, _ := c(); return h }

func (c ApplicapleFnc) Tail() Consumeable { _, t := c(); return t }

func (c ApplicapleFnc) Eval(args ...d.Native) d.Native { return c.Head().Eval(args...) }

func (c ApplicapleFnc) Ident() Parametric { return c }

func (c ApplicapleFnc) TypeFnc() TyFnc { return Application }

func (c ApplicapleFnc) TypeNat() d.TyNative { res, _ := c(); return res.TypeNat() | d.Function }

// MONADIC
func NewMonad(resource Consumeable) MonadicFnc {
	return func(args ...Parametric) (Parametric, MonadicFnc) {
		var head, tail = resource.DeCap()
		if head == nil {
			return nil, NewMonad(tail)
		}
		return head.Call(args...), NewMonad(tail)
	}
}

func (c MonadicFnc) Call(args ...Parametric) Parametric {
	return MonadicFnc(
		func(...Parametric) (Parametric, MonadicFnc) {
			return c(args...)
		})
}

func (c MonadicFnc) DeCap() (Parametric, Consumeable) { return c() }

func (c MonadicFnc) Head() Parametric { h, _ := c(); return h }

func (c MonadicFnc) Tail() Consumeable { _, t := c(); return t }

func (c MonadicFnc) Eval(args ...d.Native) d.Native { return c.Head().Eval(args...) }

func (c MonadicFnc) Ident() Parametric { return c }

func (c MonadicFnc) TypeFnc() TyFnc { return Resource }

func (c MonadicFnc) TypeNat() d.TyNative { res, _ := c(); return res.TypeNat() | d.Function }

// DATA CONTSTRUCTOR
func NewDataConstructor(
	expr func(nats ...d.Native) Native,
) DataCon {
	return DataCon(expr)
}

func (t DataCon) String() string { return t().String() }

func (t DataCon) TypeNat() d.TyNative { return d.Function | d.Data }

func (t DataCon) TypeFnc() TyFnc { return Data | Constructor }

func (t DataCon) Call(vals ...Parametric) Parametric {
	var nats = []d.Native{}
	for _, val := range vals {
		nats = append(nats, val)
	}
	return Native(func() d.Native { return t.Eval(nats...) })
}
func (t DataCon) Eval(nats ...d.Native) d.Native { return t(nats...) }

// CONSTANT
//
// constant also conains immutable data that may be an instance of a type of
// the data package, or result of a function call guarantueed to allways return
// the same value.
func (c ConstFnc) Ident() Parametric { return c() }

func (c ConstFnc) TypeFnc() TyFnc { return Function }

func (c ConstFnc) TypeNat() d.TyNative { return c().TypeNat() }

func (c ConstFnc) Eval(p ...d.Native) d.Native { return c().Eval() }

func (c ConstFnc) Call(d ...Parametric) Parametric { return c() }

///// UNARY FUNCTION
func (u UnaryFnc) TypeNat() d.TyNative { return d.Function.TypeNat() }

func (u UnaryFnc) TypeFnc() TyFnc { return Function }

func (u UnaryFnc) Ident() Parametric { return u }

func (u UnaryFnc) Eval(p ...d.Native) d.Native { return u }

func (u UnaryFnc) Call(d ...Parametric) Parametric {
	return u(d[0])
}

///// BINARY FUNCTION
func (b BinaryFnc) TypeNat() d.TyNative { return d.Function.TypeNat() }

func (b BinaryFnc) TypeFnc() TyFnc { return Function }

func (b BinaryFnc) Ident() Parametric { return b }

func (b BinaryFnc) Eval(p ...d.Native) d.Native { return b }

func (b BinaryFnc) Call(d ...Parametric) Parametric { return b(d[0], d[1]) }

///// NARY FUNCTION
func (n NaryFnc) TypeNat() d.TyNative { return d.Function.TypeNat() }

func (n NaryFnc) TypeFnc() TyFnc { return Function }

func (n NaryFnc) Ident() Parametric { return n }

func (n NaryFnc) Eval(p ...d.Native) d.Native { return n }

func (n NaryFnc) Call(d ...Parametric) Parametric { return n(d...) }
