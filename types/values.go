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
	Time
	Duration
	Error
	// SLICE BASED COLLECTIONS //
	Attr // identitys, arity,  predicates, attribute accessors...
	//Cell     // Element to contain other elements
	Slice      // [Value]
	List       // ordered, indexed, monotyped values
	MuliList   // ordered, indexed, multityped values
	AttrList   // ordered, indexed, with search/sort attributation
	RecordList // ordered, indexes list of records, search-/sortable by any field
	UniSet     // unique, monotyped values
	MuliSet    // unique, attribute mapped, multityped values		 [attr,type,val]
	AttrSet    // unique, attribute mapped, monotyped values (aka map) [attr,val]
	Record     // unique, multityped, attributed, mapped, type declared values
	// LINKED COLLECTIONS // (also slice based, but pretend not to)
	ChainedList // nodes referencing next node and value (possibly nested)
	LinkedList  // nodes referencing previous, next node and nested value (possibly nested)
	DoubleLink  // nodes referencing previous, next node and nested value (possibly nested)
	Tuple       // references a head value and nest of tail values
	Node        // node of a tree, or liked list
	Tree        // nodes referencing parent, root and a value of contained node(s)
	// FUNCTIONS
	Function
	// POINTER REFERENCE //
	Pttr // value guaranueed to contain a pointer to the instance referenced by Value
	///////////
	NATIVES

	// flat value types
	Unary = Nil | Bool | Int | Int8 | Int16 | Int32 | Uint | Uint16 |
		Uint32 | Float | Flt32 | Imag | Imag64 | Byte | Bytes | String |
		Time | Duration | Error

	// combined value types
	Nary = Ordered | Linked | Reversed | Chained | Consumed | Mapped

	// types that come with type constuctors
	Nullable = Unary | Slice

	// Slice() []Value
	Ordered = Slice | List | MuliList | AttrList | RecordList

	// Head() Value
	Linked = List | MuliList | AttrList | RecordList

	// Reversedious() Value
	Reversed = DoubleLink

	// Next() Value
	Chained = ChainedList | LinkedList | DoubleLink | Tuple | Node | Tree

	// Decap() (Value, Tupled)
	Consumed = Tuple | Node | Tree

	// Get(attr) Attribute
	Mapped = UniSet | AttrSet | MuliSet | Record

	// higher order types combined from a finite set of other types, defined by a signature
	SumTypes = Slice | List | UniSet

	// Product Types are combinations of arbitrary other types in arbitrary combination
	ProductTypes = List | MuliList | AttrList | RecordList | AttrSet |
		MuliSet | Record | ChainedList | LinkedList | DoubleLink |
		Tuple | Node | Tree
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
	pair       struct {
		one Value
		two Value
	}
	triple struct {
		one   Value
		two   Value
		three Value
	}
	quadruple struct {
		one   Value
		two   Value
		three Value
		four  Value
	}
	quintiple struct {
		one   Value
		two   Value
		three Value
		four  Value
		five  Value
	}
	sextuple struct {
		one   Value
		two   Value
		three Value
		four  Value
		five  Value
		six   Value
	}
	septupel struct {
		one   Value
		two   Value
		three Value
		four  Value
		five  Value
		six   Value
		seven Value
	}
	octuple struct {
		one   Value
		two   Value
		three Value
		four  Value
		five  Value
		six   Value
		seven Value
		eight Value
	}
)

///// HIGHER ORDER TYPES /////

///// TYPE Flag /////
func (t flag) Type() flag          { return t }
func (t flag) Eval() Value         { return t }
func (t flag) Ref() Value          { return &t }
func (t flag) DeRef() Value        { inst := t; return inst }
func (t flag) Copy() Value         { n := t; return n }
func (t flag) uint() uint          { return uint(t) }
func (t flag) len() int            { return bits.Len(uint(t)) }
func (t flag) count() int          { return bits.OnesCount(uint(t)) }
func (t flag) least() int          { return bits.TrailingZeros(uint(t)) + 1 }
func (t flag) most() int           { return bits.LeadingZeros(uint(t)) - 1 }
func (t flag) reverse() flag       { return flag(bits.Reverse(uint(t))) }
func (t flag) rotate(n int) flag   { return flag(bits.RotateLeft(uint(t), n)) }
func (t flag) toggle(v Typed) flag { return flag(uint(t) ^ v.Type().uint()) }
func (t flag) concat(v Typed) flag { return flag(uint(t) | v.Type().uint()) }
func (t flag) mask(v Typed) flag   { return flag(uint(t) &^ v.Type().uint()) }
func (t flag) match(v Typed) bool {
	if t.uint()&v.Type().uint() != 0 {
		return true
	}
	return false
}
func (t flag) meta() bool {
	if t.count() <= 1 {
		return false
	}
	return true
}

//// TYPE CONSTRUCTOR ////
type TypeImplementation func(v ...interface{}) Value
type TypeConstructor func(t ...Typed) Typed
type Instanciator func(...Value) Value
type Instance func() Value

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
		s := ""
		return Make(s)
	case Error.Type().match(t):
		var e error = fmt.Errorf("")
		return Make(e)
	case t.Type().meta():
	}
	return val
}

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
func (v slice) Type() flag      { return Slice.Type() }
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
