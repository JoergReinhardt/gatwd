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
	"sort"

	d "github.com/JoergReinhardt/godeep/data"
)

type DataType d.BitFlag

func (t DataType) Flag() d.BitFlag { return d.BitFlag(t).Flag() }
func (t DataType) Uint() uint      { return d.BitFlag(t).Uint() }

//go:generate stringer -type=DataType
const (
	Pair DataType = 1 << iota
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

type (
	// HIGHER ORDER FUNCTION TYPES
	data     func() d.Data      // <- implements data.Typed
	constant func() Data        // <- guarantueed to allways evaluate identicly
	pair     func() (a, b Data) // <- base element of all tuples and collections
	vector   func() d.Sliceable // <- indexable native golang slice of data instances
	unary    func(d Data) Data
	binary   func(a, b Data) Data
	nary     func(...Data) Data

	// parameter
	// returns previously enclosed data and another parameter instance,
	// optionaly containing the passed data, if any was passed, or the
	// previous data again.
	parameter func(Data) (Data, parameter)

	// applicative
	// parameter that contains index/key & value pair to be applyed as
	// positional, or named parameter, argument, or result of an operation
	// involving a accessable collection.
	applicative func(pair) (pair, applicative)

	// Generator
	// returns data and another generator instance. can either represent
	// endless lists, streams and the like, or consumeable data structures
	// that implement methods to be reduced on a per element basis
	generator func() (Data, generator)

	// Predicate
	//
	// returns true, when the passed data meets the enclosed condition, a
	// native boolean for use in golang control structures
	predicate func(Data) bool
)

// flag modules data closure can contain either data module instances, or
// instances of the types of the function module itself. flag get's converted
// into function module flag out of precaution.
func (dat data) Flag() d.BitFlag { return dat().Flag() }
func (dat data) Type() Flag      { return conFlag(Constant.Flag(), dat().Flag()) }
func (dat data) Eval() Data      { return dat }
func con(dat d.Data) data        { return data(func() d.Data { return dat.Eval() }) }

func (c constant) Flag() d.BitFlag { return Constant.Flag() }
func (c constant) Prec() d.BitFlag { return c().Flag() }
func (c constant) Type() Flag      { return conFlag(Constant.Flag(), c().Flag()) }
func (c constant) Eval() Data      { return c() }
func conConst(dat Data) constant   { return func() Data { return dat } }

// pair encloses two data instances
func (p pair) Flag() d.BitFlag { a, b := p(); return a.Flag() | b.Flag() }
func (p pair) Type() Flag      { return conFlag(Pair.Flag(), p.Flag()) }
func (p pair) Eval() Data      { return p }
func conPair(a, b Data) pair   { return func() (a, b Data) { return a, b } }

func (v vector) Flag() d.BitFlag { return v().Flag() }
func (v vector) Slice() []Data   { return sliceSanitize(v().(d.Sliceable).Slice()...) }
func (v vector) Type() Flag      { return conFlag(Vector.Flag(), v.Flag()) }
func conVector(dd ...d.Data) vector {
	var ddd = []d.Data{}
	for _, dat := range dd {
		ddd = append(ddd, dat)
	}
	return func() d.Sliceable {
		return d.ConChain(ddd...)
	}
}
func sliceSanitize(dd ...d.Data) []Data {
	var dat = []Data{}
	for _, ddd := range dd {
		dat = append(dat, con(ddd.Eval()))
	}
	return dat
}

///////// POLYMORPHISM ///////////
type (
	signature func() (id int, tok tokens)
	derivate  func() (id int, from int, tok tokens)
	isomorph  func() (id int, tok tokens, fnc Function)
	polymorph func() (id int, tok tokens, iso isomorphs)
	namedPoly func() (id int, name string, sig tokens, iso isomorphs)
)

func (s signature) Id() int         { id, _ := s(); return id }
func (s signature) Tokens() []Token { _, tok := s(); return tok }
func (d derivate) Id() int          { id, _, _ := d(); return id }
func (d derivate) DerivedFrom() int { _, id, _ := d(); return id }
func (d derivate) Tokens() []Token  { _, _, tok := d(); return tok }
func (i isomorph) Id() int          { id, _, _ := i(); return id }
func (i isomorph) Tokens() []Token  { _, tok, _ := i(); return tok }
func (p polymorph) Id() int         { id, _, _ := p(); return id }
func (p polymorph) Tokens() []Token { _, tok, _ := p(); return tok }
func (n namedPoly) Id() int         { id, _, _, _ := n(); return id }
func (n namedPoly) Name() string    { _, name, _, _ := n(); return name }
func (n namedPoly) Tokens() []Token { _, _, tok, _ := n(); return tok }

// isomorphic functions implement the function interface by forwarding passed
// parameters to the embedded functions eval method. TODO: handle arguments and returns
func (i isomorph) Call(d ...data) data { _, _, fn := i(); return fn.Call(d...) }

func conSignature(tok ...Token) signature {
	i := conUID()
	s := tok
	return func() (id int, sig tokens) {
		return i, s
	}
}
func conIsomorph(sig signature, fnc Function) isomorph {
	s := sig
	f := fnc
	return func() (
		id int,
		tok tokens,
		fn Function,
	) {
		id, tok = s()
		return id, tok, f
	}
}
func conPolymorph(sig signature, iso ...isomorph) polymorph {
	s := sig
	return func() (
		id int,
		tok tokens,
		iso isomorphs,
	) {
		id, tok = s()
		return id, tok, iso
	}
}
func conNamedDef(name string, pol polymorph) namedPoly {
	p := pol
	return func() (
		id int,
		name string,
		tok tokens,
		iso isomorphs,
	) {
		id, tok, iso = p()
		return id, name, tok, iso
	}
}

type signatures []signature

func (s signatures) Len() int           { return len(s) }
func (s signatures) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s signatures) Less(i, j int) bool { return s[i].Id() < s[j].Id() }
func (s signatures) hasId(id int) bool  { return s.getById(id).Id() == id }
func (s signatures) getById(id int) signature {
	var sig = s[sort.Search(len(s),
		func(i int) bool {
			return s[i].Id() >= id
		})]
	if sig.Id() == id {
		return sig
	}
	return sig
}
func sortSignatures(s signatures) signatures { sort.Sort(s); return s }

type isomorphs []isomorph

func (m isomorphs) Len() int           { return len(m) }
func (m isomorphs) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m isomorphs) Less(i, j int) bool { return m[i].Id() < m[j].Id() }
func (m isomorphs) hasId(id int) bool  { return m.getById(id).Id() == id }
func (m isomorphs) getById(id int) isomorph {
	var iso = m[sort.Search(len(m),
		func(i int) bool {
			return m[i].Id() >= id
		})]
	if iso.Id() == id {
		return iso
	}
	return iso
}
func sortIsomorphs(m isomorphs) isomorphs { sort.Sort(m); return m }

type polymorphs []polymorph

func (p polymorphs) Len() int           { return len(p) }
func (p polymorphs) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p polymorphs) Less(i, j int) bool { return p[i].Id() < p[j].Id() }
func (m polymorphs) hasId(id int) bool  { return m.getById(id).Id() == id }
func (m polymorphs) getById(id int) polymorph {
	var poly = m[sort.Search(len(m),
		func(i int) bool {
			return m[i].Id() >= id
		})]
	if poly.Id() == id {
		return poly
	}
	return poly
}
func sortPolymorphs(p polymorphs) polymorphs { sort.Sort(p); return p }
