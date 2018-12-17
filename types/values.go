package types

import (
	"math/big"
	"math/bits"
	"strconv"
	"time"
)

func New(val interface{}) Value {
	v := Value(Make(val))
	return v.Ref().(Value)
}
func Make(val interface{}) (rval Value) {
	switch val.(type) {
	case bool:
		rval = BoolVal(val.(bool))
	case int, int64:
		rval = IntVal(val.(int))
	case int8:
		rval = Int8Val(val.(int8))
	case int16:
		rval = Int16Val(val.(int16))
	case int32:
		rval = Int32Val(val.(int32))
	case uint, uint64:
		rval = UintVal(val.(uint))
	case uint16:
		rval = Uint16Val(val.(uint16))
	case uint32:
		rval = Int32Val(val.(int32))
	case float32:
		rval = Flt32Val(val.(float32))
	case float64:
		rval = FloatVal(val.(float64))
	case complex64:
		rval = ComplexVal(val.(complex64))
	case complex128:
		rval = ComplexVal(val.(complex128))
	case byte:
		rval = ByteVal(val.(byte))
	case []byte:
		rval = BytesVal(val.([]byte))
	case string:
		rval = StringVal(val.(string))
	case error:
		rval = ErrorVal{val.(error)}
	case FnType, ValType, Type:
		rval = Flag(val.(ValType))
	}
	return rval
}

//////// INTERNAL TYPE SYSTEM ///////////
type Flag uint

func (t Flag) Uint() uint         { return uint(t) }
func (t Flag) Type() Type         { return t }
func (t Flag) Flag() Flag         { return t }
func (t Flag) Ref() interface{}   { return &t }
func (t Flag) Value() Value       { return t }
func (t Flag) Copy() Value        { n := t; return n }
func (v Flag) String() string     { return v.Type().String() }
func (t Flag) Len() int           { return bits.Len(uint(t)) }
func (t Flag) Count() int         { return bits.OnesCount(uint(t)) }
func (t Flag) LeastSig() int      { return bits.TrailingZeros(uint(t)) + 1 }
func (t Flag) MostSig() int       { return bits.LeadingZeros(uint(t)) - 1 }
func (t Flag) Reverse() Flag      { return Flag(bits.Reverse(uint(t))) }
func (t Flag) Rotate(n int) Flag  { return Flag(bits.RotateLeft(uint(t), n)) }
func (t Flag) Toggle(v Flag) Flag { return Flag(uint(t) ^ v.Uint()) }
func (t Flag) Concat(v Flag) Flag { return Flag(uint(t) | v.Uint()) }
func (t Flag) Mask(v Flag) Flag   { return Flag(uint(t) &^ v.Uint()) }
func (t Flag) Match(v Flag) bool {
	if t.Uint()&v.Uint() != 0 {
		return true
	}
	return false
}

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
	Complex
	Complex64
	Byte
	Bytes
	String
	Error
	// SLICE BASED COLLECTIONS //
	Attr     // identitys, arity,  predicates, attribute accessors...
	Arry     // [Value]
	List     // ordered, indexed, monotyped values
	MuliList // ordered, indexed, multityped values
	Set      // unique, monotyped values
	AttrSet  // unique, attribute mapped, monotyped values (aka map)
	MuliSet  // unique, attribute mapped, multityped values
	Record   // unique, attribute mapped, type declared values
	// LINKED COLLECTIONS // (also slice based, but pretend not to)
	Linked       // nodes referencing next node and value (possibly nested)
	DoubleLinked // nodes referencing previous, next node and nested value (possibly nested)
	Tuple        // references a head value and nest of tail values
	Node         // node of a tree, or liked list
	Tree         // nodes referencing parent, root and a value of contained node(s)
	// INTERNAL TYPES //
	MetaType // ValType   // values
	FuncType // FnType    // functions (user defined, as well as internal)
	Intern   // InterType // instances internal data structures (for self reference)
	Native   // type(val) // instances of native go values represented by empty inerfaces

	Unary = Bool | Int | Int8 | Int16 | Int32 | Uint | Uint16 | Uint32 |
		Float | Flt32 | Complex | Complex64 | Byte | Bytes | String | Error

	Nary = List | MuliList | Set | AttrSet | MuliSet | Record | Linked |
		DoubleLinked | Tuple | Node | Tree
)

//////// native types /////////////
type (
	BoolVal      bool
	IntVal       int
	Int8Val      int8
	Int16Val     int16
	Int32Val     int32
	UintVal      uint
	Uint16Val    uint16
	Uint32Val    uint32
	FloatVal     float64
	Flt32Val     float32
	ComplexVal   complex128
	Complex64Val complex64
	ByteVal      byte
	BytesVal     []byte
	StringVal    string
	BigIntVal    *big.Int
	TimeVal      *time.Time
	ErrorVal     struct{ error }
	AttrVal      Value
	TypeVal      Flag
)

