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

///// COLLECTION ///////
///// PROPERTYS ////////
type Collected interface {
	Data
	Empty() bool //<-- no more nil pointers & 'out of index'!
}
type Countable interface {
	Len() int // <- performs mutch better on slices
}
type Splitable interface {
	Collected
	Countable
	Slice() []Data //<-- no more nil pointers & 'out of index'!
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
type Consumeable interface {
	Collected
	Head() Data
	Tail() Consumeable
	Shift() Consumeable
}
type Tupled interface {
	Consumeable
}

//////// TREES ////////
type Nodular interface {
	Collected
	NodeType() BitFlag
	Root() Nodular
}
type Nested interface {
	Nodular
	Member() []Nodular
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

// data to parse
type Token interface {
	Type() BitFlag
	String() string
}

type StateFn func(State) StateFn

func (p StateFn) Type() Typed { return StateFunc.Type() }

///
type Parametric interface {
	Options(Parameter) (Parametric, Parameter)
}

// function to change parameters and return the changed instance accompanied by
// the new ParamFn closing over the replaced arguments
type Parameter func(Parametric) (Parametric, Parameter)

func (p Parameter) Type() Typed { return Param.Flag() }
