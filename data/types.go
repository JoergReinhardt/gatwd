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

func (v TyPrimitive) TypePrim() BitFlag { return BitFlag(v) }
func (v TyPrimitive) Flag() BitFlag     { return BitFlag(v) }

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
	Primarys = Nil | Bool | Int8 | Int16 | Int32 |
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
	PairVal   struct{ l, r Primary }
	FlagSlice []BitFlag
	DataSlice []Primary
	SetString map[StrVal]Primary
	SetUint   map[UintVal]Primary
	SetInt    map[IntVal]Primary
	SetFloat  map[FltVal]Primary
	SetFlag   map[BitFlag]Primary
)

//////// ATTRIBUTE TYPE ALIAS /////////////////

/// bind the appropriate TypePrim Method to every type
func (v BitFlag) TypePrim() BitFlag   { return Flag.TypePrim() }
func (v FlagSlice) Flag() BitFlag     { return Flag.TypePrim() }
func (NilVal) TypePrim() BitFlag      { return Nil.TypePrim() }
func (v BoolVal) TypePrim() BitFlag   { return Bool.TypePrim() }
func (v IntVal) TypePrim() BitFlag    { return Int.TypePrim() }
func (v Int8Val) TypePrim() BitFlag   { return Int8.TypePrim() }
func (v Int16Val) TypePrim() BitFlag  { return Int16.TypePrim() }
func (v Int32Val) TypePrim() BitFlag  { return Int32.TypePrim() }
func (v UintVal) TypePrim() BitFlag   { return Uint.TypePrim() }
func (v Uint8Val) TypePrim() BitFlag  { return Uint8.TypePrim() }
func (v Uint16Val) TypePrim() BitFlag { return Uint16.TypePrim() }
func (v Uint32Val) TypePrim() BitFlag { return Uint32.TypePrim() }
func (v BigIntVal) TypePrim() BitFlag { return BigInt.TypePrim() }
func (v FltVal) TypePrim() BitFlag    { return Float.TypePrim() }
func (v Flt32Val) TypePrim() BitFlag  { return Flt32.TypePrim() }
func (v BigFltVal) TypePrim() BitFlag { return BigFlt.TypePrim() }
func (v ImagVal) TypePrim() BitFlag   { return Imag.TypePrim() }
func (v Imag64Val) TypePrim() BitFlag { return Imag64.TypePrim() }
func (v RatioVal) TypePrim() BitFlag  { return Ratio.TypePrim() }
func (v RuneVal) TypePrim() BitFlag   { return Rune.TypePrim() }
func (v ByteVal) TypePrim() BitFlag   { return Byte.TypePrim() }
func (v BytesVal) TypePrim() BitFlag  { return Bytes.TypePrim() }
func (v StrVal) TypePrim() BitFlag    { return String.TypePrim() }
func (v TimeVal) TypePrim() BitFlag   { return Time.TypePrim() }
func (v DuraVal) TypePrim() BitFlag   { return Duration.TypePrim() }
func (v ErrorVal) TypePrim() BitFlag  { return Error.TypePrim() }
func (v PairVal) TypePrim() BitFlag   { return Pair.TypePrim() }

///
func (NilVal) Copy() Primary      { return NilVal{} }
func (v BitFlag) Copy() Primary   { return BitFlag(v) }
func (v BoolVal) Copy() Primary   { return BoolVal(v) }
func (v IntVal) Copy() Primary    { return IntVal(v) }
func (v Int8Val) Copy() Primary   { return Int8Val(v) }
func (v Int16Val) Copy() Primary  { return Int16Val(v) }
func (v Int32Val) Copy() Primary  { return Int32Val(v) }
func (v UintVal) Copy() Primary   { return UintVal(v) }
func (v Uint8Val) Copy() Primary  { return Uint8Val(v) }
func (v Uint16Val) Copy() Primary { return Uint16Val(v) }
func (v Uint32Val) Copy() Primary { return Uint32Val(v) }
func (v BigIntVal) Copy() Primary { return BigIntVal(v) }
func (v FltVal) Copy() Primary    { return FltVal(v) }
func (v Flt32Val) Copy() Primary  { return Flt32Val(v) }
func (v BigFltVal) Copy() Primary { return BigFltVal(v) }
func (v ImagVal) Copy() Primary   { return ImagVal(v) }
func (v Imag64Val) Copy() Primary { return Imag64Val(v) }
func (v RatioVal) Copy() Primary  { return RatioVal(v) }
func (v RuneVal) Copy() Primary   { return RuneVal(v) }
func (v ByteVal) Copy() Primary   { return ByteVal(v) }
func (v BytesVal) Copy() Primary  { return BytesVal(v) }
func (v StrVal) Copy() Primary    { return StrVal(v) }
func (v TimeVal) Copy() Primary   { return TimeVal(v) }
func (v DuraVal) Copy() Primary   { return DuraVal(v) }
func (v ErrorVal) Copy() Primary  { return ErrorVal(v) }
func (v PairVal) Copy() Primary   { return PairVal{v.l, v.r} }
func (v FlagSlice) Copy() Primary {
	var nfs = DataSlice{}
	for _, dat := range v {
		nfs = append(nfs, dat)
	}
	return nfs
}

