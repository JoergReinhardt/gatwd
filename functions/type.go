package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

type (
	TyFlag    uint8
	TyFnc     d.BitFlag
	Arity     d.Int8Val
	Propertys d.Uint8Val
	TyComp    func() (string, string, []d.Typed)
)

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
func (a Arity) TypeNat() d.TyNat               { return d.Type }
func (a Arity) TypeFnc() TyFnc                 { return HigherOrder }
func (a Arity) Match(arg Arity) bool           { return a == arg }
func (a Arity) TypeName() string               { return a.String() }

///////////////////////////////////////////////////////////////////////////////
//go:generate stringer -type=TyFnc
const (
	/// KIND
	Type TyFnc = 1 << iota
	Nested
	/// FUNCTION
	Static
	Lambda
	Defined
	/// COLLECTIONS
	List
	Vector
	Set
	/// PRODUCT TYPES
	Pair
	Enum
	Tuple
	Record
	/// PARAMETERS
	Key
	Index
	//// DATA CONSTRUCTORS
	/// TRUTH
	Predicate
	True
	False
	Undecided
	/// ORDER
	Lesser
	Greater
	Equal
	/// ALTERNATIVE
	Just
	None
	Left
	Right
	/// BRANCH
	Case
	Then
	Else
	/// IMPURE
	State
	IO
	/// HIGHER ORDER TYPE
	HigherOrder

	Kind = Type | Nested

	Function = Lambda | Static | Defined

	Collection = List | Vector | Set

	Product = Pair | Enum | Tuple | Record

	Truth = True | False | Undecided

	Order = Lesser | Greater | Equal

	Switch = Case

	If = Then | Else

	Maybe = Just | None

	Either = Left | Right

	Branch = If | Switch

	Impure = State | IO

	Parametric = Kind | Function | Collection | Product |
		Truth | Order | Branch | Impure
)

// type TyFnc d.BitFlag
// encodes the kind of functional data as bitflag
func (t TyFnc) FlagType() uint8                { return 2 }
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

func NewComposedType(name string, args ...d.Native) TyComp {
	// type name is derived by passing name, delimiters, seperator and the
	// slice of types to deriveName.
	var full string
	var types []d.Typed
	name, full, types = nest(name, args)
	return func() (string, string, []d.Typed) {
		return name, full, types
	}
}

func (t TyComp) FlagType() uint8  { return 255 }
func (t TyComp) TypeComp() TyComp { return t }
func (t TyComp) TypeFnc() TyFnc   { return Nested }
func (t TyComp) TypeNat() d.TyNat { return d.Function }
func (t TyComp) String() string   { return t.TypeName() }
func (t TyComp) Flag() d.BitFlag  { return t.TypeFnc().Flag() }
func (t TyComp) Types() []d.Typed { var _, _, types = t(); return types }
func (t TyComp) FullName() string { var _, full, _ = t(); return full }
func (t TyComp) TypeName() string {
	var name, _, _ = t()
	if name == "" {
		name = t.FullName()
	}
	return name
}

// return composed type elements
func (t TyComp) CompTypes() []TyComp {
	var types = []TyComp{}
	for _, typ := range t.Types() {
		if typ.FlagType() == 255 {
			types = append(types, typ.(TyComp))
		}
	}
	return types
}

// return functional type elements
func (t TyComp) FncTypes() []TyFnc {
	var types = []TyFnc{}
	for _, typ := range t.Types() {
		if typ.FlagType() == 2 {
			types = append(types, typ.(TyFnc))
		}
	}
	return types
}

// return native type elements
func (t TyComp) NatTypes() []d.TyNat {
	var types = []d.TyNat{}
	for _, typ := range t.Types() {
		if typ.FlagType() == 1 {
			types = append(types, typ.(d.TyNat))
		}
	}
	return types
}

// evaluation yields a native bool value, by matching the native type flags of
// all it's arguments against it's elements
func (t TyComp) Eval(nats ...d.Native) d.Native {
	var nat = d.BoolVal(true)
	return nat
}

