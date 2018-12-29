package types

// VALUES AND TYPES
///////////////////
// propertys intendet for internal use
type Reproduceable interface{ Copy() Data }
type Destructable interface{ Clear() }
type Stringer interface{ String() string }

//// USER DEFINED DATA & FUNCTION TYPES ///////
type Typed interface{ Flag() BitFlag } //<- lowest common denominator
type Named interface{ Name() }
type Data interface {
	Typed
	Stringer
	Eval() Data
}
type Functional interface{ Data }

type Lambda interface {
	Functional
	Enclosed() Data
}
type NamedFuntion interface {
	Named
	Lambda
}

///// COLLECTION ///////
///// PROPERTYS ////////
type Collected interface {
	Data
	Empty() bool //<-- no more nil pointers & 'out of index'!
}

/// FLAT COLLECTIONS /////
// rarely used in functional programming, but nice to have whenever iterative
// performance is mandatory
type Ordered interface {
	Collected
	Next() Data
}
type Reverseable interface {
	Ordered
	Prev() Data
}
type Limited interface {
	Len() intVal // <- performs mutch better on slices
}

// collections that are accessable by other means than retrieving the 'next'
// element, according to list type, need accessors, to pass in attributes on
// which element(s) to access. attributes are a type alias of Data, to ensure
// type safety on argument propagation
type Associative interface {
	AttrType() Typed
	Get(Attribute) Data
}

////////// STACK ////////////
//// LAST IN FIRST OUT //////
type Stacked interface {
	Collected
	Push(Data)
	Pop() Data
	Add(...Data)
}

///////// QUEUE /////////////
//// FIRST IN FIRST OUT /////
type Queued interface {
	Collected
	Put(Data)
	Pull() Data
	Append(...Data)
}

/// NESTED COLLECTIONS /////
//// RECURSIVE LISTS ///////
type Recursive interface {
	Collected
	Head() Data
	Tail() Recursive
}
type Tupled interface {
	Recursive
	Arity() Arity // number of fields
}

//////// TREES ////////
type Nodular interface {
	Root() Nodular
}
type Nested interface {
	Nodular
	Members() Nodular
}
type Chained interface {
	Nodular
	Next() Nodular
}
type RevChained interface {
	Prev() Nodular
}
type Branched interface {
	Nodular
	Left() Nodular
	Right() Nodular
}
type Edged interface {
	Nodular
	Value() Data
}

//////////////////////////
// input item data interface
type Item interface {
	ItemType() Typed
	Idx() int
	Value() Data
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

func (t FnType) Type() Typed   { return t.Type() }
func (t FnType) Flag() BitFlag { return BitFlag(t) }

// function to change parameters and return the changed instance accompanied by
// the new ParamFn closing over the replaced arguments
type ParamFn func(Parametric) (Parametric, ParamFn)

func (p ParamFn) Type() Typed { return ParamFunc.Type() }

// data to parse
type Token interface {
	Type() BitFlag
	String() string
}
