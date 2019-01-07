package functions

import d "github.com/JoergReinhardt/godeep/data"

//////// RUNTIME TYPE SPECIFICATIONS ////////
///// GLOBAL STATE OF THE TYPE SYSTEM ///////
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

type Flag d.BitFlag

func ComposeFlag(high, low Flag) Flag {
	return Flag(d.High(d.BitFlag(high)).Flag() | d.Low(d.BitFlag(low)).Flag())
}

func (t Flag) String() string  { return DataType(t).String() }
func (t Flag) Low() Flag       { return Flag(d.Low(d.BitFlag(t)).Flag()) }
func (t Flag) High() Flag      { return Flag(d.High(d.BitFlag(t)).Flag()) }
func (t Flag) Uint() uint      { return uint(t) }
func (t Flag) Flag() d.BitFlag { return d.BitFlag(t) }

type DataType Flag

func (t DataType) Flag() d.BitFlag { return d.BitFlag(t).Flag() }
func (t DataType) Uint() uint      { return d.BitFlag(t).Uint() }
