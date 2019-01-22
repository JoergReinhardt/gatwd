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

// private methods forwarded from parameters
func (s State) len() int                                         { return f.ParamSet(s).Len() }
func (s State) accs() f.Parameters                               { return f.ParamSet(s).AccSet() }
func (s State) empty() bool                                      { return f.ParamSet(s).Empty() }
func (s State) append(p ...f.Paired) f.Parameters                { return f.ParamSet(s).Append(p...) }
func (s State) pairs() []f.Paired                                { return f.ParamSet(s).Pairs() }
func (s State) get(acc f.Functional) f.Paired                    { return f.ParamSet(s).Get(acc) }
func (s State) replace(acc f.Paired) f.Parameters                { return f.ParamSet(s).Replace(acc) }
func (s State) apply(acc ...f.Paired) ([]f.Paired, f.Parameters) { return f.ParamSet(s).Apply(acc...) }
func (s State) getIdx(acc f.Functional) int {
	idx, _ := f.ParamSet(s).GetIdx(acc)
	return idx
}

// newUID() (int, State)
// generates unique id by continously counting up. counter closure is stored in
// the state monad
func (s State) newUID() (int, State) {
	var _, praeds = s()
	var counter = f.ParamSet(s).Get(d.New("tuid")).Right().(idGenerator)
	var idx int
	idx, counter = counter()
	return idx, State(praeds.Replace(f.NewPair(d.New("tuid"), d.New(counter))).(f.ParamSet))
}

// initState()
//
// initializes the runtime state with fresh uid generator type system & ast‥.
// recursive types are defined using parametric type constructors.
// implementation relies on references to the type id/name of the presented
// type during runtime, while the list of definitions referenced to, is just a
// flat slice.
//
// the ast is a recursively nested data structure, presenting the current calls
// context as head of a pair. all other nodes are presented in a slice that can
// be accessed by nodes slice index. the context may keep references to data
// and/or parent, member, or root nodes depending on the current nodes type and
// mode of operation.
func initState() State {
	var state = State(f.NewParameters(
		f.NewPair(d.New("tuid"), d.New(initCounter())),  // uid generator
		f.NewPair(d.New("typeNames"), d.NewStringSet()), // set of type definition names
		f.NewPair(d.New("funcNames"), d.NewStringSet()), // set of function definition names
		f.NewPair(d.New("definitions"), d.NewSlice()),   // all equations name/uid grouped (t:p = 1:n)
		f.NewPair(d.New("next"), f.NewParameters()),     // current calling context (aka current node)
		f.NewPair(d.New("ctx"), f.NewParameters()),      // current calling context (aka current node)
		f.NewPair(d.New("tree"), d.NewSlice()),          // slice holding all other nodes (may be referenced by ctx parameters)
	).(f.ParamSet))

	return state
}
