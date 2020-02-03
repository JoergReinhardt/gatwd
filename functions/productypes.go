/*

PRODUCT TYPES
-------------
*/
package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

type (
	// BOOL VALUE TYPES
	Bool    bool
	Bitwise d.BitFlag

	// BOOLEAN ALGEBRA
	BoolOp    Def
	BitwiseOp Def

	// TEST & COMPARE
	Test    Def
	Compare Def

	// OPTIONAL TYPE
	Option Def

	// POLYMORPHIC EXPRESSION
	PolyDef Def

	// ALTERNATIVE TYPE
	Altern PolyDef
)

//// TRUTH VALUE
///
// truth value aliases the native bool type & returns its function type as
// either 'True', or 'False' depending on the aliased instance
func (b Bool) TypeFnc() TyFnc {
	if b {
		return True
	}
	return False
}
func (b Bool) Type() Decl                   { return Declare(b.TypeFnc()) }
func (b Bool) Or(x Bool) Bool               { return b || x }
func (b Bool) Xor(x Bool) Bool              { return b != x }
func (b Bool) And(x Bool) Bool              { return b && x }
func (b Bool) Not() Bool                    { return !b }
func (b Bool) Call(args ...Functor) Functor { return b.Call(args...) }
func (b Bool) String() string {
	if b {
		return "True"
	}
	return "False"
}
func (b Bool) Continue() (Functor, Applicative) { return b, NewNone() }
func (b Bool) Concat(seq Sequential) Applicative {
	if seq.TypeElem().Match(Truth) {
		return NewList(b).Concat(seq)
	}
	if seq.TypeElem().Match(Truth) {
		if b {
			return NewList(Bitwise(0)).Concat(seq)
		}
		return NewList(Bitwise(1)).Concat(seq)
	}
	return NewNone()
}

func (b Bitwise) String() string               { return d.BitFlag(b).String() }
func (b Bitwise) TypeFnc() TyFnc               { return Truth | Byte }
func (b Bitwise) Type() Decl                   { return Declare(Truth, Byte) }
func (b Bitwise) Match(t d.Typed) bool         { return d.BitFlag(b).Match(t) }
func (b Bitwise) InSet(bit Bitwise) bool       { return d.BitFlag(b).Match(d.BitFlag(bit)) }
func (b Bitwise) Uint() d.UintVal              { return d.BitFlag(b).Uint() }
func (b Bitwise) Not() Bitwise                 { return Bitwise(b.Uint() ^ T.Uint()) }
func (b Bitwise) And(x Bitwise) Bitwise        { return Bitwise(b.Uint() & x.Uint()) }
func (b Bitwise) Or(x Bitwise) Bitwise         { return Bitwise(b.Uint() | x.Uint()) }
func (b Bitwise) Xor(x Bitwise) Bitwise        { return Bitwise(b.Uint() ^ x.Uint()) }
func (b Bitwise) Call(args ...Functor) Functor { return b }

//// BOOLEAN OPERATIONS
///
//
func DefBoolOp(fnc func(args ...Functor) Bool) BoolOp {
	return BoolOp(Define(Lambda(func(args ...Functor) Functor {
		return fnc(args...)
		// is a bool, returns bool, may take arguments of any type
	}), Boolean, Boolean, T))
}
func (b BoolOp) Call(args ...Functor) Functor {
	return b.Call(args...)
}
func (b BoolOp) Type() Decl {
	return Declare(Truth, Truth, Truth)
}
func (b BoolOp) TypeFnc() TyFnc   { return Boolean }
func (b BoolOp) TypeId() Decl     { return Declare(Truth) }
func (b BoolOp) TypeRet() Decl    { return Declare(Truth) }
func (b BoolOp) TypeArgs() Decl   { return Declare(Truth) }
func (b BoolOp) TypeName() string { return Declare(Truth).TypeName() }
func (b BoolOp) String() string   { return b.TypeName() }

