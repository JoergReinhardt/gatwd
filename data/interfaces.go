package data

// VALUES AND TYPES
///////////////////
// propertys intendet for internal use
type Reproduceable interface{ Copy() Data }
type Destructable interface{ Clear() }
type Stringer interface{ String() string }

//// USER DEFINED DATA & FUNCTION TYPES ///////
type Typed interface{ Flag() BitFlag }  //<- lowest common denominator
type DataTyped interface{ Flag() Type } //<- lowest common denominator
type DataType interface {
	Typed
}
type Data interface {
	Typed
	Stringer
	Eval() Data
}
type Sliceable interface {
	Data
	Empty() bool
	Len() int
	Slice() []Data
}
type NativeVal interface {
	Data
	Null() func() Data
	DataFnc() func(Data) Data
}
type Vector interface {
	Data
	Len() int
	Empty() bool
	Slice() []Data
	Elem(i int) Data
	Range(i, j int) []Data
}
type NativeVec interface {
	Data
	Len() int
	Empty() bool
	Slice() []Data
	NativeSlice() interface{}
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
