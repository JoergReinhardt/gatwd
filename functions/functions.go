/*
BASE FUNCTIONS ARGUMENTS, PARAMETERS & 'APPLICABLES'

  implements arguments and parameters as well as sets there of, to pass to and
  return values from functional type implementations and apply sets of
  arguments/parameters to them.
*/
package functions

import (
	"fmt"

	d "github.com/JoergReinhardt/godeep/data"
)

// type TyHigherOrder d.BitFlag
// encodes the kind of functional data as bitflag
type TyHigherOrder d.BitFlag

func (t TyHigherOrder) Flag() d.BitFlag         { return d.BitFlag(t) }
func (t TyHigherOrder) TypeHO() TyHigherOrder   { return t }
func (t TyHigherOrder) TypePrim() d.TyPrimitive { return d.Flag }
func (t TyHigherOrder) Uint() uint              { return d.BitFlag(t).Uint() }

//go:generate stringer -type=TyHigherOrder
const (
	Data TyHigherOrder = 1 << iota
	Closure
	Function // functions are polymorph‥.
	Argument
	Parameter
	Unbound  // map key, slice index, search parameter...
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
	Internal

	Chain = Vector | Tuple | Record

	AccIndex = Vector | Chain

	AccSymbol = Tuple | AssocA | Record

	AccCollect = AccIndex | AccSymbol

	Nests = Tuple | List

	Sets = UniSet | MuliSet | AssocA | Record

	Links = Link | DLink | Node | Tree // Consumeables
)

type ( // HIGHER ORDER FUNCTION TYPES
	PrimeVal func() d.Primary // represents constructors for primary data types
	FncVal   func() Value
	PairVal  func() (a, b Value) // <- base element of all tuples and collections
	ArgVal   func(d ...Value) (Value, Argumented)
	ArgSet   func(d ...Value) ([]Value, Arguments)
	ParamVal func(d ...Paired) (Paired, Parametric)
	ParamSet func(d ...Parametric) ([]Parametric, Parameters)
)

// instanciate functionalized data
func New(inf ...interface{}) Value {
	return PrimeVal(func() d.Primary { return d.New(inf...) })
}
func NewFromData(dat d.Primary) Value {
	return PrimeVal(func() d.Primary { return dat })
}

// VALUE
//
// methods of the value type
func (dat PrimeVal) TypePrim() d.TyPrimitive { return dat().TypePrim() }
func (dat PrimeVal) TypeHO() TyHigherOrder   { return Data }
func (dat PrimeVal) Empty() bool             { return ElemEmpty(dat) }
func (dat PrimeVal) Ident() Value            { return dat }
func (dat PrimeVal) Eval(a ...d.Primary) d.Primary {
	if len(a) > 0 {
		if !dat.Empty() {
			return d.NewFromPrimary(append([]d.Primary{dat()}, a...)...)
		}
		return d.NewFromPrimary(a...)
	}
	return dat()
}
func (dat PrimeVal) Call(...d.Evaluable) d.Primary { return dat() }

func (dat FncVal) TypePrim() d.TyPrimitive { return dat().TypePrim() }
func (dat FncVal) TypeHO() TyHigherOrder   { return Data | Function }
func (dat FncVal) Empty() bool             { return ElemEmpty(dat) }
func (dat FncVal) Ident() Value            { return dat }
func (dat FncVal) Eval(a ...d.Primary) d.Primary {
	if len(a) > 0 {
		if !dat.Empty() {
			return d.NewFromPrimary(append([]d.Primary{dat()}, a...)...)
		}
		return d.NewFromPrimary(a...)
	}
	return dat()
}
func (dat FncVal) Call(...d.Evaluable) d.Primary { return dat() }

func ElemEmpty(dat Value) bool {
	if dat != nil {
		if !dat.Eval().TypePrim().Flag().Match(d.Nil.Flag()) {
			return false
		}
	}
	return true
}

