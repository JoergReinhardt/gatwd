package data

import (
	"bytes"
	"encoding/binary"
	"io"
	"math/big"
	"math/bits"
	"strings"
	"time"
)

//
// Nil
// Bool
// Int8
// Int16
// Int32
// Int
// BigInt
// Uint8
// Uint16
// Uint32
// Uint
// Flt32
// Float
// BigFlt
// Ratio
// Imag64
// Imag
// Time
// Duration
// Byte
// Rune
// Flag
// String
// Bytes
// Error

func TypePrec(a, b Native) (x, y Native) {
	// if arguments types happend to be different‥.
	if at, bt := a.Type(), b.Type(); at != bt {
		var ati, bti = at.Flag().Index(), bt.Flag().Index()
		if ati > bti {
			// return unchanged a and cast b as type of a
			return a, TypeConversionTable[bti][ati](b)
		}
		// cast a as type of b and return unchanged b
		return TypeConversionTable[ati][bti](a), b
	}
	// ‥.otherwise just return both arguments unchagend
	return a, b
}

// BOOL VALUE
func (v BoolVal) Ratio() *RatioVal     { return v.IntVal().Ratio() }
func (v BoolVal) GoRat() *big.Rat      { return (*big.Rat)(v.Ratio()) }
func (v BoolVal) GoBigInt() *big.Int   { return big.NewInt(int64(v.Int())) }
func (v BoolVal) GoBigFlt() *big.Float { return big.NewFloat(v.GoFlt()) }
func (v BoolVal) BigInt() *BigIntVal   { return (*BigIntVal)(v.GoBigInt()) }
func (v BoolVal) BigFlt() *BigFltVal   { return (*BigFltVal)(v.GoBigFlt()) }
func (v BoolVal) Unit() Native         { return BoolVal(true) }
func (v BoolVal) Int() IntVal          { return IntVal(v.GoInt()) }
func (v BoolVal) Uint() UintVal        { return UintVal(uint(v.Int())) }
func (v BoolVal) Float() FltVal        { return v.IntVal().Float() }
func (v BoolVal) Imag() ImagVal        { return v.IntVal().Imag() }
func (v BoolVal) Bool() bool           { return bool(v) }
func (v BoolVal) Idx() int             { return int(v.Int()) }
func (v BoolVal) GoFlt() float64       { return float64(v.Float()) }
func (v BoolVal) GoImag() complex128   { return complex128(v.Imag()) }
func (v BoolVal) GoUint() uint {
	if v {
		return 1
	}
	return 0
}
func (v BoolVal) GoInt() int {
	if v {
		return 1
	}
	return -1
}
func (v BoolVal) IntVal() IntVal {
	if v {
		return IntVal(1)
	}
	return IntVal(-1)
}

// operators
func (v BoolVal) Not() BoolVal            { return !v }
func (v BoolVal) Or(arg BoolVal) BoolVal  { return v || arg }
func (v BoolVal) And(arg BoolVal) BoolVal { return v && arg }

// comparators
func (v BoolVal) Equal(arg BoolVal) bool {
	if v && arg {
		return true
	}
	if !v && !arg {
		return true
	}
	return false
}
func (v BoolVal) Lesser(arg BoolVal) bool {
	if !v {
		if arg {
			return true
		}
	}
	return false
}
func (v BoolVal) Greater(arg BoolVal) bool {
	if v {
		if !arg {
			return true
		}
	}
	return false
}

// NATURAL VALUE
// Uint8Val
// conversions
func (v Uint8Val) Idx() int             { return int(v.Int()) }
func (v Uint8Val) GoInt() int           { return int(v.Int()) }
func (v Uint8Val) GoUint() uint         { return uint(v.Uint()) }
func (v Uint8Val) GoFlt() float64       { return float64(v.Float()) }
func (v Uint8Val) GoImag() complex128   { return complex128(v.Imag()) }
func (v Uint8Val) GoRat() *big.Rat      { return (*big.Rat)(v.Ratio()) }
func (v Uint8Val) GoBigInt() *big.Int   { return big.NewInt(int64(v.Int())) }
func (v Uint8Val) GoBigFlt() *big.Float { return big.NewFloat(v.GoFlt()) }
func (v Uint8Val) BigInt() *BigIntVal   { return (*BigIntVal)(v.GoBigInt()) }
func (v Uint8Val) BigFlt() *BigFltVal   { return (*BigFltVal)(v.GoBigFlt()) }
func (v Uint8Val) Unit() Native         { return Uint8Val(1) }
func (v Uint8Val) Uint() UintVal        { return UintVal(uint64(v)) }
func (v Uint8Val) Uint8() Uint8Val      { return v }
func (v Uint8Val) Uint16() Uint16Val    { return Uint16Val(uint16(v)) }
func (v Uint8Val) Uint32() Uint32Val    { return Uint32Val(uint32(v)) }
func (v Uint8Val) Int() IntVal          { return IntVal(int(v)) }
func (v Uint8Val) IntVal() IntVal       { return IntVal(int(v)) }
func (v Uint8Val) Bool() BoolVal        { return v.BoolVal().(BoolVal) }
func (v Uint8Val) Float() FltVal        { return FltVal(float64(v)) }
func (v Uint8Val) Imag() ImagVal        { return v.IntVal().Imag() }
func (v Uint8Val) Ratio() *RatioVal {
	var rat = big.NewRat(int64(v), 1)
	return (*RatioVal)(rat)
}
func (v Uint8Val) BoolVal() Native {
	if v > 0 {
		return BoolVal(true)
	}
	return BoolVal(false)
}

// operators
func (v Uint8Val) Not() Uint8Val                   { return ^v }
func (v Uint8Val) And(arg Uint8Val) Uint8Val       { return v & arg }
func (v Uint8Val) Xor(arg Uint8Val) Uint8Val       { return v ^ arg }
func (v Uint8Val) Or(arg Uint8Val) Uint8Val        { return v | arg }
func (v Uint8Val) AndNot(arg Uint8Val) Uint8Val    { return v &^ arg }
func (v Uint8Val) Negate() Uint8Val                { return -v }
func (v Uint8Val) Add(arg Uint8Val) Uint8Val       { return v + arg }
func (v Uint8Val) Substract(arg Uint8Val) Uint8Val { return v - arg }
func (v Uint8Val) Multipy(arg Uint8Val) Uint8Val   { return v * arg }
func (v Uint8Val) Power(arg Uint8Val) Uint8Val     { return v ^ arg }
func (v Uint8Val) Quotient(arg Uint8Val) Uint8Val  { return v / arg }
func (v Uint8Val) QuoRatio(arg Uint8Val) *RatioVal {
	return (*RatioVal)(big.NewRat(int64(v), int64(arg)))
}

// comparators
func (v Uint8Val) Equal(arg Uint8Val) bool   { return arg == v }
func (v Uint8Val) Lesser(arg Uint8Val) bool  { return arg < v }
func (v Uint8Val) Greater(arg Uint8Val) bool { return arg > v }

