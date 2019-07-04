package functions

import (
	"strings"

	d "github.com/joergreinhardt/gatwd/data"
)

type (
	TySig     func() (TyDef, []TyDef)
	TyDef     func() (string, []Expression)
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
	Flag_TypeCons
	Flag_DataCons
	Flag_Arity
	Flag_Prop

	Flag_Def TyFlag = 255
)

func (t TyFlag) U() d.Uint8Val { return d.Uint8Val(t) }
func (t TyFlag) Match(match d.Uint8Val) bool {
	if match == t.U() {
		return true
	}
	return false
}

//go:generate stringer -type=TyFnc
const (
	/// GENERIC TYPE
	Type TyFnc = 1 << iota
	/// FUNCTION TYPES
	Data
	Constant
	Function
	/// PARAMETER OPTIONS
	Property
	Argument
	Return
	Index
	Key
	/// TRUTH VALUE OTIONS
	True
	False
	Undecided
	/// ORDER VALUE OPTIONS
	Lesser
	Greater
	Equal
	/// BOUND VALUE OPTIONS
	Min
	Max
	/// VALUE OPTIONS
	Switch
	Case
	Then
	Else
	Just
	None
	Either
	Or
	/// SUM COLLECTION TYPES
	List
	Vector
	/// PRODUCT COLLECTION TYPES
	Pair
	Set
	Enum
	Tuple
	Record
	/// IMPURE
	State
	IO
	/// HIGHER ORDER TYPE
	HigherOrder

	Kinds = Type | Data | Constant | Function

	//// PARAMETERS
	Signature = Argument | Return
	Parameter = Key | Index | Property

	Parameters = Signature | Parameter

	//// TRUTH & COMPARE
	Truth   = True | False
	Trinary = Truth | Undecided
	Compare = Lesser | Greater | Equal

	Tests = Truth | Trinary | Compare

	//// OPTIONALS
	If     = Then | Else
	Maybe  = Just | None
	Option = Either | Or

	Branches = Switch | Case | If | Maybe | Option

	//// COLLECTIONS
	CollectSum  = List | Vector
	CollectProd = Set | Pair |
		Enum | Tuple | Record

	Collections = CollectSum | CollectProd

	AllTypes = Kinds | Parameters | Tests |
		Branches | Collections
)

//// TYPE DEFINITION
func Define(name string, retype Expression, paratypes ...Expression) TyDef {
	return func() (string, []Expression) {
		return name, append([]Expression{retype}, paratypes...)
	}
}

func (t TyDef) Type() TyDef                        { return t }
func (t TyDef) Name() string                       { var name, _ = t(); return name }
func (t TyDef) Elems() []Expression                { var _, expr = t(); return expr }
func (t TyDef) Return() Expression                 { return t.Elems()[0] }
func (t TyDef) FlagType() d.Uint8Val               { return Flag_Def.U() }
func (t TyDef) String() string                     { return t.Return().String() }
func (t TyDef) TypeFnc() TyFnc                     { return t.Return().TypeFnc() }
func (t TyDef) TypeNat() d.TyNat                   { return t.Return().TypeNat() }
func (t TyDef) Eval(args ...d.Native) d.Native     { return t.Return().Eval(args...) }
func (t TyDef) Call(args ...Expression) Expression { return t.Return().Call(args...) }
func (t TyDef) Pattern() []Expression {
	var elems = t.Elems()
	if len(elems) > 1 {
		return elems[1:]
	}
	return []Expression{}
}
func (t TyDef) Arity() Arity {
	return Arity(len(t.Pattern()))
	return Arity(0)
}
func (t TyDef) ReturnName() string {
	var retname = t.Return().TypeName()
	if strings.Contains(retname, " ") {
		retname = "(" + retname + ")"
	}
	return retname
}
func (t TyDef) PatternName() string {
	if t.Arity() > Arity(0) {
		var slice []string
		var sep = " → "
		var pattern = t.Pattern()
		if len(pattern) > 0 {
			for _, arg := range pattern {
				slice = append(slice,
					arg.TypeName())
			}
			return strings.Join(slice, sep)
		}
	}
	return ""
}
func (t TyDef) TypeName() string {
	var sep = " → "
	var name = t.Name()
	if strings.Contains(name, " ") {
		name = "(" + name + ")"
	}
	if name == "" {
		name = t.ReturnName()
	}
	if t.Arity() > Arity(0) {
		var slice []string
		slice = append(slice, t.PatternName(),
			name, t.ReturnName())
		return strings.Join(slice, sep)
	}
	return name
}
func (t TyDef) Match(typ d.Typed) bool { return true }

// type TyFnc d.BitFlag
// encodes the kind of functional data as bitflag
func (t TyFnc) TypeFnc() TyFnc                     { return Type }
func (t TyFnc) TypeNat() d.TyNat                   { return d.Type }
func (t TyFnc) Flag() d.BitFlag                    { return d.BitFlag(t) }
func (t TyFnc) Uint() uint                         { return d.BitFlag(t).Uint() }
func (t TyFnc) FlagType() d.Uint8Val               { return Flag_Functional.U() }
func (t TyFnc) Match(arg d.Typed) bool             { return t.Flag().Match(arg) }
func (t TyFnc) Call(args ...Expression) Expression { return t.TypeFnc() }
func (t TyFnc) Eval(args ...d.Native) d.Native     { return t.TypeNat() }
func (t TyFnc) Type() TyDef                        { return Define(t.TypeName(), t) }
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
		return str
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

func (p Propertys) Flag() d.BitFlag                    { return d.BitFlag(uint64(p)) }
func (p Propertys) FlagType() d.Uint8Val               { return Flag_Prop.U() }
func (p Propertys) TypeNat() d.TyNat                   { return d.Type }
func (p Propertys) TypeFnc() TyFnc                     { return Type }
func (p Propertys) TypeName() string                   { return "Propertys" }
func (p Propertys) Match(flag d.Typed) bool            { return p.Flag().Match(flag) }
func (p Propertys) Eval(args ...d.Native) d.Native     { return d.Int8Val(p) }
func (p Propertys) Call(args ...Expression) Expression { return p }
func (p Propertys) Type() TyDef {
	return Define(p.TypeName(), Property)
}

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
func (a Arity) Call(...Expression) Expression  { return NewNative(a) }
func (a Arity) Flag() d.BitFlag                { return d.BitFlag(a) }
