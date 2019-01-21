/*
INTERFACES

  interfaces provide connection between higher order function composition and
  actual implementation, as well as transition between data and function package
  types
*/
package functions

import (
	"math/big"
	"time"

	d "github.com/JoergReinhardt/godeep/data"
)

// exposes all metods of data.BitFlag on module layer. the  methods of
// data.BitFlag
// TODO: use the force luke‥. use it you goddamn'ed moron!!!
type BitFlag interface{ d.Typed }
type Function interface{ d.Data }
type Functional interface{ Call(...Function) Function }

// these interfaces reflect type class flag concatenations from the data
// package and provide a least common denominator for all members of such
// classes, to later define things like equality, operator overwrite‥.
type Nullable interface{ Null() Function }
type Bitwise interface{ Uint() uint }
type Boolean interface{ Bool() bool }
type Unsigned interface{ Uint() uint }
type Signed interface{ Int() int }
type Integer interface{ Int() int }
type Rational interface{ Rat() *big.Rat }
type Irrational interface{ Float() float64 }
type Imaginary interface{ Imag() complex128 }
type Timed interface{ Time() time.Time }
type Temporal interface{ Dura() time.Duration }
type Symbolic interface{ String() string }
type Collection interface{ Len() int }
type Numeric interface {
	Uint() uint
	Int() int
	Flt() float64
}
type Synbolic interface {
	String() string
	Bytes() []byte
}

// interface to integrate data instanciated by the data module with functional
// types instanciated by the function module.
// TODO: find more elegant way than loop to convert slices
type Paired interface {
	Function
	Left() Function
	Right() Function
	Both() (Function, Function)
}

type Typed interface {
	Name() string
}

type Ident interface {
	Function
	Ident() Function // calls enclosed fnc, with enclosed parameters
}
type Countable interface {
	Len() int // <- performs mutch better on slices
}
type Collected interface {
	Function
	Empty() bool //<-- no more nil pointers & 'out of index'!
}
type IndexedSlice interface {
	Quantified
	Elem(i int) Function
	Range(i, j int) []Function
}
type SliceOfNatives interface {
	IndexedSlice
	Native(i int) interface{}
	Natives(i, j int) []interface{}
}

// map data packages type classes, defined by binary flag composistion, to
// method sets to be implemented by higher order types.

type Argumented interface {
	Function
	Arg() Argumented
	Data() Function
	Apply(...Function) (Function, Argumented)
}
type Arguments interface {
	Function
	Len() int
	Args() []Argumented
	Data() []Function
	Get(int) Argumented
	Replace(int, Function) Arguments
	Apply(...Function) ([]Function, Arguments)
}
type Parameters interface {
	Function
	Len() int
	Pairs() []Paired
	Get(Function) Paired
	Replace(Paired) Parameters
	Apply(...Paired) ([]Paired, Parameters)
}
type Parametric interface {
	Paired
	Accs() Parametric
	Arg() Argumented
	Key() Function
	Data() Function
	Pair() Paired
	Apply(...Paired) (Paired, Parametric)
}
type Quantified interface {
	Function
	Countable
	Empty() bool
	Slice() []Function //<-- no more nil pointers & 'out of index'!
}
type Vectorized interface {
	Quantified
	Head() Function
	Tail() []Function
	DeCap() (Function, []Function)
}

//// TUPLES /////
type Tupled interface {
	Vectorized
	Flags() []d.BitFlag
}
type Recorded interface {
	Vectorized
	Tuple() Tupled
	ArgSig() []Paired // key/data-type
}

//// RECURSIVE LISTS ///////
type Recursive interface {
	Function
	Countable
	Empty() bool
	Head() Function
	Tail() Recursive
	DeCap() (Function, Recursive)
}

// LINKED LISTS
type Ordered interface {
	Collected
	Next() (Function, Ordered)
}
type Reverseable interface {
	Ordered
	Prev() Function
}

/// NESTED COLLECTIONS /////
//////// TREES ////////
type Nodular interface {
	Collected
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
	Value() Function
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
	Functional
}
