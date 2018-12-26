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
type Type BitFlag

func (v Type) Flag() BitFlag { return BitFlag(v) }

//go:generate stringer -type=Type
const (
	Nil  Type = 0
	Bool Type = 1
	Int  Type = 1 << iota
	Int8      // Int8 -> Int8
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
	Function
	///////
	Flag      // generic bitflag role
	DataType  // types of the user defineable type system
	NodeType  // types of nodes in linked trees
	TokenType // types of tokens in parsed input
	MetaType  // higher order type

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
	BitFlag         uint
	slice           []Data
	args            []BitFlag
	parms           []Attribute
	Attribute       Constant
	Constant        func() Data
	lambda          func(...Data) (args, Data)
	function        func() (lambda, strVal)
	dataClosure     func(...Data) Data
	lambdaClosure   func(...Data) Data
	functionClosure func(...Data) Data
)

/// Flag Method for all types
func (a Attribute) Flag() BitFlag {
	return a().Flag().Concat(Attr.Flag())
}
func (nilVal) Flag() BitFlag          { return Nil.Flag() }
func (a Attribute) AttrType() BitFlag { return Attr.Flag() }
func (v BitFlag) Flag() BitFlag       { return Flag.Flag() }
func (v boolVal) Flag() BitFlag       { return Bool.Flag() }
func (v intVal) Flag() BitFlag        { return Int.Flag() }
func (v int8Val) Flag() BitFlag       { return Int8.Flag() }
func (v int16Val) Flag() BitFlag      { return Int16.Flag() }
func (v int32Val) Flag() BitFlag      { return Int32.Flag() }
func (v uintVal) Flag() BitFlag       { return Uint.Flag() }
func (v uint8Val) Flag() BitFlag      { return Uint8.Flag() }
func (v uint16Val) Flag() BitFlag     { return Uint16.Flag() }
func (v uint32Val) Flag() BitFlag     { return Uint32.Flag() }
func (v bigIntVal) Flag() BitFlag     { return BigInt.Flag() }
func (v fltVal) Flag() BitFlag        { return Float.Flag() }
func (v flt32Val) Flag() BitFlag      { return Flt32.Flag() }
func (v bigFltVal) Flag() BitFlag     { return BigFlt.Flag() }
func (v imagVal) Flag() BitFlag       { return Imag.Flag() }
func (v imag64Val) Flag() BitFlag     { return Imag64.Flag() }
func (v ratioVal) Flag() BitFlag      { return Ratio.Flag() }
func (v runeVal) Flag() BitFlag       { return Rune.Flag() }
func (v byteVal) Flag() BitFlag       { return Byte.Flag() }
func (v bytesVal) Flag() BitFlag      { return Bytes.Flag() }
func (v strVal) Flag() BitFlag        { return String.Flag() }
func (v timeVal) Flag() BitFlag       { return Time.Flag() }
func (v duraVal) Flag() BitFlag       { return Duration.Flag() }
func (v slice) Flag() BitFlag         { return Chain.Flag() }
func (v errorVal) Flag() BitFlag      { return Error.Flag() }

//// BOUND TYPE FLAG METHODS ////
func (v BitFlag) Uint() uint               { return fuint(v) }
func (v BitFlag) Len() int                 { return flen(v) }
func (v BitFlag) Count() int               { return fcount(v) }
func (v BitFlag) Least() int               { return fleast(v) }
func (v BitFlag) Most() int                { return fmost(v) }
func (v BitFlag) Low(f Typed) Typed        { return flow(f) }
func (v BitFlag) High(f Typed) Typed       { return fhigh(f) }
func (v BitFlag) Reverse() BitFlag         { return frev(v) }
func (v BitFlag) Rotate(n int) BitFlag     { return frot(v, n) }
func (v BitFlag) Toggle(f BitFlag) BitFlag { return ftog(v, f) }
func (v BitFlag) Concat(f BitFlag) BitFlag { return fconc(v, f) }
func (v BitFlag) Mask(f BitFlag) BitFlag   { return fmask(v, f) }
func (v BitFlag) Match(f Typed) bool       { return fmatch(v, f) }

///// FREE TYPE FLAG METHOD IMPLEMENTATIONS /////
func fuint(t BitFlag) uint               { return uint(t) }
func flen(t BitFlag) int                 { return bits.Len(uint(t)) }
func fcount(t BitFlag) int               { return bits.OnesCount(uint(t)) }
func fleast(t BitFlag) int               { return bits.TrailingZeros(uint(t)) + 1 }
func fmost(t BitFlag) int                { return bits.LeadingZeros(uint(t)) - 1 }
func frev(t BitFlag) BitFlag             { return BitFlag(bits.Reverse(uint(t))) }
func frot(t BitFlag, n int) BitFlag      { return BitFlag(bits.RotateLeft(uint(t), n)) }
func ftog(t BitFlag, v BitFlag) BitFlag  { return BitFlag(uint(t) ^ v.Flag().Uint()) }
func fconc(t BitFlag, v BitFlag) BitFlag { return BitFlag(uint(t) | v.Flag().Uint()) }
func fmask(t BitFlag, v BitFlag) BitFlag { return BitFlag(uint(t) &^ v.Flag().Uint()) }
func fshow(f Typed) string               { return fmt.Sprintf("%64b\n", f) }
func flow(t Typed) Typed                 { return fmask(t.Flag(), BitFlag(Mask)) }
func fhigh(t Typed) Typed {
	len := flen(BitFlag(Natives))
	return fmask(frot(t.Flag(), len), frot(BitFlag(Natives), len))
}
func fmatch(t BitFlag, v Typed) bool {
	if t.Uint()&v.Flag().Uint() != 0 {
		return true
	}
	return false
}

//////// IDENTITY & TYPE-REGISTER TYPES /////////
