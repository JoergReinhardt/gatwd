/*
BASE FUNCTIONS ARGUMENTS & ACCESSABLE PRAEDICATES
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
	// returns previously enclosed data and another Args instance,
	// optionaly containing the passed data, if any was passed, or the
	// previous data again.
	Args func(d ...Argumented) (Function, Argumented)
	// ARGUMENTS
	ArgSet func(d ...Function) ([]Function, Arguments)
	// a pair to contain a position/key & value pair instead.
	Param func(d ...Paired) (Paired, Parametric)
	// ACCSET
	ParamSet func(d ...Paired) ([]Paired, Parameters)
	// wraps generic pairs of functional data
	Pair func() (a, b Function) // <- base element of all tuples and collections
)

// DATA
// closure that wraps instances of precedence types from data package
// ARGSET
// set of placeholder arguments for signatures, promises, values passed
// in a function call, partially applied values‥.
// ACCESSATTRIBUT
// shares the behaviour with that of a parameter, but yields and takes
func NewFncData(dat Function) Function { return ConstFnc(func() Function { return dat }) }
func ElemEmpty(dat Function) bool {
	if dat != nil {
		if !dat.Flag().Match(d.Nil.Flag()) {
			return false
		}
	}
	return true
}

// PAIR
// pair encloses two data instances
func NewPair(l, r Function) Paired        { return Pair(func() (Function, Function) { return l, r }) }
func (p Pair) Both() (Function, Function) { return p() }
func (p Pair) Left() Function             { l, _ := p(); return l }
func (p Pair) Right() Function            { _, r := p(); return r }
func (p Pair) Acc() Parametric            { return NewParameter(NewPair(p.Left(), p.Right())) }
func (p Pair) Arg() Argumented            { return NewArgument(p.Right()) }
func (p Pair) Flag() d.BitFlag            { a, b := p(); return a.Flag() | b.Flag() | Double.Flag() }
func (p Pair) Ident() Function            { return p }
func (p Pair) Empty() bool {
	return ElemEmpty(p.Left()) && ElemEmpty(p.Right())
}

/// Parametric
// parameters can be either retrieved, by calling the closure without passing
// parameters, or set when parameters are passed to be set.
//
// ARGUMENT
func NewArgument(do ...Function) Argumented {
	return Args(func(di ...Argumented) (Function, Argumented) {
		// if parameters where passed‥.
		if len(di) > 0 { // return former parameter‥.
			// ‥.and enclosure over newly passed parameters
			return di[0], NewArgument(di[0])
		} //‥.otherwise, pass on unaltered results from last/first call
		return do[0], NewArgument(do[0])
	})
}
func (p Args) Apply(d ...Function) (Function, Argumented) {
	if len(d) > 0 {
		return d[0], NewArgument(d...)
	}
	return p()
}
func (p Args) Data() Function     { d, _ := p(); return d }
func (p Args) Ident() Function    { return p }
func (p Args) Arg() Argumented    { return NewArgument(p.Data()) }
func (p Args) Param() Function    { return p.Data() }
func (p Args) ParamType() BitFlag { return p.Data().Flag() }
func (p Args) DataType() BitFlag  { return p.Data().Flag() }
func (p Args) ArgType() BitFlag   { return p.Data().Flag() }
func (p Args) Empty() bool        { return ElemEmpty(p.Data()) }
func (p Args) Flag() d.BitFlag {
	return p.Data().Flag() |
		d.Argument.Flag() |
		d.Parameter.Flag()
}

// ARGUMENT SET
func NewwArguments(dat ...Function) Arguments {
	return ArgSet(func(dot ...Function) ([]Function, Arguments) {
		if len(dot) > 0 {
			return dot, NewwArguments(dot...)
		}
		return dat, NewwArguments(dat...)
	})
}
func NewArgSet(dat ...Function) Arguments {
	return ArgSet(func(dot ...Function) ([]Function, Arguments) {
		return dat,
			ArgSet(
				func(...Function) ([]Function, Arguments) {
					return dat, NewwArguments(dat...)
				})

	})
}
func (a ArgSet) Flag() d.BitFlag {
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
func (a ArgSet) Args() []Argumented {
	var args = []Argumented{}
	for _, arg := range a.Data() {
		args = append(args, NewArgument(arg))
	}
	return args
}
func (a ArgSet) Data() []Function { d, _ := a(); return d }
func (a ArgSet) Len() int         { d, _ := a(); return len(d) }
func (a ArgSet) Empty() bool {
	if len(a.Args()) > 0 {
		for _, arg := range a.Args() {
			if !ElemEmpty(arg.Data()) {
				return false
			}
		}
	}
	return true
}
func (a ArgSet) ArgSet() Arguments      { _, as := a(); return as }
func (a ArgSet) Ident() Function        { return a }
func (a ArgSet) Get(idx int) Argumented { return a.Args()[idx] }
func (a ArgSet) Replace(idx int, arg Function) Arguments {
	dats, _ := a()
	dats[idx] = arg
	return NewwArguments(dats...)
}
func (a ArgSet) Apply(d ...Function) ([]Function, Arguments) {
	var dats = []Function{}
	var args = a.ArgSet()
	for i, dat := range d {
		dats = append(dats, dat)
		args = args.Replace(i, NewArgument(dat))
	}
	return dats, args
}
func ApplyArgs(ao ArgSet, args ...Function) Arguments {
	oargs, _ := ao()
	var l = len(oargs)
	if l < len(args) {
		l = len(args)
	}
	var an = make([]Function, 0, l)
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
	return NewwArguments(an...)
}

// ACCESSS ATTRIBUTE
func NewParameter(d ...Paired) Parametric {
	return Param(func(di ...Paired) (Paired, Parametric) {
		// if parameters where passed‥.
		if len(di) > 0 { // return former parameter‥.
			// ‥.and enclosure over newly passed parameters
			return di[0], NewParameter(NewPair(di[0].Left(), di[0].Right()))
		} //‥.otherwise, pass on unaltered results from last/first call
		return NewPair(d[0].Left(), d[0].Right()),
			NewParameter(NewPair(d[0].Left(), d[0].Right()))
	})
}
func (p Param) Apply(pa ...Paired) (Paired, Parametric) {
	if len(pa) > 0 {
		return pa[0], NewParameter(pa...)
	}
	return p()
}
func (p Param) Arg() Argumented            { return NewArgument(p.Pair().Right()) }
func (p Param) Ident() Function            { return p }
func (p Param) Accs() Parametric           { _, acc := p(); return acc }
func (p Param) Pair() Paired               { pa, _ := p(); return pa }
func (p Param) Key() Function              { return p.Pair().Left() }
func (p Param) Data() Function             { return p.Pair().Right() }
func (p Param) Left() Function             { return p.Pair().Left() }
func (p Param) Right() Function            { return p.Pair().Right() }
func (p Param) Both() (Function, Function) { return p.Pair().Both() }
func (p Param) Empty() bool {
	l, r := p.Pair().Both()
	return ElemEmpty(l) && ElemEmpty(r)
}
func (p Param) AccType() d.BitFlag { return p.Key().Flag() }
func (p Param) Flag() d.BitFlag {
	dat, _ := p()
	return dat.Flag() |
		d.Vector.Flag() |
		d.Parameter.Flag() |
		Accessor.Flag()
}

// ACCESS ATTRIBUTE SET
func NewParameterSet(pairs ...Paired) ParamSet {
	return ParamSet(func(pairs ...Paired) ([]Paired, Parameters) {
		return pairs, ParamSet(func(...Paired) ([]Paired, Parameters) {
			return pairs, NewParameterSet(pairs...)
		})

	})
}
func NewParameters(pairs ...Paired) Parameters {
	return ParamSet(
		func(po ...Paired) ([]Paired, Parameters) {
			if len(po) > 0 {
				return po, NewParameters(po...)
			}
			return pairs, NewParameters(pairs...)
		})
}
func (a ParamSet) GetIdx(acc Function) (int, pairSorter) {
	var ps = newPairSorter(a.Pairs()...)
	switch {
	case acc.Flag().Match(d.Symbolic.Flag()):
		ps.Sort(d.String)
	case acc.Flag().Match(d.Unsigned.Flag()):
		ps.Sort(d.Unsigned)
	case acc.Flag().Match(d.Integer.Flag()):
		ps.Sort(d.Unsigned)
	}
	return ps.Search(acc), ps
}
func (a ParamSet) Get(acc Function) Paired {
	var idx, ps = a.GetIdx(acc)
	if idx >= 0 {
		return ps[idx]
	}
	return nil
}
func (a ParamSet) Replace(acc Paired) Parameters {
	idx, ps := a.GetIdx(acc.Left())
	ps[idx] = acc
	return NewParameters(ps...)
}
func (a ParamSet) Apply(acc ...Paired) ([]Paired, Parameters) {
	if len(acc) > 0 {
		return acc, NewParameters(acc...)
	}
	return a()
}
func (a ParamSet) Flag() d.BitFlag {
	var f = d.BitFlag(0)
	for _, acc := range a.Pairs() {
		f = f | acc.Flag()
	}
	return f |
		d.Vector.Flag() |
		d.Parameter.Flag() |
		Accessor.Flag()
}
func (a ParamSet) Pairs() []Paired { pairs, _ := a(); return pairs }
func (a ParamSet) Len() int        { pairs, _ := a(); return len(pairs) }
func (a ParamSet) Empty() bool {
	if len(a.Pairs()) > 0 {
		for _, p := range a.Pairs() {
			if !ElemEmpty(p) {
				return false
			}
		}
	}
	return true
}
func (a ParamSet) AccSet() Parameters { _, set := a(); return set }
func (a ParamSet) Ident() Function    { return a }
func (a ParamSet) Append(v ...Paired) Parameters {
	return NewParameters(append(a.Pairs(), v...)...)
}
func ApplyParams(acc Parameters, praed ...Paired) Parameters {
	var ps = newPairSorter(acc.Pairs()...)
	ps.Sort(d.Type(praed[0].Left().Flag()))
	for _, p := range praed {
		idx := ps.Search(p.Left())
		if idx >= 0 {
			ps[idx] = p
			continue
		}
		ps = append(ps, p)
	}
	return NewParameters(ps...)
}