// PAIR
//
// pair encloses two data instances
func NewPair(l, r Value) Paired {
	return PairVal(func() (Value, Value) { return l, r })
}
func NewPairFromInterface(l, r interface{}) Paired {
	return PairVal(func() (Value, Value) { return New(l), New(r) })
}
func NewPairFromData(l, r d.Primary) Paired {
	return PairVal(func() (Value, Value) { return NewFromData(l), NewFromData(r) })
}
func (p PairVal) Both() (Value, Value)  { return p() }
func (p PairVal) TypeHO() TyHigherOrder { return Pair }
func (p PairVal) TypePrim() d.TyPrimitive {
	return d.Pair.TypePrim() | p.Left().TypePrim() | p.Right().TypePrim()
}
func (p PairVal) Pair() Value                   { return p }
func (p PairVal) Left() Value                   { l, _ := p(); return l }
func (p PairVal) Right() Value                  { _, r := p(); return r }
func (p PairVal) Acc() Value                    { return p.Left() }
func (p PairVal) Arg() Value                    { return p.Right() }
func (p PairVal) AccType() d.BitFlag            { return p.Left().TypePrim().Flag() }
func (p PairVal) ArgType() d.BitFlag            { return p.Right().TypePrim().Flag() }
func (p PairVal) Ident() Value                  { return p }
func (p PairVal) Eval(a ...d.Primary) d.Primary { return d.NewPair(p.Left().Eval(), p.Right().Eval()) }
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
func NewArgument(do ...Value) Argumented {
	return ArgVal(func(di ...Value) (Value, Argumented) {
		// if parameters where passed‥.
		if len(di) > 0 { // return former parameter‥.
			// ‥.and enclosure over newly passed parameters
			return di[0], NewArgument(di[0])
		} //‥.otherwise, pass on unaltered results from last/first call
		return do[0], NewArgument(do[0])
	})
}
func (p ArgVal) Apply(d ...Value) (Value, Argumented) {
	if len(d) > 0 {
		return d[0], NewArgument(d...)
	}
	return p()
}
func (p ArgVal) Arg() Value                    { k, _ := p(); return k }
func (p ArgVal) Argumented() Value             { _, d := p(); return d }
func (p ArgVal) Ident() Value                  { return p }
func (p ArgVal) Eval(a ...d.Primary) d.Primary { return p.Arg() }
func (p ArgVal) Empty() bool                   { return ElemEmpty(p.Arg()) }
func (p ArgVal) ArgType() d.TyPrimitive        { return p.Arg().TypePrim() }
func (p ArgVal) TypeHO() TyHigherOrder         { return Argument }
func (p ArgVal) TypePrim() d.TyPrimitive       { return p.ArgType() }

