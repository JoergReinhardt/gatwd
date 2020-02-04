/*

PRODUCT TYPES
-------------
*/
package functions

import (
	"fmt"

	d "github.com/joergreinhardt/gatwd/data"
)

type (
	//// TUPLE
	TupCons Def
	TupVal  Def

	//// RECORD
	RecCons Def
	RecVal  TupVal

	// POLYMORPH
	Poly Def

	// OPTIONAL
	Option Def

	// ALTERNATIVE
	Alter Poly

	// SWITCH
	Switch Poly
)

///////////////////////////////////////////////////////////////////////////////
//// DEFINE TUPLE TYPE CONSTRUCTOR
///
// defines a constructor to take arguments matching the tuple signature and
// return an instanciated tuple constant in accordance with the definition.
// the tuple value is an instance of an alias type of vector, created from
// those arguments, in case they match the signature, or none, in case they
// dont, or an instance of a partialy applied expression, in case an
// insufficient number of matching arguments has been passed.
func NewTupleCons(types ...d.Typed) TupCons {

	var (
		sym      d.Typed
		argtypes = make([]Functor, 0, len(types))
	)

	// extract name if symbol has been passed at first position,
	// else use functional type tuple as type identity
	if len(types) > 0 {
		if Kind_Symb.Match(types[0].Kind()) {
			sym = types[0].(TySym)
			if len(types) > 1 {
				types = types[1:]
			} else {
				types = types[:0]
			}
		} else {
			sym = Tuple
		}
	}

	// cast declaration cell types as functors, declare cell type
	// first, if it is a flag type, and append to slice of elements
	// later to return as vector, when constructor is called without
	// arguments (parameter overload to return constructor definition)
	for _, t := range types {
		if Kind_Nat.Match(t.Kind()) {
			argtypes = append(argtypes, Declare(t))
		}
		argtypes = append(argtypes, t.(Functor))
	}

	// data constructor for this particular tuple type
	return TupCons(Define(Lambda(func(args ...Functor) Functor {
		if len(args) > 0 {
			// returns an instance of tuple value
			return TupVal(Define(
				NewVector(args...),
				sym, DecAll(types...)))
		}
		return NewVector(argtypes...)
	}),
		Tuple, Declare(Tuple, DecAll(types...)),
		Declare(types...)))
}
func (t TupCons) Unbox() Functor { return Def(t).Unbox() }
func (t TupCons) GetCellType(idx int) d.Typed {
	if elem, ok := t.Unbox().Call().(VecVal).Get(idx); ok {
		return elem.Type()
	}
	return None
}
func (t TupCons) TypeFnc() TyFnc               { return Tuple }
func (t TupCons) Type() Decl                   { return Def(t).Type() }
func (t TupCons) TypeId() Decl                 { return Def(t).TypeId() }
func (t TupCons) TypeRet() Decl                { return Def(t).TypeRet() }
func (t TupCons) TypeArgs() Decl               { return Def(t).TypeArgs() }
func (t TupCons) Call(args ...Functor) Functor { return Def(t).Call(args...) }
func (t TupCons) String() string               { return t.TypeName() }
func (t TupCons) TypeName() string {
	return t.TypeArgs().TypeName() + " → " +
		t.TypeId().TypeName() + " → " +
		t.TypeRet().TypeName()

}

/// TUPLE VALUE
// tuple values are created by applying appropriate arguments to the
// associated tuple type definition/constructor.
func (t TupVal) Unbox() Functor { return Def(t).Unbox() }
func (t TupVal) Vector() VecVal { return t.Unbox().(VecVal) }
func (t TupVal) Get(idx int) Functor {
	if elem, ok := t.Vector().Get(idx); ok {
		return elem
	}
	return NewNone()
}
func (t TupVal) String() string                   { return t.Vector().String() }
func (t TupVal) Type() Decl                       { return Def(t).Type() }
func (t TupVal) TypeId() Decl                     { return Def(t).TypeId() }
func (t TupVal) TypeRet() Decl                    { return Def(t).TypeRet() }
func (t TupVal) TypeFnc() TyFnc                   { return Tuple }
func (t TupVal) TypeElem() Decl                   { return t.TypeRet() }
func (t TupVal) Continue() (Functor, Applicative) { return t.Vector().Continue() }
func (t TupVal) Head() Functor                    { return t.Vector().Head() }
func (t TupVal) Tail() Applicative                { return t.Vector().Tail() }
func (t TupVal) Empty() bool                      { return t.Vector().Empty() }

