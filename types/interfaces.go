package types

// VALUES AND TYPES
///////////////////
// TYPED //
type Reproduceable interface{ Copy() Evaluable }
type Printable interface{ String() string }
type Named interface{ Name() string }
type Typed interface{ Type() flag }
type Destructable interface{ Clear() }

///// FLAT VALUES ///////
type Evaluable interface {
	Reproduceable
	Printable
	Typed
	Eval() Evaluable
}

///// COLLECTIONS ///////
type Collected interface {
	Evaluable
	Empty() bool
}

/// FLAT COLLECTIONS /////
type Ordered interface {
	Collected
	Next() Evaluable
}
type Reverseable interface {
	Ordered
	Prev() Evaluable
}
type Limited interface {
	Len() int
}
type Attributed interface {
	AttrType() Typed
	Get(Evaluable) Evaluable
}

//// LAST IN FIRST OUT //////
type Stacked interface {
	Push(Evaluable)
	Pop() Evaluable
	Add(...Evaluable)
}

//// FIRST IN FIRST OUT /////
type Queued interface {
	Put(Evaluable)
	Pull() Evaluable
	Append(...Evaluable)
}

/// NESTED COLLECTIONS /////
type Nested interface {
	Decap() (Evaluable, Nested)
	Head() Evaluable
	Tail() Nested
}

///// FUNCTIONS /////////
type Parametric interface {
	Arity() int
	Fix() int
}

//////////////////////////
// input item data interface
type Item interface {
	ItemType() Typed
	Idx() int
	Value() Evaluable
}

//////////////////////////
// interfaces dealing with instances of input items
type Queue interface {
	Next()
	Peek() Item
	Current() Item
	Idx() int
}

///////////////////////
type State interface {
	Queue
	Run()
	ItemType() Typed
	State(string) StateFn
}

// takes a state and advances it. returns the next state fn to run
type StateFn func(State) StateFn

func (p StateFn) Type() Typed { return StateFunc.Type() }

// a thing that can be changed by calling the args method passing a parameter
// function. the parameter function gets called with the parametric instance as
// it's argument and returns an altered version of the instance and a new
// parameter function that closes over the old set of parameters to enable a
// possible rollback to the former instance state.
type Parametric interface {
	Params(ParamFn) (Parametric, ParamFn)
}

//go:generate stringer -type=FnType
const (
	StateFunc FnType = 1 << iota
	ParamFunc
)

type FnType uint

func (t FnType) Type() Typed { return t.Type() }
func (t FnType) Flag() flag  { return flag(t) }

// function to change parameters and return the changed instance accompanied by
// the new ParamFn closing over the replaced arguments
type ParamFn func(Parametric) (Parametric, ParamFn)

func (p ParamFn) Type() Typed { return ParamFunc.Type() }

// data to parse
type Token interface {
	Type() TokenType
	String() string
}
