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
	ConstFunc  // func() Data
	UnaryFunc  // func(Data) Data
	BinaryFunc // func(Data,Data) Data
	NnaryFunc  // func(...Data) Data
	PrediFunc  // func(Data) bool
	SetoidFunc // func(...Data) bool <-- Equal(Setoid) bool
	ApplicFunc // applicative functor
)

//// FUNCTION TYPES ////
type (
	// functional base types
	constFnc  func() Data // data.Eval() happens to be a constFn
	unaryFnc  func(d Data) Data
	binaryFnc func(a, b Data) Data
	naryFnc   func(...Data) Data
	// higher order type aliases
	cons constFnc
	unc  unaryFnc
	bnc  binaryFnc
	fnc  naryFnc
	// higher oder base types
	attrVal   constFnc // typesafe discrimination
	argVal    attrVal  // [name | idx] & data, passed in call
	paramVal  attrVal  // [name | idx] & data, passed in call
	retVal    attrVal  // [name | idx] & data returned from a function call
	accessor  func() (Data, Data)
	idxAcc    func() (int, Data)
	keyAcc    func() (string, Data)
	flagAcc   func() (BitFlag, Data)
	flagSet   []BitFlag
	paramSet  []paramVal
	argSet    []paramVal
	retValSet []retVal
	// higher order function types
	Generator func() (Data, Generator)   // contains different data each call
	Predicate func(...Data) bool         // eval truth value of data
	Condition func(Predicate, Data) Data // pass data only if enclosed pred. evals true, else Nil
)

// lower-, to  higher-order type mapping
func (constFnc) Flag() BitFlag  { return Constant.Flag() }
func (unaryFnc) Flag() BitFlag  { return Unary.Flag() }
func (binaryFnc) Flag() BitFlag { return Binary.Flag() }
func (naryFnc) Flag() BitFlag   { return Nary.Flag() }
func (cons) Flag() BitFlag      { return Constant.Flag() }
func (unc) Flag() BitFlag       { return Unary.Flag() }
func (bnc) Flag() BitFlag       { return Binary.Flag() }
func (fnc) Flag() BitFlag       { return Binary.Flag() }
func (attrVal) Flag() BitFlag   { return Attribute.Flag() }
func (argVal) Flag() BitFlag    { return Argument.Flag() }
func (paramVal) Flag() BitFlag  { return Parameter.Flag() }
func (retVal) Flag() BitFlag    { return Return.Flag() }
func (paramSet) Flag() BitFlag  { return Vector.Flag() | Parameter.Flag() }
func (argSet) Flag() BitFlag    { return Vector.Flag() | Argument.Flag() }
func (retValSet) Flag() BitFlag { return Vector.Flag() | Return.Flag() }
func (flagSet) Flag() BitFlag   { return Vector.Flag() | Flag.Flag() }

func (c cons) Eval() Data     { return c() }
func (c unc) Eval() Data      { return c }
func (c bnc) Eval() Data      { return c }
func (c fnc) Eval() Data      { return c }
func (c attrVal) Eval() Data  { return c() }
func (c argVal) Eval() Data   { return c() }
func (c paramVal) Eval() Data { return c() }
func (c retVal) Eval() Data   { return c() }

// set evaluation yields param- & retval sets data as chain
func (c paramSet) Eval() Data  { return c }
func (c retValSet) Eval() Data { return c }
func (c argSet) Eval() Data    { return c }

// evaluation of a flag-set yields bitwise OR of contained flags
func (c flagSet) Eval() Data {
	var flags BitFlag
	for _, flag := range c {
		flags = flags | flag
	}
	return flags
}
func (c cons) Call(...Data) Data  { return c() }
func (u unc) Call(d ...Data) Data { return u(d[0]) }
func (b bnc) Call(d ...Data) Data { return b(d[0], d[1]) }
func (n fnc) Call(d ...Data) Data { return n(d...) }

// generic accessor
func (a accessor) Eval() Data       { return a.Value() }
func (a accessor) Acc() Data        { acc, _ := a(); return acc }
func (a accessor) Value() Data      { _, val := a(); return val }
func (a accessor) AccType() BitFlag { return a.Acc().Flag() }
func conAcc(acc Data, val Data) accessor {
	return func() (acc Data, val Data) {
		return acc, val
	}
}

