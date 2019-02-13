package data

//// USER DEFINED DATA & FUNCTION TYPES ///////
///
// all BitFlag's implement the typed interface (as well as primary)
type Typed interface{ Flag() BitFlag }

// the main interface, all types defined here need to comply to.
type Primary interface {
	TypePrime() TyPrime
	String() string
	Evaluable
}
type BinaryMarshaler interface {
	MarshalBinary() ([]byte, error)
}

// all data types are evaluable. evaluation yields a primary instance
type Evaluable interface{ Eval(...Primary) Primary }

// the identity function returns the instance unchanged
type Identity interface{ Ident() Primary }

// deep copy
type Reproduceable interface{ Copy() Primary }

// garbage collectability
type Destructable interface{ Clear() }

// implemented by types an empty instance is defined for
type Nullable interface{ Null() Primary }

// unsignedVal and integerVal are a poor man's version of type classes and
// allow to treat the different sizes of ints and floats alike
type NaturalVal interface{ Uint() uint }
type IntegerVal interface{ Int() int }
type RealVal interface{ Float() float64 }
type SymbolicVal interface{ String() string }

// paired holds key-value pairs intendet as set accessors
type Paired interface {
	Primary
	Left() Primary
	Right() Primary
	Both() (Primary, Primary)
}

// collections are expected to know, if they are empty
type Collected interface {
	Primary
	Empty() bool //<-- no more nil pointers & 'out of index'!
}

// a slice know's it's length and can be represented in as indexable.
type Sliceable interface {
	Collected
	Len() int
	Slice() []Primary
}

// slices and set's convieniently 'mimic' the behaviour of linked list's common
// in functional programming.
type Consumeable interface {
	Collected
	Head() Primary
	Tail() Consumeable
	Shift() Consumeable
}

// mapped is the interface of all sets, that have accessors (index, or key)
type Mapped interface {
	Primary
	Keys() []Primary
	Data() []Primary
	Fields() []Paired
	Get(acc Primary) (Primary, bool)
	Set(Primary, Primary) Mapped
}
