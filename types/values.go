package types

import (
	"fmt"
	"math/big"
	"math/bits"
	"strconv"
	"strings"
	"time"
)

//////// INTERNAL TYPE SYSTEM ///////////
//
// intended to be accessable and extendable
type ValType flag

func (v ValType) Type() flag { return flag(v) }

//go:generate stringer -type=ValType
const (
	Nil  ValType = 1
	Bool ValType = 1 << iota
	Int
	Int8
	Int16
	Int32
	BigInt
	Uint
	Uint8
	Uint16
	Uint32
	Float
	Flt32
	BigFlt
	Ratio
	Imag
	Imag64
	Byte
	Rune
	Bytes
	String
	Time
	Duration
	Error
	// SLICE BASED COLLECTIONS //
	// "...life is nothing but distribution indifferences in a semi
	// permeably compartimented solution..." cell's to contain stuff, and
	// cells  are important, is what i'm saying here!
	Attr   // identitys, arity,  predicates, attribute accessors...
	Cell   // general thing to contain things and stuff...
	Chain  // [Value]
	List   // ordered, indexed, monotyped values
	AtList // ordered, indexed, with search/sort attributation
	UniSet // unique, monotyped values
	AtSet  // unique, attribute mapped, monotyped values (aka map) [attr,val]
	Record // unique, multityped, attributed, mapped, type declared values
	// LINKED COLLECTIONS // (also slice based, but pretend not to)
	Link  // nodes referencing previous, next node and nested value (possibly nested)
	DLink // nodes referencing previous, next node and nested value (possibly nested)
	Tuple // references a head value and nest of tail values
	Node  // node of a tree, or liked list
	Tree  // nodes referencing parent, root and a value of contained node(s)

	Nullable = Nil | Bool | Int | Int8 | Int16 | Int32 | BigInt | Uint |
		Uint8 | Uint16 | Uint32 | Float | Flt32 | BigFlt | Ratio | Imag |
		Imag64 | Byte | Rune | Bytes | String | Time | Duration | Error
	Elements = Attr | Cell
	Indices  = Chain | List | AtList
	Sets     = UniSet | AtSet | Record
	Links    = Link | DLink | Tuple | Node | Tree // Consumeables
	Composed = Elements | Indices | Sets | Links
	Natives  = Nullable | Composed
	Mask     = 0xFFFFFFFFFFFFFFFF ^ Natives
)

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
	uint8Val  uint8
	uint16Val uint16
	uint32Val uint32
	fltVal    float64
	flt32Val  float32
	imagVal   complex128
	imag64Val complex64
	byteVal   byte
	runeVal   rune
	bytesVal  []byte
	strVal    string
	bigIntVal big.Int
	bigFltVal big.Float
	ratioVal  big.Rat
	timeVal   time.Time
	duraVal   time.Duration
	errorVal  struct{ v error }
	//////
	attribute Evaluable
	slice     []Evaluable
)

func Make(vals ...interface{}) (rval Evaluable) {
	var val interface{}
	if len(vals) == 0 {
		return nilVal{}
	}
	if len(vals) > 1 {
		sl := newSlice()
		for _, val := range vals {
			val = val
			sl = sliceAppend(sl, Make(val))
		}
		return sl
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
	case *big.Int:
		v := bigIntVal(*val.(*big.Int))
		rval = &v
	case *big.Float:
		v := bigFltVal(*val.(*big.Float))
		rval = &v
	case *big.Rat:
		v := ratioVal(*val.(*big.Rat))
		rval = &v
	case []Evaluable:
		rval = slice(val.([]Evaluable))
	case FnType, ValType, Typed:
		rval = flag(val.(ValType))
	}
	return rval
}
func newNull(t Typed) (val Evaluable) {
	switch {
	case Nil.Type().Match(t):
		return nilVal{}
	case Bool.Type().Match(t):
		return Make(false)
	case Int.Type().Match(t):
		return Make(0)
	case Int8.Type().Match(t):
		return Make(int8(0))
	case Int16.Type().Match(t):
		return Make(int16(0))
	case Int32.Type().Match(t):
		return Make(int32(0))
	case BigInt.Type().Match(t):
		return Make(big.NewInt(0))
	case Uint.Type().Match(t):
		return Make(uint(0))
	case Uint16.Type().Match(t):
		return Make(uint16(0))
	case Uint32.Type().Match(t):
		return Make(uint32(0))
	case Float.Type().Match(t):
		return Make(float64(0))
	case Flt32.Type().Match(t):
		return Make(float32(0))
	case BigFlt.Type().Match(t):
		return Make(big.NewFloat(0))
	case Ratio.Type().Match(t):
		return Make(big.NewRat(1, 1))
	case Imag.Type().Match(t):
		return Make(complex128(float64(0)))
	case Imag64.Type().Match(t):
		return Make(complex64(float32(0)))
	case Byte.Type().Match(t):
		var b byte = 0
		return Make(b)
	case Bytes.Type().Match(t):
		var b []byte = []byte{}
		return Make(b)
	case Rune.Type().Match(t):
		var b rune = ' '
		return Make(b)
	case String.Type().Match(t):
		s := " "
		return Make(s)
	case Error.Type().Match(t):
		var e error = fmt.Errorf("")
		return Make(e)
	case t.Type().Match(BigInt):
		v := &big.Int{}
		return Make(v)
	case t.Type().Match(BigFlt):
		v := &big.Float{}
		return Make(v)
	case t.Type().Match(Ratio):
		v := &big.Rat{}
		return Make(v)
	}
	return val
}

