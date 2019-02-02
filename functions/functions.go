/*
BASE FUNCTIONS ARGUMENTS, PARAMETERS & 'APPLICABLES'

  implements arguments and parameters as well as sets there of, to pass to and
  return values from functional type implementations and apply sets of
  arguments/parameters to them.
*/
package functions

import (
	"bytes"
	"fmt"

	d "github.com/JoergReinhardt/godeep/data"
	l "github.com/JoergReinhardt/godeep/lex"
)

// type Kind d.BitFlag
// encodes the kind of functional data as bitflag
type Kind d.BitFlag

func (t Kind) Flag() d.BitFlag { return d.BitFlag(t).Flag() }
func (t Kind) Uint() uint      { return d.BitFlag(t).Uint() }

//go:generate stringer -type=Kind
const (
	Value Kind = 1 << iota
	Instance
	Polymorph
	Argument
	Parameter
	Attribut // map key, slice index, search parameter...
	Accessor // pair of Attr & Value
	Generator
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
	ArgVal func(d ...Functional) (Functional, Argumented)
	//
	// PAIR
	//
	// wraps generic pairs of functional data
	PairVal func() (a, b Functional) // <- base element of all tuples and collections
	//
	// ARGUMENTS
	//
	// collection of arguments
	ArgSet func(d ...Functional) ([]Functional, Arguments)
	//
	// PARAMETER
	//
	// a yields a position/key & value pair.
	ParamVal func(d ...Paired) (Paired, Parametric)
	//
	// PARAMETERS
	//
	// collection of parameters
	ParamSet func(d ...Parametric) ([]Parametric, Parameters)
)

// instanciate functionalized data
func New(inf ...interface{}) Functional {
	return DataVal(func() d.Data { return (d.NewFI(inf...)) })
}
func NewFromData(dat d.Data) Functional {
	return DataVal(func() d.Data { return dat })
}

// VALUE
//
// methods of the value type
func (dat DataVal) Flag() d.BitFlag            { return dat().Flag() | d.Function.Flag() }
func (dat DataVal) Kind() d.BitFlag            { return Value.Flag() }
func (dat DataVal) Empty() bool                { return ElemEmpty(dat) }
func (dat DataVal) Ident() Functional          { return dat }
func (dat DataVal) Eval() d.Data               { return dat() }
func (dat DataVal) Call(...d.Evaluable) d.Data { return dat() }
func (dat DataVal) String() string {
	var buf = bytes.NewBuffer([]byte{})
	buf.WriteString(l.Lambda.Syntax())
	buf.WriteString(l.Blank.Syntax())
	buf.WriteString(l.RightArrow.Syntax())
	buf.WriteString(l.Blank.Syntax())
	buf.WriteString(l.LeftBra.Syntax())
	buf.WriteString(dat.Flag().String())
	buf.WriteString(l.RightBra.Syntax())
	buf.WriteString(l.Blank.Syntax())
	buf.WriteString(dat.Eval().String())
	return buf.String()
}

func ElemEmpty(dat Functional) bool {
	if dat != nil {
		if !dat.Eval().Flag().Match(d.Nil.Flag()) {
			return false
		}
	}
	return true
}

