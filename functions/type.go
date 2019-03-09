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
func (a Arity) Int() int                    { return int(a) }
func (a Arity) Flag() d.BitFlag             { return d.BitFlag(a) }
func (a Arity) TypeNat() d.TyNative         { return d.Flag }
func (a Arity) TypeFnc() TyFnc              { return HigherOrder }
func (a Arity) Match(arg Arity) bool        { return a == arg }

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

func (p Propertys) TypePrime() d.TyNative       { return d.Flag }
func (p Propertys) TypeFnc() TyFnc              { return HigherOrder }
func (p Propertys) Flag() d.BitFlag             { return p.TypeFnc().Flag() }
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
func (t TyFnc) TypeHO() TyFnc             { return t }
func (t TyFnc) TypeNat() d.TyNative       { return d.Flag }
func (t TyFnc) Flag() d.BitFlag           { return d.BitFlag(t) }
func (t TyFnc) Uint() uint                { return d.BitFlag(t).Uint() }

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
	Aggregator
	Generator
	Constructor
	Functor
	Monad
	///////////
	Condition
	False
	True
	Just
	None
	Case
	Either
	Or
	If
	Else
	Error
	///////////
	Pair
	List
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
	Nests      = Tuple | List
	Sets       = UniSet | MuliSet | AssocVec | Record
	Links      = Link | DLink | Node | Tree
)

////////////////////////////////////////////////////////////////////////////////
//// (RE-) INSTANCIATE PRIMARY DATA TO IMPLEMENT FUNCTIONS VALUE INTERFACE
func (n Native) String() string                 { return n().String() }
func (n Native) Eval(args ...d.Native) d.Native { return n().Eval(args...) }
func (n Native) TypeNat() d.TyNative            { return n().TypeNat() }
func (n Native) TypeFnc() TyFnc                 { return Data }
func (n Native) Call(vals ...Parametric) Parametric {
	switch len(vals) {
	case 0:
		return NewFromData(n.Eval())
	case 1:
		return NewFromData(n.Eval(vals[0].Eval()))
	}
	var nat = n()
	for _, val := range vals {
		vals = append(vals, NewFromData(nat.Eval(val.Eval())))
	}
	return NewVector(vals...)
}
func New(inf ...interface{}) Parametric { return NewFromData(d.New(inf...)) }
func NewFromData(data ...d.Native) Native {
	var nat d.Native
	if len(data) > 0 {
		if len(data) > 1 {
			var nats = []d.Native{}
			for _, dat := range data {
				nats = append(nats, dat)
			}
			nat = d.DataSlice(nats)
		}
		nat = data[0]
	}
	return func() d.Native { return nat }
}

// type system implementation
type (
	// NATIVE DATA (aliased natives implementing parametric)
	Native  func() d.Native
	DataCon func(...d.Native) Native
	/// PURE FUNCTIONS (sole dependece on argset)
	ConstFnc  func() Parametric
	UnaryFnc  func(Parametric) Parametric
	BinaryFnc func(a, b Parametric) Parametric
	NaryFnc   func(...Parametric) Parametric
	/// TYPE SYSTEM
	TypePat  func() (Arity, Propertys, CaseVal)
	FuncDef  func() (TypePat, Parametric)
	FuncExp  func(args ...Parametric) Parametric
	ThunkExp func() []Parametric
	TypeCon  func(...Parametric) TypeId
	TypeId   func() (
		name,
		signature string,
		pattern []TypePat,
	)
	Instance func() (TypeId, Parametric)
)

// CONSTANT
//
// constant also conains immutable data that may be an instance of a type of
// the data package, or result of a function call guarantueed to allways return
// the same value.
func NewConstant(fnc func() Parametric) ConstFnc {
	return ConstFnc(func() Parametric { return fnc() })
}

func (c ConstFnc) Ident() Parametric               { return c() }
func (c ConstFnc) TypeFnc() TyFnc                  { return Function }
func (c ConstFnc) TypeNat() d.TyNative             { return c().TypeNat() }
func (c ConstFnc) Eval(p ...d.Native) d.Native     { return c().Eval() }
func (c ConstFnc) Call(d ...Parametric) Parametric { return c() }

