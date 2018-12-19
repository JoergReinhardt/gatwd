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
type ValType Flag

func (v ValType) Type() Type { return Type(Flag(v)) }
func (v ValType) Flag() Flag { return Flag(v) }

//go:generate stringer -type=ValType
const (
	Nil ValType = 0
	// FLAT VALUES
	Bool ValType = 1 << iota
	Int
	Int8
	Int16
	Int32
	Uint
	Uint16
	Uint32
	Float
	Flt32
	Imag
	Imag64
	Byte
	Bytes
	String
	Error
	// SLICE BASED COLLECTIONS //
	Attr // identitys, arity,  predicates, attribute accessors...
	Pttr // value guaranueed to contain a pointer to the instance referenced by Value
	//Cell     // Element to contain other elements
	Slice      // [Value]
	List       // ordered, indexed, monotyped values
	MuliList   // ordered, indexed, multityped values
	AttrList   // ordered, indexed, with search/sort attributation
	RecordList // ordered, indexes list of records, search-/sortable by any field
	Set        // unique, monotyped values
	AttrSet    // unique, attribute mapped, monotyped values (aka map) [attr,val]
	MuliSet    // unique, attribute mapped, multityped values		 [attr,type,val]
	Record     // unique, multityped, attributed, mapped, type declared values
	// LINKED COLLECTIONS // (also slice based, but pretend not to)
	ChainedList // nodes referencing next node and value (possibly nested)
	LinkedList  // nodes referencing previous, next node and nested value (possibly nested)
	DoubleLink  // nodes referencing previous, next node and nested value (possibly nested)
	Tuple       // references a head value and nest of tail values
	Node        // node of a tree, or liked list
	Tree        // nodes referencing parent, root and a value of contained node(s)
	// INTERNAL TYPES //
	Intern   // InterType // instances of internal data structures (for self reference)
	MetaType // ValType   // values
	FuncType // FnType    // functions (user defined, as well as internal)
	TypeCons // type constructor instanciates a meta-type & provides it's nil and identity
	///////////
	MAX_VALUE_TYPE

	// flat values
	Unary = Bool | Int | Int8 | Int16 | Int32 | Uint |
		Uint16 | Uint32 | Float | Flt32 | Imag |
		Imag64 | Byte | Bytes | String | Error

	// types that come with type constuctors
	Nullable = Unary | Slice

	// Next() Value
	Chained = ChainedList | DoubleLink | Tuple | Node | Tree

	// Head() Value
	Linked = LinkedList | DoubleLink

	// Reversedious() Value
	Reversed = DoubleLink

	// Decap() (Value, Tupled)
	Consumed = Tuple | Node | Tree

	// Slice() []Value
	Ordered = Slice | List | MuliList

	// Get(attr) Attribute
	Mapped = Set | AttrSet | MuliSet | Record

	// sum of all collections
	Nary = Chained | Linked | Reversed | Consumed | Ordered | Mapped
)

//////// INTERNAL TYPES /////////////

type (
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
	bigInt    *big.Int
	timeVal   *time.Time
	duraVal   *time.Duration
	errorVal  struct{ error }
	//////
	attribute  Value
	flag       Flag
	slice      []Value
	collection struct{ s []Value }

//	collection struct {
//		nary  int
//		cells sliceVal
//	}
)

///// TYPE FLAG /////
type Flag uint

func (t Flag) uint(u uint)         { t = Flag(u) }
func (t Flag) Uint() uint          { return uint(t) }
func (t Flag) Type() Type          { return t }
func (t Flag) Flag() Flag          { return t }
func (t Flag) Value() interface{}  { return t }
func (t Flag) Ref() interface{}    { return &t }
func (t Flag) DeRef() Value        { inst := t; return inst }
func (t Flag) Copy() Value         { n := t; return n }
func (t Flag) Len() int            { return bits.Len(uint(t)) }
func (t Flag) Count() int          { return bits.OnesCount(uint(t)) }
func (t Flag) LeastSig() int       { return bits.TrailingZeros(uint(t)) + 1 }
func (t Flag) MostSig() int        { return bits.LeadingZeros(uint(t)) - 1 }
func (t Flag) Reverse() Flag       { return Flag(bits.Reverse(uint(t))) }
func (t Flag) Rotate(n int) Flag   { return Flag(bits.RotateLeft(uint(t), n)) }
func (t Flag) Toggle(v Typed) Flag { return Flag(uint(t) ^ v.Flag().Uint()) }
func (t Flag) Concat(v Typed) Flag { return Flag(uint(t) | v.Flag().Uint()) }
func (t Flag) Mask(v Typed) Flag   { return Flag(uint(t) &^ v.Flag().Uint()) }
func (t Flag) Match(v Typed) bool {
	if t.Uint()&v.Flag().Uint() != 0 {
		return true
	}
	return false
}
func (t Flag) Meta() bool {
	if t.Count() <= 1 {
		return false
	}
	return true
}
func (t Flag) Signature() []Type {
	return []Type{}
}