// PAIR
//
// pair encloses two data instances
func NewPair(l, r Functional) Paired {
	return PairVal(func() (Functional, Functional) { return l, r })
}
func NewPairFromInterface(l, r interface{}) Paired {
	return PairVal(func() (Functional, Functional) { return New(l), New(r) })
}
func NewPairFromData(l, r d.Data) Paired {
	return PairVal(func() (Functional, Functional) { return NewFromData(l), NewFromData(r) })
}
func (p PairVal) Both() (Functional, Functional) { return p() }
func (p PairVal) Kind() d.BitFlag                { return Pair.Flag() }
func (p PairVal) Flag() d.BitFlag                { return d.Pair.Flag() | p.Left().Flag() | p.Right().Flag() }
func (p PairVal) Pair() Functional               { return p }
func (p PairVal) Left() Functional               { l, _ := p(); return l }
func (p PairVal) Right() Functional              { _, r := p(); return r }
func (p PairVal) Acc() Functional                { return p.Left() }
func (p PairVal) Arg() Functional                { return p.Right() }
func (p PairVal) AccType() d.BitFlag             { return p.Left().Flag() }
func (p PairVal) ArgType() d.BitFlag             { return p.Right().Flag() }
func (p PairVal) Ident() Functional              { return p }
func (p PairVal) Eval() d.Data                   { return NewPair(p.Left(), p.Right()) }
func (p PairVal) Empty() bool {
	return ElemEmpty(p.Left()) && ElemEmpty(p.Right())
}
func (p PairVal) String() string {
	var buf = bytes.NewBuffer([]byte{})
	buf.WriteString(p.Left().String())
	buf.WriteString(l.Colon.Syntax())
	buf.WriteString(l.Blank.Syntax())
	buf.WriteString(p.Right().String())
	return buf.String()
}

//
// ARGUMENT
//
// arguments are data instances that yield enclosed data and copy of the
// argument instance, when called with empty parameter set. when called with
// arguments, they yield the passed data and a new argument instance eclosing
// that new data instead.
func NewArgument(do ...Functional) Argumented {
	return ArgVal(func(di ...Functional) (Functional, Argumented) {
		// if parameters where passed‥.
		if len(di) > 0 { // return former parameter‥.
			// ‥.and enclosure over newly passed parameters
			return di[0], NewArgument(di[0])
		} //‥.otherwise, pass on unaltered results from last/first call
		return do[0], NewArgument(do[0])
	})
}
func (p ArgVal) Apply(d ...Functional) (Functional, Argumented) {
	if len(d) > 0 {
		return d[0], NewArgument(d...)
	}
	return p()
}
func (p ArgVal) Arg() Functional        { k, _ := p(); return k }
func (p ArgVal) Argumented() Functional { _, d := p(); return d }
func (p ArgVal) Ident() Functional      { return p }
func (p ArgVal) Eval() d.Data           { return p.Arg() }
func (p ArgVal) ArgType() d.BitFlag     { return p.Arg().Flag() }
func (p ArgVal) Empty() bool            { return ElemEmpty(p.Arg()) }
func (p ArgVal) Kind() d.BitFlag        { return d.Argument.Flag() }
func (p ArgVal) Flag() d.BitFlag        { return p.Arg().Eval().Flag() | d.Argument.Flag() }
func (p ArgVal) String() string         { return p.Arg().String() }

