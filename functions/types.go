package functions

import (
	"strings"
	s "strings"

	d "github.com/joergreinhardt/gatwd/data"
)

type (
	// FLAG TYPES
	TyKind d.Uint8Val // kind of type flag
	TyFnc  d.BitFlag  // functional primitive type
	TyAri  d.Int8Val  // flag to mark function arity (TODO: tbr?)
	TyProp d.Int8Val  // call property types
	//TyNat d.BitFlag ← defined in data

	// TYPE PATTERN
	TyExp func(...Expression) Expression
	TySym string
	TyPat []d.Typed
)

//go:generate stringer -type TyKind
const (
	Kind_BitFlag TyKind = 0 + iota
	Kind_Native
	Kind_Func
	Kind_KeyWord
	Kind_Symbol
	Kind_Value
	Kind_Token
	Kind_Arity
	Kind_Prop
	Kind_Lexi

	Kind_Comp TyKind = 255
)

func (t TyKind) U() d.Uint8Val { return d.Uint8Val(t) }
func (t TyKind) Match(match d.Uint8Val) bool {
	if match == t.U() {
		return true
	}
	return false
}

//go:generate stringer -type=TyFnc
const (
	/// GENERIC TYPE
	Type TyFnc = 1 << iota
	/// EXPRESSION TYPES
	Data
	Value
	Constant
	Generator
	Accumulator
	Constructor
	/// PARAMETER
	Property
	Argument
	Pattern
	Element
	Lexical
	Symbol
	Arity
	Index
	Key
	/// TRUTH
	True
	False
	Undecided
	/// ORDER
	Equal
	Lesser
	Greater
	/// BOUND
	Min
	Max
	/// OPTIONS
	Switch
	Case
	Just
	None
	Option
	Polymorph
	Either
	Or
	/// CLASSES
	Natural
	Integer
	Real
	Ratio
	Letter
	Text
	Bytes
	/// PRODUCT
	Vector
	List
	Set
	Pair
	/// SUM
	Enum
	Tuple
	Record
	/// COMPOUND
	Monad
	State
	IO
	/// HIGHER ORDER TYPE
	Parametric

	//// PARAMETER
	Parameter = Property | Argument | Pattern | Element |
		Lexical | Symbol | Arity | Index | Key
	//// TRUTH & COMPARE
	Truth   = True | False
	Trinary = Truth | Undecided
	Compare = Lesser | Greater | Equal

	//// BOUNDS
	Bound = Min | Max

	//// OPTIONALS
	Maybe   = Just | None
	Variant = Either | Or

	//// COLLECTIONS
	ProdTypes   = List | Vector | Enum
	SumTypes    = Set | Record | Tuple
	Collections = SumTypes | ProdTypes

	Number = Natural | Integer | Real | Ratio
	String = Letter | Text

	ALL TyFnc = 0xFFFFFFFFFFFFFFFF
)

