package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

//// PRIMITIVE TYPE CLASS INTERFACES
type Typed interface {
	//Flag() d.BitFlag
	d.Typed
}

type Reproduceable interface {
	//Copy() Native
	d.Reproduceable
}

// garbage collectability
type Destructable interface {
	//Clear()
	d.Destructable
}

type BinaryMarshaler interface {
	//MarshalBinary() ([]byte, error)
	d.BinaryMarshaler
}

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
type Paired interface {
	d.Paired
}
type Composed interface {
	d.Composed
}
type Slice interface {
	d.Sliceable
}
type Sequential interface {
	d.Sequential
}
type Mapped interface {
	d.Sequential
}

//// FUNCTIONAL CLASS
type Callable interface {
	d.Native
	TypeFnc() TyFnc
	Call(...Callable) Callable
}

// ENDOFUNCTORS
// all functors are mappable
type Functoric interface {
	Consumeable
	Map(UnaryFnc, ...Callable) Functoric
	Fold(BinaryFnc, ...Callable) Callable
}

// applicatives compose functoric operations (mappings) sequentialy
type Applicative interface {
	Functoric
	Apply(...Callable) Callable
}

// applicatives compose functoric operations (mappings) sequentialy, deciding
// which computations to perform based on the result of previous computations.
type Monadic interface {
	Applicative
	Join(f, g Consumeable) Callable
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

type Countable interface {
	Len() int // <- performs mutch better on slices
}

type Sequenced interface {
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
	GetVal(Callable) PairVal
	SetVal(Callable, Callable) Associative
	Pairs() []PairVal
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
	Sequenced
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
	Sequenced
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
