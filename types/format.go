package types

import (
	"math/big"
	"strconv"
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
	var str = "["
	if fcount(v) == 1 {
		str = str + Type(v).String()
	}
	var u = uint(1)
	var i = 0
	for i < 63 {
		if fmatch(BitFlag(u), v) {
			str = str + Type(u).String()
			if i < flen(v)-1 {
				str = str + "|"
			}
		}
		i = i + 1
		u = uint(1) << uint(i)
	}
	str = str + "]"
	return str
}
func (v flatCol) String() string {
	var slice = chainSlice(v())
	var length = len(slice)
	var str = "("
	for i, d := range slice {
		str = str + d.String()
		if i < length-1 {
			str = str + ", "
		}
	}
	str = str + ")"
	return str
}
func (v chain) String() string {
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
func referencedConsumeable(v Consumeable) string {
	var str = "["
	for i, d := range v.(Splitable).Slice() {
		str = str + d.String()
		if i < v.(Countable).Len()-1 {
			str = str + ", "
		}
	}
	str = str + "]"
	return str
}
func recursiveConsumeableString(v Consumeable) string {
	head, tail := v.Head(), v.Tail()
	str := "[" + head.String()
	if !tail.Empty() {
		str = "[" + head.String() + " " + recursiveConsumeableString(tail) + "]"
	}
	return str
}
func recolString(r reCol) string {
	head, tail := r()
	str := "[" + head.String()
	if !tail.Empty() {
		str = "[" + head.String() + " " + recolString(tail) + "]"
	}
	return str
}
func (r reCol) String() string { return recolString(r) }
func allTokens() []string {
	var str = []string{}
	var i uint
	var typ TokType = 1
	for i = 0; i < uint(len(syntax))-1; i++ {
		typ = 1 << i
		str = append(str, typ.String())
	}
	return str
}
func allSyntax() []string {
	var str = []string{}
	var i uint
	var typ TokType = 1
	for i = 0; i < uint(len(syntax))-1; i++ {
		typ = 1 << i
		str = append(str, typ.Syntax())
	}
	return str
}
func allTypes() []string {
	var str = []string{}
	var i uint
	var typ Type = 0
	for i = 0; i < uint(flen(flag(Natives)))-1; i++ {
		typ = 1 << i
		str = append(str, Type(typ).String())
	}
	return str
}
