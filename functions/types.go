package functions

import (
	"strings"
	s "strings"

	d "github.com/joergreinhardt/gatwd/data"
)

type (
	// FLAG TYPES
	//TyNat d.BitFlag ← defined in data
	TyKind d.Uint8Val // kind of type flag
	TyFnc  d.BitFlag  // functional primitive type
	TyProp d.Int8Val  // call property types

	// TYPE TAGS
	TyExp func(...Functor) Functor
	TySym string
	TyAll Decl
	TyAny Decl
	Decl  []d.Typed
)

//go:generate stringer -type TyKind
const (
	Kind_Flag TyKind = 0 + iota
	Kind_Nat
	Kind_Fnc
	Kind_Key
	Kind_Symb
	Kind_Expr
	Kind_Prop
	Kind_Lex
	Kind_Opt

	Kind_Decl TyKind = 255
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
	None
	Data
	Just
	Value
	Partial
	Constant
	Generator
	Accumulator
	Constructor
	/// PARAMETER
	Parameter
	Property
	Lexical
	Symbol
	Index
	Key
	/// TRUTH
	True
	False
	/// ORDER
	Equal
	Lesser
	Greater
	/// BOUND
	Min
	Max
	/// OPTIONS
	Polymorph
	Choice
	Either
	Or
	/// CLASSES
	Natural
	Boolean
	Integer
	Real
	Ratio
	Letter
	String
	Byte
	/// SUM
	Vector
	List
	Enum
	Set
	/// PRODUCT
	Pair
	Tuple
	Record
	HashMap
	/// COMPOUND
	Group
	Monad
	State
	IO

	//// ATOMIC
	Numbers = Natural | Integer | Real | Ratio
	Symbols = Letter | String | Byte | Truth |
		Comparison | Bound

	//// BOUNDS
	///  argument values are restricted to bounds
	Bound = Min | Max

	//// TRUTH & COMPARE
	Truth      = True | False
	Comparison = Lesser | Greater | Equal

	//// PREDICATES
	///  are true|false|(equal|undecided)
	Predicate = Truth | Comparison

	//// OPTIONALS
	///  return values of a certain type may, or may not be returned
	//   a value of alternative type, or nothing is returned instead
	Option      = Just | None
	Alternative = Either | Or
	Selection   = Option | Alternative | Choice

	//// ADDITIVE COLLECTIONS
	///  contain zero or more elements of common type
	Additives = List | Vector | Enum | Set | HashMap

	//// MULTIPLICATIVE COLLECTIONS
	///  contain zero or more elements of instances of optional, or
	//   alternative types
	Products = Pair | Tuple | Record

	//// CONTINUA
	///  return a continuation for each call until depletion
	Continua = Additives | Products

	//// TOPOLOGYS (maps between categorys)
	///  continua that enclose side effects like state, mutability, io‥.
	//   in order to be evaluated in return value continuation creation
	Topologys = Continua |
		Monad | State | IO | Group

	//// MANIFOLDS (PARAMETRIC &| POLYMORPHIC)
	Manifolds = Option | Alternative |
		Topologys | Products

	// set of all TYPES
	T TyFnc = 0xFFFFFFFFFFFFFFFF
)

// WHEN GO GENERATED FILE IS WRITE PROTECTED AND NEEDS TO BE REGENERATED
// OUTCOMMENT THIS:
//func (t TyFnc) String() string { return "standin string func" }

//// MATCHER
func IsOf(typ d.Typed, arg Functor) bool { return arg.Type().Match(typ) }

//
func IsNone(arg Functor) bool    { return arg.Type().Match(None) }
func IsAtom(arg Functor) bool    { return arg.TypeFnc().Match(Atomic) }
func IsPartial(arg Functor) bool { return arg.TypeFnc().Match(Partial) }
func IsData(arg Functor) bool    { return arg.TypeFnc().Match(Data) }
func IsCons(arg Functor) bool    { return arg.TypeFnc().Match(Constant) }
func IsComp(arg Functor) bool    { return arg.TypeFnc().Match(Comparison) }
func IsVect(arg Functor) bool    { return arg.TypeFnc().Match(Vector) }
func IsList(arg Functor) bool    { return arg.TypeFnc().Match(List) }
func IsPair(arg Functor) bool    { return arg.TypeFnc().Match(Pair) }

