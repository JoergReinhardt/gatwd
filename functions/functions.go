/*
FUNCTION GENERALIZATION

lambda calculus states, that all functions can be expressed as functions
taking one argument, by currying in additional data and behaviour. all
computation can then be expressed in those terms‥. and while that's the base
of all that's done here, and generally considered to be a great thing, it
also turns out to be a pain in the behind, when applyed to a strongly typed
language on real world problems.

to get things done anyway, data types and function signatures need to be
generalized over, in a more reasonable way. data types of arguments and
return values already get generalized by the data package using type
aliasing and adding the flag method.

functions can be further discriminated by means of arity (number & type of
input arguments) and fixity (syntactical side, on which they expect to bind
there parameter(s)). golangs capability of returning multiple values, is of
no relevance in terms of functional programming, but very usefull in
imlementing a type system on top of it. so is the ability to define methods
on function types. functions in the terms of godeep are closures, closing
over arbitrary functions together with there arguments and return values,
btw. placeholders there of and an id/signature poir for typesystem and
runtime, to handle (partial} application and evaluation.

to deal with golang index operators and conrol structures, a couple of internal
function signatures, containing non aliased types (namely bool, int & string)
will also be made avaiable for enclosure.
*/
package functions

import (
	d "github.com/JoergReinhardt/godeep/data"
)

type Kind d.BitFlag

func (t Kind) Flag() d.BitFlag { return d.BitFlag(t).Flag() }
func (t Kind) Uint() uint      { return d.BitFlag(t).Uint() }

//go:generate stringer -type=Kind
const (
	Value Kind = 1 << iota
	Parameter
	Attribut // map key, slice index, search parameter...
	Accessor // pair of Attr & Value
	Double
	Vector
	Constant
	Unary
	Binary
	Nnary
	Tuple
	List
	Chain
	UniSet
	MuliSet
	AssocA
	Record
	Link
	DLink
	Node
	Tree
	BitFlag

	Recursives = Tuple | List
	Sets       = UniSet | MuliSet | AssocA | Record
	Links      = Link | DLink | Node | Tree // Consumeables
)

type ( // HIGHER ORDER FUNCTION TYPES
	// Arg
	// returns previously enclosed data and another Arg instance,
	// optionaly containing the passed data, if any was passed, or the
	// previous data again.
	Arg func(d ...Data) (Data, Arg)
	// AccAtt
	// shares the behaviour with that of a parameter, but yields and takes
	// a pair to contain a position/key & value pair instead.
	AccAtt func(d ...Pair) (Pair, AccAtt)
	// argset
	// set of placeholder arguments for signatures, promises, values passed
	// in a function call, partially applied values‥.
	ArgSet func(d ...AccAtt) ([]AccAtt, ArgSet)
	// RetVal
	// the return has the propertys of an arg set, but enclosed to be
	// addressable as a single value
	RetVal func() []AccAtt
	// Predicate
	// returns true, when the passed data meets the enclosed condition, a
	// native boolean for use in golang control structures
	Predic func(Data) bool
	// generic function wrapper
	Val   func() Data        // <- implements data.Typed
	Const func() Data        // <- guarantueed to allways evaluate identicly
	Pair  func() (a, b Data) // <- base element of all tuples and collections
	Vec   func() d.Sliceable // <- indexable native golang slice of data instances
	Tup   func() ([]Flag, Sliceable)
)

// closure that wraps instances of precedence types from data package
func Con(dat d.Data) Val        { return Val(func() Data { return dat.Eval() }) }
func (dat Val) Flag() d.BitFlag { return dat().Flag() }
func (dat Val) Type() Flag      { return conFlag(Constant.Flag(), dat().Flag()) }
func (dat Val) String() string  { return dat().(d.Data).String() }

// constant also conains immutable data, but it may be the result of a constant experssion
func ConConst(dat Data) Const   { return func() Data { return dat } }
func (c Const) Flag() d.BitFlag { return Constant.Flag() }
func (c Const) Type() Flag      { return conFlag(Constant.Flag(), c().Flag()) }
func (c Const) String() string  { return c().(d.Data).String() }

