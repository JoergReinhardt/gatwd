package types

import (
	"fmt"
	"math/big"
	"math/bits"
	"time"
)

//////// INTERNAL TYPE SYSTEM ///////////
//
// intended to be accessable and extendable
type ValType flag

func (v ValType) Type() flag { return flag(v) }

//go:generate stringer -type=ValType
const (
	Nil      ValType = 0
	MetaType ValType = 1
	// FLAT VALUES
	Bool ValType = 1 << iota
	Int
	Int8
	Int16
	Int32
	BigInt
	Uint
	Uint16
	Uint32
	Float
	Flt32
	BigFlt
	Imag
	Imag64
	Byte
	Bytes
	String
	Time
	Duration
	Error
	// SLICE BASED COLLECTIONS //
	// "...life is nothing but distribution indifferences in a semi
	// permeably compartimented solution..." cell's to contain stuff, and
	// cells  are important, is what i'm saying here!
	Cell     // general thing to contain things and stuff...
	Attr     // identitys, arity,  predicates, attribute accessors...
	Chain    // [Value]
	List     // ordered, indexed, monotyped values
	Link     // nodes referencing previous, next node and nested value (possibly nested)
	DLink    // nodes referencing previous, next node and nested value (possibly nested)
	AttrList // ordered, indexed, with search/sort attributation
	UniSet   // unique, monotyped values
	AttrSet  // unique, attribute mapped, monotyped values (aka map) [attr,val]
	Record   // unique, multityped, attributed, mapped, type declared values
	// LINKED COLLECTIONS // (also slice based, but pretend not to)
	Tuple // references a head value and nest of tail values
	Node  // node of a tree, or liked list
	Tree  // nodes referencing parent, root and a value of contained node(s)
	// FUNCTIONS
	Function
	///////////
	NATIVE

	// flat value types
	Unary = Nil | Bool | Int | Int8 | Int16 | Int32 | Uint | Uint16 |
		Uint32 | Float | Flt32 | Imag | Imag64 | Byte | Bytes | String |
		Time | Duration | Error

	// combined value types
	Nary = Ordered | Mapped

	// types that come with type constuctors
	Nullable = Unary | Chain

	// Slice() []Value
	Ordered = Chain | List | AttrList

	// Next() Value
	Chained = Tuple | Node | Tree

	// Get(attr) Attribute
	Mapped = UniSet | AttrSet | Record

	// higher order types combined from a finite set of other types, defined by a signature
	SumTypes = Chain | List | UniSet

	// Product Types are combinations of arbitrary other types in arbitrary combination
	ProductTypes = List | AttrList | AttrSet | Record | Tuple | Node | Tree
)

var HIGH_MASK uint
var LOW_MASK uint
var nativesCount int

func initMasks() {
	var i, tmp uint
	for i = 0; i < 64; i++ {
		tmp = 1 << i
		if tmp < uint(NATIVE) {
			nativesCount = int(i) + 2
			LOW_MASK = LOW_MASK | tmp
			continue
		}
		HIGH_MASK = HIGH_MASK | tmp
	}
}

//////// INTERNAL TYPES /////////////
type (
	flag      uint
	nilVal    struct{}
	boolVal   bool
	intVal    int
	int8Val   int8
	int16Val  int16
	int32Val  int32
	uintVal   uint
	uint16Val uint16
	uint32Val uint32
	fltVal    float64
	flt32Val  float32
	imagVal   complex128
	imag64Val complex64
	byteVal   byte
	bytesVal  []byte
	strVal    string
	bigInt    struct{ v *big.Int }
	timeVal   time.Time
	duraVal   time.Duration
	errorVal  struct{ v error }
	//////
	attribute  Value
	slice      []Value
	collection struct{ s slice }
)

/// Type
func (nilVal) Type() flag       { return Nil.Type() }
func (v boolVal) Type() flag    { return Bool.Type() }
func (v intVal) Type() flag     { return Int.Type() }
func (v int8Val) Type() flag    { return Int8.Type() }
func (v int16Val) Type() flag   { return Int16.Type() }
func (v int32Val) Type() flag   { return Int32.Type() }
func (v uintVal) Type() flag    { return Uint.Type() }
func (v uint16Val) Type() flag  { return Uint16.Type() }
func (v uint32Val) Type() flag  { return Uint32.Type() }
func (v fltVal) Type() flag     { return Float.Type() }
func (v flt32Val) Type() flag   { return Flt32.Type() }
func (v imagVal) Type() flag    { return Imag.Type() }
func (v imag64Val) Type() flag  { return Imag64.Type() }
func (v byteVal) Type() flag    { return Byte.Type() }
func (v bytesVal) Type() flag   { return Bytes.Type() }
func (v strVal) Type() flag     { return String.Type() }
func (v timeVal) Type() flag    { return Time.Type() }
func (v duraVal) Type() flag    { return Duration.Type() }
func (v slice) Type() flag      { return Chain.Type() }
func (v errorVal) Type() flag   { return Error.Type() }
func (s collection) Type() flag { return Ordered.Type() }

