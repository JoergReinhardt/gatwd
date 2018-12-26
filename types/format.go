package types

import (
	"fmt"
	"math/big"
	"math/bits"
	"strconv"
	"strings"
	"time"
)

func (nilVal) String() string      { return Nil.String() }
func (v errorVal) String() string  { return v.v.Error() }
func (v errorVal) Error() errorVal { return errorVal{v.v} }
func (v boolVal) String() string   { return strconv.FormatBool(bool(v)) }
func (v intVal) String() string    { return strconv.Itoa(int(v)) }
func (v int8Val) String() string   { return strconv.Itoa(int(v)) }
func (v int16Val) String() string  { return strconv.Itoa(int(v)) }
func (v int32Val) String() string  { return strconv.Itoa(int(v)) }
func (v uintVal) String() string   { return strconv.Itoa(int(v)) }
func (v uint8Val) String() string  { return strconv.Itoa(int(v)) }
func (v uint16Val) String() string { return strconv.Itoa(int(v)) }
func (v uint32Val) String() string { return strconv.Itoa(int(v)) }
func (v byteVal) String() string   { return strconv.Itoa(int(v)) }
func (v runeVal) String() string   { return string(v) }
func (v bytesVal) String() string  { return string(v) }
func (v strVal) String() string    { return string(v) }
func (v strVal) Key() string       { return string(v) }
func (v timeVal) String() string   { return time.Time(v).String() }
func (v duraVal) String() string   { return time.Duration(v).String() }
func (v bigIntVal) String() string { return ((*big.Int)(&v)).String() }
func (v ratioVal) String() string  { return ((*big.Rat)(&v)).String() }
func (v bigFltVal) String() string { return ((*big.Float)(&v)).String() }
func (v fltVal) String() string {
	return strconv.FormatFloat(float64(v), 'G', -1, 64)
}
func (v flt32Val) String() string {
	return strconv.FormatFloat(float64(v), 'G', -1, 32)
}
func (v imagVal) String() string {
	return strconv.FormatFloat(float64(real(v)), 'G', -1, 64) + " + " +
		strconv.FormatFloat(float64(imag(v)), 'G', -1, 64) + "i"
}
func (v imag64Val) String() string {
	return strconv.FormatFloat(float64(real(v)), 'G', -1, 32) + " + " +
		strconv.FormatFloat(float64(imag(v)), 'G', -1, 32) + "i"
}
func (v BitFlag) String() string {
	if uint(bits.OnesCount(v.Uint())) == 1 {
		return Type(v).String()
	}
	len := uint(flen(v))
	str := &strings.Builder{}
	var err error
	var u, i uint
	for u < uint(Tree) {
		if v.Flag().Match(Type(u)) {
			_, err = (*str).WriteString(Type(u).String())
			if i < len-1 {
				_, err = (*str).WriteString(" | ")
			}
		}
		i = i + 1
		u = uint(1) << i
	}
	if err != nil {
		return "ERROR: could not decompose value type name to string"
	}
	return str.String()
}
func (v slice) String() string {
	var err error
	str := &strings.Builder{}
	_, err = (*str).WriteString("[")
	for i, val := range v.Slice() {
		_, err = (*str).WriteString(fmt.Sprintf("%v", val))
		if i < v.Len()-1 {
			(*str).WriteString(", ")
		}
	}
	_, err = (*str).WriteString("]")
	if err != nil {
		return "ERROR: could not concatenate slice values to string"
	}
	return str.String()
}
