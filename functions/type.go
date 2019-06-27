package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

type (
	TyDef     func() (string, Callable)
	TyFlag    d.Uint8Val
	TyFnc     d.BitFlag
	Arity     d.Int8Val
	Propertys d.Int8Val
)

//go:generate stringer -type TyFlag
const (
	Flag_BitFlag TyFlag = 0 + iota
	Flag_Native
	Flag_Functional
	Flag_Arity
	Flag_Prop

	Flag_Comp TyFlag = 255
)

func (t TyFlag) U() d.Uint8Val { return d.Uint8Val(t) }
func (t TyFlag) Match(match TyFlag) bool {
	if match == t {
		return false
	}
	return true
}

//go:generate stringer -type=TyFnc
const (
	/// KIND
	Type TyFnc = 1 << iota
	/// GENERIC FUNCTION TYPES
	Static
	Lambda
	Defined
	/// PARAMETER OPTIONS
	Key
	Index
	Return
	Argument
	Property
	/// TRUTH VALUE types
	True
	False
	Undecided
	Predicate
	/// BRANCH TYPES
	If
	Case
	Switch
	/// VALUE OPTIONS
	Then
	Else
	Just
	None
	Either
	Or
	/// IMPURE
	State
	IO
	/// HIGHER ORDER TYPE
	HigherOrder

	//// FUNCTIONS
	Function = Static | Lambda | Defined

	//// PARAMETERS
	Paramer = Key | Index | Argument |
		Return | Property

	//// COLLECTIONS
	Sum     = List | Vector
	Product = Set | Pair |
		Enum | Tuple | Record

	// TRUTH VALUES & PREDICATES
	Truth   = True | False
	Trinary = Truth | Undecided

	//// OPTIONALS
	Maybe  = Just | None
	Option = Either | Or

	Optional = Maybe | Either

	//// branch value types
	Branch = If | Case | Switch

	/// ORDER CLASS PARAMETER
	Lesser
	Greater
	Equal
	/// BOUND CLASS PARAMETER
	Min
	Max
	/// SUM COLLECTION TYPE & VALUE CONSTRUCTORS
	List
	Vector
	/// PRODUCT COLLECTION TYPE & VALUE CONSTRUCTORS
	Pair
	Set
	Enum
	Tuple
	Record
	// COLLECTION CLASS
	Collection = List | Vector | Pair |
		Set | Enum | Tuple | Record
)

//// type definition
func Define(name string, cons Callable) TyDef {
	return func() (string, Callable) { return name, cons }
}

func (t TyDef) Ident() TyDef                   { return t }
func (t TyDef) TypeComp() TyDef                { return t }
func (t TyDef) Type() Typed                    { return Type }
func (t TyDef) TypeFnc() TyFnc                 { return Type }
func (t TyDef) TypeNat() d.TyNat               { return d.Type }
func (t TyDef) Eval(nats ...d.Native) d.Native { return d.Function }
func (t TyDef) Flag() d.BitFlag                { return Type.Flag() }
func (t TyDef) FlagType() d.Uint8Val           { return Flag_Comp.U() }
func (t TyDef) String() string                 { return t.TypeName() }
func (t TyDef) TypeName() string               { var name, _ = t(); return name }
func (t TyDef) Constructor() Callable          { var _, cons = t(); return cons }
func (t TyDef) Call(args ...Callable) Callable { return t.Constructor().Call(args...) }
func (t TyDef) ConsType() TyFnc                { return t.Constructor().TypeFnc() }
func (t TyDef) Match(typ d.Typed) bool         { return true }

