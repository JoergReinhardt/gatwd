/*
BASE FUNCTIONS ARGUMENTS, PARAMETERS & 'APPLICABLES'

  implements arguments and parameters as well as sets there of, to pass to and
  return values from functional type implementations and apply sets of
  arguments/parameters to them.
*/
package functions

import (
	d "github.com/JoergReinhardt/godeep/data"
)

// type Kind d.BitFlag
// encodes the kind of functional data as bitflag
type Kind d.BitFlag

func (t Kind) Flag() d.BitFlag { return d.BitFlag(t).Flag() }
func (t Kind) Uint() uint      { return d.BitFlag(t).Uint() }

//go:generate stringer -type=Kind
const (
	Value Kind = 1 << iota
	Argument
	Parameter
	Attribut // map key, slice index, search parameter...
	Accessor // pair of Attr & Value
	Pair
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
	// basic higher oder data types are implemented as closures over
	// instances of the types of the data package. they provide a common
	// abstraction of 'precedence type'. each higher order type can be
	// reduced to it's unique precedence type in a deterministic way.
	//
	// VALUE
	//
	// most basic form of functionalized data.
	DataVal func() d.Data

	// ARGUMENT
	//
	// implementation of the applyable interface.  returns previously
	// enclosed data and another ArgVal instance, optionaly containing the
	// passed data, if any was passed, or the previous data again.
	ArgVal func(d ...d.Data) (d.Data, Argumented)
	//
	// PAIR
	//
	// wraps generic pairs of functional data
	PairVal func() (a, b d.Data) // <- base element of all tuples and collections
	//
	// ARGUMENTS
	//
	// collection of arguments
	ArgSet func(d ...d.Data) ([]d.Data, Arguments)
	//
	// PARAMETER
	//
	// a yields a position/key & value pair.
	ParamVal func(d ...Paired) (Paired, Parametric)
	//
	// PARAMETERS
	//
	// collection of parameters
	ParamSet func(d ...Paired) ([]Paired, Parameters)
)

// instanciate functionalized data
func NewValue(dat d.Data) Functional { return DataVal(func() d.Data { return dat }) }

// VALUE
//
// methods of the value type
func (dat DataVal) Flag() d.BitFlag   { return dat().Flag() | d.Function.Flag() }
func (dat DataVal) Kind() BitFlag     { return Value.Flag() }
func (dat DataVal) Empty() bool       { return ElemEmpty(dat()) }
func (dat DataVal) Ident() Functional { return dat }
func (dat DataVal) Eval() d.Data      { return dat() }

func ElemEmpty(dat d.Data) bool {
	if dat != nil {
		if !dat.Flag().Match(d.Nil.Flag()) {
			return false
		}
	}
	return true
}

// PAIR
//
// pair encloses two data instances
func NewPair(l, r d.Data) Paired         { return PairVal(func() (d.Data, d.Data) { return l, r }) }
func (p PairVal) Both() (d.Data, d.Data) { return p() }
func (p PairVal) Kind() BitFlag          { return Pair.Flag() }
func (p PairVal) Flag() d.BitFlag        { return d.Pair.Flag() | p.Left().Flag() | p.Right().Flag() }
func (p PairVal) Left() d.Data           { l, _ := p(); return l }
func (p PairVal) Right() d.Data          { _, r := p(); return r }
func (p PairVal) Acc() Parametric        { return NewParameter(NewPair(p.Left(), p.Right())) }
func (p PairVal) Arg() Argumented        { return NewArgument(p.Right()) }
func (p PairVal) Ident() Functional      { return p }
func (p PairVal) Eval() d.Data           { return NewPair(p.Left(), p.Right()) }
func (p PairVal) Empty() bool {
	return ElemEmpty(p.Left()) && ElemEmpty(p.Right())
}

//
// ARGUMENT
//
// arguments are data instances that yield enclosed data and copy of the
// argument instance, when called with empty parameter set. when called with
// arguments, they yield the passed data and a new argument instance eclosing
// that new data instead.
func NewArgument(do ...d.Data) Argumented {
	return ArgVal(func(di ...d.Data) (d.Data, Argumented) {
		// if parameters where passed‥.
		if len(di) > 0 { // return former parameter‥.
			// ‥.and enclosure over newly passed parameters
			return di[0], NewArgument(di[0])
		} //‥.otherwise, pass on unaltered results from last/first call
		return do[0], NewArgument(do[0])
	})
}
func (p ArgVal) Apply(d ...d.Data) (d.Data, Argumented) {
	if len(d) > 0 {
		return d[0], NewArgument(d...)
	}
	return p()
}
func (p ArgVal) Data() d.Data       { d, _ := p(); return d }
func (p ArgVal) Ident() Functional  { return p }
func (p ArgVal) Eval() d.Data       { return p.Data() }
func (p ArgVal) Arg() Argumented    { return NewArgument(p.Data()) }
func (p ArgVal) Param() d.Data      { return p.Data() }
func (p ArgVal) ParamType() BitFlag { return p.Data().Flag() }
func (p ArgVal) DataType() BitFlag  { return p.Data().Flag() }
func (p ArgVal) ArgType() BitFlag   { return p.Data().Flag() }
func (p ArgVal) Empty() bool        { return ElemEmpty(p.Data()) }
func (p ArgVal) Kind() BitFlag      { return d.Argument.Flag() }
func (p ArgVal) Flag() d.BitFlag    { return p.Data().Flag() | d.Argument.Flag() }

