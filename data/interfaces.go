package data

// VALUES AND TYPES
///////////////////
// propertys intendet for internal use
type Reproduceable interface{ Copy() Data }
type Destructable interface{ Clear() }
type Stringer interface{ String() string }

//// SER DEFINED DATA & FUNCTION TYPES ///////
type Typed interface{ Flag() BitFlag }  //<- lowest common denominator
type DataTyped interface{ Flag() Type } //<- lowest common denominator
type DataType interface {
	Typed
}
type Data interface {
	Typed
	Stringer
}
type Accessor interface {
	Acc() Data
	Arg() Data
}
type Paired interface {
	Left() Data
	Right() Data
	Both() (Data, Data)
}
type Mapped interface {
	Keys() []Data
	Data() []Data
	Accs() []Paired
	Get(acc Data) Data
	Set(Data, Data) Mapped
}
type Ident interface {
	Ident() Data
}
type NativeVal interface {
	Data
	Null() func() Data
	DataFnc() func(Data) Data
}
type UnsignedVal interface{ Uint() uint }
type IntegerVal interface{ Int() int }
type Sliceable interface {
	Data
	Empty() bool
	Len() int
	Slice() []Data
}
type Vectorized interface {
	Data
	Len() int
	Empty() bool
	Slice() []Data
}
type NativeVec interface {
	Data
	Len() int
	Empty() bool
	Slice() []Data
}
type Collected interface {
	Data
	Empty() bool //<-- no more nil pointers & 'out of index'!
}
type Consumeable interface {
	Collected
	Head() Data
	Tail() Consumeable
	Shift() Consumeable
}
