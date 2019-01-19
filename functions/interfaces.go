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

type Arity uint8

func (a Arity) Flag() d.BitFlag { return d.BitFlag(a) }

//go:generate stringer -type Arity
const (
	Nullary Arity = 0 + iota
	Unary
	Binary
	Ternary
	Quaternary
	Quinary
	Senary
	Septenary
	Octonary
	Novenary
	Denary
)

type Property d.BitFlag

func (p Property) Flag() d.BitFlag { return d.BitFlag(p) }

//go:generate stringer -type Property
const (
	PostFix Property = 1
	InFix   Property = 1 << iota
	PreFix
	///
	Eager
	Lazy
	///
	Right_Bound
	Left_Bound
	///
	Mutable
	Imutable
	///
	Effected
	Pure
	////
	Positional
	AccArg
	////
	Lesser
	Equal
	Greater
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
	Ident() Data // calls enclosed fnc, with enclosed parameters
}
type Function interface {
	Functional
	Call(...Data) Data // calls enclosed fnc, passes params & return
}
type FncDef interface {
	Data
	UID() int
	Arity() Arity
	Fix() Property
	Lazy() Property
	Bound() Property
	Mutable() Property
	Pure() Property
	ArgProp() Property
	ArgTypes() []Flag
	RetType() Flag
	Fnc() Function
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
	Arg() Argumented
	Data() Data
	Apply(...Data) (Data, Argumented)
}
type Arguments interface {
	Data
	Len() int
	Args() []Argumented
	Data() []Data
	Get(int) Argumented
	Replace(int, Data) Arguments
	Apply(...Data) ([]Data, Arguments)
}
type Parametric interface {
	Paired
	Acc() Parametric
	Arg() Argumented
	Key() Data
	Data() Data
	Pair() Paired
	Apply(...Paired) (Paired, Parametric)
}
type Praedicates interface {
	Data
	Len() int
	Accs() []Parametric
	Pairs() []Paired
	Get(Data) Paired
	Replace(Paired) Praedicates
	Apply(...Paired) ([]Paired, Praedicates)
}
type Quantified interface {
	Data
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
	Flags() []d.BitFlag
}
type Recorded interface {
	Vectorized
	Tuple() Tupled
	Arity() Arity
	ArgSig() []Paired // key/data-type
}

//// RECURSIVE LISTS ///////
type Recursive interface {
	Data
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
	Function
	Fix() Property
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
