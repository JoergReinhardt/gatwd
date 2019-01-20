package data

import (
	"math/big"
	"time"
)

//////// INTERNAL TYPE SYSTEM ///////////
//
// intended to be accessable and extendable
type Type BitFlag

func (v Type) Flag() BitFlag { return BitFlag(v) }

func ListAllTypes() []Type {
	var tt = []Type{}
	var i uint
	var t Type = 0
	for t < Flag {
		t = 1 << i
		i = i + 1
		tt = append(tt, Type(t))
	}
	return tt
}

//go:generate stringer -type=Type
const (
	Nil  Type = 1
	Bool Type = 1 << iota
	Int8      // Int8 -> Int8
	Int16
	Int32
	Int
	BigInt
	Uint8
	Uint16
	Uint32
	Uint
	Flt32
	Float
	BigFlt
	Ratio
	Imag64
	Imag
	Time
	Duration
	Byte
	Rune
	Bytes
	String
	Error // let's do something sophisticated here...
	//// HIGHERORDER TYPES
	Pair
	Tuple
	Record
	Vector
	List
	Set
	Function
	Argument
	Parameter
	Definition
	Flag // marks most signifficant native type & data of type bitflag

	// TYPE CLASSES
	// precedence type classes define argument types functions that accept
	// a set of possible input types
	Nullable = Nil | Bool | Int8 | Int16 | Int32 |
		Int | BigInt | Uint8 | Uint16 | Uint32 | Uint |
		Flt32 | Float | BigFlt | Ratio | Imag64 | Imag |
		Time | Duration | Byte | Rune | Bytes | String

	Bitwise = Unsigned | Byte | Flag

	Boolean = Bool | Bitwise

	Unsigned = Uint | Uint8 | Uint16 | Uint32 | Byte

	Signed = Int | Int8 | Int16 | Int32 | BigInt | Byte

	Integer = Unsigned | Signed

	Rational = Unsigned | Integer | Ratio

	Irrational = Float | Flt32 | BigFlt

	Imaginary = Imag | Imag64

	Temporal = Time | Duration

	Textual = String | Rune | Bytes

	Numeric = Integer | Rational | Irrational |
		Imaginary | Temporal

	Symbolic = Textual | Boolean | Temporal | Error

	/// here will be dragons‥.
	HigherOrder = Functional | Collection

	Collection = Pair | Tuple | Record | Vector |
		List | Set

	Functional = Function | Argument | Parameter |
		Definition | Flag

	MAX_INT Type = 0xFFFFFFFFFFFFFFFF
	Mask         = MAX_INT ^ Flag
)

//////// INTERNAL TYPES /////////////
// internal types are typealiases without any wrapping, or referencing getting
// in the way performancewise. types need to be aliased in the first place, to
// associate them with a bitflag, without having to actually asign, let alone
// attach it to the instance.
type ( ////// INTERNAL TYPES /////
	BitFlag uint
	////// TYPE ALIASES ///////
	NilVal    struct{}
	BoolVal   bool
	IntVal    int
	Int8Val   int8
	Int16Val  int16
	Int32Val  int32
	UintVal   uint
	Uint8Val  uint8
	Uint16Val uint16
	Uint32Val uint32
	FltVal    float64
	Flt32Val  float32
	ImagVal   complex128
	Imag64Val complex64
	ByteVal   byte
	RuneVal   rune
	BytesVal  []byte
	StrVal    string
	BigIntVal big.Int
	BigFltVal big.Float
	RatioVal  big.Rat
	TimeVal   time.Time
	DuraVal   time.Duration
	ErrorVal  struct{ e error }
)

//////// ATTRIBUTE TYPE ALIAS /////////////////

/// bind the appropriate Flag Method to every type
func (v BitFlag) Flag() BitFlag   { return Flag.Flag() }
func (NilVal) Flag() BitFlag      { return Nil.Flag() }
func (v BoolVal) Flag() BitFlag   { return Bool.Flag() }
func (v IntVal) Flag() BitFlag    { return Int.Flag() }
func (v Int8Val) Flag() BitFlag   { return Int8.Flag() }
func (v Int16Val) Flag() BitFlag  { return Int16.Flag() }
func (v Int32Val) Flag() BitFlag  { return Int32.Flag() }
func (v UintVal) Flag() BitFlag   { return Uint.Flag() }
func (v Uint8Val) Flag() BitFlag  { return Uint8.Flag() }
func (v Uint16Val) Flag() BitFlag { return Uint16.Flag() }
func (v Uint32Val) Flag() BitFlag { return Uint32.Flag() }
func (v BigIntVal) Flag() BitFlag { return BigInt.Flag() }
func (v FltVal) Flag() BitFlag    { return Float.Flag() }
func (v Flt32Val) Flag() BitFlag  { return Flt32.Flag() }
func (v BigFltVal) Flag() BitFlag { return BigFlt.Flag() }
func (v ImagVal) Flag() BitFlag   { return Imag.Flag() }
func (v Imag64Val) Flag() BitFlag { return Imag64.Flag() }
func (v RatioVal) Flag() BitFlag  { return Ratio.Flag() }
func (v RuneVal) Flag() BitFlag   { return Rune.Flag() }
func (v ByteVal) Flag() BitFlag   { return Byte.Flag() }
func (v BytesVal) Flag() BitFlag  { return Bytes.Flag() }
func (v StrVal) Flag() BitFlag    { return String.Flag() }
func (v TimeVal) Flag() BitFlag   { return Time.Flag() }
func (v DuraVal) Flag() BitFlag   { return Duration.Flag() }
func (v ErrorVal) Flag() BitFlag  { return Error.Flag() }

