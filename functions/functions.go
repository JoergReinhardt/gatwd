/*
BASE FUNCTIONS ARGUMENTS, PARAMETERS & 'APPLICABLES'

  implements arguments and parameters as well as sets there of, to pass to and
  return values from functional type implementations and apply sets of
  arguments/parameters to them.
*/
package functions

import (
	"fmt"

	d "github.com/JoergReinhardt/gatwd/data"
)

// type TyFnc d.BitFlag
// encodes the kind of functional data as bitflag
type TyFnc d.UintVal

func (t TyFnc) Eval(...d.Primary) d.Primary { return t }
func (t TyFnc) TypeHO() TyFnc               { return t }
func (t TyFnc) TypePrime() d.TyPrime        { return d.Flag }
func (t TyFnc) Flag() d.BitFlag             { return d.BitFlag(t) }
func (t TyFnc) Uint() uint                  { return d.BitFlag(t).Uint() }

//go:generate stringer -type=TyFnc
const (
	Type TyFnc = 1 << iota
	Data
	Constructor
	Closure
	Function // functions are polymorph‥.
	Accessor // pair of Attr & Value
	Argument
	Attribut
	Parameter
	Predicate // Praedicate(Value) Boolean
	Generator
	Unbound // map key, slice index, search parameter...
	Option
	Just
	None
	Or
	If
	Else
	Truth
	True
	False
	Enum
	Case
	Pair
	List
	Tuple
	UniSet
	MuliSet
	AssocVec
	Record
	Vector
	DLink
	Link
	Node
	Tree
	HigherOrder

	Chain = Vector | Tuple | Record

	AccIndex = Vector | Chain

	AccSymbol = Tuple | AssocVec | Record

	AccCollect = AccIndex | AccSymbol

	Nests = Tuple | List

	Sets = UniSet | MuliSet | AssocVec | Record

	Links = Link | DLink | Node | Tree // Consumeables
)

type ( // HIGHER ORDER FUNCTION TYPES
	// PRIMARY DATA
	PrimeVal func() d.Primary // represents constructors for primary data types
	// FUNCTIONAL VALUES
	FncVal func() Value
	// COLLECTIONS
	PairVal  func() (a, b Value)        // <- base element of all tuples and collections
	EnumVal  func() (Value, SumTypeFnc) // implementing 'Optional :: Maybe() bool'
	ArgVal   func(d ...Value) (Value, Argumented)
	ArgSet   func(d ...Value) ([]Value, Arguments)
	ParamVal func(d ...Paired) (Paired, Parametric)
	ParamSet func(d ...Parametric) ([]Parametric, Parameters)
	// HIGHER ORDER VALUES (ATOMIC)
	NoneVal  func()         // None and Just are a pair of optional types
	JustVal  func() Value   // implementing 'Optional :: Maybe() bool'
	OrVal    func() Value   // implementing 'Optional :: Maybe() bool'
	TrueVal  func() Boolean // boolean constants true & false
	FalseVal func() Boolean // implementing 'Boolen :: Bool() bool'
)

// instanciate functionalized data
func New(inf ...interface{}) Value {
	return PrimeVal(func() d.Primary { return d.New(inf...) })
}
func NewFromData(dat d.Primary) Value {
	return PrimeVal(func() d.Primary { return dat })
}

// PRIMARY VALUE
func (dat PrimeVal) TypePrime() d.TyPrime { return dat().TypePrime() }
func (dat PrimeVal) TypeFnc() TyFnc       { return Data }
func (dat PrimeVal) Empty() bool          { return ElemEmpty(dat) }
func (dat PrimeVal) Ident() Value         { return dat }
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

// FUNCRIONAL VALUE
func (dat FncVal) TypePrime() d.TyPrime { return dat().TypePrime() }
func (dat FncVal) TypeFnc() TyFnc       { return Function }
func (dat FncVal) Empty() bool          { return ElemEmpty(dat) }
func (dat FncVal) Ident() Value         { return dat }
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
		if !dat.Eval().TypePrime().Flag().Match(d.Nil.Flag()) {
			return false
		}
	}
	return true
}

