package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
	s "strings"
)

type (
	TyFlag    d.Uint8Val
	TyFnc     d.BitFlag
	Arity     d.Int8Val
	Propertys d.Int8Val
	TyPattern []d.Typed
	TySymbol  string
)

//go:generate stringer -type TyFlag
const (
	Flag_BitFlag TyFlag = 0 + iota
	Flag_Function
	Flag_Native
	Flag_KeyWord
	Flag_Symbol
	Flag_Token
	Flag_Arity
	Flag_Prop
	Flag_Lex

	Flag_Pattern TyFlag = 255
)

func CastFlag(elem d.Typed) Expression {
	switch {
	case Flag_Native.Match(elem.FlagType()):
		return ConPattern(elem.(d.TyNat))
	case Flag_BitFlag.Match(elem.FlagType()):
		return ConPattern(elem.(d.BitFlag))
	case Flag_Function.Match(elem.FlagType()):
		return elem.(TyFnc)
	case Flag_Prop.Match(elem.FlagType()):
		return elem.(Propertys)
	case Flag_Arity.Match(elem.FlagType()):
		return elem.(Arity)
	case Flag_KeyWord.Match(elem.FlagType()):
		return elem.(TyKeyWord)
	case Flag_Token.Match(elem.FlagType()):
		return elem.(TyTok)
	case Flag_Lex.Match(elem.FlagType()):
		return elem.(TyLex)
	case Flag_Pattern.Match(elem.FlagType()):
		return elem.(TyPattern)
	case Flag_Symbol.Match(elem.FlagType()):
		return elem.(TySymbol)
	}
	return NewNone()
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
	/// FUNCTION TYPES
	Data
	Value
	Class
	Lambda
	/// PARAMETER OPTIONS
	Property
	Argument
	Pattern
	Return
	Symbol
	Index
	Key
	/// TRUTH VALUE OTIONS
	True
	False
	Undecided
	/// ORDER VALUE OPTIONS
	EQ
	LT
	GT
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
	/// DATA TYPE CLASSES
	Numbers
	Letters
	Bytes
	Text
	/// SUM COLLECTION TYPES
	List
	Vector
	/// PRODUCT COLLECTION TYPES
	Set
	Pair
	Enum
	Tuple
	Record
	/// IMPURE
	State
	IO
	/// HIGHER ORDER TYPE
	HigherORder

	Kinds = Type | Data | Value | Class | Lambda
	//// PARAMETERS
	Params = Key | Index | Pattern | Property | Argument | Return | Symbol

	//// TRUTH & COMPARE
	Truth   = True | False
	Trinary = Truth | Undecided
	Comp    = LT | GT | EQ

	//// OPTIONALS
	If     = Then | Else
	Maybe  = Just | None
	Option = Either | Or

	Branches = Switch | Case | If | Maybe | Option

	//// COLLECTIONS
	Consumeables = List | Vector | Pair
	Enumerables  = Set | Record | Enum | Tuple
	Collections  = Consumeables | Enumerables
)

func ConSymbol(name string) TySymbol {
	return TySymbol(name)
}
func (n TySymbol) TypeFnc() TyFnc       { return Symbol }
func (n TySymbol) FlagType() d.Uint8Val { return Flag_Symbol.U() }
func (n TySymbol) Flag() d.BitFlag      { return Symbol.Flag() }
func (n TySymbol) Type() d.Typed        { return n }
func (n TySymbol) TypeName() string     { return string(n) }
func (n TySymbol) String() string       { return string(n) }
func (n TySymbol) Call(args ...Expression) Expression {
	for _, arg := range args {
		if s.Compare(arg.Type().TypeName(), string(n)) != 0 {
			return NewData(d.BoolVal(false))
		}
	}
	return NewData(d.BoolVal(true))
}
func (n TySymbol) Match(typ d.Typed) bool {
	return s.Compare(string(n), typ.TypeName()) == 0
}

func ConPattern(types ...d.Typed) TyPattern { return types }

func (p TyPattern) Elems() []d.Typed     { return p }
func (p TyPattern) Len() int             { return len(p) }
func (p TyPattern) TypeFnc() TyFnc       { return Pattern }
func (p TyPattern) Type() d.Typed        { return p.TypeFnc() }
func (p TyPattern) FlagType() d.Uint8Val { return Flag_Pattern.U() }
func (p TyPattern) Flag() d.BitFlag      { return p.TypeFnc().Flag() }
func (p TyPattern) String() string       { return p.TypeName() }
func (p TyPattern) TypeName() string     { return p.Print("(", " ", ")") }

func (p TyPattern) Pattern() TyPattern {
	var elems = make([]d.Typed, 0, p.Len())
	for _, elem := range p.Elems() {
		// filter nil & none from pattern
		if Flag_Native.Match(elem.FlagType()) {
			if elem.Match(d.Nil) {
				continue
			}
		}
		if Flag_Function.Match(elem.FlagType()) {
			if elem.Match(None) {
				continue
			}
		}
		elems = append(elems, elem)
	}
	return elems
}

