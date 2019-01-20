package data

import (
	"math/big"
	"math/bits"
	"strconv"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
)

// string nullables
func (NilVal) String() string      { return Nil.String() }
func (v ErrorVal) String() string  { return "Error: " + v.e.Error() }
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
func (v BitFlag) String() string { return StringBitFlag(v) }
func StringBitFlag(v BitFlag) string {
	var str string
	if bits.OnesCount(v.Uint()) > 1 {
		for i, f := range FlagDecompose(v) {
			str = str + Type(f).String()
			if i < len(FlagDecompose(v))-1 {
				str = str + "|"
			}
		}
	} else {
		str = Type(v).String()
	}
	return str
}
func (v Chain) String() string    { return StringSlice(", ", "[", "]", v) }
func (v ErrorVec) String() string { return StringSlice("\n", "", "", v) }

// stringer for ordered chains, without any further specification.
func StringSlice(sep, ldelim, rdelim string, s ...Data) string {
	var str string
	str = str + ldelim
	for i, d := range s {
		if FlagMatch(d.Flag(), Vector.Flag()) {
			str = str + StringSlice(sep, ldelim, rdelim, d.(Chain).Slice()...)
		} else {
			str = str + d.String()
		}
		if i < len(s)-1 {
			str = str + sep
		}
	}
	str = str + rdelim
	return str
}
func StringChainTable(v ...[]Data) string {
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
func stringChainTable(v ...Data) string {
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
