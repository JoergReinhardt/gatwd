package functions

import (
	"strings"
	s "strings"

	d "github.com/joergreinhardt/gatwd/data"
)

type (
	// FLAGS
	TyFlag d.Uint8Val
	TyFnc  d.BitFlag
	TyAri  d.Int8Val
	TyProp d.Int8Val

	// PATTERN
	TySymbol  string
	TyValue   func(...Expression) Expression
	TyPattern []d.Typed
)

//go:generate stringer -type TyFlag
const (
	Flag_BitFlag TyFlag = 0 + iota
	Flag_Native
	Flag_Function
	Flag_KeyWord
	Flag_Symbol
	Flag_Value
	Flag_Token
	Flag_Arity
	Flag_Prop
	Flag_Lex

	Flag_Pattern TyFlag = 255
)

// casts a given flag to its type, based on its flag type and returns it as
// expression
func flagToPattern(elem d.Typed) Expression {
	switch {
	case Flag_Native.Match(elem.FlagType()):
		return Def(elem.(d.TyNat))
	case Flag_BitFlag.Match(elem.FlagType()):
		return Def(elem.(d.BitFlag))
	case Flag_Function.Match(elem.FlagType()):
		return elem.(TyFnc)
	case Flag_Prop.Match(elem.FlagType()):
		return elem.(TyProp)
	case Flag_Arity.Match(elem.FlagType()):
		return elem.(TyAri)
	case Flag_KeyWord.Match(elem.FlagType()):
		return elem.(TyKeyWord)
	case Flag_Lex.Match(elem.FlagType()):
		return elem.(TyLex)
	case Flag_Pattern.Match(elem.FlagType()):
		return elem.(TyPattern)
	case Flag_Symbol.Match(elem.FlagType()):
		return elem.(TySymbol)
	case Flag_Value.Match(elem.FlagType()):
		return elem.(TyValue)
	}
	return DeclareNone()
}

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
	/// EXPRESSION TYPES
	Data
	Value
	Class
	Lambda
	Constant
	Generator
	Accumulator
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
	Then
	Else
	Just
	None
	Either
	Or
	/// CLASSES
	Numbers
	Letters
	Bytes
	Text
	/// SUM
	List
	Vector
	/// PRODUCT
	Set
	Pair
	Enum
	Tuple
	Record
	/// IMPURE
	State
	IO
	/// HIGHER ORDER TYPE
	Parametric

	Kinds = Type | Data | Value | Class | Lambda | Generator
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
	If     = Then | Else
	Maybe  = Just | None
	Option = Either | Or

	Optional = Switch | Case | If | Maybe | Option | Tuple | Record

	//// COLLECTIONS
	SumTypes    = List | Vector
	ProdTypes   = Set | Record | Enum | Tuple
	Collections = SumTypes | ProdTypes
)