//// BOOLEAN ALGEBRA FOR BOOL & BITWISE INSTANCES
var (
	OR = DefinePolymorph(DecSym("|"),
		Define(Lambda(func(args ...Functor) Functor {
			return Bool(args[0].(Bool) || args[1].(Bool))
		}),
			Truth, Truth, DecAll(Truth, Truth)),
		Define(Lambda(func(args ...Functor) Functor {
			return Bitwise(args[0].(Bitwise) | args[1].(Bitwise))
		}),
			Declare(Truth|Byte), Truth,
			DecAll(Declare(Truth|Byte), Declare(Truth|Byte))),
	)

	XOR = DefinePolymorph(DecSym("⊻"),
		Define(Lambda(func(args ...Functor) Functor {
			return Bool(args[0].(Bool) != args[1].(Bool))
		}), Truth, Truth, DecAll(Truth, Truth)),
		Define(Lambda(func(args ...Functor) Functor {
			return Bitwise(args[0].(Bitwise) ^ args[1].(Bitwise))
		}),
			Declare(Truth|Byte), Truth,
			DecAll(Declare(Truth|Byte), Declare(Truth|Byte))),
	)

	AND = DefinePolymorph(DecSym("&"),
		Define(Lambda(func(args ...Functor) Functor {
			return Bool(args[0].(Bool) && args[1].(Bool))
		}), Truth, Truth, DecAll(Truth, Truth)),
		Define(Lambda(func(args ...Functor) Functor {
			return Bitwise(args[0].(Bitwise) & args[1].(Bitwise))
		}),
			Declare(Truth|Byte), Truth,
			DecAll(Declare(Truth|Byte), Declare(Truth|Byte))),
	)

	NOT = DefinePolymorph(DecSym("¬"),
		Define(Lambda(func(args ...Functor) Functor {
			return Bool(!args[0].(Bool))
		}), Truth, Truth, Truth),
		Define(Lambda(func(args ...Functor) Functor {
			return args[0].(Bitwise).Not()
		}),
			Declare(Truth|Byte), Truth,
			DecAll(Declare(Truth|Byte), Declare(Truth|Byte))),
	)
)

//// TEST
///
// test takes a function that takes two functors to scrutinize and returns a
// boolean value to indicate test result.
func NewTest(
	atype d.Typed,
	test func(args ...Functor) bool,
) Test {
	return Test(Define(Lambda(func(args ...Functor) Functor {
		return Bool(test(args[0], args[1]))
	}), DecSym("Test"), Truth, Declare(atype, atype)))
}
func (t Test) Unbox() Functor { return Def(t).Unbox() }
func (t Test) TypeFnc() TyFnc {
	return Truth
}
func (t Test) Type() Decl {
	return Declare(
		DecSym("Test"),
		Def(t).TypeRet(),
		Def(t).TypeArgs())
}
func (t Test) String() string {
	return t.TypeFnc().TypeName()
}
func (t Test) Test(args ...Functor) bool {
	if len(args) > 1 {
		return bool(Def(t).Unbox().Call(args[0], args[1]).(Bool))
	}
	return false
}
func (t Test) Compare(a, b Functor) int {
	if t.Test(a, b) {
		return 0
	}
	return -1
}
func (t Test) Call(args ...Functor) Functor { return t(args...) }
func (t Test) Equal() Def {
	return Define(t.Unbox(), Equal, Truth, t.Type().TypeArgs())
}

//// COMPARATOR
///
// comparator takes two functors to compare and returns an integer to indicate
// the result.  if both functors are considered equal by the passed comparing
// expression, zero is retuned, a negative result, if the left argument is
// lesser and a positive result, if its greater than the right argument.
func NewComparator(
	argtype d.Typed,
	comp func(a, b Functor) int,
) Compare {
	return Compare(Define(Lambda(func(args ...Functor) Functor {
		if len(args) == 0 { // return argument type, when called empty
			if Kind_Decl.Match(argtype.Kind()) {
				return argtype.(Decl)
			}
			return Declare(Comparison,
				Declare(Lesser|Greater|Equal),
				Declare(argtype, argtype))
		}
		if comp(args[0], args[1]) < 0 {
			return Lesser
		}
		if comp(args[0], args[1]) > 0 {
			return Greater
		}
		return Equal
	}), DecSym("Compare"),
		Declare(Lesser|Greater|Equal),
		Declare(argtype, argtype)))
}
func (t Compare) Unbox() Functor               { return Def(t).Unbox() }
func (t Compare) Type() Decl                   { return Def(t).Type() }
func (t Compare) TypeRet() Decl                { return Def(t).TypeRet() }
func (t Compare) TypeArgs() Decl               { return Def(t).TypeArgs() }
func (t Compare) String() string               { return Def(t).TypeName() }
func (t Compare) TypeFnc() TyFnc               { return Comparison }
func (t Compare) Call(args ...Functor) Functor { return t(args...) }
func (t Compare) Compare(a, b Functor) TyFnc   { return Def(t).Call(a, b).(TyFnc) }
func (t Compare) Equal(a, b Functor) bool      { return t.Compare(a, b).Match(Equal) }
func (t Compare) Lesser(a, b Functor) bool     { return t.Compare(a, b).Match(Lesser) }
func (t Compare) Greater(a, b Functor) bool    { return t.Compare(a, b).Match(Greater) }

