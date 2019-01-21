/*
REGISTRY

  data type that holds the runtime state of the type system. Comes with helper
  functions to eanipulate chains of tokens when dealing with signatures during
  type checking, or construction.
*/
package run

import (
	d "github.com/JoergReinhardt/godeep/data"
	f "github.com/JoergReinhardt/godeep/functions"
)

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

type State f.ParamSet

// methods forwarded from parameters
func (s State) len() int                          { return f.ParamSet(s).Len() }
func (s State) accs() f.Parameters                { return f.ParamSet(s).AccSet() }
func (s State) empty() bool                       { return f.ParamSet(s).Empty() }
func (s State) append(p ...f.Paired) f.Parameters { return f.ParamSet(s).Append(p...) }
func (s State) pairs() []f.Paired                 { return f.ParamSet(s).Pairs() }
func (s State) get(acc f.Function) f.Paired       { return f.ParamSet(s).Get(acc) }
func (s State) getIdx(acc f.Function) int {
	idx, _ := f.ParamSet(s).GetIdx(acc)
	return idx
}
func (s State) apply(acc ...f.Paired) ([]f.Paired, f.Parameters) { return f.ParamSet(s).Apply(acc...) }
func (s State) replace(acc f.Paired) f.Parameters                { return f.ParamSet(s).Replace(acc) }

// genuine methods
func (s State) newUID() (int, State) {
	var _, praeds = s()
	var counter = f.ParamSet(s).Get(d.New("tuid")).Right().(idGenerator)
	var idx int
	idx, counter = counter()
	return idx, State(praeds.Replace(f.NewPair(d.New("tuid"), d.New(counter))).(f.ParamSet))
}

// initState()
//
// initializes the runtime state
func initState() State {
	return State(f.NewParameters(
		f.NewPair(d.New("tuid"), d.New(initCounter())),
		f.NewPair(d.New("types"), d.New(f.NewParameters())),
	).(f.ParamSet))
}