// RETURN TYPES OF THE OPTIONAL TYPE
//
// NONE
func NewNone() NoneVal                        { return NoneVal(func() {}) }
func (n NoneVal) Ident() Value                { return n }
func (n NoneVal) Eval(...d.Primary) d.Primary { return d.NilVal{}.Eval() }
func (n NoneVal) Maybe() bool                 { return false }
func (n NoneVal) Nullable() d.Primary         { return d.NilVal{} }
func (n NoneVal) TypeFnc() TyFnc              { return Option | None }
func (n NoneVal) TypePrime() d.TyPrime        { return d.Nil }
func (n NoneVal) String() string              { return "⊥" }

// JUST
func NewJustVal(v Value) JustVal {
	return JustVal(func() Value { return v })
}
func (j JustVal) Ident() Value                  { return j }
func (j JustVal) Eval(p ...d.Primary) d.Primary { return j().Eval(p...) }
func (j JustVal) Maybe() bool                   { return true }
func (j JustVal) Nullable() d.Primary           { return j.Eval() }
func (j JustVal) TypeFnc() TyFnc                { return Option | Just }
func (j JustVal) TypePrime() d.TyPrime          { return j().TypePrime() }
func (j JustVal) String() string                { return j().String() }

// OR
func NewOrVal(v Value) OrVal {
	return OrVal(func() Value { return v })
}
func (o OrVal) Ident() Value                  { return o }
func (o OrVal) Eval(p ...d.Primary) d.Primary { return o().Eval(p...) }
func (o OrVal) Maybe() bool                   { return false }
func (o OrVal) Nullable() d.Primary           { return o.Eval() }
func (o OrVal) TypeFnc() TyFnc                { return Option | Or }
func (o OrVal) TypePrime() d.TyPrime          { return o().TypePrime() }
func (o OrVal) String() string                { return o().String() }

// FUNCTIONAL TRUTH VALUES
func (t TrueVal) Call(...Value) Value {
	return NewPrimaryConstatnt(d.BoolVal(true))
}
func (t TrueVal) Ident() Value                { return t }
func (t TrueVal) Eval(...d.Primary) d.Primary { return t }
func (t TrueVal) Bool() bool                  { return true }
func (t TrueVal) TypeFnc() TyFnc              { return Truth | True }
func (t TrueVal) TypePrime() d.TyPrime        { return d.Bool }
func (t TrueVal) String() string              { return "True" }

func (f FalseVal) Call(...Value) Value {
	return NewPrimaryConstatnt(d.BoolVal(false))
}
func (f FalseVal) Ident() Value                { return f }
func (f FalseVal) Eval(...d.Primary) d.Primary { return f }
func (f FalseVal) Bool() bool                  { return false }
func (f FalseVal) TypeFnc() TyFnc              { return Truth | False }
func (f FalseVal) TypePrime() d.TyPrime        { return d.Bool }
func (f FalseVal) String() string              { return "False" }

// ENUM
func NewEnumVal(val Value, sumType SumTypeFnc) EnumVal {
	return EnumVal(func() (Value, SumTypeFnc) { return val, sumType })
}
func (e EnumVal) Ident() Value                  { return e }
func (e EnumVal) Eval(p ...d.Primary) d.Primary { return e.Value().Eval(p...) }
func (e EnumVal) Nullable() d.Primary           { return e.Eval() }
func (e EnumVal) Enum() SumTypeFnc              { _, et := e(); return et }
func (e EnumVal) Value() Value                  { value, _ := e(); return value }
func (e EnumVal) TypeFnc() TyFnc                { return Enum }
func (e EnumVal) TypePrime() d.TyPrime          { return e.Value().TypePrime() }
func (e EnumVal) String() string                { return e.Value().String() }

