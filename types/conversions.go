package types

import (
	"math/bits"
	"strconv"
	"time"
)

///// TYPE CONVERSION //////
/////
///// STRING (CONVERSION) METHODS ///////
func (nilVal) String() string      { return Nil.String() }
func (v boolVal) String() string   { return strconv.FormatBool(bool(v)) }
func (v intVal) String() string    { return strconv.Itoa(int(v)) }
func (v int8Val) String() string   { return strconv.Itoa(int(v)) }
func (v int16Val) String() string  { return strconv.Itoa(int(v)) }
func (v int32Val) String() string  { return strconv.Itoa(int(v)) }
func (v uintVal) String() string   { return strconv.Itoa(int(v)) }
func (v uint16Val) String() string { return strconv.Itoa(int(v)) }
func (v uint32Val) String() string { return strconv.Itoa(int(v)) }
func (v byteVal) String() string   { return strconv.Itoa(int(v)) }
func (v timeVal) String() string   { return time.Time(v).String() }
func (v duraVal) String() string   { return time.Duration(v).String() }
func (v bytesVal) String() string  { return string(v) }
func (v strVal) String() string    { return string(v) }
func (v strVal) Key() string       { return string(v) }
func (v errorVal) String() string  { return v.v.Error() }
func (v errorVal) Error() error    { return v.v }
func (v fltVal) String() string {
	return strconv.FormatFloat(float64(v), 'G', -1, 64)
}
func (v flt32Val) String() string {
	return strconv.FormatFloat(float64(v), 'G', -1, 32)
}
func (v slice) String() string {
	var str string
	for i, val := range v.Slice() {
		str = str + val.String()
		if i < v.Len()-1 {
			str = str + ", "
		}
	}
	return str
}
func (v imagVal) String() string {
	return strconv.FormatFloat(float64(real(v)), 'G', -1, 64) + " + " +
		strconv.FormatFloat(float64(imag(v)), 'G', -1, 64) + "i"
}
func (v imag64Val) String() string {
	return strconv.FormatFloat(float64(real(v)), 'G', -1, 32) + " + " +
		strconv.FormatFloat(float64(imag(v)), 'G', -1, 32) + "i"
}
func (s collection) String() (str string) {
	for i, v := range s.s {
		str = str + "\t" + strconv.Itoa(i) + "\t" + v.String() + "\n"
	}
	return str
}
func flagSet(f Typed, b uint) bool {
	var u uint
	u = 1 << b
	if _, ok := Typed(flag(ValType(u))).(flag); ok {
		return true
	}
	return false
}
func (v flag) String() string {
	if bits.OnesCount(v.uint()) == 1 {
		return ValType(v).String()
	}
	var str string
	var u, i uint
	for u < uint(NATIVES) {
		if v.Type().match(ValType(u)) {
			str = str + ValType(u).String() + "\n"
		}
		i = i + 1
		u = uint(1) << i
	}
	return str
}

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
func (v intVal) Idx() int        { return int(v) }     // implements Idx Attribut
func (v intVal) Key() string     { return v.String() } // implements Key Attribut
func (v intVal) FltNat() float64 { return float64(v) }
func (v intVal) IntNat() int     { return int(v) }
func (v intVal) UintNat() uint {
	if v < 0 {
		return uint(v * -1)
	}
	return uint(v)
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
func (v slice) Slice() []Value { return v }
func (v slice) Len() int       { return len(v) }
