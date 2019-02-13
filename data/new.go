package data

import (
	"math/big"
	"time"
)

func NewFromPrimary(vals ...Primary) Primary {
	var ifs = []interface{}{}
	for _, val := range vals {
		ifs = append(ifs, val.(interface{}))
	}
	return New(ifs...)
}
func NewUnboxedVector(f BitFlag, vals ...Primary) Primary { return conVec(f, vals...) }
func New(vals ...interface{}) Primary                     { dat, _ := NewWithTypeInfo(vals...); return dat }
func NewWithTypeInfo(vals ...interface{}) (rval Primary, flag BitFlag) {

	if len(vals) == 0 {
		return nil, Nil.TypePrime().Flag()
	}
	var val = vals[0]
	if len(vals) > 1 {
		var dat = DataSlice(make([]Primary, 0, len(vals)))
		for _, val := range vals {
			var d Primary
			d, flag = NewWithTypeInfo(val)
			flag = flag | d.TypePrime().Flag()
			dat = append(dat, d)
		}
		if FlagLength(flag) == 1 {
			return conVec(flag, dat...), flag
		}
		return dat, flag
	}
	switch val.(type) {
	case bool:
		rval = BoolVal(val.(bool))
	case int, int64:
		rval = IntVal(val.(int))
	case int8:
		rval = Int8Val(val.(int8))
	case int16:
		rval = Int16Val(val.(int16))
	case int32:
		rval = Int32Val(val.(int32))
	case uint, uint64:
		rval = UintVal(val.(uint))
	case uint16:
		rval = Uint16Val(val.(uint16))
	case uint32:
		rval = Int32Val(val.(int32))
	case float32:
		rval = Flt32Val(val.(float32))
	case float64:
		rval = FltVal(val.(float64))
	case complex64:
		rval = ImagVal(val.(complex64))
	case complex128:
		rval = ImagVal(val.(complex128))
	case byte:
		rval = ByteVal(val.(byte))
	case []byte:
		rval = BytesVal(val.([]byte))
	case string:
		rval = StrVal(val.(string))
	case error:
		rval = ErrorVal{val.(error)}
	case time.Time:
		rval = TimeVal(val.(time.Time))
	case time.Duration:
		rval = DuraVal(val.(time.Duration))
	case *big.Int:
		v := BigIntVal(*val.(*big.Int))
		rval = &v
	case *big.Float:
		v := BigFltVal(*val.(*big.Float))
		rval = &v
	case *big.Rat:
		v := RatioVal(*val.(*big.Rat))
		rval = &v
	case Primary:
		rval = val.(Primary)
	case []Primary:
		rval = DataSlice(val.([]Primary))
	}
	return rval, flag
}
func conVec(f BitFlag, d ...Primary) (val Primary) {
	var slice DataSlice = []Primary{}
	switch {
	case FlagMatch(f, Nil.TypePrime().Flag()):
		for _, v := range d {
			slice = append(slice, v.(NilVal))
		}
	case FlagMatch(f, Bool.TypePrime().Flag()):
		for _, v := range d {
			slice = append(slice, v.(BoolVal))
		}
	case FlagMatch(f, Int.TypePrime().Flag()):
		for _, v := range d {
			slice = append(slice, v.(IntVal))
		}
	case FlagMatch(f, Int8.TypePrime().Flag()):
		for _, v := range d {
			slice = append(slice, v.(Int8Val))
		}
	case FlagMatch(f, Int16.TypePrime().Flag()):
		for _, v := range d {
			slice = append(slice, v.(Int16Val))
		}
	case FlagMatch(f, Int32.TypePrime().Flag()):
		for _, v := range d {
			slice = append(slice, v.(Int32Val))
		}
	case FlagMatch(f, Uint.TypePrime().Flag()):
		for _, v := range d {
			slice = append(slice, v.(UintVal))
		}
	case FlagMatch(f, Uint8.TypePrime().Flag()):
		for _, v := range d {
			slice = append(slice, v.(Uint8Val))
		}
	case FlagMatch(f, Uint16.TypePrime().Flag()):
		for _, v := range d {
			slice = append(slice, v.(Uint16Val))
		}
	case FlagMatch(f, Uint32.TypePrime().Flag()):
		for _, v := range d {
			slice = append(slice, v.(Uint32Val))
		}
	case FlagMatch(f, Float.TypePrime().Flag()):
		for _, v := range d {
			slice = append(slice, v.(FltVal))
		}
	case FlagMatch(f, Flt32.TypePrime().Flag()):
		for _, v := range d {
			slice = append(slice, v.(Flt32Val))
		}
	case FlagMatch(f, Imag.TypePrime().Flag()):
		for _, v := range d {
			slice = append(slice, v.(Imag64Val))
		}
	case FlagMatch(f, Imag64.TypePrime().Flag()):
		for _, v := range d {
			slice = append(slice, v.(Imag64Val))
		}
	case FlagMatch(f, Byte.TypePrime().Flag()):
		for _, v := range d {
			slice = append(slice, v.(ByteVal))
		}
	case FlagMatch(f, Rune.TypePrime().Flag()):
		for _, v := range d {
			slice = append(slice, v.(RuneVal))
		}
	case FlagMatch(f, Bytes.TypePrime().Flag()):
		for _, v := range d {
			slice = append(slice, v.(BytesVal))
		}
	case FlagMatch(f, String.TypePrime().Flag()):
		for _, v := range d {
			slice = append(slice, v.(StrVal))
		}
	case FlagMatch(f, BigInt.TypePrime().Flag()):
		for _, v := range d {
			slice = append(slice, v.(BigIntVal))
		}
	case FlagMatch(f, BigFlt.TypePrime().Flag()):
		for _, v := range d {
			slice = append(slice, v.(BigFltVal))
		}
	case FlagMatch(f, Ratio.TypePrime().Flag()):
		for _, v := range d {
			slice = append(slice, v.(RatioVal))
		}
	case FlagMatch(f, Time.TypePrime().Flag()):
		for _, v := range d {
			slice = append(slice, v.(TimeVal))
		}
	case FlagMatch(f, Duration.TypePrime().Flag()):
		for _, v := range d {
			slice = append(slice, v.(DuraVal))
		}
	case FlagMatch(f, Error.TypePrime().Flag()):
		for _, v := range d {
			slice = append(slice, v.(ErrorVal))
		}
	}
	return slice
}
