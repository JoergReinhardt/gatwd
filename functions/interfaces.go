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
type Native interface {
	Expression
	TypeNat() d.TyNat
	Eval(...d.Native) d.Native
}
type Expression interface {
	FunctionTyped
	Callable
	Stringer
	Type() TyPattern
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
	Consume() (Expression, Consumeable)
	Head() Expression
	Tail() Consumeable
}
type Sequential interface {
	Consumeable
	TypeElem() TyPattern
	Append(...Expression) Sequential
	Cons(...Expression) Sequential
}

type Testable interface {
	Expression
	Test(...Expression) bool
}

type Compareable interface {
	Expression
	Compare(...Expression) int
}

type Countable interface {
	Len() int
}

type Listable interface {
	Sequential
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
	TypeKey() d.Typed
	TypeValue() d.Typed
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
	Sequential
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
	Sequential
	Associative
}
type Enumerable interface {
	Expression
	Next() EnumVal
	Prev() EnumVal
	EnumType() EnumType
	Alloc(d.Numeral) EnumVal
}
type Monadic interface {
	Expression
	Current() Expression
	Step(...Expression) (Expression, Monadic)
	Sequence() Sequential
}
type ProtoTyped interface {
	Expression
	Name() string
	Methods() []string
	CallMethod(string, ...Expression) Expression
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
