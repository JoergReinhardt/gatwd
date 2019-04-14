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
type Numeral interface {
	d.Native
	d.Numeral
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
type Callable interface {
	d.Native
	TypeFnc() TyFnc
	Call(...Callable) Callable
}

type Functorial interface {
	Consumeable
	MapF(UnaryFnc) FunctFnc
	Fold(BinaryFnc, Callable) Callable
}

//// PAIRS OF FUNCTIONALS
type Applicable interface {
	Functorial
	Left() Callable
	Right() Callable
	Both() (Callable, Callable)
	Apply(...Callable) (Callable, PairFnc)
}
type Monadic interface {
	Functorial
	Map(Monadic) MonaFnc
}

//// COLLECTION CLASSES
type Optional interface {
	Callable
	Maybe() bool
	Value() Callable
}
type Predictable interface {
	Callable
	True(Callable) bool
	Any(...Callable) bool
	All(...Callable) bool
}
type Distinguishable interface {
	Callable
	Case(expr ...Callable) Callable
}
type Composed interface {
	Empty() bool //<-- no more nil pointers & 'out of index'!
}

type Countable interface {
	Len() int // <- performs mutch better on slices
}

type Sequential interface {
	Slice() []Callable //<-- no more nil pointers & 'out of index'!
}

type Linked interface {
	Next() Callable
}

type Reverseable interface {
	Prev() Callable
}

type Ordered interface {
	Sort(d.TyNative)
}

type Searchable interface {
	Search(Callable) int
}

type Indexed interface {
	Get(int) Callable
	Set(int, Callable) Vectorized
}

type Generating interface {
	Callable
	Next() Optional
}

type Aggregating interface {
	Callable
	Result() Callable
	Aggregator() NaryFnc
	Aggregate(...Callable) Callable
}

type Associative interface {
	Callable
	KeyFncType() TyFnc
	ValFncType() TyFnc
	KeyNatType() d.TyNative
	ValNatType() d.TyNative
	GetVal(Callable) PairFnc
	SetVal(Callable, Callable) Associative
	Pairs() []PairFnc
}

//// CONSUMEABLE COLLECTION ///////
///
// implemented by types backed by recursive lists. consumeable is the
// behaviour map-/ & fold operations rely up on
type Consumeable interface {
	Callable
	Head() Callable
	Tail() Consumeable
	DeCap() (Callable, Consumeable)
}

//// SEQUENTIAL LIST //////
///
// implemented by types backed by a slice. sequential interface implements all
// collection type interfaces. map & fold operators rely on the consumeable
// type interface. vectorized types implement that behaviour
type Vectorized interface {
	Callable
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
	Value() Callable
}
