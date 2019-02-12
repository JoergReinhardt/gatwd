package data

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/big"
	"time"
)

//////// INTERNAL TYPE SYSTEM ///////////
//
// intended to be accessable and extendable
type TyPrimitive BitFlag

func (v TyPrimitive) TypePrim() TyPrimitive { return v }
func (v TyPrimitive) Eval(p ...Primary) Primary {
	if len(p) > 0 {
		for _, prime := range p {
			if prime.TypePrim().Flag().Match(Flag) {
				var flag = prime.(BitFlag)
				v = v | flag.TypePrim()
			}
		}
	}
	return v
}
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
	Function
	Flag // marks most signifficant native type & data of type bitflag

	// TYPE CLASSES
	// precedence type classes define argument types functions that accept
	// a set of possible input types
	Primarys = Nil | Bool | Int8 | Int16 | Int32 | Int | BigInt | Uint8 |
		Uint16 | Uint32 | Uint | Flt32 | Float | BigFlt | Ratio | Imag64 |
		Imag | Time | Duration | Byte | Rune | Bytes | String | Error

	Bitwise = Natural | Byte | Flag

	Boolean = Bool | Bitwise

	Natural = Uint | Uint8 | Uint16 | Uint32

	Integer = Int | Int8 | Int16 | Int32 | BigInt

	Rational = Natural | Ratio

	Real = Float | Flt32 | BigFlt

	Imaginary = Imag | Imag64

	Temporal = Time | Duration

	Textual = String | Rune | Bytes

	Numeric = Rational | Real | Imaginary | Temporal

	Symbolic = Textual | Boolean | Temporal | Error

	Collection = Pair | Tuple | Record | Vector | List | Set

	/// here will be dragons‥.
	HigherOrder = Function | Collection

	MAX_INT     TyPrimitive = 0xFFFFFFFFFFFFFFFF
	MaskPrimary             = MAX_INT ^ Primarys
	Mask                    = MAX_INT ^ Flag
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
func (v BitFlag) TypePrim() TyPrimitive   { return Flag.TypePrim() }
func (v FlagSlice) Flag() TyPrimitive     { return Flag.TypePrim() }
func (NilVal) TypePrim() TyPrimitive      { return Nil.TypePrim() }
func (v BoolVal) TypePrim() TyPrimitive   { return Bool.TypePrim() }
func (v IntVal) TypePrim() TyPrimitive    { return Int.TypePrim() }
func (v Int8Val) TypePrim() TyPrimitive   { return Int8.TypePrim() }
func (v Int16Val) TypePrim() TyPrimitive  { return Int16.TypePrim() }
func (v Int32Val) TypePrim() TyPrimitive  { return Int32.TypePrim() }
func (v UintVal) TypePrim() TyPrimitive   { return Uint.TypePrim() }
func (v Uint8Val) TypePrim() TyPrimitive  { return Uint8.TypePrim() }
func (v Uint16Val) TypePrim() TyPrimitive { return Uint16.TypePrim() }
func (v Uint32Val) TypePrim() TyPrimitive { return Uint32.TypePrim() }
func (v BigIntVal) TypePrim() TyPrimitive { return BigInt.TypePrim() }
func (v FltVal) TypePrim() TyPrimitive    { return Float.TypePrim() }
func (v Flt32Val) TypePrim() TyPrimitive  { return Flt32.TypePrim() }
func (v BigFltVal) TypePrim() TyPrimitive { return BigFlt.TypePrim() }
func (v ImagVal) TypePrim() TyPrimitive   { return Imag.TypePrim() }
func (v Imag64Val) TypePrim() TyPrimitive { return Imag64.TypePrim() }
func (v RatioVal) TypePrim() TyPrimitive  { return Ratio.TypePrim() }
func (v RuneVal) TypePrim() TyPrimitive   { return Rune.TypePrim() }
func (v ByteVal) TypePrim() TyPrimitive   { return Byte.TypePrim() }
func (v BytesVal) TypePrim() TyPrimitive  { return Bytes.TypePrim() }
func (v StrVal) TypePrim() TyPrimitive    { return String.TypePrim() }
func (v TimeVal) TypePrim() TyPrimitive   { return Time.TypePrim() }
func (v DuraVal) TypePrim() TyPrimitive   { return Duration.TypePrim() }
func (v ErrorVal) TypePrim() TyPrimitive  { return Error.TypePrim() }
func (v PairVal) TypePrim() TyPrimitive   { return Pair.TypePrim() }

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
			return NewUnboxedVector(v.TypePrim().Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v BoolVal) Eval(d ...Primary) Primary {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnboxedVector(v.TypePrim().Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v IntVal) Eval(d ...Primary) Primary {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnboxedVector(v.TypePrim().Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v Int8Val) Eval(d ...Primary) Primary {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnboxedVector(v.TypePrim().Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v Int16Val) Eval(d ...Primary) Primary {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnboxedVector(v.TypePrim().Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v Int32Val) Eval(d ...Primary) Primary {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnboxedVector(v.TypePrim().Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v UintVal) Eval(d ...Primary) Primary {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnboxedVector(v.TypePrim().Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v Uint8Val) Eval(d ...Primary) Primary {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnboxedVector(v.TypePrim().Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v Uint16Val) Eval(d ...Primary) Primary {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnboxedVector(v.TypePrim().Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v Uint32Val) Eval(d ...Primary) Primary {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnboxedVector(v.TypePrim().Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v BigIntVal) Eval(d ...Primary) Primary {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnboxedVector(v.TypePrim().Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v FltVal) Eval(d ...Primary) Primary {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnboxedVector(v.TypePrim().Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v Flt32Val) Eval(d ...Primary) Primary {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnboxedVector(v.TypePrim().Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v BigFltVal) Eval(d ...Primary) Primary {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnboxedVector(v.TypePrim().Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v ImagVal) Eval(d ...Primary) Primary {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnboxedVector(v.TypePrim().Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v Imag64Val) Eval(d ...Primary) Primary {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnboxedVector(v.TypePrim().Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v RatioVal) Eval(d ...Primary) Primary {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnboxedVector(v.TypePrim().Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v RuneVal) Eval(d ...Primary) Primary {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnboxedVector(v.TypePrim().Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v ByteVal) Eval(d ...Primary) Primary {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnboxedVector(v.TypePrim().Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v BytesVal) Eval(d ...Primary) Primary {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnboxedVector(v.TypePrim().Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v StrVal) Eval(d ...Primary) Primary {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnboxedVector(v.TypePrim().Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v TimeVal) Eval(d ...Primary) Primary {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnboxedVector(v.TypePrim().Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v DuraVal) Eval(d ...Primary) Primary {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnboxedVector(v.TypePrim().Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v ErrorVal) Eval(d ...Primary) Primary {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnboxedVector(v.TypePrim().Flag(), d...)
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
func (v FlagSlice) Null() Primary { return New(FlagSlice{}) }
func (v PairVal) Null() Primary   { return PairVal{NilVal{}, NilVal{}} }
func (v NilVal) Null() Primary    { return NilVal{} }
func (v BoolVal) Null() Primary   { return New(false) }
func (v IntVal) Null() Primary    { return New(0) }
func (v Int8Val) Null() Primary   { return New(int8(0)) }
func (v Int16Val) Null() Primary  { return New(int16(0)) }
func (v Int32Val) Null() Primary  { return New(int32(0)) }
func (v UintVal) Null() Primary   { return New(uint(0)) }
func (v Uint8Val) Null() Primary  { return New(uint8(0)) }
func (v Uint16Val) Null() Primary { return New(uint16(0)) }
func (v Uint32Val) Null() Primary { return New(uint32(0)) }
func (v FltVal) Null() Primary    { return New(0.0) }
func (v Flt32Val) Null() Primary  { return New(float32(0.0)) }
func (v ImagVal) Null() Primary   { return New(complex128(0.0)) }
func (v Imag64Val) Null() Primary { return New(complex64(0.0)) }
func (v ByteVal) Null() Primary   { return New(byte(0)) }
func (v BytesVal) Null() Primary  { return New([]byte{}) }
func (v RuneVal) Null() Primary   { return New(rune(' ')) }
func (v StrVal) Null() Primary    { return New(string("")) }
func (v ErrorVal) Null() Primary  { return New(error(fmt.Errorf(""))) }
func (v BigIntVal) Null() Primary { return New(big.NewInt(0)) }
func (v BigFltVal) Null() Primary { return New(big.NewFloat(0)) }
func (v RatioVal) Null() Primary  { return New(big.NewRat(1, 1)) }
func (v TimeVal) Null() Primary   { return New(time.Now()) }
func (v DuraVal) Null() Primary   { return New(time.Duration(0)) }

//// BINARY MARSHALER //////
// for bytecode serialization and stack frame encoding
func (v BitFlag) MarshalBinary() ([]byte, error) {
	var u = uint64(v)
	var buf = make([]byte, 0, binary.Size(u))
	binary.PutUvarint(buf, u)
	return buf, nil
}
func (v FlagSlice) MarshalBinary() ([]byte, error) {
	var buf = bytes.NewBuffer(make([]byte, 0, binary.Size(v)))
	err := binary.Write(buf, binary.LittleEndian, v)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), err
}
func (v PairVal) MarshalBinary() ([]byte, error) {
	buf0, err0 := v.Left().(BinaryMarshaler).MarshalBinary()
	if err0 != nil {
		return nil, err0
	}
	buf1, err1 := v.Right().(BinaryMarshaler).MarshalBinary()
	if err1 != nil {
		return nil, err1
	}
	return append(buf0, buf1...), nil
}
func (v NilVal) MarshalBinary() ([]byte, error) {
	var u = uint64(0)
	var buf = make([]byte, 0, binary.Size(u))
	binary.PutUvarint(buf, u)
	return buf, nil
}
func (v BoolVal) MarshalBinary() ([]byte, error) {
	var u = uint64(v.Uint())
	var buf = make([]byte, 0, binary.Size(u))
	binary.PutUvarint(buf, u)
	return buf, nil
}
func (v IntVal) MarshalBinary() ([]byte, error) {
	var u = uint64(v)
	var buf = make([]byte, 0, binary.Size(u))
	binary.PutUvarint(buf, u)
	return buf, nil
}
func (v Int8Val) MarshalBinary() ([]byte, error) {
	var u = uint64(v)
	var buf = make([]byte, 0, binary.Size(u))
	binary.PutUvarint(buf, u)
	return buf, nil
}
func (v Int16Val) MarshalBinary() ([]byte, error) {
	var u = uint64(v)
	var buf = make([]byte, 0, binary.Size(u))
	binary.PutUvarint(buf, u)
	return buf, nil
}
func (v Int32Val) MarshalBinary() ([]byte, error) {
	var u = uint64(v)
	var buf = make([]byte, 0, binary.Size(u))
	binary.PutUvarint(buf, u)
	return buf, nil
}
func (v UintVal) MarshalBinary() ([]byte, error) {
	var u = uint64(v)
	var buf = make([]byte, 0, binary.Size(u))
	binary.PutUvarint(buf, u)
	return buf, nil
}
func (v Uint8Val) MarshalBinary() ([]byte, error) {
	var u = uint64(v)
	var buf = make([]byte, 0, binary.Size(u))
	binary.PutUvarint(buf, u)
	return buf, nil
}
func (v Uint16Val) MarshalBinary() ([]byte, error) {
	var u = uint64(v)
	var buf = make([]byte, 0, binary.Size(u))
	binary.PutUvarint(buf, u)
	return buf, nil
}
func (v Uint32Val) MarshalBinary() ([]byte, error) {
	var u = uint64(v)
	var buf = make([]byte, 0, binary.Size(u))
	binary.PutUvarint(buf, u)
	return buf, nil
}
func (v FltVal) MarshalBinary() ([]byte, error) {
	var u = uint64(v)
	var buf = make([]byte, 0, binary.Size(u))
	binary.PutUvarint(buf, u)
	return buf, nil
}
func (v Flt32Val) MarshalBinary() ([]byte, error) {
	var u = uint64(v)
	var buf = make([]byte, 0, binary.Size(u))
	binary.PutUvarint(buf, u)
	return buf, nil
}
func (v ImagVal) MarshalBinary() ([]byte, error) {
	var n, i = uint64(real(v)), uint64(imag(v))
	var buf0 = make([]byte, 0, binary.Size(n))
	var buf1 = make([]byte, 0, binary.Size(i))
	binary.PutUvarint(buf0, n)
	binary.PutUvarint(buf1, i)
	return append(buf0, buf1...), nil
}
func (v Imag64Val) MarshalBinary() ([]byte, error) {
	var n, i = uint64(real(v)), uint64(imag(v))
	var buf0 = make([]byte, 0, binary.Size(n))
	var buf1 = make([]byte, 0, binary.Size(i))
	binary.PutUvarint(buf0, n)
	binary.PutUvarint(buf1, i)
	return append(buf0, buf1...), nil
}
func (v ByteVal) MarshalBinary() ([]byte, error) {
	var buf = make([]byte, 0, binary.Size(v))
	binary.PutUvarint(buf, uint64(v))
	return buf, nil
}
func (v BytesVal) MarshalBinary() ([]byte, error) {
	var buf = bytes.NewBuffer(make([]byte, 0, binary.Size(v)))
	err := binary.Write(buf, binary.LittleEndian, v)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), err
}
func (v RuneVal) MarshalBinary() ([]byte, error) {
	var u = uint64(v)
	var buf = make([]byte, 0, binary.Size(u))
	binary.PutUvarint(buf, u)
	return buf, nil
}
func (v StrVal) MarshalBinary() ([]byte, error) {
	var buf = bytes.NewBuffer(make([]byte, 0, binary.Size(v)))
	err := binary.Write(buf, binary.LittleEndian, []byte(v))
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), err
}
func (v ErrorVal) MarshalBinary() ([]byte, error) {
	var buf = bytes.NewBuffer(make([]byte, 0, binary.Size(v.String())))
	err := binary.Write(buf, binary.LittleEndian, []byte(v.String()))
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), err
}
func (v BigIntVal) MarshalBinary() ([]byte, error) {
	var buf = bytes.NewBuffer(make([]byte, 0, binary.Size((*big.Int)(&v).Bytes())))
	err := binary.Write(buf, binary.LittleEndian, (*big.Int)(&v).Bytes())
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), err
}
func (v BigFltVal) MarshalBinary() ([]byte, error) {
	var u, _ = (*big.Float)(&v).Uint64()
	var buf = bytes.NewBuffer(make([]byte, 0, binary.Size(u)))
	err := binary.Write(buf, binary.LittleEndian, u)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), err
}
func (v RatioVal) MarshalBinary() ([]byte, error) {
	var d, n = uint64((*big.Rat)(&v).Denom().Uint64()), uint64((*big.Rat)(&v).Num().Uint64())
	var buf0 = make([]byte, 0, binary.Size(d))
	var buf1 = make([]byte, 0, binary.Size(n))
	binary.PutUvarint(buf0, d)
	binary.PutUvarint(buf1, n)
	return append(buf0, buf1...), nil
}
func (v TimeVal) MarshalBinary() ([]byte, error) {
	var buf = make([]byte, 0, binary.Size(v))
	(*time.Time)(&v).UnmarshalBinary(buf)
	return buf, nil
}
func (v DuraVal) MarshalBinary() ([]byte, error) {
	var u = uint64(v)
	var buf = make([]byte, 0, binary.Size(u))
	binary.PutUvarint(buf, u)
	return buf, nil
}
