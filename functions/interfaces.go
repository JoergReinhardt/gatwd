package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

///// EXPRESSION INTERFACES
type Stringer interface {
	String() string
}
type FlagTyped interface {
	FlagType() d.Uint8Val
}
type Flagged interface {
	Flag() d.BitFlag
}
type Matched interface {
	Match(d.Typed) bool
}
type NameTyped interface {
	TypeName() string
}
type FunctionTyped interface {
	TypeFnc() TyFnc
}
type NativeTyped interface {
	TypeNat() d.TyNat
}
type Callable interface {
	Call(...Expression) Expression
}
type Typed interface {
	d.Typed
}
type Expression interface {
	FunctionTyped
	FlagTyped
	Callable
	Stringer
	Type() d.Typed
}
type Native interface {
	FunctionTyped
	FlagTyped
	Callable
	d.Native
	Eval(...d.Native) d.Native
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

// numberal interfacesName
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

type Textual interface {
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
	Len() int
	Keys() []Expression
	Data() []Expression
	Fields() []Paired
	Get(acc Expression) (Expression, bool)
	Delete(acc Expression) bool
	Set(Expression, Expression) Mapped
	//d.Mapped
}

///////////////////////////////////////////////////////////////////////////////
//// COLLECTION INTERFACES
///
type Consumeable interface {
	Expression
	TypeElem() d.Typed
	Head() Expression
	Tail() Consumeable
	Consume() (Expression, Consumeable)
}

type Countable interface {
	Len() int
}

type Listable interface {
	Consumeable
	Countable
}

type Sliceable interface {
	Slice() []Expression
}

type Sortable interface {
	Sort(TyFnc)
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
	//Searchable
	Sliceable
	//Sortable
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
	KeyType() d.Typed
	ValType() d.Typed
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
	KeyStr() d.StrVal
}

type KeyPaired interface {
	Keyed
	Paired
}

type Indexed interface {
	Index() int
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
