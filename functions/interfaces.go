package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

// stringer provides string representation
type Stringer interface {
	String() string
}

// all instances of data-/ & functions types carry at least one bitflag to mark
// their internal type
type Flagged interface {
	Flag() d.BitFlag
}

// all type markers derive a function from bitflags to match other instances of
// their kind, or membership in or concatenated sets there of.
type Matched interface {
	Match(d.Typed) bool
}

// flag to mark kind of type
type KindOfType interface {
	Kind() d.Uint8Val
}

// typed mirrors the data/typed interface and provides
//
// - type name, which is another string representation that might be different
//   from the flag constants name.
// - functiont to reveal the type as instance of the bitflag type
// - match method
// - string representation for each type a flag to mark the kind of type flag,
type Typed interface {
	d.Typed
	// KindOfType
	// NameTyped
	// Flagged
	// Matched
	// Stringer
}

// marks all types defined in data
type NativeTyped interface {
	TypeNat() d.TyNat
}

// marks all types defined in functions
type FunctionTyped interface {
	TypeFnc() TyFnc
}

type DynamicTyped interface {
	FunctionTyped
	NativeTyped
	Len() int
	Elements() []d.Typed
}

/// NATIVE (ALIASED)
// native interface implements data native, provides assigability of
// functionale instances implementing it to native data structures using the
// 'data/Function' type.
type Native interface {
	Functor
	TypeNat() d.TyNat
}

/// EVALUABLE
// extends the data/native interface with an eval method that takes arguments
// and returns a value of type data/native.  funtional types implementing this
// use unboxed instances of native values
type Evaluable interface {
	Native
	Eval(...d.Native) d.Native
}

/// CALLABLE
// the functor interface takes n arguments of the expression type and returns a
// value of the expression type.
type Expression interface {
	Call(...Functor) Functor
}

/// FUNCTOR
// the functor interface implements all it takes for a type to be the applied
// expression, argument, or return value in some map operation.  map calls
// functors 'call' method internaly and consults the type tags returned by its
// methods, in order to check argument types in function application, or
// control program flow based on argument values, and|or type.
type Functor interface {
	FunctionTyped
	Expression
	Stringer
	Type() Decl
}

/// CONTINUATION
type Continuous interface {
	Continue() (Functor, Applicative)
	Head() Functor
	Tail() Applicative
}

/// SEQUENTIAL
type Sequential interface {
	Functor // → Call(...Expression) Functor.(Continuation)
	Continuous
	Empty() bool
	TypeElem() Decl
	Concat(Sequential) Applicative
}

/// APPLICATIVE (FUNCTOR)
// applicative functors are defined by a method that takes its argument, boxes,
// encloses, manipulates, or useses it in a computation, to return a new
// instance of the applicatives type.
type Applicative interface {
	Sequential
	Cons(Functor) Applicative
}

type Monoid interface {
	Applicative
	Bind(f, g Functor) Monadic
}

type Ascending interface {
	Applicative
	First() Functor
}
type Descending interface {
	Ascending
	Last() Functor
}

// queues pull elements from the end of the group and put elements at its
// start
type Queued interface {
	Applicative
	Put(Functor) Queued
	Pull() (Functor, Queued)
	Append(...Functor) Queued
}
type Stacked interface {
	Applicative
	Push(Functor) Stacked
	Pop() (Functor, Stacked)
}

type Filtered interface {
	Sequential
	Filter(Testable) Applicative
	Pass(Testable) Applicative
}

type Mappable interface {
	Map(fn Functor) Applicative
}
type Foldable interface {
	Fold(acc Functor, fn func(...Functor) Functor) Applicative
}
type Flatable interface {
	Sequential
	Foldable
	Mappable
	Flatten() Applicative
}

// mapped interface is implementet by all key accessable data types
type Hashable interface {
	Len() int
	Keys() []Functor
	Values() []Functor
	Fields() []Paired
	Get(Functor) (Functor, bool)
}

type Zipped interface {
	Flatable
	Split() (l, r Applicative)
}
type Zippable interface {
	Flatable
	ZipWith(
		zipf func(l, r Sequential) Applicative, with Sequential,
	) Applicative
}