// Uint16Val
// conversions
func (v Uint16Val) Idx() int             { return int(v.Int()) }
func (v Uint16Val) GoInt() int           { return int(v.Int()) }
func (v Uint16Val) GoUint() uint         { return uint(v.Uint()) }
func (v Uint16Val) GoFlt() float64       { return float64(v.Float()) }
func (v Uint16Val) GoImag() complex128   { return complex128(v.Imag()) }
func (v Uint16Val) GoRat() *big.Rat      { return (*big.Rat)(v.Ratio()) }
func (v Uint16Val) GoBigInt() *big.Int   { return big.NewInt(int64(v.Int())) }
func (v Uint16Val) GoBigFlt() *big.Float { return big.NewFloat(v.GoFlt()) }
func (v Uint16Val) BigInt() *BigIntVal   { return (*BigIntVal)(v.GoBigInt()) }
func (v Uint16Val) BigFlt() *BigFltVal   { return (*BigFltVal)(v.GoBigFlt()) }
func (v Uint16Val) Unit() Native         { return Uint16Val(1) }
func (v Uint16Val) Uint() UintVal        { return UintVal(uint64(v)) }
func (v Uint16Val) Uint8() Uint8Val      { return Uint8Val(uint8(v)) }
func (v Uint16Val) Uint16() Uint16Val    { return v }
func (v Uint16Val) Uint32() Uint32Val    { return Uint32Val(uint32(v)) }
func (v Uint16Val) Int() IntVal          { return IntVal(int(v)) }
func (v Uint16Val) IntVal() IntVal       { return IntVal(int(v)) }
func (v Uint16Val) Bool() BoolVal        { return v.BoolVal().(BoolVal) }
func (v Uint16Val) Float() FltVal        { return FltVal(float64(v)) }
func (v Uint16Val) Imag() ImagVal        { return v.IntVal().Imag() }
func (v Uint16Val) Ratio() *RatioVal {
	var rat = big.NewRat(int64(v), 1)
	return (*RatioVal)(rat)
}
func (v Uint16Val) BoolVal() Native {
	if v > 0 {
		return BoolVal(true)
	}
	return BoolVal(false)
}

// operators
func (v Uint16Val) Not() Uint16Val                    { return ^v }
func (v Uint16Val) And(arg Uint16Val) Uint16Val       { return v & arg }
func (v Uint16Val) Xor(arg Uint16Val) Uint16Val       { return v ^ arg }
func (v Uint16Val) Or(arg Uint16Val) Uint16Val        { return v | arg }
func (v Uint16Val) AndNot(arg Uint16Val) Uint16Val    { return v &^ arg }
func (v Uint16Val) Negate() Uint16Val                 { return -v }
func (v Uint16Val) Add(arg Uint16Val) Uint16Val       { return v + arg }
func (v Uint16Val) Substract(arg Uint16Val) Uint16Val { return v - arg }
func (v Uint16Val) Multipy(arg Uint16Val) Uint16Val   { return v * arg }
func (v Uint16Val) Power(arg Uint16Val) Uint16Val     { return v ^ arg }
func (v Uint16Val) Quotient(arg Uint16Val) Uint16Val  { return v / arg }
func (v Uint16Val) QuoRatio(arg Uint16Val) *RatioVal {
	return (*RatioVal)(big.NewRat(int64(v), int64(arg)))
}

// comparators
func (v Uint16Val) Equal(arg Uint16Val) bool   { return arg == v }
func (v Uint16Val) Lesser(arg Uint16Val) bool  { return arg < v }
func (v Uint16Val) Greater(arg Uint16Val) bool { return arg > v }

// Uint32Val
// conversions
func (v Uint32Val) Idx() int             { return int(v.Int()) }
func (v Uint32Val) GoInt() int           { return int(v.Int()) }
func (v Uint32Val) GoUint() uint         { return uint(v.Uint()) }
func (v Uint32Val) GoFlt() float64       { return float64(v.Float()) }
func (v Uint32Val) GoImag() complex128   { return complex128(v.Imag()) }
func (v Uint32Val) GoRat() *big.Rat      { return (*big.Rat)(v.Ratio()) }
func (v Uint32Val) GoBigInt() *big.Int   { return big.NewInt(int64(v.Int())) }
func (v Uint32Val) GoBigFlt() *big.Float { return big.NewFloat(v.GoFlt()) }
func (v Uint32Val) BigInt() *BigIntVal   { return (*BigIntVal)(v.GoBigInt()) }
func (v Uint32Val) BigFlt() *BigFltVal   { return (*BigFltVal)(v.GoBigFlt()) }
func (v Uint32Val) Unit() Native         { return Uint32Val(1) }
func (v Uint32Val) Uint() UintVal        { return UintVal(uint64(v)) }
func (v Uint32Val) Uint8() Uint8Val      { return Uint8Val(uint8(v)) }
func (v Uint32Val) Uint16() Uint16Val    { return Uint16Val(uint16(v)) }
func (v Uint32Val) Uint32() Uint32Val    { return v }
func (v Uint32Val) Int() IntVal          { return IntVal(int(v)) }
func (v Uint32Val) IntVal() IntVal       { return IntVal(int(v)) }
func (v Uint32Val) Bool() BoolVal        { return v.BoolVal().(BoolVal) }
func (v Uint32Val) Float() FltVal        { return FltVal(float64(v)) }
func (v Uint32Val) Imag() ImagVal        { return v.IntVal().Imag() }
func (v Uint32Val) Ratio() *RatioVal {
	var rat = big.NewRat(int64(v), 1)
	return (*RatioVal)(rat)
}
func (v Uint32Val) BoolVal() Native {
	if v > 0 {
		return BoolVal(true)
	}
	return BoolVal(false)
}

// operators
func (v Uint32Val) Not() Uint32Val                    { return ^v }
func (v Uint32Val) And(arg Uint32Val) Uint32Val       { return v & arg }
func (v Uint32Val) Xor(arg Uint32Val) Uint32Val       { return v ^ arg }
func (v Uint32Val) Or(arg Uint32Val) Uint32Val        { return v | arg }
func (v Uint32Val) AndNot(arg Uint32Val) Uint32Val    { return v &^ arg }
func (v Uint32Val) Negate() Uint32Val                 { return -v }
func (v Uint32Val) Add(arg Uint32Val) Uint32Val       { return v + arg }
func (v Uint32Val) Substract(arg Uint32Val) Uint32Val { return v - arg }
func (v Uint32Val) Multipy(arg Uint32Val) Uint32Val   { return v * arg }
func (v Uint32Val) Power(arg Uint32Val) Uint32Val     { return v ^ arg }
func (v Uint32Val) Quotient(arg Uint32Val) Uint32Val  { return v / arg }
func (v Uint32Val) QuoRatio(arg Uint32Val) *RatioVal {
	return (*RatioVal)(big.NewRat(int64(v), int64(arg)))
}

// comparators
func (v Uint32Val) Equal(arg Uint32Val) bool   { return arg == v }
func (v Uint32Val) Lesser(arg Uint32Val) bool  { return arg < v }
func (v Uint32Val) Greater(arg Uint32Val) bool { return arg > v }

// UintVal
// conversions
func (v UintVal) Idx() int             { return int(v.Int()) }
func (v UintVal) GoInt() int           { return int(v.Int()) }
func (v UintVal) GoUint() uint         { return uint(v.Uint()) }
func (v UintVal) GoFlt() float64       { return float64(v.Float()) }
func (v UintVal) GoImag() complex128   { return complex128(v.Imag()) }
func (v UintVal) GoRat() *big.Rat      { return (*big.Rat)(v.Ratio()) }
func (v UintVal) GoBigInt() *big.Int   { return big.NewInt(int64(v.Int())) }
func (v UintVal) GoBigFlt() *big.Float { return big.NewFloat(v.GoFlt()) }
func (v UintVal) BigInt() *BigIntVal   { return (*BigIntVal)(v.GoBigInt()) }
func (v UintVal) BigFlt() *BigFltVal   { return (*BigFltVal)(v.GoBigFlt()) }
func (v UintVal) Unit() Native         { return UintVal(1) }
func (v UintVal) Uint() UintVal        { return v }
func (v UintVal) Uint8() Uint8Val      { return Uint8Val(uint8(v)) }
func (v UintVal) Uint16() Uint16Val    { return Uint16Val(uint16(v)) }
func (v UintVal) Uint32() Uint32Val    { return Uint32Val(uint32(v)) }
func (v UintVal) Int() IntVal          { return IntVal(int(v)) }
func (v UintVal) IntVal() IntVal       { return IntVal(int(v)) }
func (v UintVal) Bool() BoolVal        { return v.BoolVal().(BoolVal) }
func (v UintVal) Float() FltVal        { return FltVal(float64(v)) }
func (v UintVal) Imag() ImagVal        { return v.IntVal().Imag() }
func (v UintVal) Ratio() *RatioVal {
	var rat = big.NewRat(int64(v), 1)
	return (*RatioVal)(rat)
}
func (v UintVal) BoolVal() Native {
	if v > 0 {
		return BoolVal(true)
	}
	return BoolVal(false)
}

