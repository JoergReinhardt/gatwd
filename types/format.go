package types

import (
	"fmt"
	"math/big"
	"math/bits"
	"strconv"
	"strings"
	"time"
)

func (nilVal) String() strVal      { return strVal(Nil.String()) }
func (v errorVal) String() strVal  { return strVal(v.v.Error()) }
func (v errorVal) Error() errorVal { return errorVal{v.v} }
func (v boolVal) String() strVal   { return strVal(strconv.FormatBool(bool(v))) }
func (v intVal) String() strVal    { return strVal(strconv.Itoa(int(v))) }
func (v int8Val) String() strVal   { return strVal(strconv.Itoa(int(v))) }
func (v int16Val) String() strVal  { return strVal(strconv.Itoa(int(v))) }
func (v int32Val) String() strVal  { return strVal(strconv.Itoa(int(v))) }
func (v uintVal) String() strVal   { return strVal(strconv.Itoa(int(v))) }
func (v uint8Val) String() strVal  { return strVal(strconv.Itoa(int(v))) }
func (v uint16Val) String() strVal { return strVal(strconv.Itoa(int(v))) }
func (v uint32Val) String() strVal { return strVal(strconv.Itoa(int(v))) }
func (v byteVal) String() strVal   { return strVal(strconv.Itoa(int(v))) }
func (v runeVal) String() strVal   { return strVal(string(v)) }
func (v bytesVal) String() strVal  { return strVal(string(v)) }
func (v strVal) String() strVal    { return strVal(string(v)) }
func (v strVal) Key() strVal       { return strVal(string(v)) }
func (v timeVal) String() strVal   { return strVal(time.Time(v).String()) }
func (v duraVal) String() strVal   { return strVal(time.Duration(v).String()) }
func (v bigIntVal) String() strVal { return strVal(((*big.Int)(&v)).String()) }
func (v ratioVal) String() strVal  { return strVal(((*big.Rat)(&v)).String()) }
func (v bigFltVal) String() strVal { return strVal(((*big.Float)(&v)).String()) }
func (v fltVal) String() strVal {
	return strVal(strconv.FormatFloat(float64(v), 'G', -1, 64))
}
func (v flt32Val) String() strVal {
	return strVal(strconv.FormatFloat(float64(v), 'G', -1, 32))
}
func (v imagVal) String() strVal {
	return strVal(strconv.FormatFloat(float64(real(v)), 'G', -1, 64) + " + " +
		strconv.FormatFloat(float64(imag(v)), 'G', -1, 64) + "i")
}
func (v imag64Val) String() strVal {
	return strVal(strconv.FormatFloat(float64(real(v)), 'G', -1, 32) + " + " +
		strconv.FormatFloat(float64(imag(v)), 'G', -1, 32) + "i")
}
func (v flag) String() strVal {
	if uint(bits.OnesCount(v.Uint())) == 1 {
		return strVal(ValType(v).String())
	}
	len := uint(flen(v))
	str := &strings.Builder{}
	var err error
	var u, i uint
	for u < uint(Tree) {
		if v.Flag().Match(ValType(u)) {
			_, err = (*str).WriteString(ValType(u).String())
			if i < len-1 {
				_, err = (*str).WriteString(" | ")
			}
		}
		i = i + 1
		u = uint(1) << i
	}
	if err != nil {
		return strVal("ERROR: could not decompose value type name to string")
	}
	return strVal(str.String())
}
func (v slice) String() strVal {
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
	return strVal(str.String())
}
