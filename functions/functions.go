/*
BASE FUNCTIONS ARGUMENTS & ACCESSABLE PRAEDICATES

  base functional data types to wrap data instances, pairs thereof as well as
  arguments and accessors as datastructure intended to pass data between
  function calls, and assist in handling od currying, partial application‥., of
  higher order functions and the like.
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
	UnaryFnc
	BinaryFnc
	NaryFnc
	Tuple
	List
	UniSet
	MuliSet
	AssocA
	Record
	Link
	DLink
	Node
	Tree
	HigherOrder

	Chain = Vector | Tuple | Record

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
	argument func(d ...Argumented) (Data, Argumented)
	// ARGUMENTS
	arguments func(d ...Data) ([]Data, Arguments)
	// a pair to contain a position/key & value pair instead.
	praedicate func(d ...Paired) (Paired, Parametric)
	// ACCSET
	praedicates func(d ...Paired) ([]Paired, Praedicates)
	// generic functional data wrapper
	value func() Data // <- implements data.Typed
	// wraps generic pairs of functional data
	pair func() (a, b Data) // <- base element of all tuples and collections
)

// DATA
// closure that wraps instances of precedence types from data package
// ARGSET
// set of placeholder arguments for signatures, promises, values passed
// in a function call, partially applied values‥.
// ACCESSATTRIBUT
// shares the behaviour with that of a parameter, but yields and takes
func newData(dat d.Data) Data     { return value(func() Data { return dat.(d.Evaluable).Eval() }) }
func (dat value) Flag() d.BitFlag { return dat().Flag() }
func (dat value) String() string  { return dat().(d.Data).String() }
func (dat value) Ident() Data     { return dat }
func (dat value) Empty() bool     { return elemEmpty(dat) }
func elemEmpty(dat Data) bool {
	if dat != nil {
		if !dat.Flag().Match(d.Nil.Flag()) {
			return false
		}
	}
	return true
}

// PAIR
// pair encloses two data instances
func newPair(l, r Data) Paired    { return pair(func() (Data, Data) { return l, r }) }
func (p pair) Both() (Data, Data) { return p() }
func (p pair) Left() Data         { l, _ := p(); return l }
func (p pair) Right() Data        { _, r := p(); return r }
func (p pair) Acc() Parametric    { return newPraedicate(newPair(p.Left(), p.Right())) }
func (p pair) Arg() Argumented    { return newArgument(p.Right()) }
func (p pair) Flag() d.BitFlag    { a, b := p(); return a.Flag() | b.Flag() | Double.Flag() }
func (p pair) String() string     { l, r := p(); return l.String() + " " + r.String() }
func (p pair) Ident() Data        { return p }
func (p pair) Empty() bool {
	return elemEmpty(p.Left()) && elemEmpty(p.Right())
}

/// PARAMETRIZATION
// parameters can be either retrieved, by calling the closure without passing
// parameters, or set when parameters are passed to be set.
//
// ARGUMENT
func newArgument(do ...Data) Argumented {
	return argument(func(di ...Argumented) (Data, Argumented) {
		// if parameters where passed‥.
		if len(di) > 0 { // return former parameter‥.
			// ‥.and enclosure over newly passed parameters
			return di[0], newArgument(di[0])
		} //‥.otherwise, pass on unaltered results from last/first call
		return do[0], newArgument(do[0])
	})
}
func (p argument) String() string {
	d, _ := p()
	return d.Flag().String() +
		" " +
		d.String()
}
func (p argument) Apply(d ...Data) (Data, Argumented) {
	if len(d) > 0 {
		return d[0], newArgument(d...)
	}
	return p()
}
func (p argument) Data() Data         { d, _ := p(); return d }
func (p argument) Ident() Data        { return p }
func (p argument) Arg() Argumented    { return newArgument(p.Data()) }
func (p argument) Param() Data        { return p.Data() }
func (p argument) ParamType() BitFlag { return p.Data().Flag() }
func (p argument) DataType() BitFlag  { return p.Data().Flag() }
func (p argument) ArgType() BitFlag   { return p.Data().Flag() }
func (p argument) Empty() bool        { return elemEmpty(p.Data()) }
func (p argument) Flag() d.BitFlag {
	return p.Data().Flag() |
		d.Argument.Flag() |
		d.Parameter.Flag()
}

// ARGUMENT SET
func newArguments(dat ...Data) Arguments {
	return arguments(func(dot ...Data) ([]Data, Arguments) {
		if len(dot) > 0 {
			return dot, newArguments(dot...)
		}
		return dat, newArguments(dat...)
	})
}
func newArgSet(dat ...Data) Arguments {
	return arguments(func(dot ...Data) ([]Data, Arguments) {
		return dat,
			arguments(
				func(...Data) ([]Data, Arguments) {
					return dat, newArguments(dat...)
				})

	})
}
func (a arguments) String() string {
	var strdat = [][]d.Data{}
	for i, dat := range a.Args() {
		strdat = append(strdat, []d.Data{})
		strdat[i] = append(strdat[i], d.New(i), d.New(dat.String()))
	}
	return d.StringChainTable(strdat...)
}
func (a arguments) Flag() d.BitFlag {
	var f = d.BitFlag(uint(0))
	for _, arg := range a.Args() {
		f = f |
			arg.Flag() |
			d.Vector.Flag() |
			d.Argument.Flag() |
			d.Parameter.Flag()
	}
	return f
}
func (a arguments) Args() []Argumented {
	var args = []Argumented{}
	for _, arg := range a.Data() {
		args = append(args, newArgument(arg))
	}
	return args
}
func (a arguments) Data() []Data { d, _ := a(); return d }
func (a arguments) Len() int     { d, _ := a(); return len(d) }
func (a arguments) Empty() bool {
	if len(a.Args()) > 0 {
		for _, arg := range a.Args() {
			if !elemEmpty(arg.Data()) {
				return false
			}
		}
	}
	return true
}
func (a arguments) ArgSet() Arguments      { _, as := a(); return as }
func (a arguments) Ident() Data            { return a }
func (a arguments) Get(idx int) Argumented { return a.Args()[idx] }
func (a arguments) Replace(idx int, arg Data) Arguments {
	dats, _ := a()
	dats[idx] = arg
	return newArguments(dats...)
}
func (a arguments) Apply(d ...Data) ([]Data, Arguments) {
	var dats = []Data{}
	var args = a.ArgSet()
	for i, dat := range d {
		dats = append(dats, dat)
		args = args.Replace(i, newArgument(dat))
	}
	return dats, args
}
func applyArgs(ao arguments, args ...Data) Arguments {
	oargs, _ := ao()
	var l = len(oargs)
	if l < len(args) {
		l = len(args)
	}
	var an = make([]Data, 0, l)
	var i int
	for i, _ = range oargs {
		// copy old arguments to return set, if any are set at this pos.
		if oargs[i] != nil && d.Nil.Flag().Match(oargs[i].Flag()) {
			an[i] = oargs[i]
		}
		// copy new arguments to return set, if any are set at this
		// position. overwrite old arguments in case any where set at
		// this position.
		if args[i] != nil && d.Nil.Flag().Match(args[i].Flag()) {
			an[i] = args[i]
		}

	}
	return newArguments(an...)
}

// ACCESSS ATTRIBUTE
func newPraedicate(d ...Paired) Parametric {
	return praedicate(func(di ...Paired) (Paired, Parametric) {
		// if parameters where passed‥.
		if len(di) > 0 { // return former parameter‥.
			// ‥.and enclosure over newly passed parameters
			return di[0], newPraedicate(newPair(di[0].Left(), di[0].Right()))
		} //‥.otherwise, pass on unaltered results from last/first call
		return newPair(d[0].Left(), d[0].Right()),
			newPraedicate(newPair(d[0].Left(), d[0].Right()))
	})
}
func (p praedicate) Apply(pa ...Paired) (Paired, Parametric) {
	if len(pa) > 0 {
		return pa[0], newPraedicate(pa...)
	}
	return p()
}
func (p praedicate) Arg() Argumented    { return newArgument(p.Pair().Right()) }
func (p praedicate) Ident() Data        { return p }
func (p praedicate) Acc() Parametric    { _, acc := p(); return acc }
func (p praedicate) Pair() Paired       { pa, _ := p(); return pa }
func (p praedicate) Key() Data          { return p.Pair().Left() }
func (p praedicate) Data() Data         { return p.Pair().Right() }
func (p praedicate) Left() Data         { return p.Pair().Left() }
func (p praedicate) Right() Data        { return p.Pair().Right() }
func (p praedicate) Both() (Data, Data) { return p.Pair().Both() }
func (p praedicate) Empty() bool {
	l, r := p.Pair().Both()
	return elemEmpty(l) && elemEmpty(r)
}
func (p praedicate) AccType() d.BitFlag { return p.Key().Flag() }
func (p praedicate) Flag() d.BitFlag {
	dat, _ := p()
	return dat.Flag() |
		d.Vector.Flag() |
		d.Parameter.Flag() |
		Accessor.Flag()
}
func (p praedicate) String() string {
	l, r := p.Both()
	return l.String() + ": " + r.String()
}

// ACCESS ATTRIBUTE SET
func newPraedicateSet(pairs ...Paired) Praedicates {
	return praedicates(func(pairs ...Paired) ([]Paired, Praedicates) {
		return pairs, praedicates(func(...Paired) ([]Paired, Praedicates) {
			return pairs, newPraedicateSet(pairs...)
		})

	})
}
func newPraedicates(pairs ...Paired) Praedicates {
	return praedicates(
		func(po ...Paired) ([]Paired, Praedicates) {
			if len(po) > 0 {
				return po, newPraedicates(po...)
			}
			return pairs, newPraedicates(pairs...)
		})
}
func (a praedicates) getIdx(acc Data) (int, pairSorter) {
	var ps = newPairSorter(a.Pairs()...)
	switch {
	case acc.Flag().Match(d.Symbolic.Flag()):
		ps.Sort(d.String.Flag())
	case acc.Flag().Match(d.Unsigned.Flag()):
		ps.Sort(d.Unsigned.Flag())
	case acc.Flag().Match(d.Integer.Flag()):
		ps.Sort(d.Unsigned.Flag())
	}
	return ps.Search(acc), ps
}
func (a praedicates) Get(acc Data) Paired {
	var idx, ps = a.getIdx(acc)
	if idx >= 0 {
		return ps[idx]
	}
	return nil
}
func (a praedicates) Replace(acc Paired) Praedicates {
	idx, ps := a.getIdx(acc.Left())
	ps[idx] = acc
	return newPraedicates(ps...)
}
func (a praedicates) Apply(acc ...Paired) ([]Paired, Praedicates) {
	if len(acc) > 0 {
		return acc, newPraedicates(acc...)
	}
	return a()
}
func (a praedicates) String() string {
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
func (a praedicates) Flag() d.BitFlag {
	var f = d.BitFlag(0)
	for _, acc := range a.Accs() {
		f = f | acc.Flag()
	}
	return f |
		d.Vector.Flag() |
		d.Parameter.Flag() |
		Accessor.Flag()
}
func (a praedicates) Accs() (accs []Parametric) {
	pairs, _ := a()
	for _, p := range pairs {
		accs = append(accs, newPraedicate(p))
	}
	return accs
}
func (a praedicates) Pairs() []Paired { pairs, _ := a(); return pairs }
func (a praedicates) Len() int        { pairs, _ := a(); return len(pairs) }
func (a praedicates) Empty() bool {
	if len(a.Pairs()) > 0 {
		for _, p := range a.Pairs() {
			if !elemEmpty(p) {
				return false
			}
		}
	}
	return true
}
func (a praedicates) AccSet() Praedicates { _, set := a(); return set }
func (a praedicates) Ident() Data         { return a }
func (a praedicates) Append(v ...Paired) Praedicates {
	return newPraedicates(append(a.Pairs(), v...)...)
}

// pair sorter has the methods to search for a pair in-/, and sort slices of
// pairs. pairs will be sorted by the left parameter, since it references the
// accessor (key) in an accessor/value pair.
type pairSorter []Paired

func newPairSorter(p ...Paired) pairSorter { return append(pairSorter{}, p...) }
func (a pairSorter) Empty() bool {
	if len(a) > 0 {
		for _, p := range a {
			if !elemEmpty(p) {
				return false
			}
		}
	}
	return true
}
func (p pairSorter) Len() int      { return len(p) }
func (p pairSorter) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p pairSorter) Sort(f d.BitFlag) {
	less := newPraedLess(p, f)
	sort.Slice(p, less)
}
func (p pairSorter) Search(praed Data) int {
	var idx = sort.Search(len(p), newFindPraedicate(p, praed))
	// when praedicate is a precedence type encoding bit-flag
	if praed.Flag().Match(d.Flag.Flag()) {
		if d.Type(p[idx].Left().Flag()) == praed {
			return idx
		}
	}
	// otherwise check if key is equal to praedicate
	if idx < len(p) {
		if p[idx].Left() == praed {
			return idx
		}
	}
	return -1
}

func newPraedLess(accs pairSorter, f d.BitFlag) func(i, j int) bool {
	switch {
	case f.Match(d.String.Flag()):
		return func(i, j int) bool {
			chain := accs
			if strings.Compare(
				string(chain[i].(Paired).Left().String()),
				string(chain[j].(Paired).Left().String()),
			) <= 0 {
				return true
			}
			return false
		}
	case f.Match(d.Flag.Flag()):
		return func(i, j int) bool { // sort by value-, NOT accessor type
			chain := accs
			if chain[i].(Paired).Right().Flag() <
				chain[j].(Paired).Right().Flag() {
				return true
			}
			return false
		}
	case f.Match(d.Unsigned.Flag()):
		return func(i, j int) bool {
			chain := accs
			if uint(chain[i].(Paired).Left().(Unsigned).Uint()) <
				uint(chain[i].(Paired).Left().(Unsigned).Uint()) {
				return true
			}
			return false
		}
	case f.Match(d.Integer.Flag()):
		return func(i, j int) bool {
			chain := accs
			if int(chain[i].(Paired).Left().(Integer).Int()) <
				int(chain[i].(Paired).Left().(Integer).Int()) {
				return true
			}
			return false
		}
	}
	return nil
}
func newFindPraedicate(accs pairSorter, praed Data) func(i int) bool {
	var f = praed.Flag()
	var fn func(i int) bool
	switch { // parameters are accessor/value pairs to be applyed.
	case f.Match(d.Unsigned.Flag()):
		fn = func(i int) bool {
			return uint(accs[i].(Paired).Left().(Unsigned).Uint()) >=
				uint(praed.(Unsigned).Uint())
		}
	case f.Match(d.Integer.Flag()):
		fn = func(i int) bool {
			return int(accs[i].(Paired).Left().(Integer).Int()) >=
				int(praed.(Integer).Int())
		}
	case f.Match(d.String.Flag()):
		fn = func(i int) bool {
			return strings.Compare(
				accs[i].(Paired).Left().String(),
				praed.String()) >= 0
		}
	case f.Match(d.Flag.Flag()):
		fn = func(i int) bool {
			return accs[i].(Paired).Right().Flag() >=
				praed.(d.BitFlag)
		}
	}
	return fn
}
func applyPraedicates(acc Praedicates, praed ...Paired) Praedicates {
	var ps = newPairSorter(acc.Pairs()...)
	ps.Sort(praed[0].Left().Flag())
	for _, p := range praed {
		idx := ps.Search(p.Left())
		if idx >= 0 {
			ps[idx] = p
			continue
		}
		ps = append(ps, p)
	}
	return newPraedicates(ps...)
}