//// DEFINE POLYMORPHIC TYPE
///
// a parametric definition is a vector of function definitions sharing a common
// symbol.  when called with arguments, they will be folded over every
// definition in that vector to return return none, or either an instance, of a
// partial, or final value.  as long as the fold operation continues to return
// instances of partial value, without returning a final value, another
// parametric definition will be returned, defined by all remaining partial
// instances, to be applied to succeeding arguments recursively.
func DefinePolymorph(symbol d.Typed, defs ...Def) PolyDef {

	var (
		parms = NewVector()
		name  = symbol.TypeName()
		ats   = make([]d.Typed, 0, len(defs))
		rts   = make([]d.Typed, 0, len(defs))
	)

	for n, def := range defs {
		parms = parms.ConsVec(def)
		ats = append(ats, def.TypeArgs())
		rts = append(ats, def.TypeRet())
		if name == "" {
			name = name + def.TypeId().TypeName()
			if n < len(defs)-1 {
				name = name + "|"
			}
		}
	}

	return PolyDef(Define(parms, DecSym(name),
		DecAny(ats...), DecAny(rts...)))
}
func (p PolyDef) Vector() VecVal { return p.Unbox().(VecVal) }
func (p PolyDef) Unbox() Functor { return Def(p).Unbox() }
func (p PolyDef) Type() Decl     { return Def(p).Type() }
func (p PolyDef) TypeId() Decl   { return Def(p).TypeId() }
func (p PolyDef) TypeRet() Decl  { return Def(p).TypeRet() }
func (p PolyDef) TypeArgs() Decl { return Def(p).TypeArgs() }
func (p PolyDef) String() string { return Def(p).TypeName() }
func (p PolyDef) TypeFnc() TyFnc {
	if IsPartial(p.Vector().Head()) {
		return Partial | Polymorph
	}
	return Polymorph
}
func (p PolyDef) TypeName() string {
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
func (p PolyDef) Call(args ...Functor) Functor {

	var partials = []Def{}

	for _, def := range p.Vector()() {
		var result = def.Call(args...)
		if !IsNone(result) {
			if !IsPartial(result) {
				return result
			}
			partials = append(partials,
				result.(Def))
		}
	}

	if len(partials) > 0 {
		if len(partials) > 1 {
			return DefinePolymorph(
				p.TypeId(),
				partials...)
		}
		return partials[0]
	}
	return NewNone()
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
func DefineOption(def Def) Option { return Option(def) }

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
func DefineAlternative(l, r Def) Altern {
	return Altern(DefinePolymorph(Alternatives, l, r))
}

func (a Altern) TypeFnc() TyFnc { return Alternatives }
func (a Altern) Unbox() Functor { return PolyDef(a).Unbox() }
func (a Altern) Type() Decl     { return PolyDef(a).Type() }
func (a Altern) TypeId() Decl   { return PolyDef(a).TypeId() }
func (a Altern) TypeArgs() Decl { return PolyDef(a).TypeArgs() }
func (a Altern) TypeRet() Decl {
	return Declare(Alternatives, PolyDef(a).TypeRet())
}
func (a Altern) TypeName() string             { return PolyDef(a).TypeName() }
func (a Altern) String() string               { return a.TypeName() }
func (a Altern) Call(args ...Functor) Functor { return PolyDef(a).Call(args...) }
