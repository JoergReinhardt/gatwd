package types

///
//// Functional higher order types ////
// takes a state and advances it. returns the next state fn to run
type FnType BitFlag

func (t FnType) Type() Typed   { return t.Type() }
func (t FnType) Flag() BitFlag { return BitFlag(t) }
func (t FnType) Uint() uint    { return uint(t) }

//go:generate stringer -type=FnType
const (
	StateFunc FnType = 1 << iota
	ParamFunc
	ConstFunc   // func() Data
	UnaryFunc   // func(Data) Data
	BinaryFunc  // func(Data,Data) Data
	NnaryFunc   // func(...Data) Data
	PrediFunc   // func(Data) bool
	SetoidFunc  // func(...Data) bool <-- Equal(Setoid) bool
	ApplFunctor // applicative functor
)

//// INTERNAL FUNCTIONAL TYPES ////
type (
	cell struct { // idx | key :: data
		Attribute
		Data
	}
	///// FUNCTION TYPES //////
	cons      ConstFnc
	unc       UnaryFnc
	bnc       BinaryFnc
	fnc       NnaryFnc
	ConstFnc  func() Data
	UnaryFnc  func(d Data) Data
	BinaryFnc func(a, b Data) Data
	NnaryFnc  func(...Data) Data
	Argument  Data                          // type alias enables more explicity
	Attribute ConstFnc                      // attributes may carry additional behaviour
	Predicate func(...Data) bool            // <- dessicion fnc in a conditional
	Condition func(conditions ...Data) Data // <- evaluates conditionally, based on praedicate
	Guard     func(fn fnc, p Predicate)
	chain     []Data
	FLagSet   []BitFlag
	ParamSet  []Attribute
)

//// INTERFACE IMPLEMENTING ENCLOSURES /////

///////////// FUNCTIONAL STANDALONE BASE UTILS /////////////
// element empty praedicate
func composedEmpty(dat Data) bool {
	if fmatch(dat.Flag(), flag(Composed)) {
		if dat != nil {
			if !dat.(Collected).Empty() {
				return false
			}
		}
	}
	return true
}
func nullableEmpty(dat Data) bool {
	if fmatch(dat.Flag(), flag(Nullable)) {
		if dat != nil {
			if !fmatch(dat.Flag(), flag(Nil)) {
				return false
			}
		}
	}
	return true
}
func empty(d ...Data) bool {
	for _, dat := range d {
		if d != nil {
			if !composedEmpty(dat) || !nullableEmpty(dat) {
				return false
			}
		}
	}
	return true
}
func isNil(dat Data) bool {
	if fmatch(dat.Flag(), flag(Nil)) || dat != nil {
		return true
	}
	return false
}
func areNil(d ...Data) bool {
	for _, dat := range d {
		if dat == nil || isNil(dat) {
			return true
		}
	}
	return false
}

// return self
func identity(d Data) Data { return d }

// type aliases (s types/types.go)
// 'unc' --> unary, 'bnc' --> binary, cons --> func() Data
// 'fnc' --> 'func(...Data) Data', aka NnaryFnc
func arity(a int, fn fnc) fnc {
	return func(d ...Data) Data {
		if len(d) > a {
			return fn(d[:a]...)
		}
		return fn(d...)
	}
}

var ( // only pass n arguments on to closure
	unary   = func(fn fnc) fnc { return arity(1, fn) }
	binary  = func(fn fnc) fnc { return arity(2, fn) }
	trinary = func(fn fnc) fnc { return arity(3, fn) }
	quatern = func(fn fnc) fnc { return arity(4, fn) }
)

// integer generator
type count func() (int, count)

// parameter:
// 0 - initial value
// 1 - increment (decrement when param negative)
func conCount(d ...int) count {

	var out int = 0
	var init int = 0
	var step int = 1
	var cnt count

	switch len(d) {
	case 1: // initial count
		init = d[0]
	case 2: // stepsize.(
		step = d[1]
	}
	cnt = func() (int, count) {
		out = init
		init = init + step
		return out, func() (int, count) {
			out = init
			init = init + step
			return out, cnt
		}
	}
	return cnt
}
func stringer(d ...Data) Data {
	var str string
	for _, dat := range d {
		str = str + dat.String()
	}
	return conData(str)
}
func acceptDataTypes(t BitFlag, d ...Data) []Data {
	var fdat []Data
	if len(d) > 1 {
		fdat = make([]Data, 0, len(d))
		for _, dat := range d {
			if fmatch(t, dat.Flag()) {
				fdat = append(fdat, dat)
			}
		}
	}
	return fdat
}

// only pass arguments on, that match the type flag (may be concatenated)
func accept(t BitFlag, fn fnc) fnc {
	return func(d ...Data) Data {
		fdat := acceptDataTypes(t, d...)
		if len(fdat) > 0 {
			return fn(fdat...)
		}
		return nilVal{}
	}
}

// construct attribute. closes over eval method of data instance
func attr(d Data) Attribute { return conAttr(d) }

// construct parameter set
func attrs(dat ...Data) ParamSet {
	var ps = ParamSet{}
	for _, d := range dat {
		ps = append(ps, conAttr(d))
	}
	return ps
}

// construct [key|idx / value] attributed data
func elem(a Attribute, d Data) cell    { return cell{a, d} }
func nelem(a interface{}, d Data) cell { return cell{attr(conData(a)), d} }

// partial argrument application and it's helpers
func partial(fn fnc, preset ...Data) fnc {
	return func(later ...Data) Data {
		return fn(append(preset, later...)...)
	}
}
func reverse(d ...Data) []Data {
	l := len(d)
	var r = make([]Data, l)
	for i, dat := range d {
		r[l-1-i] = dat
	}
	return r
}
func reverseArgs(fn fnc) fnc {
	return func(d ...Data) Data {
		return fn(reverse(d...)...)
	}
}
func partialRight(fn fnc, preset ...Data) fnc {
	return func(later ...Data) Data {
		return fn(append(later, preset...)...)
	}
}
func not(p Predicate) Predicate {
	return func(c ...Data) bool {
		return !p(c...)
	}
}
func when(p Predicate, fn fnc) fnc {
	return func(cond ...Data) Data {
		if p(cond...) {
			return fn(cond...)
		}
		return boolVal(false)
	}
}

// curry constructor :: fx() + gy() -> fxy(g())
func curry(fn fnc, ari int) fnc {

	var fan = arity(ari, fn)
	var args = []Data{}
	var entry func(...Data) Data

	entry = func(last ...Data) Data {
		args = append(args, last...)
		if len(args) >= ari {
			// if arguments complete => concat args & return call
			return fan(append(args, last...)...)
		}
		// call entrypoint & pass on return args
		return curry(entry, ari-1)
	}
	return entry
}
