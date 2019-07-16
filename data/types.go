package data

import (
	"fmt"
	"math/big"
	s "strings"
	"time"
)

//////// INTERNAL TYPE SYSTEM ///////////
//
// intended to be accessable and extendable
type TyNat BitFlag

func (t TyNat) FlagType() Uint8Val { return 1 }
func (v TyNat) Type() TyNat        { return v }

func (t TyNat) TypeName() string {
	var count = t.Flag().Count()
	// loop to print concatenated type classes correcty
	if count > 1 {
		switch t {
		case Natives:
			return "Natives"
		case Bitwise:
			return "Bitwise"
		case Booleans:
			return "Bools"
		case Naturals:
			return "Naturals"
		case Integers:
			return "Integers"
		case Rationals:
			return "Rationals"
		case Reals:
			return "Reals"
		case Imaginarys:
			return "Imaginarys"
		case Numbers:
			return "Numbers"
		case Letters:
			return "Letters"
		case Equals:
			return "Equals"
		case Compositions:
			return "Compositions"
		case Parametric:
			return "Parametric"
		case Functional:
			return "Functional"
		}
		var delim = "|"
		var str = make([]string, 0, count)
		for _, flag := range t.Flag().Decompose() {
			str = append(str, TyNat(flag.Flag()).String())
		}
		return s.Join(str, delim)
	}
	return t.String()
}
func (v TyNat) Flag() BitFlag        { return BitFlag(v) }
func (v TyNat) Match(arg Typed) bool { return v.Flag().Match(arg) }
func (v TyNat) Eval() Native         { return v }

func FetchTypes() []TyNat {
	var tt = []TyNat{}
	var i uint
	var t TyNat = 0
	for t < Type {
		t = 1 << i
		i = i + 1
		tt = append(tt, TyNat(t))
	}
	return tt
}

//go:generate stringer -type=TyNat
const (
	Nil TyNat = 1 << iota
	Bool
	Int8
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
	Flag
	Error // let's do something sophisticated here...
	////
	Pair
	Slice
	Unboxed
	Map
	////
	Function
	Literal
	Type // marks most signifficant native type & data of type bitflag

	// TYPE CLASSES
	// precedence type classes define argument types functions that accept
	// a set of possible input types
	Natives = Nil | Bool | Int8 | Int16 | Int32 | Int | BigInt | Uint8 |
		Uint16 | Uint32 | Uint | Flt32 | Float | BigFlt | Ratio | Imag64 |
		Imag | Time | Duration | Byte | Rune | Bytes | String | Error

	Bitwise    = Naturals | Byte | Type
	Booleans   = Bool | Bitwise
	Naturals   = Uint | Uint8 | Uint16 | Uint32
	Integers   = Int | Int8 | Int16 | Int32 | BigInt
	Rationals  = Naturals | Integers | Ratio
	Reals      = Float | Flt32 | BigFlt
	Imaginarys = Imag | Imag64
	Numbers    = Rationals | Reals | Imaginarys
	Letters    = String | Rune | Bytes
	Equals     = Numbers | Letters

	Compositions = Pair | Unboxed | Slice | Map

	Parametric = Natives | Compositions

	Functional = Literal | Function | Type

	Sets = Natives | Bitwise | Booleans | Naturals | Integers | Rationals |
		Reals | Imaginarys | Numbers | Letters | Equals | Compositions |
		Parametric | Functional

	MASK         TyNat = 0xFFFFFFFFFFFFFFFF
	MASK_NATIVES       = MASK ^ Natives
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

	// COMPOSED GOLANG TYPES
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

	// SLICES OF UNALIASED NATIVE GOLANG VALUES
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

	// SLICE OF BIT FLAGS
	FlagSlice []BitFlag

	// SLICE OF GENERIC VALUES
	DataSlice []Native

	// FUNCTION VALUE
	FuncVal func(...Native) Native
)

func newUnboxed(nat TyNat) Native {
	var val Native
	switch nat {
	case Bool:
		val = BoolVec([]bool{})
	case Int:
		val = IntVec([]int{})
	case Int8:
		val = Int8Vec([]int8{})
	case Int16:
		val = Int16Vec([]int16{})
	case Int32:
		val = Int32Vec([]int32{})
	case Uint:
		val = UintVec([]uint{})
	case Uint8:
		val = Uint8Vec([]uint8{})
	case Uint16:
		val = Uint16Vec([]uint16{})
	case Uint32:
		val = Uint32Vec([]uint32{})
	case Float:
		val = FltVec([]float64{})
	case Flt32:
		val = Flt32Vec([]float32{})
	case Imag:
		val = ImagVec([]complex128{})
	case Imag64:
		val = Imag64Vec([]complex64{})
	case Byte:
		val = ByteVec([]byte{})
	case Rune:
		val = RuneVec([]rune{})
	case Bytes:
		val = BytesVec([][]byte{})
	case String:
		val = StrVec([]string{})
	case BigInt:
		val = BigIntVec([]*big.Int{})
	case BigFlt:
		val = BigFltVec([]*big.Float{})
	case Ratio:
		val = RatioVec([]*big.Rat{})
	case Time:
		val = TimeVec([]time.Time{})
	case Duration:
		val = DuraVec([]time.Duration{})
	case Error:
		val = ErrorVec([]error{})
	case Flag:
		val = FlagSet([]BitFlag{})
	}
	return val
}

