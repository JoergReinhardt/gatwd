package data

import (
	"math/big"
	"strconv"
	"time"
)

// string nullables
func (NilVal) String() string      { return Nil.String() }
func (v ErrorVal) String() string  { return v.e.Error() }
func (v ErrorVal) Error() ErrorVal { return ErrorVal{v.e} }
func (v BoolVal) String() string   { return strconv.FormatBool(bool(v)) }
func (v IntVal) String() string    { return strconv.Itoa(int(v)) }
func (v Int8Val) String() string   { return strconv.Itoa(int(v)) }
func (v Int16Val) String() string  { return strconv.Itoa(int(v)) }
func (v Int32Val) String() string  { return strconv.Itoa(int(v)) }
func (v UintVal) String() string   { return strconv.Itoa(int(v)) }
func (v Uint8Val) String() string  { return strconv.Itoa(int(v)) }
func (v Uint16Val) String() string { return strconv.Itoa(int(v)) }
func (v Uint32Val) String() string { return strconv.Itoa(int(v)) }
func (v ByteVal) String() string   { return strconv.Itoa(int(v)) }
func (v RuneVal) String() string   { return string(v) }
func (v BytesVal) String() string  { return string(v) }
func (v StrVal) String() string    { return string(v) }
func (v StrVal) Key() string       { return string(v) }
func (v TimeVal) String() string   { return time.Time(v).String() }
func (v DuraVal) String() string   { return time.Duration(v).String() }
func (v BigIntVal) String() string { return ((*big.Int)(&v)).String() }
func (v RatioVal) String() string  { return ((*big.Rat)(&v)).String() }
func (v BigFltVal) String() string { return ((*big.Float)(&v)).String() }
func (v FltVal) String() string {
	return strconv.FormatFloat(float64(v), 'G', -1, 64)
}
func (v Flt32Val) String() string {
	return strconv.FormatFloat(float64(v), 'G', -1, 32)
}
func (v ImagVal) String() string {
	return strconv.FormatFloat(float64(real(v)), 'G', -1, 64) + " + " +
		strconv.FormatFloat(float64(imag(v)), 'G', -1, 64) + "i"
}
func (v Imag64Val) String() string {
	return strconv.FormatFloat(float64(real(v)), 'G', -1, 32) + " + " +
		strconv.FormatFloat(float64(imag(v)), 'G', -1, 32) + "i"
}

// serializes bitflag to a string representation of the bitwise OR
// operation on a list of principle flags, that yielded this flag
func (v BitFlag) String() string {
	if Count(v) == 1 {
		return Type(v).String()
	}
	var str string
	var u = uint(1)
	var i = 0
	for i < 63 {
		if Match(BitFlag(u), v) {
			str = str + Type(u).String()
			if i < FlagLength(v)-1 {
				str = str + " "
			}
		}
		i = i + 1
		u = uint(1) << uint(i)
	}
	return str
}

// stringer for ordered chains, without any further specification.
func (v Chain) String() string {
	var str = "["
	for i, d := range v.Slice() {
		str = str + d.String()
		if i < v.Len()-1 {
			str = str + ", "
		}
	}
	str = str + "]"
	return str
}
func stringNativeSlice(v NativeVec) string {
	i := v.NativeSlice().([]interface{})
	var str = "["
	for n, d := range i {
		str = str + Con(d).String()
		if n < len(i)-1 {
			str = str + ", "
		}
	}
	str = str + "]"
	return str
}
func stringChain(v ...Data) string {
	var str = "["
	if s, ok := Chain(v).(NativeVec); ok {
		return stringNativeSlice(v)
	}
	for i, d := range v {
		str = str + d.String()
		if i < len(v)-1 {
			str = str + ", "
		}
	}
	str = str + "]"
	return str
}
func (v ErrorVec) String() string {
	var str string
	for i, err := range v {
		str = str +
			"ERROR " +
			strconv.Itoa(i) +
			": " +
			err.Error() +
			"\n"
	}
	return str
}