//
// ARGUMENT SET
//
// collections of arguments provide methods to apply values contained in other
// collections based on position to replace the given values and yield the
// resulting collection of arguments.
func NewwArguments(dat ...Functional) Arguments {
	return ArgSet(func(dot ...Functional) ([]Functional, Arguments) {
		if len(dot) > 0 {
			return dot, NewwArguments(dot...)
		}
		return dat, NewwArguments(dat...)
	})
}
func NewArgSet(dat ...Functional) Arguments {
	return ArgSet(func(dot ...Functional) ([]Functional, Arguments) {
		return dat,
			ArgSet(
				func(...Functional) ([]Functional, Arguments) {
					return dat, NewwArguments(dat...)
				})

	})
}
func (a ArgSet) Kind() d.BitFlag { return Argument.Flag() | Vector.Flag() }
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
func (a ArgSet) Data() []Functional { d, _ := a(); return d }
func (a ArgSet) Len() int           { d, _ := a(); return len(d) }
func (a ArgSet) Empty() bool {
	if len(a.Args()) > 0 {
		for _, arg := range a.Args() {
			if !ElemEmpty(arg.Arg()) {
				return false
			}
		}
	}
	return true
}
func (a ArgSet) ArgSet() Arguments                  { _, as := a(); return as }
func (a ArgSet) Ident() Functional                  { return a }
func (a ArgSet) Eval() d.Data                       { return a.ArgSet() }
func (a ArgSet) Get(idx int) Argumented             { return a.Args()[idx] }
func (a ArgSet) Set(idx int, dat Functional) ArgSet { a.Args()[idx] = NewArgument(dat); return a }
func (a ArgSet) Replace(idx int, arg Functional) Arguments {
	dats, _ := a()
	dats[idx] = arg
	return NewwArguments(dats...)
}
func (a ArgSet) Apply(dd ...Functional) ([]Functional, Arguments) {
	var dats = []Functional{}
	var args = a.ArgSet()
	for i, dat := range dd {
		dats = append(dats, dat)
		args = args.Replace(i, NewArgument(dat))
	}
	return dats, args
}
func (p ArgSet) String() string {
	var buf = bytes.NewBuffer([]byte{})
	buf.WriteString(l.LeftBra.Syntax())
	var args = p.Data()
	var length = len(args) - 1
	for i, arg := range args {
		buf.WriteString(arg.String())
		if i < length {
			buf.WriteString(l.Comma.Syntax())
			buf.WriteString(l.Blank.Syntax())
		}
	}
	buf.WriteString(l.RightBra.Syntax())
	return buf.String()
}
func ApplyArgs(ao ArgSet, args ...Functional) Arguments {
	oargs, _ := ao()
	var l = len(oargs)
	if l < len(args) {
		l = len(args)
	}
	var an = make([]Functional, 0, l)
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
func NewKeyValueParm(k, v Functional) Parametric { return NewParameter(NewPair(k, v)) }
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
func (p ParamVal) ApplyKeyVal(k, v Functional) (Functional, Parametric) {
	return p.Apply(NewKeyValueParm(k, v))
}
func (p ParamVal) Apply(pa ...Parametric) (Functional, Parametric) {
	if len(pa) == 0 {
		return p()
	}
	return pa[0].(ParamVal)()
}
func (p ParamVal) Ident() Functional              { return p }
func (p ParamVal) Parm() Parametric               { _, parm := p(); return parm }
func (p ParamVal) Pair() Paired                   { pa, _ := p(); return pa }
func (p ParamVal) Arg() Functional                { return p.Pair().Right() }
func (p ParamVal) Acc() Functional                { return p.Pair().Left() }
func (p ParamVal) ArgType() d.BitFlag             { return p.Arg().Flag() }
func (p ParamVal) AccType() d.BitFlag             { return p.Acc().Flag() }
func (p ParamVal) Eval() d.Data                   { return NewPair(p.Acc(), p.Arg()) }
func (p ParamVal) Both() (Functional, Functional) { return p.Pair().Both() }
func (p ParamVal) Empty() bool {
	l, r := p.Pair().Both()
	return ElemEmpty(l) && ElemEmpty(r)
}
func (p ParamVal) Flag() d.BitFlag { return p.Pair().Flag() | d.Parameter.Flag() }
func (p ParamVal) Kind() d.BitFlag { return Parameter.Flag() }
func (p ParamVal) String() string  { return p.Pair().String() }

// PARAMETERS
//
// collection of parameters has the methods to apply another collection of
// parameters and replace the contained ones based on accessor (order doesn't
// matter).
func NewParameterSet(parms ...Parametric) ParamSet {
	return ParamSet(func(parms ...Parametric) ([]Parametric, Parameters) {
		return parms, ParamSet(func(...Parametric) ([]Parametric, Parameters) {
			return parms, NewParameterSet(parms...)
		})

	})
}
func NewParameters(pairs ...Paired) Parameters {
	var parms []Parametric
	for _, parm := range pairs {
		parms = append(parms, NewParameter(parm))
	}
	return ParamSet(
		func(po ...Parametric) ([]Parametric, Parameters) {
			if len(po) > 0 {
				return po, NewParameterSet(po...)
			}
			return parms, NewParameterSet(parms...)
		})
}
func (a ParamSet) AppendKeyValue(k, v Functional) Parameters {
	return NewParameterSet(append(a.Parms(), NewKeyValueParm(k, v))...)
}
func (a ParamSet) GetIdx(acc Functional) (int, pairSorter) {
	var ps = newPairSorter(a.Pairs()...)
	switch {
	case acc.Eval().Flag().Match(d.Symbolic.Flag()):
		ps.Sort(d.String)
	case acc.Eval().Flag().Match(d.Unsigned.Flag()):
		ps.Sort(d.Unsigned)
	case acc.Eval().Flag().Match(d.Integer.Flag()):
		ps.Sort(d.Unsigned)
	}
	var idx = ps.Search(acc)
	if idx != -1 {
		return idx, ps
	}
	return -1, ps
}
func (a ParamSet) Get(acc Functional) Paired {
	var idx, ps = a.GetIdx(acc)
	fmt.Println(idx)
	if idx >= 0 {
		return ps[idx]
	}
	return NewPairFromInterface(d.NilVal{}, d.NilVal{})
}
func (a ParamSet) Set(acc, val Functional) ParamSet {
	idx, ps := a.GetIdx(acc)
	ps[idx] = NewPair(acc, val)
	return NewParameters(ps...).(ParamSet)
}
func (a ParamSet) Replace(acc Paired) Parameters {
	idx, ps := a.GetIdx(acc.Left())
	ps[idx] = acc
	return NewParameters(ps...)
}
func (a ParamSet) ReplaceKeyValue(k, v Functional) Parameters {
	return a.Replace(NewPair(k, v))
}
func (a ParamSet) ApplyKeyValue(k, v Functional) ([]Parametric, Parameters) {
	return a.Apply(NewKeyValueParm(k, v))
}
func (a ParamSet) Apply(args ...Parametric) ([]Parametric, Parameters) {
	if len(args) == 0 {
		return a()
	}
	ps := newPairSorter(a.Pairs()...)
	for _, arg := range args {
		idx := ps.GetIdx(arg.Acc())
		if idx != -1 {
			ps[idx] = arg.Pair()
			continue
		}
		ps = append(ps, arg.Pair())
		ps.Sort(d.Type(arg.Acc().Eval().Flag()))
	}
	parameters := NewParameters(ps...)
	return parameters.Parms(), parameters
}
func (a ParamSet) Kind() d.BitFlag { return Parameter.Flag() }
func (a ParamSet) Flag() d.BitFlag {
	var f = d.BitFlag(0)
	for _, pair := range a.Pairs() {
		f = f | pair.Flag()
	}
	return f | d.Vector.Flag() | d.Parameter.Flag()
}
func (a ParamSet) Parms() []Parametric { parms, _ := a(); return parms }
func (a ParamSet) Pairs() []Paired {
	var pairs = []Paired{}
	for _, parm := range a.Parms() {
		pairs = append(pairs, NewPair(parm.Acc(), parm.Arg()))
	}
	return pairs
}
func (a ParamSet) Len() int { pairs, _ := a(); return len(pairs) }
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
func (p ParamSet) String() string {
	var buf = bytes.NewBuffer([]byte{})
	buf.WriteString(l.LeftBra.Syntax())
	var parms = p.Parms()
	var length = len(parms) - 1
	for i, parm := range parms {
		buf.WriteString(parm.String())
		if i < length {
			buf.WriteString(l.Comma.Syntax())
			buf.WriteString(l.Blank.Syntax())
		}
	}
	buf.WriteString(l.RightBra.Syntax())
	return buf.String()
}
func (a ParamSet) Ident() Functional { return a }
func (a ParamSet) Eval() d.Data      { return a }
func (a ParamSet) Append(v ...Paired) Parameters {
	return NewParameters(append(a.Pairs(), v...)...)
}
func ApplyParams(acc Parameters, praed ...Paired) Parameters {
	var ps = newPairSorter(acc.Pairs()...)
	ps.Sort(d.Type(praed[0].Left().Eval().Flag()))
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
