package functions

import (
	d "github.com/JoergReinhardt/gatwd/data"
)

// unique identifyer of every higher order type, returns unique id, name,
// signature, native- & functional type of a given type.
type TypeIdent interface {
	Value
	Name() string
	Pattern() string
}
type Typed interface {
	Value() interface{}
	Name() string
}

// a type pattern matcher is associated with a types data-, or type-constructor
// and has the ability to either match, or don't match a given input type to
// determine if either a derived type, or a data value instance of the
// designated type can be created from that input.
type PatternMatcher interface {
	Match(...Typed) bool
}

// data constructor takes instance(s) of input type(s) and generates data
// instances of the designated type, in case the input type matches the pattern
// matcher.
type DataConstructor interface {
	TypeIdent
	PatternMatcher
	Con(...Value) Value
}

// type constructor takes type argument(s) & derives new types in case the
// pattern of the input type(s) are matched by the pattern matcher.
type TypeConstructor interface {
	TypeIdent
	PatternMatcher
	Con(...TypeIdent) TypeIdent
}

// polymorphic type definition returns collections of all, data and type
// constructors defined for a given type.
type TypeDefinition interface {
	TypeIdent
	DataCons() []DataConstructor
	TypeCons() []TypeConstructor
}

type Equation interface {
	Name() string
	Pattern() string
	PatternMatcher
	Callable
}
type FunctionDefinition interface {
	Name() string
	Equations() []Equation
}

// import native value interface from data package
type Native interface {
	d.Native
}

// values are at least a native value, accompanyed by a functional type, that
// can be Evaluated to a native (constant wrapping an atomic native in the
// simplest case)
type Value interface {
	Native
	TypeFnc() TyFnc
	Call(...Value) Value
}
type Callable interface {
	Value
}

// nullable 'classes'
type Nullable interface{ d.Nullable }
type Bitwise interface{ d.Binary }
type Boolean interface{ d.Boolean }
type Natural interface{ d.Natural }
type Integer interface{ d.Integer }
type Rational interface{ d.Rational }
type Real interface{ d.Real }
type Imaginary interface{ d.Imaginary }
type Number interface{ d.Number }
type Letter interface{ d.Letter }
type Text interface{ d.Text }
type Printable interface{ d.Printable }
type PairedNative interface{ d.Paired }
type ComposedNative interface{ d.Composed }
type SliceNative interface{ d.Sliceable }
type SequentialNative interface{ d.Sequential }
type MappedNatives interface{ d.Sequential }

// FUNCTIONAL KIND INTERFACES
//
// data augmented by the functional interface can be discriminated by it's
// kind. all types of a common kind need to provide that kinds interface
// methods.
//
// Paired interface{}
//
// the paired interface is the functional version of data packages paired type.
// it returns two values, omiting the encapsulation by an additional struct
// instance the data type brings with it. it can be implemented by that type,
// or anyhing else that returns two values.
type Paired interface {
	Value
	Left() Value
	Right() Value
	Both() (Value, Value)
}
type ByteCode interface {
	Value
	Bytes() []byte
}

/////////////////////////////////
/// closure over fnc that may, or may not return a result.
// returned value contains either value, or none Val
type Optional interface {
	Maybe() bool
	Value() Value
}

// 'APPLYABLE' DATA
//
// gatwd aims for imutable values since data in functional programming is
// conceptionally mandatory to be imutable. go on the other hand is inherently
// procedual and relys on highly volatile mutable values, especially when
// dealing with things like inidices and loop counters. there is no way to keep
// code from mutating values on the syntax layer either. gatwd provides a
// functional abstraction to implement data that can be applyed as parameter or
// argument, without assignment. the changed data is returned as closure over
// the new data instead. this is implemented by providing closure function
// types that close over data instances. the function signature of applyable
// types is variadic, and the implementations internaly discriminate between
// the case where arguments have been passed against the case where the
// function has been called with an empty argument set. in the first case, the
// return value will be the new data and a new instance of the applyable type,
// now enclosing the new data. called without arguments instead, the current
// data and an identical instance of the applyable type will be returned.
//
// Argumented interface{}
//
// the 'simplest' applyable type provides methods to be cast as either of type
// data, or argumented and implements 'Functional'. the interface mirrors the
// applyable function type (it's applyable behaviour, so to speak) in it's
// 'Apply(...Data) (d.Data, Argumented)' method.
type Argumented interface {
	Value
	ArgType() d.TyNative
	Arg() Value
	Apply(...Value) (Value, Argumented)
}

// Arguments interface{}
//
// interface all lists of arguments must implement to gain applyability
type Arguments interface {
	Value
	Len() int
	Args() []Argumented
	Data() []Value
	Get(int) Argumented
	Replace(int, Value) Arguments
	Apply(...Value) ([]Value, Arguments)
}

// Parametric interface{}
//
// an argument with a distinct identifyer is a parameter. the identifyer can be
// a name, like for instance in a list of tupled function arguments, but also
// an integer denoting it's position in a tuple-, or slice, as well as a
// search, or sort praedicate that provides additional information for what to
// search-/, or sort by.
type Parametric interface {
	Value
	Parm() Parametric
	Arg() Value
	Acc() Value
	Both() (Value, Value)
	Left() Value
	Right() Value
	Pair() Paired
	Apply(...Parametric) (Value, Parametric)
}