// operators
func (v UintVal) Not() UintVal                  { return ^v }
func (v UintVal) And(arg UintVal) UintVal       { return v & arg }
func (v UintVal) Xor(arg UintVal) UintVal       { return v ^ arg }
func (v UintVal) Or(arg UintVal) UintVal        { return v | arg }
func (v UintVal) AndNot(arg UintVal) UintVal    { return v &^ arg }
func (v UintVal) Negate() UintVal               { return -v }
func (v UintVal) Add(arg UintVal) UintVal       { return v + arg }
func (v UintVal) Substract(arg UintVal) UintVal { return v - arg }
func (v UintVal) Multipy(arg UintVal) UintVal   { return v * arg }
func (v UintVal) Power(arg UintVal) UintVal     { return v ^ arg }
func (v UintVal) Quotient(arg UintVal) UintVal  { return v / arg }
func (v UintVal) QuoRatio(arg UintVal) *RatioVal {
	return (*RatioVal)(big.NewRat(int64(v), int64(arg)))
}

// comparators
func (v UintVal) Equal(arg UintVal) bool   { return arg == v }
func (v UintVal) Lesser(arg UintVal) bool  { return arg < v }
func (v UintVal) Greater(arg UintVal) bool { return arg > v }

// INTEGER VALUE
// Int8Val
func (v Int8Val) GoInt() int           { return int(v) }
func (v Int8Val) GoFlt() float64       { return float64(v) }
func (v Int8Val) GoUint() uint         { return uint(v) }
func (v Int8Val) GoImag() complex128   { return complex128(v.Imag()) }
func (v Int8Val) GoRat() *big.Rat      { return (*big.Rat)(v.Ratio()) }
func (v Int8Val) GoBigInt() *big.Int   { return big.NewInt(int64(v)) }
func (v Int8Val) GoBigFlt() *big.Float { return big.NewFloat(float64(v)) }
func (v Int8Val) BigInt() *BigIntVal   { return (*BigIntVal)(v.GoBigInt()) }
func (v Int8Val) BigFlt() *BigFltVal   { return (*BigFltVal)(v.GoBigFlt()) }
func (v Int8Val) Unit() Native         { return UintVal(uint(v)) }
func (v Int8Val) Byte() Native         { return Int8Val(byte(v.Int8())) }
func (v Int8Val) Rune() Native         { return RuneVal(rune(v.Int32())) }
func (v Int8Val) Flag() Native         { return BitFlag(uint(v)) }
func (v Int8Val) Int8() Int8Val        { return v }
func (v Int8Val) Int16() Int16Val      { return Int16Val(int16(v)) }
func (v Int8Val) Int32() Int32Val      { return Int32Val(int32(v)) }
func (v Int8Val) Int() IntVal          { return IntVal(int64(v)) }
func (v Int8Val) Float() FltVal        { return FltVal(float64(v)) }
func (v Int8Val) Imag() ImagVal        { return ImagVal(complex(v.Float(), 1.0)) }
func (v Int8Val) Idx() int             { return int(v) }
func (v Int8Val) Ratio() *RatioVal {
	var rat = big.NewRat(1, int64(v))
	return (*RatioVal)(rat)
}
func (v Int8Val) Bool() BoolVal {
	if v > 0 {
		return BoolVal(true)
	}
	return BoolVal(false)
}
func (v Int8Val) Truth() Native {
	if v < 0 {
		return BoolVal(false)
	}
	if v > 0 {
		return BoolVal(true)
	}
	return NilVal{}
}
func (v Int8Val) Uint() UintVal {
	if v < 0 {
		return UintVal(v * -1)
	}
	return UintVal(v)
}

// operators
func (v Int8Val) Not() Int8Val                  { return ^v }
func (v Int8Val) And(arg Int8Val) Int8Val       { return v & arg }
func (v Int8Val) Xor(arg Int8Val) Int8Val       { return v ^ arg }
func (v Int8Val) Or(arg Int8Val) Int8Val        { return v | arg }
func (v Int8Val) AndNot(arg Int8Val) Int8Val    { return v &^ arg }
func (v Int8Val) Negate() Int8Val               { return -v }
func (v Int8Val) Add(arg Int8Val) Int8Val       { return v + arg }
func (v Int8Val) Substract(arg Int8Val) Int8Val { return v - arg }
func (v Int8Val) Multipy(arg Int8Val) Int8Val   { return v * arg }
func (v Int8Val) Power(arg Int8Val) Int8Val     { return v ^ arg }
func (v Int8Val) Quotient(arg Int8Val) Int8Val  { return v / arg }

// comparators
func (v Int8Val) Equal(arg Int8Val) bool   { return arg == v }
func (v Int8Val) Lesser(arg Int8Val) bool  { return arg < v }
func (v Int8Val) Greater(arg Int8Val) bool { return arg > v }

// Int16Val
func (v Int16Val) GoInt() int           { return int(v) }
func (v Int16Val) GoFlt() float64       { return float64(v) }
func (v Int16Val) GoUint() uint         { return uint(v) }
func (v Int16Val) GoImag() complex128   { return complex128(v.Imag()) }
func (v Int16Val) GoRat() *big.Rat      { return (*big.Rat)(v.Ratio()) }
func (v Int16Val) GoBigInt() *big.Int   { return big.NewInt(int64(v)) }
func (v Int16Val) GoBigFlt() *big.Float { return big.NewFloat(float64(v)) }
func (v Int16Val) BigInt() *BigIntVal   { return (*BigIntVal)(v.GoBigInt()) }
func (v Int16Val) BigFlt() *BigFltVal   { return (*BigFltVal)(v.GoBigFlt()) }
func (v Int16Val) Unit() Native         { return UintVal(uint(v)) }
func (v Int16Val) Byte() Native         { return Int16Val(byte(v.Int16())) }
func (v Int16Val) Rune() Native         { return RuneVal(rune(v.Int32())) }
func (v Int16Val) Flag() Native         { return BitFlag(uint(v)) }
func (v Int16Val) Int8() Int8Val        { return Int8Val(int8(v)) }
func (v Int16Val) Int16() Int16Val      { return v }
func (v Int16Val) Int32() Int32Val      { return Int32Val(int32(v)) }
func (v Int16Val) Int() IntVal          { return IntVal(int64(v)) }
func (v Int16Val) Float() FltVal        { return FltVal(float64(v)) }
func (v Int16Val) Imag() ImagVal        { return ImagVal(complex(v.Float(), 1.0)) }
func (v Int16Val) Idx() int             { return int(v) }
func (v Int16Val) Ratio() *RatioVal {
	var rat = big.NewRat(1, int64(v))
	return (*RatioVal)(rat)
}
func (v Int16Val) Bool() BoolVal {
	if v > 0 {
		return BoolVal(true)
	}
	return BoolVal(false)
}
func (v Int16Val) Truth() Native {
	if v < 0 {
		return BoolVal(false)
	}
	if v > 0 {
		return BoolVal(true)
	}
	return NilVal{}
}
func (v Int16Val) Uint() UintVal {
	if v < 0 {
		return UintVal(v * -1)
	}
	return UintVal(v)
}

// operators
func (v Int16Val) Not() Int16Val                   { return ^v }
func (v Int16Val) And(arg Int16Val) Int16Val       { return v & arg }
func (v Int16Val) Xor(arg Int16Val) Int16Val       { return v ^ arg }
func (v Int16Val) Or(arg Int16Val) Int16Val        { return v | arg }
func (v Int16Val) AndNot(arg Int16Val) Int16Val    { return v &^ arg }
func (v Int16Val) Negate() Int16Val                { return -v }
func (v Int16Val) Add(arg Int16Val) Int16Val       { return v + arg }
func (v Int16Val) Substract(arg Int16Val) Int16Val { return v - arg }
func (v Int16Val) Multipy(arg Int16Val) Int16Val   { return v * arg }
func (v Int16Val) Power(arg Int16Val) Int16Val     { return v ^ arg }
func (v Int16Val) Quotient(arg Int16Val) Int16Val  { return v / arg }