///
func (NilVal) Copy() Data      { return NilVal{} }
func (v BitFlag) Copy() Data   { return BitFlag(v) }
func (v BoolVal) Copy() Data   { return BoolVal(v) }
func (v IntVal) Copy() Data    { return IntVal(v) }
func (v Int8Val) Copy() Data   { return Int8Val(v) }
func (v Int16Val) Copy() Data  { return Int16Val(v) }
func (v Int32Val) Copy() Data  { return Int32Val(v) }
func (v UintVal) Copy() Data   { return UintVal(v) }
func (v Uint8Val) Copy() Data  { return Uint8Val(v) }
func (v Uint16Val) Copy() Data { return Uint16Val(v) }
func (v Uint32Val) Copy() Data { return Uint32Val(v) }
func (v BigIntVal) Copy() Data { return BigIntVal(v) }
func (v FltVal) Copy() Data    { return FltVal(v) }
func (v Flt32Val) Copy() Data  { return Flt32Val(v) }
func (v BigFltVal) Copy() Data { return BigFltVal(v) }
func (v ImagVal) Copy() Data   { return ImagVal(v) }
func (v Imag64Val) Copy() Data { return Imag64Val(v) }
func (v RatioVal) Copy() Data  { return RatioVal(v) }
func (v RuneVal) Copy() Data   { return RuneVal(v) }
func (v ByteVal) Copy() Data   { return ByteVal(v) }
func (v BytesVal) Copy() Data  { return BytesVal(v) }
func (v StrVal) Copy() Data    { return StrVal(v) }
func (v TimeVal) Copy() Data   { return TimeVal(v) }
func (v DuraVal) Copy() Data   { return DuraVal(v) }
func (v ErrorVal) Copy() Data  { return ErrorVal(v) }

///
func (NilVal) Eval() Data      { return NilVal{} }
func (v BitFlag) Eval() Data   { return v }
func (v BoolVal) Eval() Data   { return v }
func (v IntVal) Eval() Data    { return v }
func (v Int8Val) Eval() Data   { return v }
func (v Int16Val) Eval() Data  { return v }
func (v Int32Val) Eval() Data  { return v }
func (v UintVal) Eval() Data   { return v }
func (v Uint8Val) Eval() Data  { return v }
func (v Uint16Val) Eval() Data { return v }
func (v Uint32Val) Eval() Data { return v }
func (v BigIntVal) Eval() Data { return v }
func (v FltVal) Eval() Data    { return v }
func (v Flt32Val) Eval() Data  { return v }
func (v BigFltVal) Eval() Data { return v }
func (v ImagVal) Eval() Data   { return v }
func (v Imag64Val) Eval() Data { return v }
func (v RatioVal) Eval() Data  { return v }
func (v RuneVal) Eval() Data   { return v }
func (v ByteVal) Eval() Data   { return v }
func (v BytesVal) Eval() Data  { return v }
func (v StrVal) Eval() Data    { return v }
func (v TimeVal) Eval() Data   { return v }
func (v DuraVal) Eval() Data   { return v }
func (v ErrorVal) Eval() Data  { return v }

///
func (NilVal) Ident() Data      { return NilVal{} }
func (v BitFlag) Ident() Data   { return v }
func (v BoolVal) Ident() Data   { return v }
func (v IntVal) Ident() Data    { return v }
func (v Int8Val) Ident() Data   { return v }
func (v Int16Val) Ident() Data  { return v }
func (v Int32Val) Ident() Data  { return v }
func (v UintVal) Ident() Data   { return v }
func (v Uint8Val) Ident() Data  { return v }
func (v Uint16Val) Ident() Data { return v }
func (v Uint32Val) Ident() Data { return v }
func (v BigIntVal) Ident() Data { return v }
func (v FltVal) Ident() Data    { return v }
func (v Flt32Val) Ident() Data  { return v }
func (v BigFltVal) Ident() Data { return v }
func (v ImagVal) Ident() Data   { return v }
func (v Imag64Val) Ident() Data { return v }
func (v RatioVal) Ident() Data  { return v }
func (v RuneVal) Ident() Data   { return v }
func (v ByteVal) Ident() Data   { return v }
func (v BytesVal) Ident() Data  { return v }
func (v StrVal) Ident() Data    { return v }
func (v TimeVal) Ident() Data   { return v }
func (v DuraVal) Ident() Data   { return v }
func (v ErrorVal) Ident() Data  { return v }

//// native nullable typed ///////
func (v NilVal) Null() struct{}       { return struct{}{} }
func (v BoolVal) Null() bool          { return false }
func (v IntVal) Null() int            { return 0 }
func (v Int8Val) Null() int8          { return 0 }
func (v Int16Val) Null() int16        { return 0 }
func (v Int32Val) Null() int32        { return 0 }
func (v UintVal) Null() uint          { return 0 }
func (v Uint8Val) Null() uint8        { return 0 }
func (v Uint16Val) Null() uint16      { return 0 }
func (v Uint32Val) Null() uint32      { return 0 }
func (v FltVal) Null() float64        { return 0 }
func (v Flt32Val) Null() float32      { return 0 }
func (v ImagVal) Null() complex128    { return complex128(0.0) }
func (v Imag64Val) Null() complex64   { return complex64(0.0) }
func (v ByteVal) Null() byte          { return byte(0) }
func (v RuneVal) Null() rune          { return rune(' ') }
func (v StrVal) Null() string         { return string("") }
func (v BigIntVal) Null() *big.Int    { return big.NewInt(0) }
func (v BigFltVal) Null() *big.Float  { return big.NewFloat(0) }
func (v RatioVal) Null() *big.Rat     { return big.NewRat(1, 1) }
func (v TimeVal) Null() time.Time     { return time.Now() }
func (v DuraVal) Null() time.Duration { return time.Duration(0) }