///
func (NilVal) Eval(d ...Primary) Primary { return NilVal{} }
func (v BitFlag) Eval(d ...Primary) Primary {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnBoxed(v.TypePrim(), d...)
		}
		return d[0]
	}
	return v
}
func (v BoolVal) Eval(d ...Primary) Primary {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnBoxed(v.TypePrim(), d...)
		}
		return d[0]
	}
	return v
}
func (v IntVal) Eval(d ...Primary) Primary {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnBoxed(v.TypePrim(), d...)
		}
		return d[0]
	}
	return v
}
func (v Int8Val) Eval(d ...Primary) Primary {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnBoxed(v.TypePrim(), d...)
		}
		return d[0]
	}
	return v
}
func (v Int16Val) Eval(d ...Primary) Primary {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnBoxed(v.TypePrim(), d...)
		}
		return d[0]
	}
	return v
}
func (v Int32Val) Eval(d ...Primary) Primary {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnBoxed(v.TypePrim(), d...)
		}
		return d[0]
	}
	return v
}
func (v UintVal) Eval(d ...Primary) Primary {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnBoxed(v.TypePrim(), d...)
		}
		return d[0]
	}
	return v
}
func (v Uint8Val) Eval(d ...Primary) Primary {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnBoxed(v.TypePrim(), d...)
		}
		return d[0]
	}
	return v
}
func (v Uint16Val) Eval(d ...Primary) Primary {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnBoxed(v.TypePrim(), d...)
		}
		return d[0]
	}
	return v
}
func (v Uint32Val) Eval(d ...Primary) Primary {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnBoxed(v.TypePrim(), d...)
		}
		return d[0]
	}
	return v
}
func (v BigIntVal) Eval(d ...Primary) Primary {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnBoxed(v.TypePrim(), d...)
		}
		return d[0]
	}
	return v
}
func (v FltVal) Eval(d ...Primary) Primary {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnBoxed(v.TypePrim(), d...)
		}
		return d[0]
	}
	return v
}
func (v Flt32Val) Eval(d ...Primary) Primary {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnBoxed(v.TypePrim(), d...)
		}
		return d[0]
	}
	return v
}
func (v BigFltVal) Eval(d ...Primary) Primary {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnBoxed(v.TypePrim(), d...)
		}
		return d[0]
	}
	return v
}
func (v ImagVal) Eval(d ...Primary) Primary {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnBoxed(v.TypePrim(), d...)
		}
		return d[0]
	}
	return v
}
func (v Imag64Val) Eval(d ...Primary) Primary {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnBoxed(v.TypePrim(), d...)
		}
		return d[0]
	}
	return v
}
func (v RatioVal) Eval(d ...Primary) Primary {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnBoxed(v.TypePrim(), d...)
		}
		return d[0]
	}
	return v
}
func (v RuneVal) Eval(d ...Primary) Primary {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnBoxed(v.TypePrim(), d...)
		}
		return d[0]
	}
	return v
}
func (v ByteVal) Eval(d ...Primary) Primary {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnBoxed(v.TypePrim(), d...)
		}
		return d[0]
	}
	return v
}
func (v BytesVal) Eval(d ...Primary) Primary {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnBoxed(v.TypePrim(), d...)
		}
		return d[0]
	}
	return v
}
func (v StrVal) Eval(d ...Primary) Primary {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnBoxed(v.TypePrim(), d...)
		}
		return d[0]
	}
	return v
}
func (v TimeVal) Eval(d ...Primary) Primary {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnBoxed(v.TypePrim(), d...)
		}
		return d[0]
	}
	return v
}
func (v DuraVal) Eval(d ...Primary) Primary {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnBoxed(v.TypePrim(), d...)
		}
		return d[0]
	}
	return v
}
func (v ErrorVal) Eval(d ...Primary) Primary {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnBoxed(v.TypePrim(), d...)
		}
		return d[0]
	}
	return v
}
func (v PairVal) Eval(d ...Primary) Primary {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnBoxed(v.TypePrim(), d...)
		}
		return d[0]
	}
	return v
}

