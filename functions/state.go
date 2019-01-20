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

// functional implementing methods
func (i idGenerator) String() string  { return "state monad" }
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

type State parameters

// methods forwarded from parameters
func (s State) len() int                      { return parameters(s).Len() }
func (s State) accs() []Parametric            { return parameters(s).Accs() }
func (s State) empty() bool                   { return parameters(s).Empty() }
func (s State) append(p ...Paired) Parameters { return parameters(s).Append(p...) }
func (s State) pairs() []Paired               { return parameters(s).Pairs() }
func (s State) get(acc Data) Paired           { return parameters(s).Get(acc) }
func (s State) getIdx(acc Data) int {
	idx, _ := parameters(s).getIdx(acc)
	return idx
}
func (s State) apply(acc ...Paired) ([]Paired, Parameters) { return parameters(s).Apply(acc...) }
func (s State) replace(acc Paired) Parameters              { return parameters(s).Replace(acc) }

// genuine methods
func (s State) newUID() (int, State) {
	var _, praeds = s()
	var counter = parameters(s).Get(d.New("tuid")).Right().(idGenerator)
	var idx int
	idx, counter = counter()
	return idx, State(praeds.Replace(newPair(d.New("tuid"), d.New(counter))).(parameters))
}

// initState()
//
// initializes the runtime state
func initState() State {
	return State(newParameters(
		newPair(d.New("tuid"), d.New(initCounter())),
		newPair(d.New("types"), d.New(newParameters())),
	).(parameters))
}
