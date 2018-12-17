package types

// VALUES AND TYPES
///////////////////
// TYPED //
// is THE interface each and every part needs to implement to even become
// comprehendible to the library types are typesafe, but also intended to be
// automagicly inferred whenever that's possible. a bitflag is providing
// composable type identificators.
type TypeFlag uint

type Typed interface {
	Type() Type
}

// typemarkers for buildtin and user defined typed need to implement the type
// interface
type Type interface {
	Uint() uint
	String() string
	Type() Type
	Flag() Flag
	Concat(Flag) Flag
	Toggle(Flag) Flag
	Mask(Flag) Flag
	Match(Flag) bool
}

// VALUE PROPERTYS //
// interface needs to be implementd by allmost every part, even most of
// the core data types implement it. That selfreferentiality provides flexibility.
type Value interface {
	Typed
	Value() Value
	Ref() interface{}
	String() string
	Copy() Value
}

///// LIST BEHAVIOUR /////
type Size interface {
	Len() int
}
type Array interface {
	Slice() []Value
}
type LiFo interface {
	Pull() Value
	Put(Value)
	Append(...Value)
}
type FiFo interface {
	Pop() Value
	Push(Value)
	Add(...Value)
}
type First interface {
	First() Value
}
type Last interface {
	Last() Value
}

//// LIST COMPOSITIONS ////
type Accessible interface {
	Getter
	Setter
}
type DoubleEnded interface {
	First
	Last
}
type Listed interface {
	Accessible
	DoubleEnded
	Array
	Size
}

///// LINKED ELEMENTS /////
type Consumeable interface {
	Decap() (Value, Tupled)
}
type TailLinked interface {
	Tail() Tupled
}
type HeadLinked interface {
	Head() Value
}
type DoublyLinked interface {
	Previous() Element
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
}
type Void interface {
	Empty() bool
}
type Attribute interface {
	Attr() Value
	AttrType() Type
}
type Ordered interface {
	Idx() int
}
type Mapped interface {
	Key() string
}
type IdenAcc interface {
	Identity() string
}
type Getter interface {
	Get(Attribute) Value
}
type Setter interface {
	Set(Attribute, Value)
}
type Iterable interface {
	Next() (Value, Iterable)
}

//////////////////////////
type Element interface {
	Value
	Arity // aka Unary() bool
	Void  // aka Empty() bool
}
type Cell interface {
	Attribute
	Element
}
type Tupled interface {
	Element
	HeadLinked
	Consumeable
	TailLinked
}
type OrderedTup interface {
	Ordered
	Tupled
	Size
}
type MappedTup interface {
	OrderedTup
	Mapped
}
type TupleNest interface {
	Tupled
}
type Nodular interface {
	Attribute
	Rooted
	Arity
}
type NestNodul interface {
	Nodular
	Nested
	Size
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
	ItemType() Type
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
	ItemType() Type
	State(string) StateFn
}

// takes a state and advances it. returns the next state fn to run
type StateFn func(State) StateFn

func (p StateFn) Type() Type { return StateFunc.Type() }

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

type FnType Flag

func (t FnType) Type() Type { return t.Type() }
func (t FnType) Flag() Flag { return Flag(t) }

// function to change parameters and return the changed instance accompanied by
// the new ParamFn closing over the replaced arguments
type ParamFn func(Parametric) (Parametric, ParamFn)

func (p ParamFn) Type() Type { return ParamFunc.Type() }
