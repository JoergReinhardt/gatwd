package data

import (
	"fmt"
	"math/big"
	"time"
)

//////// INTERNAL TYPE SYSTEM ///////////
//
// intended to be accessable and extendable
type TyPrimitive BitFlag

func (v TyPrimitive) Flag() BitFlag { return BitFlag(v) }

func ListAllTypes() []TyPrimitive {
	var tt = []TyPrimitive{}
	var i uint
	var t TyPrimitive = 0
	for t < Flag {
		t = 1 << i
		i = i + 1
		tt = append(tt, TyPrimitive(t))
	}
	return tt
}

//go:generate stringer -type=TyPrimitive
const (
	Nil  TyPrimitive = 1
	Bool TyPrimitive = 1 << iota
	Int8             // Int8 -> Int8
	Int16
	Int32
	Int
	BigInt
	Uint8
	Uint16
	Uint32
	Uint
	Flt32
	Float
	BigFlt
	Ratio
	Imag64
	Imag
	Time
	Duration
	Byte
	Rune
	Bytes
	String
	Error // let's do something sophisticated here...
	//// HIGHERORDER TYPES
	Pair
	Tuple
	Record
	Vector
	List
	Set
	Argument
	Parameter
	Function
	Object
	Flag // marks most signifficant native type & data of type bitflag

	// TYPE CLASSES
	// precedence type classes define argument types functions that accept
	// a set of possible input types
	Nullables = Nil | Bool | Int8 | Int16 | Int32 |
		Int | BigInt | Uint8 | Uint16 | Uint32 | Uint |
		Flt32 | Float | BigFlt | Ratio | Imag64 | Imag |
		Time | Duration | Byte | Rune | Bytes | String

	Bitwise = Unsigned | Byte | Flag

	Boolean = Bool | Bitwise

	Unsigned = Uint | Uint8 | Uint16 | Uint32 | Byte

	Signed = Int | Int8 | Int16 | Int32 | BigInt | Byte

	Integer = Unsigned | Signed

	Rational = Unsigned | Integer | Ratio

	Irrational = Float | Flt32 | BigFlt

	Imaginary = Imag | Imag64

	Temporal = Time | Duration

	Textual = String | Rune | Bytes

	Numeric = Integer | Rational | Irrational |
		Imaginary | Temporal

	Symbolic = Textual | Boolean | Temporal | Error

	/// here will be dragons‥.
	HigherOrder = Functional | Collection

	Collection = Pair | Tuple | Record | Vector |
		List | Set

	Functional = Function | Argument | Parameter |
		Flag

	MAX_INT TyPrimitive = 0xFFFFFFFFFFFFFFFF
	Mask                = MAX_INT ^ Flag
)

//////// INTERNAL TYPES /////////////
// internal types are typealiases without any wrapping, or referencing getting
// in the way performancewise. types need to be aliased in the first place, to
// associate them with a bitflag, without having to actually asign, let alone
// attach it to the instance.
type ( ////// INTERNAL TYPES /////
	BitFlag uint
	////// TYPE ALIASES ///////
	NilVal    struct{}
	BoolVal   bool
	IntVal    int
	Int8Val   int8
	Int16Val  int16
	Int32Val  int32
	UintVal   uint
	Uint8Val  uint8
	Uint16Val uint16
	Uint32Val uint32
	FltVal    float64
	Flt32Val  float32
	ImagVal   complex128
	Imag64Val complex64
	ByteVal   byte
	RuneVal   rune
	BytesVal  []byte
	StrVal    string
	BigIntVal big.Int
	BigFltVal big.Float
	RatioVal  big.Rat
	TimeVal   time.Time
	DuraVal   time.Duration
	ErrorVal  struct{ e error }
	PairVal   struct{ l, r Data }
	FlagSlice []BitFlag
	DataSlice []Data
	SetString map[StrVal]Data
	SetUint   map[UintVal]Data
	SetInt    map[IntVal]Data
	SetFloat  map[FltVal]Data
	SetFlag   map[BitFlag]Data
)

//////// ATTRIBUTE TYPE ALIAS /////////////////