// functional type flag expresses the type of a functional value
func (t TyFnc) TypeFnc() TyFnc                     { return Type }
func (t TyFnc) TypeNat() d.TyNat                   { return d.Type }
func (t TyFnc) Flag() d.BitFlag                    { return d.BitFlag(t) }
func (t TyFnc) Uint() d.UintVal                    { return d.BitFlag(t).Uint() }
func (t TyFnc) FlagType() d.Uint8Val               { return Flag_Function.U() }
func (t TyFnc) Call(args ...Expression) Expression { return t.TypeFnc() }
func (t TyFnc) Type() TyPattern                    { return Def(t) }
func (t TyFnc) Match(arg d.Typed) bool             { return t.Flag().Match(arg) }
func (t TyFnc) TypeName() string {
	var count = t.Flag().Count()
	// loop to print concatenated type classes correcty
	if count > 1 {
		switch t {
		case Kinds:
			return "Kinds"
		case Parameter:
			return "Parameter"
		case Truth:
			return "Truth"
		case Trinary:
			return "Trinary"
		case Compare:
			return "Compare"
		case If:
			return "If"
		case Maybe:
			return "Maybe"
		case Option:
			return "Option"
		case Optional:
			return "Branche"
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
func (p TyProp) FlagType() d.Uint8Val               { return Flag_Prop.U() }
func (p TyProp) Type() TyPattern                    { return Def(p) }
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

func (a TyAri) FlagType() d.Uint8Val          { return Flag_Arity.U() }
func (a TyAri) Flag() d.BitFlag               { return d.BitFlag(a) }
func (a TyAri) Type() TyPattern               { return Def(a) }
func (a TyAri) TypeNat() d.TyNat              { return d.Type }
func (a TyAri) TypeFnc() TyFnc                { return Arity }
func (a TyAri) Int() int                      { return int(a) }
func (a TyAri) Match(arg d.Typed) bool        { return a == arg }
func (a TyAri) TypeName() string              { return a.String() }
func (a TyAri) Call(...Expression) Expression { return DecData(d.IntVal(int(a))) }

// type flag representing pattern elements that define a symbol
func DefSym(name string) TySymbol {
	return TySymbol(name)
}
func (n TySymbol) FlagType() d.Uint8Val { return Flag_Symbol.U() }
func (n TySymbol) Flag() d.BitFlag      { return Symbol.Flag() }
func (n TySymbol) Type() TyPattern      { return Def(n) }
func (n TySymbol) TypeFnc() TyFnc       { return Symbol }
func (n TySymbol) String() string       { return string(n) }
func (n TySymbol) TypeName() string     { return string(n) }
func (n TySymbol) Call(args ...Expression) Expression {
	for _, arg := range args {
		if s.Compare(arg.Type().TypeName(), string(n)) != 0 {
			return DecData(d.BoolVal(false))
		}
	}
	return DecData(d.BoolVal(true))
}
func (n TySymbol) Match(typ d.Typed) bool {
	if Flag_Symbol.Match(typ.FlagType()) {
		return s.Compare(string(n),
			string(typ.(TySymbol))) == 0
	}
	return s.Compare(string(n), typ.TypeName()) == 0
}

// type flag representing a pattern element that represents a value
func DefValNative(nat d.Native) TyValue {
	return DefVal(DecNative(nat))
}
func DefValGo(val interface{}) TyValue {
	return DefVal(DecNative(Declare(val)))
}
func DefVal(expr Expression) TyValue {
	return func(args ...Expression) Expression {
		if len(args) > 0 {
			return expr.Call(args...)
		}
		return expr
	}
}
func (n TyValue) FlagType() d.Uint8Val               { return Flag_Value.U() }
func (n TyValue) Flag() d.BitFlag                    { return Value.Flag() }
func (n TyValue) Type() TyPattern                    { return Def(n) }
func (n TyValue) TypeFnc() TyFnc                     { return Value }
func (n TyValue) String() string                     { return n().String() }
func (n TyValue) TypeName() string                   { return n().Type().TypeName() }
func (n TyValue) Call(args ...Expression) Expression { return n(args...) }

func (n TyValue) Match(typ d.Typed) bool {
	if Flag_Value.Match(typ.FlagType()) {
		return true
	}
	return false
}

// pattern of type, property, arity, symbol & value flags
func Def(types ...d.Typed) TyPattern { return types }

// elems yields all elements contained in the pattern
func (p TyPattern) Type() TyPattern               { return p }
func (p TyPattern) Types() []d.Typed              { return p }
func (p TyPattern) Call(...Expression) Expression { return p }
func (p TyPattern) Len() int                      { return len(p.Types()) }
func (p TyPattern) String() string                { return p.TypeName() }
func (p TyPattern) FlagType() d.Uint8Val          { return Flag_Pattern.U() }
func (p TyPattern) Flag() d.BitFlag               { return p.TypeFnc().Flag() }
func (p TyPattern) TypeFnc() TyFnc                { return Pattern }
func (p TyPattern) Get(idx int) TyPattern {
	if idx < p.Len() {
		return p.Patterns()[idx]
	}
	return Def(None)
}

// head yields the first pattern element cast as expression
func (p TyPattern) Head() Expression {
	if p.Len() > 0 {
		var head = p.Patterns()[0]
		return head
	}
	return nil
}

// type-head yields first pattern element as typed
func (p TyPattern) HeadTyped() d.Typed { return p.Head().(d.Typed) }

// type-head yields first pattern element as typed
func (p TyPattern) HeadPattern() TyPattern { return p.Head().(TyPattern) }

// tail yields a consumeable consisting all pattern elements but the first one
// cast as slice of expressions
func (p TyPattern) Tail() Consumeable {
	if p.Len() > 1 {
		return Def(p.Types()[1:]...)
	}
	return TyPattern([]d.Typed{})
}

// tail-type yields a type pattern consisting of all pattern elements but the
// first one
func (p TyPattern) TailPattern() TyPattern {
	if p.Len() > 0 {
		return p.Types()[1:]
	}
	return []d.Typed{}
}

// consume uses head & tail to implement consumeable
func (p TyPattern) Consume() (Expression, Consumeable) { return p.Head(), p.Tail() }

// type-consume works like consume, but yields the head cast as typed & the
// tail as a type pattern
func (p TyPattern) ConsumeTyped() (d.Typed, TyPattern) {
	if p.Len() > 1 {
		return p.Patterns()[0], Def(p.Types()[1:]...)
	}
	if p.Len() > 0 {
		return p.Patterns()[0], []d.Typed{}
	}
	return None, []d.Typed{}
}

// pattern-consume works like type consume, but yields the head converted to,
// or cast as type pattern
func (p TyPattern) ConsumePattern() (TyPattern, TyPattern) {
	return p.HeadPattern(), p.TailPattern()
}

func (p TyPattern) Append(args ...Expression) Consumeable {
	var types = append(make([]d.Typed, 0, p.Len()+len(args)), p...)
	for _, arg := range args {
		if arg.TypeFnc().Match(Type) {
			if typed, ok := arg.(d.Typed); ok {
				types = append(types, typed)
			}
		}
	}
	return TyPattern(types)
}

// pattern yields a slice of type patterns, with all none & nil elements
// filtered out
func (p TyPattern) Patterns() []TyPattern {
	var pattern = make([]TyPattern, 0, p.Len())
	for _, typ := range p.Types() {
		if Flag_Pattern.Match(typ.FlagType()) {
			pattern = append(pattern, typ.(TyPattern))
			continue
		}
		pattern = append(pattern, Def(typ))
	}
	return pattern
}

// expressions that take arguments are expected to also have a type identity
// and return type.
func (p TyPattern) TypeArguments() TyPattern {
	if p.Len() > 2 {
		return p.Patterns()[0]
	}
	return Def(None)
}
func (p TyPattern) ArgumentsName() string {
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
func (p TyPattern) TypeIdent() TyPattern {
	if p.Len() > 2 {
		return p.Patterns()[1]
	}
	if p.Len() > 0 {
		return p.Patterns()[0]
	}
	return Def(None)
}
func (p TyPattern) IdentName() string {
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
func (p TyPattern) TypeReturn() TyPattern {
	if p.Len() > 2 {
		return p.Patterns()[2]
	}
	if p.Len() > 1 {
		return p.Patterns()[1]
	}
	return Def(None)
}
func (p TyPattern) ReturnName() string {
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
func (p TyPattern) TypeElem() TyPattern { return p.TypeIdent() }

func (p TyPattern) TypeName() string {
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
func (p TyPattern) Print(ldelim, sep, rdelim string) string {
	var names = make([]string, 0, p.Len())
	for _, typ := range p.Types() {
		// element is instance of data/typed → print type-name
		if !Flag_Pattern.Match(typ.FlagType()) {
			names = append(names, typ.TypeName())
			continue
		}
		// element is a type pattern
		var pat = typ.(TyPattern)
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
func (p TyPattern) Match(typ d.Typed) bool {
	if Flag_Pattern.Match(typ.FlagType()) {
		if pat, ok := typ.(TyPattern); ok {
			return p.MatchTypes(pat.Types()...)
		}
	}
	return p.MatchTypes(typ)
}

// match-args takes multiple expression arguments and matches their types
// against the elements of the pattern.
func (p TyPattern) MatchTypes(types ...d.Typed) bool {
	var elems, match = p.short(types...)
	for n, elem := range elems {
		if !elem.Match(match[n]) {
			return false
		}
	}
	return true
}
func (p TyPattern) MatchArgs(args ...Expression) bool {
	var types = make([]d.Typed, 0, len(args))
	for _, arg := range args {
		types = append(types, arg.Type())
	}
	return p.MatchTypes(types...)
}

// matches multiple type flags against its elements in order. should there be
// more, or less arguments than pattern elements, the shorter sequence will be
// matched.
func (p TyPattern) short(types ...d.Typed) (elems, match []d.Typed) {
	// if number of arguments is not equal to number of elements, find
	// shorter sequence
	if p.Len() > len(types) {
		elems, match = types, p.Types()
	} else {
		elems, match = p.Types(), types
	}
	return elems, match
}
