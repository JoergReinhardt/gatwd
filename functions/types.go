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
	TyProp d.Int8Val  // call property types
	//TyNat d.BitFlag ← defined in data

	// TYPE PATTERN
	TySym  string
	TyComp []d.Typed
	TyExp  func(...Expression) Expression
)

//go:generate stringer -type TyKind
const (
	Kind_BitFlag TyKind = 0 + iota
	Kind_Nat
	Kind_Fnc
	Kind_Key
	Kind_Sym
	Kind_Expr
	Kind_Prop
	Kind_Lex

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
	Element
	Lexical
	Symbol
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
	HashMap
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

// helper functions, to convert between slices of data/typed & ty-pattern
// instances
func typedToComp(typ []d.Typed) []TyComp {
	var pat = make([]TyComp, 0, len(typ))
	for _, t := range typ {
		if Kind_Comp.Match(t.Kind()) {
			pat = append(pat, t.(TyComp))
			continue
		}
		pat = append(pat, Def(t))
	}
	return pat
}
func typedToExpr(typ []d.Typed) []Expression {
	var elems = make([]Expression, 0, len(typ))
	for _, comp := range typedToComp(typ) {
		elems = append(elems, comp)
	}
	return elems
}
func compToTyped(pat []TyComp) []d.Typed {
	var typ = make([]d.Typed, 0, len(pat))
	for _, p := range pat {
		typ = append(typ, p)
	}
	return typ
}
func compToExpr(comps []TyComp) []Expression {
	var elems = make([]Expression, 0, len(comps))
	for _, comp := range comps {
		elems = append(elems, comp)
	}
	return elems
}

// functional type flag expresses the type of a functional value
func (t TyFnc) TypeFnc() TyFnc                     { return Type }
func (t TyFnc) TypeNat() d.TyNat                   { return d.Type }
func (t TyFnc) Flag() d.BitFlag                    { return d.BitFlag(t) }
func (t TyFnc) Uint() d.UintVal                    { return d.BitFlag(t).Uint() }
func (t TyFnc) Kind() d.Uint8Val                   { return Kind_Fnc.U() }
func (t TyFnc) Call(args ...Expression) Expression { return t.TypeFnc() }
func (t TyFnc) Type() TyComp                       { return Def(t) }
func (t TyFnc) Match(arg d.Typed) bool             { return t.Flag().Match(arg) }
func (t TyFnc) TypeName() string {
	var count = t.Flag().Count()
	// loop to print concatenated type classes correcty
	if count > 1 {
		switch t {
		case ALL:
			return "*"
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
	Default TyProp = 0 + iota
	PostFix
	PreFix
	InFix
	Atomic
	Thunk
	Eager
	Lazy
	Right
	Left
	Mutable
	Imutable
	Primitive
	Effected
	Pure
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
func (p TyProp) Type() TyComp                       { return Def(p) }
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
func (p TyProp) RightBound() bool { return p.Flag().Match(Right.Flag()) }
func (p TyProp) LeftBound() bool  { return !p.Flag().Match(Right.Flag()) }
func (p TyProp) Mutable() bool    { return p.Flag().Match(Mutable.Flag()) }
func (p TyProp) Imutable() bool   { return !p.Flag().Match(Mutable.Flag()) }
func (p TyProp) SideEffect() bool { return p.Flag().Match(Effected.Flag()) }
func (p TyProp) Pure() bool       { return !p.Flag().Match(Effected.Flag()) }
func (p TyProp) Primitive() bool  { return p.Flag().Match(Primitive.Flag()) }
func (p TyProp) Parametric() bool { return !p.Flag().Match(Primitive.Flag()) }

//// TYPE SYMBOL
///
// type flag representing pattern elements that define symbols
func DefSym(name string) TySym   { return TySym(name) }
func (n TySym) Kind() d.Uint8Val { return Kind_Sym.U() }
func (n TySym) Flag() d.BitFlag  { return Symbol.Flag() }
func (n TySym) Type() TyComp     { return Def(n) }
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
	return s.Compare(string(n), typ.TypeName()) == 0
}

//// TYPE EXPRESSION
///
// type flag representing a parametric element in a type pattern by a value, or
// type-expression, expecting and returning typeflags as its values.
func DefValGo(val interface{}) TyExp {
	return TyExp(func(...Expression) Expression { return Dat(val) })
}
func DefValNat(nat d.Native) TyExp {
	return TyExp(func(...Expression) Expression { return Box(nat) })
}
func DefExpr(expr Expression) TyExp {
	return func(args ...Expression) Expression {
		if len(args) > 0 {
			return expr.Call(args...)
		}
		return expr.Call()
	}
}
func (n TyExp) Kind() d.Uint8Val                   { return Kind_Expr.U() }
func (n TyExp) Flag() d.BitFlag                    { return Value.Flag() }
func (n TyExp) Type() TyComp                       { return Def(n) }
func (n TyExp) TypeFnc() TyFnc                     { return Value }
func (n TyExp) String() string                     { return n().String() }
func (n TyExp) TypeName() string                   { return n().Type().TypeName() }
func (n TyExp) Value() Expression                  { return n() }
func (n TyExp) Call(args ...Expression) Expression { return n(args...) }

func (n TyExp) Match(typ d.Typed) bool {
	if Kind_Expr.Match(typ.Kind()) {
		var expr = typ.(TyExp).Value()
		// call this type expression with the arguments value, if
		// argument is data
		if expr.Type().MatchAnyType(Data) {
			if n.Call(expr).Type().Match(None) {
				return false
			}
		}
		// the argument expression passing this type expression as value
		if expr.Call(n.Value()).Type().Match(None) {
			return false
		}
	}
	// if the argument is a composed type pattern, apply that pattern as
	// argument and check if the result is none
	if Kind_Comp.Match(typ.Kind()) {
		if n.Call(typ.(TyComp)).Type().Match(None) {
			return false
		}
	}
	if n.Call(Def(typ)).Type().Match(None) {
		return false
	}
	return true
}

///////////////////////////////////////////////////////////////////////////////
//// TYPE PATTERN
///
// defines a new type according to a slice of possibly nested data/typed
// instances.  arguments are expected to be passed in troublesome irish order:
// -  I type *identity*
// -  R *return* value type
// -  A *argument* type set
func Def(types ...d.Typed) TyComp {
	return types
}

func (p TyComp) TypeIdent() TyComp {
	if p.Len() > 0 {
		return p.Pattern()[0]
	}
	return Def(None)
}

func (p TyComp) TypeReturn() TyComp {
	if p.Len() > 1 {
		return p.Pattern()[1]
	}
	return Def(None)
}

func (p TyComp) TypeArguments() TyComp {
	if p.Len() > 2 {
		return p.Pattern()[2]
	}
	return Def(None)
}

func (p TyComp) TypePropertys() []TyComp {
	if p.Len() > 2 {
		return p.Pattern()[2:]
	}
	return []TyComp{}
}

func (p TyComp) Match(typ d.Typed) bool {
	if Kind_Comp.Match(typ.Kind()) {
		return p.MatchTypes(typ.(TyComp).Types()...)
	}
	return p[0].Match(typ)
}

// match-args takes multiple expression arguments and matches their types
// against the elements of the pattern.
func (p TyComp) MatchArgs(args ...Expression) bool {
	var head, tail = p.ConsumeTyped()
	for _, arg := range args {
		if head == nil {
			break
		}
		if !head.Match(arg.Type()) {
			return false
		}
		head, tail = tail.ConsumeTyped()
	}
	return true
}

// match-types takes multiple types and matches them against an equal number of
// pattern elements one at a time, starting with the first one. if the number
// of arguments and elements differ, the shorter list will be evaluated.
func (p TyComp) MatchTypes(types ...d.Typed) bool {
	var short, long = p.sortLength(types...)
	for n, elem := range short {
		if !elem.Match(long[n]) {
			return false
		}
	}
	return true
}

// matches if any of the arguments matches any of the patterns elements
func (p TyComp) MatchAnyType(args ...d.Typed) bool {
	var head, tail = p.ConsumeTyped()
	for _, arg := range args {
		if head == nil {
			return false
		}
		if head.Match(arg) {
			return true
		}
		head, tail = tail.ConsumeTyped()
	}
	return false
}

// returns true if the arguments type matches any of the patterns types
func (p TyComp) MatchAnyArg(args ...Expression) bool {
	var types = make([]d.Typed, 0, len(args))
	for _, arg := range args {
		types = append(types, arg.Type())
	}
	return p.MatchAnyType(types...)
}

// matches multiple type flags against its elements in order. should there be
// more, or less arguments than pattern elements, the shorter sequence will be
// matched.
func (p TyComp) sortLength(types ...d.Typed) (short, long []d.Typed) {
	// if number of arguments is not equal to number of elements, find
	// shorter sequence
	if p.Len() > len(types) {
		short, long = types, p.Types()
	} else {
		short, long = p.Types(), types
	}
	return short, long
}

func (p TyComp) HeadTyped() d.Typed {
	if len(p) > 0 {
		return p[0]
	}
	return nil
}
func (p TyComp) TailTyped() TyComp {
	if len(p) > 1 {
		return p[1:]
	}
	return nil
}
func (p TyComp) ConsumeTyped() (d.Typed, TyComp) {
	return p.HeadTyped(), p.TailTyped()
}

// pattern yields a slice of type patterns, with all none & nil elements
// filtered out
func (p TyComp) Elements() []d.Typed {
	var elems = make([]d.Typed, 0, p.Count())
	for _, elem := range p {
		if Kind_Nat.Match(elem.Kind()) {
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
func (p TyComp) Fields() []TyComp {
	var elems = make([]TyComp, 0, p.Count())
	for _, elem := range p.Elements() {
		if Kind_Nat.Match(elem.Kind()) {
			if elem.Match(d.Nil) {
				continue
			}
		}
		if elem.Match(None) {
			continue
		}
		if Kind_Comp.Match(elem.Kind()) {
			elems = append(elems, elem.(TyComp))
			continue
		}
		elems = append(elems, Def(elem))
	}
	return elems
}

func (p TyComp) ArgumentsName() string {
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
func (p TyComp) IdentName() string {
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
func (p TyComp) ReturnName() string {
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

func (p TyComp) TypeName() string {
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

// type-elem yields the first elements typed
func (p TyComp) TypeElem() TyComp { return p.TypeIdent() }

// elems yields all elements contained in the pattern
func (p TyComp) Type() TyComp                  { return p }
func (p TyComp) Types() []d.Typed              { return p }
func (p TyComp) Call(...Expression) Expression { return p } // ← TODO: match arg instances
func (p TyComp) Len() int                      { return len(p.Types()) }
func (p TyComp) String() string                { return p.TypeName() }
func (p TyComp) Kind() d.Uint8Val              { return Kind_Comp.U() }
func (p TyComp) Flag() d.BitFlag               { return p.TypeFnc().Flag() }
func (p TyComp) TypeFnc() TyFnc                { return Type }

// length of elements excluding fields set to none
func (p TyComp) Count() int {
	var count int
	for _, elem := range p {
		if !elem.Match(None) {
			count = count + 1
		}
	}
	return count
}
func (p TyComp) Get(idx int) TyComp {
	if idx < p.Len() {
		return p.Pattern()[idx]
	}
	return Def(None)
}

// head yields the first pattern element cast as expression
func (p TyComp) Head() Expression {
	if p.Len() > 0 {
		var head = p.Pattern()[0]
		return head
	}
	return nil
}

// type-head yields first pattern element as typed
func (p TyComp) HeadPattern() TyComp { return p.Head().(TyComp) }

// tail yields a consumeable consisting all pattern elements but the first one
// cast as slice of expressions
func (p TyComp) Tail() Traversable {
	if p.Len() > 1 {
		return Def(p.Types()[1:]...)
	}
	return TyComp([]d.Typed{})
}

// tail-type yields a type pattern consisting of all pattern elements but the
// first one
func (p TyComp) TailPattern() TyComp {
	if p.Len() > 0 {
		return p.Types()[1:]
	}
	return []d.Typed{}
}

// consume uses head & tail to implement consumeable
func (p TyComp) Traverse() (Expression, Traversable) { return p.Head(), p.Tail() }

// pattern-consume works like type consume, but yields the head converted to,
// or cast as type pattern
func (p TyComp) ConsumePattern() (TyComp, TyComp) {
	return p.HeadPattern(), p.TailPattern()
}

func (p TyComp) Cons(args ...Expression) Sequential {
	var types = make([]Expression, 0, p.Len())
	for _, pat := range p {
		types = append(types, pat.(TyComp))
	}
	return NewVector(append(args, types...)...)
}

func (p TyComp) Append(args ...Expression) Sequential {
	var types = make([]Expression, 0, p.Len())
	for _, pat := range p {
		types = append(types, pat.(TyComp))
	}
	return NewVector(append(types, args...)...)
}

// pattern yields a slice of type patterns, with all none & nil elements
// filtered out
func (p TyComp) Pattern() []TyComp {
	var pattern = make([]TyComp, 0, p.Len())
	for _, typ := range p.Types() {
		if Kind_Comp.Match(typ.Kind()) {
			pattern = append(pattern, typ.(TyComp))
			continue
		}
		pattern = append(pattern, Def(typ))
	}
	return pattern
}

// print converts pattern to string, seperating the elements with a seperator
// and putting sub patterns in delimiters. seperator and delimiters are passed
// to the method. sub patterns are printed recursively.
func (p TyComp) Print(ldelim, sep, rdelim string) string {
	var names = make([]string, 0, p.Len())
	for _, typ := range p.Types() {
		// element is instance of data/typed → print type-name
		if !Kind_Comp.Match(typ.Kind()) {
			names = append(names, typ.TypeName())
			continue
		}
		// element is a type pattern
		var pat = typ.(TyComp)
		// print type pattern with delimiters and separator
		names = append(names, pat.Print(ldelim, sep, rdelim))
	}
	// print elements wrapped in delimiters, seperated by seperator
	return ldelim + strings.Join(names, sep) + rdelim
}

// bool methods
func (p TyComp) HasIdentity() bool {
	if p.Count() > 0 {
		return true
	}
	return false
}
func (p TyComp) HasReturnType() bool {
	if p.Count() > 1 {
		return true
	}
	return false
}
func (p TyComp) HasArguments() bool {
	if p.Count() > 2 {
		return true
	}
	return false
}

// one element pattern is a type identity
func (p TyComp) IsIdentity() bool {
	if p.Count() == 1 {
		return true
	}
	return false
}
func (p TyComp) IsAtomic() bool {
	if p.IsIdentity() {
		return !strings.ContainsAny(p.Elements()[0].TypeName(), " |,:")
	}
	return false
}
func (p TyComp) IsTruth() bool {
	if p.Count() == 1 {
		return p.Elements()[0].Match(Truth)
	}
	return false
}
func (p TyComp) IsTrinary() bool {
	if p.Count() == 1 {
		return p.Elements()[0].Match(Trinary)
	}
	return false
}
func (p TyComp) IsCompare() bool {
	if p.Count() == 1 {
		return p.Elements()[0].Match(Compare)
	}
	return false
}
func (p TyComp) IsData() bool {
	if p.Count() == 2 {
		return p.Elements()[0].Match(Data)
	}
	return false
}
func (p TyComp) IsPair() bool {
	if p.Count() == 2 {
		return p.Elements()[0].Match(Pair)
	}
	return false
}
func (p TyComp) IsVector() bool {
	if p.Count() == 2 {
		return p.Elements()[0].Match(Vector)
	}
	return false
}
func (p TyComp) IsList() bool {
	if p.Count() == 2 {
		return p.Elements()[0].Match(List)
	}
	return false
}
func (p TyComp) IsCollection() bool {
	if p.Count() == 2 {
		return p.Elements()[0].Match(Collections)
	}
	return false
}
func (p TyComp) IsEnum() bool {
	if p.Count() == 2 {
		return p.Elements()[0].Match(Enum)
	}
	return false
}
func (p TyComp) IsTuple() bool {
	if p.Count() == 2 {
		return p.Elements()[0].Match(Tuple)
	}
	return false
}
func (p TyComp) IsRecord() bool {
	if p.Count() == 2 {
		return p.Elements()[0].Match(Record)
	}
	return false
}
func (p TyComp) IsSet() bool {
	if p.Count() == 2 {
		return p.Elements()[0].Match(Set)
	}
	return false
}
func (p TyComp) IsSwitch() bool {
	if p.Count() == 2 {
		return p.Elements()[0].Match(Switch)
	}
	return false
}
func (p TyComp) IsNumber() bool {
	if p.Count() == 2 {
		return p.Elements()[0].Match(Number)
	}
	return false
}
func (p TyComp) IsString() bool {
	if p.Count() == 2 {
		return p.Elements()[0].Match(String)
	}
	return false
}
func (p TyComp) IsBytes() bool {
	if p.Count() == 2 {
		return p.Elements()[0].Match(Bytes)
	}
	return false
}
func (p TyComp) IsSumType() bool {
	if p.Count() == 2 && (p.IsList() || p.IsVector() || p.IsEnum()) {
		return true
	}
	return false
}
func (p TyComp) IsProductType() bool {
	if p.Count() == 2 && (p.IsTuple() || p.IsRecord() || p.IsPair() || p.IsSet()) {
		return true
	}
	return false
}
func (p TyComp) IsCase() bool {
	if p.Count() == 3 {
		return p.Elements()[1].Match(Case)
	}
	return false
}
func (p TyComp) IsMaybe() bool {
	if p.Count() == 3 {
		return p.Elements()[1].Match(Maybe)
	}
	return false
}
func (p TyComp) IsAlternative() bool {
	if p.Count() == 3 {
		return p.Elements()[1].Match(Variant)
	}
	return false
}
func (p TyComp) IsFunction() bool {
	if p.Count() == 3 {
		return p.Elements()[1].Match(Value)
	}
	return false
}
func (p TyComp) IsParametric() bool {
	if p.Count() == 3 {
		return p.Elements()[1].Match(Parametric)
	}
	return false
}

// two element pattern is a constant type returning a value type
func (p TyComp) HasData() bool        { return p.MatchAnyType(Data) }
func (p TyComp) HasPair() bool        { return p.MatchAnyType(Pair) }
func (p TyComp) HasEnum() bool        { return p.MatchAnyType(Enum) }
func (p TyComp) HasTuple() bool       { return p.MatchAnyType(Tuple) }
func (p TyComp) HasRecord() bool      { return p.MatchAnyType(Record) }
func (p TyComp) HasTruth() bool       { return p.MatchAnyType(Truth) }
func (p TyComp) HasTrinary() bool     { return p.MatchAnyType(Trinary) }
func (p TyComp) HasCompare() bool     { return p.MatchAnyType(Compare) }
func (p TyComp) HasBound() bool       { return p.MatchAnyType(Min, Max) }
func (p TyComp) HasMaybe() bool       { return p.MatchAnyType(Maybe) }
func (p TyComp) HasAlternative() bool { return p.MatchAnyType(Variant) }
func (p TyComp) HasNumber() bool      { return p.MatchAnyType(Number) }
func (p TyComp) HasString() bool      { return p.MatchAnyType(String) }
func (p TyComp) HasBytes() bool       { return p.MatchAnyType(Bytes) }
func (p TyComp) HasCollection() bool {
	return p.MatchAnyType(
		List, Vector, Tuple, Enum, Record)
}
func (p TyComp) HasReturn() bool {
	if p.Count() >= 2 {
		return true
	}
	return false
}
func (p TyComp) HasArgs() bool {
	if p.Count() >= 3 {
		return true
	}
	return false
}