///// UNARY FUNCTION
func NewUnaryFnc(fnc func(f Parametric) Parametric) UnaryFnc {
	return UnaryFnc(func(f Parametric) Parametric { return fnc(f) })
}
func (u UnaryFnc) TypeNat() d.TyNative         { return d.Function.TypeNat() }
func (u UnaryFnc) TypeFnc() TyFnc              { return Function }
func (u UnaryFnc) Ident() Parametric           { return u }
func (u UnaryFnc) Eval(p ...d.Native) d.Native { return u }
func (u UnaryFnc) Call(d ...Parametric) Parametric {
	return u(d[0])
}

///// BINARY FUNCTION
func NewBinaryFnc(fnc func(a, b Parametric) Parametric) BinaryFnc {
	return BinaryFnc(func(a, b Parametric) Parametric { return fnc(a, b) })
}
func (b BinaryFnc) TypeNat() d.TyNative             { return d.Function.TypeNat() }
func (b BinaryFnc) TypeFnc() TyFnc                  { return Function }
func (b BinaryFnc) Ident() Parametric               { return b }
func (b BinaryFnc) Eval(p ...d.Native) d.Native     { return b }
func (b BinaryFnc) Call(d ...Parametric) Parametric { return b(d[0], d[1]) }

///// NARY FUNCTION
func NewNaryFnc(fnc func(f ...Parametric) Parametric) NaryFnc {
	return NaryFnc(func(f ...Parametric) Parametric { return fnc(f...) })
}
func (n NaryFnc) TypeNat() d.TyNative             { return d.Function.TypeNat() }
func (n NaryFnc) TypeFnc() TyFnc                  { return Function }
func (n NaryFnc) Ident() Parametric               { return n }
func (n NaryFnc) Eval(p ...d.Native) d.Native     { return n }
func (n NaryFnc) Call(d ...Parametric) Parametric { return n(d...) }

/// TYPE IDENT
func NewTypeId(
	name, signature string, patterns ...TypePat,
) TypeId {
	return func() (string, string, []TypePat) {
		return name, signature, patterns
	}

}
func (t TypeId) Name() string                      { n, _, _ := t(); return n }
func (t TypeId) Signature() string                 { _, s, _ := t(); return s }
func (t TypeId) Patterns() []TypePat               { _, _, p := t(); return p }
func (t TypeId) String() string                    { return t.Name() + " = " + t.Signature() }
func (t TypeId) TypeNat() d.TyNative               { return d.Function | d.Type }
func (t TypeId) TypeFnc() TyFnc                    { return Type }
func (t TypeId) Eval(nat ...d.Native) d.Native     { return t }
func (t TypeId) Call(val ...Parametric) Parametric { return t }

/// DATA, EXPRESSION & TYPE CONSTRUCTORS
type TyExpr uint8

//go:generate stringer -type=TyExpr
const (
	// heap
	TypeDefinition TyExpr = 1
	TypeContructor TyExpr = 1 << iota
	DataContructor        // func(inf ...interface{}) native
	DataInstance          // native
	FncDefinition         // expression
	FncApplication
	ThunkExpression
	CaseExpression
)

// TYPE CONSTRUCTOR
func NewTypeConstructor(
	expr func(types ...Parametric) TypeId,
) TypeCon {
	return TypeCon(expr)
}
func (t TypeCon) String() string                     { return t().String() }
func (t TypeCon) TypeCon() TyExpr                    { return TypeContructor }
func (t TypeCon) TypeNat() d.TyNative                { return d.Function | d.Type }
func (t TypeCon) TypeFnc() TyFnc                     { return Type | Constructor }
func (t TypeCon) Call(vals ...Parametric) Parametric { return t(vals...) }
func (t TypeCon) Eval(nats ...d.Native) d.Native {
	var parms = []Parametric{}
	for _, nat := range nats {
		parms = append(parms, NewFromData(nat))
	}
	return t.Call(parms...)
}

// DATA CONTSTRUCTOR
func NewDataConstructor(
	expr func(nats ...d.Native) Native,
) DataCon {
	return DataCon(expr)
}
func (t DataCon) String() string      { return t().String() }
func (t DataCon) TypeCon() TyExpr     { return DataContructor }
func (t DataCon) TypeNat() d.TyNative { return d.Function | d.Data }
func (t DataCon) TypeFnc() TyFnc      { return Data | Constructor }
func (t DataCon) Call(vals ...Parametric) Parametric {
	var nats = []d.Native{}
	for _, val := range vals {
		nats = append(nats, val)
	}
	return Native(func() d.Native { return t.Eval(nats...) })
}
func (t DataCon) Eval(nats ...d.Native) d.Native { return t(nats...) }