// call yields a functional bool value, by matching all type flags of all
// arguments against its elements.
func (t TyComp) Call(args ...Callable) Callable {
	var val Callable
	return val
}

// match takes an instance of the native typed interface, casts it according to
// flag type and uses call & evaluation to yield a bool value, indicating if
// the type matches.
func (t TyComp) Match(typ d.Typed) bool {
	return true
}

// name is derived by recursive concatenation of functional & composed type
// names, seperated by blank, and by optional a seperator and delimited by
// optional delimiters.
func nest(name string, args []d.Native) (string, string, []d.Typed) {

	var types = []d.Typed{}
	var typ d.Typed
	var num = len(args)
	var full, sep = "", " → "

	for n, arg := range args {

		if fnc, ok := arg.(Callable); ok {
			typ = nestFunctional(fnc)
		} else {
			if nat, ok := arg.(d.Native); ok {
				typ = nestNative(nat)
			}
		}

		types = append(types, typ)
		// call types full method to get its type full
		full = full + typ.TypeName()
		// element separation
		if n < num-1 {
			// elements are seperated by optional seperator and
			// followed by a mandatory blank.
			full = full + sep
		}
	}
	// embed resulting name in left-/ and right delimiter
	return name, full, types
}

func nestFunctional(fnc Callable) d.Typed {

	var typ d.Typed

	// if this is a type flag, assume or concatenated type parameters
	if fnc.TypeFnc().Match(Type) {
		if flag, ok := fnc.(TyFnc); ok {
			return nestFncParametric(flag)
		}
	}

	// if function type matches nested, return composed type
	if fnc.TypeFnc().Match(Nested) {
		if nest, ok := fnc.(CompTyped); ok {
			return nest.TypeFnc()
		}
	}

	// or concatenate instances sub types
	switch {
	}

	return typ
}

func nestNative(nat d.Native) d.Typed {

	var typ d.Typed

	// if this is a type flag, assume or concatenated type parameters
	if nat.TypeNat().Match(d.Type) {
		if flag, ok := nat.(d.TyNat); ok {
			return nestNativeParametric(flag)
		}
	}

	// compose collection types
	if nat.TypeNat().Match(d.Compositions) {
		switch {
		case nat.TypeNat().Match(Pair):
			if pair, ok := nat.(d.PairVal); ok {
				typ = NewComposedType(
					pair.TypeName(),
					pair.LeftType(),
					pair.RightType(),
				)
			}
		case nat.TypeNat().Match(d.Unboxed):
			if ubox, ok := nat.(d.Sliceable); ok {
				typ = NewComposedType(
					ubox.TypeName(),
					ubox.SubType(),
				)
			}
		case nat.TypeNat().Match(d.Slice):
			if slice, ok := nat.(d.Sliceable); ok {
				typ = NewComposedType(
					slice.TypeName(),
					slice.SubType(),
				)
			}
		case nat.TypeNat().Match(d.Map):
			if set, ok := nat.(d.Mapped); ok {
				typ = NewComposedType(
					set.TypeName(),
					set.KeyType(),
					set.ValType(),
				)
			}
		default: // atomic native instance
			typ = nat.TypeNat()
		}
	}
	return typ
}

func nestNativeParametric(typ d.TyNat) d.Typed {

	if typ.Match(d.Flag) || typ.Flag().Count() == 1 {
		return typ
	}

	var nats = []d.Native{}

	for _, flag := range typ.Flag().Decompose() {
		nats = append(nats, flag)
	}

	return NewComposedType(typ.TypeName(), nats...)
}

func nestFncParametric(typ TyFnc) d.Typed {

	if typ.Match(Type) || typ.Flag().Count() == 1 {
		return typ
	}

	var nats = []d.Native{}

	for _, flag := range typ.Flag().Decompose() {
		nats = append(nats, flag)
	}

	return NewComposedType(typ.TypeName(), nats...)
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
func (p Propertys) TypeNat() d.TyNat               { return d.Type }
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
