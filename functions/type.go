package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
	"github.com/joergreinhardt/gatwd/lex"
)

type (
	TyFnc     d.BitFlag
	Arity     d.Uint8Val
	Propertys d.Uint8Val
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
	/// TYPE CLASSES
	Numbers
	Strings
	Bytes
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
	/// IO
	IO
	Shared
	Receive
	Transmit
	Control
	Channel
	Buffer
	Reader
	Writer
	ReadWriter
	Condition
	WaitGroup
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

	Consumeables = Collections | Functor | Applicable | Monad | IO
)

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

func (p Propertys) PostFix() bool                  { return p.Flag().Match(PostFix.Flag()) }
func (p Propertys) InFix() bool                    { return !p.Flag().Match(PostFix.Flag()) }
func (p Propertys) Atomic() bool                   { return p.Flag().Match(Atomic.Flag()) }
func (p Propertys) Thunk() bool                    { return !p.Flag().Match(Atomic.Flag()) }
func (p Propertys) Eager() bool                    { return p.Flag().Match(Eager.Flag()) }
func (p Propertys) Lazy() bool                     { return !p.Flag().Match(Eager.Flag()) }
func (p Propertys) RightBound() bool               { return p.Flag().Match(RightBound.Flag()) }
func (p Propertys) LeftBound() bool                { return !p.Flag().Match(RightBound.Flag()) }
func (p Propertys) Mutable() bool                  { return p.Flag().Match(Mutable.Flag()) }
func (p Propertys) Imutable() bool                 { return !p.Flag().Match(Mutable.Flag()) }
func (p Propertys) SideEffect() bool               { return p.Flag().Match(SideEffect.Flag()) }
func (p Propertys) Pure() bool                     { return !p.Flag().Match(SideEffect.Flag()) }
func (p Propertys) Primitive() bool                { return p.Flag().Match(Primitive.Flag()) }
func (p Propertys) Parametric() bool               { return !p.Flag().Match(Primitive.Flag()) }
func (p Propertys) TypeNat() d.TyNat               { return d.Flag }
func (p Propertys) TypeFnc() TyFnc                 { return HigherOrder }
func (p Propertys) Flag() d.BitFlag                { return d.BitFlag(uint64(p)) }
func FlagToProp(flag d.BitFlag) Propertys          { return Propertys(uint8(flag.Uint())) }
func (p Propertys) Eval(a ...d.Native) d.Native    { return p }
func (p Propertys) Call(args ...Callable) Callable { return p }
func (p Propertys) Match(flag d.BitFlag) bool      { return p.Flag().Match(flag) }
func (p Propertys) MatchProperty(arg Propertys) bool {
	if p&arg != 0 {
		return true
	}
	return false
}
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
func (a Arity) Int() int                  { return int(a) }
func (a Arity) Flag() d.BitFlag           { return d.BitFlag(a) }
func (a Arity) TypeNat() d.TyNat          { return d.Flag }
func (a Arity) TypeFnc() TyFnc            { return HigherOrder }
func (a Arity) Match(arg Arity) bool      { return a == arg }

// type TyFnc d.BitFlag
// encodes the kind of functional data as bitflag
func (t TyFnc) FlagType() int8 { return 1 }
func (t TyFnc) TypeName() string {
	var count = t.Flag().Count()
	if count > 1 {
		var str = "["
		for i, flag := range t.Flag().Decompose() {
			str = str + TyFnc(flag.Flag()).String()
			if i < count-1 {
				str = str + " "
			}
		}
		return str + "]"
	}
	return t.String()
}
func (t TyFnc) TypeFnc() TyFnc                 { return Type }
func (t TyFnc) TypeNat() d.TyNat               { return d.Flag }
func (t TyFnc) Call(args ...Callable) Callable { return t.TypeFnc() }
func (t TyFnc) Eval(args ...d.Native) d.Native { return t.TypeNat() }
func (t TyFnc) Flag() d.BitFlag                { return d.BitFlag(t) }
func (t TyFnc) Match(arg d.Typed) bool         { return t.Flag().Match(arg) }
func (t TyFnc) Uint() uint                     { return d.BitFlag(t).Uint() }
