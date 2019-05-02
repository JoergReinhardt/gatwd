package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

//// interfaces imported from data
type Typed interface {
	// Flag() d.BitFlag
	d.Typed
}

type Evaluable interface {
	// Eval(...Native) Native
	d.Evaluable
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

type Nullable interface {
	// Null()
	d.Native
	d.Nullable
}
type Discrete interface {
	// Unit()
	d.Discrete
}
type Boolean interface {
	// Bool() bool
	d.Boolean
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
	// Nullable
	// Discrete
	// Boolean
	// Natural
	// Integer
	// Rational
	// Real
	// Imaginary
	d.Numeral
}

type Text interface {
	// String() string
	d.Text
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

type Paired interface {
	Callable
	Pair() Paired
	Left() Callable
	Right() Callable
	Both() (Callable, Callable)
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

///////////////////////////////////////////////////////////////////
//// FUNCTIONAL INTERFACES
///
// Callable interface
//
// is the smallest common denominator of all functional types. all arguments
// and return values either are instances, or compositions of values,
// implementing the data/Native interface. callables can themselfes be treatet
// as data and need to implement the Native interface to. a callable should
// return it's return types native type flag when <TypeNat() d.TyNative> is
// called. TypeFnc() TyFnc returns the funtional type of a callable
//
// callables usually return the enclosed, or computed value when called.
// callables enclosing static data just return it, when called without
// arguments. if arguments are passed, they are either considered function
// arguments, when the callable is an expression, or data to be CONVERTED to an
// instance of that callables type, which forms the base of all functors,
// applicatives, monads‥. ability to convert, wrap, or box native values to add
// functional behaviour.
type Callable interface {
	d.Native
	// TypeNat() d.TyNative
	// String() string
	// Eval(...d.Native) d.Native
	TypeFnc() TyFnc
	Call(...Callable) Callable
}

type Parametric interface {
	Callable
	Type() Typed
	TypeName() string
}

// branched yields two callable return values
type Branched interface {
	Callable
	Left() Callable
	Right() Callable
	Both() (Callable, Callable)
}

// swaps position of branched return values either per call, or yielding a
// new instance containing the values in swapped position
type Swappable interface {
	Callable
	Swap() (Callable, Callable)
	SwappedPair() PairVal
}

// associated values are pairs of values, associated to one another and to any
// collection of associates, in as the left field is either key, or index to
// access an associated value as an element in a collection. in optionals right
// field indicates the optional type of the result yielded, in maybes it
// indicates succssess, etc‥.
type Accociated interface {
	Callable
	Key() Callable
	Value() Callable
}

type Keyed interface {
	Callable
	KeyStr() string
}

type Indexed interface {
	Callable
	Index() string
}

// access elements directly by index position
type IndexAssoc interface {
	Callable
	Get(int) Callable
	Set(int, Callable) Vectorized
}

//// CONSUMEABLE
///
// all types that implement some type of collection, or step by step recursion,
// are 'enumerated', as in expected to work with every given type as a single
// instance, or collection there of. consumeable implements that behaviour.
// the behaviour the map-/ & fold operators rely on <Head() Callable> as 'unit'
// function, which forms the base of all functors, applicatives, monads‥.
type Consumeable interface {
	Callable
	Head() Callable
	Tail() Consumeable
	DeCap() (Callable, Consumeable)
}

//// FUNCTOR
// callables that are consumeables providing a 'map' method, are functors.
type Functoric interface {
	Consumeable
	Map(UnaryExpr, ...Callable) Functoric
	Fold(BinaryExpr, ...Callable) Callable
}

// applicatives are infix operators that compose functor operations
// sequentialy. the bind function contains the 'composing behaviour', i.e.
// knows how to compose the expected return types of it's operands, to yield
// the applicatives boxed return type.
type Applicative interface {
	Functoric
	Bind(...Callable) Callable
}

// monads are functors with two additional operations, one of which is
// implemented by a type constructor that knows how to convert any given type
// to be the boxed monad type. every empty, or depleted monad, behaves like a
// (data) type constructor, when called providing arguments. arguments are
// wrapped to be of the monad type. calling a monad that is not empty, results
// in argument convertion and composition of the arguments with the preexisting
// elements contained by the monad and yields a new instance of the monad type.
// for a list, that effectually appends the arguments as new elements to the
// preexisting list and returns that newly created list. the 'unit' behaviour
// is implemented by calling the monad without arguments, which will yield the
// first element as head and all preceeding elements as tail value.
type Monadic interface {
	Applicative
	Join(f, g Consumeable) Callable
}

// a predicate makes things predictable. for enumerated types it can also
// distinguish between cases where all elements of a list are considered true
// and those where some element in the list turns out to be true. this
// implements early return and works on infinite lists to.
type Predictable interface {
	Callable
	True(Callable) bool
	Any(...Callable) bool
	All(...Callable) bool
}

// provides branching by any distinguishable property. aka 'case switch'
type Distinguishable interface {
	Callable
	Case(expr ...Callable) Callable
}

// has numerous elements and knows it
type Countable interface {
	Len() int // <- performs mutch better on slices
}

// collection of elements in a fixed sequence accessable by index, aka
// 'slice/array'
type Sequenced interface {
	Slice() []Callable
}

// ordered collections can be sorted by‥.
type Sortable interface {
	Sort(d.TyNat)
}

// ‥.and searched after based on a predicate
type Searchable interface {
	Search(Callable) int
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
	Callable
	Result() Callable
	Aggregator() NaryExpr
	Aggregate(...Callable) Callable
}

// associative values and collections have either a key, or index position, to
// associate them with their position in a collection
type Associative interface {
	Callable
	KeyFncType() TyFnc
	ValFncType() TyFnc
	KeyNatType() d.TyNat
	ValNatType() d.TyNat
	GetVal(Callable) Callable
	SetVal(Callable, Callable) Associative
	Pairs() []Paired
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

type Edged interface {
	Nodular
	From() Nodular
	To() Nodular
}

type Leaved interface {
	Nodular
	Value() Callable
}
