package types

import (
	"math/bits"
	"strconv"
)

///// TYPE CONVERSION //////
// BOOL -> VALUE
func (v boolVal) Int() intVal {
	if v {
		return intVal(1)
	}
	return intVal(-1)
}
func (v boolVal) IntNat() int {
	if v {
		return 1
	}
	return -1
}
func (v boolVal) UintNat() uint {
	if v {
		return 1
	}
	return 0
}

// VALUE -> BOOL
func (v intVal) Bool() boolVal {
	if v == 1 {
		return boolVal(true)
	}
	return boolVal(false)
}
func (v strVal) Bool() boolVal {
	s, err := strconv.ParseBool(string(v))
	if err != nil {
		return false
	}
	return boolVal(s)
}

// INT -> VALUE
func (v intVal) Idx() intVal { return v } // implements Idx Attribut
//func (v intVal) Key() strVal    { return v.String() } // implements Key Attribut
func (v intVal) FltNat() fltVal { return fltVal(v) }
func (v intVal) IntNat() intVal { return v }
func (v intVal) UintNat() uintVal {
	if v < 0 {
		return uintVal(v * -1)
	}
	return uintVal(v)
}

// VALUE -> INT
func (v int8Val) Int() intVal   { return intVal(int(v)) }
func (v int16Val) Int() intVal  { return intVal(int(v)) }
func (v int32Val) Int() intVal  { return intVal(int(v)) }
func (v uintVal) Int() intVal   { return intVal(int(v)) }
func (v uint16Val) Int() intVal { return intVal(int(v)) }
func (v uint32Val) Int() intVal { return intVal(int(v)) }
func (v fltVal) Int() intVal    { return intVal(int(v)) }
func (v flt32Val) Int() intVal  { return intVal(int(v)) }
func (v byteVal) Int() intVal   { return intVal(int(v)) }
func (v imagVal) Real() intVal  { return intVal(real(v)) }
func (v imagVal) Imag() intVal  { return intVal(imag(v)) }
func (v strVal) Int() intVal {
	s, err := strconv.Atoi(string(v))
	if err != nil {
		return -1
	}
	return intVal(s)
}

// VALUE -> FLOAT
func (v uintVal) Float() fltVal { return fltVal(v.Int().Float()) }
func (v intVal) Float() fltVal  { return fltVal(v.FltNat()) }
func (v strVal) Float() fltVal {
	s, err := strconv.ParseFloat(string(v), 64)
	if err != nil {
		return -1
	}
	return fltVal(s)
}

// VALUE -> UINT
func (v uintVal) Uint() uintVal { return v }
func (v uintVal) UintNat() uint { return uint(v) }
func (v intVal) Uint() uintVal  { return uintVal(v.UintNat()) }
func (v strVal) Uint() uintVal {
	u, err := strconv.ParseUint(string(v), 10, 64)
	if err != nil {
		return 0
	}
	return uintVal(u)
}
func (v boolVal) Uint() uintVal {
	if v {
		return uintVal(1)
	}
	return uintVal(0)
}

// INTEGERS FOR DEDICATED PURPOSES
func (v uintVal) Len() intVal  { return intVal(bits.Len64(uint64(v))) }
func (v byteVal) Len() intVal  { return intVal(bits.Len64(uint64(v))) }
func (v bytesVal) Len() intVal { return intVal(len(v)) }
func (v strVal) Len() intVal   { return intVal(len(string(v))) }

// SLICE ->
func (v chain) Slice() []Data { return v }
func (v chain) Len() int      { return len(v) }
