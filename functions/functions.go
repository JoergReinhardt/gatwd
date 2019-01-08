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
	Pair Kind = 1 << iota
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

	Recursives = Tuple | List
	Sets       = UniSet | MuliSet | AssocA | Record
	Links      = Link | DLink | Node | Tree // Consumeables
)

type ( // HIGHER ORDER FUNCTION TYPES
	// parameter
	// returns previously enclosed data and another parameter instance,
	// optionaly containing the passed data, if any was passed, or the
	// previous data again.
	parameter func(d ...Data) (Data, parameter)
	// Predicate
	// returns true, when the passed data meets the enclosed condition, a
	// native boolean for use in golang control structures
	predicate func(Data) bool
	// generic function wrapper
	data     func() Data        // <- implements data.Typed
	constant func() Data        // <- guarantueed to allways evaluate identicly
	pair     func() (a, b Data) // <- base element of all tuples and collections
	vector   func() d.Sliceable // <- indexable native golang slice of data instances
)

// parameters can be retrieved, by calling the closure without passing
// parameters, or set, when parameters are indenet to be set
func conParam(do Data) parameter {
	return func(di ...Data) (Data, parameter) {
		// if parameters where passed‥.
		if len(di) > 0 { // return former parameter‥.
			// ‥.and enclosure over newly passed parameters
			return di[0], conParam(di[0])
		} //‥.otherwise, pass on unaltered results from last/first call
		return do, conParam(do)
	}
}
func (p parameter) Flag() d.BitFlag { d, p := p(); return d.Flag() }
func (p parameter) Type() Flag      { d, _ := p(); return conFlag(Constant.Flag(), d.Flag()) }

// closure that wraps instances of precedence types from data package
func con(dat d.Data) data        { return data(func() Data { return dat.Eval() }) }
func (dat data) Flag() d.BitFlag { return dat().Flag() }
func (dat data) Type() Flag      { return conFlag(Constant.Flag(), dat().Flag()) }
func (dat data) String() string  { return dat().(d.Data).String() }

// constant also conains immutable data, but it may be the result of a constant experssion
func conConst(dat Data) constant   { return func() Data { return dat } }
func (c constant) Flag() d.BitFlag { return Constant.Flag() }
func (c constant) Type() Flag      { return conFlag(Constant.Flag(), c().Flag()) }

// pair encloses two data instances
func conPair(l, r Data) pair   { return func() (Data, Data) { return l, r } }
func (p pair) Flag() d.BitFlag { a, b := p(); return a.Flag() | b.Flag() }
func (p pair) Type() Flag      { return conFlag(Pair.Flag(), p.Flag()) }
func (p pair) String() string  { l, r := p(); return l.String() + ": " + r.String() }

// vector keeps a slice of data instances
func conVector(dd ...d.Data) vector {
	var ddd = []d.Data{}
	for _, dat := range dd {
		ddd = append(ddd, dat)
	}
	return func() d.Sliceable {
		return d.ChainToNativeSlice(d.ConChain(ddd...))
	}
}
func (v vector) Flag() d.BitFlag { return v().Flag() }
func (v vector) Slice() []Data   { return sliceFunctionalize(v().(d.Sliceable).Slice()...) }
func (v vector) Type() Flag      { return conFlag(Vector.Flag(), v.Flag()) }
func (v vector) String() string  { return v().String() }

// helper to type alias slices, initially initialized by the data package
func sliceFunctionalize(dd ...d.Data) []Data {
	var dat = []Data{}
	for _, ddd := range dd {
		dat = append(dat, con(ddd.Eval()))
	}
	return dat
}

///////// PARAMETRIZATION //////////
type (
	unary  func(Data) Data
	binary func(a, b Data) Data
	nary   func(...Data) Data
)
