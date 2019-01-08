package functions

//////// RUNTIME TYPE SPECIFICATIONS ////////
///// UID & USER DEFINED TYPE REGISTRATION ////
// TODO: make that portable, serializeable, parallelizeable, modular,
// selfcontained, distributely executed, and all the good things. by wrapping it all in a state monad

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

var (
	uid   = genCount()
	sig   = patterns{}
	iso   = isomorphs{}  // sig & fnc
	poly  = polymorphs{} // []sig & []fnc
	names = map[string]patterns{}
)