type Ordered interface {
	Applicative
	Lesser(Ordered) bool
	Greater(Ordered) bool
	Equal(Ordered) bool
}

///////////////////////////////////////////////////////////////////////////////
// interface to implement by all conditional types
type Testable interface {
	Functor
	Test(a, b Functor) bool
}

// interface to implement by ordered, sort-/ & searchable types
type Compareable interface {
	Functor
	Compare(Functor) int
}

// types with known number of elements ,or known length
type Countable interface {
	Len() int
}

// interface provides indexable representation of elements
type Sliceable interface {
	Slice() []Functor
}

// swapping two elements in (pair-) position, or index
type Swapable interface {
	Swap() (Functor, Functor)
}

// fields in pairs can be swapped
type SwapablePaired interface {
	Swapable
	SwappedPair() Paired
}

// interface implementet by types sortable by some praedicate
type Sortable interface {
	Sort(
		lesser func(a, b Functor) bool,
	) Applicative
}

// interface implementet by types searchable by some praedicate
type Searchable interface {
	Search(
		match Functor,
		compare func(a, b Functor) int,
	) Functor
}

//// INDEX ASSOCIATIONS
///
// elements that can tell their index position
type Indexed interface {
	Index() int
}

// interface implementet by types associating their elements with an index
type Selectable interface {
	Get(int) (Functor, bool)
	Set(int, Functor) (RandomAcc, bool)
}

// index value pair
type IndexPaired interface {
	Indexed
	Paired
}

// interface to accumulate propertys of a vector
type RandomAcc interface {
	Selectable
	Searchable
	Sliceable
	Sortable
	Countable
	Prefix() VecVal
	Suffix() VecVal
	First() Functor
	Last() Functor
}

//// KEY ASSOCIATIONS
///
// implementet by value types that are accessed in a collection by a key
type Associated interface {
	Key() Functor
	Value() Functor
}

// implemented by keyed values, where the key is of type string
type Keyed interface {
	KeyStr() string
}

// pairs associated by key
type KeyPaired interface {
	Paired
}

// implementet by types that provide random access to internal elements based
// on keys, or indices
type Associative interface {
	TypeKey() d.Typed
	TypeValue() d.Typed
}

// implementet by types that provide random access to internal elements based
// on keys
type KeyAssociated interface {
	Pairs() []Paired
	GetVal(Functor) (Functor, bool)
	SetVal(Functor, Functor) (AssociativeCollected, bool)
}

// COMPOSED TYPES
// interface to be implemented by all pairs
type Paired interface {
	Swapable
	Functor
	Associated
	Associative
	Empty() bool
	Pair() Paired
	Left() Functor
	Right() Functor
	Both() (Functor, Functor)
}

// interface to be implemented by all collections providing random element
// access
type AssociativeCollected interface {
	KeyAssociated
	Applicative
	Associative
}

// extends the consumeable interface to work with collections of pairs
type ConsumeablePaired interface {
	Applicative
	Associative
	HeadPair() Paired
	TailPairs() ConsumeablePaired
	ConsumePair() (Paired, ConsumeablePaired)
}

// interface to be implementet by enumerable types
type Enumerable interface {
	Functor
	Next() Enumerable
	Prev() Enumerable
	Create(d.Numeral) Enumerable
}

// monadic interface generalizes step wise sequential progress of computations
// like i/o operations, evaluation of lists of commands, batch processing of
// files line by line, etc‥.
type Monadic interface {
	Functor
	Current() Functor
	Sequence() Applicative
}

// interface to implement by dynamicly declared and defined types
type ProtoTyped interface {
	Functor
	Name() string
	Methods() []string
	CallM(string, ...Functor) Functor
}

///////////////////////////////////////////////////////////////////////////////
//// PROTOTYPE COMPOUNDS
////
//// TREES & GRAPHS
///
type Linked interface {
	Sliceable
	Next() Nodular
}

type DoubleLinked interface {
	Prev() Nodular
}

type Nodular interface {
	Linked
	Functor
	Root() Nodular
}

type Branched interface {
	Nodular
	Members() Nodular
}

type Leaved interface {
	Nodular
	Value() Functor
}

type Edged interface {
	Nodular
	From() Nodular
	To() Nodular
}