func newNull(nat TyNat) Native {
	var val Native
	switch nat {
	case Nil:
		val = NilVal{}
	case Bool:
		val = BoolVal(false)
	case Int8:
		val = Int8Val(int8(0))
	case Int16:
		val = Int16Val(int16(0))
	case Int32:
		val = Int32Val(int32(0))
	case Int:
		val = IntVal(0)
	case BigInt:
		val = BigIntVal(*big.NewInt(0))
	case Uint8:
		val = Uint8Val(uint8(0))
	case Uint16:
		val = Uint16Val(uint16(0))
	case Uint32:
		val = Uint32Val(uint32(0))
	case Uint:
		val = UintVal(uint(0))
	case Flt32:
		val = Flt32Val(float32(0.0))
	case Float:
		val = FltVal(float64(0.0))
	case BigFlt:
		val = BigFltVal(*big.NewFloat(0.0))
	case Ratio:
		val = RatioVal(*big.NewRat(1, 1))
	case Imag64:
		val = Imag64Val(complex64(0.0))
	case Imag:
		val = Imag64Val(complex128(0.0))
	case Time:
		val = TimeVal(time.Now())
	case Duration:
		val = DuraVal(time.Duration(0))
	case Byte:
		val = ByteVal(byte(0))
	case Rune:
		val = RuneVal(rune(' '))
	case Bytes:
		val = BytesVal([]byte{})
	case String:
		val = StrVal("")
	case Error:
		val = ErrorVal{error(fmt.Errorf(""))}
	default:
		val = NilVal{}
	}
	return val
}

func (v FuncVal) Eval(args ...Native) Native { return v(args...) }

// yields a null value of the methods type
func (v FlagSlice) Null() FlagSlice { return FlagSlice(FlagSlice{}) }
func (v BitFlag) Null() BitFlag     { return BitFlag(BitFlag(0)) }
func (v PairVal) Null() PairVal     { return PairVal(PairVal{NilVal{}, NilVal{}}) }

func (v NilVal) Null() NilVal       { return NilVal(NilVal{}) }
func (v BoolVal) Null() BoolVal     { return BoolVal(false) }
func (v Int8Val) Null() Int8Val     { return Int8Val(int8(0)) }
func (v Int16Val) Null() Int16Val   { return Int16Val(int16(0)) }
func (v Int32Val) Null() Int32Val   { return Int32Val(int32(0)) }
func (v IntVal) Null() IntVal       { return IntVal(0) }
func (v BigIntVal) Null() BigIntVal { return BigIntVal(*big.NewInt(0)) }
func (v Uint8Val) Null() Uint8Val   { return Uint8Val(uint8(0)) }
func (v Uint16Val) Null() Uint16Val { return Uint16Val(uint16(0)) }
func (v Uint32Val) Null() Uint32Val { return Uint32Val(uint32(0)) }
func (v UintVal) Null() UintVal     { return UintVal(uint(0)) }
func (v Flt32Val) Null() Flt32Val   { return Flt32Val(float32(0.0)) }
func (v FltVal) Null() FltVal       { return FltVal(0.0) }
func (v BigFltVal) Null() BigFltVal { return BigFltVal(*big.NewFloat(0)) }
func (v RatioVal) Null() RatioVal   { return RatioVal(*big.NewRat(1, 1)) }
func (v Imag64Val) Null() Imag64Val { return Imag64Val(complex64(0.0)) }
func (v ImagVal) Null() ImagVal     { return ImagVal(complex128(0.0)) }
func (v TimeVal) Null() TimeVal     { return TimeVal(time.Now()) }
func (v DuraVal) Null() DuraVal     { return DuraVal(time.Duration(0)) }
func (v ByteVal) Null() ByteVal     { return ByteVal(byte(0)) }
func (v RuneVal) Null() RuneVal     { return RuneVal(rune(' ')) }
func (v BytesVal) Null() BytesVal   { return BytesVal([]byte{}) }
func (v StrVal) Null() StrVal       { return StrVal(string("")) }
func (v ErrorVal) Null() ErrorVal   { return ErrorVal{error(fmt.Errorf(""))} }

/// bind the corresponding Type Method to every type
func (v BitFlag) Type() TyNat   { return Flag }
func (v FlagSlice) Flag() TyNat { return Flag | Slice }

