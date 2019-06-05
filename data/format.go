package data

import (
	"bytes"
	"math/big"
	"math/bits"
	"strconv"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
)

//// NATIVE SLICES /////

func (v NilVec) String() string    { return StringSlice(", ", "[", "]", v) }
func (v BoolVec) String() string   { return StringSlice(", ", "[", "]", v) }
func (v IntVec) String() string    { return StringSlice(", ", "[", "]", v) }
func (v Int8Vec) String() string   { return StringSlice(", ", "[", "]", v) }
func (v Int16Vec) String() string  { return StringSlice(", ", "[", "]", v) }
func (v Int32Vec) String() string  { return StringSlice(", ", "[", "]", v) }
func (v UintVec) String() string   { return StringSlice(", ", "[", "]", v) }
func (v Uint8Vec) String() string  { return StringSlice(", ", "[", "]", v) }
func (v Uint16Vec) String() string { return StringSlice(", ", "[", "]", v) }
func (v Uint32Vec) String() string { return StringSlice(", ", "[", "]", v) }
func (v FltVec) String() string    { return StringSlice(", ", "[", "]", v) }
func (v Flt32Vec) String() string  { return StringSlice(", ", "[", "]", v) }
func (v ImagVec) String() string   { return StringSlice(", ", "[", "]", v) }
func (v Imag64Vec) String() string { return StringSlice(", ", "[", "]", v) }
func (v ByteVec) String() string   { return string([]byte(v)) }
func (v RuneVec) String() string   { return StringSlice(", ", "[", "]", v) }
func (v BytesVec) String() string  { return StringSlice(", ", "[", "]", v) }
func (v StrVec) String() string    { return StringSlice(", ", "[", "]", v) }
func (v BigIntVec) String() string { return StringSlice(", ", "[", "]", v) }
func (v BigFltVec) String() string { return StringSlice(", ", "[", "]", v) }
func (v RatioVec) String() string  { return StringSlice(", ", "[", "]", v) }
func (v TimeVec) String() string   { return StringSlice(", ", "[", "]", v) }
func (v DuraVec) String() string   { return StringSlice(", ", "[", "]", v) }
func (v FlagSet) String() string   { return StringSlice(", ", "[", "]", v) }
func (v SetVal) String() string    { return StringSlice(", ", "[", "]", v) }

// string nullables
func (NilVal) String() string      { return Nil.String() }
func (v ErrorVal) String() string  { return "Error: " + v.E.Error() }
func (v ErrorVal) Error() ErrorVal { return ErrorVal{v.E} }
func (v BoolVal) String() string   { return strconv.FormatBool(bool(v)) }
func (v IntVal) String() string    { return strconv.Itoa(int(v)) }
func (v Int8Val) String() string   { return strconv.Itoa(int(v)) }
func (v Int16Val) String() string  { return strconv.Itoa(int(v)) }
func (v Int32Val) String() string  { return strconv.Itoa(int(v)) }
func (v UintVal) String() string   { return strconv.Itoa(int(v)) }
func (v Uint8Val) String() string  { return strconv.Itoa(int(v)) }
func (v Uint16Val) String() string { return strconv.Itoa(int(v)) }
func (v Uint32Val) String() string { return strconv.Itoa(int(v)) }
func (v RuneVal) String() string   { return string(v) }
func (v StrVal) Key() string       { return string(v) }
func (v TimeVal) String() string   { return "" + time.Time(v).String() }
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
func (v DataSlice) String() string { return StringSlice(", ", "[", "]", v.Slice()...) }

// serializes bitflag to a string representation of the bitwise OR
// operation on a list of principle flags, that yielded this flag
func (v BitFlag) String() string { return StringBitFlag(v) }
func StringBitFlag(v BitFlag) string {
	var str string
	if bits.OnesCount(v.Uint()) > 1 {
		for i, f := range FlagDecompose(v) {
			str = str + f.(TyNat).String()
			if i < len(FlagDecompose(v))-1 {
				str = str + "∙"
			}
		}
	} else {
		str = TyNat(v).String()
	}
	return str
}
func (v ErrorVec) String() string { return StringSlice("\n", "", "", v) }

// stringer for ordered chains, without any further specification.
func StringSlice(sep, ldelim, rdelim string, s ...Native) string {
	var str string
	str = str + ldelim
	for i, d := range s {
		if FlagMatch(d.TypeNat().Flag(), Slice.TypeNat().Flag()) {
			str = str + StringSlice(sep, ldelim, rdelim, d.(DataSlice).Slice()...)
		}
		if i < len(s)-1 {
			str = str + sep
		}
	}
	str = str + rdelim
	return str
}

func StringChainTable(v ...[]Native) string {
	var str = &strings.Builder{}
	var tab = tablewriter.NewWriter(str)
	tab.SetBorder(false)
	tab.SetColumnSeparator(" ")
	tab.SetAlignment(tablewriter.ALIGN_LEFT)
	for _, dr := range v {
		var row = []string{}
		for _, d := range dr {
			row = append(row, d.String())
		}
		tab.Append(row)
	}
	tab.Render()
	return str.String()
}

func stringChainTable(v ...Native) string {
	str := &strings.Builder{}
	tab := tablewriter.NewWriter(str)
	for i, d := range v {
		row := []string{
			strconv.Itoa(i), d.String(),
		}
		tab.Append(row)
	}
	tab.Render()
	return str.String()
}

//// SETS ////
func (p PairVal) String() string {
	return p.Left().String() + ": " + p.Right().String()
}

func (s SetInt) String() string {
	var str = bytes.NewBuffer([]byte{})
	for k, v := range s {
		str.WriteString(k.String())
		str.WriteString(": ")
		str.WriteString(v.String())
	}
	return str.String()
}

func (s SetUint) String() string {
	var str = bytes.NewBuffer([]byte{})
	for k, v := range s {
		str.WriteString(k.String())
		str.WriteString(": ")
		str.WriteString(v.String())
	}
	return str.String()
}

func (s SetFloat) String() string {
	var str = bytes.NewBuffer([]byte{})
	for k, v := range s {
		str.WriteString(k.String())
		str.WriteString(": ")
		str.WriteString(v.String())
	}
	return str.String()
}

func (s SetFlag) String() string {
	var str = bytes.NewBuffer([]byte{})
	for k, v := range s {
		str.WriteString(k.String())
		str.WriteString(": ")
		str.WriteString(v.String())
	}
	return str.String()
}

func (s SetString) String() string {
	var str = bytes.NewBuffer([]byte{})
	for k, v := range s {
		str.WriteString(k.String())
		str.WriteString(": ")
		str.WriteString(v.String())
	}
	return str.String()
}