// Parameters interface{}
//
// a list of instances of the parametric type.
type Parameters interface {
	Value
	Len() int
	Pairs() []Paired
	Parms() []Parametric
	Get(Value) Paired
	Replace(Paired) Parameters
	ReplaceKeyValue(k, v Value) Parameters
	Apply(...Parametric) ([]Parametric, Parameters)
	AppendKeyValue(k, v Value) Parameters
}

// Composed interface{}
//
// everything that has a length is countable. that's usually a collection but
// not neccessariely. A notable example of an excemption is the BitFlag, which
// can be concatenated from multiple flags by bitwise OR concatenation and
// provides a 'Len() int' method to check for, and a 'Deompose() []BitFlag'
// method to yield the components.
type Composed interface {
	Empty() bool //<-- no more nil pointers & 'out of index'!
}

// Sliceable interface{}
//
// the collected interface is implemented by all collection types. they need to
// be countable and also provide a method to check if they're empty or not.
// that's neccessary, since the way to implement a performant empty check
// highly depends on the type (recursive vs. flat)
type Sliceable interface {
	Composed
	Slice() []Value //<-- no more nil pointers & 'out of index'!
}

// Quantified interface{}
// is a collection type that can provide a flat slice representation of it's
// data
type Quantified interface {
	Len() int // <- performs mutch better on slices
}

// Vectorized interface{}
//
// is a quantified (indexable slice) type that mimics the pattern common in
// functional programming, to have a 'had' and 'tail' method. since array like
// data types are inherently different from recursive datatypes, that mimicry
// only goes so far‥. note the difference in return type when compared to the
// 'Recursive' inerface, that provides methods of the same name. the interface
// allows programmers to implement structurally identical 'business logic' in
// all the cases, where that difference doesn't matter, by expecting and
// asserting arguments and return values to be of a more  general typ more
// general interface type like 'functional', 'Countable', 'Collected'‥.
// whenever neccessary instances can be discriminating and treated as either
// 'Vectorized', or 'Recursive'.
type Sequential interface {
	Head() Value
	Tail() []Value
	DeCap() (Value, []Value)
}
type Vectorized interface {
	Value
	Sliceable
	Quantified
	// !!! not to be mixe with recursive !!!
	Sequential
	Search(Value) int
	Sort(d.TyNative)
	Get(int) Value
	Set(int, Value) Vectorized
}

//// RECURSIVE LISTS ///////
//
// recursive functional pendant to an array.
type Recursive interface {
	Value
	Quantified
	// !!! recursive data structures can't provide slices !!!
	Head() Value
	Tail() ListFnc
	DeCap() (Value, ListFnc)
}

//// TUPLES /////
//
// Tupled interface{}
//
// common propertys of all tuple types, commonly used in functional data type
// used for instanced as argument set and composit retrun type during function
// application. each unique combination of types and the order they are
// contained in, constitutes a unique higher order type named by it's unique
// signature.
type Tupled interface {
	Vectorized
	Flags() []d.BitFlag
}

// Recorded{}
//
// a tuple containing pairs of key/value data can be addressed by.
type Associative interface {
	GetVal(Value) Paired
	SetVal(Value, Value) Associative
}
type Recorded interface {
	Vectorized
	Associative
	Tuple() Tupled
	ArgSig() []Paired // key/data-type
}
type Accessable interface {
	Associative
	Pairs() []Paired
	Head() Paired
	Tail() []Paired
	DeCap() (Paired, []Paired)
}

// LINKED LISTS
//
// Ordered interface{}
// all recursive collections are inherently linked and may be accessed
// sequentially until depletion, to gain an ordered list. the ordered interface
// makes that explicit, provides a more convienient way to deal with it and/or
// allows slice based collections to be used in protty much the same way.
type Ordered interface {
	Sliceable
	Next() (Value, Ordered)
}

// Reverseable interface{}
//
// to implement reversability a linked data structure needs to provide
// additional reference to it's predeccessor node
type Reverseable interface {
	Ordered
	Prev() Value
}

/// NESTED COLLECTIONS /////
//
// just like with linked lists, every recursice data structure can also be
// considered a nested datasructure. nestedness also manifests in data
// structure (lists of lists of...) , after recursive processing, nested data
// doesn't have to be linear either (as linked lists are) and can be considered
// acyclic graphs. (acyclic, since nesting can't be infinite)
//
//////// TREES ////////
//
// tree interfaces implement the superset of linked lists, as well as nested
// structures acyclic graphs. the 'chained' and 'Root' interface types by
// providing absolute reference to a starting point and explicit order, can
// also provide representation of cyclic graphs implementation of a ringbuffer
// as simplemost example, more complicated structures can be implemented by
// including an edge implementation, optionally implemented by a set of
// different edge types (or colours).
type Nodular interface {
	Sliceable
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
	Value() Value
}

type StateFnc func() StateFnc

func (s StateFnc) Run() {
	for state := s(); state != nil; {
		state = state()
	}
}
