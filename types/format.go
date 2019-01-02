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
func (a keyAcc) String() string {
	return "[" + a.Acc().String() + "]⇒ ( " + a.Value().Eval().String() + ")"
}
func (a idxAcc) String() string   { return "[" + a.Acc().String() + "]⇒ (" + a.Value().String() + ")" }
func (a accessor) String() string { return "[" + a.Acc().String() + "]⇒ (" + a.Value().String() + ")" }
func (c paramSet) String() string {
	var str string
	for i, p := range c {
		str = str + p().String()
		if i < len(c)-1 {
			str = str + ", "
		}
	}
	return str
}
func (c argSet) String() string {
	var str string
	for i, a := range c {
		str = str + a().String()
		if i < len(c)-1 {
			str = str + ", "
		}
	}
	return str
}
func (c retValSet) String() string {
	var str string
	for i, r := range c {
		str = str + r().String()
		if i < len(c)-1 {
			str = str + ", "
		}
	}
	return str
}

func (c cons) String() string      { return "λ → " + c().String() }
func (c unc) String() string       { return "d → λ → d" }
func (c bnc) String() string       { return "d → d → λ → d" }
func (c fnc) String() string       { return "d → ‥.→ λ → d" }
func (c constFnc) String() string  { return "λ → " + c().String() }
func (c unaryFnc) String() string  { return "Data → λ → Data" }
func (c binaryFnc) String() string { return "Data → Data → λ → Data" }
func (c naryFnc) String() string   { return "Data → ‥.→ λ → Data" }
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
