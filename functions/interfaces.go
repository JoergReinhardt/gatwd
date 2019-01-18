/*
INTERFACES

  interfaces provide connection between higher order function composition and
  actual implementation, as well as transition between data and function package
  types
*/
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
	"math/big"
	"time"

	d "github.com/JoergReinhardt/godeep/data"
)

// exposes all metods of data.BitFlag on module layer. the  methods of
// data.BitFlag
// TODO: use the force luke‥. use it you goddamn'ed moron!!!
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

type Arity int8
type Fixity int8
type Equality int8
type Evaluation bool

func (l Evaluation) String() string {
	if l {
		return "Lazy"
	}
	return "Eager"
}

type Boundness bool

func (b Boundness) String() string {
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

type SideEffectedness bool

func (b SideEffectedness) String() string {
	if b {
		return "Side Effected"
	}
	return "Pure"
}

//go:generate stringer -type Arity
//go:generate stringer -type Fixity
//go:generate stringer -type Equality
const (
	Constant_Function Arity            = 0
	Unary_Function    Arity            = 1
	Binary_Function   Arity            = 2
	PostFix           Fixity           = -1
	InFix             Fixity           = 0
	PreFix            Fixity           = 1
	Lesser            Equality         = -1
	Equal             Equality         = 0
	Greater           Equality         = 1
	Eager             Evaluation       = false
	Lazy              Evaluation       = true
	Right_Bound       Boundness        = false
	Left_Bound        Boundness        = true
	Mutable           Mutability       = true
	Immutable         Mutability       = false
	Side_Effected     SideEffectedness = true
	Pure              SideEffectedness = false
)

// interface to integrate data instanciated by the data module with functional
// types instanciated by the function module.
// TODO: find more elegant way than loop to convert slices
type Data interface {
	Flag() d.BitFlag
	String() string
}
type Paired interface {
	Data
	Left() Data
	Right() Data
	Both() (Data, Data)
}

type Typed interface{ Type() Flag }

type Functional interface {
	Data
	Typed
	Eval() Data // calls enclosed fnc, with enclosed parameters
}
type Functor interface {
	Functional
	Call(...Data) Data // calls enclosed fnc, passes params & return
}
type Countable interface {
	Len() int // <- performs mutch better on slices
}
type Collected interface {
	Data
	Empty() bool //<-- no more nil pointers & 'out of index'!
}
type IndexedSlice interface {
	Quantified
	Elem(i int) Data
	Range(i, j int) []Data
}
type SliceOfNatives interface {
	IndexedSlice
	Native(i int) interface{}
	Natives(i, j int) []interface{}
}

// map data packages type classes, defined by binary flag composistion, to
// method sets to be implemented by higher order types.
type Nullable interface{ Null() Data }
type Numeric interface {
	Uint() uint
	Int() int
	Flt() float64
	Imag() complex128
	BitWise() uint
	Dura() time.Duration
}
type Synbolic interface {
	String() string
	Bytes() []byte
	Time() time.Time
}
type Unsigned interface{ Uint() uint }
type Integer interface{ Int() int }
type Rational interface{ Rat() *big.Rat }
type Irrational interface{ Float() float64 }
type Imaginary interface{ Imag() complex128 }
type BinaryData interface{ Bytes() []byte }
type Symbolic interface{ String() string }
type BitWise interface{ Bytes() uint }
type Temporal interface {
	Time() time.Time
	Dura() time.Duration
}
type Collection interface{ Len() int }

type Argumented interface {
	Data
	Typed
	Arg() Argumented
	Data() Data
	Set(...Data) (Data, Argumented)
}
type Arguments interface {
	Args() []Argumented
	Set(...Argumented) ([]Argumented, Arguments)
}
type Parametric interface {
	Paired
	Acc() Parametric
	Arg() Argumented
	Key() Data
	Data() Data
	Pair() Paired
	Set(...Paired) (Paired, Parametric)
}
type Preadicates interface {
	Accs() []Parametric
	Pairs() []Paired
	Set(...Paired) ([]Paired, Preadicates)
}
type Quantified interface {
	Functional
	Countable
	Empty() bool
	Slice() []Data //<-- no more nil pointers & 'out of index'!
}
type Vectorized interface {
	Quantified
	Head() Data
	Tail() []Data
	DeCap() (Data, []Data)
}

//// TUPLES /////
type Tupled interface {
	Vectorized
	Arity() Arity
	Sig() []d.BitFlag
}

//// RECURSIVE LISTS ///////
type Recursive interface {
	Functional
	Countable
	Empty() bool
	Head() Data
	Tail() Recursive
	DeCap() (Data, Recursive)
}

// LINKED LISTS
type Ordered interface {
	Collected
	Next() (Data, Ordered)
}
type Reverseable interface {
	Ordered
	Prev() value
}

/// NESTED COLLECTIONS /////
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

// to be handled by the runtime, all that is defined, declared, instanciated‥.
// needs to be identifieable by unique numeric id.
type Identified interface{ Id() int }

// type definitions and variable declarations can be anonymous, or named.  In
// the latter case they need to provide the name method.
type Named interface{ Name() string }

// operators expect their parameters within syntactic context to either be
// left, right, or on both sides of the operators position.
type Operator interface {
	Functor
	Fix() Fixity
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
	Type() d.BitFlag
	Flag() d.BitFlag
	String() string
}