func (v BoolVal) Type() Type      { return Bool.Type() }
func (v IntVal) Type() Type       { return Int.Type() }
func (v Int8Val) Type() Type      { return Int8.Type() }
func (v Int16Val) Type() Type     { return Int16.Type() }
func (v Int32Val) Type() Type     { return Int32.Type() }
func (v UintVal) Type() Type      { return Uint.Type() }
func (v Uint16Val) Type() Type    { return Uint16.Type() }
func (v Uint32Val) Type() Type    { return Uint32.Type() }
func (v FloatVal) Type() Type     { return Float.Type() }
func (v ComplexVal) Type() Type   { return Complex.Type() }
func (v Complex64Val) Type() Type { return Complex64.Type() }
func (v ByteVal) Type() Type      { return Byte.Type() }
func (v BytesVal) Type() Type     { return Bytes.Type() }
func (v StringVal) Type() Type    { return String.Type() }
func (v ErrorVal) Type() Type     { return Error.Type() }
func (v TypeVal) Type() Type      { return MetaType.Type() }

///// STRING (CONVERSION) METHODS ///////
func (v BoolVal) String() string   { return strconv.FormatBool(bool(v)) }
func (v IntVal) String() string    { return strconv.Itoa(int(v)) }
func (v Int8Val) String() string   { return strconv.Itoa(int(v)) }
func (v Int16Val) String() string  { return strconv.Itoa(int(v)) }
func (v Int32Val) String() string  { return strconv.Itoa(int(v)) }
func (v UintVal) String() string   { return strconv.Itoa(int(v)) }
func (v Uint16Val) String() string { return strconv.Itoa(int(v)) }
func (v Uint32Val) String() string { return strconv.Itoa(int(v)) }
func (v FloatVal) String() string  { return strconv.FormatFloat(float64(v), 'G', -1, 64) }
func (v Flt32Val) String() string  { return strconv.FormatFloat(float64(v), 'G', -1, 32) }
func (v ByteVal) String() string   { return strconv.Itoa(int(v)) }
func (v BytesVal) String() string  { return string(v) }
func (v StringVal) String() string { return string(v) }
func (v StringVal) Key() string    { return string(v) }
func (v ErrorVal) String() string  { return v.error.Error() }
func (v ErrorVal) Error() error    { return v.error }
func (v ComplexVal) String() string {
	return strconv.FormatFloat(float64(real(v)), 'G', -1, 64) + " + " +
		strconv.FormatFloat(float64(imag(v)), 'G', -1, 64) + "i"
}
func (v Complex64Val) String() string {
	return strconv.FormatFloat(float64(real(v)), 'G', -1, 32) + " + " +
		strconv.FormatFloat(float64(imag(v)), 'G', -1, 32) + "i"
}

///// TYPE CONVERSION //////
// BOOL -> VALUE
func (v BoolVal) Int() IntVal {
	if v {
		return IntVal(1)
	}
	return IntVal(-1)
}
func (v BoolVal) IntNat() int {
	if v {
		return 1
	}
	return -1
}
func (v BoolVal) UintNat() uint {
	if v {
		return 1
	}
	return 0
}

// VALUE -> BOOL
func (v IntVal) Bool() BoolVal {
	if v == 1 {
		return BoolVal(true)
	}
	return BoolVal(false)
}
func (v StringVal) Bool() BoolVal {
	s, err := strconv.ParseBool(string(v))
	if err != nil {
		return false
	}
	return BoolVal(s)
}

// INT -> VALUE
func (v IntVal) Idx() int        { return int(v) }     // implements Idx Attribut
func (v IntVal) Key() string     { return v.String() } // implements Key Attribut
func (v IntVal) FltNat() float64 { return float64(v) }
func (v IntVal) IntNat() int     { return int(v) }
func (v IntVal) UintNat() uint {
	if v < 0 {
		return uint(v * -1)
	}
	return uint(v)
}

// VALUE -> INT
func (v Int8Val) Int() IntVal     { return IntVal(int(v)) }
func (v Int16Val) Int() IntVal    { return IntVal(int(v)) }
func (v Int32Val) Int() IntVal    { return IntVal(int(v)) }
func (v UintVal) Int() IntVal     { return IntVal(int(v)) }
func (v Uint16Val) Int() IntVal   { return IntVal(int(v)) }
func (v Uint32Val) Int() IntVal   { return IntVal(int(v)) }
func (v FloatVal) Int() IntVal    { return IntVal(int(v)) }
func (v Flt32Val) Int() IntVal    { return IntVal(int(v)) }
func (v ByteVal) Int() IntVal     { return IntVal(int(v)) }
func (v ComplexVal) Real() IntVal { return IntVal(real(v)) }
func (v ComplexVal) Imag() IntVal { return IntVal(imag(v)) }
func (v StringVal) Int() IntVal {
	s, err := strconv.Atoi(string(v))
	if err != nil {
		return -1
	}
	return IntVal(s)
}

