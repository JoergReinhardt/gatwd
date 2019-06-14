package data

import (
	"math/big"
	"math/bits"
	"math/cmplx"
	"strconv"
	"time"
)

// converts a numeral to an instance of another numeral type
func castNumberAs(num Numeral, typ TyNat) Native {
	if typ.Match(Numbers) {
		switch typ {
		case BigInt:
			var bi = BigIntVal(*big.NewInt(int64(num.Int())))
			return &bi
		case BigFlt:
			var bf = BigFltVal(*big.NewFloat(num.Float()))
			return &bf
		case Ratio:
			var rt = RatioVal(*num.Rat())
			return &rt
		case Imag:
			return ImagVal(num.Imag())
		case Imag64:
			var im = num.Imag()
			if cmplx.Abs(im) <= -3.4028234663852886e+38 &&
				cmplx.Abs(im) >= 3.4028234663852886e+38 {
				return Imag64Val(complex64(im))
			}
		case Int:
			if num.Int() >= -9223372036854775808 &&
				num.Int() <= 9223372036854775807 {
				return IntVal(num.Int())
			}
		case Int8:
			if num.Int() >= -128 &&
				num.Int() <= 127 {
				return Int8Val(num.Int())
			}
		case Int16:
			if num.Int() >= -32768 &&
				num.Int() <= -32767 {
				return Int16Val(num.Int())
			}
		case Int32:
			if num.Int() >= -2147483648 &&
				num.Int() <= -2147483647 {
				return Int16Val(num.Int())
			}
		case Uint:
			if num.Uint() >= 0 &&
				num.Uint() <= 18446744073709551615 {
				return UintVal(num.Uint())
			}
		case Uint8:
			if num.Uint() >= 0 &&
				num.Int() <= 255 {
				return Uint8Val(num.Uint())
			}
		case Uint16:
			if num.Uint() >= 0 &&
				num.Int() <= 65535 {
				return Uint16Val(num.Uint())
			}
		case Uint32:
			if num.Uint() >= 0 &&
				num.Int() <= 4294967295 {
				return Uint32Val(num.Uint())
			}
		case Float:
			if num.Float() >= -1.7976931348623157e+308 &&
				num.Float() <= 1.7976931348623157e+308 {
				return FltVal(num.Float())
			}
		case Flt32:
			if num.Float() >= -3.4028234663852886e+38 &&
				num.Float() <= 3.4028234663852886e+38 {
				return Flt32Val(num.Float())
			}
		case String:
			return StrVal(num.String())
		}
	}
	return NilVal{}
}

// either parse number, or duration from string, or return length of string
func (v StrVal) Number() Native {
	if u, err := v.Bool(); err != nil {
		return BoolVal(u)
	}
	if u, err := v.Uint(); err != nil {
		return UintVal(u)
	}
	if i, err := v.Int(); err != nil {
		return IntVal(i)
	}
	if f, err := v.Float(); err != nil {
		return FltVal(f)
	}
	if d, err := v.Duration(); err != nil {
		return DuraVal(d)
	}
	return IntVal(v.Len())
}

// BOOL VALUE
func (v BoolVal) Unit() Native { return BoolVal(true) }
func (b BoolVal) Bool() bool   { return bool(b) }
func (v BoolVal) Uint() uint {
	if v {
		return 1
	}
	return 0
}
func (v BoolVal) Int() int {
	if v {
		return 1
	}
	return -1
}
func (v BoolVal) Integer() IntVal {
	if v {
		return IntVal(1)
	}
	return IntVal(-1)
}
func (v BoolVal) Float() float64   { return v.Integer().Float() }
func (v BoolVal) Rat() *big.Rat    { return v.Integer().Rat() }
func (v BoolVal) Imag() complex128 { return v.Integer().Imag() }

// NATURAL VALUE
func (v UintVal) Unit() Native { return UintVal(1) }
func (v UintVal) Truth() Native {
	if v > 0 {
		return BoolVal(true)
	}
	return BoolVal(false)
}
func (v UintVal) Uint() uint       { return uint(v) }
func (v UintVal) Int() int         { return int(v) }
func (v UintVal) Integer() IntVal  { return IntVal(int(v)) }
func (v UintVal) Bool() bool       { return bool(v.Truth().(BoolVal)) }
func (v UintVal) Float() float64   { return float64(v) }
func (v UintVal) Rat() *big.Rat    { return big.NewRat(int64(v), 1) }
func (v UintVal) Imag() complex128 { return v.Integer().Imag() }