// PAIR
func NewPair(l, r Value) Paired {
	return PairVal(func() (Value, Value) { return l, r })
}
func NewPairFromInterface(l, r interface{}) Paired {
	return PairVal(func() (Value, Value) { return New(l), New(r) })
}
func NewPairFromData(l, r d.Primary) Paired {
	return PairVal(func() (Value, Value) { return NewFromData(l), NewFromData(r) })
}
func (p PairVal) Both() (Value, Value) { return p() }
func (p PairVal) TypeFnc() TyFnc       { return Pair | Function }
func (p PairVal) TypePrime() d.TyPrime {
	return d.Pair.TypePrime() | p.Left().TypePrime() | p.Right().TypePrime()
}
func (p PairVal) Pair() Value                   { return p }
func (p PairVal) Left() Value                   { l, _ := p(); return l }
func (p PairVal) Right() Value                  { _, r := p(); return r }
func (p PairVal) Acc() Value                    { return p.Left() }
func (p PairVal) Arg() Value                    { return p.Right() }
func (p PairVal) AccType() d.BitFlag            { return p.Left().TypePrime().Flag() }
func (p PairVal) ArgType() d.BitFlag            { return p.Right().TypePrime().Flag() }
func (p PairVal) Ident() Value                  { return p }
func (p PairVal) Eval(a ...d.Primary) d.Primary { return d.NewPair(p.Left().Eval(), p.Right().Eval()) }
func (p PairVal) Empty() bool {
	return ElemEmpty(p.Left()) && ElemEmpty(p.Right())
}

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
func (p ArgVal) ArgType() d.TyPrime            { return p.Arg().TypePrime() }
func (p ArgVal) TypeFnc() TyFnc                { return Argument }
func (p ArgVal) TypePrime() d.TyPrime          { return p.ArgType() }

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
func (a ArgSet) TypeFnc() TyFnc { return Argument | Vector }
func (a ArgSet) TypePrime() d.TyPrime {
	var f = d.BitFlag(uint(0))
	for _, arg := range a.Args() {
		f = f.Concat(arg.TypePrime().Flag())
	}
	f = f | d.Vector.TypePrime().Flag()
	return d.TyPrime(f)
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
		if oargs[i] != nil && d.Nil.TypePrime().Flag().Match(oargs[i].TypePrime().Flag()) {
			an[i] = oargs[i]
		}
		// copy new arguments to return set, if any are set at this
		// position. overwrite old arguments in case any where set at
		// this position.
		if args[i] != nil && d.Nil.TypePrime().Flag().Match(args[i].TypePrime().Flag()) {
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
func (p ParamVal) ArgType() d.BitFlag            { return p.Arg().TypePrime().Flag() }
func (p ParamVal) AccType() d.BitFlag            { return p.Acc().TypePrime().Flag() }
func (p ParamVal) Eval(a ...d.Primary) d.Primary { return NewPair(p.Acc(), p.Arg()) }
func (p ParamVal) Both() (Value, Value)          { return p.Pair().Both() }
func (p ParamVal) Empty() bool {
	l, r := p.Pair().Both()
	return ElemEmpty(l) && ElemEmpty(r)
}
func (p ParamVal) TypePrime() d.TyPrime { return p.Pair().TypePrime() }
func (p ParamVal) TypeFnc() TyFnc       { return Parameter }

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
	case acc.Eval().TypePrime().Flag().Match(d.Symbolic.TypePrime().Flag()):
		ps.Sort(d.String)
	case acc.Eval().TypePrime().Flag().Match(d.Natural.TypePrime().Flag()):
		ps.Sort(d.Natural)
	case acc.Eval().TypePrime().Flag().Match(d.Integer.TypePrime().Flag()):
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
		ps.Sort(d.TyPrime(arg.Acc().Eval().TypePrime()))
	}
	parameters := NewParameters(ps...)
	return parameters.Parms(), parameters
}
func (a ParamSet) TypeFnc() TyFnc { return Parameter | Sets }
func (a ParamSet) TypePrime() d.TyPrime {
	var f = d.BitFlag(0).Flag()
	for _, pair := range a.Pairs() {
		f = f | pair.TypePrime().Flag()
	}
	return d.TyPrime(f | d.Vector.TypePrime().Flag())
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
	ps.Sort(d.TyPrime(praed[0].Left().Eval().TypePrime()))
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