// comparators
func (v Int16Val) Equal(arg Int16Val) bool   { return arg == v }
func (v Int16Val) Lesser(arg Int16Val) bool  { return arg < v }
func (v Int16Val) Greater(arg Int16Val) bool { return arg > v }

// Int32Val
func (v Int32Val) GoInt() int           { return int(v) }
func (v Int32Val) GoFlt() float64       { return float64(v) }
func (v Int32Val) GoUint() uint         { return uint(v) }
func (v Int32Val) GoImag() complex128   { return complex128(v.Imag()) }
func (v Int32Val) GoRat() *big.Rat      { return (*big.Rat)(v.Ratio()) }
func (v Int32Val) GoBigInt() *big.Int   { return big.NewInt(int64(v)) }
func (v Int32Val) GoBigFlt() *big.Float { return big.NewFloat(float64(v)) }
func (v Int32Val) BigInt() *BigIntVal   { return (*BigIntVal)(v.GoBigInt()) }
func (v Int32Val) BigFlt() *BigFltVal   { return (*BigFltVal)(v.GoBigFlt()) }
func (v Int32Val) Unit() Native         { return UintVal(uint(v)) }
func (v Int32Val) Byte() Native         { return Int32Val(byte(v.Int32())) }
func (v Int32Val) Rune() Native         { return RuneVal(rune(v.Int32())) }
func (v Int32Val) Flag() Native         { return BitFlag(uint(v)) }
func (v Int32Val) Int8Val() Int8Val     { return Int8Val(int8(v)) }
func (v Int32Val) Int16() Int16Val      { return Int16Val(int16(v)) }
func (v Int32Val) Int32() Int32Val      { return v }
func (v Int32Val) Int() IntVal          { return IntVal(int64(v)) }
func (v Int32Val) Float() FltVal        { return FltVal(float64(v)) }
func (v Int32Val) Imag() ImagVal        { return ImagVal(complex(v.Float(), 1.0)) }
func (v Int32Val) Idx() int             { return int(v) }
func (v Int32Val) Ratio() *RatioVal {
	var rat = big.NewRat(1, int64(v))
	return (*RatioVal)(rat)
}
func (v Int32Val) Bool() BoolVal {
	if v > 0 {
		return BoolVal(true)
	}
	return BoolVal(false)
}
func (v Int32Val) Truth() Native {
	if v < 0 {
		return BoolVal(false)
	}
	if v > 0 {
		return BoolVal(true)
	}
	return NilVal{}
}
func (v Int32Val) Uint() UintVal {
	if v < 0 {
		return UintVal(v * -1)
	}
	return UintVal(v)
}

// operators
func (v Int32Val) Not() Int32Val                   { return ^v }
func (v Int32Val) And(arg Int32Val) Int32Val       { return v & arg }
func (v Int32Val) Xor(arg Int32Val) Int32Val       { return v ^ arg }
func (v Int32Val) Or(arg Int32Val) Int32Val        { return v | arg }
func (v Int32Val) AndNot(arg Int32Val) Int32Val    { return v &^ arg }
func (v Int32Val) Negate() Int32Val                { return -v }
func (v Int32Val) Add(arg Int32Val) Int32Val       { return v + arg }
func (v Int32Val) Substract(arg Int32Val) Int32Val { return v - arg }
func (v Int32Val) Multipy(arg Int32Val) Int32Val   { return v * arg }
func (v Int32Val) Power(arg Int32Val) Int32Val     { return v ^ arg }
func (v Int32Val) Quotient(arg Int32Val) Int32Val  { return v / arg }

// comparators
func (v Int32Val) Equal(arg Int32Val) bool   { return arg == v }
func (v Int32Val) Lesser(arg Int32Val) bool  { return arg < v }
func (v Int32Val) Greater(arg Int32Val) bool { return arg > v }

// IntVal
// conversions
func (v IntVal) GoInt() int           { return int(v) }
func (v IntVal) GoFlt() float64       { return float64(v) }
func (v IntVal) GoUint() uint         { return uint(v) }
func (v IntVal) GoImag() complex128   { return complex128(v.Imag()) }
func (v IntVal) GoRat() *big.Rat      { return (*big.Rat)(v.Ratio()) }
func (v IntVal) GoBigInt() *big.Int   { return big.NewInt(int64(v)) }
func (v IntVal) GoBigFlt() *big.Float { return big.NewFloat(float64(v)) }
func (v IntVal) BigInt() *BigIntVal   { return (*BigIntVal)(v.GoBigInt()) }
func (v IntVal) BigFlt() *BigFltVal   { return (*BigFltVal)(v.GoBigFlt()) }
func (v IntVal) Unit() Native         { return UintVal(uint(v)) }
func (v IntVal) Byte() Native         { return IntVal(byte(v.Int8())) }
func (v IntVal) Rune() Native         { return RuneVal(rune(v.Int32())) }
func (v IntVal) Flag() Native         { return BitFlag(uint(v)) }
func (v IntVal) IntVal() IntVal       { return v }
func (v IntVal) Int8() IntVal         { return IntVal(int8(v)) }
func (v IntVal) Int16() Int16Val      { return Int16Val(int16(v)) }
func (v IntVal) Int32() Int32Val      { return Int32Val(int32(v)) }
func (v IntVal) Int() IntVal          { return v }
func (v IntVal) Float() FltVal        { return FltVal(float64(v)) }
func (v IntVal) Imag() ImagVal        { return ImagVal(complex(v.Float(), 1.0)) }
func (v IntVal) Idx() int             { return int(v) }
func (v IntVal) Ratio() *RatioVal {
	var rat = big.NewRat(1, int64(v))
	return (*RatioVal)(rat)
}
func (v IntVal) Bool() BoolVal {
	if v > 0 {
		return BoolVal(true)
	}
	return BoolVal(false)
}
func (v IntVal) Truth() Native {
	if v < 0 {
		return BoolVal(false)
	}
	if v > 0 {
		return BoolVal(true)
	}
	return NilVal{}
}
func (v IntVal) Uint() UintVal {
	if v < 0 {
		return UintVal(v * -1)
	}
	return UintVal(v)
}

// operators
func (v IntVal) Not() IntVal                 { return ^v }
func (v IntVal) And(arg IntVal) IntVal       { return v & arg }
func (v IntVal) Xor(arg IntVal) IntVal       { return v ^ arg }
func (v IntVal) Or(arg IntVal) IntVal        { return v | arg }
func (v IntVal) AndNot(arg IntVal) IntVal    { return v &^ arg }
func (v IntVal) Negate() IntVal              { return -v }
func (v IntVal) Add(arg IntVal) IntVal       { return v + arg }
func (v IntVal) Substract(arg IntVal) IntVal { return v - arg }
func (v IntVal) Multipy(arg IntVal) IntVal   { return v * arg }
func (v IntVal) Power(arg IntVal) IntVal     { return v ^ arg }
func (v IntVal) Quotient(arg IntVal) IntVal  { return v / arg }

// comparators
func (v IntVal) Equal(arg IntVal) bool   { return arg == v }
func (v IntVal) Lesser(arg IntVal) bool  { return arg < v }
func (v IntVal) Greater(arg IntVal) bool { return arg > v }