/// Type
func (nilVal) Type() flag      { return Nil.Type() }
func (v flag) Type() flag      { return v }
func (v boolVal) Type() flag   { return Bool.Type() }
func (v intVal) Type() flag    { return Int.Type() }
func (v int8Val) Type() flag   { return Int8.Type() }
func (v int16Val) Type() flag  { return Int16.Type() }
func (v int32Val) Type() flag  { return Int32.Type() }
func (v uintVal) Type() flag   { return Uint.Type() }
func (v uint8Val) Type() flag  { return Uint8.Type() }
func (v uint16Val) Type() flag { return Uint16.Type() }
func (v uint32Val) Type() flag { return Uint32.Type() }
func (v bigIntVal) Type() flag { return BigInt.Type() }
func (v fltVal) Type() flag    { return Float.Type() }
func (v flt32Val) Type() flag  { return Flt32.Type() }
func (v bigFltVal) Type() flag { return BigFlt.Type() }
func (v imagVal) Type() flag   { return Imag.Type() }
func (v imag64Val) Type() flag { return Imag64.Type() }
func (v ratioVal) Type() flag  { return Ratio.Type() }
func (v runeVal) Type() flag   { return Rune.Type() }
func (v byteVal) Type() flag   { return Byte.Type() }
func (v bytesVal) Type() flag  { return Bytes.Type() }
func (v strVal) Type() flag    { return String.Type() }
func (v timeVal) Type() flag   { return Time.Type() }
func (v duraVal) Type() flag   { return Duration.Type() }
func (v slice) Type() flag     { return Chain.Type() }
func (v errorVal) Type() flag  { return Error.Type() }

/// VALUE
func (v nilVal) Eval() Evaluable    { return v }
func (t flag) Eval() Evaluable      { return t }
func (v boolVal) Eval() Evaluable   { return v }
func (v intVal) Eval() Evaluable    { return v }
func (v int8Val) Eval() Evaluable   { return v }
func (v int16Val) Eval() Evaluable  { return v }
func (v int32Val) Eval() Evaluable  { return v }
func (v bigIntVal) Eval() Evaluable { return v }
func (v uintVal) Eval() Evaluable   { return v }
func (v uint8Val) Eval() Evaluable  { return v }
func (v uint16Val) Eval() Evaluable { return v }
func (v uint32Val) Eval() Evaluable { return v }
func (v imagVal) Eval() Evaluable   { return v }
func (v imag64Val) Eval() Evaluable { return v }
func (v bigFltVal) Eval() Evaluable { return v }
func (v flt32Val) Eval() Evaluable  { return v }
func (v fltVal) Eval() Evaluable    { return v }
func (v ratioVal) Eval() Evaluable  { return v }
func (v byteVal) Eval() Evaluable   { return v }
func (v runeVal) Eval() Evaluable   { return v }
func (v bytesVal) Eval() Evaluable  { return v }
func (v strVal) Eval() Evaluable    { return v }
func (v slice) Eval() Evaluable     { return v }
func (v errorVal) Eval() Evaluable  { return v }
func (v timeVal) Eval() Evaluable   { return v }
func (v duraVal) Eval() Evaluable   { return v }

