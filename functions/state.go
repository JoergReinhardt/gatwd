/*
REGISTRY

  data type that holds the runtime state of the type system. Comes with helper
  functions to eanipulate chains of tokens when dealing with signatures during
  type checking, or construction.
*/
package functions

import d "github.com/JoergReinhardt/godeep/data"

/////////////////////////////////////////////////////////////////////
type idGenerator func() (int, idGenerator)

func (i idGenerator) String() string  { return "UID()" }
func (u idGenerator) Flag() d.BitFlag { return d.Function.Flag() }

func initCounter() idGenerator {
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

type State praedicates

func (s State) len() int                                  { return praedicates(s).Len() }
func (s State) accs() []Parametric                        { return praedicates(s).Accs() }
func (s State) pairs() []Paired                           { return praedicates(s).Pairs() }
func (s State) apply(p ...Paired) ([]Paired, Praedicates) { return praedicates(s).Apply(p...) }
func (s State) NewUID() (int, State) {
	var _, praeds = s()
	var counter = praedicates(s).Get(d.New("UID")).Right().(idGenerator)
	var idx int
	idx, counter = counter()
	return idx, State(praeds.Replace(newPair(d.New("UID"), d.New(counter))).(praedicates))
}

// initState()
//
// initializes the runtime state
func initState() State {
	return State(newPraedicates(
		newPair(d.New("UID"), d.New(initCounter())),
	).(praedicates))
}
