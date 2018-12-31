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
	Const  // constant Data and fucntions
	Attr   // attribute special type
	Param  // parameter function
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
	Flag // generic bitflag role

	Nullable = Nil | Bool | Int | Int8 | Int16 | Int32 | BigInt | Uint |
		Uint8 | Uint16 | Uint32 | Float | Flt32 | BigFlt | Ratio | Imag |
		Imag64 | Byte | Rune | Bytes | String | Time | Duration |
		Attr | Error

	Numbers = Bool | Int | Int8 | Int16 | Int32 | BigInt | Uint | Uint8 |
		Uint16 | Uint32 | Float | Flt32 | BigFlt | Ratio | Imag |
		Imag64

	Internals = Flag

	Recursives = Tuple | List
	Chains     = Chain | AtList
	Sets       = UniSet | AtSet | Record
	Links      = Link | DLink | Node | Tree // Consumeables
	Composed   = Recursives | Chains | Sets | Links
	Natives    = Function | Nullable | Composed | Internals

	MAX_INT Type = 0xFFFFFFFFFFFFFFFF
	Mask         = MAX_INT ^ Natives
)

//////// INTERNAL TYPES /////////////
// internal types are typealiases without any wrapping, or referencing getting
// in the way performancewise. types need to be aliased in the first place, to
// associate them with a bitflag, without having to actually asign, let alone
// attach it to the instance.
type ( ////// INTERNAL TYPES /////
	BitFlag uint
	////// TYPE ALIASES ///////
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
)

func (c chain) Eval() Data    { return c }
func (c FLagSet) Eval() Data  { return c }
func (c ParamSet) Eval() Data { return c }
func (c cell) Eval() Data     { return c }
func (c cons) Eval() Data     { return c }

func (c unc) Eval() Data { return c }
func (c bnc) Eval() Data { return c }
func (c fnc) Eval() Data { return c }

func (c cell) Flag() BitFlag      { return Flag.Flag() }
func (c chain) Flag() BitFlag     { return Chain.Flag() }
func (c FLagSet) Flag() BitFlag   { return Flag.Flag() }
func (c ParamSet) Flag() BitFlag  { return Param.Flag() }
func (c ConstFnc) Flag() BitFlag  { return Const.Flag() }
func (c cons) Flag() BitFlag      { return Function.Flag() }
func (c unc) Flag() BitFlag       { return Function.Flag() }
func (c bnc) Flag() BitFlag       { return Function.Flag() }
func (c fnc) Flag() BitFlag       { return Function.Flag() }
func (c UnaryFnc) Flag() BitFlag  { return Function.Flag() }
func (c BinaryFnc) Flag() BitFlag { return Function.Flag() }
func (c NnaryFnc) Flag() BitFlag  { return Function.Flag() }

func (c cell) String() string      { return c.Data.String() + ": " + c.Data.String() }
func (c cons) String() string      { return c().String() }
func (c unc) String() string       { return "λ" }
func (c bnc) String() string       { return "λ" }
func (c fnc) String() string       { return "λ" }
func (c FLagSet) String() string   { return fshow((Flag | List)) }
func (c ParamSet) String() string  { return Param.Flag().String() }
func (c ConstFnc) String() string  { return "λ" }
func (c UnaryFnc) String() string  { return "λ" }
func (c BinaryFnc) String() string { return "λ" }
func (c NnaryFnc) String() string  { return "λ" }

//////// ATTRIBUTE TYPE ALIAS /////////////////
func conAttr(d Data) Attribute { return Attribute(d.Eval) }
func paramSetToData(p ParamSet) []Data {
	var data = []Data{}
	if len(p) == 0 {
		return []Data{nilVal{}}
	}
	for _, parm := range p {
		data = append(data, parm())
	}
	return data
}

/// bind the appropriate Flag Method to every type
func (a Attribute) Flag() BitFlag {
	return a().Flag().Concat(Attr.Flag())
}
func (v BitFlag) Flag() BitFlag       { return v }
func (nilVal) Flag() BitFlag          { return Nil.Flag() }
func (a Attribute) AttrType() BitFlag { return Attr.Flag() }
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
func (v errorVal) Flag() BitFlag      { return Error.Flag() }

///

func (nilVal) Eval() Data      { return nilVal{} }
func (v BitFlag) Eval() Data   { return v }
func (v boolVal) Eval() Data   { return v }
func (v intVal) Eval() Data    { return v }
func (v int8Val) Eval() Data   { return v }
func (v int16Val) Eval() Data  { return v }
func (v int32Val) Eval() Data  { return v }
func (v uintVal) Eval() Data   { return v }
func (v uint8Val) Eval() Data  { return v }
func (v uint16Val) Eval() Data { return v }
func (v uint32Val) Eval() Data { return v }
func (v bigIntVal) Eval() Data { return v }
func (v fltVal) Eval() Data    { return v }
func (v flt32Val) Eval() Data  { return v }
func (v bigFltVal) Eval() Data { return v }
func (v imagVal) Eval() Data   { return v }
func (v imag64Val) Eval() Data { return v }
func (v ratioVal) Eval() Data  { return v }
func (v runeVal) Eval() Data   { return v }
func (v byteVal) Eval() Data   { return v }
func (v bytesVal) Eval() Data  { return v }
func (v strVal) Eval() Data    { return v }
func (v timeVal) Eval() Data   { return v }
func (v duraVal) Eval() Data   { return v }
func (v errorVal) Eval() Data  { return v }

