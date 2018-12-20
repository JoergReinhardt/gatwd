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
type typedFn Instance

func (t typedFn) Type() flag { return Typed(t()).Type() }

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

type dataFn Instance

func (t dataFn) Eval() Value    { return Value(t()).Eval() }
func (t dataFn) Ref() Value     { return Value(t()).Ref() }
func (t dataFn) DeRef() Value   { return Value(t()).DeRef() }
func (t dataFn) Copy() Value    { return Value(t()).Copy() }
func (t dataFn) String() string { return Value(t()).String() }

///// LINKED ELEMENTS /////
type Voidable interface {
	Empty() bool
}
type Chained interface {
	Value
	Len() int
	Values() []Value
}
type chainFn Instance

func (s chainFn) Len() int        { return s().(Chained).Len() }
func (s chainFn) Empty() bool     { return s().(Voidable).Empty() }
func (s chainFn) Values() []Value { return s().(Chained).Values() }

type Consumed interface {
	Decap() (Value, Tupular)
	Head() Value
	Tail() Value
}
type recursiveListFn recursiveFn
type recursiveFn Instance

func (s recursiveFn) Decap() (Value, Tupular) { return s().(Consumed).Decap() }
func (s recursiveFn) Head() Value             { return s().(Consumed).Head() }
func (s recursiveFn) Tail() Value             { return s().(Consumed).Tail() }

///// ATTRIBUTE ACCESSORS /////
type Attributed interface {
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
type IndexAt interface {
	Idx() int
}
type StringAt interface {
	Key() string
}
type IdxGet interface {
	Get(IndexAt) Value
}
type IdxSet interface {
	Set(IndexAt, Value)
}
type StrGet interface {
	Get(StringAt) Value
}
type StrSet interface {
	Set(StringAt, Value)
}
type Getter interface {
	Get(Attribute) Value
}
type Setter interface {
	Set(Attribute, Value)
}

///////////////////////////
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
	Consumed
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

//// LIST COMPOSITIONS ////
type Collected interface {
	Voidable
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
	AttrByType
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
