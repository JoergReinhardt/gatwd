package functions

import (
	"strings"

	d "github.com/joergreinhardt/gatwd/data"
	"github.com/joergreinhardt/gatwd/lex"
)

//go:generate stringer -type=TyFnc
const (
	/// KIND FLAGS ///
	Type TyFnc = 1 << iota
	Native
	Data
	Expression
	Key
	Index
	/// EXPRESSION CALL PROPERTYS
	CallArity
	CallPropertys
	/// COLLECTION TYPES
	List
	Vector
	Tuple
	Record
	Enum
	Set
	Pair
	/// FUNCTORS AND MONADS
	Applicable
	Operator
	Functor
	Monad
	/// MONADIC SUB TYPES
	Undecided
	False
	True
	Equal
	Lesser
	Greater
	Just
	None
	Case
	Switch
	Either
	Or
	If
	Else
	Do
	While
	/// HIGHER ORDER TYPE
	HigherOrder

	Collections = List | Vector | Tuple | Record | Enum |
		Set | Pair

	Options = Undecided | False | True | Equal | Lesser |
		Greater | Just | None | Case | Switch | Either |
		Or | If | Else | Do | While

	Parameters = CallPropertys | CallArity

	Kinds = Type | Native | Data | Expression

	Truth = Undecided | False | True

	Ordered = Equal | Lesser | Greater

	Maybe = Just | None

	CaseSwitch = Case | Switch

	Alternatives = Either | Or

	Branch = If | Else

	Continue = Do | While

	Functors = Applicable | Operator | Functor | Monad
)

// expression type, call propertys & arity
type (
	TySig     func() (string, Typed, Typed)
	TySum     func() (TySig, []TySig)
	TyProd    func() (TySig, TyCon)
	TyCon     func(...Typed) Typed
	TyFnc     d.BitFlag
	Arity     d.Uint8Val
	Propertys d.Uint8Val
)

///////////////////////////////////////////////////////////////////////////////
func NewTyProd(sig TySig, con TyCon) TyProd {
	return func() (TySig, TyCon) { return sig, con }
}
func (t TyProd) Signature() TySig {
	var sig, _ = t()
	return sig
}
func (t TyProd) String() string {
	return t.Signature().String()
}
func (t TyProd) TypeName() string {
	return t.Signature().TypeName()
}
func (t TyProd) FlagType() int8 {
	return t.Signature().FlagType()
}
func (t TyProd) TypeNat() d.TyNat {
	return t.Signature().TypeNat()
}
func (t TyProd) TypeFnc() TyFnc {
	return t.Signature().TypeFnc()
}
func (t TyProd) Flag() d.BitFlag {
	return t.Signature().Flag()
}
func (t TyProd) Eval(args ...d.Native) d.Native {
	return t.Signature().Eval(args...)
}
func (t TyProd) Call(args ...Callable) Callable {
	return t.Signature().Call(args...)
}

///////////////////////////////////////////////////////////////////////////////
func NewTySum(args ...TySig) TySum {
	return func() (TySig, []TySig) {
		if len(args) > 0 {
			if len(args) > 1 {
				return args[0], args[1:]
			}
			return args[0], []TySig{}
		}
		return NewTySig(d.Nil, None, "None"),
			[]TySig{}
	}
}

func (t TySum) Signature() TySig {
	var sig, _ = t()
	return sig
}
func (t TySum) String() string {
	return t.Signature().String()
}
func (t TySum) TypeName() string {
	return t.Signature().TypeName()
}
func (t TySum) FlagType() int8 {
	return t.Signature().FlagType()
}
func (t TySum) TypeNat() d.TyNat {
	return t.Signature().TypeNat()
}
func (t TySum) TypeFnc() TyFnc {
	return t.Signature().TypeFnc()
}
func (t TySum) Flag() d.BitFlag {
	return t.Signature().Flag()
}
func (t TySum) Eval(args ...d.Native) d.Native {
	return t.Signature().Eval(args...)
}
func (t TySum) Call(args ...Callable) Callable {
	return t.Signature().Call(args...)
}

