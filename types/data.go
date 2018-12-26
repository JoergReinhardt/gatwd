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
	Nil  ValType = 0
	Bool ValType = 1
	Int  ValType = 1 << iota
	Int8         // Int8 -> Int8
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
	Attr   // attribute special type
	Error  // let's do something sophisticated here...
	Tuple  // references a head value and nest of tail values
	List   // ordered, indexed, monotyped values
	Chain  // [Value]
	AtList // ordered, indexed, with search/sort attributation
	UniSet // unique, monotyped values
	AtSet  // unique, attribute mapped, monotyped values (aka map) [attr,val]
	Record // unique, multityped, attributed, mapped, type declared values
	Link   // nodes referencing previous, next node and nested value
	DLink  // nodes referencing previous, next node and nested value
	Node   // node of a tree, or liked list
	Tree   // nodes referencing parent, root and a value of contained node(s)

	Nullable = Nil | Bool | Int | Int8 | Int16 | Int32 | BigInt | Uint |
		Uint8 | Uint16 | Uint32 | Float | Flt32 | BigFlt | Ratio | Imag |
		Imag64 | Byte | Rune | Bytes | String | Time | Duration |
		Attr | Error

	Numbers = Bool | Int | Int8 | Int16 | Int32 | BigInt | Uint | Uint8 |
		Uint16 | Uint32 | Float | Flt32 | BigFlt | Ratio | Imag |
		Imag64

	Elements = Tuple | List
	Indices  = Chain | AtList
	Sets     = UniSet | AtSet | Record
	Links    = Link | DLink | Node | Tree // Consumeables
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
	attribute Data
	slice     []Data
)

func data(vals ...interface{}) (rval Data) {
	var val interface{}
	if len(vals) == 0 {
		return nilVal{}
	}
	if len(vals) > 1 {
		sl := newSlice()
		for _, val := range vals {
			val = val
			sl = slicePut(sl, data(val))
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
	case Data:
		rval = val.(Data)
	case []Data:
		rval = slice(val.([]Data))
	case FnType, ValType, Typed:
		rval = flag(val.(ValType))
	}
	return rval
}

//// GENERATE NULL VALUE OF EACH TYPE ////////
func newNull(t Typed) (val Data) {
	switch {
	case Nil.Type().Match(t):
		return nilVal{}
	case Bool.Type().Match(t):
		return data(false)
	case Int.Type().Match(t):
		return data(0)
	case Int8.Type().Match(t):
		return data(int8(0))
	case Int16.Type().Match(t):
		return data(int16(0))
	case Int32.Type().Match(t):
		return data(int32(0))
	case BigInt.Type().Match(t):
		return data(big.NewInt(0))
	case Uint.Type().Match(t):
		return data(uint(0))
	case Uint16.Type().Match(t):
		return data(uint16(0))
	case Uint32.Type().Match(t):
		return data(uint32(0))
	case Float.Type().Match(t):
		return data(float64(0))
	case Flt32.Type().Match(t):
		return data(float32(0))
	case BigFlt.Type().Match(t):
		return data(big.NewFloat(0))
	case Ratio.Type().Match(t):
		return data(big.NewRat(1, 1))
	case Imag.Type().Match(t):
		return data(complex128(float64(0)))
	case Imag64.Type().Match(t):
		return data(complex64(float32(0)))
	case Byte.Type().Match(t):
		var b byte = 0
		return data(b)
	case Bytes.Type().Match(t):
		var b []byte = []byte{}
		return data(b)
	case Rune.Type().Match(t):
		var b rune = ' '
		return data(b)
	case String.Type().Match(t):
		s := " "
		return data(s)
	case Error.Type().Match(t):
		var e error = fmt.Errorf("")
		return data(e)
	case t.Type().Match(BigInt):
		v := &big.Int{}
		return data(v)
	case t.Type().Match(BigFlt):
		v := &big.Float{}
		return data(v)
	case t.Type().Match(Ratio):
		v := &big.Rat{}
		return data(v)
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

func (nilVal) String() strVal      { return strVal(Nil.String()) }
func (v errorVal) String() strVal  { return strVal(v.v.Error()) }
func (v errorVal) Error() errorVal { return errorVal{v.v} }
func (v boolVal) String() strVal   { return strVal(strconv.FormatBool(bool(v))) }
func (v intVal) String() strVal    { return strVal(strconv.Itoa(int(v))) }
func (v int8Val) String() strVal   { return strVal(strconv.Itoa(int(v))) }
func (v int16Val) String() strVal  { return strVal(strconv.Itoa(int(v))) }
func (v int32Val) String() strVal  { return strVal(strconv.Itoa(int(v))) }
func (v uintVal) String() strVal   { return strVal(strconv.Itoa(int(v))) }
func (v uint8Val) String() strVal  { return strVal(strconv.Itoa(int(v))) }
func (v uint16Val) String() strVal { return strVal(strconv.Itoa(int(v))) }
func (v uint32Val) String() strVal { return strVal(strconv.Itoa(int(v))) }
func (v byteVal) String() strVal   { return strVal(strconv.Itoa(int(v))) }
func (v runeVal) String() strVal   { return strVal(string(v)) }
func (v bytesVal) String() strVal  { return strVal(string(v)) }
func (v strVal) String() strVal    { return strVal(string(v)) }
func (v strVal) Key() strVal       { return strVal(string(v)) }
func (v timeVal) String() strVal   { return strVal(time.Time(v).String()) }
func (v duraVal) String() strVal   { return strVal(time.Duration(v).String()) }
func (v bigIntVal) String() strVal { return strVal(((*big.Int)(&v)).String()) }
func (v ratioVal) String() strVal  { return strVal(((*big.Rat)(&v)).String()) }
func (v bigFltVal) String() strVal { return strVal(((*big.Float)(&v)).String()) }
func (v fltVal) String() strVal {
	return strVal(strconv.FormatFloat(float64(v), 'G', -1, 64))
}
func (v flt32Val) String() strVal {
	return strVal(strconv.FormatFloat(float64(v), 'G', -1, 32))
}
func (v imagVal) String() strVal {
	return strVal(strconv.FormatFloat(float64(real(v)), 'G', -1, 64) + " + " +
		strconv.FormatFloat(float64(imag(v)), 'G', -1, 64) + "i")
}
func (v imag64Val) String() strVal {
	return strVal(strconv.FormatFloat(float64(real(v)), 'G', -1, 32) + " + " +
		strconv.FormatFloat(float64(imag(v)), 'G', -1, 32) + "i")
}
func (v flag) String() strVal {
	if uint(bits.OnesCount(v.Uint())) == 1 {
		return strVal(ValType(v).String())
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
		return strVal("ERROR: could not decompose value type name to string")
	}
	return strVal(str.String())
}
func (v slice) String() strVal {
	var err error
	str := &strings.Builder{}
	_, err = (*str).WriteString("[")
	for i, val := range v.Slice() {
		_, err = (*str).WriteString(fmt.Sprintf("%v", val))
		if i < v.Len()-1 {
			(*str).WriteString(", ")
		}
	}
	_, err = (*str).WriteString("]")
	if err != nil {
		return "ERROR: could not concatenate slice values to string"
	}
	return strVal(str.String())
}
