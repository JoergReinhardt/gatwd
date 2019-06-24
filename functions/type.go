package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

type TyFlag uint8

func (t TyFlag) Match(match TyFlag) bool {
	if t&^match != 0 {
		return false
	}
	return true
}

//go:generate stringer -type TyFlag
const (
	Flag_BitFlag    TyFlag = 0
	Flag_Native            = 1
	Flag_Functional        = 1 << iota
	Flag_Property
	Flag_Syntax
	Flag_Tuple
	Flag_Record
	Flag_Signature
)

type (
	TyFnc     d.BitFlag
	Arity     d.Int8Val
	Propertys d.Uint8Val
)

//// CALL ARITY
///
// arity of well defined callables
//
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

func (a Arity) Eval(args ...d.Native) d.Native { return d.Int8Val(a) }
func (a Arity) Call(...Callable) Callable      { return NewNative(a.Eval()) }
func (a Arity) Int() int                       { return int(a) }
func (a Arity) Flag() d.BitFlag                { return d.BitFlag(a) }
func (a Arity) TypeNat() d.TyNat               { return d.Flag }
func (a Arity) TypeFnc() TyFnc                 { return HigherOrder }
func (a Arity) Match(arg Arity) bool           { return a == arg }
func (a Arity) TypeName() string               { return a.String() }

///////////////////////////////////////////////////////////////////////////////
//go:generate stringer -type=TyFnc
const (
	/// KINDS
	Type TyFnc = 1 << iota
	Data
	Nullable
	Function
	/// COLLECTIONS
	Enum
	List
	Vector
	Set
	Pair
	Tuple
	Record
	/// PARAMETERS
	Key
	Index
	/// CONSTRUCTORS
	Predicate
	True
	False
	Undecided
	Lesser
	Greater
	Equal
	Just
	None
	Case
	Switch
	Left
	Right
	If
	Else
	IO
	/// HIGHER ORDER TYPE
	HigherOrder

	Collection = Enum | List | Vector | Set | Pair | Tuple | Record

	Truth = True | False | Undecided

	Ordering = Lesser | Greater | Equal

	Maybe = Just | None

	Either = Left | Right

	Branch = If | Else

	Sets = Collection | Truth | Ordering | Maybe | Either | Branch
)

// type TyFnc d.BitFlag
// encodes the kind of functional data as bitflag
func (t TyFnc) FlagType() uint8                { return 2 }
func (t TyFnc) TypeFnc() TyFnc                 { return Type }
func (t TyFnc) TypeNat() d.TyNat               { return d.Flag }
func (t TyFnc) Flag() d.BitFlag                { return d.BitFlag(t) }
func (t TyFnc) Uint() uint                     { return d.BitFlag(t).Uint() }
func (t TyFnc) Match(arg d.Typed) bool         { return t.Flag().Match(arg) }
func (t TyFnc) Call(args ...Callable) Callable { return t.TypeFnc() }
func (t TyFnc) Eval(args ...d.Native) d.Native { return t.TypeNat() }
func (t TyFnc) TypeName() string {
	var delim = " "
	var count = t.Flag().Count()
	// loop to print concatenated type classes correcty
	if count > 1 {
		var str string
		for i, flag := range t.Flag().Decompose() {
			str = str + TyFnc(flag.Flag()).String()
			if i < count-1 {
				str = str + delim
			}
		}
		return str
	}
	return t.String()
}

//// CALL PROPERTYS
///
// propertys of well defined callables
//
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

func FlagToProp(flag d.BitFlag) Propertys          { return Propertys(uint8(flag.Uint())) }
func (p Propertys) Flag() d.BitFlag                { return d.BitFlag(uint64(p)) }
func (p Propertys) FlagType() uint8                { return 3 }
func (p Propertys) TypeNat() d.TyNat               { return d.Flag }
func (p Propertys) TypeFnc() TyFnc                 { return HigherOrder }
func (p Propertys) TypeName() string               { return "Propertys" }
func (p Propertys) Match(flag d.Typed) bool        { return p.Flag().Match(flag) }
func (p Propertys) Eval(args ...d.Native) d.Native { return d.Int8Val(p) }
func (p Propertys) Call(args ...Callable) Callable { return p }
func (p Propertys) MatchProperty(arg Propertys) bool {
	if p&arg != 0 {
		return true
	}
	return false
}

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
