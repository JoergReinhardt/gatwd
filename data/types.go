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
type TyNative BitFlag

func (v TyNative) TypeNat() TyNative { return v }
func (v TyNative) Eval(p ...Native) Native {
	if len(p) > 0 {
		for _, prime := range p {
			if prime.TypeNat().Flag().Match(Flag) {
				var flag = prime.(BitFlag)
				v = v | flag.TypeNat()
			}
		}
	}
	return v
}
func (v TyNative) Flag() BitFlag { return BitFlag(v) }

func ListAllTypes() []TyNative {
	var tt = []TyNative{}
	var i uint
	var t TyNative = 0
	for t < Flag {
		t = 1 << i
		i = i + 1
		tt = append(tt, TyNative(t))
	}
	return tt
}

//go:generate stringer -type=TyNative
const (
	Nil  TyNative = 1
	Bool TyNative = 1 << iota
	Int8          // Int8 -> Int8
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
	Pipe
	Buffer
	Reader
	Writer
	Channel
	Error // let's do something sophisticated here...
	//// HIGHERORDER TYPES
	Nat
	Fnc
	FlatType
	SumType
	ProdType
	Pair
	Tuple
	Record
	Vector
	List
	Set
	Data
	Expression
	Function
	Instance
	Flag // marks most signifficant native type & data of type bitflag

	// TYPE CLASSES
	// precedence type classes define argument types functions that accept
	// a set of possible input types
	Natives = Nil | Bool | Int8 | Int16 | Int32 | Int | BigInt | Uint8 |
		Uint16 | Uint32 | Uint | Flt32 | Float | BigFlt | Ratio | Imag64 |
		Imag | Time | Duration | Byte | Rune | Bytes | String | Error

	Bitwise  = Naturals | Byte | Flag
	Booleans = Bool | Bitwise
	Naturals = Uint | Uint8 | Uint16 | Uint32

	Integers   = Int | Int8 | Int16 | Int32 | BigInt
	Rationals  = Naturals | Ratio
	Reals      = Float | Flt32 | BigFlt
	Imaginarys = Imag | Imag64
	Numbers    = Rationals | Reals | Imaginarys
	Letters    = String | Rune | Bytes
	Equals     = Numbers | Letters
	Streams    = Reader | Writer | Pipe

	Compositions = Pair | Tuple | Record | Vector | List | Set
	Type         = FlatType | SumType | ProdType
	Functional   = Compositions | Type

	MASK         TyNative = 0xFFFFFFFFFFFFFFFF
	MASK_NATIVES          = MASK ^ Natives
)

//////// INTERNAL TYPES /////////////
// internal types are typealiases without any wrapping, or referencing getting
// in the way performancewise. types need to be aliased in the first place, to
// associate them with a bitflag, without having to actually asign, let alone
// attach it to the instance.
type ( // NATIVE GOLANG TYPES
	NilVal    struct{}
	BitFlag   uint
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

	// COMPLEX GOLANG TYPES
	BigIntVal big.Int
	BigFltVal big.Float
	RatioVal  big.Rat
	TimeVal   time.Time
	DuraVal   time.Duration
	ErrorVal  struct{ E error }
	PairVal   struct{ L, R Native }

	// SETS OF NATIVES
	SetString map[StrVal]Native
	SetUint   map[UintVal]Native
	SetInt    map[IntVal]Native
	SetFloat  map[FltVal]Native
	SetFlag   map[BitFlag]Native
	SetVal    map[Native]Native

	// GENERIC SLICE
	DataSlice []Native

	// SLICE OF BIT FLAGS
	FlagSlice []BitFlag

	// SLICES OF UNALIASED NATIVES
	InterfaceSlice []interface{}
	NilVec         []struct{}
	BoolVec        []bool
	IntVec         []int
	Int8Vec        []int8
	Int16Vec       []int16
	Int32Vec       []int32
	UintVec        []uint
	Uint8Vec       []uint8
	Uint16Vec      []uint16
	Uint32Vec      []uint32
	FltVec         []float64
	Flt32Vec       []float32
	ImagVec        []complex128
	Imag64Vec      []complex64
	ByteVec        []byte
	RuneVec        []rune
	BytesVec       [][]byte
	StrVec         []string
	BigIntVec      []*big.Int
	BigFltVec      []*big.Float
	RatioVec       []*big.Rat
	TimeVec        []time.Time
	DuraVec        []time.Duration
	ErrorVec       []error
	FlagSet        []BitFlag
)

