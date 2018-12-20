package types

// VALUES AND TYPES
///////////////////
// TYPED //
// is THE interface each and every part needs to implement to even become
// comprehendible to the library types are typesafe, but also intended to be
// automagicly inferred whenever that's possible. a bitflag is providing
// composable type identificators.

// typemarkers for buildtin and user defined typed need to implement the type
// interface
type Typed interface {
	Type() flag
}

// VALUE PROPERTYS //
// interface needs to be implementd by allmost every part, even most of
// the core data types implement it. That selfreferentiality provides flexibility.
type Value interface {
	Typed
	Eval() Value
	Ref() Value
	DeRef() Value
	Copy() Value
	String() string
}

//// LIST COMPOSITIONS ////
type Collected interface {
	Voidable
	Sliced
}
type IdxCollected interface {
	OrdinalGetter
	OrdinalSetter
}
type DoubleEnded interface {
	Topped
	Bottomed
}
type Listed interface {
	IdxCollected // Get | Set
	Sliced       // Sliced
}
type MultiTypedList interface {
	AttrByType
	Listed
}

///// LINKED ELEMENTS /////
type Consumeable interface {
	Decap() (Value, Tupular)
	Head() Value
	Tail() Value
}

///// LIST BEHAVIOUR /////
type Voidable interface {
	Empty() bool
}
type Sliced interface {
	Value
	Len() int
	Values() []Value
}
type Attributeable interface {
	Attributes() []Attribute
	Values() []Value
}
type AttrBySlice interface {
	Elements() []Cellular // Cell[OrdAttr,Value]
}
type AttrByKey interface {
	Keys() []Attribute
	Fields() []Cellular // Cell[StrAttr,Value]
}
type AttrByType interface {
	Attr() []Attribute
	Members() []Cellular // Cell[StrAttr,Value]
}
type Attribute interface {
	Attr() Value
	AttrType() Typed
}
type OrdinalAttr interface {
	Idx() int
}
type StringAttr interface {
	Key() string
}
type OrdinalGetter interface {
	Get(OrdinalAttr) Value
}
type OrdinalSetter interface {
	Set(OrdinalAttr, Value)
}
type StringGetter interface {
	Get(StringAttr) Value
}
type StringSetter interface {
	Set(StringAttr, Value)
}
type Getter interface {
	Get(Attribute) Value
}
type Setter interface {
	Set(Attribute, Value)
}
type Stacked interface {
	Pull() Value
	Put(Value)
	Append(...Value)
}
type Queued interface {
	Pop() Value
	Push(Value)
	Add(...Value)
}
type Topped interface {
	First() Value
}
type Bottomed interface {
	Last() Value
}
type Referenced interface {
	HeadReferenced
	Next() Value
}
type TailReferenced interface {
	Tail() Tupular
}
type HeadReferenced interface {
	Head() Value
}
type DoublyReferenced interface {
	Reversedious() Elementar
}
type Stackable interface {
	Push(Value)
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
type Arity interface {
	Arity() int
	Unary() bool
}
type Iterable interface {
	Next() (Value, Iterable)
}

//////////////////////////
type Elementar interface {
	Voidable // aka Empty() bool
	Arity    // aka Unary() bool
	Value
}
type Cellular interface {
	Attribute // accessor attribute
	Value
}
type Tupular interface {
	Elementar
	Consumeable
}
type Nodular interface {
	Attribute
	Rooted
	Arity
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
	Value
}

//////////////////////////
// input item data interface
type Item interface {
	ItemType() Typed
	Idx() int
	Value() Value
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