// type TyFnc d.BitFlag
// encodes the kind of functional data as bitflag
func (t TyFnc) FlagType() d.Uint8Val           { return Flag_Functional.U() }
func (t TyFnc) Type() Typed                    { return Type }
func (t TyFnc) TypeFnc() TyFnc                 { return Type }
func (t TyFnc) TypeNat() d.TyNat               { return d.Type }
func (t TyFnc) Flag() d.BitFlag                { return d.BitFlag(t) }
func (t TyFnc) Uint() uint                     { return d.BitFlag(t).Uint() }
func (t TyFnc) Match(arg d.Typed) bool         { return t.Flag().Match(arg) }
func (t TyFnc) Call(args ...Callable) Callable { return t.TypeFnc() }
func (t TyFnc) Eval(args ...d.Native) d.Native { return t.TypeNat() }
func (t TyFnc) TypeName() string {
	var count = t.Flag().Count()
	// loop to print concatenated type classes correcty
	if count > 1 {
		var delim = "|"
		var str string
		for i, flag := range t.Flag().Decompose() {
			str = str + TyFnc(flag.Flag()).String()
			if i < count-1 {
				str = str + delim
			}
		}
		return "[" + str + "]"
	}
	return t.String()
}

//// CALL PROPERTYS
///
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
	RightBound
	// ⌐: Left_Bound
	Mutable
	// ⌐: Imutable
	SideEffect
	// ⌐: Pure
	Primitive
	// ⌐: Parametric
)

// CALL PROPERTY FLAG
func (p Propertys) MatchProperty(arg Propertys) bool {
	if p&arg != 0 {
		return true
	}
	return false
}

// PROPERTY CONVIENIENCE METHODS
func (p Propertys) PostFix() bool    { return p.Flag().Match(PostFix.Flag()) }
func (p Propertys) InFix() bool      { return !p.Flag().Match(PostFix.Flag()) }
func (p Propertys) Atomic() bool     { return p.Flag().Match(Atomic.Flag()) }
func (p Propertys) Thunk() bool      { return !p.Flag().Match(Atomic.Flag()) }
func (p Propertys) Eager() bool      { return p.Flag().Match(Eager.Flag()) }
func (p Propertys) Lazy() bool       { return !p.Flag().Match(Eager.Flag()) }
func (p Propertys) RightBound() bool { return p.Flag().Match(RightBound.Flag()) }
func (p Propertys) LeftBound() bool  { return !p.Flag().Match(RightBound.Flag()) }
func (p Propertys) Mutable() bool    { return p.Flag().Match(Mutable.Flag()) }
func (p Propertys) Imutable() bool   { return !p.Flag().Match(Mutable.Flag()) }
func (p Propertys) SideEffect() bool { return p.Flag().Match(SideEffect.Flag()) }
func (p Propertys) Pure() bool       { return !p.Flag().Match(SideEffect.Flag()) }
func (p Propertys) Primitive() bool  { return p.Flag().Match(Primitive.Flag()) }
func (p Propertys) Parametric() bool { return !p.Flag().Match(Primitive.Flag()) }

func FlagToProp(flag d.BitFlag) Propertys { return Propertys(uint8(flag.Uint())) }

func (p Propertys) Flag() d.BitFlag                { return d.BitFlag(uint64(p)) }
func (p Propertys) FlagType() d.Uint8Val           { return Flag_Prop.U() }
func (p Propertys) TypeNat() d.TyNat               { return d.Type }
func (p Propertys) TypeFnc() TyFnc                 { return Type }
func (p Propertys) Type() Typed                    { return Property }
func (p Propertys) TypeName() string               { return "Propertys" }
func (p Propertys) Match(flag d.Typed) bool        { return p.Flag().Match(flag) }
func (p Propertys) Eval(args ...d.Native) d.Native { return d.Int8Val(p) }
func (p Propertys) Call(args ...Callable) Callable { return p }

//// CALL ARITY
///
// arity of well defined callables
//
//go:generate stringer -type Arity
const (
	Nary Arity = -1 + iota
	Nullary
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

func (a Arity) Eval(args ...d.Native) d.Native { return a }
func (a Arity) FlagType() d.Uint8Val           { return Flag_Arity.U() }
func (a Arity) Int() int                       { return int(a) }
func (a Arity) TypeFnc() TyFnc                 { return Type }
func (a Arity) TypeNat() d.TyNat               { return d.Type }
func (a Arity) Match(arg d.Typed) bool         { return a == arg }
func (a Arity) TypeName() string               { return a.String() }
func (a Arity) Call(...Callable) Callable      { return NewNative(a) }
func (a Arity) Flag() d.BitFlag                { return d.BitFlag(a) }