/// bind the appropriate TypeNat Method to every type
func (v BitFlag) TypeNat() TyNative   { return Flag.TypeNat() }
func (v FlagSlice) Flag() TyNative    { return Flag.TypeNat() }
func (NilVal) TypeNat() TyNative      { return Nil.TypeNat() }
func (v BoolVal) TypeNat() TyNative   { return Bool.TypeNat() }
func (v IntVal) TypeNat() TyNative    { return Int.TypeNat() }
func (v Int8Val) TypeNat() TyNative   { return Int8.TypeNat() }
func (v Int16Val) TypeNat() TyNative  { return Int16.TypeNat() }
func (v Int32Val) TypeNat() TyNative  { return Int32.TypeNat() }
func (v UintVal) TypeNat() TyNative   { return Uint.TypeNat() }
func (v Uint8Val) TypeNat() TyNative  { return Uint8.TypeNat() }
func (v Uint16Val) TypeNat() TyNative { return Uint16.TypeNat() }
func (v Uint32Val) TypeNat() TyNative { return Uint32.TypeNat() }
func (v BigIntVal) TypeNat() TyNative { return BigInt.TypeNat() }
func (v FltVal) TypeNat() TyNative    { return Float.TypeNat() }
func (v Flt32Val) TypeNat() TyNative  { return Flt32.TypeNat() }
func (v BigFltVal) TypeNat() TyNative { return BigFlt.TypeNat() }
func (v ImagVal) TypeNat() TyNative   { return Imag.TypeNat() }
func (v Imag64Val) TypeNat() TyNative { return Imag64.TypeNat() }
func (v RatioVal) TypeNat() TyNative  { return Ratio.TypeNat() }
func (v RuneVal) TypeNat() TyNative   { return Rune.TypeNat() }
func (v ByteVal) TypeNat() TyNative   { return Byte.TypeNat() }
func (v BytesVal) TypeNat() TyNative  { return Bytes.TypeNat() }
func (v StrVal) TypeNat() TyNative    { return String.TypeNat() }
func (v TimeVal) TypeNat() TyNative   { return Time.TypeNat() }
func (v DuraVal) TypeNat() TyNative   { return Duration.TypeNat() }
func (v ErrorVal) TypeNat() TyNative  { return Error.TypeNat() }
func (v PairVal) TypeNat() TyNative   { return Pair.TypeNat() }

///
func (NilVal) Copy() Native      { return NilVal{} }
func (v BitFlag) Copy() Native   { return BitFlag(v) }
func (v BoolVal) Copy() Native   { return BoolVal(v) }
func (v IntVal) Copy() Native    { return IntVal(v) }
func (v Int8Val) Copy() Native   { return Int8Val(v) }
func (v Int16Val) Copy() Native  { return Int16Val(v) }
func (v Int32Val) Copy() Native  { return Int32Val(v) }
func (v UintVal) Copy() Native   { return UintVal(v) }
func (v Uint8Val) Copy() Native  { return Uint8Val(v) }
func (v Uint16Val) Copy() Native { return Uint16Val(v) }
func (v Uint32Val) Copy() Native { return Uint32Val(v) }
func (v BigIntVal) Copy() Native { return BigIntVal(v) }
func (v FltVal) Copy() Native    { return FltVal(v) }
func (v Flt32Val) Copy() Native  { return Flt32Val(v) }
func (v BigFltVal) Copy() Native { return BigFltVal(v) }
func (v ImagVal) Copy() Native   { return ImagVal(v) }
func (v Imag64Val) Copy() Native { return Imag64Val(v) }
func (v RatioVal) Copy() Native  { return RatioVal(v) }
func (v RuneVal) Copy() Native   { return RuneVal(v) }
func (v ByteVal) Copy() Native   { return ByteVal(v) }
func (v BytesVal) Copy() Native  { return BytesVal(v) }
func (v StrVal) Copy() Native    { return StrVal(v) }
func (v TimeVal) Copy() Native   { return TimeVal(v) }
func (v DuraVal) Copy() Native   { return DuraVal(v) }
func (v ErrorVal) Copy() Native  { return ErrorVal(v) }
func (v PairVal) Copy() Native   { return PairVal{v.L, v.R} }
func (v FlagSlice) Copy() Native {
	var nfs = DataSlice{}
	for _, dat := range v {
		nfs = append(nfs, dat)
	}
	return nfs
}