/////// TYPECONSTRUCTORS ///////
//
// type constructor
type InstanceConstructor func(v ...interface{}) Value
type TypeConstructor func(t ...Type) Type

func (t TypeConstructor) Identity() (val Value) {
	return val
}

func New(v Value) Value          { return v.Ref().(Value) } // new pointer ref
func newVal(v interface{}) Value { return Val(v) }
func Val(val interface{}) (rval Value) {
	switch val.(type) {
	case bool:
		rval = boolVal(val.(bool))
	case int, int64:
		rval = intVal(val.(int))
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
	case []Value:
		rval = val.(slice)
	case FnType, ValType, Type:
		rval = Flag(val.(ValType))
	}
	return rval
}
func (t TypeConstructor) Null() (val Value) {
	switch {
	case Nil.Flag().Match(t()):
		return nilVal{}
	case Bool.Flag().Match(t()):
		return Val(false)
	case Int.Flag().Match(t()):
		return Val(0)
	case Int8.Flag().Match(t()):
		return Val(int8(0))
	case Int16.Flag().Match(t()):
		return Val(int16(0))
	case Int32.Flag().Match(t()):
		return Val(int32(0))
	case Uint.Flag().Match(t()):
		return Val(uint(0))
	case Uint16.Flag().Match(t()):
		return Val(uint16(0))
	case Uint32.Flag().Match(t()):
		return Val(uint32(0))
	case Float.Flag().Match(t()):
		return Val(float64(0))
	case Flt32.Flag().Match(t()):
		return Val(float32(0))
	case Imag.Flag().Match(t()):
		return Val(complex128(float64(0)))
	case Imag64.Flag().Match(t()):
		return Val(complex64(float32(0)))
	case Byte.Flag().Match(t()):
		var b byte = 0
		return Val(b)
	case Bytes.Flag().Match(t()):
		var b []byte = []byte{}
		return Val(b)
	case String.Flag().Match(t()):
		s := ""
		return Val(s)
	case Error.Flag().Match(t()):
		var e error = fmt.Errorf("")
		return Val(e)
	case t().Flag().Meta():
	}
	return val
}

/// FLAG
func (nilVal) Flag() Flag       { return Nil.Flag() }
func (v boolVal) Flag() Flag    { return Bool.Flag() }
func (v intVal) Flag() Flag     { return Int.Flag() }
func (v int8Val) Flag() Flag    { return Int8.Flag() }
func (v int16Val) Flag() Flag   { return Int16.Flag() }
func (v int32Val) Flag() Flag   { return Int32.Flag() }
func (v uintVal) Flag() Flag    { return Uint.Flag() }
func (v uint16Val) Flag() Flag  { return Uint16.Flag() }
func (v uint32Val) Flag() Flag  { return Uint32.Flag() }
func (v fltVal) Flag() Flag     { return Float.Flag() }
func (v flt32Val) Flag() Flag   { return Flt32.Flag() }
func (v imagVal) Flag() Flag    { return Imag.Flag() }
func (v imag64Val) Flag() Flag  { return Imag64.Flag() }
func (v byteVal) Flag() Flag    { return Byte.Flag() }
func (v bytesVal) Flag() Flag   { return Bytes.Flag() }
func (v strVal) Flag() Flag     { return String.Flag() }
func (v slice) Flag() Flag      { return Slice.Flag() }
func (v errorVal) Flag() Flag   { return Error.Flag() }
func (v flag) Flag() Flag       { return MetaType.Flag() }
func (s collection) Flag() Flag { return Ordered.Flag() }

/// TYPE
func (nilVal) Type() Type       { return Nil.Type() }
func (v boolVal) Type() Type    { return Bool.Type() }
func (v intVal) Type() Type     { return Int.Type() }
func (v int8Val) Type() Type    { return Int8.Type() }
func (v int16Val) Type() Type   { return Int16.Type() }
func (v int32Val) Type() Type   { return Int32.Type() }
func (v uintVal) Type() Type    { return Uint.Type() }
func (v uint16Val) Type() Type  { return Uint16.Type() }
func (v uint32Val) Type() Type  { return Uint32.Type() }
func (v fltVal) Type() Type     { return Float.Type() }
func (v flt32Val) Type() Type   { return Flt32.Type() }
func (v imagVal) Type() Type    { return Imag.Type() }
func (v imag64Val) Type() Type  { return Imag64.Type() }
func (v byteVal) Type() Type    { return Byte.Type() }
func (v bytesVal) Type() Type   { return Bytes.Type() }
func (v strVal) Type() Type     { return String.Type() }
func (v slice) Type() Type      { return Slice.Type() }
func (v errorVal) Type() Type   { return Error.Type() }
func (v flag) Type() Type       { return MetaType.Type() }
func (s collection) Type() Type { return Ordered.Type() }