// INTEGER VALUE
func (v IntVal) Unit() Native { return IntVal(1) }
func (v IntVal) Bool() bool {
	if v > 0 {
		return true
	}
	return false
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
func (v IntVal) Natural() UintVal {
	if v < 0 {
		return UintVal(v * -1)
	}
	return UintVal(v)
}
func (v IntVal) Uint() uint       { return uint(v.Natural()) }
func (v IntVal) Int() int         { return int(v) } // implements Idx Attribut
func (v IntVal) Integer() IntVal  { return v }      // implements Idx Attribut
func (v IntVal) Float() float64   { return float64(v.Int()) }
func (v IntVal) Rat() *big.Rat    { return big.NewRat(1, int64(v)) } // implements Idx Attribut
func (v IntVal) Imag() complex128 { return complex(v.Float(), 1.0) } // implements Idx Attribut
func (v IntVal) Idx() int         { return int(v) }                  // implements Idx Attribut

// REAL VALUE
func (v FltVal) Unit() Native { return FltVal(1.0) }
func (v FltVal) Bool() bool {
	if v > 0.0 {
		return true
	}
	return false
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
func (v FltVal) Uint() uint       { return uint(v) }
func (v FltVal) Int() int         { return int(v) }
func (v FltVal) Integer() IntVal  { return IntVal(int(v)) }
func (v FltVal) Float() float64   { return float64(v) }
func (v FltVal) Rat() *big.Rat    { return big.NewRat(int64(1), int64(1)).SetFloat64(v.Float()) }
func (v FltVal) Imag() complex128 { return complex(v, 1.0) }

// RATIONAL VALUE
func (v RatioVal) Unit() Native { return RatioVal(*big.NewRat(1, 1)) }
func (v RatioVal) Bool() bool {
	if v.Int() > 0 {
		return true
	}
	return false
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
func (v RatioVal) Uint() uint       { return uint(v.Int()) }
func (v RatioVal) Int() int         { var num, _ = v.Rat().Float64(); return int(num) }
func (v RatioVal) Integer() IntVal  { return IntVal(v.Int()) }
func (v RatioVal) Float() float64   { var flt, _ = v.Rat().Float64(); return flt }
func (v RatioVal) Rat() *big.Rat    { return (*big.Rat)(&v) }
func (v RatioVal) Imag() complex128 { return complex(v.Float(), 1.0) }
func (v RatioVal) Numerator() int   { return int(v.Rat().Num().Int64()) }
func (v RatioVal) Denominator() int { return int(v.Rat().Denom().Int64()) }
func (v RatioVal) BothInt() (int, int) {
	return int(v.Rat().Num().Int64()), int(v.Rat().Denom().Int64())
}
func (v RatioVal) Both() (Native, Native) { return IntVal(v.Numerator()), IntVal(v.Denominator()) }
func (v RatioVal) Left() Native           { return IntVal(v.Numerator()) }
func (v RatioVal) Right() Native          { return IntVal(v.Denominator()) }

// IMAGINARY VALUE
func (v ImagVal) Bool() bool {
	if real(v) > 0 {
		return true
	}
	return false
}
func (v ImagVal) Unit() Native                  { return ImagVal(complex(0, 0)) }
func (v ImagVal) Uint() uint                    { return uint(real(v)) }
func (v ImagVal) Int() int                      { return int(real(v)) }
func (v ImagVal) Integer() IntVal               { return IntVal(real(v)) }
func (v ImagVal) Float() float64                { return float64(real(v)) }
func (v ImagVal) Rat() *big.Rat                 { return big.NewRat(int64(real(v)), int64(imag(v))) }
func (v ImagVal) Imag() complex128              { return complex128(v) }
func (v ImagVal) Imaginary() float64            { return imag(v) }
func (v ImagVal) Real() float64                 { return real(v) }
func (v ImagVal) BothFloat() (float64, float64) { return real(v), imag(v) }
func (v ImagVal) Both() (Native, Native)        { return FltVal(real(v)), FltVal(imag(v)) }
func (v ImagVal) Left() Native                  { return FltVal(v.Real()) }
func (v ImagVal) Right() Native                 { return FltVal(v.Imaginary()) }

/// TIME VALUE
func (v TimeVal) Time() time.Time  { return time.Time(v) }
func (v TimeVal) Uint() uint       { return uint(time.Time(v).Unix()) }
func (v TimeVal) Natural() UintVal { return UintVal(uint(time.Time(v).Unix())) }
func (v TimeVal) Int() int         { return int(time.Time(v).Unix()) }
func (v TimeVal) Integer() IntVal  { return IntVal(time.Time(v).Unix()) }
func (v TimeVal) Bool() bool       { return IntVal(v.Int()).Bool() }
func (v TimeVal) Rat() *big.Rat    { return IntVal(v.Int()).Rat() }
func (v TimeVal) Float() float64   { return IntVal(v.Int()).Float() }
func (v TimeVal) Imag() complex128 { return IntVal(v.Int()).Imag() }

/// DURATION VALUE
func (v DuraVal) Duration() time.Duration { return time.Duration(v) }
func (v DuraVal) Uint() uint              { return uint(v) }
func (v DuraVal) Natural() UintVal        { return UintVal(v.Uint()) }
func (v DuraVal) Int() int                { return int(v) }
func (v DuraVal) Integer() IntVal         { return IntVal(v.Int()) }
func (v DuraVal) Bool() bool              { return IntVal(v.Int()).Bool() }
func (v DuraVal) Rat() *big.Rat           { return IntVal(v.Int()).Rat() }
func (v DuraVal) Float() float64          { return IntVal(v.Int()).Float() }
func (v DuraVal) Imag() complex128        { return IntVal(v.Int()).Imag() }

func (v ByteVal) Bool() bool {
	if v > ByteVal(0) {
		return true
	}
	return false
}
func (v ByteVal) Unit() byte       { return byte(0) }
func (v ByteVal) Uint() uint       { return uint(v) }
func (v ByteVal) Natural() UintVal { return UintVal(uint(v)) }
func (v ByteVal) Int() int         { return int(v) }
func (v ByteVal) Integer() IntVal  { return IntVal(int(v)) }
func (v ByteVal) Rat() *big.Rat    { return IntVal(int(v)).Rat() }
func (v ByteVal) Float() float64   { return IntVal(int(v)).Float() }
func (v ByteVal) Imag() complex128 { return IntVal(int(v)).Imag() }
func (v ByteVal) Byte() byte       { return byte(v) }
func (v ByteVal) Bytes() []byte    { return []byte{v.Byte()} }
func (v ByteVal) Rune() rune       { return rune(v.Byte()) }
func (v ByteVal) String() string   { return string(v.Bytes()) }
func (v ByteVal) Len() int         { return bits.Len8(uint8(v.Uint())) }

func (v BytesVal) Unit() []byte     { return []byte{byte(0)} }
func (v BytesVal) Bytes() []byte    { return []byte(v) }
func (v BytesVal) ByteVec() ByteVec { return ByteVec(v) }
func (v BytesVal) String() string   { return string(v) }
func (v BytesVal) Runes() []rune    { return []rune(v.String()) }
func (v BytesVal) RuneVec() RuneVec { return RuneVec(v.Runes()) }
func (v BytesVal) Buffer() RuneVec  { return RuneVec(v.Runes()) }
func (v BytesVal) Len() int         { return len(v.Bytes()) }
func (v BytesVal) Bool() bool {
	for _, b := range v {
		if b > byte(0) {
			return true
		}
	}
	return false
}

/// STRING VALUE
func (v StrVal) Unit() Native { return StrVal(" ") }
func (v StrVal) Bool() (bool, error) {
	var truth, err = strconv.ParseBool(string(v))
	if err != nil {
		return false, err
	}
	return truth, nil
}
func (v StrVal) BoolVal() Native {
	var b, err = v.Bool()
	if err != nil {
		return NilVal{}
	}
	return BoolVal(b)
}
func (v StrVal) Bytes() BytesVal { return []byte(string(v)) }
func (v StrVal) String() string  { return string(v) }
func (v StrVal) Runes() RuneVec  { return []rune(string(v)) }
func (v StrVal) Uint() (uint, error) {
	u, err := strconv.ParseUint(string(v), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(u), nil
}
func (v StrVal) UintVal() Native {
	var u, err = v.Uint()
	if err != nil {
		return NilVal{}
	}
	return UintVal(u)
}
func (v StrVal) Len() int { return int(len(string(v))) }

// parse string to integer, float, duration, or time
func (v StrVal) Int() (int, error) {
	var s, err = strconv.Atoi(string(v))
	if err != nil {
		return 0, err
	}
	return int(s), nil
}
func (v StrVal) IntVal() Native {
	var val, err = v.Int()
	if err != nil {
		return NilVal{}
	}
	return IntVal(val)
}
func (v StrVal) Float() (float64, error) {
	var s, err = strconv.ParseFloat(string(v), 64)
	if err != nil {
		return 0.0, err
	}
	return float64(FltVal(s)), nil
}
func (v StrVal) FltVal() Native {
	var flt, err = v.Float()
	if err != nil {
		return NilVal{}
	}
	return FltVal(flt)
}
func (v StrVal) Duration() (time.Duration, error) {
	var d, err = time.ParseDuration(v.String())
	if err != nil {
		return time.Duration(0), err
	}
	return d, nil
}
func (v StrVal) DuraVal() Native {
	var dura, err = v.Duration()
	if err != nil {
		return NilVal{}
	}
	return DuraVal(dura)
}
func (v StrVal) Time(layout string) (time.Time, error) {
	t, err := time.Parse(layout, v.String())
	if err != nil {
		return time.Now(), err
	}
	return t, nil
}
func (v StrVal) TimeVal(layout string) Native {
	var tim, err = v.Time(layout)
	if err != nil {
		return NilVal{}
	}
	return TimeVal(tim)
}