/// bind the appropriate Flag Method to every type
func (v BitFlag) Flag() BitFlag   { return Flag.Flag() }
func (v FlagSlice) Flag() BitFlag { return Flag.Flag() }
func (NilVal) Flag() BitFlag      { return Nil.Flag() }
func (v BoolVal) Flag() BitFlag   { return Bool.Flag() }
func (v IntVal) Flag() BitFlag    { return Int.Flag() }
func (v Int8Val) Flag() BitFlag   { return Int8.Flag() }
func (v Int16Val) Flag() BitFlag  { return Int16.Flag() }
func (v Int32Val) Flag() BitFlag  { return Int32.Flag() }
func (v UintVal) Flag() BitFlag   { return Uint.Flag() }
func (v Uint8Val) Flag() BitFlag  { return Uint8.Flag() }
func (v Uint16Val) Flag() BitFlag { return Uint16.Flag() }
func (v Uint32Val) Flag() BitFlag { return Uint32.Flag() }
func (v BigIntVal) Flag() BitFlag { return BigInt.Flag() }
func (v FltVal) Flag() BitFlag    { return Float.Flag() }
func (v Flt32Val) Flag() BitFlag  { return Flt32.Flag() }
func (v BigFltVal) Flag() BitFlag { return BigFlt.Flag() }
func (v ImagVal) Flag() BitFlag   { return Imag.Flag() }
func (v Imag64Val) Flag() BitFlag { return Imag64.Flag() }
func (v RatioVal) Flag() BitFlag  { return Ratio.Flag() }
func (v RuneVal) Flag() BitFlag   { return Rune.Flag() }
func (v ByteVal) Flag() BitFlag   { return Byte.Flag() }
func (v BytesVal) Flag() BitFlag  { return Bytes.Flag() }
func (v StrVal) Flag() BitFlag    { return String.Flag() }
func (v TimeVal) Flag() BitFlag   { return Time.Flag() }
func (v DuraVal) Flag() BitFlag   { return Duration.Flag() }
func (v ErrorVal) Flag() BitFlag  { return Error.Flag() }
func (v PairVal) Flag() BitFlag   { return Pair.Flag() }

///
func (NilVal) Copy() Data      { return NilVal{} }
func (v BitFlag) Copy() Data   { return BitFlag(v) }
func (v BoolVal) Copy() Data   { return BoolVal(v) }
func (v IntVal) Copy() Data    { return IntVal(v) }
func (v Int8Val) Copy() Data   { return Int8Val(v) }
func (v Int16Val) Copy() Data  { return Int16Val(v) }
func (v Int32Val) Copy() Data  { return Int32Val(v) }
func (v UintVal) Copy() Data   { return UintVal(v) }
func (v Uint8Val) Copy() Data  { return Uint8Val(v) }
func (v Uint16Val) Copy() Data { return Uint16Val(v) }
func (v Uint32Val) Copy() Data { return Uint32Val(v) }
func (v BigIntVal) Copy() Data { return BigIntVal(v) }
func (v FltVal) Copy() Data    { return FltVal(v) }
func (v Flt32Val) Copy() Data  { return Flt32Val(v) }
func (v BigFltVal) Copy() Data { return BigFltVal(v) }
func (v ImagVal) Copy() Data   { return ImagVal(v) }
func (v Imag64Val) Copy() Data { return Imag64Val(v) }
func (v RatioVal) Copy() Data  { return RatioVal(v) }
func (v RuneVal) Copy() Data   { return RuneVal(v) }
func (v ByteVal) Copy() Data   { return ByteVal(v) }
func (v BytesVal) Copy() Data  { return BytesVal(v) }
func (v StrVal) Copy() Data    { return StrVal(v) }
func (v TimeVal) Copy() Data   { return TimeVal(v) }
func (v DuraVal) Copy() Data   { return DuraVal(v) }
func (v ErrorVal) Copy() Data  { return ErrorVal(v) }
func (v PairVal) Copy() Data   { return PairVal{v.l, v.r} }
func (v FlagSlice) Copy() Data {
	var nfs = DataSlice{}
	for _, dat := range v {
		nfs = append(nfs, dat)
	}
	return nfs
}

