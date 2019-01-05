package data

import (
	"fmt"
	"math/big"
	"math/bits"
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
	Flag // marks most signifficant native type & data of type bitflag
	Function

	Nullable = Nil | Bool | Int | Int8 | Int16 | Int32 | BigInt | Uint |
		Uint8 | Uint16 | Uint32 | Float | Flt32 | BigFlt | Ratio | Imag |
		Imag64 | Byte | Rune | Bytes | String | Time | Duration | Error

	Numbers = Bool | Int | Int8 | Int16 | Int32 | BigInt | Uint | Uint8 |
		Uint16 | Uint32 | Float | Flt32 | BigFlt | Ratio | Imag |
		Imag64

	Unsigned = Uint | Uint8 | Uint16 | Uint32

	Integer = Int | Int8 | Int16 | Int32 | BigInt

	Rational = Integer | Ratio

	Irrational = Float | Flt32 | BigFlt

	Imaginary = Imag | Imag64

	Temporal = Time | Duration

	Symbolic = Byte | Rune | Bytes | String | Error | Flag

	Enumerable = Map | Slice

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
	ErrorVal  struct{ v error }
	Native    struct{ i interface{} }
)

//////// ATTRIBUTE TYPE ALIAS /////////////////

/// bind the appropriate Flag Method to every type
func (v BitFlag) Flag() BitFlag   { return v }
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

//// BOUND TYPE FLAG METHODS ////
func (v BitFlag) Uint() uint               { return uint(v) }
func (v BitFlag) Len() int                 { return FlagLength(v) }
func (v BitFlag) Count() int               { return Count(v) }
func (v BitFlag) Least() int               { return LeastSig(v) }
func (v BitFlag) Most() int                { return MostSig(v) }
func (v BitFlag) Low(f BitFlag) BitFlag    { return Low(f).Flag() }
func (v BitFlag) High(f BitFlag) BitFlag   { return High(f).Flag() }
func (v BitFlag) Reverse() BitFlag         { return Reverse(v).Flag() }
func (v BitFlag) Rotate(n int) BitFlag     { return Rotate(v, n).Flag() }
func (v BitFlag) Toggle(f BitFlag) BitFlag { return Toggle(v, f).Flag() }
func (v BitFlag) Concat(f BitFlag) BitFlag { return Concat(v, f).Flag() }
func (v BitFlag) Mask(f BitFlag) BitFlag   { return MaskFlag(v, f).Flag() }
func (v BitFlag) Match(f BitFlag) bool     { return Match(v, f) }

///// FREE TYPE FLAG METHOD IMPLEMENTATIONS /////
func flag(t Typed) BitFlag              { return t.Flag() }
func FlagLength(t Typed) int            { return bits.Len(t.Flag().Uint()) }
func Count(t Typed) int                 { return bits.OnesCount(t.Flag().Uint()) }
func LeastSig(t Typed) int              { return bits.TrailingZeros(t.Flag().Uint()) + 1 }
func MostSig(t Typed) int               { return bits.LeadingZeros(t.Flag().Uint()) - 1 }
func Reverse(t Typed) BitFlag           { return BitFlag(bits.Reverse(t.Flag().Uint())) }
func Rotate(t Typed, n int) BitFlag     { return BitFlag(bits.RotateLeft(t.Flag().Uint(), n)) }
func Toggle(t Typed, v Typed) BitFlag   { return BitFlag(t.Flag().Uint() ^ v.Flag().Uint()) }
func Concat(t Typed, v Typed) BitFlag   { return BitFlag(t.Flag().Uint() | v.Flag().Uint()) }
func MaskFlag(t Typed, v Typed) BitFlag { return BitFlag(t.Flag().Uint() &^ v.Flag().Uint()) }
func Show(f Typed) string               { return fmt.Sprintf("%64b\n", f) }
func Low(t Typed) Typed                 { return MaskFlag(t.Flag(), Typed(Mask)) }
func High(t BitFlag) BitFlag {
	len := FlagLength(BitFlag(Flag))
	return MaskFlag(Rotate(t.Flag(), len), Rotate(BitFlag(Flag), len))
}
func Match(t BitFlag, v BitFlag) bool {
	if t.Uint()&v.Flag().Uint() != 0 {
		return true
	}
	return false
}