///
func (NilVal) Ident() Primary      { return NilVal{} }
func (v BitFlag) Ident() Primary   { return v }
func (v BoolVal) Ident() Primary   { return v }
func (v IntVal) Ident() Primary    { return v }
func (v Int8Val) Ident() Primary   { return v }
func (v Int16Val) Ident() Primary  { return v }
func (v Int32Val) Ident() Primary  { return v }
func (v UintVal) Ident() Primary   { return v }
func (v Uint8Val) Ident() Primary  { return v }
func (v Uint16Val) Ident() Primary { return v }
func (v Uint32Val) Ident() Primary { return v }
func (v BigIntVal) Ident() Primary { return v }
func (v FltVal) Ident() Primary    { return v }
func (v Flt32Val) Ident() Primary  { return v }
func (v BigFltVal) Ident() Primary { return v }
func (v ImagVal) Ident() Primary   { return v }
func (v Imag64Val) Ident() Primary { return v }
func (v RatioVal) Ident() Primary  { return v }
func (v RuneVal) Ident() Primary   { return v }
func (v ByteVal) Ident() Primary   { return v }
func (v BytesVal) Ident() Primary  { return v }
func (v StrVal) Ident() Primary    { return v }
func (v TimeVal) Ident() Primary   { return v }
func (v DuraVal) Ident() Primary   { return v }
func (v ErrorVal) Ident() Primary  { return v }
func (v PairVal) Ident() Primary   { return v }

//// native nullable typed ///////
func (v BitFlag) Null() Primary   { return Nil.TypePrim() }
func (v FlagSlice) Null() Primary { return NewFromNative(FlagSlice{}) }
func (v PairVal) Null() Primary   { return PairVal{NilVal{}, NilVal{}} }
func (v NilVal) Null() Primary    { return NilVal{} }
func (v BoolVal) Null() Primary   { return NewFromNative(false) }
func (v IntVal) Null() Primary    { return NewFromNative(0) }
func (v Int8Val) Null() Primary   { return NewFromNative(int8(0)) }
func (v Int16Val) Null() Primary  { return NewFromNative(int16(0)) }
func (v Int32Val) Null() Primary  { return NewFromNative(int32(0)) }
func (v UintVal) Null() Primary   { return NewFromNative(uint(0)) }
func (v Uint8Val) Null() Primary  { return NewFromNative(uint8(0)) }
func (v Uint16Val) Null() Primary { return NewFromNative(uint16(0)) }
func (v Uint32Val) Null() Primary { return NewFromNative(uint32(0)) }
func (v FltVal) Null() Primary    { return NewFromNative(0.0) }
func (v Flt32Val) Null() Primary  { return NewFromNative(float32(0.0)) }
func (v ImagVal) Null() Primary   { return NewFromNative(complex128(0.0)) }
func (v Imag64Val) Null() Primary { return NewFromNative(complex64(0.0)) }
func (v ByteVal) Null() Primary   { return NewFromNative(byte(0)) }
func (v BytesVal) Null() Primary  { return NewFromNative([]byte{}) }
func (v RuneVal) Null() Primary   { return NewFromNative(rune(' ')) }
func (v StrVal) Null() Primary    { return NewFromNative(string("")) }
func (v ErrorVal) Null() Primary  { return NewFromNative(error(fmt.Errorf(""))) }
func (v BigIntVal) Null() Primary { return NewFromNative(big.NewInt(0)) }
func (v BigFltVal) Null() Primary { return NewFromNative(big.NewFloat(0)) }
func (v RatioVal) Null() Primary  { return NewFromNative(big.NewRat(1, 1)) }
func (v TimeVal) Null() Primary   { return NewFromNative(time.Now()) }
func (v DuraVal) Null() Primary   { return NewFromNative(time.Duration(0)) }
