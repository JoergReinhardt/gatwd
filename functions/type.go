package functions

import (
	"sync"

	d "github.com/joergreinhardt/gatwd/data"
)

//go:generate stringer -type=TyFnc
const (
	Type TyFnc = 1 << iota
	Data
	Function
	Constructor
	///////////
	Endofunctor
	Applicaple
	Operator
	Functor
	Monad
	///////////
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
	//////////
	Truth
	Number
	Symbol
	Error
	Pair
	Tuple
	Enum
	Set
	List
	Vector
	Record
	///////////
	HigherOrder

	Kind = Data | Function

	Morphisms = Applicaple | Constructor | Functor | Monad

	Option = Just | None | Case | Switch |
		Either | Or | If | Else | Truth

	Boxed = Pair | Enum | Option

	Collection = List | Vector | Record | Set
)

func initTypeSystem() []func(...Callable) (HOTypeCon, bool) {

	var reg = &struct {
		Lock *sync.RWMutex
		Map  map[string]int
		Reg  []HOTypeCon
	}{
		Lock: &sync.RWMutex{},
		Map:  make(map[string]int),
		Reg:  []HOTypeCon{},
	}

	var methods []func(...Callable) (HOTypeCon, bool)

	methods = []func(...Callable) (HOTypeCon, bool){

		// BY INDEX
		func(args ...Callable) (HOTypeCon, bool) {

			if len(args) > 0 {
				if arg, ok := args[0].Eval().(d.IntVal); ok {
					var idx = arg.Int()
					if idx < len(reg.Reg) {
						return reg.Reg[idx], true
					}
				}
			}
			return nil, false
		},

		// BY NAME
		func(args ...Callable) (HOTypeCon, bool) {

			if len(args) > 0 {
				var name = args[0].String()
				if idx, ok := reg.Map[name]; ok {
					if idx < len(reg.Reg) {
						return reg.Reg[idx], true
					}
				}
			}
			return nil, false
		},

		// CREATE
		func(args ...Callable) (HOTypeCon, bool) {

			if len(args) > 0 {

				var arg = args[0]
				var match = arg.TypeFnc().Flag().Match

				if match(Symbol) {
					//if str, ok := arg.(Text); ok {
					//}

				}
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
	// ⌐: Thunk
	Eager
	// ⌐: Lazy
	Right
	// ⌐: Left_Bound
	Mutable
	// ⌐: Imutable
	SideEffect
	// ⌐: Pure
	Primitive
	// ⌐: Parametric
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
type TyFnc d.BitFlag

func (t TyFnc) TypeFnc() TyFnc                 { return Type }
func (t TyFnc) TypeNat() d.TyNative            { return d.Flag }
func (t TyFnc) Call(args ...Callable) Callable { return t.TypeFnc() }
func (t TyFnc) Eval(args ...d.Native) d.Native { return t.TypeNat() }
func (t TyFnc) Flag() d.BitFlag                { return d.BitFlag(t) }
func (t TyFnc) Uint() uint                     { return d.BitFlag(t).Uint() }
