package functions

import (
	"sync"

	d "github.com/joergreinhardt/gatwd/data"
)

//go:generate stringer -type=TyFnc
const (
	/// TYPE RELATED FLAGS ///
	Type TyFnc = 1 << iota
	TypeSum
	TypeProduct
	Constructor
	Expression
	ExprProperys
	ExprArity
	Data
	/// FUNCTORS AND MONADS ///
	Endofunctor
	Applicable
	Operator
	Functor
	Monad
	/// MONADIC SUB TYPES ///
	False
	True
	Just
	None
	Case
	Switch
	Either
	Or
	If
	Else
	While
	Do
	/// TYPE CLASSES ///
	Equality
	Truth
	Order
	Number
	Symbol
	Error
	/// COLLECTION TYPES ///
	Pair
	Tuple
	Enum
	Set
	List
	Record
	Vector
	/// HIGHER ORDER TYPE IS THE HIGHEST FLAG ///
	HigherOrder

	Kind = Data | Expression

	ExprProps = ExprProperys | ExprArity

	Morphisms = Constructor | Operator | Functor |
		Endofunctor | Applicable | Monad

	Options = False | True | Just | None | Case |
		Switch | Either | Or | If | Else |
		While | Do

	Collections = Pair | Tuple | Enum | Set |
		List | Vector | Record

	Classes = Truth | Equality | Order | Number |
		Symbol | Error
)

//// TYPE SYSTEM
///
// the type system provides stateful methods, enclosing the list of
// typeconstructors and map of type names to be shared by all higher order
// types. types can be looked up and/or be created concurrently.
//
//  TypeMethod is the common signature of all higher order type constructors.
//  they return the higher order type that has been looked up, or created and a
//  boolean to indicate if lookup/creation succeeded.
type TypeMethod func(...Callable) (HOTypeCon, bool)

// TypeMethods is a slice of TypeMethod function pointers and provides golang
// methods to access the methods and the underlying data structure in a
// typesafe manner that encapsulates all possible side effects.
type TypeMethods []TypeMethod

// look up type by uid
func (t TypeMethods) LookupIdx(uid Callable) (HOTypeCon, bool)  { return t[0](uid) }
func (t TypeMethods) LookupIdxNative(uid int) (HOTypeCon, bool) { return t[0](New(uid)) }

// look up type by name
func (t TypeMethods) LookupName(name Callable) (HOTypeCon, bool)     { return t[1](name) }
func (t TypeMethods) LookupNameNative(name string) (HOTypeCon, bool) { return t[1](New(name)) }

// create a new higher order type
func (t TypeMethods) Create(args ...Callable) (HOTypeCon, bool) { return t[2](args...) }

func initTypeSystem() []TypeMethod {

	var reg = &struct {
		Lock *sync.RWMutex
		Map  map[string]int
		Reg  []HOTypeCon
	}{
		Lock: &sync.RWMutex{},
		Map:  make(map[string]int),
		Reg:  []HOTypeCon{},
	}

	var methods TypeMethods

	methods = TypeMethods{

		// - 0 - LOOKUP BY INDEX
		func(args ...Callable) (HOTypeCon, bool) {

			if len(args) > 0 {

				if arg, ok := args[0].Eval().(d.IntVal); ok {

					var idx = arg.Int()

					if idx < len(reg.Reg) {

						reg.Lock.Lock()
						defer reg.Lock.Unlock()

						if reg.Reg[idx] != nil {

							return reg.Reg[idx], true
						}
					}
				}
			}
			return nil, false
		},

		// - 1 - LOOKUP BY NAME
		func(args ...Callable) (HOTypeCon, bool) {

			if len(args) > 0 {

				var name = args[0].String()

				reg.Lock.Lock()
				defer reg.Lock.Unlock()

				if idx, ok := reg.Map[name]; ok {

					if idx < len(reg.Reg) {

						if reg.Reg[idx] != nil {

							return reg.Reg[idx], true
						}
					}
				}
			}
			return nil, false
		},

		// - 2 - CREATE
		func(args ...Callable) (HOTypeCon, bool) {

			if len(args) > 0 {
			}

			return nil, false
		}}

	return methods
}

//////////////////////////////////////////////////////////////////////
//// SEMANTIC CALL PROPERTYS
///
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
func (a Arity) Signature() []Callable {
	return []Callable{
		NewFromFlag(Type),
		NewFromFlag(ExprArity),
	}
}
func (a Arity) Match(arg Arity) bool { return a == arg }

// properys relevant for application
type Propertys d.Uint8Val

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

func (p Propertys) PostFix() bool    { return p.Flag().Match(PostFix) }
func (p Propertys) InFix() bool      { return !p.Flag().Match(PostFix) }
func (p Propertys) Atomic() bool     { return p.Flag().Match(Atomic) }
func (p Propertys) Thunk() bool      { return !p.Flag().Match(Atomic) }
func (p Propertys) Eager() bool      { return p.Flag().Match(Eager) }
func (p Propertys) Lazy() bool       { return !p.Flag().Match(Eager) }
func (p Propertys) RightBound() bool { return p.Flag().Match(RightBound) }
func (p Propertys) LeftBound() bool  { return !p.Flag().Match(RightBound) }
func (p Propertys) Mutable() bool    { return p.Flag().Match(Mutable) }
func (p Propertys) Imutable() bool   { return !p.Flag().Match(Mutable) }
func (p Propertys) SideEffect() bool { return p.Flag().Match(SideEffect) }
func (p Propertys) Pure() bool       { return !p.Flag().Match(SideEffect) }
func (p Propertys) Primitive() bool  { return p.Flag().Match(Primitive) }
func (p Propertys) Parametric() bool { return !p.Flag().Match(Primitive) }

func (p Propertys) TypePrime() d.TyNative { return d.Flag }
func (p Propertys) TypeFnc() TyFnc        { return HigherOrder }
func (p Propertys) Signature() []Callable {
	return []Callable{
		NewFromFlag(Type),
		NewFromFlag(ExprProperys),
	}
}

func (p Propertys) Flag() d.BitFlag             { return d.BitFlag(uint64(p)) }
func FlagToProp(flag d.BitFlag) Propertys       { return Propertys(uint8(flag.Uint())) }
func (p Propertys) Eval(a ...d.Native) d.Native { return p.Flag() }

func (p Propertys) MatchProperty(arg Propertys) bool {
	if p&arg != 0 {
		return true
	}
	return false
}

func (p Propertys) Match(flag d.BitFlag) bool { return p.Flag().Match(flag) }
func (p Propertys) Print() string {

	var flags = p.Flag().Decompose()
	var str string
	var l = len(flags)

	if l > 1 {
		for i, flag := range flags {
			str = str + FlagToProp(flag).String()
			if i < l-1 {
				str = str + " "
			}
		}
	}

	return p.String()
}

// type TyFnc d.BitFlag
// encodes the kind of functional data as bitflag
type TyFnc d.BitFlag

func (t TyFnc) TypeFnc() TyFnc                 { return Type }
func (t TyFnc) TypeNat() d.TyNative            { return d.Flag }
func (t TyFnc) Call(args ...Callable) Callable { return t.TypeFnc() }
func (t TyFnc) Eval(args ...d.Native) d.Native { return t.TypeNat() }
func (t TyFnc) Flag() d.BitFlag                { return d.BitFlag(t) }
func (c TyFnc) Signature() []Callable {
	return []Callable{
		NewFromFlag(Type),
		NewFromFlag(Expression),
	}
}
func (t TyFnc) Uint() uint { return d.BitFlag(t).Uint() }