// REAL VALUE
func (v FltVal) Unit() Native         { return FltVal(1.0) }
func (v FltVal) Idx() int             { return int(v) }
func (v FltVal) GoInt() int           { return int(v) }
func (v FltVal) GoUint() uint         { return uint(v) }
func (v FltVal) GoFlt() float64       { return float64(v) }
func (v FltVal) GoImag() complex128   { return complex128(v.Imag()) }
func (v FltVal) GoRat() *big.Rat      { return (*big.Rat)(v.Ratio()) }
func (v FltVal) GoBigInt() *big.Int   { return big.NewInt(int64(v.GoInt())) }
func (v FltVal) GoBigFlt() *big.Float { return big.NewFloat(v.GoFlt()) }
func (v FltVal) BigInt() *BigIntVal   { return (*BigIntVal)(v.GoBigInt()) }
func (v FltVal) BigFlt() *BigFltVal   { return (*BigFltVal)(v.GoBigFlt()) }
func (v FltVal) Uint() UintVal        { return UintVal(uint(v)) }
func (v FltVal) Int() IntVal          { return IntVal(int(v)) }
func (v FltVal) Float() FltVal        { return v }
func (v FltVal) Flt32() Flt32Val      { return Flt32Val(float32(v)) }
func (v FltVal) Imag() ImagVal        { return ImagVal(complex(v, 1.0)) }
func (v FltVal) Ratio() *RatioVal {
	var rat = big.NewRat(int64(1), int64(1)).SetFloat64(v.GoFlt())
	return (*RatioVal)(rat)
}
func (v FltVal) Bool() BoolVal {
	if v > 0.0 {
		return BoolVal(true)
	}
	return BoolVal(false)
}
func (v FltVal) Truth() Native {
	if v < 0.0 {
		return BoolVal(false)
	}
	if v > 0.0 {
		return BoolVal(true)
	}
	return NilVal{}
}

// operators
func (v FltVal) Negate() FltVal              { return -v }
func (v FltVal) Add(arg FltVal) FltVal       { return v + arg }
func (v FltVal) Substract(arg FltVal) FltVal { return v - arg }
func (v FltVal) Multipy(arg FltVal) FltVal   { return v * arg }
func (v FltVal) Quotient(arg FltVal) FltVal  { return v / arg }

// comparators
func (v FltVal) Equal(arg FltVal) bool   { return v == arg }
func (v FltVal) Lesser(arg FltVal) bool  { return v < arg }
func (v FltVal) Greater(arg FltVal) bool { return v > arg }

// REAL VALUE
func (v Flt32Val) Unit() Native         { return Flt32Val(1.0) }
func (v Flt32Val) Idx() int             { return int(v) }
func (v Flt32Val) GoInt() int           { return int(v) }
func (v Flt32Val) GoUint() uint         { return uint(v) }
func (v Flt32Val) GoFlt() float64       { return float64(v) }
func (v Flt32Val) GoImag() complex128   { return complex128(v.Imag()) }
func (v Flt32Val) GoRat() *big.Rat      { return (*big.Rat)(v.Ratio()) }
func (v Flt32Val) GoBigInt() *big.Int   { return big.NewInt(int64(v.GoInt())) }
func (v Flt32Val) GoBigFlt() *big.Float { return big.NewFloat(v.GoFlt()) }
func (v Flt32Val) BigInt() *BigIntVal   { return (*BigIntVal)(v.GoBigInt()) }
func (v Flt32Val) BigFlt() *BigFltVal   { return (*BigFltVal)(v.GoBigFlt()) }
func (v Flt32Val) Uint() UintVal        { return UintVal(uint(v)) }
func (v Flt32Val) Int() IntVal          { return IntVal(int(v)) }
func (v Flt32Val) Float() FltVal        { return FltVal(float64(v)) }
func (v Flt32Val) Flt32() Flt32Val      { return v }
func (v Flt32Val) Imag() ImagVal        { return ImagVal(complex(v, 1.0)) }
func (v Flt32Val) Ratio() *RatioVal {
	var rat = big.NewRat(int64(1), int64(1)).SetFloat64(v.GoFlt())
	return (*RatioVal)(rat)
}
func (v Flt32Val) Bool() BoolVal {
	if v > 0.0 {
		return BoolVal(true)
	}
	return BoolVal(false)
}
func (v Flt32Val) Truth() Native {
	if v < 0.0 {
		return BoolVal(false)
	}
	if v > 0.0 {
		return BoolVal(true)
	}
	return NilVal{}
}

// operators
func (v Flt32Val) Negate() Flt32Val                { return -v }
func (v Flt32Val) Add(arg Flt32Val) Flt32Val       { return v + arg }
func (v Flt32Val) Substract(arg Flt32Val) Flt32Val { return v - arg }
func (v Flt32Val) Multipy(arg Flt32Val) Flt32Val   { return v * arg }
func (v Flt32Val) Quotient(arg Flt32Val) Flt32Val  { return v / arg }

// comparators
func (v Flt32Val) Equal(arg Flt32Val) bool   { return v == arg }
func (v Flt32Val) Lesser(arg Flt32Val) bool  { return v < arg }
func (v Flt32Val) Greater(arg Flt32Val) bool { return v > arg }

// RATIONAL VALUE
func (v RatioVal) Idx() int               { return int(v.Int()) }
func (v RatioVal) GoInt() int             { return int(v.Int()) }
func (v RatioVal) GoUint() uint           { return uint(v.Uint()) }
func (v RatioVal) GoFlt() float64         { return float64(v.Float()) }
func (v RatioVal) GoImag() complex128     { return complex128(v.Imag()) }
func (v RatioVal) GoRat() *big.Rat        { return (*big.Rat)(&v) }
func (v RatioVal) GoBigInt() *big.Int     { return big.NewInt(int64(v.GoInt())) }
func (v RatioVal) GoBigFlt() *big.Float   { return big.NewFloat(v.GoFlt()) }
func (v RatioVal) BigInt() *BigIntVal     { return (*BigIntVal)(v.GoBigInt()) }
func (v RatioVal) BigFlt() *BigFltVal     { return (*BigFltVal)(v.GoBigFlt()) }
func (v RatioVal) Unit() Native           { return RatioVal(*big.NewRat(1, 1)) }
func (v RatioVal) Uint() UintVal          { return UintVal(uint(v.Int())) }
func (v RatioVal) Int() IntVal            { var num, _ = v.Rat().Float64(); return IntVal(int(num)) }
func (v RatioVal) Float() FltVal          { var flt, _ = v.Rat().Float64(); return FltVal(flt) }
func (v RatioVal) Rat() *big.Rat          { return (*big.Rat)(&v) }
func (v RatioVal) Imag() ImagVal          { return ImagVal(complex(v.Float(), 1.0)) }
func (v RatioVal) Numerator() IntVal      { return IntVal(int(v.Rat().Num().Int64())) }
func (v RatioVal) Denominator() IntVal    { return IntVal(int(v.Rat().Denom().Int64())) }
func (v RatioVal) Both() (Native, Native) { return IntVal(v.Numerator()), IntVal(v.Denominator()) }
func (v RatioVal) Left() Native           { return IntVal(v.Numerator()) }
func (v RatioVal) Right() Native          { return IntVal(v.Denominator()) }
func (v RatioVal) BothInt() (IntVal, IntVal) {
	return IntVal(int(v.Rat().Num().Int64())), IntVal(int(v.Rat().Denom().Int64()))
}
func (v RatioVal) Bool() BoolVal {
	if v.Int() > 0 {
		return BoolVal(true)
	}
	return BoolVal(false)
}
func (v RatioVal) Truth() Native {
	if v.Int() > 0 {
		return BoolVal(true)
	}
	if v.Int() < 0 {
		return BoolVal(true)
	}
	return NilVal{}
}

