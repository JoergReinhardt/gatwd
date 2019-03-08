package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

type Arity d.Uint8Val

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

func (a Arity) Eval(v ...d.Native) d.Native { return a }
func (a Arity) Int() int                    { return int(a) }
func (a Arity) Flag() d.BitFlag             { return d.BitFlag(a) }
func (a Arity) TypeNat() d.TyNative         { return d.Flag }
func (a Arity) TypeFnc() TyFnc              { return HigherOrder }
func (a Arity) Match(arg Arity) bool        { return a == arg }

// properys relevant for application
type Propertys d.Uint8Val

//go:generate stringer -type Propertys
const (
	Default Propertys = 0
	PostFix Propertys = 1
	InFix   Propertys = 1 + iota
	// ⌐: PreFix
	Atomic
	// ⌐: Composit
	Eager
	// ⌐: Lazy
	Right
	// ⌐: Left_Bound
	Mutable
	// ⌐: Imutable
	SideEffect
	// ⌐: Pure
	Primitive
	// ⌐: Functional
)

func (p Propertys) TypePrime() d.TyNative       { return d.Flag }
func (p Propertys) TypeFnc() TyFnc              { return HigherOrder }
func (p Propertys) Flag() d.BitFlag             { return p.TypeFnc().Flag() }
func (p Propertys) Eval(a ...d.Native) d.Native { return p.Flag() }
func (p Propertys) Match(arg Propertys) bool {
	if p&arg != 0 {
		return true
	}
	return false
}

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
	Expression
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
	Case
	Either
	Or
	If
	Else
	Error
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
	IO
	///////////
	HigherOrder

	Truth      = True | False
	Option     = Just | None
	EitherOr   = Either | Or
	IfElse     = If | Else
	Chain      = Vector | Tuple | Record
	AccIndex   = Vector | Chain
	AccSymbol  = Tuple | AssocVec | Record
	AccCollect = AccIndex | AccSymbol
	Nests      = Tuple | List
	Sets       = UniSet | MuliSet | AssocVec | Record
	Links      = Link | DLink | Node | Tree
)

////////////////////////////////////////////////////////////////////////////////
//// (RE-) INSTANCIATE PRIMARY DATA TO IMPLEMENT FUNCTIONS VALUE INTERFACE
func New(inf ...interface{}) Functional { return NewFromData(d.New(inf...)) }
func NewFromData(data ...d.Native) Native {
	var nat d.Native
	if len(data) > 0 {
		if len(data) > 1 {
			var nats = []d.Native{}
			for _, dat := range data {
				nats = append(nats, dat)
			}
			nat = d.DataSlice(nats)
		}
		nat = data[0]
	}
	return Native(func() d.Native { return nat })
}

func (n Native) String() string                 { return n().String() }
func (n Native) Eval(args ...d.Native) d.Native { return n().Eval(args...) }
func (n Native) TypeNat() d.TyNative            { return n().TypeNat() }
func (n Native) TypeFnc() TyFnc                 { return Data }
func (n Native) Call(vals ...Functional) Functional {
	switch len(vals) {
	case 0:
		return NewFromData(n.Eval())
	case 1:
		return NewFromData(n.Eval(vals[0].Eval()))
	}
	var nat = n()
	for _, val := range vals {
		vals = append(vals, NewFromData(nat.Eval(val.Eval())))
	}
	return NewVector(vals...)
}

// type system implementation
type (
	Native  func() d.Native
	TypeCon func(...TypeId) TypeId
	DataCon func(...d.Native) Native
	ExprCon func(...Functional) Functional
	TypePat func() (string, CaseVal)
	TypeId  func() (
		id int,
		name,
		signature string,
		pattern []TypePat,
	)
	Instance func() (TypeId, Functional)
)