// call envoced without arguments, returns all cell values wrapped in a vector
// of mixed type elements. when arguments are passed, they are expected to be
// integer index accessors, in which case (the) element(s) associated with the
// passed index value(s) will be returned.
func (t TupVal) Call(args ...Functor) Functor {
	if len(args) > 0 { // assume arguments are index cell accessors
		if len(args) > 1 { // retrieve sequence of cells
			var (
				elems = make([]Functor, 0, len(args))
				types = make([]d.Typed, 0, len(args))
			)
			for _, arg := range args {
				if !arg.Type().Match(None) {
					elems = append(elems, t.Call(arg))
					types = append(types, arg.Type())
				}
			}
			return TupVal(Define( // return tuple of chosen cells
				NewVector(elems...), Tuple,
				DecAll(types...)))
		}
		if args[0].TypeFnc().Match(Atomic) {
			if eve, ok := args[0].(Evaluable); ok {
				if ok := eve.Eval().
					Type().Match(d.Int); ok {
					if elem, ok := t.Call().(VecVal).Get(
						eve.Eval().(d.Integer).GoInt(),
					); ok {
						return elem
					}
				}
			}
		}
	}
	return t.Vector()
}

// create an anonymous ad-hoc tuple from a bunch of arguments
func AllocTuple(args ...Functor) TupVal {
	var types = make([]d.Typed, 0, len(args))
	for _, arg := range args {
		types = append(types, arg.Type())
	}
	return TupVal(Define( // return tuple of chosen cells
		NewVector(args...), Tuple,
		DecAll(types...)))
}

///////////////////////////////////////////////////////////////////////////////
//// RECORD CONSTRUCTOR DEFINITION
///
// alloc-record expects key/value pairs as arguments to derive field
// type names and define field type constructors and then applys the

func NewRecordCon(defs ...Def) RecCons {
	var (
		cons  TupCons            // tuple constructor
		names = d.NewStringMap() // name to index position map
		types = make([]d.Typed, 0, len(defs))
	)
	// range defs to extract type ids & names
	for pos, def := range defs {
		types = append(types, def.TypeId())
		names = names.Set(
			d.StrVal(def.TypeId().TypeName()),
			d.IntVal(pos),
		).(d.MapString)
	}

	cons = NewTupleCons(types...)

	return RecCons(Define(Lambda(func(args ...Functor) Functor {

		var tup = cons.Call(args...).(TupVal)

		return RecVal(Define(Lambda(func(args ...Functor) Functor {
			// element acess by index, org key
			if len(args) > 0 {
			}

			return RecVal(tup)

		}), Record, cons.TypeRet()))
	}), Record, cons.TypeRet(), cons.TypeArgs()))
}
func (r RecCons) Call(args ...Functor) Functor { return Def(r).Call(args...) }
func (r RecCons) Unbox() Functor               { return Def(r).Unbox() }
func (r RecCons) Type() Decl                   { return Def(r).Type() }
func (r RecCons) TypeId() Decl                 { return Def(r).TypeId() }
func (r RecCons) TypeRet() Decl                { return Def(r).TypeRet() }
func (r RecCons) TypeArgs() Decl               { return Def(r).TypeArgs() }
func (r RecCons) TypeFnc() TyFnc               { return Record }
func (r RecCons) String() string               { return r.TypeName() }
func (r RecCons) TypeName() string             { return "" }

