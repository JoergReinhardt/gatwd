package data

import (
	"math/big"
)

type Stringer interface {
	String() string
}
type Flagged interface {
	Flag() BitFlag
}
type FlagTyped interface {
	FlagType() Uint8Val
}
type Matched interface {
	Match(Typed) bool
}
type NativeTyped interface {
	Type() TyNat
}
type NameTyped interface {
	TypeName() string
}

// typed needs to not have the NativeTyped interface, to stay interchangeable
// with types from other packages
type Typed interface {
	FlagTyped
	NameTyped
	Flagged
	Matched
	Stringer
}

// the main interface, all native types need to implement.
type Native interface {
	NativeTyped
	Stringer
	//	Type() Typed
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
	GoBytes() []byte
	Unit() Native
}

// unsignedVal and integerVal are a poor man's version of type classes and
// allow to treat the different sizes of ints and floats alike

type Boolean interface {
	Bool() BoolVal
	GoBool() bool
}

type Natural interface {
	Uint() UintVal
	GoUint() uint
}

type Integer interface {
	Int() IntVal
	GoInt() int
	Idx() int
}

type Rational interface {
	GoRat() *big.Rat
}

type Real interface {
	Float() FltVal
	GoFlt() float64
}

type Imaginary interface {
	Imag() ImagVal
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
	String() StrVal
}

type Serializeable interface {
	MarshalBinary() (BytesVal, error)
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
	Tail() DataSlice
	Shift() (Native, DataSlice)
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
	ElemType() Typed
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
	Has(acc Native) bool
	Get(acc Native) (Native, bool)
	Set(Native, Native) Mapped
	Delete(acc Native) bool
	KeyType() Typed
	ValType() Typed
}