// accessable by index
func (a idxAcc) Eval() Data       { return a.Value() }
func (a idxAcc) Acc() intVal      { acc, _ := a(); return intVal(acc) }
func (a idxAcc) Value() Data      { _, val := a(); return val }
func (a idxAcc) AccType() BitFlag { return Int.Flag() }
func (a idxAcc) Idx() int         { return int(a.Acc()) }
func conAccIdx(idx int, val Data) idxAcc {
	return func() (idx int, val Data) {
		return idx, val
	}
}

// accessable by key
func (a keyAcc) Eval() Data       { return a.Value() }
func (a keyAcc) Acc() strVal      { acc, _ := a(); return strVal(acc) }
func (a keyAcc) Value() Data      { _, val := a(); return val }
func (a keyAcc) AccType() BitFlag { return String.Flag() }
func conAccKey(key string, val Data) keyAcc {
	return func() (key string, val Data) {
		return key, val
	}
}

// accessable by flag
func (a flagAcc) Eval() Data       { return a.Value() }
func (a flagAcc) Acc() BitFlag     { acc, _ := a(); return BitFlag(acc) }
func (a flagAcc) Value() Data      { _, val := a(); return val }
func (a flagAcc) AccType() BitFlag { return Flag.Flag() }
func (a flagAcc) Key() BitFlag     { return BitFlag(a.Acc()) }
func conAccFlag(key string, val Data) keyAcc {
	return func() (key string, val Data) {
		return key, val
	}
}

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
func limNary(a int, fn fnc) fnc {
	return func(d ...Data) Data {
		if len(d) > a {
			return fn(d[:a]...)
		}
		return fn(d...)
	}
}

var ( // only pass n arguments on to closure
	unary   = func(fn fnc) fnc { return fnc(limNary(1, fn)) }
	binary  = func(fn fnc) fnc { return fnc(limNary(2, fn)) }
	trinary = func(fn fnc) fnc { return fnc(limNary(3, fn)) }
	quatern = func(fn fnc) fnc { return fnc(limNary(4, fn)) }
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
func stringer(d ...Data) string {
	var str string
	for _, dat := range d {
		str = str + dat.String()
	}
	return str
}

// filters array of bit flags to only contain accepted types
func acceptDataTypes(t BitFlag, d ...Data) []Data {
	var fdat []Data
	if len(d) > 0 {
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
func conAttr(d Data) attrVal  { return attrVal(d.Eval) }
func conParm(d Data) paramVal { return paramVal(d.Eval) }
func paramSetToData(p paramSet) []Data {
	var data = []Data{}
	if len(p) == 0 {
		return []Data{nilVal{}}
	}
	for _, parm := range p {
		data = append(data, parm())
	}
	return data
}

// construct parameter set
func attrs(dat ...Data) paramSet {
	var ps = paramSet{}
	for _, d := range dat {
		ps = append(ps, conParm(d))
	}
	return ps
}

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

// strict curry operator exclusively deals with the case of currying one
// parameter of a binary function and returs a unary that expects the second
// parameter and will return the result of the composed call.
func curryBinary(fn bnc) func(Data) unc { // less specific: unc itself
	return func(a Data) unc { // less specific: Data itself
		return func(b Data) Data { // isomorph func() Data
			return fn(a, b)
		}
	}
}
func curryNary(fn fnc, arity int) Functional { // unc | bnc | fnc
	return fn
}

// default curry takes an optional arity parameter, or assumes the correct
// number of parameters has been passed and ommits the check.
func curry(fn fnc, arity ...int) fnc {
	var args = []Data{}
	var ari = arity
	var disp fnc
	var fan fnc
	//<- optionaly guard arity
	if len(arity) > 0 {
		fan = limNary(arity[0], fn)
	} else {
		fan = fnc(fn)
	}

	disp = fnc(func(last ...Data) Data {
		args = append(args, last...)
		if len(args) >= ari[0] {
			// if arguments complete => concat args & return call
			return fan(append(args, last...)...)
		}
		// call entrypoint & pass on return vals
		return curry(disp, ari[0]-1)
	})
	return disp
}
