/*
TYPE FLAG

  compose id, precedence type flag and kind flag to form part of a types unique
  idetity, to first pass filter tokens by combination of propertys in a highly
  efficient manner.
*/
package functions

import (
	d "github.com/JoergReinhardt/godeep/data"
)

///// BIT_FLAG ////////
func composeHighLow(high, low BitFlag) BitFlag {
	return d.FlagHigh(BitFlag(high).Flag() | d.FlagLow(BitFlag(low)).Flag())
}

type Flag func() (uid int, kind Kind, prec d.BitFlag)

func newFlag(uid int, kind Kind, prec d.BitFlag) Flag {
	return func() (id int, k Kind, p d.BitFlag) { return uid, Kind(kind), prec }
}
func (t Flag) Flag() d.BitFlag { return HigherOrder.Flag() }             // higher order type
func (t Flag) Type() Flag      { return t }                              // higher order type
func (t Flag) UID() int        { uid, _, _ := t(); return uid }          // precedence type
func (t Flag) Kind() BitFlag   { _, k, _ := t(); return Kind(k).Flag() } // precedence type
func (t Flag) Prec() d.BitFlag { _, _, p := t(); return p.Flag() }       // precedence type

type FlagSet func() []Flag

func (f FlagSet) Flag() d.BitFlag { return d.HigherOrder.Flag() }
func (f FlagSet) Ident() Data     { return f }

func newFlagSet(fs ...Flag) FlagSet {
	fo := make([]Flag, 0, len(fs))
	for _, f := range fs {
		fo = append(fo, f)
	}
	return func() []Flag { return fo }
}
