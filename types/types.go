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

func (v ValType) Flag() flag { return flag(v) }

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
	///////
	BitFlag   // generic bitflag role
	DataType  // types of the user defineable type system
	NodeType  // types of nodes in linked trees
	TokenType // types of tokens in parsed input
	HOType    // higher order type

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
	flag            uint
	slice           []Data
	args            []flag
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
func (c Constant) Flag() flag      { return c().Flag() }
func (a Attribute) Flag() flag     { return Attr.Flag() }
func (a Attribute) AttrType() flag { return a().Flag() }
func (v flag) Flag() flag          { return v }
func (nilVal) Flag() flag          { return BitFlag.Flag() }
func (v boolVal) Flag() flag       { return Bool.Flag() }
func (v intVal) Flag() flag        { return Int.Flag() }
func (v int8Val) Flag() flag       { return Int8.Flag() }
func (v int16Val) Flag() flag      { return Int16.Flag() }
func (v int32Val) Flag() flag      { return Int32.Flag() }
func (v uintVal) Flag() flag       { return Uint.Flag() }
func (v uint8Val) Type() flag      { return Uint8.Flag() }
func (v uint16Val) Flag() flag     { return Uint16.Flag() }
func (v uint32Val) Type() flag     { return Uint32.Flag() }
func (v bigIntVal) Flag() flag     { return BigInt.Flag() }
func (v fltVal) Flag() flag        { return Float.Flag() }
func (v flt32Val) Flag() flag      { return Flt32.Flag() }
func (v bigFltVal) Flag() flag     { return BigFlt.Flag() }
func (v imagVal) Flag() flag       { return Imag.Flag() }
func (v imag64Val) Type() flag     { return Imag64.Flag() }
func (v ratioVal) Flag() flag      { return Ratio.Flag() }
func (v runeVal) Type() flag       { return Rune.Flag() }
func (v byteVal) Flag() flag       { return Byte.Flag() }
func (v bytesVal) Flag() flag      { return Bytes.Flag() }
func (v strVal) Flag() flag        { return String.Flag() }
func (v timeVal) Flag() flag       { return Time.Flag() }
func (v duraVal) Flag() flag       { return Duration.Flag() }
func (v slice) Flag() flag         { return Chain.Flag() }
func (v errorVal) Flag() flag      { return Error.Flag() }

//// BOUND TYPE FLAG METHODS ////
func (t flag) Uint() uint         { return fuint(t) }
func (t flag) Len() int           { return flen(t) }
func (t flag) Count() int         { return fcount(t) }
func (t flag) Least() int         { return fleast(t) }
func (t flag) Most() int          { return fmost(t) }
func (t flag) Low(v Typed) Typed  { return flow(t) }
func (t flag) High(v Typed) Typed { return fhigh(t) }
func (t flag) Reverse() flag      { return frev(t) }
func (t flag) Rotate(n int) flag  { return frot(t, n) }
func (t flag) Toggle(v flag) flag { return ftog(t, v) }
func (t flag) Concat(v flag) flag { return fconc(t, v) }
func (t flag) Mask(v flag) flag   { return fmask(t, v) }
func (t flag) Match(v Typed) bool { return fmatch(t, v) }

///// FREE TYPE FLAG METHOD IMPLEMENTATIONS /////
func fuint(t flag) uint         { return uint(t) }
func flen(t flag) int           { return bits.Len(uint(t)) }
func fcount(t flag) int         { return bits.OnesCount(uint(t)) }
func fleast(t flag) int         { return bits.TrailingZeros(uint(t)) + 1 }
func fmost(t flag) int          { return bits.LeadingZeros(uint(t)) - 1 }
func frev(t flag) flag          { return flag(bits.Reverse(uint(t))) }
func frot(t flag, n int) flag   { return flag(bits.RotateLeft(uint(t), n)) }
func ftog(t flag, v flag) flag  { return flag(uint(t) ^ v.Flag().Uint()) }
func fconc(t flag, v flag) flag { return flag(uint(t) | v.Flag().Uint()) }
func fmask(t flag, v flag) flag { return flag(uint(t) &^ v.Flag().Uint()) }
func fshow(f Typed) string      { return fmt.Sprintf("%64b\n", f) }
func flow(t Typed) Typed        { return fmask(t.Flag(), flag(Mask)) }
func fhigh(t Typed) Typed {
	len := flen(flag(Natives))
	return fmask(frot(t.Flag(), len), frot(flag(Natives), len))
}
func fmatch(t flag, v Typed) bool {
	if t.Uint()&v.Flag().Uint() != 0 {
		return true
	}
	return false
}

//////// IDENTITY & TYPE-REGISTER TYPES /////////
