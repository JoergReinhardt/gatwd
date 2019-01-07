package functions

import d "github.com/JoergReinhardt/godeep/data"

//////// RUNTIME TYPE SPECIFICATIONS ////////
///// UID & USER DEFINED TYPE REGISTRATION ////
// TODO: make that portable, serializeable, parallelizeable, modular,
// selfcontained, distributely executed, and all the good things. by wrapping it all in a state monad
var (
	uid   = genCount()
	sig   = signatures{}
	iso   = isomorphs{}  // sig & fnc
	poly  = polymorphs{} // []sig & []fnc
	names = map[string]polymorph{}
)

type idGenerator func() (int, idGenerator)

func genCount() idGenerator {
	return func() (int, idGenerator) {
		var id int
		var gen idGenerator
		gen = func() (int, idGenerator) {
			id = id + 1
			return id, gen
		}
		return id, gen
	}
}

func conUID() int { var id int; id, uid = uid(); return id }

///// BIT_FLAG ////////
func composeHighLow(high, low d.BitFlag) d.BitFlag {
	return d.High(d.BitFlag(high).Flag() | d.Low(d.BitFlag(low)).Flag())
}

type Flag func() (high DataType, prec d.BitFlag)

func conFlag(high, prec d.BitFlag) Flag {
	return func() (h DataType, l d.BitFlag) { return DataType(high), prec }
}
func (t Flag) Type() Flag      { return t }                               // higher order type
func (t Flag) Flag() d.BitFlag { _, l := t(); return DataType(l).Flag() } // precedence type
func (t Flag) String() string {
	high, prec := t()
	return DataType(high).String() + "||" + d.Type(prec).String()
}
