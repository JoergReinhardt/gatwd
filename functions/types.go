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
	TyExp func(...Expression) Expression
	TySym string
	TyAlt TyDef
	TyDef []d.Typed
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
	Kind_Opt

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
	Undecided
	/// ORDER
	Equal
	Lesser
	Greater
	/// BOUND
	Min
	Max
	/// OPTIONS
	Polymorph
	Switch
	Case
	Just
	None
	Either
	Or
	/// CLASSES
	Natural
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
	Functor
	Applicative
	Monad
	State
	IO

	//// TRUTH & COMPARE
	Truth   = True | False
	Trinary = Truth | Undecided
	Compare = Lesser | Greater | Equal

	//// BOUNDS
	Bound = Min | Max

	//// OPTIONALS
	Optionals    = Just | None
	Alternatives = Either | Or

	//// COLLECTIONS
	Collections = List | Vector | Enum | Set
	Products    = Pair | Tuple | Record | HashMap

	//// TOPOLOGYS (maps between categorys)
	Topologys = Functor | Applicative | Monad |
		State | IO | Group

	//// CONTINUA
	Continua = Collections | Products | Topologys

	//// ATOMIC
	Numbers = Natural | Integer | Real | Ratio
	Symbols = Letter | String | Byte | Truth |
		Trinary | Compare | Bound

	//// MANIFOLDS (PARAMETRIC &| POLYMORPHIC)
	Manifolds = Optionals | Alternatives |
		Topologys | Products

	// set of all TYPES
	T TyFnc = 0xFFFFFFFFFFFFFFFF
)

// WHEN GO GENERATED FILE IS WRITE PROTECTED AND NEEDS TO BE REGENERATED
// OUTCOMMENT THIS:
//func (t TyFnc) String() string { return "standin string func" }

//// MATCHER
func IsOf(typ d.Typed, arg Expression) bool { return arg.Type().Match(typ) }
func IsNone(arg Expression) bool            { return arg.Type().Match(None) }
func IsData(arg Expression) bool            { return arg.Type().Match(Data) }
func IsCons(arg Expression) bool            { return arg.Type().Match(Constant) }
func IsComp(arg Expression) bool            { return arg.Type().Match(Compare) }
func IsBound(arg Expression) bool           { return arg.Type().Match(Bound) }
func IsTruth(arg Expression) bool           { return arg.Type().Match(Truth) }
func IsTrinary(arg Expression) bool         { return arg.Type().Match(Trinary) }
func IsJust(arg Expression) bool            { return arg.Type().Match(Just) }
func IsEither(arg Expression) bool          { return arg.Type().Match(Either) }
func IsOr(arg Expression) bool              { return arg.Type().Match(Or) }
func IsVect(arg Expression) bool            { return arg.Type().Match(Vector) }
func IsList(arg Expression) bool            { return arg.Type().Match(List) }
func IsPair(arg Expression) bool            { return arg.Type().Match(Pair) }
func IsType(arg Expression) bool            { return arg.Type().Match(Type) }
func IsProdT(arg Expression) bool           { return arg.Type().Match(Products) }
func IsContin(arg Expression) bool          { return arg.Type().Match(Continua) }
func IsText(arg Expression) bool            { return arg.Type().Match(Symbols) }
func IsNumber(arg Expression) bool          { return arg.Type().Match(Numbers) }

