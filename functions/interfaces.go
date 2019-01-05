package functions

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

type Named interface{ Name() string }
type Fixity uint8

const (
	PostFix Fixity = 0
	InFix   Fixity = 1
	PreFix  Fixity = 2
)

type Function interface { // RENAME: FunctionType
	Eval(Data) Data // of types Argumemnt & Return
}
type Instance interface {
	Function
	Flag() Flag
	Ari() int8
	Fix() Fixity
	Sig() Signature
	Fnc() []Implement
}

///// COLLECTION ///////
///// PROPERTYS ////////
type Collected interface {
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
type Accessable interface {
	AccType() // 0: int | 1: string | 3: bitflag
	Value(Data)
}
type KeyAccessable interface {
	Accessable
	Key(string) Data
}
type IdxAccessable interface {
	Accessable
	Idx(int) Data
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
type Reduceable interface {
	Collected
	Head() Data
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

// data to parse
type Token interface {
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
	Value() Data
}
