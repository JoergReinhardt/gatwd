package functions

import d "github.com/JoergReinhardt/godeep/data"

///// BIT_FLAG ////////
func composeHighLow(high, low d.BitFlag) d.BitFlag {
	return d.High(d.BitFlag(high).Flag() | d.Low(d.BitFlag(low)).Flag())
}

type Flag func() (tid int, kind Kind, prec d.BitFlag)

func conFlag(kind, prec d.BitFlag) Flag {
	return func() (t int, h Kind, l d.BitFlag) { tid := 0; return tid, Kind(kind), prec }
}
func (t Flag) Kind() Flag      { return t }                              // higher order type
func (t Flag) TypeId() int     { i, _, _ := t(); return i }              // precedence type
func (t Flag) Flag() d.BitFlag { _, _, l := t(); return Kind(l).Flag() } // precedence type
func (t Flag) String() string {
	_, kind, prec := t()
	return Kind(kind).String() + "||" + d.Type(prec).String()
}
