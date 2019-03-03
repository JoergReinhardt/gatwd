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
	// ⌐: Composit
	Eager
	// ⌐: Lazy
	Right
	// ⌐: Left_Bound
	Mutable
	// ⌐: Imutable
	SideEffect
	// ⌐: Pure
	Primitive
	// ⌐: Functional
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
	Application
	///////////
	Variable
	Function
	Closure
	Resourceful
	///////////
	Argument
	Parameter
	Accessor
	Attribut
	Predicate
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
	EitherOr   = Either | Or
	IfElse     = If | Else
	Chain      = Vector | Tuple | Record
	AccIndex   = Vector | Chain
	AccSymbol  = Tuple | AssocVec | Record
	AccCollect = AccIndex | AccSymbol
	Nests      = Tuple | List
	Sets       = UniSet | MuliSet | AssocVec | Record
	Links      = Link | DLink | Node | Tree
)

// type system implementation
type (
	Native   func() d.Native
	TypeId   func() (name, signature string)
	DataCon  func(...Functional) Instance
	TypeCon  func(...TypeId) TypeId
	Instance func() (TypeId, Functional)
	FncDef   func() (
		id TypeId,
		props Propertys,
		expr Functional,
		equation string,
		argTypes []TypeId,
	)
)

////////////////////////////////////////////////////////////////////////////////
//// (RE-) INSTANCIATE PRIMARY DATA TO IMPLEMENT FUNCTIONS VALUE INTERFACE
///
// allocate new functional data values from untyped data, or data that has been
// initialized to implement data packages 'native' interface before.
func New(inf ...interface{}) Functional       { return conNative(d.New(inf...)) }
func NewFromData(data ...d.Native) Functional { return conNative(d.NewFromPrimary(data...)) }

func conNative(nat d.Native) Native             { return func() d.Native { return nat } }
func (n Native) String() string                 { return n().String() }
func (n Native) Eval(args ...d.Native) d.Native { return n().Eval(args...) }
func (n Native) TypeNat() d.TyNative            { return n().TypeNat() }
func (n Native) TypeFnc() TyFnc                 { return Data }
func (n Native) Call(vals ...Functional) Functional {
	switch len(vals) {
	case 0:
		return conNative(n.Eval())
	case 1:
		return conNative(n.Eval(vals[0].Eval()))
	}
	var nat = n()
	for _, val := range vals {
		vals = append(vals, conNative(nat.Eval(val.Eval())))
	}
	return NewVector(vals...)
}

////////////////////////////////////////////////////////////////////////////////
//// TYPE IDENT
///
// types are identified by name and defined by a signature signature
func (t TypeId) Name() string                      { n, _ := t(); return n }
func (t TypeId) Signature() string                 { _, p := t(); return p }
func (t TypeId) String() string                    { return t.Name() + " = " + t.Signature() }
func (t TypeId) TypeNat() d.TyNative               { return d.Function | d.Type }
func (t TypeId) TypeFnc() TyFnc                    { return Type }
func (t TypeId) Eval(nat ...d.Native) d.Native     { return t }
func (t TypeId) Call(val ...Functional) Functional { return t }

func NewTypeId(
	name, pattern string,
) TypeId {
	return func() (string, string) {
		return name, pattern
	}

}

///////////////////////////////////////////////////////////////////////////////
//// FUNCTION DEFINITION
///
// function definition returns a list of argument types, call propertys bitwise
// encoded as 8bit flag & and the expression defining the function body
func DefineFnc(
	id TypeId,
	props Propertys,
	expr Functional, equation string,
	args ...TypeId,
) FncDef {
	return FncDef(func() (TypeId, Propertys, Functional, string, []TypeId) {
		return id, props, expr, equation, args
	})
}
func (fd FncDef) Id() TypeId          { tid, _, _, _, _ := fd(); return tid }
func (fd FncDef) Props() Propertys    { _, prop, _, _, _ := fd(); return prop }
func (fd FncDef) Expr() Functional    { _, _, expr, _, _ := fd(); return expr }
func (fd FncDef) Equation() string    { _, _, _, equa, _ := fd(); return equa }
func (fd FncDef) ArgTypes() []TypeId  { _, _, _, _, args := fd(); return args }
func (fd FncDef) Arity() Arity        { return Arity(len(fd.ArgTypes())) }
func (fd FncDef) TypeFnc() TyFnc      { return Type | Function | Definition }
func (fd FncDef) TypeNat() d.TyNative { return d.Function | d.Type }

///////////////////////////////////////////////////////////////////////////////
//// DATA CONSTRUCTOR
///
// constructs instances of functional data value types from previously
// untyped-, or data initialized as instance of a native type from the data
// package, from it's arguments.  when called without arguments, constructor
// returns ident and signature as defined during declaration.
func NewDataConstructor(name, pattern string) DataCon {
	return func(val ...Functional) Instance {
		switch len(val) {
		case 0: // reveals the type this constructor constructs & a
			// 'none' instance
			return func() (TypeId, Functional) {
				return NewTypeId(name, pattern), NewNone()
			}
		case 1:
			return func() (TypeId, Functional) {
				return NewTypeId(name, pattern), val[0]
			}
		}
		return func() (TypeId, Functional) {
			return NewTypeId(name, pattern), NewVector(val...)
		}
	}
}
func (c DataCon) Ident() Functional   { return c }
func (c DataCon) TypeFnc() TyFnc      { return Type | Data | Constructor }
func (c DataCon) TypeNat() d.TyNative { return d.Type | d.Data }
func (c DataCon) String() string {
	tid, _ := c()()
	return tid.String()
}
func (c DataCon) Eval(n ...d.Native) d.Native {
	switch len(n) {
	case 0:
		return d.NilVal{}
	case 1:
		return n[0]
	}
	return d.DataSlice(n)
}
func (c DataCon) Call(d ...Functional) Functional {
	switch len(d) {
	case 0:
		return NewNone()
	case 1:
		return d[0]
	}
	return NewVector(d...)
}
