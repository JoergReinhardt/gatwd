package types

// VALUES AND TYPES
///////////////////
// TYPED //
type Reproduceable interface{ Copy() Evaluable }
type Printable interface{ String() string }
type Named interface{ Name() string }
type Typed interface{ Type() flag }

type Evaluable interface {
	Reproduceable
	Printable
	Typed
	Eval() Evaluable
	ref() Evaluable
}
type Callable interface {
	Arity() int
}
type Applyable interface {
	Fix() int
}

///// LINEAR ELEMENTS /////
type Voidable interface {
	Empty() bool
}
type Size interface {
	Len() int
}
type Chained interface {
	Size
	Voidable
	Values() []Evaluable
}

// one of the very few exceptions to the 'everything is an expression' rule
type Destructable interface {
	Clear()
}

//// RECURSIVE ELEMENTS //////
// designated 'main mode of transportation' in the world of ffp
type Consumed interface {
	Cellular
	Decap() (Evaluable, Tupular)
	Head() Evaluable
	Tail() Evaluable
}

///// ATTRIBUTATION /////
type Attributed interface {
	Attributes() []Attribute
	Values() []Evaluable
}
type Attribute interface {
	Attr() Evaluable
	AttrType() Typed
}
type AttrBySlice interface {
	Elements() []Cellular // Cell[OrdAttr,Value]
}
type AttrByKey interface {
	Keys() []Attribute
	Fields() []Cellular // Cell[StrAttr,Value]
}
type AttrByMembership interface {
	Attr() []Attribute
	Members() []Cellular // Cell[StrAttr,Value]
}

/////
type Any interface {
	Any(func(a Attribute, v Evaluable) bool) bool
}
type All interface {
	All(func(a Attribute, v Evaluable) bool) bool
}
type Membership interface {
	Match() bool
}

///// COLLECTION ACCESSORS //////
type IndexAt interface {
	Idx() int
}
type StringAt interface {
	Key() string
}
type IdxGet interface {
	Get(IndexAt) Evaluable
}
type IdxSet interface {
	Set(IndexAt, Evaluable)
}
type StrGet interface {
	Get(StringAt) Evaluable
}
type StrSet interface {
	Set(StringAt, Evaluable)
}
type Getter interface {
	Get(Attribute) Evaluable
}
type Setter interface {
	Set(Attribute, Evaluable)
}

///////////////////////////
type Stacked interface {
	Pull() Evaluable
	Put(Evaluable)
	Append(...Evaluable)
}
type Queued interface {
	Pop() Evaluable
	Push(Evaluable)
	Add(...Evaluable)
}
type Topped interface {
	First() Evaluable
}
type Bottomed interface {
	Last() Evaluable
}
type Referenced interface {
	HeadReferenced
	Next() Evaluable
}
type TailReferenced interface {
	Tail() Tupular
}
type HeadReferenced interface {
	Head() Evaluable
}
type DoublyReferenced interface {
	Reverse() Evaluable
}
type Stackable interface {
	Push(Evaluable)
	Pop()
}
type Rooted interface {
	Root() Nodular
}
type Parented interface {
	Parent() Nodular
}
type Nested interface {
	Nest() []Nodular
}
type Identifyable interface {
	Ident() Attribute
}
type Iterable interface {
	Next() (Evaluable, Iterable)
}

//////////////////////////
type Cellular interface {
	Evaluable
	Voidable // aka Empty() bool
}
type AttributedCell interface {
	Cellular
	Attribute
}
type Tupular interface {
	Consumed
}
type Nodular interface {
	Attribute
	Rooted
}
type NestNodul interface {
	Nodular
	Nested
}
type ChildNodul interface {
	Parented
	Nodular
}
type BranchNodul interface {
	ChildNodul
	Left() Nodular
	Right() Nodular
}
type LeaveNodule interface {
	ChildNodul
	Evaluable
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

//// LIST COMPOSITIONS ////
type Collected interface {
	Chained
}
type IdxCollected interface {
	IdxGet
	IdxSet
}
type DoubleEnded interface {
	Topped
	Bottomed
}
type Listed interface {
	IdxCollected // Get | Set
	Chained      // Sliced
}
type MultiTypedList interface {
	AttrByMembership
	Listed
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
