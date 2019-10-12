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
type Matched interface {
	Match(Typed) bool
}
type NativeTyped interface {
	Type() TyNat
}
type NameOfType interface {
	TypeName() string
}

// typed needs to not have the NativeTyped interface, to stay interchangeable
// with types from other packages
type Typed interface {
	KindOfFlag
	NameOfType
	Flagged
	Matched
	Stringer
}
type KindOfFlag interface {
	Kind() Uint8Val
}

// the main interface, all native types need to implement.
type Native interface {
	NativeTyped
	Stringer
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
type NaturalAriOps interface {
	Inc() Natural
	Dec() Natural
	AddU(arg UintVal) UintVal
	SubstractU(arg UintVal) UintVal
	MultipyU(arg UintVal) UintVal
	PowerU(arg UintVal) UintVal
	QuotientU(arg UintVal) UintVal
	QuoRatioU(arg UintVal) *RatioVal
}
type NaturalBoolOps interface {
	NotU() UintVal
	AndU(arg UintVal) UintVal
	XorU(arg UintVal) UintVal
	OrU(arg UintVal) UintVal
	AndNotU(arg UintVal) UintVal
}
type NaturalComparators interface {
	EqualU(arg UintVal) bool
	LesserU(arg UintVal) bool
	GreaterU(arg UintVal) bool
}

type Integer interface {
	Int() IntVal
	BigInt() *BigIntVal
	GoInt() int
	Idx() int
}
type IntegerAriOps interface {
	Inc() Integer
	Dec() Integer
	AddI(arg UintVal) UintVal
	SubstractI(arg UintVal) UintVal
	MultipyI(arg UintVal) UintVal
	PowerI(arg UintVal) UintVal
	QuotientI(arg UintVal) UintVal
	QuoRatioI(arg UintVal) *RatioVal
}
type IntegerBoolOps interface {
	NotI() IntVal
	AndI(arg IntVal) IntVal
	XorI(arg IntVal) IntVal
	OrI(arg IntVal) IntVal
	AndNotI(arg IntVal) IntVal
}
type IntegerComparators interface {
	EqualI(arg IntVal) bool
	LesserI(arg IntVal) bool
	GreaterI(arg IntVal) bool
}

type Real interface {
	Float() FltVal
	GoFlt() float64
}
type RealOps interface {
	NegateR() FltVal
	AddR(arg FltVal) FltVal
	SubstractR(arg FltVal) FltVal
	MultipyR(arg FltVal) FltVal
	QuotientR(arg FltVal) FltVal
}
type RealComparators interface {
	EqualR(arg FltVal) bool
	LesserR(arg FltVal) bool
	GreaterR(arg FltVal) bool
}

type Rational interface {
	GoRat() *big.Rat
}
type RationalOps interface {
	Negate() *RatioVal
	Invert() *RatioVal
	Add(arg *RatioVal) *RatioVal
	Substract(arg *RatioVal) *RatioVal
	Multipy(arg *RatioVal) *RatioVal
	Quotient(arg *RatioVal) *RatioVal
}
type RationalComparators interface {
	Lesser(arg *RatioVal) bool
	Greater(arg *RatioVal) bool
	Equal(arg *RatioVal) bool
}

type Imaginary interface {
	Imag() ImagVal
	GoImag() complex128
}

type Numeral interface {
	Native
	Natural
	Integer
	Real
	Rational
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
	TypeKey() TyNat
	TypeValue() TyNat
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

// indexable sequence of native instances
type Sliced interface {
	Slice() []Native
}

// indexable sequence methods
type Sliceable interface {
	Sliced
	Composed
	Len() int
	Copy() Native
	Get(Native) Native
	GetInt(int) Native
	Range(s, e int) Sliceable
	TypeElem() Typed
}

// data mutability interface
type Mutable interface {
	Sliceable
	Set(s, arg Native)
	SetInt(int, Native)
}

// mapped is the interface of all key accessable hash maps
type Mapped interface {
	Native
	Sliced
	Len() int
	First() Paired
	Keys() []Native
	Data() []Native
	Fields() []Paired
	Has(acc Native) bool
	Get(acc Native) (Native, bool)
	Set(Native, Native) Mapped
	Delete(acc Native) bool
	TypeKey() Typed
	TypeValue() Typed
}
