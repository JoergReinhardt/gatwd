package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

//// PRIMITIVE TYPE CLASS INTERFACES
type Nullable interface {
	d.Nullable
}
type Bitwise interface {
	d.Binary
}
type Boolean interface {
	d.Boolean
}
type Natural interface {
	d.Natural
}
type Integer interface {
	d.Integer
}
type Rational interface {
	d.Rational
}
type Real interface {
	d.Real
}
type Imaginary interface {
	d.Imaginary
}
type Number interface {
	d.Number
}
type Letter interface {
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
type Functional interface {
	d.Native
	TypeFnc() TyFnc
	Call(...Functional) Functional
}

//// PAIRS OF FUNCTIONALS
type Paired interface {
	Functional
	Left() Functional
	Right() Functional
	Both() (Functional, Functional)
}

//// COLLECTION CLASSES
type Optional interface {
	Functional
	Maybe() bool
	Value() Functional
}

type Composed interface {
	Empty() bool //<-- no more nil pointers & 'out of index'!
}

type Countable interface {
	Len() int // <- performs mutch better on slices
}

type Sequential interface {
	Slice() []Functional //<-- no more nil pointers & 'out of index'!
}

type Linked interface {
	Next() Functional
}

type Reverseable interface {
	Prev() Functional
}

type Ordered interface {
	Sort(d.TyNative)
}

type Searchable interface {
	Search(Functional) int
}

type Indexed interface {
	Get(int) Functional
	Set(int, Functional) Vectorized
}

type Associative interface {
	Functional
	GetVal(Functional) Paired
	SetVal(Functional, Functional) Associative
	Pairs() []Paired
}

//// CONSUMEABLE COLLECTION ///////
///
// implemented by types backed by recursive lists. consumeable is the
// behaviour map-/ & fold operations rely up on
type Consumeable interface {
	Functional
	Composed
	Head() Functional
	Tail() Consumeable
	DeCap() (Functional, Consumeable)
}

//// SEQUENTIAL LIST //////
///
// implemented by types backed by a slice. sequential interface implements all
// collection type interfaces. map & fold operators rely on the consumeable
// type interface. vectorized types implement that behaviour
type Vectorized interface {
	Composed
	Sequential
	Searchable
	Ordered
	Indexed
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
	Value() Functional
}

//// STATE MONAD
type State interface {
	Run()
}
type StateFnc func() StateFnc

func (s StateFnc) Run() {
	var state = s()
	for state != nil {
		state = state()
	}
}

//// ERROR
type ErrorFnc func() error

func (e ErrorFnc) Error() string { return e().Error() }