/// VALUE
func (v nilVal) Eval() Value     { return v }
func (v boolVal) Eval() Value    { return v }
func (v intVal) Eval() Value     { return v }
func (v int8Val) Eval() Value    { return v }
func (v int16Val) Eval() Value   { return v }
func (v int32Val) Eval() Value   { return v }
func (v uintVal) Eval() Value    { return v }
func (v uint16Val) Eval() Value  { return v }
func (v uint32Val) Eval() Value  { return v }
func (v imagVal) Eval() Value    { return v }
func (v imag64Val) Eval() Value  { return v }
func (v flt32Val) Eval() Value   { return v }
func (v fltVal) Eval() Value     { return v }
func (v byteVal) Eval() Value    { return v }
func (v bytesVal) Eval() Value   { return v }
func (v strVal) Eval() Value     { return v }
func (v slice) Eval() Value      { return v }
func (v errorVal) Eval() Value   { return v }
func (v collection) Eval() Value { return v }
func (v timeVal) Eval() Value    { return v }
func (v duraVal) Eval() Value    { return v }

/// REFERENCE
func (v nilVal) Ref() Value     { return nil }
func (v boolVal) Ref() Value    { return &v }
func (v intVal) Ref() Value     { return &v }
func (v int8Val) Ref() Value    { return &v }
func (v int16Val) Ref() Value   { return &v }
func (v int32Val) Ref() Value   { return &v }
func (v uintVal) Ref() Value    { return &v }
func (v uint16Val) Ref() Value  { return &v }
func (v uint32Val) Ref() Value  { return &v }
func (v fltVal) Ref() Value     { return &v }
func (v flt32Val) Ref() Value   { return &v }
func (v imagVal) Ref() Value    { return &v }
func (v imag64Val) Ref() Value  { return &v }
func (v byteVal) Ref() Value    { return &v }
func (v bytesVal) Ref() Value   { return &v }
func (v strVal) Ref() Value     { return &v }
func (v slice) Ref() Value      { return &v }
func (v errorVal) Ref() Value   { return &v }
func (v collection) Ref() Value { return &v }
func (v timeVal) Ref() Value    { return &v }
func (v duraVal) Ref() Value    { return &v }

/// DEREFERENCE
func (v nilVal) DeRef() Value     { return nil }
func (v boolVal) DeRef() Value    { inst := *(v.Eval().(*boolVal)); return inst }
func (v intVal) DeRef() Value     { inst := *(v.Eval().(*intVal)); return inst }
func (v int16Val) DeRef() Value   { inst := *(v.Eval().(*int16Val)); return inst }
func (v int8Val) DeRef() Value    { inst := *(v.Eval().(*int8Val)); return inst }
func (v int32Val) DeRef() Value   { inst := *(v.Eval().(*int32Val)); return inst }
func (v uintVal) DeRef() Value    { inst := *(v.Eval().(*uintVal)); return inst }
func (v uint16Val) DeRef() Value  { inst := *(v.Eval().(*uint16Val)); return inst }
func (v uint32Val) DeRef() Value  { inst := *(v.Eval().(*uint32Val)); return inst }
func (v fltVal) DeRef() Value     { inst := *(v.Eval().(*fltVal)); return inst }
func (v flt32Val) DeRef() Value   { inst := *(v.Eval().(*flt32Val)); return inst }
func (v imagVal) DeRef() Value    { inst := *(v.Eval().(*imagVal)); return inst }
func (v imag64Val) DeRef() Value  { inst := *(v.Eval().(*imag64Val)); return inst }
func (v byteVal) DeRef() Value    { inst := *(v.Eval().(*byteVal)); return inst }
func (v bytesVal) DeRef() Value   { inst := *(v.Eval().(*bytesVal)); return inst }
func (v strVal) DeRef() Value     { inst := *(v.Eval().(*strVal)); return inst }
func (v timeVal) DeRef() Value    { inst := *(v.Eval().(*timeVal)); return inst }
func (v duraVal) DeRef() Value    { inst := *(v.Eval().(*duraVal)); return inst }
func (v slice) DeRef() Value      { inst := *(v.Eval().(*slice)); return inst }
func (v errorVal) DeRef() Value   { inst := *(v.Eval().(*errorVal)); return inst }
func (v collection) DeRef() Value { inst := *(v.Eval().(*collection)); return inst }