// operators
func (v *RatioVal) Negate() *RatioVal { return (*RatioVal)((*big.Rat)(v).Neg((*big.Rat)(v))) }
func (v *RatioVal) Invert() *RatioVal { return (*RatioVal)((*big.Rat)(v).Inv((*big.Rat)(v))) }
func (v *RatioVal) Add(arg *RatioVal) *RatioVal {
	return (*RatioVal)((*big.Rat)(v).Add((*big.Rat)(v), (*big.Rat)(arg)))
}
func (v *RatioVal) Substract(arg *RatioVal) *RatioVal {
	return (*RatioVal)((*big.Rat)(v).Sub((*big.Rat)(v), (*big.Rat)(arg)))
}
func (v *RatioVal) Multipy(arg *RatioVal) *RatioVal {
	return (*RatioVal)((*big.Rat)(v).Mul((*big.Rat)(v), (*big.Rat)(arg)))
}
func (v *RatioVal) Quotient(arg *RatioVal) *RatioVal {
	return (*RatioVal)((*big.Rat)(v).Quo((*big.Rat)(v), (*big.Rat)(arg)))
}

// comparators
func (v *RatioVal) Cmp(arg *RatioVal) int      { return v.Cmp(arg) }
func (v *RatioVal) Lesser(arg *RatioVal) bool  { return v.Cmp(arg) < 0 }
func (v *RatioVal) Greater(arg *RatioVal) bool { return v.Cmp(arg) > 0 }
func (v *RatioVal) Equal(arg *RatioVal) bool   { return v.Cmp(arg) == 0 }

// IMAGINARY VALUE
func (v Imag64Val) Idx() int                    { return int(v.Int()) }
func (v Imag64Val) GoInt() int                  { return int(v.Int()) }
func (v Imag64Val) GoUint() uint                { return uint(v.Uint()) }
func (v Imag64Val) GoFlt() float64              { return float64(v.Float()) }
func (v Imag64Val) GoImag() complex128          { return complex128(v.Imag()) }
func (v Imag64Val) GoRat() *big.Rat             { return (*big.Rat)(v.Ratio()) }
func (v Imag64Val) GoBigInt() *big.Int          { return big.NewInt(int64(v.GoInt())) }
func (v Imag64Val) GoBigFlt() *big.Float        { return big.NewFloat(v.GoFlt()) }
func (v Imag64Val) BigInt() *BigIntVal          { return (*BigIntVal)(v.GoBigInt()) }
func (v Imag64Val) BigFlt() *BigFltVal          { return (*BigFltVal)(v.GoBigFlt()) }
func (v Imag64Val) Unit() Native                { return Imag64Val(complex(0, 0)) }
func (v Imag64Val) Uint() UintVal               { return UintVal(uint(real(v))) }
func (v Imag64Val) Int() IntVal                 { return IntVal(int(real(v))) }
func (v Imag64Val) IntVal() IntVal              { return IntVal(real(v)) }
func (v Imag64Val) Float() FltVal               { return FltVal(float64(real(v))) }
func (v Imag64Val) Ratio() *big.Rat             { return big.NewRat(int64(real(v)), int64(imag(v))) }
func (v Imag64Val) Imag() ImagVal               { return ImagVal(complex64(v)) }
func (v Imag64Val) Imag64() Imag64Val           { return v }
func (v Imag64Val) Imaginary() FltVal           { return FltVal(imag(v)) }
func (v Imag64Val) Real() FltVal                { return FltVal(real(v)) }
func (v Imag64Val) BothFloat() (FltVal, FltVal) { return FltVal(real(v)), FltVal(imag(v)) }
func (v Imag64Val) Both() (Native, Native)      { return FltVal(real(v)), FltVal(imag(v)) }
func (v Imag64Val) Left() Native                { return FltVal(v.Real()) }
func (v Imag64Val) Right() Native               { return FltVal(v.Imaginary()) }
func (v Imag64Val) Bool() BoolVal {
	if real(v) > 0 {
		return BoolVal(true)
	}
	return BoolVal(false)
}
func (v Imag64Val) Equal(arg Imag64Val) bool { return v == arg }

// IMAGINARY VALUE
func (v ImagVal) Idx() int                    { return int(v.Int()) }
func (v ImagVal) GoInt() int                  { return int(v.Int()) }
func (v ImagVal) GoUint() uint                { return uint(v.Uint()) }
func (v ImagVal) GoFlt() float64              { return float64(v.Float()) }
func (v ImagVal) GoImag() complex128          { return complex128(v.Imag()) }
func (v ImagVal) GoRat() *big.Rat             { return (*big.Rat)(v.Ratio()) }
func (v ImagVal) GoBigInt() *big.Int          { return big.NewInt(int64(v.GoInt())) }
func (v ImagVal) GoBigFlt() *big.Float        { return big.NewFloat(v.GoFlt()) }
func (v ImagVal) BigInt() *BigIntVal          { return (*BigIntVal)(v.GoBigInt()) }
func (v ImagVal) BigFlt() *BigFltVal          { return (*BigFltVal)(v.GoBigFlt()) }
func (v ImagVal) Unit() Native                { return ImagVal(complex(0, 0)) }
func (v ImagVal) Uint() UintVal               { return UintVal(uint(real(v))) }
func (v ImagVal) Int() IntVal                 { return IntVal(int(real(v))) }
func (v ImagVal) IntVal() IntVal              { return IntVal(real(v)) }
func (v ImagVal) Float() FltVal               { return FltVal(float64(real(v))) }
func (v ImagVal) Ratio() *big.Rat             { return big.NewRat(int64(real(v)), int64(imag(v))) }
func (v ImagVal) Imag() ImagVal               { return v }
func (v ImagVal) Imag64() Imag64Val           { return Imag64Val(complex64(v)) }
func (v ImagVal) Imaginary() FltVal           { return FltVal(imag(v)) }
func (v ImagVal) Real() FltVal                { return FltVal(real(v)) }
func (v ImagVal) BothFloat() (FltVal, FltVal) { return FltVal(real(v)), FltVal(imag(v)) }
func (v ImagVal) Both() (Native, Native)      { return FltVal(real(v)), FltVal(imag(v)) }
func (v ImagVal) Left() Native                { return FltVal(v.Real()) }
func (v ImagVal) Right() Native               { return FltVal(v.Imaginary()) }
func (v ImagVal) Bool() BoolVal {
	if real(v) > 0 {
		return BoolVal(true)
	}
	return BoolVal(false)
}
func (v ImagVal) Equal(arg ImagVal) bool { return v == arg }

/// BIG INT VALUE
func (v *BigIntVal) Int64() int64      { return v.Int64() }
func (v BigIntVal) Idx() int           { return int(v.Int()) }
func (v BigIntVal) GoInt() int         { return int(v.Int()) }
func (v BigIntVal) GoUint() uint       { return uint(v.Uint()) }
func (v BigIntVal) GoFlt() float64     { return float64(v.Int64()) }
func (v BigIntVal) GoImag() complex128 { return complex128(v.Imag()) }
func (v BigIntVal) GoRat() *big.Rat {
	return (*big.Rat)(big.NewRat(int64(v.GoInt()), 1))
}
func (v BigIntVal) GoBigInt() *big.Int   { return (*big.Int)(&v) }
func (v BigIntVal) GoBigFlt() *big.Float { return big.NewFloat(v.GoFlt()) }
func (v BigIntVal) BigInt() *BigIntVal   { return (*BigIntVal)(v.GoBigInt()) }
func (v BigIntVal) BigFlt() *BigFltVal   { return (*BigFltVal)(v.GoBigFlt()) }
func (v BigIntVal) Bool() BoolVal        { return IntVal(v.Int()).Bool() }
func (v BigIntVal) Int() IntVal          { return IntVal(int(v.Int64())) }
func (v BigIntVal) Uint() UintVal        { return UintVal(uint(v.GoBigInt().Uint64())) }
func (v BigIntVal) Float() FltVal        { return FltVal(float64(v.GoFlt())) }
func (v BigIntVal) atio() *RatioVal      { return IntVal(v.Int()).Ratio() }
func (v BigIntVal) Imag() ImagVal        { return IntVal(v.Int()).Imag() }