///////////////////////////////////////////////////////////////////////////////
func NewTySig(nat, fnt Typed, names ...string) TySig {
	return func() (string, Typed, Typed) {
		return strings.Join(names, "·"), nat, fnt
	}
}

// returns the name given to a dynamicly defined type by definition
func (t TySig) TypeName() string {
	var name, _, _ = t()
	return name
}

func (t TySig) FlagType() int8 { return 3 }

// type-native method reflects the derived typed native type
func (t TySig) TypeNat() d.TyNat {
	var _, nat, _ = t()
	return nat.(d.TyNat)
}

// type-functional method reflects the derived types functional type
func (t TySig) TypeFnc() TyFnc {
	var _, _, fnt = t()
	return fnt.(TyFnc)
}

// string concatenates the string representations of both type flags with the
// types given name, if one exists
func (t TySig) String() string {
	var name, nat, fnt = t()
	return nat.String() + "·" + fnt.String() + "·" + name
}

// flag method returns the native type as bitflag
func (t TySig) Flag() d.BitFlag { return t.TypeNat().Flag() }

// eval method OR concatenates argument flags with the types native type flag.
func (t TySig) Eval(args ...d.Native) d.Native {
	var flag = t.Flag()
	if len(args) > 0 {
		for _, arg := range args {
			flag = flag | arg.TypeNat().Flag()
		}
	}
	return flag
}

// call OR concatenates the arguments functional types.
func (t TySig) Call(args ...Callable) Callable {
	var flag = t.TypeFnc().Flag()
	if len(args) > 0 {
		for _, arg := range args {
			flag = flag | arg.TypeNat().Flag()
		}
	}
	return NewFromData(flag)
}

///////////////////////////////////////////////////////////////////////////////
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

func (p Propertys) TypeNat() d.TyNat { return d.Flag }
func (p Propertys) TypeFnc() TyFnc   { return HigherOrder }

func (p Propertys) Flag() d.BitFlag { return d.BitFlag(uint64(p)) }

func FlagToProp(flag d.BitFlag) Propertys { return Propertys(uint8(flag.Uint())) }

func (p Propertys) Eval(a ...d.Native) d.Native { return p }

func (p Propertys) Call(args ...Callable) Callable { return p }

func (p Propertys) MatchProperty(arg Propertys) bool {
	if p&arg != 0 {
		return true
	}
	return false
}

func (p Propertys) Match(flag d.BitFlag) bool { return p.Flag().Match(flag) }
func (p Propertys) Print() string {

	var flags = p.Flag().Decompose()
	var str string
	var l = len(flags)

	if l > 1 {
		for i, typed := range flags {

			if typed.FlagType() == 0 {

				str = str + typed.(d.TyNat).String()
			}

			if typed.FlagType() == 1 {

				str = str + typed.(TyFnc).String()
			}

			if typed.FlagType() == 2 {

				str = str + typed.(lex.TySyntax).String()
			}

			if i < l-1 {
				str = str + " "
			}
		}
	}

	return p.String()
}

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

func (a Arity) Eval(...d.Native) d.Native { return d.Int8Val(a) }

func (a Arity) Call(...Callable) Callable { return NewFromData(a.Eval()) }

func (a Arity) Int() int         { return int(a) }
func (a Arity) Flag() d.BitFlag  { return d.BitFlag(a) }
func (a Arity) TypeNat() d.TyNat { return d.Flag }
func (a Arity) TypeFnc() TyFnc   { return HigherOrder }

func (a Arity) Match(arg Arity) bool { return a == arg }

// type TyFnc d.BitFlag
// encodes the kind of functional data as bitflag
func (t TyFnc) FlagType() int8                 { return 2 }
func (t TyFnc) TypeName() string               { return t.String() }
func (t TyFnc) TypeFnc() TyFnc                 { return Type }
func (t TyFnc) TypeNat() d.TyNat               { return d.Flag }
func (t TyFnc) Call(args ...Callable) Callable { return t.TypeFnc() }
func (t TyFnc) Eval(args ...d.Native) d.Native { return t.TypeNat() }
func (t TyFnc) Flag() d.BitFlag                { return d.BitFlag(t) }
func (t TyFnc) Uint() uint                     { return d.BitFlag(t).Uint() }
