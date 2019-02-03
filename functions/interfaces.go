/*
INTERFACES

  interfaces provide connection between higher order function composition and
  actual implementation, as well as transition between data and function package
  types
*/
package functions

import (
	"math/big"
	"time"

	d "github.com/JoergReinhardt/godeep/data"
)

// BitFlag interface{}
//
// alias for the  bitflag interface from data package.  currently kind of
// useless, since functions from the data package expect bitfags to be of type
// data.BitFlag. => TODO: make usefull
type BitFlag interface{ d.Data }

// nullable 'classes'
//
// functions package provides an interface for each group of datas types
// grouped by common methods they provide. the groups are defined as bitwise
// concatenations of flags of all the types providing the common method. this
// builds the base for implementing higher order type classes.
type Nullable interface{ Null() Functional }
type Bitwise interface{ Uint() uint }
type Boolean interface{ Bool() bool }
type Unsigned interface{ Uint() uint }
type Signed interface{ Int() int }
type Integer interface{ Int() int }
type Rational interface{ Rat() *big.Rat }
type Irrational interface{ Float() float64 }
type Imaginary interface{ Imag() complex128 }
type Timed interface{ Time() time.Time }
type Temporal interface{ Dura() time.Duration }
type Symbolic interface{ String() string }
type Collection interface{ Len() int }
type Numeric interface {
	Uint() uint
	Int() int
	Flt() float64
}
type Synbolic interface {
	String() string
	Bytes() []byte
}

// Functional interface{}
//
// augments instances of data packages types with a 'kind' method. 'kinds of
// data' are sets of functional types providing common methods. this maps data
// packages 'classes' and there associated methods in the functions package, to
// the concept of type class.
type Functional interface {
	d.Data
	Kind() d.BitFlag
	Eval() d.Data
}
type Function interface {
	Functional
	Call(...Functional) Functional
}

// Identified interface{}
//
// common type to address all symbols defined within the internal type system
// by numeric unique id.
type Identified interface{ Id() int }

// Named interface{}
//
// named interface implementations provide an internal name symbols can be
// addressed by. names are not unique!
//
// names get used in different contexts. the type system holds a name and
// string representation of it's signature for every type that has been
// defined. for type definitions that don't provide a distinct name (ad-hoc
// tuple type literals for instance), the typesignature will be the name those
// will be defined by. since the typesystem allows polymophism, multiple
// implementations with different id's may exist for every defined type name.
// usually at least a type constructor, to generate the types signature and a
// data constructor, to instanciate a value of the type.
type Named interface{ Name() string }

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
	Functional
	Left() Functional
	Right() Functional
	Both() (Functional, Functional)
}

// Typed interface{}
//
// internal higher order types have a name. due to polymorphism that name
// almost allways addresses multiple instances of the base data types, defined
// by the package
type Typed interface {
	Name() string
}

// SliceOfNatives interface{}
//
// possible gains in performance is the reason for providing slices in the
// first place. having slices of native instances, instead of slices of
// interface instances, makes this even more true. the actual slice types are
// defined in the data package and share methods with names chosen by
// convention and return types dependent on the native slice in particular, so
// names and return types can be inferred during 'writing the code time' by the
// programmer and hardcoded using assertion syntax (instance.(NatTypeVal)). not
// being 'generic' here, is not considered a problem, since all possible
// usecases of native slices, that could bring an actual gain in performance,
// involve knowing and discriminting by which type it's been dealt with anyway.
type SliceOfNatives interface {
	IndexedSlice
	Native(i int) interface{}
	Natives(i, j int) []interface{}
}

// 'APPLYABLE' DATA
//
// godeep aims for imutable values since data in functional programming is
// conceptionally mandatory to be imutable. go on the other hand is inherently
// procedual and relys on highly volatile mutable values, especially when
// dealing with things like inidices and loop counters. there is no way to keep
// code from mutating values on the syntax layer either. godeep provides a
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
	Functional
	ArgType() d.BitFlag
	Arg() Functional
	Apply(...Functional) (Functional, Argumented)
}