//
// ARGUMENT SET
//
// collections of arguments provide methods to apply values contained in other
// collections based on position to replace the given values and yield the
// resulting collection of arguments.
func NewwArguments(dat ...d.Data) Arguments {
	return ArgSet(func(dot ...d.Data) ([]d.Data, Arguments) {
		if len(dot) > 0 {
			return dot, NewwArguments(dot...)
		}
		return dat, NewwArguments(dat...)
	})
}
func NewArgSet(dat ...d.Data) Arguments {
	return ArgSet(func(dot ...d.Data) ([]d.Data, Arguments) {
		return dat,
			ArgSet(
				func(...d.Data) ([]d.Data, Arguments) {
					return dat, NewwArguments(dat...)
				})

	})
}
func (a ArgSet) Kind() BitFlag { return Argument.Flag() | Vector.Flag() }
func (a ArgSet) Flag() d.BitFlag {
	var f = d.BitFlag(uint(0))
	for _, arg := range a.Args() {
		f = f.Concat(arg.Flag())
	}
	f = f | d.Argument.Flag() | d.Vector.Flag()
	return f
}
func (a ArgSet) Args() []Argumented {
	var args = []Argumented{}
	for _, arg := range a.Data() {
		args = append(args, NewArgument(arg))
	}
	return args
}
func (a ArgSet) Data() []d.Data { d, _ := a(); return d }
func (a ArgSet) Len() int       { d, _ := a(); return len(d) }
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
func (a ArgSet) ArgSet() Arguments              { _, as := a(); return as }
func (a ArgSet) Ident() Functional              { return a }
func (a ArgSet) Eval() d.Data                   { return a.ArgSet() }
func (a ArgSet) Get(idx int) Argumented         { return a.Args()[idx] }
func (a ArgSet) Set(idx int, dat d.Data) ArgSet { a.Args()[idx] = NewArgument(dat); return a }
func (a ArgSet) Replace(idx int, arg d.Data) Arguments {
	dats, _ := a()
	dats[idx] = arg
	return NewwArguments(dats...)
}
func (a ArgSet) Apply(dd ...d.Data) ([]d.Data, Arguments) {
	var dats = []d.Data{}
	var args = a.ArgSet()
	for i, dat := range dd {
		dats = append(dats, dat)
		args = args.Replace(i, NewArgument(dat))
	}
	return dats, args
}
func ApplyArgs(ao ArgSet, args ...d.Data) Arguments {
	oargs, _ := ao()
	var l = len(oargs)
	if l < len(args) {
		l = len(args)
	}
	var an = make([]d.Data, 0, l)
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

// PARAMETRIC
//
// parameteric values carry an accessor additional to the enclosed argument.
// that accessor can be used as key, search praedicate, order in a list, among
// other things.
func NewKeyValueParm(k, v d.Data) Parametric { return NewParameter(NewPair(k, v)) }
func NewParameter(dd ...Paired) Parametric {
	return ParamVal(func(di ...Paired) (Paired, Parametric) {
		// if parameters where passed‥.
		if len(di) > 0 { // return former parameter‥.
			// ‥.and enclosure over newly passed parameters
			return di[0], NewParameter(NewPair(di[0].Left(), di[0].Right()))
		} //‥.otherwise, pass on unaltered results from last/first call
		return NewPair(dd[0].Left(), dd[0].Right()),
			NewParameter(NewPair(dd[0].Left(), dd[0].Right()))
	})
}
func (p ParamVal) Apply(pa ...Paired) (Paired, Parametric) {
	if len(pa) > 0 {
		return pa[0], NewParameter(pa...)
	}
	return p()
}
func (p ParamVal) Arg() Argumented        { return NewArgument(p.Pair().Right()) }
func (p ParamVal) Ident() Functional      { return p }
func (p ParamVal) Eval() d.Data           { return NewPair(p.Left(), p.Right()) }
func (p ParamVal) Accs() Parametric       { _, acc := p(); return acc }
func (p ParamVal) Pair() Paired           { pa, _ := p(); return pa }
func (p ParamVal) Key() d.Data            { return p.Pair().Left() }
func (p ParamVal) Data() d.Data           { return p.Pair().Right() }
func (p ParamVal) Left() d.Data           { return p.Pair().Left() }
func (p ParamVal) Right() d.Data          { return p.Pair().Right() }
func (p ParamVal) Both() (d.Data, d.Data) { return p.Pair().Both() }
func (p ParamVal) Empty() bool {
	l, r := p.Pair().Both()
	return ElemEmpty(l) && ElemEmpty(r)
}
func (p ParamVal) AccType() d.BitFlag { return p.Key().Flag() }
func (p ParamVal) Flag() d.BitFlag    { return p.Pair().Flag() | d.Parameter.Flag() }
func (p ParamVal) Kind() BitFlag      { return Parameter.Flag() }

// PARAMETERS
//
// collection of parameters has the methods to apply another collection of
// parameters and replace the contained ones based on accessor (order doesn't
// matter).
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
func (a ParamSet) GetIdx(acc d.Data) (int, pairSorter) {
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
func (a ParamSet) Get(acc d.Data) Paired {
	var idx, ps = a.GetIdx(acc)
	if idx >= 0 {
		return ps[idx]
	}
	return nil
}
func (a ParamSet) Set(acc d.Data, key, val d.Data) ParamSet {
	idx, ps := a.GetIdx(acc)
	ps[idx] = NewParameter(NewPair(key, val))
	return NewParameterSet(ps...)
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
func (a ParamSet) Kind() BitFlag { return Parameter.Flag() }
func (a ParamSet) Flag() d.BitFlag {
	var f = d.BitFlag(0)
	for _, pair := range a.Pairs() {
		f = f | pair.Flag()
	}
	return f | d.Vector.Flag() | d.Parameter.Flag()
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
func (a ParamSet) Ident() Functional  { return a }
func (a ParamSet) Eval() d.Data       { return NewVector(a.AccSet()) }
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
