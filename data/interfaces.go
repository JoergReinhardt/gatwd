package data

// VALUES AND TYPES
///////////////////
// propertys intendet for internal use
type Reproduceable interface{ Copy() Primary }
type Destructable interface{ Clear() }
type Stringer interface{ String() string }

//// SER DEFINED DATA & FUNCTION TYPES ///////
type Primary interface {
	TypePrim() BitFlag
	String() string
}
type Ident interface {
	Primary
	Ident() Primary
}
type Evaluable interface {
	Eval() Primary
}

type Nullable interface {
	Primary
	Null() Primary
}
type Paired interface {
	Primary
	Left() Primary
	Right() Primary
	Both() (Primary, Primary)
}
type Mapped interface {
	Primary
	Keys() []Primary
	Data() []Primary
	Fields() []Paired
	Get(acc Primary) (Primary, bool)
	Set(Primary, Primary) Mapped
}
type UnsignedVal interface{ Uint() uint }
type IntegerVal interface{ Int() int }
type Collected interface {
	Primary
	Empty() bool //<-- no more nil pointers & 'out of index'!
}
type Sliceable interface {
	Collected
	Len() int
	Slice() []Primary
}
type Consumeable interface {
	Collected
	Head() Primary
	Tail() Consumeable
	Shift() Consumeable
}