func (r RecVal) Len() int                     { return Def(r).Len() }
func (r RecVal) TypeFnc() TyFnc               { return Record }
func (r RecVal) Type() Decl                   { return Def(r).Type() }
func (r RecVal) TypeId() Decl                 { return Def(r).TypeId() }
func (r RecVal) TypeRet() Decl                { return Def(r).TypeRet() }
func (r RecVal) TypeArgs() Decl               { return Def(r).TypeArgs() }
func (r RecVal) TypeName() string             { return Def(r).TypeName() }
func (r RecVal) String() string               { return r.Unbox().String() }
func (r RecVal) Unbox() Functor               { return Def(r).Unbox() }
func (r RecVal) Call(args ...Functor) Functor { return r.Unbox().Call(args...) }

///////////////////////////////////////////////////////////////////////////////
//// DEFINE POLYMORPHIC TYPE
///
// a parametric definition is a vector of function definitions sharing a common
// symbol.  when called with arguments, they will be folded over every
// definition in that vector to return return none, or either an instance, of a
// partial, or final value.  as long as the fold operation continues to return
// instances of partial value, without returning a final value, another
// parametric definition will be returned, defined by all remaining partial
// instances, to be applied to succeeding arguments recursively.
func NewPolyMorph(symbol d.Typed, defs ...Def) Poly {

	var (
		parms = NewVector()       // vector to hold definitions
		name  = symbol.TypeName() // string version of type symbol

		ats = make([]d.Typed, 0, len(defs)) // slice of argument types
		rts = make([]d.Typed, 0, len(defs)) // slice of return types
	)

	for n, def := range defs { // range over definitions
		parms = parms.ConsVec(def) // concatenate to vectors
		ats = append(ats, def.TypeArgs())
		rts = append(ats, def.TypeRet())
		if name == "" { // if the name has not been set‥.
			// compose name from dot cocatenating of all subtype names
			name = name + def.TypeId().TypeName()
			if n < len(defs)-1 {
				name = name + "."
			}
		}
	}

	// define Polymorph from vector of parameters, declare name &
	// concatenate return & argument types
	return Poly(Define(parms, DecSym(name),
		DecAny(ats...), DecAny(rts...)))
}

// vector returns a vector of all (remaining) definitions
func (p Poly) Unbox() Functor { return Def(p).Unbox() }
func (p Poly) Vector() VecVal { return p.Unbox().(VecVal) }
func (p Poly) Len() int       { return p.Vector().Len() }
func (p Poly) Type() Decl     { return Def(p).Type() }
func (p Poly) TypeId() Decl   { return Def(p).TypeId() }
func (p Poly) TypeRet() Decl  { return Def(p).TypeRet() }
func (p Poly) TypeArgs() Decl { return Def(p).TypeArgs() }
func (p Poly) String() string { return Def(p).TypeName() }
func (p Poly) TypeFnc() TyFnc {
	// set function type to partial, if its the case
	if IsPartial(p.Vector().Head()) {
		return Partial | Polymorph
	}
	return Polymorph
}
func (p Poly) TypeName() string {
	//var str = "case x in\n"
	var str = p.TypeId().TypeName() + " ∷\n"

	for _, f := range p.Vector()() {
		var def = f.(Def)
		str = str + "\t" + def.TypeArgs().TypeName() +
			" ＝ " + def.TypeId().TypeName() +
			" → " + def.TypeRet().TypeName() + "\n"
	}

	return str
}

