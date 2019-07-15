package data

import (
	"bytes"
	"encoding/binary"
	"io"
	"math/big"
	"math/bits"
	"math/cmplx"
	"strings"
	"time"
)

// converts a numeral to an instance of another numeral type
func CastNumeral(num Numeral, typ TyNat) Native {
	if typ.Match(Numbers) {
		switch typ {
		case BigInt:
			var bi = BigIntVal(*big.NewInt(int64(num.GoInt())))
			return &bi
		case BigFlt:
			var bf = BigFltVal(*big.NewFloat(num.GoFlt()))
			return &bf
		case Ratio:
			var rt = RatioVal(*num.GoRat())
			return &rt
		case Imag:
			return ImagVal(num.GoImag())
		case Imag64:
			var im = num.GoImag()
			if cmplx.Abs(im) <= -3.4028234663852886e+38 &&
				cmplx.Abs(im) >= 3.4028234663852886e+38 {
				return Imag64Val(complex64(im))
			}
		case Int:
			if num.GoInt() >= -9223372036854775808 &&
				num.GoInt() <= 9223372036854775807 {
				return IntVal(num.GoInt())
			}
		case Int8:
			if num.GoInt() >= -128 &&
				num.GoInt() <= 127 {
				return Int8Val(num.GoInt())
			}
		case Int16:
			if num.GoInt() >= -32768 &&
				num.GoInt() <= -32767 {
				return Int16Val(num.GoInt())
			}
		case Int32:
			if num.GoInt() >= -2147483648 &&
				num.GoInt() <= -2147483647 {
				return Int16Val(num.GoInt())
			}
		case Uint:
			if num.GoUint() >= 0 &&
				num.GoUint() <= 18446744073709551615 {
				return UintVal(num.GoUint())
			}
		case Uint8:
			if num.GoUint() >= 0 &&
				num.GoInt() <= 255 {
				return Uint8Val(num.GoUint())
			}
		case Uint16:
			if num.GoUint() >= 0 &&
				num.GoInt() <= 65535 {
				return Uint16Val(num.GoUint())
			}
		case Uint32:
			if num.GoUint() >= 0 &&
				num.GoInt() <= 4294967295 {
				return Uint32Val(num.GoUint())
			}
		case Float:
			if num.GoFlt() >= -1.7976931348623157e+308 &&
				num.GoFlt() <= 1.7976931348623157e+308 {
				return FltVal(num.GoFlt())
			}
		case Flt32:
			if num.GoFlt() >= -3.4028234663852886e+38 &&
				num.GoFlt() <= 3.4028234663852886e+38 {
				return Flt32Val(num.GoFlt())
			}
		case String:
			return StrVal(num.String())
		}
	}
	return NilVal{}
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
func (v BoolVal) Float() FltVal { return v.IntVal().Float() }
func (v BoolVal) Imag() ImagVal { return v.IntVal().Imag() }

// NATURAL VALUE
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
func (v UintVal) Uint() UintVal        { return UintVal(uint(v)) }
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

// INTEGER VALUE
func (v IntVal) GoInt() int           { return int(v) }
func (v IntVal) GoFlt() float64       { return float64(v) }
func (v IntVal) GoUint() uint         { return uint(v) }
func (v IntVal) GoImag() complex128   { return complex128(v.Imag()) }
func (v IntVal) GoRat() *big.Rat      { return (*big.Rat)(v.Ratio()) }
func (v IntVal) GoBigInt() *big.Int   { return big.NewInt(int64(v)) }
func (v IntVal) GoBigFlt() *big.Float { return big.NewFloat(float64(v)) }
func (v IntVal) BigInt() *BigIntVal   { return (*BigIntVal)(v.GoBigInt()) }
func (v IntVal) BigFlt() *BigFltVal   { return (*BigFltVal)(v.GoBigFlt()) }
func (v IntVal) Unit() Native         { return IntVal(1) }
func (v IntVal) IntVal() IntVal       { return v }
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
func (v ImagVal) Imag() ImagVal               { return ImagVal(complex128(v)) }
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

/// BIG INT VALUE
func (v *BigIntVal) Int64() int64        { return v.Int64() }
func (v BigIntVal) Idx() int             { return int(v.Int()) }
func (v BigIntVal) GoInt() int           { return int(v.Int()) }
func (v BigIntVal) GoUint() uint         { return uint(v.Uint()) }
func (v BigIntVal) GoFlt() float64       { return float64(v.Int64()) }
func (v BigIntVal) GoImag() complex128   { return complex128(v.Imag()) }
func (v BigIntVal) GoRat() *big.Rat      { return (*big.Rat)(v.Ratio()) }
func (v BigIntVal) GoBigInt() *big.Int   { return (*big.Int)(&v) }
func (v BigIntVal) GoBigFlt() *big.Float { return big.NewFloat(v.GoFlt()) }
func (v BigIntVal) BigInt() *BigIntVal   { return (*BigIntVal)(v.GoBigInt()) }
func (v BigIntVal) BigFlt() *BigFltVal   { return (*BigFltVal)(v.GoBigFlt()) }
func (v BigIntVal) Bool() BoolVal        { return IntVal(v.Int()).Bool() }
func (v BigIntVal) Int() IntVal          { return IntVal(int(v.Int64())) }
func (v BigIntVal) Uint() UintVal        { return UintVal(uint(v.GoBigInt().Uint64())) }
func (v BigIntVal) Float() FltVal        { return FltVal(float64(v.GoFlt())) }
func (v BigIntVal) Ratio() *RatioVal     { return IntVal(v.Int()).Ratio() }
func (v BigIntVal) Imag() ImagVal        { return IntVal(v.Int()).Imag() }

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

/// BYTE VALUE
func (v ByteVal) Bool() bool {
	if v > ByteVal(0) {
		return true
	}
	return false
}
func (v ByteVal) Idx() int             { return int(v.Int()) }
func (v ByteVal) String() string       { return string(v.Bytes()) }
func (v ByteVal) GoByte() byte         { return byte(v) }
func (v ByteVal) GoInt() int           { return int(v.Int()) }
func (v ByteVal) GoUint() uint         { return uint(v.Uint()) }
func (v ByteVal) GoFlt() float64       { return float64(v.Float()) }
func (v ByteVal) GoImag() complex128   { return complex128(v.Imag()) }
func (v ByteVal) GoRat() *big.Rat      { return (*big.Rat)(v.Ratio()) }
func (v ByteVal) GoBigInt() *big.Int   { return big.NewInt(int64(v.GoInt())) }
func (v ByteVal) GoBigFlt() *big.Float { return big.NewFloat(v.GoFlt()) }
func (v ByteVal) BigInt() *BigIntVal   { return (*BigIntVal)(v.GoBigInt()) }
func (v ByteVal) BigFlt() *BigFltVal   { return (*BigFltVal)(v.GoBigFlt()) }
func (v ByteVal) Bytes() BytesVal      { return BytesVal([]byte{v.GoByte()}) }
func (v ByteVal) Unit() Native         { return ByteVal(byte(0)) }
func (v ByteVal) Uint() UintVal        { return UintVal(uint(v)) }
func (v ByteVal) Int() IntVal          { return IntVal(int(v)) }
func (v ByteVal) Ratio() *RatioVal     { return IntVal(int(v)).Ratio() }
func (v ByteVal) Float() FltVal        { return FltVal(float64(v)) }
func (v ByteVal) Imag() ImagVal        { return IntVal(int(v)).Imag() }
func (v ByteVal) Byte() ByteVal        { return ByteVal(byte(v)) }
func (v ByteVal) Rune() RuneVal        { return RuneVal(rune(v.Byte())) }
func (v ByteVal) Len() IntVal          { return IntVal(bits.Len8(uint8(v.Uint()))) }

/// BYTE SLICE VALUE
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