/// COPY
func (t flag) Copy() Evaluable      { n := t; return n }
func (n nilVal) Copy() Evaluable    { return nilVal(struct{}{}) }
func (v boolVal) Copy() Evaluable   { var r boolVal = v; return r }
func (v int32Val) Copy() Evaluable  { var r int32Val = v; return r }
func (v int16Val) Copy() Evaluable  { var r int16Val = v; return r }
func (v int8Val) Copy() Evaluable   { var r int8Val = v; return r }
func (v intVal) Copy() Evaluable    { var r intVal = v; return r }
func (v bigIntVal) Copy() Evaluable { var r bigIntVal = v; return r }
func (v uint32Val) Copy() Evaluable { var r uint32Val = v; return r }
func (v uint16Val) Copy() Evaluable { var r uint16Val = v; return r }
func (v uint8Val) Copy() Evaluable  { var r uint8Val = v; return r }
func (v uintVal) Copy() Evaluable   { var r uintVal = v; return r }
func (v fltVal) Copy() Evaluable    { var r fltVal = v; return r }
func (v flt32Val) Copy() Evaluable  { var r flt32Val = v; return r }
func (v bigFltVal) Copy() Evaluable { var r bigFltVal = v; return r }
func (v imagVal) Copy() Evaluable   { var r imagVal = v; return r }
func (v imag64Val) Copy() Evaluable { var r imag64Val = v; return r }
func (v ratioVal) Copy() Evaluable  { var r ratioVal = v; return r }
func (v byteVal) Copy() Evaluable   { var r byteVal = v; return r }
func (v runeVal) Copy() Evaluable   { var r runeVal = v; return r }
func (v bytesVal) Copy() Evaluable  { var r bytesVal = v; return r }
func (v strVal) Copy() Evaluable    { var r strVal = v; return r }
func (v timeVal) Copy() Evaluable   { var r timeVal = v; return r }
func (v duraVal) Copy() Evaluable   { var r duraVal = v; return r }
func (v slice) Copy() Evaluable     { var ret = []Evaluable{}; return slice(append(ret, v)) }
func (v errorVal) Copy() Evaluable  { var r errorVal = v; return r }

/// STRING
func (nilVal) String() string      { return Nil.String() }
func (v errorVal) String() string  { return v.v.Error() }
func (v errorVal) Error() error    { return v.v }
func (v boolVal) String() string   { return strconv.FormatBool(bool(v)) }
func (v intVal) String() string    { return strconv.Itoa(int(v)) }
func (v int8Val) String() string   { return strconv.Itoa(int(v)) }
func (v int16Val) String() string  { return strconv.Itoa(int(v)) }
func (v int32Val) String() string  { return strconv.Itoa(int(v)) }
func (v uintVal) String() string   { return strconv.Itoa(int(v)) }
func (v uint8Val) String() string  { return strconv.Itoa(int(v)) }
func (v uint16Val) String() string { return strconv.Itoa(int(v)) }
func (v uint32Val) String() string { return strconv.Itoa(int(v)) }
func (v byteVal) String() string   { return strconv.Itoa(int(v)) }
func (v runeVal) String() string   { return string(v) }
func (v bytesVal) String() string  { return string(v) }
func (v strVal) String() string    { return string(v) }
func (v strVal) Key() string       { return string(v) }
func (v timeVal) String() string   { return time.Time(v).String() }
func (v duraVal) String() string   { return time.Duration(v).String() }
func (v bigIntVal) String() string { return ((*big.Int)(&v)).String() }
func (v ratioVal) String() string  { return ((*big.Rat)(&v)).String() }
func (v bigFltVal) String() string { return ((*big.Float)(&v)).String() }
func (v fltVal) String() string {
	return strconv.FormatFloat(float64(v), 'G', -1, 64)
}
func (v flt32Val) String() string {
	return strconv.FormatFloat(float64(v), 'G', -1, 32)
}
func (v imagVal) String() string {
	return strconv.FormatFloat(float64(real(v)), 'G', -1, 64) + " + " +
		strconv.FormatFloat(float64(imag(v)), 'G', -1, 64) + "i"
}
func (v imag64Val) String() string {
	return strconv.FormatFloat(float64(real(v)), 'G', -1, 32) + " + " +
		strconv.FormatFloat(float64(imag(v)), 'G', -1, 32) + "i"
}
func (v flag) String() string {
	if uint(bits.OnesCount(v.Uint())) == 1 {
		return ValType(v).String()
	}
	len := uint(flen(v))
	str := &strings.Builder{}
	var err error
	var u, i uint
	for u < uint(Tree) {
		if v.Type().Match(ValType(u)) {
			_, err = (*str).WriteString(ValType(u).String())
			if i < len-1 {
				_, err = (*str).WriteString(" | ")
			}
		}
		i = i + 1
		u = uint(1) << i
	}
	if err != nil {
		return "ERROR: could not decompose value type name to string"
	}
	return str.String()
}
func (v slice) String() string {
	var err error
	str := &strings.Builder{}
	_, err = (*str).WriteString("[")
	for i, val := range v.Slice() {
		_, err = (*str).WriteString(val.String())
		if i < v.Len()-1 {
			(*str).WriteString(", ")
		}
	}
	_, err = (*str).WriteString("]")
	if err != nil {
		return "ERROR: could not concatenate slice values to string"
	}
	return str.String()
}
