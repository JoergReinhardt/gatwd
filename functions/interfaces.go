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
type Native interface {
	Expression
	TypeNat() d.TyNat
}

// extends the data/native interface with an eval method that takes arguments
// and returns a value of type data/native.  funtional types implementing this
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
// continuation is shared by all collections, generators, monads, side effects,
// etc‥., it returns the current state of computation, io operation, current
// elemet in collection‥., as head element and a group of computations to
// continue on as its tail.  a given instance will always return the same head
// and tail when called without, or with identical arguments.
//
// continuations are not necessarily collections of elements, but they do
// always return tails of the sequential type, since there may be more than one
// 'computation' necessary to evaluate the continuation .
//
// execution is performed lazily and infinite lists may be handled.
//
// CAVEAT: a continuations 'call' method always has to return a continuation,
// when called, in order to satisfy constraints for map, apply, fold, bind
//
// C.Call(args...).(Continuation) ← has to work!
type Continuation interface {
	// Call(...Expression) <Continuation>
	Expression
	Empty() bool
	TypeElem() TyComp
	Head() Expression
	Tail() Group
	Continue() (Expression, Group)
}

// a group has elemets with a binary operation (takes two arguments which have
// to be members of the set, returns a set membe) defined up on.  for numeric
// types that might be a arithmetic operation, concatenation for strings, etc‥.
//
// groups expose a cons operation to instanciate a new element from the group
// an apply itself and the new instance to the binary operation to return a new
// instance of an element from the group.  for collections that pre-, or
// appends an element. continuations take those arguments to either apply them
// to the current state and return the result, or create and add new
// computations to continue on.
type Group interface {
	Continuation
	Cons(...Expression) Group
	ConsGroup(Group) Group
	Concat(Group) Group
}

type Directional interface {
	Group
	First() Expression
	Suffix() Directional
	Append(Group) Directional
	Prepend(Group) Directional
	AppendArgs(...Expression) Directional
	PrependArgs(...Expression) Directional
}
type BiDirectional interface {
	Last() Expression
	Prefix() BiDirectional
}

// stack pushes new elements as first element to the group & pops the last
// element that has been added from the sequence
type Stack interface {
	Group
	Pop() (Expression, Stack)
	Push(...Expression) Stack
}

// queues pull elements from the end of the group and put elements at its
// start
type Queue interface {
	Group
	Pull() (Expression, Queue)
	Put(...Expression) Queue
}

type Filtered interface {
	Continuation
	Filter(Testable) Group
	Pass(Testable) Group
}

type Functorial interface {
	Continuation
	Map(fn Expression) Group
	Fold(acc Expression, fn func(...Expression) Expression) Group
	Flatten() Group
}

// mapped interface is implementet by all key accessable data types
type Mapped interface {
	Len() int
	Keys() []Expression
	Values() []Expression
	Fields() []Paired
	Get(Expression) (Expression, bool)
}

type Zipped interface {
	Functorial
	Split() (l, r Group)
}
type Zippable interface {
	Functorial
	ZipWith(
		zipf func(l, r Continuation) Group, with Continuation,
	) Group
}
type Applicable interface {
	Functorial
	Apply(func(
		Group, ...Expression,
	) (
		Expression,
		Continuation,
	)) Group
}

type Monoidal interface {
	Applicable
	Bind(Expression, Functorial) Group
}

type Ordered interface {
	Group
	Lesser(Ordered) bool
	Greater(Ordered) bool
	Equal(Ordered) bool
}

///////////////////////////////////////////////////////////////////////////////
// interface to implement by all conditional types
type Testable interface {
	Expression
	Test(Expression) bool
}

// interface to implement by ordered, sort-/ & searchable types
type Compareable interface {
	Expression
	Compare(Expression) int
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
	Sort(
		lesser func(a, b Expression) bool,
	) Group
}

// interface implementet by types searchable by some praedicate
type Searchable interface {
	Search(
		match Expression,
		compare func(a, b Expression) int,
	) Expression
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
	Searchable
	Sliceable
	Sortable
	Countable
	Prefix() VecVal
	Suffix() VecVal
	First() Expression
	Last() Expression
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
	Group
	Associative
}

// extends the consumeable interface to work with collections of pairs
type ConsumeablePaired interface {
	Group
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
	EnumType() EnumCon
	Alloc(d.Numeral) EnumVal
}

// monadic interface generalizes step wise sequential progress of computations
// like i/o operations, evaluation of lists of commands, batch processing of
// files line by line, etc‥.
type Monadic interface {
	Expression
	Current() Expression
	Sequence() Group
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
