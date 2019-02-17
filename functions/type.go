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
	Data
	///////////
	Definition
	Application
	///////////
	Variable
	Function
	Closure
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
	Either
	False
	True
	Just
	None
	If
	Else
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

	EitherOr = If | Else

	Chain = Vector | Tuple | Record

	AccIndex = Vector | Chain

	AccSymbol = Tuple | AssocVec | Record

	AccCollect = AccIndex | AccSymbol

	Nests = Tuple | List

	Sets = UniSet | MuliSet | AssocVec | Record

	Links = Link | DLink | Node | Tree // Consumeables
)

type (
	// TYPE IDENTITY & CONSTRUCTION
	TypeFnc func() (
		name d.StrVal,
		signature d.StrVal,
		constructors []TypeDef,
	)
	SumTypeFnc  func() (TypeDef, []TypeDef)
	ProdTypeFnc func(t ...TypeDef) (ProdTypeFnc, []ProdTypeFnc)
)

func NewTypeFnc(
	name string,
	signature string,
	cons ...TypeDef,
) TypeFnc {
	return TypeFnc(func() (
		name, sig d.StrVal,
		cons []TypeDef,
	) {
		return d.StrVal(name), d.StrVal(signature), cons
	})
}
func (t TypeFnc) Name() d.StrVal      { name, _, _ := t(); return name }
func (t TypeFnc) Signature() d.StrVal { _, sig, _ := t(); return sig }
func (t TypeFnc) Cons() []TypeDef     { _, _, cons := t(); return cons }
func (t TypeFnc) TypeNat() d.TyNative { return d.Type }
func (t TypeFnc) TypeFnc() TyFnc      { return Type }
func (t TypeFnc) Ident() Value        { return t }

func (t TypeFnc) Call(...Value) Value { return t }
func (t TypeFnc) String() string      { return t.Name().String() }
func (t TypeFnc) Eval(p ...d.Native) d.Native {
	return t.Name()
}

// SUM TYPE
func NewSumType(sum func() (TypeDef, []TypeDef)) SumTypeFnc { return SumTypeFnc(sum) }
func (e SumTypeFnc) Type() TypeFnc                          { t, _ := e(); return t.(TypeFnc) }
func (e SumTypeFnc) Sum() []TypeDef                         { _, sum := e(); return sum }
func (e SumTypeFnc) Name() d.StrVal                         { return e.Type().Name() }
func (e SumTypeFnc) Signature() d.StrVal                    { return e.Type().Signature() }
func (e SumTypeFnc) TypeNat() d.TyNative                    { return d.Sum }
func (e SumTypeFnc) TypeFnc() TyFnc                         { return Type }
func (e SumTypeFnc) Ident() Value                           { return e }

func (e SumTypeFnc) Eval(dat ...d.Native) d.Native { return d.StrVal(e.String()) }
func (e SumTypeFnc) Cons(t ...TypeDef) SumTypeFnc  { return e.Cons() }
func (e SumTypeFnc) Call(v ...Value) Value {
	t, set := e()
	var vals = []Value{}
	for _, val := range set {
		vals = append(vals, val)
	}
	return NewPair(t, NewVector(vals...))
}
func (e SumTypeFnc) String() string {
	var mem, set = e()
	var str = mem.Name().String() + "∈"
	for i, enum := range set {
		str = str + enum.Name().String()
		if i < len(set)-1 {
			str = str + "|"
		}
	}
	return str
}
