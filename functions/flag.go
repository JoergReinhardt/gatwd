package functions

import (
	d "github.com/JoergReinhardt/godeep/data"
)

///// BIT_FLAG ////////
func composeHighLow(high, low BitFlag) BitFlag {
	return d.FlagHigh(BitFlag(high).Flag() | d.FlagLow(BitFlag(low)).Flag())
}

type Flag func() (kind Kind, prec BitFlag)

func newFlag(kind Kind, prec BitFlag) Flag {
	return func() (k Kind, p BitFlag) { return Kind(kind), prec }
}
func (t Flag) Flag() d.BitFlag { return d.Flag.Flag() | t.Prec() }         // higher order type
func (t Flag) Type() Flag      { return newFlag(Internal, d.Flag.Flag()) } // higher order type
func (t Flag) Kind() BitFlag   { k, _ := t(); return Kind(k).Flag() }      // precedence type
func (t Flag) Prec() d.BitFlag { _, p := t(); return p.Flag() }            // precedence type
func (t Flag) String() string {
	kind, prec := t()
	return Kind(kind).String() + "||" + prec.String()
}

type FlagSet func() []Flag

func newFlagSet(fs ...Flag) FlagSet {
	fo := make([]Flag, 0, len(fs))
	for _, f := range fs {
		fo = append(fo, f)
	}
	return func() []Flag { return fo }
}