// pair encloses two data instances
func ConPair(l, r Data) Pair      { return func() (Data, Data) { return l, r } }
func (p Pair) Both() (Data, Data) { return p() }
func (p Pair) Left() Data         { l, _ := p(); return l }
func (p Pair) Right() Data        { _, r := p(); return r }
func (p Pair) Flag() d.BitFlag    { a, b := p(); return a.Flag() | b.Flag() }
func (p Pair) Type() Flag         { return conFlag(Double.Flag(), p.Flag()) }
func (p Pair) String() string     { l, r := p(); return l.String() + " " + r.String() }

// vector keeps a slice of data instances
func ConVec(dd ...d.Data) Vec {
	var ddd = []d.Data{}
	for _, dat := range dd {
		ddd = append(ddd, dat)
	}
	return func() d.Sliceable {
		return d.ChainToNativeSlice(d.ConChain(ddd...))
	}
}

// implements functions/sliceable interface
func (v Vec) Slice() []Data   { return sliceFunctionalize(v().(d.NativeVec).Slice()...) }
func (v Vec) Len() int        { return v().(d.NativeVec).Len() }
func (v Vec) Empty() bool     { return v().(d.NativeVec).Empty() }
func (v Vec) Flag() d.BitFlag { return v().Flag() }
func (v Vec) Type() Flag      { return conFlag(Vector.Flag(), v().Flag()) }
func (v Vec) String() string  { return v().String() }

// helper to type alias slices, initially initialized by the data package
func sliceFunctionalize(dd ...d.Data) []Data {
	var dat = []Data{}
	for _, ddd := range dd {
		dat = append(dat, Con(ddd.Eval()))
	}
	return dat
}

///////// PARAMETRIZATION //////////
// parameters can be retrieved, by calling the closure without passing
// parameters, or set, when parameters are indenet to be set
func ConParm(do Data) Arg {
	return func(di ...Data) (Data, Arg) {
		// if parameters where passed‥.
		if len(di) > 0 { // return former parameter‥.
			// ‥.and enclosure over newly passed parameters
			return di[0], ConParm(di[0])
		} //‥.otherwise, pass on unaltered results from last/first call
		return do, ConParm(do)
	}
}
func (p Arg) Param() Arg      { _, pa := p(); return pa }
func (p Arg) Data() Data      { d, _ := p(); return d }
func (p Arg) Arg() Data       { d, _ := p(); return d }
func (p Arg) Flag() d.BitFlag { d, _ := p(); return d.Flag() }
func (p Arg) Type() Flag      { d, _ := p(); return conFlag(Parameter.Flag(), d.Flag()) }
func (p Arg) String() string  { d, _ := p(); return "Parm: " + d.String() }

func ConAcc(do Pair) AccAtt {
	return func(di ...Pair) (Pair, AccAtt) {
		// if parameters where passed‥.
		if len(di) > 0 { // return former parameter‥.
			// ‥.and enclosure over newly passed parameters
			return di[0], ConAcc(di[0])
		} //‥.otherwise, pass on unaltered results from last/first call
		return do, ConAcc(do)
	}
}
func (p AccAtt) Param() AccAtt      { _, pa := p(); return pa }
func (p AccAtt) Data() Paired       { d, _ := p(); return d }
func (p AccAtt) Both() (Data, Data) { l, r := p.Data().Both(); return l, r }
func (p AccAtt) Acc() Data          { return p.Data().Left() }
func (p AccAtt) Left() Data         { return p.Data().Left() }
func (p AccAtt) Arg() Data          { return p.Data().Right() }
func (p AccAtt) Right() Data        { return p.Data().Right() }
func (p AccAtt) Flag() d.BitFlag    { d, _ := p(); return d.Flag() }
func (p AccAtt) Type() Flag         { d, _ := p(); return conFlag(Accessor.Flag(), d.Flag()) }
func (p AccAtt) String() string     { l, r := p.Both(); return l.String() + ": " + r.String() }

type (
	unary  func(Data) Data
	binary func(a, b Data) Data
	nary   func(...Data) Data
)