///
func (NilVal) Eval(d ...Data) Data { return NilVal{} }
func (v BitFlag) Eval(d ...Data) Data {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnBoxed(v.Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v BoolVal) Eval(d ...Data) Data {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnBoxed(v.Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v IntVal) Eval(d ...Data) Data {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnBoxed(v.Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v Int8Val) Eval(d ...Data) Data {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnBoxed(v.Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v Int16Val) Eval(d ...Data) Data {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnBoxed(v.Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v Int32Val) Eval(d ...Data) Data {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnBoxed(v.Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v UintVal) Eval(d ...Data) Data {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnBoxed(v.Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v Uint8Val) Eval(d ...Data) Data {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnBoxed(v.Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v Uint16Val) Eval(d ...Data) Data {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnBoxed(v.Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v Uint32Val) Eval(d ...Data) Data {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnBoxed(v.Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v BigIntVal) Eval(d ...Data) Data {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnBoxed(v.Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v FltVal) Eval(d ...Data) Data {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnBoxed(v.Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v Flt32Val) Eval(d ...Data) Data {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnBoxed(v.Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v BigFltVal) Eval(d ...Data) Data {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnBoxed(v.Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v ImagVal) Eval(d ...Data) Data {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnBoxed(v.Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v Imag64Val) Eval(d ...Data) Data {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnBoxed(v.Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v RatioVal) Eval(d ...Data) Data {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnBoxed(v.Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v RuneVal) Eval(d ...Data) Data {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnBoxed(v.Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v ByteVal) Eval(d ...Data) Data {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnBoxed(v.Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v BytesVal) Eval(d ...Data) Data {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnBoxed(v.Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v StrVal) Eval(d ...Data) Data {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnBoxed(v.Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v TimeVal) Eval(d ...Data) Data {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnBoxed(v.Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v DuraVal) Eval(d ...Data) Data {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnBoxed(v.Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v ErrorVal) Eval(d ...Data) Data {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnBoxed(v.Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v PairVal) Eval(d ...Data) Data {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnBoxed(v.Flag(), d...)
		}
		return d[0]
	}
	return v
}

///
func (NilVal) Ident() Data      { return NilVal{} }
func (v BitFlag) Ident() Data   { return v }
func (v BoolVal) Ident() Data   { return v }
func (v IntVal) Ident() Data    { return v }
func (v Int8Val) Ident() Data   { return v }
func (v Int16Val) Ident() Data  { return v }
func (v Int32Val) Ident() Data  { return v }
func (v UintVal) Ident() Data   { return v }
func (v Uint8Val) Ident() Data  { return v }
func (v Uint16Val) Ident() Data { return v }
func (v Uint32Val) Ident() Data { return v }
func (v BigIntVal) Ident() Data { return v }
func (v FltVal) Ident() Data    { return v }
func (v Flt32Val) Ident() Data  { return v }
func (v BigFltVal) Ident() Data { return v }
func (v ImagVal) Ident() Data   { return v }
func (v Imag64Val) Ident() Data { return v }
func (v RatioVal) Ident() Data  { return v }
func (v RuneVal) Ident() Data   { return v }
func (v ByteVal) Ident() Data   { return v }
func (v BytesVal) Ident() Data  { return v }
func (v StrVal) Ident() Data    { return v }
func (v TimeVal) Ident() Data   { return v }
func (v DuraVal) Ident() Data   { return v }
func (v ErrorVal) Ident() Data  { return v }
func (v PairVal) Ident() Data   { return v }

//// native nullable typed ///////
func (v BitFlag) Null() Data   { return Nil.Flag() }
func (v FlagSlice) Null() Data { return NewFromNative(FlagSlice{}) }
func (v PairVal) Null() Data   { return PairVal{NilVal{}, NilVal{}} }
func (v NilVal) Null() Data    { return NilVal{} }
func (v BoolVal) Null() Data   { return NewFromNative(false) }
func (v IntVal) Null() Data    { return NewFromNative(0) }
func (v Int8Val) Null() Data   { return NewFromNative(int8(0)) }
func (v Int16Val) Null() Data  { return NewFromNative(int16(0)) }
func (v Int32Val) Null() Data  { return NewFromNative(int32(0)) }
func (v UintVal) Null() Data   { return NewFromNative(uint(0)) }
func (v Uint8Val) Null() Data  { return NewFromNative(uint8(0)) }
func (v Uint16Val) Null() Data { return NewFromNative(uint16(0)) }
func (v Uint32Val) Null() Data { return NewFromNative(uint32(0)) }
func (v FltVal) Null() Data    { return NewFromNative(0.0) }
func (v Flt32Val) Null() Data  { return NewFromNative(float32(0.0)) }
func (v ImagVal) Null() Data   { return NewFromNative(complex128(0.0)) }
func (v Imag64Val) Null() Data { return NewFromNative(complex64(0.0)) }
func (v ByteVal) Null() Data   { return NewFromNative(byte(0)) }
func (v BytesVal) Null() Data  { return NewFromNative([]byte{}) }
func (v RuneVal) Null() Data   { return NewFromNative(rune(' ')) }
func (v StrVal) Null() Data    { return NewFromNative(string("")) }
func (v ErrorVal) Null() Data  { return NewFromNative(error(fmt.Errorf(""))) }
func (v BigIntVal) Null() Data { return NewFromNative(big.NewInt(0)) }
func (v BigFltVal) Null() Data { return NewFromNative(big.NewFloat(0)) }
func (v RatioVal) Null() Data  { return NewFromNative(big.NewRat(1, 1)) }
func (v TimeVal) Null() Data   { return NewFromNative(time.Now()) }
func (v DuraVal) Null() Data   { return NewFromNative(time.Duration(0)) }
