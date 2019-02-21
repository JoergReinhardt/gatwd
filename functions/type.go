package functions

import (
	d "github.com/JoergReinhardt/gatwd/data"
)

// type TyFnc d.BitFlag
// encodes the kind of functional data as bitflag
type TyFnc d.UintVal

func (t TyFnc) Eval(...d.Native) d.Native { return t }
func (t TyFnc) TypeHO() TyFnc             { return t }
func (t TyFnc) TypeNat() d.TyNative       { return d.Flag }
func (t TyFnc) Flag() d.BitFlag           { return d.BitFlag(t) }
func (t TyFnc) Uint() uint                { return d.BitFlag(t).Uint() }

//go:generate stringer -type=TyFnc
const (
	Type TyFnc = 1 << iota
	Instance
	Data
	///////////
	Definition
	Application
	///////////
	Variable
	Function
	Closure
	Resourceful
	///////////
	Argument
	Parameter
	Accessor
	Attribut
	Predicate
	Generator
	Constructor
	Functor
	Monad
	///////////
	Condition
	False
	True
	Just
	None
	If
	Else
	Either
	Or
	Case
	///////////
	Pair
	List
	Tuple
	UniSet
	MuliSet
	AssocVec
	Record
	Vector
	DLink
	Link
	Node
	Tree
	///////////
	HigherOrder

	Truth = True | False

	Option = Just | None

	EitherOr = Either | Or

	IfElse = If | Else

	Chain = Vector | Tuple | Record

	AccIndex = Vector | Chain

	AccSymbol = Tuple | AssocVec | Record

	AccCollect = AccIndex | AccSymbol

	Nests = Tuple | List

	Sets = UniSet | MuliSet | AssocVec | Record

	Links = Link | DLink | Node | Tree // Consumeables
)

// type functions
//
// type functions are an intermediate manifestations of higher order types.

// TYPE IDENT
//
// type ident function binds the name of a unique type to one of it's possibly
// many definitions, each having a distict pattern/native-type/functional-type
// combination of features. the linker will combine all type related functions
// (data-, type constructors), carrying the same type name, to be the entire
// definition of that type.
type TypeIdentFnc func() (
	name, pattern string,
	tnat d.TyNative,
	tfnc TyFnc,
)

func (t TypeIdentFnc) Name() string                  { n, _, _, _ := t(); return n }
func (t TypeIdentFnc) Pattern() string               { _, p, _, _ := t(); return p }
func (t TypeIdentFnc) TypeNat() d.TyNative           { _, _, n, _ := t(); return d.Type | d.Function | n }
func (t TypeIdentFnc) TypeFnc() TyFnc                { _, _, _, f := t(); return Type | f }
func (t TypeIdentFnc) String() string                { return t.Name() + " = " + t.Pattern() }
func (t TypeIdentFnc) Eval(nat ...d.Native) d.Native { return t }
func (t TypeIdentFnc) Call(val ...Value) Value       { return t }

func NewTypeId(
	name, pattern string,
	tnat d.TyNative,
	tfnc TyFnc,
) TypeIdentFnc {
	return func() (string, string, d.TyNative, TyFnc) {
		return name, pattern, tnat, tfnc
	}
}

// PATTERN MATCHER
type PatternMatcherFnc func(tid ...Typed) bool

func (p PatternMatcherFnc) Match(t ...Typed) bool { return p(t...) }

func NewPatternMatcher(f func(t ...Typed) bool) PatternMatcherFnc {
	return PatternMatcherFnc(f)
}

// TYPE CONSTRUCTOR
//
// type constructor derives types whenever sum-, or product types are applyed
// to types for the first time.
type TypeConstructorFnc func(...Value) (
	name, pattern string,
	tfnc TyFnc,
	tnat d.TyNative,
	matcher PatternMatcherFnc,
	con func(...TypeIdent) TypeIdent,
)

func (d TypeConstructorFnc) Name() string                  { name, _, _, _, _, _ := d(); return name }
func (d TypeConstructorFnc) Pattern() string               { _, pat, _, _, _, _ := d(); return pat }
func (d TypeConstructorFnc) TypeFnc() TyFnc                { _, _, tfnc, _, _, _ := d(); return tfnc }
func (d TypeConstructorFnc) TypeNat() d.TyNative           { _, _, _, tnat, _, _ := d(); return tnat }
func (d TypeConstructorFnc) Match(it ...Typed) bool        { _, _, _, _, pm, _ := d(); return pm(it...) }
func (d TypeConstructorFnc) Con(t ...TypeIdent) TypeIdent  { _, _, _, _, _, con := d(); return con(t...) }
func (d TypeConstructorFnc) String() string                { return d.Name() }
func (d TypeConstructorFnc) Eval(nat ...d.Native) d.Native { return d.Con().Eval(nat...) }
func (d TypeConstructorFnc) Call(val ...Value) Value       { return d.Con().Call(val...) }

func NewTypeConstructor(
	name, pattern string,
	tfnc TyFnc,
	tnat d.TyNative,
	match PatternMatcherFnc,
	con func(...TypeIdent) TypeIdent,
) TypeConstructorFnc {
	return func(vals ...Value) (
		string, string,
		TyFnc, d.TyNative,
		PatternMatcherFnc,
		func(...TypeIdent) TypeIdent,
	) {
		return name, pattern, tfnc, tnat, match, con
	}
}

// DATA CONSTRUCTOR
//
// function to construct a data instance from input arguments. with inbuildt
// capacity to validate the type of the input against the type expected as
// argument by the constructor.
//
// a data constructor may also be satisfyed regarding it's arguments, in which
// case it returns the enclosed instance of the value, when called without
// arguments.
type DataConstructorFnc func(...Value) (
	name, pattern string,
	tfnc TyFnc,
	tnat d.TyNative,
	match PatternMatcherFnc,
	con Callable,
)

func (d DataConstructorFnc) Name() string           { name, _, _, _, _, _ := d(); return name }
func (d DataConstructorFnc) Pattern() string        { _, pat, _, _, _, _ := d(); return pat }
func (d DataConstructorFnc) TypeFnc() TyFnc         { _, _, tfnc, _, _, _ := d(); return tfnc }
func (d DataConstructorFnc) TypeNat() d.TyNative    { _, _, _, tnat, _, _ := d(); return tnat }
func (d DataConstructorFnc) Match(it ...Typed) bool { _, _, _, _, pm, _ := d(); return pm(it...) }
func (d DataConstructorFnc) Con(vals ...Value) Value {
	_, _, _, _, _, con := d()
	return con.Call(vals...)
}
func (d DataConstructorFnc) String() string                { return d.Name() }
func (d DataConstructorFnc) Eval(nat ...d.Native) d.Native { return d.Con().Eval(nat...) }
func (d DataConstructorFnc) Call(val ...Value) Value       { return d.Con().Call(val...) }

func NewDataConstructor(
	name, pattern string,
	pm PatternMatcherFnc,
	con Callable,
) DataConstructorFnc {
	return func(vals ...Value) (
		string, string,
		TyFnc, d.TyNative,
		PatternMatcherFnc,
		Callable,
	) {
		return name, pattern, con.TypeFnc(), con.TypeNat(), pm, con
	}
}