//
// ARGUMENT SET
//
// collections of arguments provide methods to apply values contained in other
// collections based on position to replace the given values and yield the
// resulting collection of arguments.
func NewwArguments(dat ...Value) Arguments {
	return ArgSet(func(dot ...Value) ([]Value, Arguments) {
		if len(dot) > 0 {
			return dot, NewwArguments(dot...)
		}
		return dat, NewwArguments(dat...)
	})
}
func NewArgSet(dat ...Value) Arguments {
	return ArgSet(func(dot ...Value) ([]Value, Arguments) {
		return dat,
			ArgSet(
				func(...Value) ([]Value, Arguments) {
					return dat, NewwArguments(dat...)
				})

	})
}
func (a ArgSet) TypeHO() TyHigherOrder { return Argument | Vector }
func (a ArgSet) TypePrim() d.TyPrimitive {
	var f = d.BitFlag(uint(0))
	for _, arg := range a.Args() {
		f = f.Concat(arg.TypePrim().Flag())
	}
	f = f | d.Vector.TypePrim().Flag()
	return d.TyPrimitive(f)
}
func (a ArgSet) Args() []Argumented {
	var args = []Argumented{}
	for _, arg := range a.Data() {
		args = append(args, NewArgument(arg))
	}
	return args
}
func (a ArgSet) Data() []Value { d, _ := a(); return d }
func (a ArgSet) Len() int      { d, _ := a(); return len(d) }
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
func (a ArgSet) ArgSet() Arguments             { _, as := a(); return as }
func (a ArgSet) Ident() Value                  { return a }
func (a ArgSet) Eval(p ...d.Primary) d.Primary { return a.ArgSet() }
func (a ArgSet) Get(idx int) Argumented        { return a.Args()[idx] }
func (a ArgSet) Set(idx int, dat Value) ArgSet { a.Args()[idx] = NewArgument(dat); return a }
func (a ArgSet) Replace(idx int, arg Value) Arguments {
	dats, _ := a()
	dats[idx] = arg
	return NewwArguments(dats...)
}
func (a ArgSet) Apply(dd ...Value) ([]Value, Arguments) {
	var dats = []Value{}
	var args = a.ArgSet()
	for i, dat := range dd {
		dats = append(dats, dat)
		args = args.Replace(i, NewArgument(dat))
	}
	return dats, args
}
func ApplyArgs(ao ArgSet, args ...Value) Arguments {
	oargs, _ := ao()
	var l = len(oargs)
	if l < len(args) {
		l = len(args)
	}
	var an = make([]Value, 0, l)
	var i int
	for i, _ = range oargs {
		// copy old arguments to return set, if any are set at this pos.
		if oargs[i] != nil && d.Nil.TypePrim().Flag().Match(oargs[i].TypePrim().Flag()) {
			an[i] = oargs[i]
		}
		// copy new arguments to return set, if any are set at this
		// position. overwrite old arguments in case any where set at
		// this position.
		if args[i] != nil && d.Nil.TypePrim().Flag().Match(args[i].TypePrim().Flag()) {
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
func NewKeyValueParm(k, v Value) Parametric { return NewParameter(NewPair(k, v)) }
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
func (p ParamVal) ApplyKeyVal(k, v Value) (Value, Parametric) {
	return p.Apply(NewKeyValueParm(k, v))
}
func (p ParamVal) Apply(pa ...Parametric) (Value, Parametric) {
	if len(pa) == 0 {
		return p()
	}
	return pa[0].(ParamVal)()
}
func (p ParamVal) Ident() Value                  { return p }
func (p ParamVal) Parm() Parametric              { _, parm := p(); return parm }
func (p ParamVal) Pair() Paired                  { pa, _ := p(); return pa }
func (p ParamVal) Left() Value                   { return p.Pair().Left() }
func (p ParamVal) Right() Value                  { return p.Pair().Right() }
func (p ParamVal) Arg() Value                    { return p.Pair().Right() }
func (p ParamVal) Acc() Value                    { return p.Pair().Left() }
func (p ParamVal) ArgType() d.BitFlag            { return p.Arg().TypePrim().Flag() }
func (p ParamVal) AccType() d.BitFlag            { return p.Acc().TypePrim().Flag() }
func (p ParamVal) Eval(a ...d.Primary) d.Primary { return NewPair(p.Acc(), p.Arg()) }
func (p ParamVal) Both() (Value, Value)          { return p.Pair().Both() }
func (p ParamVal) Empty() bool {
	l, r := p.Pair().Both()
	return ElemEmpty(l) && ElemEmpty(r)
}
func (p ParamVal) TypePrim() d.TyPrimitive { return p.Pair().TypePrim() }
func (p ParamVal) TypeHO() TyHigherOrder   { return Parameter }

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
func (a ParamSet) AppendKeyValue(k, v Value) Parameters {
	return NewParameterSet(append(a.Parms(), NewKeyValueParm(k, v))...)
}
func (a ParamSet) GetIdx(acc Value) (int, pairSorter) {
	var ps = newPairSorter(a.Pairs()...)
	switch {
	case acc.Eval().TypePrim().Flag().Match(d.Symbolic.TypePrim().Flag()):
		ps.Sort(d.String)
	case acc.Eval().TypePrim().Flag().Match(d.Natural.TypePrim().Flag()):
		ps.Sort(d.Natural)
	case acc.Eval().TypePrim().Flag().Match(d.Integer.TypePrim().Flag()):
		ps.Sort(d.Natural)
	}
	var idx = ps.Search(acc)
	if idx != -1 {
		return idx, ps
	}
	return -1, ps
}
func (a ParamSet) Get(acc Value) Paired {
	var idx, ps = a.GetIdx(acc)
	fmt.Println(idx)
	if idx >= 0 {
		return ps[idx]
	}
	return NewPairFromInterface(d.NilVal{}, d.NilVal{})
}
func (a ParamSet) Set(acc, val Value) ParamSet {
	idx, ps := a.GetIdx(acc)
	ps[idx] = NewPair(acc, val)
	return NewParameters(ps...).(ParamSet)
}
func (a ParamSet) Replace(acc Paired) Parameters {
	idx, ps := a.GetIdx(acc.Left())
	ps[idx] = acc
	return NewParameters(ps...)
}
func (a ParamSet) ReplaceKeyValue(k, v Value) Parameters {
	return a.Replace(NewPair(k, v))
}
func (a ParamSet) ApplyKeyValue(k, v Value) ([]Parametric, Parameters) {
	return a.Apply(NewKeyValueParm(k, v))
}
func (a ParamSet) Apply(args ...Parametric) ([]Parametric, Parameters) {
	if len(args) == 0 {
		return a()
	}
	ps := newPairSorter(a.Pairs()...)
	for _, arg := range args {
		idx := ps.Search(arg.Acc())
		if idx != -1 {
			ps[idx] = arg.Pair()
			continue
		}
		ps = append(ps, arg.Pair())
		ps.Sort(d.TyPrimitive(arg.Acc().Eval().TypePrim()))
	}
	parameters := NewParameters(ps...)
	return parameters.Parms(), parameters
}
func (a ParamSet) TypeHO() TyHigherOrder { return Parameter }
func (a ParamSet) TypePrim() d.TyPrimitive {
	var f = d.BitFlag(0).Flag()
	for _, pair := range a.Pairs() {
		f = f | pair.TypePrim().Flag()
	}
	return d.TyPrimitive(f | d.Vector.TypePrim().Flag())
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
func (a ParamSet) Ident() Value                  { return a }
func (a ParamSet) Eval(p ...d.Primary) d.Primary { return a }
func (a ParamSet) Append(v ...Paired) Parameters {
	return NewParameters(append(a.Pairs(), v...)...)
}
func ApplyParams(acc Parameters, praed ...Paired) Parameters {
	var ps = newPairSorter(acc.Pairs()...)
	ps.Sort(d.TyPrimitive(praed[0].Left().Eval().TypePrim()))
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