/// TYPE PATTERN
func NewTypePattern(pattern string, cas CaseVal) TypePat {
	return TypePat(func() (string, CaseVal) {
		return pattern, cas
	})
}
func (p TypePat) pattern() string   { pat, _ := p(); return pat }
func (p TypePat) caseExpr() CaseVal { _, exp := p(); return exp }
func (p TypePat) String() string    { return p.pattern() }
func (p TypePat) TypeNat() d.TyNative {
	return d.Function | p.caseExpr().TypeNat()
}
func (p TypePat) TypeFnc() TyFnc {
	return Case | p.caseExpr().TypeFnc()
}
func (p TypePat) Call(val ...Functional) Functional {
	return p.caseExpr().Call(val...)
}
func (p TypePat) Eval(nat ...d.Native) d.Native {
	return p.caseExpr().Eval(nat...)
}

/// TYPE IDENT
func NewTypeId(
	id int, name, signature string, patterns ...TypePat,
) TypeId {
	return func() (int, string, string, []TypePat) {
		return id, name, signature, patterns
	}

}
func (t TypeId) Id() int                           { i, _, _, _ := t(); return i }
func (t TypeId) Name() string                      { _, n, _, _ := t(); return n }
func (t TypeId) Signature() string                 { _, _, s, _ := t(); return s }
func (t TypeId) Patterns() []TypePat               { _, _, _, p := t(); return p }
func (t TypeId) String() string                    { return t.Name() + " = " + t.Signature() }
func (t TypeId) TypeNat() d.TyNative               { return d.Function | d.Type }
func (t TypeId) TypeFnc() TyFnc                    { return Type }
func (t TypeId) Eval(nat ...d.Native) d.Native     { return t }
func (t TypeId) Call(val ...Functional) Functional { return t }

/// DATA, EXPRESSION & TYPE CONSTRUCTORS
type TyCon int8

//go:generate stringer -type=TyCon
const (
	DataContructor TyCon = -1
	ExprContructor TyCon = 0
	TypeContructor TyCon = 1
)

// type, data & expression contructors wrap the call function of the passed
// expression as constructing function
func NewTypeConstructor(
	expr func(types ...TypeId) TypeId,
) TypeCon {
	return TypeCon(expr)
}
func (t TypeCon) String() string      { return t().String() }
func (t TypeCon) TypeCon() TyCon      { return TypeContructor }
func (t TypeCon) TypeNat() d.TyNative { return d.Function | d.Type }
func (t TypeCon) TypeFnc() TyFnc      { return Type | Constructor }
func (t TypeCon) Call(vals ...Functional) Functional {
	var tids = []TypeId{}
	for _, val := range vals {
		tids = append(tids, val.(TypeId))
	}
	return t(tids...)
}
func (t TypeCon) Eval(nats ...d.Native) d.Native {
	var tids = []TypeId{}
	for _, nat := range nats {
		tids = append(tids, nat.(TypeId))
	}
	return t(tids...)
}

func NewDataConstructor(
	expr func(nats ...d.Native) Native,
) DataCon {
	return DataCon(expr)
}
func (t DataCon) String() string      { return t().String() }
func (t DataCon) TypeCon() TyCon      { return DataContructor }
func (t DataCon) TypeNat() d.TyNative { return d.Function | d.Data }
func (t DataCon) TypeFnc() TyFnc      { return Data | Constructor }
func (t DataCon) Call(vals ...Functional) Functional {
	var nats = []d.Native{}
	for _, val := range vals {
		nats = append(nats, val)
	}
	return t(nats...)
}
func (t DataCon) Eval(nats ...d.Native) d.Native { return t(nats...) }

func NewExprConstructor(
	expr func(expr ...Functional) Functional,
) ExprCon {
	return ExprCon(expr)
}
func (t ExprCon) String() string                    { return t().String() }
func (t ExprCon) TypeCon() TyCon                    { return ExprContructor }
func (t ExprCon) TypeNat() d.TyNative               { return d.Function | d.Expression }
func (t ExprCon) TypeFnc() TyFnc                    { return Expression | Constructor }
func (t ExprCon) Call(val ...Functional) Functional { return t(val...) }
func (t ExprCon) Eval(nats ...d.Native) d.Native {
	var fncs = []Functional{}
	for _, nat := range nats {
		fncs = append(fncs, NewFromData(nat))
	}
	return t(fncs...)
}
