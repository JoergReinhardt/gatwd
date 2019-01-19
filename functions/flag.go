/*
TYPE FLAG

  compose id, precedence type flag and kind flag to form part of a types unique
  idetity, to first pass filter tokens by combination of propertys in a highly
  efficient manner.
*/
package functions

import (
	"math/bits"

	d "github.com/JoergReinhardt/godeep/data"
)

///// BIT_FLAG ////////
func composeHighLow(high, low BitFlag) BitFlag {
	return d.FlagHigh(BitFlag(high).Flag() | d.FlagLow(BitFlag(low)).Flag())
}

type Flag func() (kind Kind, prec d.BitFlag)

func newFlag(kind Kind, prec d.BitFlag) Flag {
	return func() (k Kind, p d.BitFlag) { return Kind(kind), prec }
}
func (t Flag) Flag() d.BitFlag { return HigherOrder.Flag() }             // higher order type
func (t Flag) Type() Flag      { return newFlag(HigherOrder, t.Flag()) } // higher order type
func (t Flag) Kind() BitFlag   { k, _ := t(); return Kind(k).Flag() }    // precedence type
func (t Flag) Prec() d.BitFlag { _, p := t(); return p.Flag() }          // precedence type
func (t Flag) String() string {
	var str string
	k, p := t()
	if bits.OnesCount(k.Uint()) > 1 {
		var flags = d.FlagDecompose(d.BitFlag(k))
		for i, f := range flags {
			str = str + Kind(f).String()
			if i < len(flags)-1 {
				str = str + "|"
			}
		}
	} else {
		str = Kind(k).String()
	}
	return str + ":" + d.StringBitFlag(d.BitFlag(p))
}

type FlagSet func() []Flag

func (f FlagSet) Flag() d.BitFlag { return d.HigherOrder.Flag() }
func (f FlagSet) Ident() Data     { return f }
func (f FlagSet) String() string {
	var str = "["
	var l = len(f())
	for i, flag := range f() {
		str = str + flag.String()
		if i < l-1 {
			str = str + ", "
		}
	}
	return str + "]"
}

func newFlagSet(fs ...Flag) FlagSet {
	fo := make([]Flag, 0, len(fs))
	for _, f := range fs {
		fo = append(fo, f)
	}
	return func() []Flag { return fo }
}
