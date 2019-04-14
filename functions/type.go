package functions

import (
	"bytes"
	d "github.com/joergreinhardt/gatwd/data"
)

//go:generate stringer -type=TyFnc
const (
	Type TyFnc = 1 << iota
	Data
	Function
	///////////
	Applicaple
	Constructor
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

// higher order types are defined, created & enumerated dynamicly during
// runtime & identified by a unique number
type TyHO uint

type (
	TypeId  func(arg ...Callable) (int, string, TyFnc, d.TyNative, []Callable)
	TypeReg func(args ...Callable) []TypeId
)

//// TYPE-UID
///
// unique id functionhof some registered, distinct type, yielding uid, name,
// functional & native type and a list of all definitions for that type.
//
// init type id takes a new uid, functional- and native type flags and
// optionally a name and yeilds a new type id function, with empty definition
// set.
func initTypeId(uid int, tf TyFnc, tn d.TyNative, names ...string) TypeId {
	return newTypeId(uid, tf, tn, []Callable{}, names...)
}

// newTypeId takes a list of definitions additional to the initial arguments.
func newTypeId(
	uid int,
	tfnc TyFnc,
	tnat d.TyNative,
	defs []Callable,
	names ...string,
) TypeId {

	// construct name for encloseure
	var name = constructTypeName(tfnc, tnat, names...)
	// allocate list of definitions from passed lis, for enclosure
	var definitions = defs

	// return enclosure literal
	return func(args ...Callable) (
		int,
		string,
		TyFnc,
		d.TyNative,
		[]Callable,
	) {
		// assign current definition to result as fallback for the case
		// where no arguments are passed to a call
		var result = definitions

		// check number of passed arguments
		if len(args) == 0 {
			// range through args and try applying args to
			// different methods
			for _, arg := range args {

				// LOOKUP
				if def, ok := lookupDef(arg, definitions); ok {
					result = append(result, def)
					continue
				}

				// APPEND
				if def, ok := appendDef(arg, definitions); ok {
					result = append(result, def)
					continue
				}

				// REPLACE
				if def, ok := replaceDef(arg, definitions); ok {
					result = append(result, def)
					continue
				}
				// jump to process next argument scince this
				// one failed to process at all
				continue
			}
		}
		// yield updated, or current type id
		return uid, name, tfnc, tnat, result
	}
}

// get length of set of definitions, to determine what the next uid to generate
// must be.
func (t TypeId) nextUid() int { return len(t.Definitions()) }

// concat predeccessor names, with base type names
func constructTypeName(tfnc TyFnc, tnat d.TyNative, names ...string) string {
	// string buffer for name concatenation
	var strbuf = bytes.NewBuffer([]byte{})

	// concat all passed name segments
	for _, subname := range names {

		strbuf.WriteString(subname)
		// divide with spaces
		strbuf.WriteString(" ")
	}

	// append base type names
	strbuf.WriteString(tfnc.String() + " " + tnat.String())

	// render full name for enclosure
	return strbuf.String()
}

// lookup definition
func lookupDef(arg Callable, defs []Callable) (Callable, bool) {
	if uid, ok := arg.Eval().(d.IntVal); ok {
		var n = int(uid)
		if n < len(defs) {
			return defs[n], true
		}
	}
	return NewNoOp(), false
}

// append definition
func appendDef(arg Callable, defs []Callable) (Callable, bool) {

	if typ, ok := arg.(TypeId); ok {

		var uid = len(defs)

		defs = append(
			defs,
			typ.Definitions()...,
		)

		return newTypeId(
			uid,
			typ.TypeFnc(),
			typ.TypeNat(),
			defs,
			typ.Name(),
		), true
	}

	return nil, false
}

// replace existing definition
func replaceDef(arg Callable, defs []Callable) (TypeId, bool) {

	if pair, ok := arg.(PairFnc); ok {
		// left is expected to be the uid
		// (index position), right is supposed
		// to be the defining callable.
		id, typ := pair()

		// if uid is an index‥.
		if uid, ok := id.Eval().(d.IntVal); ok {

			// copy of definition list
			var result = defs
			var n = int(uid)

			// in case that index allready exists
			if n < len(defs) {

				var inst TypeId
				var ok bool

				// if instance is TypeId
				if inst, ok = typ.(TypeId); ok {
					// assign new definition to existing index
					defs[n] = inst
					// update result with updated index
					result = defs

				}

				// return fresh copy of updated type id
				return newTypeId(
						inst.Uid(),
						inst.TypeFnc(),
						inst.TypeNat(),
						result,
						inst.Name(),
					),
					true
			}

		}
	}
	return nil, false
}

// convienience lookupN method, takes a variadic number of integers, to lookup
// and return the corresponding type id functions.
func (t TypeId) LookupDefs(args ...int) []Callable {

	var result = []Callable{}

	for _, arg := range args {

		// convert argument to native type
		var uid = NewNative(d.IntVal(arg))
		// pass on uid, by uid and append returns
		// de-slice every single value from the results slice
		var _, _, _, _, defs = t(uid)

		result = append(result, defs[arg])
	}

	return result
}

// looks up a single type id function by it's uid
func (t TypeId) LookupDef(arg int) Callable {

	// convert argument to native type
	var uid = NewNative(d.IntVal(arg))
	// pass on uid, by uid and append returns
	// de-slice every single value from the results slice
	var _, _, _, _, result = t(uid)

	if len(result) > 0 {
		return result[0]
	}
	return NewNoOp()
}

func (t TypeId) AppendDefs(args ...Callable) TypeId {

	// generate new instance to enclose updated list of definitions
	var result = t

	// range over arguments, apply one by one as argument to results
	// AppenOne, overwrite result with every iteration.
	for _, arg := range args {
		result = result.AppendDef(arg)
	}

	// return final version of the type id
	return result
}

// pass one argument to get appendet to the set of definitions for this type.
// yields a reference to the updated type id function
func (t TypeId) AppendDef(arg Callable) TypeId {

	// call function, passing the argument first, to yeild updated result
	var uid, name, tfnc, tnat, defs = t(arg)

	// return fresh instance with updated definition set
	return newTypeId(uid, tfnc, tnat, defs, name)
}

func (t TypeId) Ident() Callable         { return t }
func (t TypeId) String() string          { return t.Name() }
func (t TypeId) Uid() int                { uid, _, _, _, _ := t(); return uid }
func (t TypeId) Name() string            { _, name, _, _, _ := t(); return name }
func (t TypeId) TypeFnc() TyFnc          { _, _, tfnc, _, _ := t(); return tfnc }
func (t TypeId) TypeNat() d.TyNative     { _, _, _, tnat, _ := t(); return tnat }
func (t TypeId) Definitions() []Callable { _, _, _, _, defs := t(); return defs }

// apply args to the function
func (t TypeId) Call(args ...Callable) Callable {
	var u, n, tf, tn, d = t(args...)
	return newTypeId(u, tf, tn, d, n)
}

// evaluation of the type id function yields the types uid
func (t TypeId) Eval(...d.Native) d.Native { return d.IntVal(t.Uid()) }

//////////////////////////////////////////////////////////////////////////////
//// TYPE REGISTRY
///
// a type registry takes either no arguments, to return the vector of all
// previously defined types sorted by uid, one, or more type identitys, to add
// to the vector of defined types, one, or more uint values, to perform a type
// lookup on

func NewTypeReg() TypeReg {

	var registry = []TypeId{}

	return func(args ...Callable) []TypeId {

		var result = registry

		if len(args) > 0 {
		}

		return result
	}
}

func lookupType(arg Callable, registry []Callable) (Callable, bool) {

	//if uid, ok := arg.Eval().(d.IntVal); ok {

	//var sort = dataSorter()

	//	sort.Sort()
	//}

	return nil, false
}

func (t TypeReg) Ident() Callable     { return t }
func (t TypeReg) TypeFnc() TyFnc      { return HigherOrder }
func (t TypeReg) TypeNat() d.TyNative { return d.Type }
func (t TypeReg) Call(args ...Callable) Callable {
	var nargs []d.Native
	return NewNative(t.Eval(nargs...))
}
func (t TypeReg) Eval(args ...d.Native) d.Native {
	var result = NewVector()
	return result
}
func (t TypeReg) String() string { return t.Eval().String() }

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
type TyFnc d.UintVal

func (t TyFnc) Eval(...d.Native) d.Native { return t }
func (t TyFnc) TypeNat() d.TyNative       { return d.Flag }
func (t TyFnc) Flag() d.BitFlag           { return d.BitFlag(t) }
func (t TyFnc) Uint() uint                { return d.BitFlag(t).Uint() }