// functional type flag expresses the type of a functional value
func (t TyFnc) TypeFnc() TyFnc                     { return Type }
func (t TyFnc) TypeNat() d.TyNat                   { return d.Type }
func (t TyFnc) Flag() d.BitFlag                    { return d.BitFlag(t) }
func (t TyFnc) Uint() d.UintVal                    { return d.BitFlag(t).Uint() }
func (t TyFnc) Kind() d.Uint8Val                   { return Kind_Func.U() }
func (t TyFnc) Call(args ...Expression) Expression { return t.TypeFnc() }
func (t TyFnc) Type() TyPat                        { return Def(t) }
func (t TyFnc) Match(arg d.Typed) bool             { return t.Flag().Match(arg) }
func (t TyFnc) TypeName() string {
	var count = t.Flag().Count()
	// loop to print concatenated type classes correcty
	if count > 1 {
		switch t {
		case Parameter:
			return "Parameter"
		case Truth:
			return "Truth"
		case Trinary:
			return "Trinary"
		case Compare:
			return "Compare"
		case Maybe:
			return "Maybe"
		case Variant:
			return "Option"
		case Bound:
			return "Bound"
		case SumTypes:
			return "SumTypes"
		case ProdTypes:
			return "ProductTypes"
		case Collections:
			return "Collections"
		}
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

/// CALL PROPERTYS
//
//go:generate stringer -type TyProp
const (
	Default TyProp = 0
	PostFix TyProp = 1
	InFix   TyProp = 1 + iota
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
func (p TyProp) MatchProperty(arg TyProp) bool {
	if p&arg != 0 {
		return true
	}
	return false
}

// PROPERTY CONVIENIENCE METHODS
func flagToProp(flag d.BitFlag) TyProp { return TyProp(flag.Uint()) }

func (p TyProp) Flag() d.BitFlag                    { return d.BitFlag(uint64(p)) }
func (p TyProp) Kind() d.Uint8Val                   { return Kind_Prop.U() }
func (p TyProp) Type() TyPat                        { return Def(p) }
func (p TyProp) TypeFnc() TyFnc                     { return Property }
func (p TyProp) TypeNat() d.TyNat                   { return d.Type }
func (p TyProp) TypeName() string                   { return "Propertys" }
func (p TyProp) Match(flag d.Typed) bool            { return p.Flag().Match(flag) }
func (p TyProp) Call(args ...Expression) Expression { return p }

func (p TyProp) PostFix() bool    { return p.Flag().Match(PostFix.Flag()) }
func (p TyProp) InFix() bool      { return !p.Flag().Match(PostFix.Flag()) }
func (p TyProp) Atomic() bool     { return p.Flag().Match(Atomic.Flag()) }
func (p TyProp) Thunk() bool      { return !p.Flag().Match(Atomic.Flag()) }
func (p TyProp) Eager() bool      { return p.Flag().Match(Eager.Flag()) }
func (p TyProp) Lazy() bool       { return !p.Flag().Match(Eager.Flag()) }
func (p TyProp) RightBound() bool { return p.Flag().Match(RightBound.Flag()) }
func (p TyProp) LeftBound() bool  { return !p.Flag().Match(RightBound.Flag()) }
func (p TyProp) Mutable() bool    { return p.Flag().Match(Mutable.Flag()) }
func (p TyProp) Imutable() bool   { return !p.Flag().Match(Mutable.Flag()) }
func (p TyProp) SideEffect() bool { return p.Flag().Match(SideEffect.Flag()) }
func (p TyProp) Pure() bool       { return !p.Flag().Match(SideEffect.Flag()) }
func (p TyProp) Primitive() bool  { return p.Flag().Match(Primitive.Flag()) }
func (p TyProp) Parametric() bool { return !p.Flag().Match(Primitive.Flag()) }

/// CALL ARITY
//
// arity of well defined callables
//
//go:generate stringer -type TyAri
const (
	Nary TyAri = -1 + iota
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

func (a TyAri) Kind() d.Uint8Val              { return Kind_Arity.U() }
func (a TyAri) Flag() d.BitFlag               { return d.BitFlag(a) }
func (a TyAri) Type() TyPat                   { return Def(a) }
func (a TyAri) TypeNat() d.TyNat              { return d.Type }
func (a TyAri) TypeFnc() TyFnc                { return Arity }
func (a TyAri) Int() int                      { return int(a) }
func (a TyAri) Match(arg d.Typed) bool        { return a == arg }
func (a TyAri) TypeName() string              { return a.String() }
func (a TyAri) Call(...Expression) Expression { return Box(d.IntVal(int(a))) }

//// TYPE SYMBOL
///
// type flag representing pattern elements that define symbols
func DefSym(name string) TySym   { return TySym(name) }
func (n TySym) Kind() d.Uint8Val { return Kind_Symbol.U() }
func (n TySym) Flag() d.BitFlag  { return Symbol.Flag() }
func (n TySym) Type() TyPat      { return Def(n) }
func (n TySym) TypeFnc() TyFnc   { return Symbol }
func (n TySym) String() string   { return n.TypeName() }
func (n TySym) TypeName() string {
	if strings.Contains(string(n), " ") {
		return "(" + string(n) + ")"
	}
	return string(n)
}
func (n TySym) Call(args ...Expression) Expression {
	for _, arg := range args {
		if s.Compare(arg.Type().TypeName(), string(n)) != 0 {
			return Box(d.BoolVal(false))
		}
	}
	return Box(d.BoolVal(true))
}
func (n TySym) Match(typ d.Typed) bool {
	if Kind_Symbol.Match(typ.Kind()) {
		return s.Compare(string(n),
			string(typ.(TySym))) == 0
	}
	return s.Compare(string(n), typ.TypeName()) == 0
}

//// TYPE EXPRESSION
///
// type flag representing a parametric element in a type pattern by a value, or
// type-expression, expecting and returning typeflags as its values.
func DefValGo(val interface{}) TyExp {
	return DefVal(Dat(d.New(val)))
}
func DefValNat(nat d.Native) TyExp {
	return DefVal(Dat(nat))
}
func DefVal(expr Expression) TyExp {
	return func(args ...Expression) Expression {
		if len(args) > 0 {
			return expr.Call(args...)
		}
		return expr.Call()
	}
}
func (n TyExp) Kind() d.Uint8Val                   { return Kind_Value.U() }
func (n TyExp) Flag() d.BitFlag                    { return Value.Flag() }
func (n TyExp) Type() TyPat                        { return Def(n) }
func (n TyExp) TypeFnc() TyFnc                     { return Value }
func (n TyExp) String() string                     { return n().String() }
func (n TyExp) TypeName() string                   { return n().Type().TypeName() }
func (n TyExp) Value() Expression                  { return n() }
func (n TyExp) Call(args ...Expression) Expression { return n(args...) }

func (n TyExp) Match(typ d.Typed) bool {
	if Kind_Value.Match(typ.Kind()) {
		return true
	}
	return false
}

///////////////////////////////////////////////////////////////////////////////
//// TYPE PATTERN
///
// pattern of type, property, arity, symbol & value flags
func Def(args ...d.Typed) TyPat { return args }

// elems yields all elements contained in the pattern
func (p TyPat) Type() TyPat                   { return p }
func (p TyPat) Types() []d.Typed              { return p }
func (p TyPat) Call(...Expression) Expression { return p } // ← TODO: match arg instances
func (p TyPat) Len() int                      { return len(p.Types()) }
func (p TyPat) String() string                { return p.TypeName() }
func (p TyPat) Kind() d.Uint8Val              { return Kind_Comp.U() }
func (p TyPat) Flag() d.BitFlag               { return p.TypeFnc().Flag() }
func (p TyPat) TypeFnc() TyFnc                { return Pattern }

// length of elements excluding fields set to none
func (p TyPat) Count() int {
	var count int
	for _, elem := range p {
		if !elem.Match(None) {
			count = count + 1
		}
	}
	return count
}
func (p TyPat) Get(idx int) TyPat {
	if idx < p.Len() {
		return p.Pattern()[idx]
	}
	return Def(None)
}

// head yields the first pattern element cast as expression
func (p TyPat) Head() Expression {
	if p.Len() > 0 {
		var head = p.Pattern()[0]
		return head
	}
	return nil
}

// type-head yields first pattern element as typed
func (p TyPat) HeadTyped() d.Typed { return p.Head().(d.Typed) }

// type-head yields first pattern element as typed
func (p TyPat) HeadPattern() TyPat { return p.Head().(TyPat) }

// tail yields a consumeable consisting all pattern elements but the first one
// cast as slice of expressions
func (p TyPat) Tail() Consumeable {
	if p.Len() > 1 {
		return Def(p.Types()[1:]...)
	}
	return TyPat([]d.Typed{})
}

// tail-type yields a type pattern consisting of all pattern elements but the
// first one
func (p TyPat) TailPattern() TyPat {
	if p.Len() > 0 {
		return p.Types()[1:]
	}
	return []d.Typed{}
}

// consume uses head & tail to implement consumeable
func (p TyPat) Consume() (Expression, Consumeable) { return p.Head(), p.Tail() }

// type-consume works like consume, but yields the head cast as typed & the
// tail as a type pattern
func (p TyPat) ConsumeTyped() (d.Typed, TyPat) {
	if p.Len() > 1 {
		return p.Pattern()[0], Def(p.Types()[1:]...)
	}
	if p.Len() > 0 {
		return p.Pattern()[0], []d.Typed{}
	}
	return None, []d.Typed{}
}

// pattern-consume works like type consume, but yields the head converted to,
// or cast as type pattern
func (p TyPat) ConsumePattern() (TyPat, TyPat) {
	return p.HeadPattern(), p.TailPattern()
}

func (p TyPat) Cons(args ...Expression) Sequential {
	var types = make([]Expression, 0, p.Len())
	for _, pat := range p {
		types = append(types, pat.(TyPat))
	}
	return NewVector(append(args, types...)...)
}

func (p TyPat) Append(args ...Expression) Sequential {
	var types = make([]Expression, 0, p.Len())
	for _, pat := range p {
		types = append(types, pat.(TyPat))
	}
	return NewVector(append(types, args...)...)
}

// pattern yields a slice of type patterns, with all none & nil elements
// filtered out
func (p TyPat) Pattern() []TyPat {
	var pattern = make([]TyPat, 0, p.Len())
	for _, typ := range p.Types() {
		if Kind_Comp.Match(typ.Kind()) {
			pattern = append(pattern, typ.(TyPat))
			continue
		}
		pattern = append(pattern, Def(typ))
	}
	return pattern
}

// pattern yields a slice of type patterns, with all none & nil elements
// filtered out
func (p TyPat) Elements() []d.Typed {
	var elems = make([]d.Typed, 0, p.Count())
	for _, elem := range p {
		if Kind_Native.Match(elem.Kind()) {
			if elem.Match(d.Nil) {
				continue
			}
		}
		if elem.Match(None) {
			continue
		}
		elems = append(elems, elem)
	}
	return elems
}
func (p TyPat) Fields() []TyPat {
	var elems = make([]TyPat, 0, p.Count())
	for _, elem := range p.Elements() {
		if Kind_Native.Match(elem.Kind()) {
			if elem.Match(d.Nil) {
				continue
			}
		}
		if elem.Match(None) {
			continue
		}
		if Kind_Comp.Match(elem.Kind()) {
			elems = append(elems, elem.(TyPat))
			continue
		}
		elems = append(elems, Def(elem))
	}
	return elems
}

// expressions that take arguments are expected to also have a type identity
// and return type.
func (p TyPat) TypeArguments() TyPat {
	if p.Len() > 2 {
		return p.Pattern()[0]
	}
	return Def(None)
}
func (p TyPat) ArgumentsName() string {
	if p.TypeArguments().Len() > 0 {
		if !p.TypeArguments().Match(None) {
			if p.TypeArguments().Len() > 1 {
				var ldelim, sep, rdelim = "", " → ", ""
				return p.TypeArguments().Print(
					ldelim, sep, rdelim,
				)
			}
			return p.TypeArguments().Print("", " ", "")
		}
	}
	return ""
}

// each type is expected to have a type identity, which is the second last
// element in the types pattern
func (p TyPat) TypeIdent() TyPat {
	if p.Len() > 2 {
		return p.Pattern()[1]
	}
	if p.Len() > 0 {
		return p.Pattern()[0]
	}
	return Def(None)
}
func (p TyPat) IdentName() string {
	if p.TypeIdent().Len() > 0 {
		if !p.TypeIdent().Match(None) {
			if p.TypeIdent().Len() > 1 {
				var ldelim, sep, rdelim = "(", " ", ")"
				switch {
				case p.TypeIdent().Match(List | Vector):
					ldelim, sep, rdelim = "[", " ", "]"
				case p.TypeIdent().Match(Set):
					ldelim, sep, rdelim = "{", " ", "}"
				case p.TypeIdent().Match(Record):
					ldelim, sep, rdelim = "{", " ∷ ", "}"
				case p.TypeIdent().Match(Tuple):
					ldelim, sep, rdelim = "(", " | ", ")"
				case p.TypeIdent().Match(Enum):
					ldelim, sep, rdelim = "[", " | ", "]"
				}
				return p.TypeIdent().Print(
					ldelim, sep, rdelim,
				)
			}
			return p.TypeIdent().Print("", " ", "")
		}
	}
	return ""
}

// each type is expected to have an return type, which equals the last element
// in the types pattern
func (p TyPat) TypeReturn() TyPat {
	if p.Len() > 2 {
		return p.Pattern()[2]
	}
	if p.Len() > 1 {
		return p.Pattern()[1]
	}
	return Def(None)
}
func (p TyPat) ReturnName() string {
	if p.TypeReturn().Len() > 0 {
		if !p.TypeReturn().Match(None) {
			if p.TypeReturn().Len() > 1 {
				var ldelim, sep, rdelim = "(", " ", ")"
				switch {
				case p.TypeIdent().Match(List | Vector):
					ldelim, sep, rdelim = "[", " ", "]"
				case p.TypeIdent().Match(Set):
					ldelim, sep, rdelim = "{", " ", "}"
				case p.TypeIdent().Match(Record):
					ldelim, sep, rdelim = "{", " ∷ ", "}"
				case p.TypeIdent().Match(Tuple):
					ldelim, sep, rdelim = "(", " | ", ")"
				case p.TypeIdent().Match(Enum):
					ldelim, sep, rdelim = "[", " | ", "]"
				}
				return p.TypeReturn().Print(
					ldelim, sep, rdelim,
				)
			}
			return p.TypeReturn().Print("", " ", "")
		}
	}
	return ""
}

// type-elem yields the first elements typed
func (p TyPat) TypeElem() TyPat { return p.TypeIdent() }

func (p TyPat) TypeName() string {
	var strs = []string{}
	if !p.TypeArguments().Match(None) {
		strs = append(strs, p.ArgumentsName())
	}
	if !p.TypeIdent().Match(None) {
		strs = append(strs, p.IdentName())
	}
	if !p.TypeReturn().Match(None) {
		strs = append(strs, p.ReturnName())
	}
	return strings.Join(strs, " → ")
}

// print converts pattern to string, seperating the elements with a seperator
// and putting sub patterns in delimiters. seperator and delimiters are passed
// to the method. sub patterns are printed recursively.
func (p TyPat) Print(ldelim, sep, rdelim string) string {
	var names = make([]string, 0, p.Len())
	for _, typ := range p.Types() {
		// element is instance of data/typed → print type-name
		if !Kind_Comp.Match(typ.Kind()) {
			names = append(names, typ.TypeName())
			continue
		}
		// element is a type pattern
		var pat = typ.(TyPat)
		// print type pattern with delimiters and separator
		names = append(names, pat.Print(ldelim, sep, rdelim))
	}
	// print elements wrapped in delimiters, seperated by seperator
	return ldelim + strings.Join(names, sep) + rdelim
}

// match takes its argument, evaluated by passing it to the match-args method
// and yields the resulting bool. should the argument be a pattern itself, all
// its sub elements are evaluated to match sub patterns recursively, when
// called by match-all method.
func (p TyPat) Match(typ d.Typed) bool {
	if Kind_Comp.Match(typ.Kind()) {
		if pat, ok := typ.(TyPat); ok {
			return p.MatchTypes(pat.Types()...)
		}
	}
	return p.MatchTypes(typ)
}

// match-types takes multiple types and matches them against an equal number of
// arguments starting with the first one
func (p TyPat) MatchTypes(types ...d.Typed) bool {
	var short, long = p.sortLength(types...)
	for n, elem := range short {
		if !elem.Match(long[n]) {
			return false
		}
	}
	return true
}
func (p TyPat) MatchAnyType(args ...d.Typed) bool {
	for _, elem := range p {
		for _, arg := range args {
			if elem.Match(arg) {
				return true
			}
		}
	}
	return false
}

// match-args takes multiple expression arguments and matches their types
// against the elements of the pattern.
func (p TyPat) MatchArgs(args ...Expression) bool {
	var types = make([]d.Typed, 0, len(args))
	for _, arg := range args {
		types = append(types, arg.Type())
	}
	return p.MatchTypes(types...)
}
func (p TyPat) MatchAnyArg(args ...Expression) bool {
	var types = make([]d.Typed, 0, len(args))
	for _, arg := range args {
		types = append(types, arg.Type())
	}
	return p.MatchAnyType(types...)
}

// matches multiple type flags against its elements in order. should there be
// more, or less arguments than pattern elements, the shorter sequence will be
// matched.
func (p TyPat) sortLength(types ...d.Typed) (short, long []d.Typed) {
	// if number of arguments is not equal to number of elements, find
	// shorter sequence
	if p.Len() > len(types) {
		short, long = types, p.Types()
	} else {
		short, long = p.Types(), types
	}
	return short, long
}

// bool methods
func (p TyPat) HasIdentity() bool {
	if p.Count() > 0 {
		return true
	}
	return false
}
func (p TyPat) HasReturnType() bool {
	if p.Count() > 1 {
		return true
	}
	return false
}
func (p TyPat) HasArguments() bool {
	if p.Count() > 2 {
		return true
	}
	return false
}

// one element pattern is a type identity
func (p TyPat) IsIdentity() bool {
	if p.Count() == 1 {
		return true
	}
	return false
}
func (p TyPat) IsAtomic() bool {
	if p.IsIdentity() {
		return !strings.ContainsAny(p.Elements()[0].TypeName(), " |,:")
	}
	return false
}
func (p TyPat) IsTruth() bool {
	if p.Count() == 1 {
		return p.Elements()[0].Match(Truth)
	}
	return false
}
func (p TyPat) IsTrinary() bool {
	if p.Count() == 1 {
		return p.Elements()[0].Match(Trinary)
	}
	return false
}
func (p TyPat) IsCompare() bool {
	if p.Count() == 1 {
		return p.Elements()[0].Match(Compare)
	}
	return false
}
func (p TyPat) IsParameter() bool {
	if p.Count() == 2 {
		return p.Elements()[0].Match(Parameter)
	}
	return false
}
func (p TyPat) IsData() bool {
	if p.Count() == 2 {
		return p.Elements()[0].Match(Data)
	}
	return false
}
func (p TyPat) IsPair() bool {
	if p.Count() == 2 {
		return p.Elements()[0].Match(Pair)
	}
	return false
}
func (p TyPat) IsVector() bool {
	if p.Count() == 2 {
		return p.Elements()[0].Match(Vector)
	}
	return false
}
func (p TyPat) IsList() bool {
	if p.Count() == 2 {
		return p.Elements()[0].Match(List)
	}
	return false
}
func (p TyPat) IsCollection() bool {
	if p.Count() == 2 {
		return p.Elements()[0].Match(Collections)
	}
	return false
}
func (p TyPat) IsEnum() bool {
	if p.Count() == 2 {
		return p.Elements()[0].Match(Enum)
	}
	return false
}
func (p TyPat) IsTuple() bool {
	if p.Count() == 2 {
		return p.Elements()[0].Match(Tuple)
	}
	return false
}
func (p TyPat) IsRecord() bool {
	if p.Count() == 2 {
		return p.Elements()[0].Match(Record)
	}
	return false
}
func (p TyPat) IsSet() bool {
	if p.Count() == 2 {
		return p.Elements()[0].Match(Set)
	}
	return false
}
func (p TyPat) IsSwitch() bool {
	if p.Count() == 2 {
		return p.Elements()[0].Match(Switch)
	}
	return false
}
func (p TyPat) IsNumber() bool {
	if p.Count() == 2 {
		return p.Elements()[0].Match(Number)
	}
	return false
}
func (p TyPat) IsString() bool {
	if p.Count() == 2 {
		return p.Elements()[0].Match(String)
	}
	return false
}
func (p TyPat) IsBytes() bool {
	if p.Count() == 2 {
		return p.Elements()[0].Match(Bytes)
	}
	return false
}
func (p TyPat) IsSumType() bool {
	if p.Count() == 2 && (p.IsList() || p.IsVector() || p.IsEnum()) {
		return true
	}
	return false
}
func (p TyPat) IsProductType() bool {
	if p.Count() == 2 && (p.IsTuple() || p.IsRecord() || p.IsPair() || p.IsSet()) {
		return true
	}
	return false
}
func (p TyPat) IsCase() bool {
	if p.Count() == 3 {
		return p.Elements()[1].Match(Case)
	}
	return false
}
func (p TyPat) IsMaybe() bool {
	if p.Count() == 3 {
		return p.Elements()[1].Match(Maybe)
	}
	return false
}
func (p TyPat) IsAlternative() bool {
	if p.Count() == 3 {
		return p.Elements()[1].Match(Variant)
	}
	return false
}
func (p TyPat) IsFunction() bool {
	if p.Count() == 3 {
		return p.Elements()[1].Match(Value)
	}
	return false
}
func (p TyPat) IsParametric() bool {
	if p.Count() == 3 {
		return p.Elements()[1].Match(Parametric)
	}
	return false
}

// two element pattern is a constant type returning a value type
func (p TyPat) HasData() bool        { return p.MatchAnyType(Data) }
func (p TyPat) HasPair() bool        { return p.MatchAnyType(Pair) }
func (p TyPat) HasEnum() bool        { return p.MatchAnyType(Enum) }
func (p TyPat) HasTuple() bool       { return p.MatchAnyType(Tuple) }
func (p TyPat) HasRecord() bool      { return p.MatchAnyType(Record) }
func (p TyPat) HasParameter() bool   { return p.MatchAnyType(Parameter) }
func (p TyPat) HasTruth() bool       { return p.MatchAnyType(Truth) }
func (p TyPat) HasTrinary() bool     { return p.MatchAnyType(Trinary) }
func (p TyPat) HasCompare() bool     { return p.MatchAnyType(Compare) }
func (p TyPat) HasBound() bool       { return p.MatchAnyType(Min, Max) }
func (p TyPat) HasMaybe() bool       { return p.MatchAnyType(Maybe) }
func (p TyPat) HasAlternative() bool { return p.MatchAnyType(Variant) }
func (p TyPat) HasNumber() bool      { return p.MatchAnyType(Number) }
func (p TyPat) HasString() bool      { return p.MatchAnyType(String) }
func (p TyPat) HasBytes() bool       { return p.MatchAnyType(Bytes) }
func (p TyPat) HasCollection() bool {
	return p.MatchAnyType(
		List, Vector, Tuple, Enum, Record)
}
func (p TyPat) HasReturn() bool {
	if p.Count() >= 2 {
		return true
	}
	return false
}
func (p TyPat) HasArgs() bool {
	if p.Count() >= 3 {
		return true
	}
	return false
}