//
func IsBound(arg Functor) bool  { return arg.Type().Match(Bound) }
func IsTruth(arg Functor) bool  { return arg.Type().Match(Truth) }
func IsEither(arg Functor) bool { return arg.Type().Match(Either) }
func IsOr(arg Functor) bool     { return arg.Type().Match(Or) }
func IsType(arg Functor) bool   { return arg.Type().Match(Type) }
func IsProdT(arg Functor) bool  { return arg.Type().Match(Products) }
func IsContin(arg Functor) bool { return arg.Type().Match(Continua) }
func IsText(arg Functor) bool   { return arg.Type().Match(Symbols) }
func IsNumber(arg Functor) bool { return arg.Type().Match(Numbers) }

// helper functions, to convert between slices of data/typed & ty-pattern
// instances
func typedToComp(typ []d.Typed) []Decl {
	var pat = make([]Decl, 0, len(typ))
	for _, t := range typ {
		if Kind_Decl.Match(t.Kind()) {
			pat = append(pat, t.(Decl))
			continue
		}
		pat = append(pat, Declare(t))
	}
	return pat
}
func typedToExpr(typ []d.Typed) []Functor {
	var elems = make([]Functor, 0, len(typ))
	for _, comp := range typedToComp(typ) {
		elems = append(elems, comp)
	}
	return elems
}
func compToTyped(pat []Decl) []d.Typed {
	var typ = make([]d.Typed, 0, len(pat))
	for _, p := range pat {
		typ = append(typ, p)
	}
	return typ
}
func compToExpr(comps []Decl) []Functor {
	var elems = make([]Functor, 0, len(comps))
	for _, comp := range comps {
		elems = append(elems, comp)
	}
	return elems
}