/// REFERENCE
func (v nilVal) Ref() interface{}     { return nil }
func (v nilVal) DeRef() Value         { return nil }
func (v boolVal) Ref() interface{}    { return &v }
func (v boolVal) DeRef() Value        { inst := *(v.Value().(*boolVal)); return inst }
func (v intVal) Ref() interface{}     { return &v }
func (v intVal) DeRef() Value         { inst := *(v.Value().(*intVal)); return inst }
func (v int16Val) Ref() interface{}   { return &v }
func (v int16Val) DeRef() Value       { inst := *(v.Value().(*int16Val)); return inst }
func (v int32Val) Ref() interface{}   { return &v }
func (v int32Val) DeRef() Value       { inst := *(v.Value().(*int32Val)); return inst }
func (v uintVal) Ref() interface{}    { return &v }
func (v uintVal) DeRef() Value        { inst := *(v.Value().(*uintVal)); return inst }
func (v uint16Val) Ref() interface{}  { return &v }
func (v uint16Val) DeRef() Value      { inst := *(v.Value().(*uint16Val)); return inst }
func (v uint32Val) Ref() interface{}  { return &v }
func (v uint32Val) DeRef() Value      { inst := *(v.Value().(*uint32Val)); return inst }
func (v fltVal) Ref() interface{}     { return &v }
func (v fltVal) DeRef() Value         { inst := *(v.Value().(*fltVal)); return inst }
func (v flt32Val) Ref() interface{}   { return &v }
func (v flt32Val) DeRef() Value       { inst := *(v.Value().(*flt32Val)); return inst }
func (v imagVal) Ref() interface{}    { return &v }
func (v imagVal) DeRef() Value        { inst := *(v.Value().(*imagVal)); return inst }
func (v imag64Val) Ref() interface{}  { return &v }
func (v imag64Val) DeRef() Value      { inst := *(v.Value().(*imag64Val)); return inst }
func (v byteVal) Ref() interface{}    { return &v }
func (v byteVal) DeRef() Value        { inst := *(v.Value().(*byteVal)); return inst }
func (v bytesVal) Ref() interface{}   { return &v }
func (v bytesVal) DeRef() Value       { inst := *(v.Value().(*bytesVal)); return inst }
func (v strVal) Ref() interface{}     { return &v }
func (v strVal) DeRef() Value         { inst := *(v.Value().(*strVal)); return inst }
func (v slice) Ref() interface{}      { return &v }
func (v slice) DeRef() Value          { inst := *(v.Value().(*slice)); return inst }
func (v errorVal) Ref() interface{}   { return &v }
func (v errorVal) DeRef() Value       { inst := *(v.Value().(*errorVal)); return inst }
func (v collection) Ref() interface{} { return &v }
func (v collection) DeRef() Value     { inst := *(v.Value().(*collection)); return inst }

/// VALUE
func (n nilVal) Value() interface{}     { return n }
func (v boolVal) Value() interface{}    { return v }
func (v intVal) Value() interface{}     { return v }
func (v int8Val) Value() interface{}    { return v }
func (v int16Val) Value() interface{}   { return v }
func (v int32Val) Value() interface{}   { return v }
func (v uintVal) Value() interface{}    { return v }
func (v uint16Val) Value() interface{}  { return v }
func (v uint32Val) Value() interface{}  { return v }
func (v imagVal) Value() interface{}    { return v }
func (v imag64Val) Value() interface{}  { return v }
func (v flt32Val) Value() interface{}   { return v }
func (v fltVal) Value() interface{}     { return v }
func (v byteVal) Value() interface{}    { return v }
func (v bytesVal) Value() interface{}   { return v }
func (v strVal) Value() interface{}     { return v }
func (v slice) Value() interface{}      { return v }
func (v errorVal) Value() interface{}   { return v }
func (v collection) Value() interface{} { return v }

/// COPY
func (v int32Val) Copy() Value  { var r int32Val = v; return r }
func (v int16Val) Copy() Value  { var r int16Val = v; return r }
func (v intVal) Copy() Value    { var r intVal = v; return r }
func (n nilVal) Copy() Value    { return nilVal(struct{}{}) }
func (v fltVal) Copy() Value    { var r fltVal = v; return r }
func (v uint32Val) Copy() Value { var r uint32Val = v; return r }
func (v uint16Val) Copy() Value { var r uint16Val = v; return r }
func (v uintVal) Copy() Value   { var r uintVal = v; return r }
func (v boolVal) Copy() Value   { var r boolVal = v; return r }
func (v flt32Val) Copy() Value  { var r flt32Val = v; return r }
func (v imagVal) Copy() Value   { var r imagVal = v; return r }
func (v imag64Val) Copy() Value { var r imag64Val = v; return r }
func (v byteVal) Copy() Value   { var r byteVal = v; return r }
func (v bytesVal) Copy() Value  { var r bytesVal = v; return r }
func (v strVal) Copy() Value    { var r strVal = v; return r }
func (v slice) Copy() Value     { var ret = []Value{}; return slice(append(ret, v)) }
func (v errorVal) Copy() Value  { var r errorVal = v; return r }
func (s collection) Copy() (v Value) {
	sl := []Value{}
	for _, val := range s.s {
		sl = append(sl, val.Copy())
	}
	v = &collection{sl}
	return v
}
