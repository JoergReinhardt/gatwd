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

//go:generate stringer -type=Type
const (
	Nil  Type = 1
	Bool Type = 1 << iota
	Int
	Int8 // Int8 -> Int8
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
	Error // let's do something sophisticated here...
	Slice
	Map
	Function
	Flag // marks most signifficant native type & data of type bitflag
	Native

	Nullable = Nil | Bool | Int | Int8 | Int16 | Int32 | BigInt | Uint |
		Uint8 | Uint16 | Uint32 | Float | Flt32 | BigFlt | Ratio | Imag |
		Imag64 | Byte | Rune | Bytes | String | Time | Duration | Error

	Numeral = Bool | Int | Int8 | Int16 | Int32 | BigInt | Uint | Uint8 |
		Uint16 | Uint32 | Float | Flt32 | BigFlt | Ratio | Imag |
		Imag64

	Unsigned = Uint | Uint8 | Uint16 | Uint32

	Integer = Int | Int8 | Int16 | Int32 | BigInt

	Rational = Integer | Ratio

	Irrational = Float | Flt32 | BigFlt

	Imaginary = Imag | Imag64

	Temporal = Time | Duration

	Symbolic = Byte | Rune | Bytes | String | Error

	Collections = Map | Slice

	Bitwise = Unsigned | Byte | Flag

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

//
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