// helper functions, to convert between slices of data/typed & ty-pattern
// instances
func typedToComp(typ []d.Typed) []TyDef {
	var pat = make([]TyDef, 0, len(typ))
	for _, t := range typ {
		if Kind_Comp.Match(t.Kind()) {
			pat = append(pat, t.(TyDef))
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
func compToTyped(pat []TyDef) []d.Typed {
	var typ = make([]d.Typed, 0, len(pat))
	for _, p := range pat {
		typ = append(typ, p)
	}
	return typ
}
func compToExpr(comps []TyDef) []Expression {
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
func (t TyFnc) Type() TyDef                        { return Def(t) }
func (t TyFnc) Match(arg d.Typed) bool             { return t.Flag().Match(arg) }
func (t TyFnc) TypeName() string {
	var count = t.Flag().Count()
	// loop to print concatenated type classes correcty
	if count > 1 {
		switch t {
		case T:
			return "*"
		case Truth:
			return "Truth"
		case Trinary:
			return "Trinary"
		case Compare:
			return "Compare"
		case Bound:
			return "Bound"
		case Optionals:
			return "Optional"
		case Collections:
			return "Collections"
		case Products:
			return "Products"
		case Topologys:
			return "Topologys"
		case Continua:
			return "Continua"
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

func (p TyProp) Flag() d.BitFlag                    { return d.BitFlag(uint64(p)) }
func (p TyProp) Kind() d.Uint8Val                   { return Kind_Prop.U() }
func (p TyProp) Type() TyDef                        { return Def(p) }
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
func DefSym(symbol string) TySym { return TySym(symbol) }
func (n TySym) Kind() d.Uint8Val { return Kind_Sym.U() }
func (n TySym) Flag() d.BitFlag  { return Symbol.Flag() }
func (n TySym) Type() TyDef      { return Def(n) }
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

//// SET OF ALTERNATIVE TYPES
///
// type flag representing pattern elements that define symbols
func DefAlt(types ...d.Typed) TyAlt { return TyAlt(Def(types...)) }
func (n TyAlt) TypeFnc() TyFnc      { return Or }
func (n TyAlt) Flag() d.BitFlag     { return Or.Flag() }
func (n TyAlt) Type() TyDef         { return TyDef(n) }
func (n TyAlt) Kind() d.Uint8Val    { return Kind_Opt.U() }
func (n TyAlt) String() string      { return n.TypeName() }
func (n TyAlt) TypeName() string {
	var str string // = "["
	for i, t := range n {
		str = str + t.TypeName()
		if i < len(n)-1 {
			str = str + "|"
		}
	}
	return str //+ "]"
}

// matches when any of its members matches the arguments type
func (n TyAlt) Match(arg d.Typed) bool {
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
func (n TyAlt) Call(args ...Expression) Expression {
	for _, arg := range args {
		if !n.Match(arg.Type()) {
			return Box(d.BoolVal(false))
		}
	}
	return Box(d.BoolVal(true))
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
func (n TyExp) Type() TyDef                        { return Def(n) }
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
		if n.Call(typ.(TyDef)).Type().Match(None) {
			return false
		}
	}
	if n.Call(Def(typ)).Type().Match(None) {
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
func Def(types ...d.Typed) TyDef {
	return types
}

func (p TyDef) TypeId() TyDef {
	if p.Len() > 0 {
		return p.Pattern()[0]
	}
	return Def(None)
}

func (p TyDef) TypeRet() TyDef {
	if p.Len() > 1 {
		return p.Pattern()[1]
	}
	return Def(None)
}

func (p TyDef) TypeArgs() TyDef {
	if p.Len() > 2 {
		return p.Pattern()[2]
	}
	return Def(None)
}

func (p TyDef) TypePropertys() []TyDef {
	if p.Len() > 2 {
		return p.Pattern()[2:]
	}
	return []TyDef{}
}

func (p TyDef) Match(typ d.Typed) bool {
	if Kind_Comp.Match(typ.Kind()) {
		return p.MatchTypes(typ.(TyDef).Types()...)
	}
	return p[0].Match(typ)
}

// match-args takes multiple expression arguments and matches their types
// against the elements of the pattern.
func (p TyDef) MatchArgs(args ...Expression) bool {
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
func (p TyDef) MatchTypes(types ...d.Typed) bool {
	var short, long = p.sortLength(types...)
	for n, elem := range short {
		if !elem.Match(long[n]) {
			return false
		}
	}
	return true
}

// matches if any of the arguments matches any of the patterns elements
func (p TyDef) MatchAnyType(args ...d.Typed) bool {
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
func (p TyDef) MatchAnyArg(args ...Expression) bool {
	var types = make([]d.Typed, 0, len(args))
	for _, arg := range args {
		types = append(types, arg.Type())
	}
	return p.MatchAnyType(types...)
}

// matches multiple type flags against its elements in order. should there be
// more, or less arguments than pattern elements, the shorter sequence will be
// matched.
func (p TyDef) sortLength(types ...d.Typed) (short, long []d.Typed) {
	// if number of arguments is not equal to number of elements, find
	// shorter sequence
	if p.Len() > len(types) {
		short, long = types, p.Types()
	} else {
		short, long = p.Types(), types
	}
	return short, long
}

func (p TyDef) HeadTyped() d.Typed {
	if len(p) > 0 {
		return p[0]
	}
	return nil
}
func (p TyDef) TailTyped() TyDef {
	if len(p) > 1 {
		return p[1:]
	}
	return nil
}
func (p TyDef) ConsumeTyped() (d.Typed, TyDef) {
	return p.HeadTyped(), p.TailTyped()
}

// elements returns the instnces of d.Typed initially passed to the
// constructor.
func (p TyDef) Elements() []d.Typed {
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
func (p TyDef) Fields() []TyDef {
	var elems = make([]TyDef, 0, p.Count())
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
			elems = append(elems, elem.(TyDef))
			continue
		}
		elems = append(elems, Def(elem))
	}
	return elems
}

// print returns a string representation of a pattern, seperating the elements
// with a seperator and putting sub patterns in delimiters. seperator and
// delimiters are passed to the method. sub patterns are printed recursively.
func (p TyDef) print(ldelim, sep, rdelim string) string {
	var names = make([]string, 0, p.Len())
	for _, typ := range p.Types() {
		names = append(names, typ.TypeName())
	}
	// print elements wrapped in delimiters, seperated by seperator
	return ldelim + strings.Join(names, sep) + rdelim
}

func (p TyDef) ArgumentsName() string {
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
func (p TyDef) IdentName() string {
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
func (p TyDef) ReturnName() string {
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

func (p TyDef) TypeName() string {
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
func (p TyDef) TypeElem() TyDef { return p.TypeId() }

// elems yields all elements contained in the pattern
func (p TyDef) Types() []d.Typed              { return p }
func (p TyDef) Call(...Expression) Expression { return p } // ← TODO: match arg instances
func (p TyDef) Len() int                      { return len(p.Types()) }
func (p TyDef) Empty() bool                   { return p.Len() == 0 }
func (p TyDef) String() string                { return p.TypeName() }
func (p TyDef) Kind() d.Uint8Val              { return Kind_Comp.U() }
func (p TyDef) Flag() d.BitFlag               { return p.TypeFnc().Flag() }
func (p TyDef) Type() TyDef                   { return p }
func (p TyDef) TypeFnc() TyFnc                { return Type }

// length of elements excluding fields set to none
func (p TyDef) Count() int {
	var count int
	for _, elem := range p {
		if !elem.Match(None) {
			count = count + 1
		}
	}
	return count
}
func (p TyDef) Get(idx int) TyDef {
	if idx < p.Len() {
		return p.Pattern()[idx]
	}
	return Def(None)
}

// head yields the first pattern element cast as expression
func (p TyDef) Head() Expression {
	if p.Len() > 0 {
		var head = p.Pattern()[0]
		return head
	}
	return nil
}

// type-head yields first pattern element as typed
func (p TyDef) HeadPattern() TyDef { return p.Head().(TyDef) }

// tail yields a consumeable consisting all pattern elements but the first one
// cast as slice of expressions
func (p TyDef) Tail() Grouped {
	if p.Len() > 1 {
		return Def(p.Types()[1:]...)
	}
	return TyDef([]d.Typed{})
}

// tail-type yields a type pattern consisting of all pattern elements but the
// first one
func (p TyDef) TailPattern() TyDef {
	if p.Len() > 0 {
		return p.Types()[1:]
	}
	return []d.Typed{}
}

// consume uses head & tail to implement consumeable
func (p TyDef) Continue() (Expression, Grouped) { return p.Head(), p.Tail() }

// pattern-consume works like type consume, but yields the head converted to,
// or cast as type pattern
func (p TyDef) ConsumePattern() (TyDef, TyDef) {
	return p.HeadPattern(), p.TailPattern()
}

func (p TyDef) ConsGroup(con Grouped) Grouped {
	var types = make([]d.Typed, 0, p.Len())
	for head, cons := con.Continue(); !cons.Empty(); {
		if Kind_Comp.Match(head.Type().Kind()) {
			types = append(types, head.(TyDef))
			continue
		}
		types = append(types, head.Type())
	}
	return Def(types...)
}
func (p TyDef) Cons(arg Expression) Grouped {
	if IsType(arg) {
		return Def(p, arg.(d.Typed))
	}
	return Def(p, arg.Type())
}

func (p TyDef) Concat(grp Continued) Grouped {
	var slice = make([]Expression, 0, len(p))
	for _, t := range p {
		slice = append(slice, t.(TyDef))
	}
	return NewList(slice...).Concat(grp)
}
func (p TyDef) Append(args ...Expression) Grouped {
	var types = make([]Expression, 0, p.Len())
	for _, pat := range p {
		types = append(types, pat.(TyDef))
	}
	return NewVector(append(types, args...)...)
}

// pattern yields a slice of type patterns, with all none & nil elements
// filtered out
func (p TyDef) Pattern() []TyDef {
	var pattern = make([]TyDef, 0, p.Len())
	for _, typ := range p.Types() {
		if Kind_Comp.Match(typ.Kind()) {
			pattern = append(pattern, typ.(TyDef))
			continue
		}
		pattern = append(pattern, Def(typ))
	}
	return pattern
}

// bool methods
func (p TyDef) HasIdentity() bool {
	if p.Count() > 0 {
		return true
	}
	return false
}
func (p TyDef) HasReturnType() bool {
	if p.Count() > 1 {
		return true
	}
	return false
}
func (p TyDef) HasArguments() bool {
	if p.Count() > 2 {
		return true
	}
	return false
}

// one element pattern is a type identity
func (p TyDef) IsIdentity() bool {
	if p.Count() == 1 {
		return true
	}
	return false
}
func (p TyDef) IsAtomic() bool {
	if p.IsIdentity() {
		return !strings.ContainsAny(p.Elements()[0].TypeName(), " |,:")
	}
	return false
}
func (p TyDef) IsTruth() bool {
	if p.Count() == 1 {
		return p.Elements()[0].Match(Truth)
	}
	return false
}
func (p TyDef) IsTrinary() bool {
	if p.Count() == 1 {
		return p.Elements()[0].Match(Trinary)
	}
	return false
}
func (p TyDef) IsCompare() bool {
	if p.Count() == 1 {
		return p.Elements()[0].Match(Compare)
	}
	return false
}
func (p TyDef) IsData() bool {
	if p.Count() == 2 {
		return p.Elements()[0].Match(Data)
	}
	return false
}
func (p TyDef) IsPair() bool {
	if p.Count() == 2 {
		return p.Elements()[0].Match(Pair)
	}
	return false
}
func (p TyDef) IsVector() bool {
	if p.Count() == 2 {
		return p.Elements()[0].Match(Vector)
	}
	return false
}
func (p TyDef) IsList() bool {
	if p.Count() == 2 {
		return p.Elements()[0].Match(List)
	}
	return false
}
func (p TyDef) IsFunctor() bool {
	if p.Count() == 2 {
		return p.Elements()[0].Match(Continua)
	}
	return false
}
func (p TyDef) IsEnum() bool {
	if p.Count() == 2 {
		return p.Elements()[0].Match(Enum)
	}
	return false
}
func (p TyDef) IsTuple() bool {
	if p.Count() == 2 {
		return p.Elements()[0].Match(Tuple)
	}
	return false
}
func (p TyDef) IsRecord() bool {
	if p.Count() == 2 {
		return p.Elements()[0].Match(Record)
	}
	return false
}
func (p TyDef) IsSet() bool {
	if p.Count() == 2 {
		return p.Elements()[0].Match(Set)
	}
	return false
}
func (p TyDef) IsSwitch() bool {
	if p.Count() == 2 {
		return p.Elements()[0].Match(Switch)
	}
	return false
}
func (p TyDef) IsNumber() bool {
	if p.Count() == 2 {
		return p.Elements()[0].Match(Numbers)
	}
	return false
}
func (p TyDef) IsString() bool {
	if p.Count() == 2 {
		return p.Elements()[0].Match(String)
	}
	return false
}
func (p TyDef) IsByte() bool {
	if p.Count() == 2 {
		return p.Elements()[0].Match(Byte)
	}
	return false
}
func (p TyDef) IsSumType() bool {
	if p.Count() == 2 && (p.IsList() || p.IsVector() || p.IsEnum()) {
		return true
	}
	return false
}
func (p TyDef) IsProductType() bool {
	if p.Count() == 2 && (p.IsTuple() || p.IsRecord() || p.IsPair() || p.IsSet()) {
		return true
	}
	return false
}
func (p TyDef) IsCase() bool {
	if p.Count() == 3 {
		return p.Elements()[1].Match(Case)
	}
	return false
}
func (p TyDef) IsMaybe() bool {
	if p.Count() == 3 {
		return p.Elements()[1].Match(Optionals)
	}
	return false
}
func (p TyDef) IsOption() bool {
	if p.Count() == 3 {
		return p.Elements()[1].Match(Alternatives)
	}
	return false
}
func (p TyDef) IsFunction() bool {
	if p.Count() == 3 {
		return p.Elements()[1].Match(Value)
	}
	return false
}
func (p TyDef) IsParametric() bool {
	if p.Count() == 3 {
		return p.Elements()[1].Match(Polymorph)
	}
	return false
}

// two element pattern is a constant type returning a value type
func (p TyDef) HasData() bool        { return p.MatchAnyType(Data) }
func (p TyDef) HasPair() bool        { return p.MatchAnyType(Pair) }
func (p TyDef) HasEnum() bool        { return p.MatchAnyType(Enum) }
func (p TyDef) HasTuple() bool       { return p.MatchAnyType(Tuple) }
func (p TyDef) HasRecord() bool      { return p.MatchAnyType(Record) }
func (p TyDef) HasTruth() bool       { return p.MatchAnyType(Truth) }
func (p TyDef) HasTrinary() bool     { return p.MatchAnyType(Trinary) }
func (p TyDef) HasCompare() bool     { return p.MatchAnyType(Compare) }
func (p TyDef) HasBound() bool       { return p.MatchAnyType(Min, Max) }
func (p TyDef) HasMaybe() bool       { return p.MatchAnyType(Optionals) }
func (p TyDef) HasAlternative() bool { return p.MatchAnyType(Alternatives) }
func (p TyDef) HasNumber() bool      { return p.MatchAnyType(Numbers) }
func (p TyDef) HasString() bool      { return p.MatchAnyType(String) }
func (p TyDef) HasByte() bool        { return p.MatchAnyType(Byte) }
func (p TyDef) HasCollection() bool {
	return p.MatchAnyType(
		List, Vector, Tuple, Enum, Record)
}
func (p TyDef) HasReturn() bool {
	if p.Count() >= 2 {
		return true
	}
	return false
}
func (p TyDef) HasArgs() bool {
	if p.Count() >= 3 {
		return true
	}
	return false
}