// SPECIAL INTEGERS
func (v UintVal) Len() IntVal   { return IntVal(bits.Len64(uint64(v))) }
func (v ByteVal) Len() IntVal   { return IntVal(bits.Len64(uint64(v))) }
func (v BytesVal) Len() IntVal  { return IntVal(len(v)) }
func (v StringVal) Len() IntVal { return IntVal(len(string(v))) }

// VALUE -> FLOAT
func (v UintVal) Float() FloatVal { return FloatVal(v.Int().Float()) }
func (v IntVal) Float() FloatVal  { return FloatVal(v.FltNat()) }
func (v StringVal) Float() FloatVal {
	s, err := strconv.ParseFloat(string(v), 64)
	if err != nil {
		return -1
	}
	return FloatVal(s)
}

// VALUE -> UINT
func (v UintVal) Uint() UintVal { return v }
func (v UintVal) UintNat() uint { return uint(v) }
func (v IntVal) Uint() UintVal  { return UintVal(v.UintNat()) }
func (v StringVal) Uint() UintVal {
	u, err := strconv.ParseUint(string(v), 10, 64)
	if err != nil {
		return 0
	}
	return UintVal(u)
}
func (v BoolVal) Uint() UintVal {
	if v {
		return UintVal(1)
	}
	return UintVal(0)
}

///// methods implementing the value interface

func (v BoolVal) Ref() interface{} { return &v }
func (v BoolVal) Value() Value     { return v }
func (v BoolVal) Copy() Value      { var r BoolVal = v; return r }

func (v IntVal) Ref() interface{} { return &v }
func (v IntVal) Value() Value     { return v }
func (v IntVal) Copy() Value      { var r IntVal = v; return r }

func (v Int8Val) Value() Value     { return v }
func (v Int8Val) Ref() interface{} { return &v }
func (v Int8Val) Copy() Value      { var r Int8Val = v; return r }

func (v Int16Val) Value() Value     { return v }
func (v Int16Val) Ref() interface{} { return &v }
func (v Int16Val) Copy() Value      { var r Int16Val = v; return r }

func (v Int32Val) Value() Value     { return v }
func (v Int32Val) Ref() interface{} { return &v }
func (v Int32Val) Copy() Value      { var r Int32Val = v; return r }

func (v UintVal) Value() Value     { return v }
func (v UintVal) Ref() interface{} { return &v }
func (v UintVal) Copy() Value      { var r UintVal = v; return r }

func (v Uint16Val) Value() Value     { return v }
func (v Uint16Val) Ref() interface{} { return &v }
func (v Uint16Val) Copy() Value      { var r Uint16Val = v; return r }

func (v Uint32Val) Value() Value     { return v }
func (v Uint32Val) Ref() interface{} { return &v }
func (v Uint32Val) Copy() Value      { var r Uint32Val = v; return r }

func (v FloatVal) Value() Value     { return v }
func (v FloatVal) Ref() interface{} { return &v }
func (v FloatVal) Copy() Value      { var r FloatVal = v; return r }

func (v Flt32Val) Value() Value     { return v }
func (v Flt32Val) Ref() interface{} { return &v }
func (v Flt32Val) Type() Type       { return Flt32.Type() }
func (v Flt32Val) Copy() Value      { var r Flt32Val = v; return r }

func (v ComplexVal) Value() Value     { return v }
func (v ComplexVal) Ref() interface{} { return &v }
func (v ComplexVal) Copy() Value      { var r ComplexVal = v; return r }

func (v Complex64Val) Value() Value     { return v }
func (v Complex64Val) Ref() interface{} { return &v }
func (v Complex64Val) Copy() Value      { var r Complex64Val = v; return r }

func (v ByteVal) Value() Value     { return v }
func (v ByteVal) Ref() interface{} { return &v }
func (v ByteVal) Copy() Value      { var r ByteVal = v; return r }

func (v BytesVal) Value() Value     { return v }
func (v BytesVal) Ref() interface{} { return &v }
func (v BytesVal) Copy() Value      { var r BytesVal = v; return r }

func (v StringVal) Value() Value     { return v }
func (v StringVal) Ref() interface{} { return &v }
func (v StringVal) Copy() Value      { var r StringVal = v; return r }

func (v ErrorVal) Value() Value     { return v }
func (v ErrorVal) Ref() interface{} { return &v }
func (v ErrorVal) Copy() Value      { var r ErrorVal = v; return r }