// functional type flag expresses the type of a functional value
func (t TyFnc) TypeFnc() TyFnc               { return t }
func (t TyFnc) TypeNat() d.TyNat             { return d.Type }
func (t TyFnc) Flag() d.BitFlag              { return d.BitFlag(t) }
func (t TyFnc) Uint() d.UintVal              { return d.BitFlag(t).Uint() }
func (t TyFnc) Kind() d.Uint8Val             { return Kind_Fnc.U() }
func (t TyFnc) Call(args ...Functor) Functor { return t.TypeFnc() }
func (t TyFnc) Type() Decl                   { return Declare(t) }
func (t TyFnc) Match(arg d.Typed) bool       { return t.Flag().Match(arg) }
func (t TyFnc) TypeName() string {
	var count = t.Flag().Count()
	// loop to print concatenated type classes correcty
	if count > 1 {
		switch t {
		case T:
			return "✱"
		case Predicate:
			return "Praedicates"
		case Bound:
			return "Bound"
		case Alternative:
			return "Alternatives"
		case Additives:
			return "Additives"
		case Products:
			return "Products"
		case Continua:
			return "Continua"
		case Topologys:
			return "Topologys"
		case Manifolds:
			return "Manifolds"
		case Numbers:
			return "Numbers"
		case Symbols:
			return "Symbols"
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

func (p TyProp) Flag() d.BitFlag              { return d.BitFlag(uint64(p)) }
func (p TyProp) Kind() d.Uint8Val             { return Kind_Prop.U() }
func (p TyProp) Type() Decl                   { return Declare(p) }
func (p TyProp) TypeFnc() TyFnc               { return Property }
func (p TyProp) TypeNat() d.TyNat             { return d.Type }
func (p TyProp) TypeName() string             { return "Propertys" }
func (p TyProp) Match(flag d.Typed) bool      { return p.Flag().Match(flag) }
func (p TyProp) Call(args ...Functor) Functor { return p }

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
func DecSym(symbol string) TySym { return TySym(symbol) }
func (n TySym) Kind() d.Uint8Val { return Kind_Symb.U() }
func (n TySym) Flag() d.BitFlag  { return Symbol.Flag() }
func (n TySym) Type() Decl       { return Declare(n) }
func (n TySym) TypeFnc() TyFnc   { return Symbol }
func (n TySym) String() string   { return n.TypeName() }
func (n TySym) TypeName() string {
	if strings.Contains(string(n), " ") {
		return "(" + string(n) + ")"
	}
	return string(n)
}
func (n TySym) Call(args ...Functor) Functor {
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

//// SET OF TYPES OF A SCALAR TYPE (TUPLE, RECORD‥.)
///
// type flag representing pattern elements that define symbols
func DecAll(types ...d.Typed) TyAll { return TyAll(Declare(types...)) }
func (n TyAll) TypeFnc() TyFnc      { return Or }
func (n TyAll) Flag() d.BitFlag     { return Or.Flag() }
func (n TyAll) Type() Decl          { return Decl(n) }
func (n TyAll) Elements() []d.Typed { return Decl(n).Elements() }
func (n TyAll) Len() int            { return len(Decl(n).Elements()) }
func (n TyAll) Kind() d.Uint8Val    { return Kind_Opt.U() }
func (n TyAll) String() string      { return n.TypeName() }
func (n TyAll) TypeName() string {
	var str string = "("
	for i, t := range n {
		str = str + t.TypeName()
		if i < len(n)-1 {
			str = str + " "
		}
	}
	return str + "("
}

// matches when any of its members matches the arguments type
func (n TyAll) Match(arg d.Typed) bool {
	for _, typ := range n {
		if typ.Match(arg) {
			return true
		}
	}
	return false
}

// call method lifts arguments types and applys them to match method one by
// one.  returns true, if all passed arguements are in the set of optional
// types.
func (n TyAll) Call(args ...Functor) Functor {
	for _, arg := range args {
		if !n.Match(arg.Type()) {
			return Box(d.BoolVal(false))
		}
	}
	return Box(d.BoolVal(true))
}

//// SET OF mutual exclusive OPTIONS
///
// type flag representing pattern elements that define symbols
func DecAny(types ...d.Typed) TyAny { return TyAny(Declare(types...)) }
func (n TyAny) TypeFnc() TyFnc      { return Or }
func (n TyAny) Flag() d.BitFlag     { return Or.Flag() }
func (n TyAny) Type() Decl          { return Decl(n) }
func (n TyAny) Elements() []d.Typed { return Decl(n).Elements() }
func (n TyAny) Len() int            { return len(Decl(n).Elements()) }
func (n TyAny) Kind() d.Uint8Val    { return Kind_Opt.U() }
func (n TyAny) String() string      { return n.TypeName() }
func (n TyAny) TypeName() string {
	var str string = "["
	for i, t := range n {
		str = str + t.TypeName()
		if i < len(n)-1 {
			str = str + "|"
		}
	}
	return str + "]"
}

// matches when any of its members matches the arguments type
func (n TyAny) Match(arg d.Typed) bool {
	for _, typ := range n {
		if typ.Match(arg) {
			return true
		}
	}
	return false
}

// call method lifts arguments types and applys them to match method one by
// one.  returns true, if any of the passed arguements is in the set of
// optional types.
func (n TyAny) Call(args ...Functor) Functor {
	for _, arg := range args {
		if n.Match(arg.Type()) {
			return Box(d.BoolVal(true))
		}
	}
	return Box(d.BoolVal(false))
}

//// TYPE EXPRESSION
///
// type flag representing a parametric element in a type pattern by a value, or
// type-expression, expecting and returning typeflags as its values.
func DecGoVal(val interface{}) TyExp {
	return TyExp(func(...Functor) Functor { return Dat(val) })
}
func DecNatVal(nat d.Native) TyExp {
	return TyExp(func(...Functor) Functor { return Box(nat) })
}
func DecExpr(expr Functor) TyExp {
	return func(args ...Functor) Functor {
		if len(args) > 0 {
			return expr.Call(args...)
		}
		return expr.Call()
	}
}
func (n TyExp) Kind() d.Uint8Val             { return Kind_Expr.U() }
func (n TyExp) Flag() d.BitFlag              { return Value.Flag() }
func (n TyExp) Type() Decl                   { return Declare(n) }
func (n TyExp) TypeFnc() TyFnc               { return Value }
func (n TyExp) String() string               { return n().String() }
func (n TyExp) TypeName() string             { return n().Type().TypeName() }
func (n TyExp) Value() Functor               { return n() }
func (n TyExp) Call(args ...Functor) Functor { return n(args...) }

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
	if Kind_Decl.Match(typ.Kind()) {
		if n.Call(typ.(Decl)).Type().Match(None) {
			return false
		}
	}
	if n.Call(Declare(typ)).Type().Match(None) {
		return false
	}
	return true
}

///////////////////////////////////////////////////////////////////////////////
//// TYPE DEFINITION
///
// defines a new type according to a slice of possibly nested data/typed
// instances.  arguments are expected to be passed in troublesome irish order:
// -  I type *identity*
// -  R *return* value type
// -  A *argument* type set (!optional!)
func Declare(types ...d.Typed) Decl {
	return types
}

func (p Decl) TypeId() Decl {
	if p.Len() > 0 {
		return p.Pattern()[0]
	}
	return Declare(None)
}

func (p Decl) TypeRet() Decl {
	if p.Len() > 1 {
		return p.Pattern()[1]
	}
	return Declare(None)
}

func (p Decl) TypeArgs() Decl {
	if p.Len() > 2 {
		return p.Pattern()[2]
	}
	return Declare(None)
}

func (p Decl) TypePropertys() []Decl {
	if p.Len() > 2 {
		return p.Pattern()[2:]
	}
	return []Decl{}
}

func (p Decl) Match(typ d.Typed) bool {
	if Kind_Decl.Match(typ.Kind()) {
		return p.MatchTypes(typ.(Decl).Types()...)
	}
	return p[0].Match(typ)
}

// match-args takes multiple expression arguments and matches their types
// against the elements of the pattern.
func (p Decl) MatchArgs(args ...Functor) bool {
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
func (p Decl) MatchTypes(types ...d.Typed) bool {
	var short, long = p.sortLength(types...)
	for n, elem := range short {
		if !elem.Match(long[n]) {
			return false
		}
	}
	return true
}

// matches if any of the arguments matches any of the patterns elements
func (p Decl) MatchAnyType(args ...d.Typed) bool {
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
func (p Decl) MatchAnyArg(args ...Functor) bool {
	var types = make([]d.Typed, 0, len(args))
	for _, arg := range args {
		types = append(types, arg.Type())
	}
	return p.MatchAnyType(types...)
}

// matches multiple type flags against its elements in order. should there be
// more, or less arguments than pattern elements, the shorter sequence will be
// matched.
func (p Decl) sortLength(types ...d.Typed) (short, long []d.Typed) {
	// if number of arguments is not equal to number of elements, find
	// shorter sequence
	if p.Len() > len(types) {
		short, long = types, p.Types()
	} else {
		short, long = p.Types(), types
	}
	return short, long
}

func (p Decl) HeadTyped() d.Typed {
	if len(p) > 0 {
		return p[0]
	}
	return nil
}
func (p Decl) TailTyped() Decl {
	if len(p) > 1 {
		return p[1:]
	}
	return nil
}
func (p Decl) ConsumeTyped() (d.Typed, Decl) {
	return p.HeadTyped(), p.TailTyped()
}

// elements returns the instnces of d.Typed initially passed to the
// constructor.
func (p Decl) Elements() []d.Typed {
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

// fields returns each element as instance of composed-type, either casting the
// element as such, or instanciating one from the d.Typed interface isntance
func (p Decl) Fields() []Decl {
	var elems = make([]Decl, 0, p.Count())
	for _, elem := range p.Elements() {
		if Kind_Nat.Match(elem.Kind()) {
			if elem.Match(d.Nil) {
				continue
			}
		}
		if elem.Match(None) {
			continue
		}
		if Kind_Decl.Match(elem.Kind()) {
			elems = append(elems, elem.(Decl))
			continue
		}
		elems = append(elems, Declare(elem))
	}
	return elems
}

// print returns a string representation of a pattern, seperating the elements
// with a seperator and putting sub patterns in delimiters. seperator and
// delimiters are passed to the method. sub patterns are printed recursively.
func (p Decl) print(ldelim, sep, rdelim string) string {
	var names = make([]string, 0, p.Len())
	for _, typ := range p.Types() {
		names = append(names, typ.TypeName())
	}
	// print elements wrapped in delimiters, seperated by seperator
	return ldelim + strings.Join(names, sep) + rdelim
}

func (p Decl) ArgumentsName() string {
	if p.TypeArgs().Len() > 0 {
		if !p.TypeArgs().Match(None) {
			if p.TypeArgs().Len() > 1 {
				var ldelim, sep, rdelim = "", " → ", ""
				return p.TypeArgs().print(
					ldelim, sep, rdelim,
				)
			}
			return p.TypeArgs().print("", " ", "")
		}
	}
	return ""
}
func (p Decl) IdentName() string {
	if p.TypeId().Len() > 0 {
		if !p.TypeId().Match(None) {
			if p.TypeId().Len() > 1 {
				var ldelim, sep, rdelim = "(", " ", ")"
				switch {
				case p.TypeId().Match(List | Vector):
					ldelim, sep, rdelim = "[", " ", "]"
				case p.TypeId().Match(Set):
					ldelim, sep, rdelim = "{", " ", "}"
				case p.TypeId().Match(Record):
					ldelim, sep, rdelim = "{", " ∷ ", "}"
				case p.TypeId().Match(Tuple):
					ldelim, sep, rdelim = "(", " | ", ")"
				case p.TypeId().Match(Enum):
					ldelim, sep, rdelim = "[", " | ", "]"
				}
				return p.TypeId().print(
					ldelim, sep, rdelim,
				)
			}
			return p.TypeId().print("", " ", "")
		}
	}
	return ""
}
func (p Decl) ReturnName() string {
	if p.TypeRet().Len() > 0 {
		if !p.TypeRet().Match(None) {
			if p.TypeRet().Len() > 1 {
				var ldelim, sep, rdelim = "(", " ", ")"
				switch {
				case p.TypeId().Match(List | Vector):
					ldelim, sep, rdelim = "[", " ", "]"
				case p.TypeId().Match(Set):
					ldelim, sep, rdelim = "{", " ", "}"
				case p.TypeId().Match(Record):
					ldelim, sep, rdelim = "{", " ∷ ", "}"
				case p.TypeId().Match(Tuple):
					ldelim, sep, rdelim = "(", " | ", ")"
				case p.TypeId().Match(Enum):
					ldelim, sep, rdelim = "[", " | ", "]"
				}
				return p.TypeRet().print(
					ldelim, sep, rdelim,
				)
			}
			return p.TypeRet().print("", " ", "")
		}
	}
	return ""
}

func (p Decl) TypeName() string {
	var strs = []string{}
	if !p.TypeArgs().Match(None) {
		strs = append(strs, p.ArgumentsName())
	}
	if !p.TypeId().Match(None) {
		strs = append(strs, p.IdentName())
	}
	if !p.TypeRet().Match(None) {
		strs = append(strs, p.ReturnName())
	}
	return strings.Join(strs, " → ")
}

// type-elem yields the first elements typed
func (p Decl) TypeElem() Decl { return p.TypeId() }

// elems yields all elements contained in the pattern
func (p Decl) Types() []d.Typed        { return p }
func (p Decl) Call(...Functor) Functor { return p } // ← TODO: match arg instances
func (p Decl) Len() int                { return len(p.Types()) }
func (p Decl) Empty() bool             { return p.Len() == 0 }
func (p Decl) String() string          { return p.TypeName() }
func (p Decl) Kind() d.Uint8Val        { return Kind_Decl.U() }
func (p Decl) Flag() d.BitFlag         { return p.TypeFnc().Flag() }
func (p Decl) Type() Decl              { return p }
func (p Decl) TypeFnc() TyFnc          { return Type }

// length of elements excluding fields set to none
func (p Decl) Count() int {
	var count int
	for _, elem := range p {
		if !elem.Match(None) {
			count = count + 1
		}
	}
	return count
}
func (p Decl) Get(idx int) Decl {
	if idx < p.Len() {
		return p.Pattern()[idx]
	}
	return Declare(None)
}

// head yields the first pattern element cast as expression
func (p Decl) Head() Functor {
	if p.Len() > 0 {
		var head = p.Pattern()[0]
		return head
	}
	return nil
}

// type-head yields first pattern element as typed
func (p Decl) HeadPattern() Decl { return p.Head().(Decl) }

// tail yields a consumeable consisting all pattern elements but the first one
// cast as slice of expressions
func (p Decl) Tail() Applicative {
	if p.Len() > 1 {
		return Declare(p.Types()[1:]...)
	}
	return Decl([]d.Typed{})
}

// tail-type yields a type pattern consisting of all pattern elements but the
// first one
func (p Decl) TailPattern() Decl {
	if p.Len() > 0 {
		return p.Types()[1:]
	}
	return []d.Typed{}
}

// consume uses head & tail to implement consumeable
func (p Decl) Continue() (Functor, Applicative) { return p.Head(), p.Tail() }

// pattern-consume works like type consume, but yields the head converted to,
// or cast as type pattern
func (p Decl) ConsumePattern() (Decl, Decl) {
	return p.HeadPattern(), p.TailPattern()
}

func (p Decl) ConsGroup(con Applicative) Applicative {
	var types = make([]d.Typed, 0, p.Len())
	for head, cons := con.Continue(); !cons.Empty(); {
		if Kind_Decl.Match(head.Type().Kind()) {
			types = append(types, head.(Decl))
			continue
		}
		types = append(types, head.Type())
	}
	return Declare(types...)
}
func (p Decl) Cons(arg Functor) Applicative {
	if IsType(arg) {
		return Declare(p, arg.(d.Typed))
	}
	return Declare(p, arg.Type())
}

func (p Decl) Concat(grp Sequential) Applicative {
	var slice = make([]Functor, 0, len(p))
	for _, t := range p {
		slice = append(slice, t.(Decl))
	}
	return NewList(slice...).Concat(grp)
}
func (p Decl) Append(args ...Functor) Applicative {
	var types = make([]Functor, 0, p.Len())
	for _, pat := range p {
		types = append(types, pat.(Decl))
	}
	return NewVector(append(types, args...)...)
}

// pattern yields a slice of type patterns, with all none & nil elements
// filtered out
func (p Decl) Pattern() []Decl {
	var pattern = make([]Decl, 0, p.Len())
	for _, typ := range p.Types() {
		if Kind_Decl.Match(typ.Kind()) {
			pattern = append(pattern, typ.(Decl))
			continue
		}
		pattern = append(pattern, Declare(typ))
	}
	return pattern
}

// bool methods
func (p Decl) HasIdentity() bool {
	if p.Count() > 0 {
		return true
	}
	return false
}
func (p Decl) HasReturnType() bool {
	if p.Count() > 1 {
		return true
	}
	return false
}
func (p Decl) HasArguments() bool {
	if p.Count() > 2 {
		return true
	}
	return false
}

// one element pattern is a type identity
func (p Decl) IsIdentity() bool {
	if p.Count() == 1 {
		return true
	}
	return false
}
func (p Decl) IsAtomic() bool {
	if p.IsIdentity() {
		return !strings.ContainsAny(p.Elements()[0].TypeName(), " |,:")
	}
	return false
}
func (p Decl) IsTruth() bool {
	if p.Count() == 1 {
		return p.Elements()[0].Match(Truth)
	}
	return false
}
func (p Decl) IsCompare() bool {
	if p.Count() == 1 {
		return p.Elements()[0].Match(Comparison)
	}
	return false
}
func (p Decl) IsData() bool {
	if p.Count() == 2 {
		return p.Elements()[0].Match(Data)
	}
	return false
}
func (p Decl) IsPair() bool {
	if p.Count() == 2 {
		return p.Elements()[0].Match(Pair)
	}
	return false
}
func (p Decl) IsVector() bool {
	if p.Count() == 2 {
		return p.Elements()[0].Match(Vector)
	}
	return false
}
func (p Decl) IsList() bool {
	if p.Count() == 2 {
		return p.Elements()[0].Match(List)
	}
	return false
}
func (p Decl) IsFunctor() bool {
	if p.Count() == 2 {
		return p.Elements()[0].Match(Continua)
	}
	return false
}
func (p Decl) IsEnum() bool {
	if p.Count() == 2 {
		return p.Elements()[0].Match(Enum)
	}
	return false
}
func (p Decl) IsTuple() bool {
	if p.Count() == 2 {
		return p.Elements()[0].Match(Tuple)
	}
	return false
}
func (p Decl) IsRecord() bool {
	if p.Count() == 2 {
		return p.Elements()[0].Match(Record)
	}
	return false
}
func (p Decl) IsSet() bool {
	if p.Count() == 2 {
		return p.Elements()[0].Match(Set)
	}
	return false
}
func (p Decl) IsSwitch() bool {
	if p.Count() == 2 {
		return p.Elements()[0].Match(Choice)
	}
	return false
}
func (p Decl) IsNumber() bool {
	if p.Count() == 2 {
		return p.Elements()[0].Match(Numbers)
	}
	return false
}
func (p Decl) IsString() bool {
	if p.Count() == 2 {
		return p.Elements()[0].Match(String)
	}
	return false
}
func (p Decl) IsByte() bool {
	if p.Count() == 2 {
		return p.Elements()[0].Match(Byte)
	}
	return false
}
func (p Decl) IsSumType() bool {
	if p.Count() == 2 && (p.IsList() || p.IsVector() || p.IsEnum()) {
		return true
	}
	return false
}
func (p Decl) IsProductType() bool {
	if p.Count() == 2 && (p.IsTuple() || p.IsRecord() || p.IsPair() || p.IsSet()) {
		return true
	}
	return false
}
func (p Decl) IsCase() bool {
	if p.Count() == 3 {
		return p.Elements()[1].Match(Option)
	}
	return false
}
func (p Decl) IsMaybe() bool {
	if p.Count() == 3 {
		return p.Elements()[1].Match(Option)
	}
	return false
}
func (p Decl) IsOption() bool {
	if p.Count() == 3 {
		return p.Elements()[1].Match(Alternative)
	}
	return false
}
func (p Decl) IsFunction() bool {
	if p.Count() == 3 {
		return p.Elements()[1].Match(Value)
	}
	return false
}
func (p Decl) IsParametric() bool {
	if p.Count() == 3 {
		return p.Elements()[1].Match(Polymorph)
	}
	return false
}

// two element pattern is a constant type returning a value type
func (p Decl) HasData() bool        { return p.MatchAnyType(Data) }
func (p Decl) HasPair() bool        { return p.MatchAnyType(Pair) }
func (p Decl) HasEnum() bool        { return p.MatchAnyType(Enum) }
func (p Decl) HasTuple() bool       { return p.MatchAnyType(Tuple) }
func (p Decl) HasRecord() bool      { return p.MatchAnyType(Record) }
func (p Decl) HasTruth() bool       { return p.MatchAnyType(Truth) }
func (p Decl) HasCompare() bool     { return p.MatchAnyType(Comparison) }
func (p Decl) HasBound() bool       { return p.MatchAnyType(Min, Max) }
func (p Decl) HasMaybe() bool       { return p.MatchAnyType(Option) }
func (p Decl) HasAlternative() bool { return p.MatchAnyType(Alternative) }
func (p Decl) HasNumber() bool      { return p.MatchAnyType(Numbers) }
func (p Decl) HasString() bool      { return p.MatchAnyType(String) }
func (p Decl) HasByte() bool        { return p.MatchAnyType(Byte) }
func (p Decl) HasCollection() bool {
	return p.MatchAnyType(
		List, Vector, Tuple, Enum, Record)
}
func (p Decl) HasReturn() bool {
	if p.Count() >= 2 {
		return true
	}
	return false
}
func (p Decl) HasArgs() bool {
	if p.Count() >= 3 {
		return true
	}
	return false
}
