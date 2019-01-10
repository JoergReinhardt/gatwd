package functions

import (
//	l "github.com/JoergReinhardt/godeep/lang"
)

/////////////////////////
//type StateFn func(State) StateFn
//
//func (p StateFn) Type() Typed { return StateFunc.Type() }

///
//type Parametric interface {
//	Options(paramVal) (Parametric, Parametric)
//}

import (
	d "github.com/JoergReinhardt/godeep/data"
)

type Arity int8
type Fixity int8
type Equality int8
type Lazynes bool

func (l Lazynes) String() string {
	if l {
		return "Lazy"
	}
	return "Eager"
}

type Bind bool

func (b Bind) String() string {
	if b {
		return "Left_Binding"
	}
	return "Right_Binding"
}

type Mutability bool

func (b Mutability) String() string {
	if b {
		return "Mutable"
	}
	return "Immutable"
}

type SideEffects bool

func (b SideEffects) String() string {
	if b {
		return "Side Effected"
	}
	return "Pure"
}

//go:generate stringer -type Arity
//go:generate stringer -type Fixity
//go:generate stringer -type Equality
const (
	Constant_Function Arity       = 0
	Unary_Function    Arity       = 1
	Binary_Function   Arity       = 2
	PostFix           Fixity      = -1
	InFix             Fixity      = 0
	PreFix            Fixity      = 1
	Lesser            Equality    = -1
	Equal             Equality    = 0
	Greater           Equality    = 1
	Eager             Lazynes     = false
	Lazy              Lazynes     = true
	Right_Binding     Bind        = false
	Left_Binding      Bind        = true
	Mutable           Mutability  = true
	Immutable         Mutability  = false
	Side_Effected     SideEffects = true
	Pure              SideEffects = false
)

// to be handled by the runtime, all that is defined, declared, instanciated‥.
// needs to be identifieable by unique numeric id.
type Identified interface{ Id() int }

// type definitions and variable declarations can be anonymous, or named.  In
// the latter case they need to provide the name method.
type Named interface{ Name() string }

// interface to wrap data from the data module and function module specific
// data alike
type Data interface {
	Flag() d.BitFlag
	String() string
}

// forward methods of data.BitFlag
type BitFlag interface {
	Eval() d.Data
	Flag() d.BitFlag
	String() string
	Uint() uint
	Len() int
	Count() int
	Least() int
	Most() int
	High(d.BitFlag) d.BitFlag
	Reverse() d.BitFlag
	Rotate(n int) d.BitFlag
	Toggle(f d.BitFlag) d.BitFlag
	Concat(f d.BitFlag) d.BitFlag
	Mask(f d.BitFlag) d.BitFlag
	Match(f d.BitFlag) bool
}

// least invasive, general abbreveation of a golang function in terms of
// godeeps typesystem: it can be called, optionally using no to n parameters of
// the generalized data type and returns a value, also of general data type
type Functional interface {
	Type() Flag
	Eval() Data // calls enclosed fnc, with enclosed parameters
}

// least invasive wrapper to represent a function and it's runtime parameters
// within godeeps typesystem
type Functor interface {
	Functional
	Call(...Data) Data // calls enclosed fnc, passes params & return
}

// lifts the vectorized interface from the data package onto the functions
// level to make those interchangeable.
type Vectorized interface {
	Data
	Len() int
	Empty() bool
	Slice() []Data
}
type Collected interface {
	Data
	Empty() bool //<-- no more nil pointers & 'out of index'!
}

// operators expect their parameters within syntactic context to either be
// left, right, or on both sides of the operators position.
type Operator interface {
	Functor
	Fix() Fixity
}

///// COLLECTION ///////
///// PROPERTYS ////////
type Paired interface {
	Left() Data
	Right() Data
	Both() (Data, Data)
}
type Countable interface {
	Len() int // <- performs mutch better on slices
}
type Sliceable interface {
	Data
	Countable
	Empty() bool
	Slice() []Data //<-- no more nil pointers & 'out of index'!
}
type AccessableSlice interface {
	Sliceable
	Elem(i int) Data
	Range(i, j int) []Data
}
type SliceOfNatives interface {
	AccessableSlice
	Native(i int) interface{}
	Natives(i, j int) []interface{}
}

/// FLAT COLLECTIONS /////
// rarely used in functional programming, but nice to have whenever iterative
// performance is mandatory
type Ordered interface {
	Collected
	Next() value
}
type Reverseable interface {
	Ordered
	Prev() value
}

// collections that are accessable by other means than retrieving the 'next'
// element, according to list type, need accessors, to pass in attributes on
// which element(s) to access. attributes are a type alias of Data, to ensure
// type safety on argument propagation
type Accessable interface {
	AccType() // 0: int | 1: string | 3: bitflag
	Value(value)
}
type KeyAccessable interface {
	Accessable
	Key(string) value
}
type IdxAccessable interface {
	Accessable
	Idx(int) value
}

////////// STACK ////////////
//// LAST IN FIRST OUT //////
type Stacked interface {
	Collected
	Push(value)
	Pop() value
	Add(...value)
}

///////// QUEUE /////////////
//// FIRST IN FIRST OUT /////
type Queued interface {
	Collected
	Put(value)
	Pull() value
	Append(...value)
}

/// NESTED COLLECTIONS /////
//// RECURSIVE LISTS ///////
type Reduceable interface {
	Collected
	Head() value
	Tail() Reduceable
	Shift() Reduceable
}
type Tupled interface {
	Reduceable
}

//////////////////////////
// input item data interface
type Item interface {
	ItemType() d.BitFlag
	Idx() int
	Value() value
}

//////////////////////////
// interfaces dealing with instances of input items
type Queue interface {
	Next()
	Peek() Item
	Current() Item
	Idx() int
}

// data to parse
type Token interface {
	Type() TokType
	Flag() d.BitFlag
	String() string
}

//////// TREES ////////
type Nodular interface {
	Collected
	NodeType() Flag
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
	Value() value
}
