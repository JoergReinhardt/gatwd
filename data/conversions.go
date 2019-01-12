package data

import (
	"math/bits"
	"strconv"
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

// INT -> VALUE
func (v IntVal) Integer() int { return int(v) } // implements Idx Attribut
func (v IntVal) Idx() IntVal  { return v }      // implements Idx Attribut
//func (v intVal) Key() strVal    { return v.String() } // implements Key Attribut
func (v IntVal) FltNat() FltVal { return FltVal(v) }
func (v IntVal) IntNat() IntVal { return v }
func (v IntVal) UintNat() UintVal {
	if v < 0 {
		return UintVal(v * -1)
	}
	return UintVal(v)
}

// VALUE -> INT
func (v Int8Val) Int() IntVal   { return IntVal(int(v)) }
func (v Int16Val) Int() IntVal  { return IntVal(int(v)) }
func (v Int32Val) Int() IntVal  { return IntVal(int(v)) }
func (v UintVal) Int() IntVal   { return IntVal(int(v)) }
func (v Uint16Val) Int() IntVal { return IntVal(int(v)) }
func (v Uint32Val) Int() IntVal { return IntVal(int(v)) }
func (v FltVal) Int() IntVal    { return IntVal(int(v)) }
func (v Flt32Val) Int() IntVal  { return IntVal(int(v)) }
func (v ByteVal) Int() IntVal   { return IntVal(int(v)) }
func (v ImagVal) Real() IntVal  { return IntVal(real(v)) }
func (v ImagVal) Imag() IntVal  { return IntVal(imag(v)) }
func (v StrVal) Int() IntVal {
	s, err := strconv.Atoi(string(v))
	if err != nil {
		return -1
	}
	return IntVal(s)
}

// VALUE -> FLOAT
func (v UintVal) Float() FltVal { return FltVal(v.Int().Float()) }
func (v IntVal) Float() FltVal  { return FltVal(v.FltNat()) }
func (v StrVal) Float() FltVal {
	s, err := strconv.ParseFloat(string(v), 64)
	if err != nil {
		return -1
	}
	return FltVal(s)
}

// VALUE -> UINT
func (v UintVal) Uint() UintVal { return v }
func (v UintVal) UintNat() uint { return uint(v) }
func (v IntVal) Uint() UintVal  { return UintVal(v.UintNat()) }
func (v StrVal) Uint() UintVal {
	u, err := strconv.ParseUint(string(v), 10, 64)
	if err != nil {
		return 0
	}
	return UintVal(u)
}
func (v BoolVal) Uint() UintVal {
	if v {
		return UintVal(1)
	}
	return UintVal(0)
}

// INTEGERS FOR DEDICATED PURPOSES
func (v UintVal) Len() IntVal  { return IntVal(bits.Len64(uint64(v))) }
func (v ByteVal) Len() IntVal  { return IntVal(bits.Len64(uint64(v))) }
func (v BytesVal) Len() IntVal { return IntVal(len(v)) }
func (v StrVal) Len() IntVal   { return IntVal(len(string(v))) }