// Arguments interface{}
//
// interface all lists of arguments must implement to gain applyability
type Arguments interface {
	Functional
	Len() int
	Args() []Argumented
	Data() []Functional
	Get(int) Argumented
	Replace(int, Functional) Arguments
	Apply(...Functional) ([]Functional, Arguments)
}

// Parametric interface{}
//
// an argument with a distinct identifyer is a parameter. the identifyer can be
// a name, like for instance in a list of tupled function arguments, but also
// an integer denoting it's position in a tuple-, or slice, as well as a
// search, or sort praedicate that provides additional information for what to
// search-/, or sort by.
type Parametric interface {
	Functional
	Parm() Parametric
	Arg() Functional
	Acc() Functional
	Both() (Functional, Functional)
	Pair() Paired
	Apply(...Parametric) (Functional, Parametric)
}

// Parameters interface{}
//
// a list of instances of the parametric type.
type Parameters interface {
	Functional
	Len() int
	Pairs() []Paired
	Parms() []Parametric
	Get(Functional) Paired
	Replace(Paired) Parameters
	ReplaceKeyValue(k, v Functional) Parameters
	Apply(...Parametric) ([]Parametric, Parameters)
	AppendKeyValue(k, v Functional) Parameters
}

// Countable interface{}
//
// everything that has a length is countable. that's usually a collection but
// not neccessariely. A notable example of an excemption is the BitFlag, which
// can be concatenated from multiple flags by bitwise OR concatenation and
// provides a 'Len() int' method to check for, and a 'Deompose() []BitFlag'
// method to yield the components.
type Countable interface {
	Len() int // <- performs mutch better on slices
}

// Collected interface{}
//
// the collected interface is implemented by all collection types. they need to
// be countable and also provide a method to check if they're empty or not.
// that's neccessary, since the way to implement a performant empty check
// highly depends on the type (recursive vs. flat)
type Collected interface {
	Functional
	Countable
	Empty() bool //<-- no more nil pointers & 'out of index'!
}

// IndexedSlice interface{}
//
// the semantic concept of array indexing (T[i] -> T) doesn't play a huge role
// in functional programming, since data structures are commonly recursive and
// don't provide for element-wise access. go on the other hand highly relates
// on loops and slices that are accessable by index. to not loose a great deal
// of possible performance and ease of use, the concept of indexing has to be
// introduced. that's most probably 'inpure' in a mathematical sense, otherwise
// functional languages wouldn't go such long ways to circumvent it. godeep
// will provide element-wise accessable slices, while avoiding to use them at
// any point where it signifficantly could break abstrace concepts relied up on
// by functional programming‥.  this been said, we'll see how good this will
// turn out to work.
type IndexedSlice interface {
	Quantified
	Elem(i int) Functional
	Range(i, j int) []Functional
}

// Quantified interface{}
// is a collection type that can provide a flat slice representation of it's
// data
type Quantified interface {
	Collected
	Slice() []Functional //<-- no more nil pointers & 'out of index'!
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
type Vectorized interface {
	Quantified
	Head() Functional
	Tail() []Functional
	DeCap() (Functional, []Functional)
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
type Recorded interface {
	Vectorized
	Tuple() Tupled
	Get(Functional) Paired
	ArgSig() []Paired // key/data-type
}

//// RECURSIVE LISTS ///////
//
// recursive functional pendant to an array.
type Recursive interface {
	Functional
	Countable
	Empty() bool
	Head() Functional
	Tail() Recursive
	DeCap() (Functional, Recursive)
}

// LINKED LISTS
//
// Ordered interface{}
//j
// all recursive collections are inherently linked and may be accessed
// sequentially until depletion, to gain an ordered list. the ordered interface
// makes that explicit, provides a more convienient way to deal with it and/or
// allows slice based collections to be used in protty much the same way.
type Ordered interface {
	Collected
	Next() (Functional, Ordered)
}

// Reverseable interface{}
//
// to implement reversability a linked data structure needs to provide
// additional reference to it's predeccessor node
type Reverseable interface {
	Ordered
	Prev() Functional
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
	Collected
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