// call evaluates application of arguments to each enclosed definition, as long
// as partial results are yielded, until either a final result, or an instance
// of none is returned
func (p Poly) Call(args ...Functor) Functor {

	// allocate slice to hold partialy applied return values
	var partials = []Def{}

	for _, def := range p.Vector()() { // range over constructors
		var result = def.Call(args...) // apply args
		if !IsNone(result) {           // not none‥.
			if !IsPartial(result) { // ‥.not partial‥.
				return result // ‥.⇒ first & final result
			} // else append to partials
			partials = append(partials,
				result.(Def))
		}
	}

	// as long as no final result has been yielded, but partial
	// applications are returned, keep calm &carry on
	if len(partials) > 0 { // if there are partials‥.
		if len(partials) > 1 { //‥.if there is more than one‥.
			return NewPolyMorph( //‥.define polymorphic‥.
				p.TypeId(), partials...) //‥.from partials‥.
		}
		return partials[0] //‥.⇒ single remaining partial
	}
	return NewNone() //‥.⇒ none, when no args passed
}

//// OPTIONAL TYPE DEFINITION
///
// when arguments are applied to function definitions either a value is
// returned, or an instance of none, in the case where passed arguments fail to
// match the functions signature for instance.  user defined functions may also
// selectively return none, not depending on the on arguments types, but their
// values instead, in reaction to encapsulated side effects, or as natural
// result of the computation they where applied to.  the optional type exposes
// that property explicitly by being defined with a return type of Just | None.
func NewOption(def Def) Option { return Option(def) }

func (o Option) TypeFnc() TyFnc   { return Options }
func (o Option) Type() Decl       { return Def(o).Type() }
func (o Option) TypeId() Decl     { return Def(o).TypeId() }
func (o Option) TypeArgs() Decl   { return Def(o).TypeArgs() }
func (o Option) TypeName() string { return Def(o).TypeName() }
func (o Option) String() string   { return o.TypeName() }
func (o Option) TypeRet() Decl {
	return Declare(Options, DecAny(Def(o).TypeRet(), None))
}
func (o Option) Unbox() Functor { return Def(o).Unbox() }
func (o Option) Call(args ...Functor) Functor {
	if len(args) > 0 {
		return o.Unbox().Call(args...)
	}
	return o.Unbox().Call()
}

//// ALTERNATIVES
///
// alternatives returns a value of either the first, or second of the two
// defined return type alternatives.
func NewAlternative(l, r Def) Alter {
	fmt.Printf("l: %s, r: %s\n", l, r)
	return Alter(NewPolyMorph(Declare(
		Alternatives,
		DecAny(
			DecAll(Either, l.TypeId()),
			DecAll(Or, r.TypeId()),
		),
		DecAny(
			l.TypeArgs(),
			r.TypeArgs(),
		),
	), l, r))
}

func (a Alter) TypeFnc() TyFnc { return Alternatives }
func (a Alter) Unbox() Functor { return Poly(a).Unbox() }
func (a Alter) Type() Decl     { return Poly(a).Type() }
func (a Alter) TypeId() Decl   { return Poly(a).TypeId() }
func (a Alter) TypeArgs() Decl { return Poly(a).TypeArgs() }
func (a Alter) TypeRet() Decl {
	return Declare(Alternatives, Poly(a).TypeRet())
}
func (a Alter) TypeName() string             { return Poly(a).TypeName() }
func (a Alter) String() string               { return a.TypeName() }
func (a Alter) Call(args ...Functor) Functor { return Poly(a).Call(args...) }

//// SWITCH
///
// alternatives returns a value of either the first, or second of the two
// defined return type alternatives.
func NewSwitch(args ...Def) Switch {
	return Switch(NewPolyMorph(Choice, args...))
}

func (a Switch) TypeFnc() TyFnc { return Choice }
func (a Switch) Unbox() Functor { return Poly(a).Unbox() }
func (a Switch) Type() Decl     { return Poly(a).Type() }
func (a Switch) TypeId() Decl   { return Poly(a).TypeId() }
func (a Switch) TypeArgs() Decl { return Poly(a).TypeArgs() }
func (a Switch) TypeRet() Decl {
	return Declare(Choice, Poly(a).TypeRet())
}
func (a Switch) TypeName() string             { return Poly(a).TypeName() }
func (a Switch) String() string               { return a.TypeName() }
func (a Switch) Call(args ...Functor) Functor { return Poly(a).Call(args...) }
