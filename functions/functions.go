/*
FUNCTION GENERALIZATION

ambda calculus states, that all functions can be expressed as functions
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
	"strings"

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
	Internal

	AccIndex = Vector | Chain

	AccSymbol = Tuple | AssocA | Record

	AccCollect = AccIndex | AccSymbol

	Nests = Tuple | List

	Sets = UniSet | MuliSet | AssocA | Record

	Links = Link | DLink | Node | Tree // Consumeables
)

type ( // HIGHER ORDER FUNCTION TYPES
	// ARGUMENT
	// returns previously enclosed data and another argument instance,
	// optionaly containing the passed data, if any was passed, or the
	// previous data again.
	argument func(d ...Data) (Data, argument)
	// ARGSET
	// set of placeholder arguments for signatures, promises, values passed
	// in a function call, partially applied values‥.
	argSet func(d ...Data) ([]Data, argSet)
	// ACCESSATTRIBUT
	// shares the behaviour with that of a parameter, but yields and takes
	// a pair to contain a position/key & value pair instead.
	accessAttribut func(d ...Paired) (Paired, accessAttribut)
	// ACCSET
	accSet func(d ...Paired) ([]Paired, accSet)
	// returnValue
	// the return has the propertys of an arg set, but enclosed to be
	// addressable as a single value
	returnValue func() []accessAttribut
	// generic function wrapper
	value    func() Data        // <- implements data.Typed
	constant func() Data        // <- guarantueed to allways evaluate identicly
	pair     func() (a, b Data) // <- base element of all tuples and collections
	vector   func() d.Sliceable // <- indexable native golang slice of data instances
	tuple    func() (Data, Sliceable)
)

type (
	unary  func(Data) Data
	binary func(a, b Data) Data
	nary   func(...Data) Data
)

// DATA
// closure that wraps instances of precedence types from data package
func newData(dat d.Data) value    { return value(func() Data { return dat.(d.Evaluable).Eval() }) }
func (dat value) Flag() d.BitFlag { return dat().Flag() }
func (dat value) Type() Flag      { return newFlag(Constant, dat().Flag()) }
func (dat value) String() string  { return dat().(d.Data).String() }

// CONSTANT
// constant also conains immutable data, but it may be the result of a constant experssion
func newConstant(dat Data) constant { return func() Data { return dat } }
func (c constant) Flag() d.BitFlag  { return Constant.Flag() }
func (c constant) Type() Flag       { return newFlag(Constant, c().Flag()) }
func (c constant) String() string   { return c().(d.Data).String() }

// PAIR
// pair encloses two data instances
func newPair(l, r Data) pair      { return func() (Data, Data) { return l, r } }
func (p pair) Both() (Data, Data) { return p() }
func (p pair) Left() Data         { l, _ := p(); return l }
func (p pair) Right() Data        { _, r := p(); return r }
func (p pair) Acc() Data          { return p.Left() }
func (p pair) Arg() Data          { return p.Right() }
func (p pair) Flag() d.BitFlag    { a, b := p(); return a.Flag() | b.Flag() | Double.Flag() }
func (p pair) Type() Flag         { return newFlag(Double, p.Flag()) }
func (p pair) String() string     { l, r := p(); return l.String() + " " + r.String() }

// VECTOR
// vector keeps a slice of data instances
func newVector(dd ...d.Data) vector {
	return func() d.Sliceable {
		return d.NewChain(dd...)
	}
}

// implements functions/sliceable interface
func (v vector) Len() int        { return v().(d.NativeVec).Len() }
func (v vector) Empty() bool     { return v().(d.NativeVec).Empty() }
func (v vector) Flag() d.BitFlag { return v().Flag() | Vector.Flag() }
func (v vector) Type() Flag {
	return newFlag(Vector,
		d.Slice.Flag()|
			d.Parameter.Flag()|
			v().Flag())
}
func (v vector) String() string { return v().String() }
func (v vector) Slice() []d.Data {
	var vo = []d.Data{}
	for _, val := range v().Slice() {
		vo = append(vo, val)
	}
	return vo
}

/// PARAMETRIZATION
// parameters can be either retrieved, by calling the closure without passing
// parameters, or set when parameters are passed to be set.
//
// ARGUMENT
func newArgument(do Data) argument {
	return func(di ...Data) (Data, argument) {
		// if parameters where passed‥.
		if len(di) > 0 { // return former parameter‥.
			// ‥.and enclosure over newly passed parameters
			return di[0], newArgument(di[0])
		} //‥.otherwise, pass on unaltered results from last/first call
		return do, newArgument(do)
	}
}
func (p argument) String() string {
	d, _ := p()
	return d.Flag().String() +
		" " +
		d.String()
}
func (p argument) Data() Data         { d, _ := p(); return d }
func (p argument) Arg() Data          { return p.Data() }
func (p argument) Param() Data        { return p.Data() }
func (p argument) ParamType() BitFlag { return p.Data().Flag() }
func (p argument) DataType() BitFlag  { return p.Data().Flag() }
func (p argument) ArgType() BitFlag   { return p.Data().Flag() }
func (p argument) Type() Flag         { return newFlag(Attribut, p.Data().Flag()) }
func (p argument) Flag() d.BitFlag {
	return p.Data().Flag() |
		d.Argument.Flag() |
		d.Parameter.Flag()
}

// ARGUMENT SET
func newArgSet(args ...Data) argSet {
	return func(a ...Data) ([]Data, argSet) {
		if len(a) > 0 {
			return a, newArgSet(a...)
		}
		return args, newArgSet(args...)
	}
}
func (a argSet) String() string {
	var strdat = [][]d.Data{}
	for i, dat := range a.Args() {
		strdat = append(strdat, []d.Data{})
		strdat[i] = append(strdat[i], d.New(i), d.New(dat.String()))
	}
	return d.StringChainTable(strdat...)
}
func (a argSet) Type() Flag { return newFlag(Attribut, a.Flag()) }
func (a argSet) Flag() d.BitFlag {
	var f = d.BitFlag(uint(0))
	for _, arg := range a.Args() {
		f = f |
			arg.Flag() |
			d.Slice.Flag() |
			d.Argument.Flag() |
			d.Parameter.Flag()
	}
	return f
}
func (a argSet) Args() []Data   { d, _ := a(); return d }
func (a argSet) ArgSet() argSet { _, as := a(); return as }
func applyArgs(ao argSet, args ...argument) argSet {
	oargs, _ := ao()
	var l = len(oargs)
	if l < len(args) {
		l = len(args)
	}
	var an = make([]Data, 0, l)
	var i int
	for i, _ = range an {
		// copy old arguments to return set, if any are set at this pos.
		if oargs[i] != nil && d.FlagMatch(oargs[i].Flag(), d.Nil.Flag()) {
			an[i] = oargs[i]
		}
		// copy new arguments to return set, if any are set at this
		// position. overwrite old arguments in case any where set at
		// this position.
		if args[i] != nil && d.FlagMatch(args[i].Flag(), d.Nil.Flag()) {
			an[i] = args[i]
		}

	}
	return newArgSet(an...)
}

// ACCESSS ATTRIBUTE
func newAccAttribute(l, r Data) accessAttribut {
	return func(di ...Paired) (Paired, accessAttribut) {
		// if parameters where passed‥.
		if len(di) > 0 { // return former parameter‥.
			// ‥.and enclosure over newly passed parameters
			return di[0], newAccAttribute(di[0].Left(), di[0].Right())
		} //‥.otherwise, pass on unaltered results from last/first call
		return newPair(l, r), newAccAttribute(l, r)
	}
}
func (p accessAttribut) Param() accessAttribut { _, pa := p(); return pa }
func (p accessAttribut) Data() Paired          { d, _ := p(); return d }
func (p accessAttribut) Arg() Data             { return p.Data().Right() }
func (p accessAttribut) Acc() Data             { return p.Data().Left() }
func (p accessAttribut) Left() Data            { return p.Data().Left() }
func (p accessAttribut) Right() Data           { return p.Data().Right() }
func (p accessAttribut) Both() (Data, Data)    { return p.Left(), p.Right() }
func (p accessAttribut) AccType() d.BitFlag    { return p.Acc().Flag() }
func (p accessAttribut) Flag() d.BitFlag {
	dat, _ := p()
	return dat.Flag() |
		d.Slice.Flag() |
		d.Parameter.Flag() |
		Accessor.Flag()
}
func (p accessAttribut) Type() Flag {
	d, _ := p()
	return newFlag(Accessor, d.Flag())
}
func (p accessAttribut) String() string {
	l, r := p.Both()
	return l.String() + ": " + r.String()
}

// TUPLE
func (tup tuple) Flag() d.BitFlag {
	da, _ := tup()
	return da.Flag() |
		d.Parameter.Flag() |
		Accessor.Flag()
}
func (tup tuple) Type() Flag     { d, _ := tup(); return newFlag(Tuple, d.Flag()) }
func (tup tuple) String() string { d, c := tup(); return d.String() + " " + c.String() }

// ACCESS ATTRIBUTE SET
func newAccSet(accAttr ...Paired) accSet {
	return func(acc ...Paired) ([]Paired, accSet) {
		if len(acc) > 0 {
			return acc, newAccSet(acc...)
		}
		return accAttr, newAccSet(accAttr...)
	}
}
func (a accSet) String() string {
	var strout = [][]d.Data{}
	for i, pa := range a.Accs() {
		strout = append(strout, []d.Data{})
		strout[i] = append(
			strout[i],
			d.New(i),
			d.New(pa.Left().String()),
			d.New(pa.Right().String()))
	}
	return d.StringChainTable(strout...)
}
func (a accSet) Flag() d.BitFlag {
	var f = d.BitFlag(0)
	for _, acc := range a.Accs() {
		f = f | acc.Flag()
	}
	return f |
		d.Slice.Flag() |
		d.Parameter.Flag() |
		Accessor.Flag()
}
func (a accSet) Type() Flag                { return newFlag(AccCollect, a.Flag()) }
func (a accSet) Accs() []Paired            { acc, _ := a(); return acc }
func (a accSet) AccSet() accSet            { _, set := a(); return set }
func (a accSet) Append(v ...Paired) accSet { return newAccSet(append(a.Accs(), v...)...) }

// pair sorter has the methods to search for a pair in-/, and sort slices of
// pairs. pairs will be sorted by the left parameter, since it references the
// accessor (key) in an accessor/value pair.
type pairSorter []Paired

func newPairSorte() pairSorter     { return []Paired{} }
func (p pairSorter) Len() int      { return len(p) }
func (p pairSorter) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p pairSorter) Sort(acc Paired) {
	less := newAccLess(p, acc)
	sort.Slice(p, less)
}
func (p pairSorter) Search(acc Paired) Paired {
	search := newAccSearch(p, acc)
	var idx = sort.Search(len(p), search)
	return p[idx]
}

func newAccLess(accs pairSorter, acc Paired) func(i, j int) bool {
	f := acc.Flag()
	switch {
	case d.FlagMatch(d.String.Flag(), f):
		return func(i, j int) bool {
			chain := accs
			if strings.Compare(
				string(chain[i].(Accessable).Acc().String()),
				string(chain[j].(Accessable).Acc().String()),
			) <= 0 {
				return true
			}
			return false
		}
	case d.FlagMatch(d.Flag.Flag(), f):
		return func(i, j int) bool { // sort by value-, NOT accessor type
			chain := accs
			if chain[i].(Accessable).Arg().Flag() <
				chain[j].(Accessable).Arg().Flag() {
				return true
			}
			return false
		}
	case d.FlagMatch(d.Unsigned.Flag(), f):
		return func(i, j int) bool {
			chain := accs
			if uint(chain[i].(Accessable).Acc().(Unsigned).Uint()) <
				uint(chain[j].(Accessable).Acc().(Unsigned).Uint()) {
				return true
			}
			return false
		}
	case d.FlagMatch(d.Integer.Flag(), f):
		return func(i, j int) bool {
			chain := accs
			if int(chain[i].(Accessable).Acc().(Integer).Int()) <
				int(chain[j].(Accessable).Acc().(Integer).Int()) {
				return true
			}
			return false
		}
	}
	return nil
}
func newAccSearch(accs pairSorter, acc Paired) func(i int) bool {
	var fn func(i int) bool
	f := acc.Flag()
	switch { // parameters are accessor/value pairs to be applyed.
	case d.FlagMatch(d.Unsigned.Flag(), f):
		fn = func(i int) bool {
			return uint(accs[i].(Accessable).Acc().(Unsigned).Uint()) >=
				uint(acc.(Accessable).Acc().(Unsigned).Uint())
		}
	case d.FlagMatch(d.Integer.Flag(), f):
		fn = func(i int) bool {
			return int(accs[i].(Accessable).Acc().(Integer).Int()) >=
				int(acc.(Accessable).Acc().(Integer).Int())
		}
	case d.FlagMatch(d.String.Flag(), f):
		fn = func(i int) bool {
			return strings.Compare(
				accs[i].(Accessable).Acc().String(),
				acc.(Accessable).Acc().String()) >= 0
		}
	case d.FlagMatch(d.Flag.Flag(), f):
		fn = func(i int) bool {
			return accs[i].(Accessable).Acc().Flag() >=
				acc.(Accessable).Acc().Flag()
		}
	}
	return fn
}
