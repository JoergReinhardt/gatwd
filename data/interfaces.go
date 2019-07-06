package data

import (
	"math/big"
)

// the main interface, all native types need to implement.
type Native interface {
	Eval() Native
	TypeNat() TyNat
	String() string
	TypeName() string
}

// all BitFlag's implement the typed interface (as well as primary)
type Typed interface {
	Native
	Flag() BitFlag
	FlagType() Uint8Val
	Match(Typed) bool
}

type BinaryMarshaler interface {
	MarshalBinary() ([]byte, error)
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
type Discrete interface {
	Unit() Native
}

// unsignedVal and integerVal are a poor man's version of type classes and
// allow to treat the different sizes of ints and floats alike

type Boolean interface {
	GoBool() bool
}

type Natural interface {
	GoUint() uint
}

type Integer interface {
	GoInt() int
	Idx() int
}

type Rational interface {
	GoRat() *big.Rat
}

type Real interface {
	GoFlt() float64
}

type Imaginary interface {
	GoImag() complex128
}

type Numeral interface {
	Native
	Natural
	Integer
	Rational
	Real
	Imaginary
}

type Raw interface {
	GoBytes() []byte
}

type Letter interface {
	GoRune() rune
	GoByte() byte
}

type Text interface {
	String() string
}

type Serializeable interface {
	MarshalBinary() ([]byte, error)
}

type Printable interface {
	String() string
	GoBytes() []byte
	GoRunes() []rune
}

// paired holds key-value pairs intendet as set accessors
type Paired interface {
	Native
	Left() Native
	Right() Native
	Both() (Native, Native)
	LeftType() TyNat
	RightType() TyNat
}

// collections are expected nothing more, but to know, if they are empty
type Composed interface {
	Native
	Empty() bool //<-- no more nil pointers & 'out of index'!
}

// a slice know's it's length and can be represented in as indexable.
type Sequential interface {
	Composed
	Head() Native
	Tail() Sequential
	Shift() Sequential
}
type Sliced interface {
	Slice() []Native
}
type Sliceable interface {
	Sliced
	Composed
	Len() int
	Copy() Native
	Get(Native) Native
	GetInt(int) Native
	Range(s, e int) Sliceable
	ElemType() TyNat
}

type Mutable interface {
	Sliceable
	Set(s, arg Native)
	SetInt(int, Native)
}

// mapped is the interface of all sets, that have accessors (index, or key)
type Mapped interface {
	Native
	Sliced
	Len() int
	Keys() []Native
	Data() []Native
	Fields() []Paired
	Get(acc Native) (Native, bool)
	Delete(acc Native) bool
	Set(Native, Native) Mapped
	KeyType() TyNat
	ValType() TyNat
}