/// COPY
func (v int32Val) Copy() Value       { var r int32Val = v; return r }
func (v int16Val) Copy() Value       { var r int16Val = v; return r }
func (v int8Val) Copy() Value        { var r int8Val = v; return r }
func (v intVal) Copy() Value         { var r intVal = v; return r }
func (n nilVal) Copy() Value         { return nilVal(struct{}{}) }
func (v fltVal) Copy() Value         { var r fltVal = v; return r }
func (v uint32Val) Copy() Value      { var r uint32Val = v; return r }
func (v uint16Val) Copy() Value      { var r uint16Val = v; return r }
func (v uintVal) Copy() Value        { var r uintVal = v; return r }
func (v boolVal) Copy() Value        { var r boolVal = v; return r }
func (v flt32Val) Copy() Value       { var r flt32Val = v; return r }
func (v imagVal) Copy() Value        { var r imagVal = v; return r }
func (v imag64Val) Copy() Value      { var r imag64Val = v; return r }
func (v byteVal) Copy() Value        { var r byteVal = v; return r }
func (v bytesVal) Copy() Value       { var r bytesVal = v; return r }
func (v strVal) Copy() Value         { var r strVal = v; return r }
func (v timeVal) Copy() Value        { var r timeVal = v; return r }
func (v duraVal) Copy() Value        { var r duraVal = v; return r }
func (v slice) Copy() Value          { var ret = []Value{}; return slice(append(ret, v)) }
func (v errorVal) Copy() Value       { var r errorVal = v; return r }
func (s collection) Copy() (v Value) { return collection{s.s.Copy().(slice)} }

func New(vals ...interface{}) (rval Value) {
	sli := make([]Value, 0, len(vals))
	for _, val := range vals {
		sli = append(sli, Make(val).Ref())
	}
	return Make(sli).Ref()
}
func Make(vals ...interface{}) (rval Value) {
	var val interface{}
	if len(vals) > 1 {
		for _, val := range vals {
			val = val
		}
	}
	val = vals[0]
	switch val.(type) {
	case bool:
		rval = boolVal(val.(bool))
	case int, int64:
		rval = intVal(val.(int))
	case int8:
		rval = int8Val(val.(int8))
	case int16:
		rval = int16Val(val.(int16))
	case int32:
		rval = int32Val(val.(int32))
	case uint, uint64:
		rval = uintVal(val.(uint))
	case uint16:
		rval = uint16Val(val.(uint16))
	case uint32:
		rval = int32Val(val.(int32))
	case float32:
		rval = flt32Val(val.(float32))
	case float64:
		rval = fltVal(val.(float64))
	case complex64:
		rval = imagVal(val.(complex64))
	case complex128:
		rval = imagVal(val.(complex128))
	case byte:
		rval = byteVal(val.(byte))
	case []byte:
		rval = bytesVal(val.([]byte))
	case string:
		rval = strVal(val.(string))
	case error:
		rval = errorVal{val.(error)}
	case time.Time:
		rval = timeVal(val.(time.Time))
	case time.Duration:
		rval = duraVal(val.(time.Duration))
	case []Value:
		rval = slice(val.([]Value))
	case FnType, ValType, Typed:
		rval = flag(val.(ValType))
	}
	return rval
}
func newNull(t Typed) (val Value) {
	switch {
	case Nil.Type().match(t):
		return nilVal{}
	case Bool.Type().match(t):
		return Make(false)
	case Int.Type().match(t):
		return Make(0)
	case Int8.Type().match(t):
		return Make(int8(0))
	case Int16.Type().match(t):
		return Make(int16(0))
	case Int32.Type().match(t):
		return Make(int32(0))
	case Uint.Type().match(t):
		return Make(uint(0))
	case Uint16.Type().match(t):
		return Make(uint16(0))
	case Uint32.Type().match(t):
		return Make(uint32(0))
	case Float.Type().match(t):
		return Make(float64(0))
	case Flt32.Type().match(t):
		return Make(float32(0))
	case Imag.Type().match(t):
		return Make(complex128(float64(0)))
	case Imag64.Type().match(t):
		return Make(complex64(float32(0)))
	case Byte.Type().match(t):
		var b byte = 0
		return Make(b)
	case Bytes.Type().match(t):
		var b []byte = []byte{}
		return Make(b)
	case String.Type().match(t):
		s := " "
		return Make(s)
	case Error.Type().match(t):
		var e error = fmt.Errorf("")
		return Make(e)
	case t.Type().meta():
	}
	return val
}