// operators
func (v BigIntVal) Add(arg BigIntVal) BigIntVal       { return v.Add(arg) }
func (v BigIntVal) Substract(arg BigIntVal) BigIntVal { return v.Substract(arg) }
func (v BigIntVal) Multipy(arg BigIntVal) BigIntVal   { return v.Multipy(arg) }
func (v BigIntVal) Quotient(arg BigIntVal) BigIntVal  { return v.Quotient(arg) }
func (v BigIntVal) Power(arg BigIntVal) BigIntVal     { return v.Power(arg) }
func (v BigIntVal) And(arg BigIntVal) BigIntVal       { return v.And(arg) }
func (v BigIntVal) Xor(arg BigIntVal) BigIntVal       { return v.Xor(arg) }
func (v BigIntVal) Or(arg BigIntVal) BigIntVal        { return v.Or(arg) }
func (v BigIntVal) AndNot(arg BigIntVal) BigIntVal    { return v.AndNot(arg) }
func (v BigIntVal) Not() BigIntVal                    { return v.Not() }

// comparators
func (v BigIntVal) Cmp(arg BigIntVal) int      { return v.GoBigInt().Cmp(arg.GoBigInt()) }
func (v BigIntVal) Equal(arg BigIntVal) bool   { return v.Cmp(arg) == 0 }
func (v BigIntVal) Lesser(arg BigIntVal) bool  { return v.Cmp(arg) < 0 }
func (v BigIntVal) Greater(arg BigIntVal) bool { return v.Cmp(arg) > 0 }

/// BIG FLOAT VALUE
func (v *BigFltVal) Float64() int64      { return v.Float64() }
func (v BigFltVal) Idx() int             { return int(v.Int()) }
func (v BigFltVal) GoInt() int           { return int(v.Int()) }
func (v BigFltVal) GoUint() uint         { return uint(v.Uint()) }
func (v BigFltVal) GoFlt() float64       { return float64(v.BigFlt().Float64()) }
func (v BigFltVal) GoImag() complex128   { return complex128(v.Imag()) }
func (v BigFltVal) GoRat() *big.Rat      { return (*big.Rat)(v.Ratio()) }
func (v BigFltVal) GoBigInt() *big.Int   { return big.NewInt(int64(v.Int())) }
func (v BigFltVal) GoBigFlt() *big.Float { return (*big.Float)(&v) }
func (v BigFltVal) BigInt() *BigIntVal   { return (*BigIntVal)(v.GoBigInt()) }
func (v BigFltVal) BigFlt() *BigFltVal   { return (*BigFltVal)(v.GoBigFlt()) }
func (v BigFltVal) Bool() BoolVal        { return IntVal(v.Int()).Bool() }
func (v BigFltVal) Int() IntVal          { return IntVal(int(v.GoBigInt().Int64())) }
func (v BigFltVal) Uint() UintVal        { return UintVal(uint(v.GoBigInt().Uint64())) }
func (v BigFltVal) Float() FltVal        { return FltVal(float64(v.Float64())) }
func (v BigFltVal) Ratio() *RatioVal     { return IntVal(v.Int()).Ratio() }
func (v BigFltVal) Imag() ImagVal        { return IntVal(v.Int()).Imag() }

// operators
func (v BigFltVal) Add(arg BigFltVal) BigFltVal       { return v.Add(arg) }
func (v BigFltVal) Substract(arg BigFltVal) BigFltVal { return v.Substract(arg) }
func (v BigFltVal) Multipy(arg BigFltVal) BigFltVal   { return v.Multipy(arg) }
func (v BigFltVal) Quotient(arg BigFltVal) BigFltVal  { return v.Quotient(arg) }

// comparators
func (v BigFltVal) Cmp(arg BigFltVal) int      { return v.GoBigFlt().Cmp(arg.GoBigFlt()) }
func (v BigFltVal) Equal(arg BigFltVal) bool   { return v.Cmp(arg) == 0 }
func (v BigFltVal) Lesser(arg BigFltVal) bool  { return v.Cmp(arg) < 0 }
func (v BigFltVal) Greater(arg BigFltVal) bool { return v.Cmp(arg) > 0 }

/// TIME VALUE
func (v TimeVal) Idx() int             { return int(v.Int()) }
func (v TimeVal) GoInt() int           { return int(v.Int()) }
func (v TimeVal) GoUint() uint         { return uint(v.Uint()) }
func (v TimeVal) GoFlt() float64       { return float64(v.Float()) }
func (v TimeVal) GoImag() complex128   { return complex128(v.Imag()) }
func (v TimeVal) GoRat() *big.Rat      { return (*big.Rat)(v.Ratio()) }
func (v TimeVal) GoBigInt() *big.Int   { return big.NewInt(int64(v.GoInt())) }
func (v TimeVal) GoBigFlt() *big.Float { return big.NewFloat(v.GoFlt()) }
func (v TimeVal) BigInt() *BigIntVal   { return (*BigIntVal)(v.GoBigInt()) }
func (v TimeVal) BigFlt() *BigFltVal   { return (*BigFltVal)(v.GoBigFlt()) }
func (v TimeVal) Time() time.Time      { return time.Time(v) }
func (v TimeVal) Uint() UintVal        { return UintVal(uint(time.Time(v).Unix())) }
func (v TimeVal) UintVal() UintVal     { return UintVal(uint(time.Time(v).Unix())) }
func (v TimeVal) Int() IntVal          { return IntVal(int(time.Time(v).Unix())) }
func (v TimeVal) IntVal() IntVal       { return IntVal(time.Time(v).Unix()) }
func (v TimeVal) Bool() BoolVal        { return IntVal(v.Int()).Bool() }
func (v TimeVal) Ratio() *RatioVal     { return IntVal(v.Int()).Ratio() }
func (v TimeVal) Float() FltVal        { return IntVal(v.Int()).Float() }
func (v TimeVal) Imag() ImagVal        { return IntVal(v.Int()).Imag() }
func (v TimeVal) ANSIC() StrVal        { return StrVal(time.ANSIC) }

// operators
func (v TimeVal) Add(arg DuraVal) TimeVal       { return v.Add(arg) }
func (v TimeVal) Substract(arg DuraVal) TimeVal { return v.Substract(arg) }

// comparators
func (v TimeVal) Lesser(arg TimeVal) bool  { return time.Time(v).Before(time.Time(arg)) }
func (v TimeVal) Greater(arg TimeVal) bool { return time.Time(v).After(time.Time(arg)) }
func (v TimeVal) Equal(arg TimeVal) bool   { return time.Time(v).Equal(time.Time(arg)) }

/// DURATION VALUE
func (v DuraVal) Idx() int                { return int(v.Int()) }
func (v DuraVal) GoInt() int              { return int(v.Int()) }
func (v DuraVal) GoUint() uint            { return uint(v.Uint()) }
func (v DuraVal) GoFlt() float64          { return float64(v.Float()) }
func (v DuraVal) GoImag() complex128      { return complex128(v.Imag()) }
func (v DuraVal) GoRat() *big.Rat         { return (*big.Rat)(v.Ratio()) }
func (v DuraVal) GoBigInt() *big.Int      { return big.NewInt(int64(v.GoInt())) }
func (v DuraVal) GoBigFlt() *big.Float    { return big.NewFloat(v.GoFlt()) }
func (v DuraVal) BigInt() *BigIntVal      { return (*BigIntVal)(v.GoBigInt()) }
func (v DuraVal) BigFlt() *BigFltVal      { return (*BigFltVal)(v.GoBigFlt()) }
func (v DuraVal) Duration() time.Duration { return time.Duration(v) }
func (v DuraVal) Uint() UintVal           { return UintVal(uint(v)) }
func (v DuraVal) UintVal() UintVal        { return UintVal(v.Uint()) }
func (v DuraVal) Int() IntVal             { return IntVal(int(v)) }
func (v DuraVal) IntVal() IntVal          { return IntVal(v.Int()) }
func (v DuraVal) Bool() BoolVal           { return IntVal(v.Int()).Bool() }
func (v DuraVal) Ratio() *RatioVal        { return IntVal(v.Int()).Ratio() }
func (v DuraVal) Float() FltVal           { return IntVal(v.Int()).Float() }
func (v DuraVal) Imag() ImagVal           { return IntVal(v.Int()).Imag() }

