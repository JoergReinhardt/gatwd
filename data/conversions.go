package data

import (
	"math/big"
	"math/bits"
	"strconv"
	"time"
)

//func TypeCast(vals ...Data, b BitFlag) Data        { return NewData(vals...) }
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
func (v StrVal) Bool() BoolVal {
	s, err := strconv.ParseBool(string(v))
	if err != nil {
		return false
	}
	return BoolVal(s)
}
func (v StrVal) Bytes() BytesVal { return []byte(string(v)) }
func (v StrVal) Runes() RuneVec  { return []rune(string(v)) }
func (v StrVal) Uints() Uint32Vec {
	var runes = []rune(string(v))
	var uints = make([]uint32, 0, len(runes))
	for _, r := range runes {
		uints = append(uints, uint32(r))
	}
	return uints
}

// INT -> VALUE
func (v IntVal) Rat() *big.Rat   { return big.NewRat(1, int64(v)) } // implements Idx Attribut
func (v IntVal) Int() int        { return int(v) }                  // implements Idx Attribut
func (v IntVal) Integer() IntVal { return v }                       // implements Idx Attribut
func (v IntVal) Idx() int        { return int(v) }                  // implements Idx Attribut
//func (v intVal) Key() strVal    { return v.String() } // implements Key Attribut
func (v IntVal) FltNat() FltVal { return FltVal(v) }
func (v IntVal) IntNat() IntVal { return v }
func (v IntVal) UintNat() UintVal {
	if v < 0 {
		return UintVal(v * -1)
	}
	return UintVal(v)
}

// FLOAT -> VALUE
func (v FltVal) Float() float64    { return float64(v) }
func (v Flt32Val) Float() float64  { return float64(v) }
func (v BigFltVal) Float() float64 { f, _ := (*big.Float)(&v).Float64(); return f }

// UINT -> VALUE
func (v Uint8Val) Byte() byte { return byte(v) }

// VALUE -> UINT
func (v Uint8Val) Uint() uint  { return uint(v) }
func (v Uint16Val) Uint() uint { return uint(v) }
func (v Uint32Val) Uint() uint { return uint(v) }
func (v UintVal) Uint() uint   { return uint(v) }
func (v FltVal) Uint() uint    { return uint(v) }
func (v Flt32Val) Uint() uint  { return uint(v) }
func (v ByteVal) Uint() uint   { return uint(v) }
func (v ImagVal) Uint() uint   { return uint(real(v)) }

// VALUE -> INT
func (v Int8Val) Int() int   { return int(v) }
func (v Int16Val) Int() int  { return int(v) }
func (v Int32Val) Int() int  { return int(v) }
func (v UintVal) Int() int   { return int(v) }
func (v Uint16Val) Int() int { return int(v) }
func (v Uint32Val) Int() int { return int(v) }
func (v FltVal) Int() int    { return int(v) }
func (v Flt32Val) Int() int  { return int(v) }
func (v ByteVal) Int() int   { return int(v) }
func (v TimeVal) Int() int   { return int(time.Time(v).Unix()) }
func (v DuraVal) Int() int   { return int(v) }
func (v ImagVal) Real() int  { return int(real(v)) }
func (v ImagVal) Imag() int  { return int(imag(v)) }
func (v StrVal) Int() int {
	s, err := strconv.Atoi(string(v))
	if err != nil {
		return -1
	}
	return int(s)
}

// VALUE -> FLOAT
func (v UintVal) Float() float64 { return v.Float() }
func (v IntVal) Float() float64  { return float64(v.FltNat()) }
func (v StrVal) Float() float64 {
	s, err := strconv.ParseFloat(string(v), 64)
	if err != nil {
		return -1
	}
	return float64(FltVal(s))
}

// VALUE -> UINT
func (v IntVal) Uint() uint { return uint(v.UintNat()) }
func (v StrVal) Uint() uint {
	u, err := strconv.ParseUint(string(v), 10, 64)
	if err != nil {
		return 0
	}
	return uint(u)
}

func (b BoolVal) Bool() bool { return bool(b) }

func (v BoolVal) Uint() uint {
	if v {
		return uint(1)
	}
	return uint(0)
}

// INTEGERS FOR DEDICATED PURPOSES
func (v UintVal) Len() int  { return int(bits.Len64(uint64(v))) }
func (v ByteVal) Len() int  { return int(bits.Len64(uint64(v))) }
func (v BytesVal) Len() int { return int(len(v)) }
func (v StrVal) Len() int   { return int(len(string(v))) }