///
func (NilVal) Eval(d ...Native) Native { return NilVal{} }
func (v BitFlag) Eval(d ...Native) Native {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnboxedVector(v.TypeNat().Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v BoolVal) Eval(d ...Native) Native {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnboxedVector(v.TypeNat().Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v IntVal) Eval(d ...Native) Native {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnboxedVector(v.TypeNat().Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v Int8Val) Eval(d ...Native) Native {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnboxedVector(v.TypeNat().Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v Int16Val) Eval(d ...Native) Native {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnboxedVector(v.TypeNat().Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v Int32Val) Eval(d ...Native) Native {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnboxedVector(v.TypeNat().Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v UintVal) Eval(d ...Native) Native {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnboxedVector(v.TypeNat().Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v Uint8Val) Eval(d ...Native) Native {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnboxedVector(v.TypeNat().Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v Uint16Val) Eval(d ...Native) Native {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnboxedVector(v.TypeNat().Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v Uint32Val) Eval(d ...Native) Native {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnboxedVector(v.TypeNat().Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v BigIntVal) Eval(d ...Native) Native {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnboxedVector(v.TypeNat().Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v FltVal) Eval(d ...Native) Native {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnboxedVector(v.TypeNat().Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v Flt32Val) Eval(d ...Native) Native {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnboxedVector(v.TypeNat().Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v BigFltVal) Eval(d ...Native) Native {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnboxedVector(v.TypeNat().Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v ImagVal) Eval(d ...Native) Native {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnboxedVector(v.TypeNat().Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v Imag64Val) Eval(d ...Native) Native {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnboxedVector(v.TypeNat().Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v RatioVal) Eval(d ...Native) Native {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnboxedVector(v.TypeNat().Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v RuneVal) Eval(d ...Native) Native {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnboxedVector(v.TypeNat().Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v ByteVal) Eval(d ...Native) Native {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnboxedVector(v.TypeNat().Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v BytesVal) Eval(d ...Native) Native {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnboxedVector(v.TypeNat().Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v StrVal) Eval(d ...Native) Native {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnboxedVector(v.TypeNat().Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v TimeVal) Eval(d ...Native) Native {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnboxedVector(v.TypeNat().Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v DuraVal) Eval(d ...Native) Native {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnboxedVector(v.TypeNat().Flag(), d...)
		}
		return d[0]
	}
	return v
}
func (v ErrorVal) Eval(d ...Native) Native {
	if len(d) > 0 {
		if len(d) > 1 {
			return NewUnboxedVector(v.TypeNat().Flag(), d...)
		}
		return d[0]
	}
	return v
}

///
func (NilVal) Ident() Native      { return NilVal{} }
func (v BitFlag) Ident() Native   { return v }
func (v BoolVal) Ident() Native   { return v }
func (v IntVal) Ident() Native    { return v }
func (v Int8Val) Ident() Native   { return v }
func (v Int16Val) Ident() Native  { return v }
func (v Int32Val) Ident() Native  { return v }
func (v UintVal) Ident() Native   { return v }
func (v Uint8Val) Ident() Native  { return v }
func (v Uint16Val) Ident() Native { return v }
func (v Uint32Val) Ident() Native { return v }
func (v BigIntVal) Ident() Native { return v }
func (v FltVal) Ident() Native    { return v }
func (v Flt32Val) Ident() Native  { return v }
func (v BigFltVal) Ident() Native { return v }
func (v ImagVal) Ident() Native   { return v }
func (v Imag64Val) Ident() Native { return v }
func (v RatioVal) Ident() Native  { return v }
func (v RuneVal) Ident() Native   { return v }
func (v ByteVal) Ident() Native   { return v }
func (v BytesVal) Ident() Native  { return v }
func (v StrVal) Ident() Native    { return v }
func (v TimeVal) Ident() Native   { return v }
func (v DuraVal) Ident() Native   { return v }
func (v ErrorVal) Ident() Native  { return v }
func (v PairVal) Ident() Native   { return v }

//// native nullable typed ///////
func (v BitFlag) Null() Native   { return Nil.TypeNat() }
func (v FlagSlice) Null() Native { return New(FlagSlice{}) }
func (v PairVal) Null() Native   { return PairVal{NilVal{}, NilVal{}} }
func (v NilVal) Null() Native    { return NilVal{} }
func (v BoolVal) Null() Native   { return New(false) }
func (v IntVal) Null() Native    { return New(0) }
func (v Int8Val) Null() Native   { return New(int8(0)) }
func (v Int16Val) Null() Native  { return New(int16(0)) }
func (v Int32Val) Null() Native  { return New(int32(0)) }
func (v UintVal) Null() Native   { return New(uint(0)) }
func (v Uint8Val) Null() Native  { return New(uint8(0)) }
func (v Uint16Val) Null() Native { return New(uint16(0)) }
func (v Uint32Val) Null() Native { return New(uint32(0)) }
func (v FltVal) Null() Native    { return New(0.0) }
func (v Flt32Val) Null() Native  { return New(float32(0.0)) }
func (v ImagVal) Null() Native   { return New(complex128(0.0)) }
func (v Imag64Val) Null() Native { return New(complex64(0.0)) }
func (v ByteVal) Null() Native   { return New(byte(0)) }
func (v BytesVal) Null() Native  { return New([]byte{}) }
func (v RuneVal) Null() Native   { return New(rune(' ')) }
func (v StrVal) Null() Native    { return New(string("")) }
func (v ErrorVal) Null() Native  { return New(error(fmt.Errorf(""))) }
func (v BigIntVal) Null() Native { return New(big.NewInt(0)) }
func (v BigFltVal) Null() Native { return New(big.NewFloat(0)) }
func (v RatioVal) Null() Native  { return New(big.NewRat(1, 1)) }
func (v TimeVal) Null() Native   { return New(time.Now()) }
func (v DuraVal) Null() Native   { return New(time.Duration(0)) }

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