func (p TyPattern) TypeElem() d.Typed {
	if p.Len() > 0 {
		return p.Pattern()[0]
	}
	return Argument
}
func (p TyPattern) Head() Expression {
	if p.Len() > 0 {
		var elem = p.Pattern()[0]
		return CastFlag(elem)
	}
	return nil
}
func (p TyPattern) TypeHead() d.Typed { return p.Head().(TyPattern) }
func (p TyPattern) Tail() Consumeable {
	if p.Len() > 0 {
		return TyPattern(p.Pattern()[1:])
	}
	return TyPattern([]d.Typed{})
}
func (p TyPattern) TypeTail() TyPattern {
	if p.Len() > 0 {
		return p.Elems()[1:]
	}
	return []d.Typed{}
}
func (p TyPattern) TypeConsume() (d.Typed, TyPattern) {
	if p.Len() > 1 {
		return p.Pattern()[0], p.Pattern()[1:]
	}
	if p.Len() > 0 {
		return p.Pattern()[0], []d.Typed{}
	}
	return None, []d.Typed{}
}
func (p TyPattern) Consume() (Expression, Consumeable) {
	return p.Head(), p.Tail()
}
func (p TyPattern) Print(ldelim, sep, rdelim string) string {
	if p.Len() > 1 {
		var slice = make([]string, 0, p.Len())
		for _, elem := range p.Pattern() {
			// recursively print pattern types
			if Flag_Pattern.Match(elem.FlagType()) {
				slice = append(
					slice,
					elem.(TyPattern).Print(
						ldelim, sep, rdelim))
				continue
			}
			if Flag_Function.Match(elem.FlagType()) {
				if elem.Match(Data) {
					slice = append(slice, elem.TypeName())
					continue
				}
			}
			// append type name for all other types
			slice = append(slice, elem.TypeName())
		}
		return ldelim + s.Join(slice, sep) + rdelim
	}
	if p.Len() > 0 {
		var head = p.Elems()[0]
		if !head.Match(Type) {
			return ldelim + head.TypeName() + rdelim
		}
	}
	return ""
}

func (p TyPattern) Call(args ...Expression) Expression {
	var types = make([]d.Typed, 0, len(args))
	for _, arg := range args {
		types = append(types, arg.Type())
	}
	return NewData(d.BoolVal(p.MatchAll(types...)))
}
func (p TyPattern) Match(typ d.Typed) bool {
	if Flag_Pattern.Match(typ.FlagType()) {
		return p.MatchAll(typ.(TyPattern).Pattern()...)
	}
	return p.MatchAll(typ)
}

// matches n'th element of pattern
func (p TyPattern) MatchN(idx int, typ d.Typed) bool {
	if idx < p.Len() {
		if p[idx].FlagType() == typ.FlagType() {
			if p[idx].Match(typ) {
				return true
			}
		}
	}
	return false
}

// matches n'th element of pattern
func (p TyPattern) MatchAll(types ...d.Typed) bool {
	var elems, match []d.Typed
	if p.Len() > len(types) {
		elems, match = types, p
	} else {
		elems, match = p, types
	}
	for n, elem := range elems {
		if elem.FlagType() != match[n].FlagType() ||
			!elem.Match(match[n]) {
			return false
		}
	}
	return true
}

// type TyFnc d.BitFlag
// encodes the kind of functional data as bitflag
func (t TyFnc) TypeFnc() TyFnc                     { return Type }
func (t TyFnc) TypeNat() d.TyNat                   { return d.Type }
func (t TyFnc) Flag() d.BitFlag                    { return d.BitFlag(t) }
func (t TyFnc) Uint() d.UintVal                    { return d.BitFlag(t).Uint() }
func (t TyFnc) FlagType() d.Uint8Val               { return Flag_Function.U() }
func (t TyFnc) Call(args ...Expression) Expression { return t.TypeFnc() }
func (t TyFnc) Type() d.Typed                      { return t }
func (t TyFnc) Match(arg d.Typed) bool             { return t.Flag().Match(arg) }
func (t TyFnc) TypeName() string {
	var count = t.Flag().Count()
	// loop to print concatenated type classes correcty
	if count > 1 {
		switch t {
		case Kinds:
			return "Kinds"
		case Params:
			return "Params"
		case Truth:
			return "Truth"
		case Trinary:
			return "Trinary"
		case Comp:
			return "Comp"
		case If:
			return "Cond"
		case Maybe:
			return "Maybe"
		case Option:
			return "Option"
		case Branches:
			return "Branches"
		case Consumeables:
			return "Consumeables"
		case Collections:
			return "Collections"
		case Enumerables:
			return "Enumerables"
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

func FlagToProp(flag d.BitFlag) Propertys { return Propertys(flag.Uint()) }

func (p Propertys) Flag() d.BitFlag                    { return d.BitFlag(uint64(p)) }
func (p Propertys) FlagType() d.Uint8Val               { return Flag_Prop.U() }
func (p Propertys) TypeNat() d.TyNat                   { return d.Type }
func (p Propertys) TypeFnc() TyFnc                     { return Type }
func (p Propertys) TypeName() string                   { return "Propertys" }
func (p Propertys) Match(flag d.Typed) bool            { return p.Flag().Match(flag) }
func (p Propertys) Eval() d.Native                     { return d.Int8Val(p) }
func (p Propertys) Call(args ...Expression) Expression { return p }
func (p Propertys) Type() d.Typed                      { return p }

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

func (a Arity) Eval() d.Native                { return a }
func (a Arity) FlagType() d.Uint8Val          { return Flag_Arity.U() }
func (a Arity) Int() int                      { return int(a) }
func (a Arity) Type() d.Typed                 { return Type }
func (a Arity) TypeFnc() TyFnc                { return Type }
func (a Arity) TypeNat() d.TyNat              { return d.Type }
func (a Arity) Match(arg d.Typed) bool        { return a == arg }
func (a Arity) TypeName() string              { return a.String() }
func (a Arity) Call(...Expression) Expression { return NewData(a) }
func (a Arity) Flag() d.BitFlag               { return d.BitFlag(a) }
