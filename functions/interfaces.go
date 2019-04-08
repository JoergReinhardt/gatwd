package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

//// PRIMITIVE TYPE CLASS INTERFACES
type Nullable interface {
	d.Native
	d.Nullable
}
type Discrete interface {
	d.Native
	d.Discrete
}
type Boolean interface {
	d.Native
	d.Boolean
}
type Natural interface {
	d.Natural
}
type Integer interface {
	d.Native
	d.Integer
}
type Rational interface {
	d.Native
	d.Rational
}
type Real interface {
	d.Native
	d.Real
}
type Imaginary interface {
	d.Native
	d.Imaginary
}
type Number interface {
	d.Native
	d.Number
}
type Letter interface {
	d.Native
	d.Letter
}
type Text interface {
	d.Text
}
type Printable interface {
	d.Printable
}
type PairedNative interface {
	d.Paired
}
type ComposedNative interface {
	d.Composed
}
type SliceNative interface {
	d.Sliceable
}
type SequentialNative interface {
	d.Sequential
}
type MappedNatives interface {
	d.Sequential
}

//// FUNCTIONAL CLASS
type Parametric interface {
	d.Native
	TypeFnc() TyFnc
	Call(...Parametric) Parametric
}

//// PAIRS OF FUNCTIONALS
type Paired interface {
	Resourceful
	Left() Parametric
	Right() Parametric
	Both() (Parametric, Parametric)
}

//// COLLECTION CLASSES
type Optional interface {
	Parametric
	Maybe() bool
	Value() Parametric
}
type Predictable interface {
	Parametric
	True(Parametric) bool
	Any(...Parametric) bool
	All(...Parametric) bool
}
type Distinguishable interface {
	Parametric
	Case(expr ...Parametric) Parametric
}
type Composed interface {
	Empty() bool //<-- no more nil pointers & 'out of index'!
}

type Countable interface {
	Len() int // <- performs mutch better on slices
}

type Sequential interface {
	Slice() []Parametric //<-- no more nil pointers & 'out of index'!
}

type Linked interface {
	Next() Parametric
}

type Reverseable interface {
	Prev() Parametric
}

type Ordered interface {
	Sort(d.TyNative)
}

type Searchable interface {
	Search(Parametric) int
}

type Indexed interface {
	Get(int) Parametric
	Set(int, Parametric) Vectorized
}

type Resourceful interface {
	Parametric
}

type Generating interface {
	Parametric
	Next() Optional
}

type Aggregating interface {
	Parametric
	Result() Parametric
	Aggregator() NaryFnc
	Aggregate(...Parametric) Parametric
}
type Monadic interface {
	Parametric
}

type Associative interface {
	Parametric
	AccFncType() TyFnc
	AccNatType() d.TyNative
	GetVal(Parametric) Paired
	SetVal(Parametric, Parametric) Associative
	Pairs() []Paired
}

//// CONSUMEABLE COLLECTION ///////
///
// implemented by types backed by recursive lists. consumeable is the
// behaviour map-/ & fold operations rely up on
type Consumeable interface {
	Parametric
	Head() Parametric
	Tail() Consumeable
	DeCap() (Parametric, Consumeable)
}

//// SEQUENTIAL LIST //////
///
// implemented by types backed by a slice. sequential interface implements all
// collection type interfaces. map & fold operators rely on the consumeable
// type interface. vectorized types implement that behaviour
type Vectorized interface {
	Parametric
	Sequential
	Searchable
	Ordered
	Indexed
}

/// INSTANCES
type Instanciated interface {
	Uid() uint
}

/// ITEMS & TOKENS
// data to parse
type Token interface {
	d.Native
	TypeTok() TyToken
	Data() d.Native
}

//// TREES & GRAPHS
///
type Nodular interface {
	Sequential
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
	From() Nodular
	To() Nodular
}

type Leaved interface {
	Nodular
	Value() Parametric
}