///
func (nilFnc) Flag() BitFlag      { return fconc(Nil, Function).Flag() }
func (v boolFnc) Flag() BitFlag   { return fconc(Bool, Function).Flag() }
func (v intFnc) Flag() BitFlag    { return fconc(Int, Function).Flag() }
func (v int8Fnc) Flag() BitFlag   { return fconc(Int8, Function).Flag() }
func (v int16Fnc) Flag() BitFlag  { return fconc(Int16, Function).Flag() }
func (v int32Fnc) Flag() BitFlag  { return fconc(Int32, Function).Flag() }
func (v uintFnc) Flag() BitFlag   { return fconc(Uint, Function).Flag() }
func (v uint8Fnc) Flag() BitFlag  { return fconc(Uint8, Function).Flag() }
func (v uint16Fnc) Flag() BitFlag { return fconc(Uint16, Function).Flag() }
func (v uint32Fnc) Flag() BitFlag { return fconc(Uint32, Function).Flag() }
func (v bigIntFnc) Flag() BitFlag { return fconc(BigInt, Function).Flag() }
func (v fltFnc) Flag() BitFlag    { return fconc(Float, Function).Flag() }
func (v flt32Fnc) Flag() BitFlag  { return fconc(Flt32, Function).Flag() }
func (v bigFltFnc) Flag() BitFlag { return fconc(BigFlt, Function).Flag() }
func (v imagFnc) Flag() BitFlag   { return fconc(Imag, Function).Flag() }
func (v imag64Fnc) Flag() BitFlag { return fconc(Imag64, Function).Flag() }
func (v ratioFnc) Flag() BitFlag  { return fconc(Ratio, Function).Flag() }
func (v runeFnc) Flag() BitFlag   { return fconc(Rune, Function).Flag() }
func (v byteFnc) Flag() BitFlag   { return fconc(Byte, Function).Flag() }
func (v bytesFnc) Flag() BitFlag  { return fconc(Bytes, Function).Flag() }
func (v strFnc) Flag() BitFlag    { return fconc(String, Function).Flag() }
func (v timeFnc) Flag() BitFlag   { return fconc(Time, Function).Flag() }
func (v duraFnc) Flag() BitFlag   { return fconc(Duration, Function).Flag() }
func (v errorFnc) Flag() BitFlag  { return fconc(Error, Function).Flag() }

//// BOUND TYPE FLAG METHODS ////
func (v BitFlag) Uint() uint               { return uint(v) }
func (v BitFlag) Len() int                 { return flen(v) }
func (v BitFlag) Count() int               { return fcount(v) }
func (v BitFlag) Least() int               { return fleast(v) }
func (v BitFlag) Most() int                { return fmost(v) }
func (v BitFlag) Low(f BitFlag) BitFlag    { return flow(f).Flag() }
func (v BitFlag) High(f BitFlag) BitFlag   { return fhigh(f).Flag() }
func (v BitFlag) Reverse() BitFlag         { return frev(v).Flag() }
func (v BitFlag) Rotate(n int) BitFlag     { return frot(v, n).Flag() }
func (v BitFlag) Toggle(f BitFlag) BitFlag { return ftog(v, f).Flag() }
func (v BitFlag) Concat(f BitFlag) BitFlag { return fconc(v, f).Flag() }
func (v BitFlag) Mask(f BitFlag) BitFlag   { return fmask(v, f).Flag() }
func (v BitFlag) Match(f BitFlag) bool     { return fmatch(v, f) }

///// FREE TYPE FLAG METHOD IMPLEMENTATIONS /////
func flag(t Typed) BitFlag           { return t.Flag() }
func flen(t Typed) int               { return bits.Len(t.Flag().Uint()) }
func fcount(t Typed) int             { return bits.OnesCount(t.Flag().Uint()) }
func fleast(t Typed) int             { return bits.TrailingZeros(t.Flag().Uint()) + 1 }
func fmost(t Typed) int              { return bits.LeadingZeros(t.Flag().Uint()) - 1 }
func frev(t Typed) BitFlag           { return BitFlag(bits.Reverse(t.Flag().Uint())) }
func frot(t Typed, n int) BitFlag    { return BitFlag(bits.RotateLeft(t.Flag().Uint(), n)) }
func ftog(t Typed, v Typed) BitFlag  { return BitFlag(t.Flag().Uint() ^ v.Flag().Uint()) }
func fconc(t Typed, v Typed) BitFlag { return BitFlag(t.Flag().Uint() | v.Flag().Uint()) }
func fmask(t Typed, v Typed) BitFlag { return BitFlag(t.Flag().Uint() &^ v.Flag().Uint()) }
func fshow(f Typed) string           { return fmt.Sprintf("%64b\n", f) }
func flow(t Typed) Typed             { return fmask(t.Flag(), Typed(Mask)) }
func fhigh(t BitFlag) BitFlag {
	len := flen(BitFlag(Natives))
	return fmask(frot(t.Flag(), len), frot(BitFlag(Natives), len))
}
func fmatch(t BitFlag, v BitFlag) bool {
	if t.Uint()&v.Flag().Uint() != 0 {
		return true
	}
	return false
}
