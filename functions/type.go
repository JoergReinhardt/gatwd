package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
	"github.com/joergreinhardt/gatwd/lex"
)

type (
	TyComp    func() (string, []Typed)
	TyFnc     d.BitFlag
	Arity     d.Int8Val
	Propertys d.Uint8Val
)

//go:generate stringer -type=TyFnc
const (
	/// KIND FLAGS ///
	Type TyFnc = 1 << iota
	Data
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
	Element
	List
	Vector
	Tuple
	Record
	Enum
	Set
	Pair
	/// FUNCTORS AND MONADS
	Constructor
	Functor
	Applicable
	Monad
	/// MONADIC SUB TYPES
	Undecided
	Predicate
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
	Buffer
	Reader
	Writer
	/// HIGHER ORDER TYPE
	HigherOrder

	Collections = List | Vector | Tuple | Record | Enum |
		Set | Pair

	Options = Undecided | False | True | Equal | Lesser |
		Greater | Just | None | Case | Switch | Either |
		Or | If | Else | Do | While

	Parameters = CallPropertys | CallArity

	Kinds = Type | Data | Functor

	Truth = Undecided | False | True

	Ordered = Equal | Lesser | Greater

	Maybe = Just | None

	Alternatives = Either | Or

	Branch = If | Else

	Continue = Do | While

	IO = Buffer | Reader | Writer

	Consumeables = Collections | Applicable | Monad | IO
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

func FlagToProp(flag d.BitFlag) Propertys          { return Propertys(uint8(flag.Uint())) }
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
func (p Propertys) Eval() d.Native                 { return d.Int8Val(p) }
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

			if typed.FlagType() == 1 {

				str = str + typed.(d.TyNat).String()
			}

			if typed.FlagType() == 2 {

				str = str + typed.(TyFnc).String()
			}

			if typed.FlagType() == 3 {

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

func (a Arity) Eval() d.Native            { return d.Int8Val(a) }
func (a Arity) Call(...Callable) Callable { return NewNative(a.Eval()) }
func (a Arity) Int() int                  { return int(a) }
func (a Arity) Flag() d.BitFlag           { return d.BitFlag(a) }
func (a Arity) TypeNat() d.TyNat          { return d.Flag }
func (a Arity) TypeFnc() TyFnc            { return HigherOrder }
func (a Arity) Match(arg Arity) bool      { return a == arg }

// type TyFnc d.BitFlag
// encodes the kind of functional data as bitflag
func (t TyFnc) FlagType() uint8 { return 2 }
func (t TyFnc) TypeName() string {
	var count = t.Flag().Count()
	if count > 1 {
		var str string
		for i, flag := range t.Flag().Decompose() {
			str = str + TyFnc(flag.Flag()).String()
			if i < count-1 {
				str = str + "·"
			}
		}
		return str
	}
	return t.String()
}
func (t TyFnc) TypeFnc() TyFnc                 { return Type }
func (t TyFnc) TypeNat() d.TyNat               { return d.Flag }
func (t TyFnc) Call(args ...Callable) Callable { return t.TypeFnc() }
func (t TyFnc) Eval() d.Native                 { return t.TypeNat() }
func (t TyFnc) Flag() d.BitFlag                { return d.BitFlag(t) }
func (t TyFnc) Match(arg d.Typed) bool         { return t.Flag().Match(arg) }
func (t TyFnc) Uint() uint                     { return d.BitFlag(t).Uint() }

///////////////////////////////////////////////////////////////////////////////
//// COMPOSED TYPE
///
// composition type to define higher order types. it returns a type name that
// has either been passed during creation, or derived from the type names of
// it's elements and a slice of instances implementing the typed interface to
// implement recursively nested higher order types of arbitrary depth and
// complexity.
func NewComposedType(name string, types ...Typed) TyComp {
	if name == "" {
		name = concatTypeNames(types...)
	}
	return func() (string, []Typed) {
		return name, types
	}
}

func concatTypeNames(types ...Typed) string {
	var str string
	var length = len(types)
	for n, t := range types {
		str = str + t.TypeName()
		if n < length-1 {
			str = str + " "
		}
	}
	return str
}

// higher order type has the highest possible value assigned as its flag type
func (t TyComp) FlagType() uint8 { return 254 }

// return isolated type name
func (t TyComp) TypeName() string { name, _ := t(); return name }

// return isolated slice of typed instances
func (t TyComp) AllFlags() []Typed { _, flags := t(); return flags }

// returns all native type flags as slice of typed instances
func (t TyComp) NatFlags() []d.TyNat {
	var flags = []d.TyNat{}
	for _, flag := range t.AllFlags() {
		if flag.FlagType() == 1 {
			flags = append(flags, flag.(d.TyNat))
		}
	}
	return flags
}

// returns all functional type flags as slice of typed instances
func (t TyComp) FncFlags() []TyFnc {
	var flags = []TyFnc{}
	for _, flag := range t.AllFlags() {
		if flag.FlagType() == 2 {
			flags = append(flags, flag.(TyFnc))
		}
	}
	return flags
}

// OR concatenate all flags
func (t TyComp) Flag() d.BitFlag {
	var flags d.BitFlag
	for _, flag := range t.AllFlags() {
		flags = flags | flag.Flag()
	}
	return flags
}

// OR concatenate all function types
func (t TyComp) TypeFnc() TyFnc {
	var flags TyFnc
	for _, flag := range t.FncFlags() {
		flags = flags | flag
	}
	return flags
}

// OR concatenate all native types
func (t TyComp) TypeNat() d.TyNat {
	var flags d.TyNat
	for _, flag := range t.NatFlags() {
		flags = flags | flag
	}
	return flags
}

// call method returns a vector instance containing all functional type flags
func (t TyComp) Call(...Callable) Callable {
	var args = []Callable{}
	for _, arg := range t.FncFlags() {
		args = append(args, arg)
	}
	return NewVector(args...)
}

// eval method renders data slice of all native type flags
func (t TyComp) Eval() d.Native { return d.NewSlice() }

// matching function for composed higher order types will get a wee bit more
// complicated, since it will have to implement of vital parts of the higher
// order type system
func (t TyComp) Match(arg d.Typed) bool { return t.Flag().Match(arg) }

// returns string representation of recursively nested type, also non trivial
func (t TyComp) String() string { return t.TypeName() }
