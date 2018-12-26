package types

import (
	"fmt"
	"math/big"

	"time"
)

func conData(vals ...interface{}) (rval Data) {
	var val interface{}
	if len(vals) == 0 {
		return nilVal{}
	}
	if len(vals) > 1 {
		sl := newSlice()
		for _, val := range vals {
			sl = slicePut(sl, conData(val))
		}
		return sl
	}
	val = vals[0]
	switch val.(type) {
	case bool:
		rval = boolVal(val.(bool))
	case int, int64:
		rval = intVal(val.(int))
	case int8:
		rval = int8Val(val.(int8))
	case int16:
		rval = int16Val(val.(int16))
	case int32:
		rval = int32Val(val.(int32))
	case uint, uint64:
		rval = uintVal(val.(uint))
	case uint16:
		rval = uint16Val(val.(uint16))
	case uint32:
		rval = int32Val(val.(int32))
	case float32:
		rval = flt32Val(val.(float32))
	case float64:
		rval = fltVal(val.(float64))
	case complex64:
		rval = imagVal(val.(complex64))
	case complex128:
		rval = imagVal(val.(complex128))
	case byte:
		rval = byteVal(val.(byte))
	case []byte:
		rval = bytesVal(val.([]byte))
	case string:
		rval = strVal(val.(string))
	case error:
		rval = errorVal{val.(error)}
	case time.Time:
		rval = timeVal(val.(time.Time))
	case time.Duration:
		rval = duraVal(val.(time.Duration))
	case *big.Int:
		v := bigIntVal(*val.(*big.Int))
		rval = &v
	case *big.Float:
		v := bigFltVal(*val.(*big.Float))
		rval = &v
	case *big.Rat:
		v := ratioVal(*val.(*big.Rat))
		rval = &v
	case Data:
		rval = val.(Data)
	case []Data:
		rval = slice(val.([]Data))
	case FnType, Type, Typed:
		rval = BitFlag(val.(Type))
	}
	return rval
}

//// GENERATE NULL VALUE OF EACH TYPE ////////
func conNull(t Typed) (val Data) {
	switch {
	case Nil.Flag().Match(t):
		return nilVal{}
	case Bool.Flag().Match(t):
		return conData(false)
	case Int.Flag().Match(t):
		return conData(0)
	case Int8.Flag().Match(t):
		return conData(int8(0))
	case Int16.Flag().Match(t):
		return conData(int16(0))
	case Int32.Flag().Match(t):
		return conData(int32(0))
	case BigInt.Flag().Match(t):
		return conData(big.NewInt(0))
	case Uint.Flag().Match(t):
		return conData(uint(0))
	case Uint16.Flag().Match(t):
		return conData(uint16(0))
	case Uint32.Flag().Match(t):
		return conData(uint32(0))
	case Float.Flag().Match(t):
		return conData(float64(0))
	case Flt32.Flag().Match(t):
		return conData(float32(0))
	case BigFlt.Flag().Match(t):
		return conData(big.NewFloat(0))
	case Ratio.Flag().Match(t):
		return conData(big.NewRat(1, 1))
	case Imag.Flag().Match(t):
		return conData(complex128(float64(0)))
	case Imag64.Flag().Match(t):
		return conData(complex64(float32(0)))
	case Byte.Flag().Match(t):
		var b = 0
		return conData(b)
	case Bytes.Flag().Match(t):
		var b = []byte{}
		return conData(b)
	case Rune.Flag().Match(t):
		var b = ' '
		return conData(b)
	case String.Flag().Match(t):
		s := " "
		return conData(s)
	case Error.Flag().Match(t):
		var e = fmt.Errorf("")
		return conData(e)
	case t.Flag().Match(BigInt):
		v := &big.Int{}
		return conData(v)
	case t.Flag().Match(BigFlt):
		v := &big.Float{}
		return conData(v)
	case t.Flag().Match(Ratio):
		v := &big.Rat{}
		return conData(v)
	}
	return val
}