func (v NilVal) Type() TyNat    { return Nil }
func (v BoolVal) Type() TyNat   { return Bool }
func (v IntVal) Type() TyNat    { return Int }
func (v Int8Val) Type() TyNat   { return Int8 }
func (v Int16Val) Type() TyNat  { return Int16 }
func (v Int32Val) Type() TyNat  { return Int32 }
func (v UintVal) Type() TyNat   { return Uint }
func (v Uint8Val) Type() TyNat  { return Uint8 }
func (v Uint16Val) Type() TyNat { return Uint16 }
func (v Uint32Val) Type() TyNat { return Uint32 }
func (v BigIntVal) Type() TyNat { return BigInt }
func (v FltVal) Type() TyNat    { return Float }
func (v Flt32Val) Type() TyNat  { return Flt32 }
func (v BigFltVal) Type() TyNat { return BigFlt }
func (v ImagVal) Type() TyNat   { return Imag }
func (v Imag64Val) Type() TyNat { return Imag64 }
func (v RatioVal) Type() TyNat  { return Ratio }
func (v RuneVal) Type() TyNat   { return Rune }
func (v ByteVal) Type() TyNat   { return Byte }
func (v BytesVal) Type() TyNat  { return Bytes }
func (v StrVal) Type() TyNat    { return String }
func (v TimeVal) Type() TyNat   { return Time }
func (v DuraVal) Type() TyNat   { return Duration }
func (v ErrorVal) Type() TyNat  { return Error }
func (v FuncVal) Type() TyNat   { return Function }

//func (v BitFlag) Type() Typed   { return Type }
//func (v FlagSlice) Type() Typed { return Flag | Slice }
//func (v NilVal) Type() Typed    { return Nil }
//func (v BoolVal) Type() Typed   { return Bool }
//func (v IntVal) Type() Typed    { return Int }
//func (v Int8Val) Type() Typed   { return Int8 }
//func (v Int16Val) Type() Typed  { return Int16 }
//func (v Int32Val) Type() Typed  { return Int32 }
//func (v UintVal) Type() Typed   { return Uint }
//func (v Uint8Val) Type() Typed  { return Uint8 }
//func (v Uint16Val) Type() Typed { return Uint16 }
//func (v Uint32Val) Type() Typed { return Uint32 }
//func (v BigIntVal) Type() Typed { return BigInt }
//func (v FltVal) Type() Typed    { return Float }
//func (v Flt32Val) Type() Typed  { return Flt32 }
//func (v BigFltVal) Type() Typed { return BigFlt }
//func (v ImagVal) Type() Typed   { return Imag }
//func (v Imag64Val) Type() Typed { return Imag64 }
//func (v RatioVal) Type() Typed  { return Ratio }
//func (v RuneVal) Type() Typed   { return Rune }
//func (v ByteVal) Type() Typed   { return Byte }
//func (v BytesVal) Type() Typed  { return Bytes }
//func (v StrVal) Type() Typed    { return String }
//func (v TimeVal) Type() Typed   { return Time }
//func (v DuraVal) Type() Typed   { return Duration }
//func (v ErrorVal) Type() Typed  { return Error }
//func (v FuncVal) Type() Typed   { return Function }

// provide a deep copy method
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
func (v FuncVal) Copy() Native   { return FuncVal(v) }
func (v PairVal) Copy() Native   { return PairVal{v.L, v.R} }
func (v FlagSlice) Copy() Native {
	var nfs = DataSlice{}
	for _, dat := range v {
		nfs = append(nfs, dat)
	}
	return nfs
}

// ident returns the instance as it's given type
func (NilVal) Ident() NilVal         { return NilVal{} }
func (v BitFlag) Ident() BitFlag     { return v }
func (v BoolVal) Ident() BoolVal     { return v }
func (v IntVal) Ident() IntVal       { return v }
func (v Int8Val) Ident() Int8Val     { return v }
func (v Int16Val) Ident() Int16Val   { return v }
func (v Int32Val) Ident() Int32Val   { return v }
func (v UintVal) Ident() UintVal     { return v }
func (v Uint8Val) Ident() Uint8Val   { return v }
func (v Uint16Val) Ident() Uint16Val { return v }
func (v Uint32Val) Ident() Uint32Val { return v }
func (v BigIntVal) Ident() BigIntVal { return v }
func (v FltVal) Ident() FltVal       { return v }
func (v Flt32Val) Ident() Flt32Val   { return v }
func (v BigFltVal) Ident() BigFltVal { return v }
func (v ImagVal) Ident() ImagVal     { return v }
func (v Imag64Val) Ident() Imag64Val { return v }
func (v RatioVal) Ident() RatioVal   { return v }
func (v RuneVal) Ident() RuneVal     { return v }
func (v ByteVal) Ident() ByteVal     { return v }
func (v BytesVal) Ident() BytesVal   { return v }
func (v StrVal) Ident() StrVal       { return v }
func (v TimeVal) Ident() TimeVal     { return v }
func (v DuraVal) Ident() DuraVal     { return v }
func (v ErrorVal) Ident() ErrorVal   { return v }
func (v FuncVal) Ident() FuncVal     { return v }
