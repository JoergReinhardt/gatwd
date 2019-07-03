package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

//// NATIVE INTERFACE
///
type Flagged interface {
	Match(d.Typed) bool
	Flag() d.BitFlag
}
type Typed interface {
	Type() TyDef
	TypeFnc() TyFnc
	TypeName() string
	FlagType() d.Uint8Val
}

type Evaluable interface {
	Eval(...d.Native) d.Native
}

type Expression interface {
	Typed
	Call(...Expression) Expression
	Eval(...d.Native) d.Native
	TypeNat() d.TyNat
	String() string
}

type Reproduceable interface {
	// Copy() Native
	d.Reproduceable
}

//  garbage collectability
type Destructable interface {
	// Clear()
	d.Destructable
}

// numberal interfaces
type Boolean interface {
	// Bool() bool
	d.Boolean
}
type Discrete interface {
	// Unit()
	d.Discrete
}
type Natural interface {
	// Uint() uint
	d.Natural
}
type Integer interface {
	// Int() int
	d.Integer
}
type Rational interface {
	// Rat() *big.Rat
	d.Rational
}
type Real interface {
	// Float() float64
	d.Real
}
type Imaginary interface {
	// Imag() complex128
	d.Imaginary
}
type Numeral interface {
	Expression
	Discrete
	Boolean
	Natural
	Integer
	Rational
	Real
	Imaginary
}

type Text interface {
	// String() string
	d.Text
}

type Raw interface {
	// Bytes() []byte
	d.Raw
}

type Letter interface {
	// Rune() rune
	// Byte() byte
	d.Letter
}

type BinaryMarshaler interface {
	// MarshalBinary() ([]byte, error)
	d.BinaryMarshaler
}

type Serializeable interface {
	// MarshalBinary() ([]byte, error)
	d.Serializeable
}

type Printable interface {
	// String() string
	// Bytes() []byte
	// Runes() []rune
	d.Printable
}

type Verifyable interface {
	Bool() bool
}

type Composed interface {
	// Empty()
	d.Composed
}

type Slice interface {
	// Composed
	// Len() int
	// Slice() []Native
	d.Sliceable
}

type Sequential interface {
	// Head() Native
	// Tail() Sequential
	// Shift() Sequential
	d.Sequential
}

type Mapped interface {
	// Len() int
	// Keys() []Native
	// Data() []Native
	// Fields() []Paired
	// Get(acc Native) (Native, bool)
	// Delete(acc Native) bool
	// Set(Native, Native) Mapped
	d.Mapped
}

///////////////////////////////////////////////////////////////////////////////
//// COLLECTION INTERFACES
///
type Consumeable interface {
	Expression
	Head() Expression
	Tail() Consumeable
	Consume() (Expression, Consumeable)
}

type Countable interface {
	Len() int // <- performs mutch better on slices
}

type Listable interface {
	Consumeable
	Countable
}

type Sliceable interface {
	Slice() []Expression
}

type Sortable interface {
	Sort(d.TyNat)
}

type Searchable interface {
	Search(Expression) int
}

type IndexAssociated interface {
	Get(int) (Expression, bool)
	Set(int, Expression) (Vectorized, bool)
}

type Vectorized interface {
	IndexAssociated
	Searchable
	Sliceable
	Sortable
	Countable
}

type Swapable interface {
	Swap() (Expression, Expression)
}

type SwapablePaired interface {
	Swapable
	SwappedPair() Paired
}

type Ordered interface {
	Swapable
	Less(Expression) bool
}

type Associated interface {
	Key() Expression
	Value() Expression
}

type Associative interface {
	KeyNatType() d.TyNat
	ValNatType() d.TyNat
	KeyType() TyDef
	ValType() TyDef
}

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

type Keyed interface {
	KeyStr() string
}

type KeyPaired interface {
	Keyed
	Paired
}

type Indexed interface {
	Index() string
}
type IndexPaired interface {
	Indexed
	Paired
}
type ConsumeablePaired interface {
	Countable
	Consumeable
	Associative
	HeadPair() Paired
	TailPairs() ConsumeablePaired
	ConsumePair() (Paired, ConsumeablePaired)
}
type IndexablePaired interface {
	ConsumeablePaired
	IndexAssociated
}

type KeyAssociated interface {
	Pairs() []Paired
	GetVal(Expression) (Expression, bool)
	SetVal(Expression, Expression) (AssociativeCollected, bool)
}
type AssociativeCollected interface {
	KeyAssociated
	Consumeable
	Associative
	Countable
}

//// TREES & GRAPHS
///
type Nodular interface {
	Sliceable
	Root() Nodular
}

type Chained interface {
	Nodular
	Next() Nodular
}

type RevChained interface {
	Prev() Nodular
}

type Edged interface {
	Nodular
	From() Nodular
	To() Nodular
}

type Leaved interface {
	Nodular
	Value() Expression
}
