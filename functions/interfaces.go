package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

//// interfaces imported from data
type Flagged interface {
	Match(d.Typed) bool
	Flag() d.BitFlag
}
type Typed interface {
	Type() TyDef
	TypeFnc() TyFnc
	TypeName() string
	FlagType() d.Uint8Val
}

type Evaluable interface {
	Eval(...d.Native) d.Native
}

type Expression interface {
	Typed
	Call(...Expression) Expression
	Eval(...d.Native) d.Native
	TypeNat() d.TyNat
	String() string
}

type Reproduceable interface {
	// Copy() Native
	d.Reproduceable
}

//  garbage collectability
type Destructable interface {
	// Clear()
	d.Destructable
}

// numberal interfaces
type Boolean interface {
	// Bool() bool
	d.Boolean
}
type Discrete interface {
	// Unit()
	d.Discrete
}
type Natural interface {
	// Uint() uint
	d.Natural
}
type Integer interface {
	// Int() int
	d.Integer
}
type Rational interface {
	// Rat() *big.Rat
	d.Rational
}
type Real interface {
	// Float() float64
	d.Real
}
type Imaginary interface {
	// Imag() complex128
	d.Imaginary
}
type Numeral interface {
	Expression
	Discrete
	Boolean
	Natural
	Integer
	Rational
	Real
	Imaginary
}

type Text interface {
	// String() string
	d.Text
}

type Raw interface {
	// Bytes() []byte
	d.Raw
}

type Letter interface {
	// Rune() rune
	// Byte() byte
	d.Letter
}

type BinaryMarshaler interface {
	// MarshalBinary() ([]byte, error)
	d.BinaryMarshaler
}

type Serializeable interface {
	// MarshalBinary() ([]byte, error)
	d.Serializeable
}

type Printable interface {
	// String() string
	// Bytes() []byte
	// Runes() []rune
	d.Printable
}

type Verifyable interface {
	Bool() bool
}

type Constructing interface {
	Expression
	Const() Expression
}

type Paired interface {
	Consumeable
	Empty() bool
	Pair() Paired
	Left() Expression
	Right() Expression
	Both() (Expression, Expression)
	KeyNatType() d.TyNat
	ValNatType() d.TyNat
	KeyType() TyDef
	ValType() TyDef
}

type Composed interface {
	// Empty()
	d.Composed
}

type Slice interface {
	// Composed
	// Len() int
	// Slice() []Native
	d.Sliceable
}

type Sequential interface {
	// Head() Native
	// Tail() Sequential
	// Shift() Sequential
	d.Sequential
}

type Mapped interface {
	// Len() int
	// Keys() []Native
	// Data() []Native
	// Fields() []Paired
	// Get(acc Native) (Native, bool)
	// Delete(acc Native) bool
	// Set(Native, Native) Mapped
	d.Mapped
}

// branched yields two callable return values
type Branched interface {
	Expression
	Left() Expression
	Right() Expression
	Both() (Expression, Expression)
}

// swaps position of branched return values either per call, or yielding a
// new instance containing the values in swapped position
type Swappable interface {
	Expression
	Swap() (Expression, Expression)
	SwappedPair() Paired
}

// associated values are pairs of values, associated to one another and to any
// collection of associates, in as the left field is either key, or index to
// access an associated value as an element in a collection. in optionals right
// field indicates the optional type of the result yielded, in maybes it
// indicates succssess, etc‥.
type Accociated interface {
	Expression
	Key() Expression
	Value() Expression
}

type Keyed interface {
	Expression
	KeyStr() string
}

type Indexed interface {
	Expression
	Index() string
}

// access elements directly by index position
type IndexAssoc interface {
	Expression
	Get(int) (Expression, bool)
	Set(int, Expression) (Vectorized, bool)
}

//// CONSUMEABLE
///
// all types that implement some type of collection, or step by step recursion,
// are 'enumerated', as in expected to work with every given type as a single
// instance, or collection there of. consumeable implements that behaviour.
// the behaviour the map-/ & fold operators rely on <Head() Callable> as 'unit'
// function, which forms the base of all functors, applicatives, monads‥.
type Consumeable interface {
	Expression
	Head() Expression
	Tail() Consumeable
	Consume() (Expression, Consumeable)
}

//// CONSUMEABLE PAIRS
type ConsumeablePairs interface {
	Consumeable
	HeadPair() Paired
	TailPairs() ConsumeablePairs
	ConsumePair() (Paired, ConsumeablePairs)
}

//// FUNCTOR

// has numerous elements and knows it
type Countable interface {
	Len() int // <- performs mutch better on slices
}

// collection of elements in a fixed sequence accessable by index, aka
// 'slice/array'
type Sequenced interface {
	Slice() []Expression
}

// ordered collections can be sorted by‥.
type Sortable interface {
	Sort(d.TyNat)
}

// ‥.and searched after based on a predicate
type Searchable interface {
	Search(Expression) int
}

// combines common functions provided by all vector shaped data
type Vectorized interface {
	Sequenced
	Searchable
	Sortable
	IndexAssoc
}

// bahaviour of aggregators, that take one, or many arguments per call and
// yield the current result of a computation aggregating those passed
// arguments. base of fold-l behaviour.
type Aggregating interface {
	Expression
	Result() Expression
	Aggregate(...Expression) Expression
}

// associative values and collections have either a key, or index position, to
// associate them with their position in a collection
type Associative interface {
	Expression
	KeyType() TyDef
	ValType() TyDef
	KeyNatType() d.TyNat
	ValNatType() d.TyNat
	GetVal(Expression) (Expression, bool)
	SetVal(Expression, Expression) (Associative, bool)
	Pairs() []Paired
}

//// TREES & GRAPHS
///
type Nodular interface {
	Sequenced
	Root() Nodular
}

type Chained interface {
	Nodular
	Next() Nodular
}

type RevChained interface {
	Prev() Nodular
}

type Edged interface {
	Nodular
	From() Nodular
	To() Nodular
}

type Leaved interface {
	Nodular
	Value() Expression
}
