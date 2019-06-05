package data

import "math/big"

//// USER DEFINED DATA & FUNCTION TYPES ///////
///
// all BitFlag's implement the typed interface (as well as primary)
type Typed interface {
	Native
	Flag() BitFlag
	FlagType() int8
	Match(Typed) bool
	TypeName() string
}

// the main interface, all types defined here need to comply to.
type Native interface {
	TypeNat() TyNat
	String() string
	Evaluable
}

type BinaryMarshaler interface {
	MarshalBinary() ([]byte, error)
}

// all data types are evaluable. evaluation yields a primary instance
type Evaluable interface {
	Eval(...Native) Native
}

// the identity function returns the instance unchanged
type Identity interface {
	Ident() Native
}

// deep copy
type Reproduceable interface {
	Copy() Native
}

// garbage collectability
type Destructable interface {
	Clear()
}

// implemented by types an empty instance is defined for
type Nullable interface {
	Null() Native
}
type Discrete interface {
	Unit() Native
}

// unsignedVal and integerVal are a poor man's version of type classes and
// allow to treat the different sizes of ints and floats alike

type Boolean interface {
	Bool() bool
}

type Natural interface {
	Uint() uint
}

type Integer interface {
	Int() int
}

type Rational interface {
	Rat() *big.Rat
}

type Real interface {
	Float() float64
}

type Imaginary interface {
	Imag() complex128
}

type Numeral interface {
	Native
	Nullable
	Discrete
	Boolean
	Natural
	Integer
	Rational
	Real
	Imaginary
	Integer() IntVal
}

type Raw interface {
	Bytes() []byte
}

type Letter interface {
	Rune() rune
	Byte() byte
}

type Text interface {
	String() string
}

type Serializeable interface {
	MarshalBinary() ([]byte, error)
}

type Printable interface {
	String() string
	Bytes() []byte
	Runes() []rune
}

// paired holds key-value pairs intendet as set accessors
type Paired interface {
	Native
	Left() Native
	Right() Native
	Both() (Native, Native)
}

// collections are expected nothing more, but to know, if they are empty
type Composed interface {
	Native
	Empty() bool //<-- no more nil pointers & 'out of index'!
}

// a slice know's it's length and can be represented in as indexable.
type Sliceable interface {
	Composed
	Len() int
	Slice() []Native
	Get(Native) Native
	GetInt(int) Native
	Range(s, e int) Sliceable
	Copy() Native
	Null() Native
}
type Mutable interface {
	Sliceable
	Set(s, arg Native)
	SetInt(int, Native)
}

// slices and set's convieniently 'mimic' the behaviour of linked list's common
// in functional programming.
type Sequential interface {
	Composed
	Head() Native
	Tail() Sequential
	Shift() Sequential
}

// mapped is the interface of all sets, that have accessors (index, or key)
type Mapped interface {
	Native
	Len() int
	Keys() []Native
	Data() []Native
	Fields() []Paired
	Get(acc Native) (Native, bool)
	Delete(acc Native) bool
	Set(Native, Native) Mapped
}
