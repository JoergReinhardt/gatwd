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

// flag to mark kind of type
type KindOfType interface {
	Kind() d.Uint8Val
}

// marks all types defined in data
type NativeTyped interface {
	TypeNat() d.TyNat
}

// marks all types defined in functions
type FunctionTyped interface {
	TypeFnc() TyFnc
}

// native interface implements data native, provides assigability of
// functionale instances implementing it to native data structures using the
// 'data/Function' type.
// flag.
type Native interface {
	Expression
	TypeNat() d.TyNat
}

// extends the data/native interface with an eval method that takes arguments
// and returns a value of type data/native. funtional types implementing this
// use unboxed instances of native values
type NatEval interface {
	Native
	Eval(...d.Native) d.Native
}

// the functor interface takes n arguments of the expression type and returns a
// value of the expression type.
type Evaluable interface {
	Call(...Expression) Expression
}

// the expression interface is the least common denominator of all functional
// type primitives.
type Expression interface {
	FunctionTyped
	Evaluable
	Stringer
	Type() TyComp
}

///////////////////////////////////////////////////////////////////////////////
//// COLLECTION INTERFACES
///
// mapped interface is implementet by all key accessable data types
type Mapped interface {
	Len() int
	Keys() []string
	Values() []Expression
	Fields() []KeyPair
	Get(string) (Expression, bool)
	//d.Mapped
}

// consumeable is shared by all collections, continuations, side effects, etc‥.
// it returns the current head of a collection, last result, or input in a
// series of computations, data i/o operations, etc‥.
// a given instance will always return the same head and tail, the returned
// tail when consumed, will return the next step and so on. return values can
// either be passed on recurively as continuations, or be reassigned to the
// same values in a loop, thereby implementing a functional trampolin to
// flatten recursive calls.
// execution is performed lazily and infinite lists can be handled.
type Traversable interface {
	Expression
	Head() Expression
	Tail() Traversable
	Traverse() (Expression, Traversable)
}

// new elements can be pre-/ and appended to at the front and end of sequences.
// this is even true for infinite lists, since appending is performed lazily,
// which in the case of appending to an infinite list, may as well be never.
type Sequential interface {
	Traversable
	TypeElem() TyComp
	Cons(...Expression) Sequential // default op, list = front, vector = back
	Consume() (Expression, Sequential)
}
type Ordered interface {
	Sequential
	Swapable
	Less(Expression) bool
	Append(...Expression) Sequential
	Prepend(...Expression) Sequential
}

// interface to implement by all conditional types
type Testable interface {
	Expression
	Test(...Expression) bool
}

// interface to implement by ordered, sort-/ & searchable types
type Compareable interface {
	Expression
	Compare(...Expression) int
}

// types with known number of elements ,or known length
type Countable interface {
	Len() int
}

// interface provides indexable representation of elements
type Sliceable interface {
	Slice() []Expression
}

// swapping two elements in (pair-) position, or index
type Swapable interface {
	Swap() (Expression, Expression)
}

// fields in pairs can be swapped
type SwapablePaired interface {
	Swapable
	SwappedPair() Paired
}

// interface implementet by types sortable by some praedicate
type Sortable interface {
	Sort(Expression)
}

// interface implementet by types searchable by some praedicate
type Searchable interface {
	Search(Expression) int
}

//// INDEX ASSOCIATIONS
///
// elements that can tell their index position
type Indexed interface {
	Index() int
}

// interface implementet by types associating their elements with an index
type IndexAssociated interface {
	Get(int) (Expression, bool)
	Set(int, Expression) (Vectorized, bool)
}

// index value pair
type IndexPaired interface {
	Indexed
	Paired
}

// interface to accumulate propertys of a vector
type Vectorized interface {
	IndexAssociated
	//Searchable
	Sliceable
	//Sortable
	Countable
}

//// KEY ASSOCIATIONS
///
// implementet by value types that are accessed in a collection by a key
type Associated interface {
	Key() Expression
	Value() Expression
}

// implemented by keyed values, where the key is of type string
type Keyed interface {
	KeyStr() string
}

// pairs associated by key
type KeyPaired interface {
	Keyed
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
	GetVal(Expression) (Expression, bool)
	SetVal(Expression, Expression) (AssociativeCollected, bool)
}

// COMPOSED TYPES
// interface to be implemented by all pairs
type Paired interface {
	Swapable
	Expression
	Associated
	Associative
	Empty() bool
	Pair() Paired
	Left() Expression
	Right() Expression
	Both() (Expression, Expression)
}

// interface to be implemented by all collections providing random element
// access
type AssociativeCollected interface {
	KeyAssociated
	Sequential
	Associative
}

// extends the consumeable interface to work with collections of pairs
type ConsumeablePaired interface {
	Sequential
	Associative
	HeadPair() Paired
	TailPairs() ConsumeablePaired
	ConsumePair() (Paired, ConsumeablePaired)
}

// interface to be implementet by enumerable types
type Enumerable interface {
	Expression
	Next() EnumVal
	Prev() EnumVal
	EnumType() EnumDef
	Alloc(d.Numeral) EnumVal
}

// monadic interface generalizes step wise sequential progress of computations
// like i/o operations, evaluation of lists of commands, batch processing of
// files line by line, etc‥.
type Monadic interface {
	Expression
	Current() Expression
	Step(...Expression) (Expression, Monadic)
	Sequence() Sequential
}

// interface to implement by dynamicly declared and defined types
type ProtoTyped interface {
	Expression
	Name() string
	Methods() []string
	CallM(string, ...Expression) Expression
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
	Expression
	Root() Nodular
}

type Branched interface {
	Nodular
	Members() Nodular
}

type Leaved interface {
	Nodular
	Value() Expression
}

type Edged interface {
	Nodular
	From() Nodular
	To() Nodular
}
