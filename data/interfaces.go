package data

// VALUES AND TYPES
///////////////////
// propertys intendet for internal use
type Reproduceable interface{ Copy() Data }
type Destructable interface{ Clear() }
type Stringer interface{ String() string }

//// SER DEFINED DATA & FUNCTION TYPES ///////
type Data interface {
	Flag() BitFlag
	String() string
}
type Ident interface {
	Data
	Ident() Data
}
type Evaluable interface {
	Eval() Data
}

type Nullable interface {
	Data
	Null() Data
}
type Paired interface {
	Left() Data
	Right() Data
	Both() (Data, Data)
}
type Mapped interface {
	Data
	Keys() []Data
	Data() []Data
	Fields() []Paired
	Get(acc Data) (Data, bool)
	Set(Data, Data) Mapped
}
type UnsignedVal interface{ Uint() uint }
type IntegerVal interface{ Int() int }
type Collected interface {
	Data
	Empty() bool //<-- no more nil pointers & 'out of index'!
}
type Sliceable interface {
	Collected
	Len() int
	Slice() []Data
}
type Consumeable interface {
	Collected
	Head() Data
	Tail() Consumeable
	Shift() Consumeable
}