//// BOUND TYPE FLAG METHODS ////
func (t flag) uint() uint          { return fuint(t) }
func (t flag) len() int            { return flen(t) }
func (t flag) count() int          { return fcount(t) }
func (t flag) least() int          { return fleast(t) }
func (t flag) most() int           { return fmost(t) }
func (t flag) reverse() flag       { return frev(t) }
func (t flag) rotate(n int) flag   { return frot(t, n) }
func (t flag) toggle(v Typed) flag { return ftog(t, v) }
func (t flag) concat(v Typed) flag { return fconc(t, v) }
func (t flag) mask(v Typed) flag   { return fmask(t, v) }
func (t flag) match(v Typed) bool  { return fmatch(t, v) }
func (t flag) meta() bool          { return fmeta(t) }

// ...is a typed value all by itself
func (t flag) Type() flag   { return t }
func (t flag) Eval() Value  { return t }
func (t flag) Ref() Value   { return &t }
func (t flag) DeRef() Value { inst := t; return inst }
func (t flag) Copy() Value  { n := t; return n }

///// FREE TYPE FLAG METHOD IMPLEMENTATIONS /////
func fmatch(t flag, v Typed) bool {
	if t.uint()&v.Type().uint() != 0 {
		return true
	}
	return false
}
func fmeta(t flag) bool {
	if t.count() <= 1 {
		return false
	}
	return true
}
func fsplit(f flag) (left, right flag) {
	left = frot(
		f,
		flen(
			flag(NATIVE),
		),
	)
	return left, f
}
func fshow(f flag) string {
	return fmt.Sprintf("%64b\n", f)
}
func fuint(t flag) uint          { return uint(t) }
func flen(t flag) int            { return bits.Len(uint(t)) }
func fcount(t flag) int          { return bits.OnesCount(uint(t)) }
func fleast(t flag) int          { return bits.TrailingZeros(uint(t)) + 1 }
func fmost(t flag) int           { return bits.LeadingZeros(uint(t)) - 1 }
func frev(t flag) flag           { return flag(bits.Reverse(uint(t))) }
func frot(t flag, n int) flag    { return flag(bits.RotateLeft(uint(t), n)) }
func ftog(t flag, v Typed) flag  { return flag(uint(t) ^ v.Type().uint()) }
func fconc(t flag, v Typed) flag { return flag(uint(t) | v.Type().uint()) }
func fmask(t flag, v Typed) flag { return flag(uint(t) &^ v.Type().uint()) }

///// HIGHER ORDER TYPES /////
//// CONSTRUCTORS ////
type Constructor func(...Value) Value

type Instance func() Value

// eight bits to mark type class
type TypeClass uint8

//go:generate stringer -type=TypeClass
const (
	metaTypeName TypeClass = 0 + iota
	dataConstructor
	typeConstructor
	typeConverter
	functionSignature
	functionBody
)

func lenID(s []flag) int   { return len(s) }
func lenMS(s []constr) int { return len(s) }

func less(a, b flag) bool { return less(b, a) }
func more(a, b flag) bool { return less(a, b) }
func lessID(a, b flag) bool {
	if a <= b {
		return true
	}
	return false
}
func swapID(a, b int, s []flag) []flag {
	tmp := s[a]
	s[a] = s[b]
	s[b] = tmp
	return s
}

type constr struct {
	id  flag        // reference to type id
	sig []flag      // type signature, param/retval def, etc...
	fnc Constructor // constructs instances of types, data, signatures...
}
type typeReg struct {
	id  []flag
	con []constr
}

func newTypeReg() *typeReg {
	return &typeReg{
		[]flag{},
		[]constr{},
	}
}

var types *typeReg

func initTypeReg() {
	types = newTypeReg()
}