// comparators
func (v DuraVal) Equal(arg DuraVal) bool   { return v.Int() == arg.Int() }
func (v DuraVal) Lesser(arg DuraVal) bool  { return v.Int() < arg.Int() }
func (v DuraVal) Greater(arg DuraVal) bool { return v.Int() > arg.Int() }

/// BYTE VALUE
func (v ByteVal) Bool() bool {
	if v > ByteVal(0) {
		return true
	}
	return false
}
func (v ByteVal) Idx() int                    { return int(v.Int()) }
func (v ByteVal) String() string              { return string(v.Bytes()) }
func (v ByteVal) GoByte() byte                { return byte(v) }
func (v ByteVal) GoInt() int                  { return int(v.Int()) }
func (v ByteVal) GoUint() uint                { return uint(v.Uint()) }
func (v ByteVal) GoFlt() float64              { return float64(v.Float()) }
func (v ByteVal) GoImag() complex128          { return complex128(v.Imag()) }
func (v ByteVal) GoRat() *big.Rat             { return (*big.Rat)(v.Ratio()) }
func (v ByteVal) GoBigInt() *big.Int          { return big.NewInt(int64(v.GoInt())) }
func (v ByteVal) GoBigFlt() *big.Float        { return big.NewFloat(v.GoFlt()) }
func (v ByteVal) BigInt() *BigIntVal          { return (*BigIntVal)(v.GoBigInt()) }
func (v ByteVal) BigFlt() *BigFltVal          { return (*BigFltVal)(v.GoBigFlt()) }
func (v ByteVal) Bytes() BytesVal             { return BytesVal([]byte{v.GoByte()}) }
func (v ByteVal) Unit() Native                { return ByteVal(byte(0)) }
func (v ByteVal) Uint() UintVal               { return UintVal(uint(v)) }
func (v ByteVal) Int() IntVal                 { return IntVal(int(v)) }
func (v ByteVal) Ratio() *RatioVal            { return IntVal(int(v)).Ratio() }
func (v ByteVal) Float() FltVal               { return FltVal(float64(v)) }
func (v ByteVal) Imag() ImagVal               { return IntVal(int(v)).Imag() }
func (v ByteVal) Byte() ByteVal               { return ByteVal(byte(v)) }
func (v ByteVal) Rune() RuneVal               { return RuneVal(rune(v.Byte())) }
func (v ByteVal) Len() IntVal                 { return IntVal(bits.Len8(uint8(v.Uint()))) }
func (v ByteVal) Concat(arg ByteVal) BytesVal { return BytesVal([]byte{byte(v), byte(arg)}) }

// operators
func (v ByteVal) And(arg ByteVal) ByteVal    { return v & arg }
func (v ByteVal) Xor(arg ByteVal) ByteVal    { return v ^ arg }
func (v ByteVal) Or(arg ByteVal) ByteVal     { return v | arg }
func (v ByteVal) AndNot(arg ByteVal) ByteVal { return v &^ arg }
func (v ByteVal) Not() ByteVal               { return ^v }

// comparators
func (v ByteVal) Equal(arg ByteVal) bool   { return v == arg }
func (v ByteVal) Lesser(arg ByteVal) bool  { return v < arg }
func (v ByteVal) Greater(arg ByteVal) bool { return v > arg }

/// BYTE SLICE VALUE
// conversions
func (v BytesVal) String() string            { return string(v) }
func (v BytesVal) GoBytes() []byte           { return []byte(v) }
func (v BytesVal) GoRunes() []rune           { return []rune(v.String()) }
func (v BytesVal) ByteBuffer() *bytes.Buffer { return bytes.NewBuffer(v) }
func (v BytesVal) ByteReader() io.ByteReader { return bytes.NewReader(v) }
func (v BytesVal) StrVal() StrVal            { return StrVal(v.String()) }
func (v BytesVal) Unit() BytesVal            { return BytesVal([]byte{byte(0)}) }
func (v BytesVal) Bytes() ByteVec            { return ByteVec(v) }
func (v BytesVal) RuneVec() RuneVec          { return RuneVec(v.GoRunes()) }
func (v BytesVal) Len() IntVal               { return IntVal(len(v.Bytes())) }
func (v BytesVal) Concat(args ...ByteVal) BytesVal {
	for _, arg := range args {
		v = append(v, byte(arg))
	}
	return v
}
func (v BytesVal) UintNative() Native {
	u, err := binary.ReadUvarint(v.ByteReader())
	if err != nil {
		return NewError(err)
	}
	return UintVal(u)
}
func (v BytesVal) IntNative() Native {
	i, err := binary.ReadVarint(v.ByteReader())
	if err != nil {
		return NewError(err)
	}
	return IntVal(i)
}
func (v BytesVal) Bool() BoolVal {
	for _, b := range v {
		if b > byte(0) {
			return BoolVal(true)
		}
	}
	return BoolVal(false)
}

/// STRING VALUE
// conversions
func (v StrVal) String() string                  { return string(v) }
func (v StrVal) StringBuffer() *strings.Reader   { return strings.NewReader(v.String()) }
func (v StrVal) Unit() Native                    { return StrVal(" ") }
func (v StrVal) Runes() RuneVec                  { return RuneVec([]rune(string(v))) }
func (v StrVal) Len() IntVal                     { return IntVal(int(len(string(v)))) }
func (v StrVal) Bytes() BytesVal                 { return []byte(string(v)) }
func (v StrVal) DurationNative() Native          { return v.ReadDuraVal() }
func (v StrVal) TimeNative(layout string) Native { return v.ReadTimeVal(layout) }
func (v StrVal) NumberNative() Native {
	if _, err := v.ReadBool(); err == nil {
		return v.ReadBoolVal()
	}
	if _, err := v.ReadUint(); err == nil {
		return v.ReadUintVal()
	}
	if _, err := v.ReadInt(); err == nil {
		return v.ReadIntVal()
	}
	if _, err := v.ReadFloat(); err == nil {
		return v.ReadFloatVal()
	}
	return NewNil()
}

// operators
func (v StrVal) Concat(arg StrVal) StrVal { return StrVal(string(v) + string(arg)) }
func (v StrVal) Add(arg StrVal) StrVal    { return v.Concat(arg) }
func (v StrVal) Split(arg StrVal) (val []StrVal) {
	var slice = strings.Split(string(v), string(arg))
	val = make([]StrVal, 0, len(slice))
	for _, str := range slice {
		val = append(val, StrVal(str))
	}
	return val
}
func (v StrVec) Join(arg StrVal) (val []StrVal) {
	val = make([]StrVal, 0, len(v))
	for _, str := range v {
		val = append(val, StrVal(str))
	}
	return val
}

// comparators
func (v StrVal) EqualFold(arg StrVal) bool { return v.EqualFold(arg) }
func (v StrVal) Compare(arg StrVal) int    { return strings.Compare(string(v), string(arg)) }
func (v StrVal) Equal(arg StrVal) bool     { return v.Compare(arg) == 0 }
func (v StrVal) Lesser(arg StrVal) bool    { return v.Compare(arg) < 0 }
func (v StrVal) Greater(arg StrVal) bool   { return v.Compare(arg) > 0 }
func (v StrVal) HasPrefix(arg StrVal) bool { return strings.HasPrefix(string(v), string(arg)) }
func (v StrVal) HasSuffix(arg StrVal) bool { return strings.HasSuffix(string(v), string(arg)) }
func (v StrVal) Contains(arg StrVal) bool  { return strings.Contains(string(v), string(arg)) }
func (v StrVal) ContainsAny(arg StrVal) bool {
	return strings.ContainsAny(string(v), string(arg))
}
func (v StrVal) ContainsRune(arg RuneVal) bool {
	return strings.ContainsRune(string(v), rune(arg))
}
