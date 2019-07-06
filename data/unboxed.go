package data

import (
	"math/big"
	"time"
)

// expects all arguments to match the type passed in as flag
func NewUnboxed(flag TyNat, args ...Native) Sliceable {

	var d Sliceable

	switch flag {
	case Nil:
		d = NilVec{}
		for _, _ = range args {
			d = append(d.(NilVec), struct{}{})
		}

	case Bool:
		d = BoolVec{}
		for _, dat := range args {
			d = append(d.(BoolVec), bool(dat.(BoolVal)))
		}

	case Int:
		d = IntVec{}
		for _, dat := range args {
			d = append(d.(IntVec), int(dat.(IntVal)))
		}

	case Int8:
		d = Int8Vec{}
		for _, dat := range args {
			d = append(d.(Int8Vec), int8(dat.(Int8Val)))
		}

	case Int16:
		d = Int16Vec{}
		for _, dat := range args {
			d = append(d.(Int16Vec), int16(dat.(Int16Val)))
		}

	case Int32:
		d = Int32Vec{}
		for _, dat := range args {
			d = append(d.(Int32Vec), int32(dat.(Int32Val)))
		}

	case Uint:
		d = UintVec{}
		for _, dat := range args {
			d = append(d.(UintVec), uint(dat.(UintVal)))
		}

	case Uint8:
		d = Uint8Vec{}
		for _, dat := range args {
			d = append(d.(Uint8Vec), uint8(dat.(Uint8Val)))
		}

	case Uint16:
		d = Uint16Vec{}
		for _, dat := range args {
			d = append(d.(Uint16Vec), uint16(dat.(Uint16Val)))
		}

	case Uint32:
		d = Uint32Vec{}
		for _, dat := range args {
			d = append(d.(Uint32Vec), uint32(dat.(Uint32Val)))
		}

	case Float:
		d = FltVec{}
		for _, dat := range args {
			d = append(d.(FltVec), float64(dat.(FltVal)))
		}

	case Flt32:
		d = Flt32Vec{}
		for _, dat := range args {
			d = append(d.(Flt32Vec), float32(dat.(Flt32Val)))
		}

	case Imag:
		d = ImagVec{}
		for _, dat := range args {
			d = append(d.(ImagVec), complex128(dat.(ImagVal)))
		}

	case Imag64:
		d = Imag64Vec{}
		for _, dat := range args {
			d = append(d.(Imag64Vec), complex64(dat.(Imag64Val)))
		}

	case Byte:
		d = ByteVec{}
		for _, dat := range args {
			d = append(d.(ByteVec), byte(dat.(ByteVal)))
		}

	case Rune:
		d = RuneVec{}
		for _, dat := range args {
			d = append(d.(RuneVec), rune(dat.(RuneVal)))
		}

	case Bytes:
		d = BytesVec{}
		for _, dat := range args {
			d = append(d.(BytesVec), []byte(dat.(BytesVal)))
		}

	case String:
		d = StrVec{}
		for _, dat := range args {
			d = append(d.(StrVec), string(dat.(StrVal)))
		}

	case BigInt:
		d = BigIntVec{}
		for _, dat := range args {
			bi := dat.(BigIntVal)
			d = append(d.(BigIntVec), big.NewInt(((*big.Int)(&bi)).Int64()))
		}

	case BigFlt:
		d = BigFltVec{}
		for _, dat := range args {
			bf := dat.(BigFltVal)
			f, _ := ((*big.Float)(&bf)).Float64()
			d = append(d.(BigFltVec), big.NewFloat(f))
		}

	case Ratio:
		d = RatioVec{}
		for _, dat := range args {
			rat := dat.(RatioVal)
			d = append(d.(RatioVec), big.NewRat(
				((*big.Rat)(&rat)).Num().Int64(),
				((*big.Rat)(&rat)).Denom().Int64(),
			))
		}

	case Time:
		d = TimeVec{}
		for _, dat := range args {
			d = append(d.(TimeVec), time.Time(dat.(TimeVal)))
		}

	case Duration:
		d = DuraVec{}
		for _, dat := range args {
			d = append(d.(DuraVec), time.Duration(dat.(DuraVal)))
		}

	case Error:
		d = ErrorVec{}
		for _, dat := range args {
			d = append(d.(ErrorVec), error(dat.(ErrorVal).E))
		}
	}
	return d
}
func (v NilVec) Eval() Native    { return v }
func (v ErrorVec) Eval() Native  { return v }
func (v BoolVec) Eval() Native   { return v }
func (v IntVec) Eval() Native    { return v }
func (v Int8Vec) Eval() Native   { return v }
func (v Int16Vec) Eval() Native  { return v }
func (v Int32Vec) Eval() Native  { return v }
func (v UintVec) Eval() Native   { return v }
func (v Uint8Vec) Eval() Native  { return v }
func (v Uint16Vec) Eval() Native { return v }
func (v Uint32Vec) Eval() Native { return v }
func (v FltVec) Eval() Native    { return v }
func (v Flt32Vec) Eval() Native  { return v }
func (v ImagVec) Eval() Native   { return v }
func (v Imag64Vec) Eval() Native { return v }
func (v ByteVec) Eval() Native   { return v }
func (v RuneVec) Eval() Native   { return v }
func (v BytesVec) Eval() Native  { return v }
func (v StrVec) Eval() Native    { return v }
func (v BigIntVec) Eval() Native { return v }
func (v BigFltVec) Eval() Native { return v }
func (v RatioVec) Eval() Native  { return v }
func (v TimeVec) Eval() Native   { return v }
func (v DuraVec) Eval() Native   { return v }
func (v FlagSet) Eval() Native   { return v }

func (v ByteVec) Bytes() []byte { return []byte(v) }

func (v *ByteVec) Set(i int, b byte) { (*v)[i] = b }

func (v *ByteVec) Insert(i, j int, b byte) {
	var s = []byte(*v)
	s = append(s, byte(0))
	copy(s[i+1:], s[i:])
	s[i] = b
	*v = ByteVec(s)
}

func (v *ByteVec) InsertSlice(i, j int, b ...byte) {
	var s = []byte(*v)
	*v = ByteVec(append(s[:i], append(b, s[i:]...)...))
}

func (v *ByteVec) Cut(i, j int) {
	var s = []byte(*v)
	copy(s[i:], s[j:])
	// to prevent a possib. mem leak
	for k, n := len(s)-j+i, len(s); k < n; k++ {
		s[k] = byte(0)
	}
	*v = ByteVec(s[:len(s)-j+i])
}

func (v *ByteVec) Delete(i int) {
	var s = []byte(*v)
	copy(s[i:], s[i+1:])
	s[len(s)-1] = byte(0)
	*v = ByteVec(s[:len(s)-1])
}

func (v InterfaceSlice) GetInt(i int) interface{} { return v[i] }
func (v NilVec) GetInt(i int) Native              { return NilVal(v[i]) }
func (v BoolVec) GetInt(i int) Native             { return BoolVal(v[i]) }
func (v IntVec) GetInt(i int) Native              { return IntVal(v[i]) }
func (v Int8Vec) GetInt(i int) Native             { return Int8Val(v[i]) }
func (v Int16Vec) GetInt(i int) Native            { return Int16Val(v[i]) }
func (v Int32Vec) GetInt(i int) Native            { return Int32Val(v[i]) }
func (v UintVec) GetInt(i int) Native             { return UintVal(v[i]) }
func (v Uint8Vec) GetInt(i int) Native            { return Uint8Val(v[i]) }
func (v Uint16Vec) GetInt(i int) Native           { return Uint16Val(v[i]) }
func (v Uint32Vec) GetInt(i int) Native           { return Uint32Val(v[i]) }
func (v FltVec) GetInt(i int) Native              { return FltVal(v[i]) }
func (v Flt32Vec) GetInt(i int) Native            { return Flt32Val(v[i]) }
func (v ImagVec) GetInt(i int) Native             { return ImagVal(v[i]) }
func (v Imag64Vec) GetInt(i int) Native           { return Imag64Val(v[i]) }
func (v ByteVec) GetInt(i int) Native             { return ByteVal(v[i]) }
func (v RuneVec) GetInt(i int) Native             { return RuneVal(v[i]) }
func (v BytesVec) GetInt(i int) Native            { return BytesVal(v[i]) }
func (v StrVec) GetInt(i int) Native              { return StrVal(v[i]) }
func (v BigIntVec) GetInt(i int) Native           { return BigIntVal((*(*big.Int)(v[i]))) }
func (v BigFltVec) GetInt(i int) Native           { return BigFltVal((*(*big.Float)(v[i]))) }
func (v RatioVec) GetInt(i int) Native            { return RatioVal((*(*big.Rat)(v[i]))) }
func (v TimeVec) GetInt(i int) Native             { return TimeVal(v[i]) }
func (v DuraVec) GetInt(i int) Native             { return DuraVal(v[i]) }
func (v ErrorVec) GetInt(i int) Native            { return ErrorVal{v[i]} }
func (v FlagSet) GetInt(i int) Native             { return BitFlag(v[i]) }

func (v InterfaceSlice) Get(i Native) interface{} { return v[i.(IntVal).Idx()] }
func (v NilVec) Get(i Native) Native              { return NilVal(v[i.(IntVal).Idx()]) }
func (v BoolVec) Get(i Native) Native             { return BoolVal(v[i.(IntVal).Idx()]) }
func (v IntVec) Get(i Native) Native              { return IntVal(v[i.(IntVal).Idx()]) }
func (v Int8Vec) Get(i Native) Native             { return Int8Val(v[i.(IntVal).Idx()]) }
func (v Int16Vec) Get(i Native) Native            { return Int16Val(v[i.(IntVal).Idx()]) }
func (v Int32Vec) Get(i Native) Native            { return Int32Val(v[i.(IntVal).Idx()]) }
func (v UintVec) Get(i Native) Native             { return UintVal(v[i.(IntVal).Idx()]) }
func (v Uint8Vec) Get(i Native) Native            { return Uint8Val(v[i.(IntVal).Idx()]) }
func (v Uint16Vec) Get(i Native) Native           { return Uint16Val(v[i.(IntVal).Idx()]) }
func (v Uint32Vec) Get(i Native) Native           { return Uint32Val(v[i.(IntVal).Idx()]) }
func (v FltVec) Get(i Native) Native              { return FltVal(v[i.(IntVal).Idx()]) }
func (v Flt32Vec) Get(i Native) Native            { return Flt32Val(v[i.(IntVal).Idx()]) }
func (v ImagVec) Get(i Native) Native             { return ImagVal(v[i.(IntVal).Idx()]) }
func (v Imag64Vec) Get(i Native) Native           { return Imag64Val(v[i.(IntVal).Idx()]) }
func (v ByteVec) Get(i Native) Native             { return ByteVal(v[i.(IntVal).Idx()]) }
func (v RuneVec) Get(i Native) Native             { return RuneVal(v[i.(IntVal).Idx()]) }
func (v BytesVec) Get(i Native) Native            { return BytesVal(v[i.(IntVal).Idx()]) }
func (v StrVec) Get(i Native) Native              { return StrVal(v[i.(IntVal).Idx()]) }
func (v BigIntVec) Get(i Native) Native           { return BigIntVal((*(*big.Int)(v[i.(IntVal).Idx()]))) }
func (v BigFltVec) Get(i Native) Native           { return BigFltVal((*(*big.Float)(v[i.(IntVal).Idx()]))) }
func (v RatioVec) Get(i Native) Native            { return RatioVal((*(*big.Rat)(v[i.(IntVal).Idx()]))) }
func (v TimeVec) Get(i Native) Native             { return TimeVal(v[i.(IntVal).Idx()]) }
func (v DuraVec) Get(i Native) Native             { return DuraVal(v[i.(IntVal).Idx()]) }
func (v ErrorVec) Get(i Native) Native            { return ErrorVal{v[i.(IntVal).Idx()]} }
func (v FlagSet) Get(i Native) Native             { return BitFlag(v[i.(IntVal).Idx()]) }

func (v NilVec) Range(i, j int) Sliceable    { return NilVec(v[i:j]) }
func (v BoolVec) Range(i, j int) Sliceable   { return BoolVec(v[i:j]) }
func (v IntVec) Range(i, j int) Sliceable    { return IntVec(v[i:j]) }
func (v Int8Vec) Range(i, j int) Sliceable   { return Int8Vec(v[i:j]) }
func (v Int16Vec) Range(i, j int) Sliceable  { return Int16Vec(v[i:j]) }
func (v Int32Vec) Range(i, j int) Sliceable  { return Int32Vec(v[i:j]) }
func (v UintVec) Range(i, j int) Sliceable   { return UintVec(v[i:j]) }
func (v Uint8Vec) Range(i, j int) Sliceable  { return Uint8Vec(v[i:j]) }
func (v Uint16Vec) Range(i, j int) Sliceable { return Uint16Vec(v[i:j]) }
func (v Uint32Vec) Range(i, j int) Sliceable { return Uint32Vec(v[i:j]) }
func (v FltVec) Range(i, j int) Sliceable    { return FltVec(v[i:j]) }
func (v Flt32Vec) Range(i, j int) Sliceable  { return Flt32Vec(v[i:j]) }
func (v ImagVec) Range(i, j int) Sliceable   { return ImagVec(v[i:j]) }
func (v Imag64Vec) Range(i, j int) Sliceable { return Imag64Vec(v[i:j]) }
func (v ByteVec) Range(i, j int) Sliceable   { return ByteVec(v[i:j]) }
func (v RuneVec) Range(i, j int) Sliceable   { return RuneVec(v[i:j]) }
func (v BytesVec) Range(i, j int) Sliceable  { return BytesVec(v[i:j]) }
func (v StrVec) Range(i, j int) Sliceable    { return StrVec(v[i:j]) }
func (v BigIntVec) Range(i, j int) Sliceable { return BigIntVec(v[i:j]) }
func (v BigFltVec) Range(i, j int) Sliceable { return BigFltVec(v[i:j]) }
func (v RatioVec) Range(i, j int) Sliceable  { return RatioVec(v[i:j]) }
func (v TimeVec) Range(i, j int) Sliceable   { return TimeVec(v[i:j]) }
func (v DuraVec) Range(i, j int) Sliceable   { return DuraVec(v[i:j]) }
func (v ErrorVec) Range(i, j int) Sliceable  { return ErrorVec(v[i:j]) }
func (v FlagSet) Range(i, j int) Sliceable   { return FlagSet(v[i:j]) }

func (v InterfaceSlice) nat(i int) interface{}    { return v[i] }
func (v NilVec) Native(i int) struct{}            { return v[i] }
func (v BoolVec) Native(i int) bool               { return v[i] }
func (v IntVec) Native(i int) int                 { return v[i] }
func (v Int8Vec) Native(i int) int8               { return v[i] }
func (v Int16Vec) Native(i int) int16             { return v[i] }
func (v Int32Vec) Native(i int) int32             { return v[i] }
func (v UintVec) Native(i int) uint               { return v[i] }
func (v Uint8Vec) Native(i int) uint8             { return v[i] }
func (v Uint16Vec) Native(i int) uint16           { return v[i] }
func (v Uint32Vec) Native(i int) uint32           { return v[i] }
func (v FltVec) Native(i int) float64             { return v[i] }
func (v Flt32Vec) Native(i int) float32           { return v[i] }
func (v ImagVec) Native(i int) complex128         { return v[i] }
func (v Imag64Vec) Native(i int) complex64        { return v[i] }
func (v ByteVec) Native(i int) byte               { return v[i] }
func (v RuneVec) Native(i int) rune               { return v[i] }
func (v BytesVec) Native(i int) []byte            { return v[i] }
func (v StrVec) Native(i int) string              { return v[i] }
func (v BigIntVec) Native(i int) *big.Int         { return v[i] }
func (v BigFltVec) Native(i int) *big.Float       { return v[i] }
func (v RatioVec) Native(i int) *big.Rat          { return v[i] }
func (v TimeVec) Native(i int) time.Time          { return v[i] }
func (v DuraVec) Native(i int) time.Duration      { return v[i] }
func (v FlagSet) Native(i int) BitFlag            { return v[i] }
func (v ErrorVec) Native(i int) struct{ e error } { return struct{ e error }{v[i]} }

func (v NilVec) RangeNative(i, j int) []struct{}       { return NilVec(v[i:j]) }
func (v BoolVec) RangeNative(i, j int) []bool          { return BoolVec(v[i:j]) }
func (v IntVec) RangeNative(i, j int) []int            { return IntVec(v[i:j]) }
func (v Int8Vec) RangeNative(i, j int) []int8          { return Int8Vec(v[i:j]) }
func (v Int16Vec) RangeNative(i, j int) []int16        { return Int16Vec(v[i:j]) }
func (v Int32Vec) RangeNative(i, j int) []int32        { return Int32Vec(v[i:j]) }
func (v UintVec) RangeNative(i, j int) []uint          { return UintVec(v[i:j]) }
func (v Uint8Vec) RangeNative(i, j int) []uint8        { return Uint8Vec(v[i:j]) }
func (v Uint16Vec) RangeNative(i, j int) []uint16      { return Uint16Vec(v[i:j]) }
func (v Uint32Vec) RangeNative(i, j int) []uint32      { return Uint32Vec(v[i:j]) }
func (v FltVec) RangeNative(i, j int) []float64        { return FltVec(v[i:j]) }
func (v Flt32Vec) RangeNative(i, j int) []float32      { return Flt32Vec(v[i:j]) }
func (v ImagVec) RangeNative(i, j int) []complex128    { return ImagVec(v[i:j]) }
func (v Imag64Vec) RangeNative(i, j int) []complex64   { return Imag64Vec(v[i:j]) }
func (v ByteVec) RangeNative(i, j int) []byte          { return ByteVec(v[i:j]) }
func (v RuneVec) RangeNative(i, j int) []rune          { return RuneVec(v[i:j]) }
func (v BytesVec) RangeNative(i, j int) [][]byte       { return BytesVec(v[i:j]) }
func (v StrVec) RangeNative(i, j int) []string         { return StrVec(v[i:j]) }
func (v BigIntVec) RangeNative(i, j int) []*big.Int    { return BigIntVec(v[i:j]) }
func (v BigFltVec) RangeNative(i, j int) []*big.Float  { return BigFltVec(v[i:j]) }
func (v RatioVec) RangeNative(i, j int) []*big.Rat     { return RatioVec(v[i:j]) }
func (v TimeVec) RangeNative(i, j int) []time.Time     { return TimeVec(v[i:j]) }
func (v DuraVec) RangeNative(i, j int) []time.Duration { return DuraVec(v[i:j]) }
func (v ErrorVec) RangeNative(i, j int) []error        { return ErrorVec(v[i:j]) }
func (v FlagSet) RangeNative(i, j int) []BitFlag       { return FlagSet(v[i:j]) }

func (v NilVec) TypeNat() TyNat    { return Unboxed.TypeNat() }
func (v BoolVec) TypeNat() TyNat   { return Unboxed.TypeNat() }
func (v IntVec) TypeNat() TyNat    { return Unboxed.TypeNat() }
func (v Int8Vec) TypeNat() TyNat   { return Unboxed.TypeNat() }
func (v Int16Vec) TypeNat() TyNat  { return Unboxed.TypeNat() }
func (v Int32Vec) TypeNat() TyNat  { return Unboxed.TypeNat() }
func (v UintVec) TypeNat() TyNat   { return Unboxed.TypeNat() }
func (v Uint8Vec) TypeNat() TyNat  { return Unboxed.TypeNat() }
func (v Uint16Vec) TypeNat() TyNat { return Unboxed.TypeNat() }
func (v Uint32Vec) TypeNat() TyNat { return Unboxed.TypeNat() }
func (v FltVec) TypeNat() TyNat    { return Unboxed.TypeNat() }
func (v Flt32Vec) TypeNat() TyNat  { return Unboxed.TypeNat() }
func (v ImagVec) TypeNat() TyNat   { return Unboxed.TypeNat() }
func (v Imag64Vec) TypeNat() TyNat { return Unboxed.TypeNat() }
func (v ByteVec) TypeNat() TyNat   { return Unboxed.TypeNat() }
func (v RuneVec) TypeNat() TyNat   { return Unboxed.TypeNat() }
func (v BytesVec) TypeNat() TyNat  { return Unboxed.TypeNat() }
func (v StrVec) TypeNat() TyNat    { return Unboxed.TypeNat() }
func (v BigIntVec) TypeNat() TyNat { return Unboxed.TypeNat() }
func (v BigFltVec) TypeNat() TyNat { return Unboxed.TypeNat() }
func (v RatioVec) TypeNat() TyNat  { return Unboxed.TypeNat() }
func (v TimeVec) TypeNat() TyNat   { return Unboxed.TypeNat() }
func (v DuraVec) TypeNat() TyNat   { return Unboxed.TypeNat() }
func (v ErrorVec) TypeNat() TyNat  { return Unboxed.TypeNat() }
func (v FlagSet) TypeNat() TyNat   { return Unboxed.TypeNat() }

func (v NilVec) Slice() []Native {
	var slice = []Native{}
	for _, nat := range v {
		slice = append(slice, NilVal(nat))
	}
	return slice
}
func (v BoolVec) Slice() []Native {
	var slice = []Native{}
	for _, nat := range v {
		slice = append(slice, BoolVal(nat))
	}
	return slice
}
func (v IntVec) Slice() []Native {
	var slice = []Native{}
	for _, nat := range v {
		slice = append(slice, IntVal(nat))
	}
	return slice
}
func (v Int8Vec) Slice() []Native {
	var slice = []Native{}
	for _, nat := range v {
		slice = append(slice, Int8Val(nat))
	}
	return slice
}
func (v Int16Vec) Slice() []Native {
	var slice = []Native{}
	for _, nat := range v {
		slice = append(slice, Int16Val(nat))
	}
	return slice
}
func (v Int32Vec) Slice() []Native {
	var slice = []Native{}
	for _, nat := range v {
		slice = append(slice, Int32Val(nat))
	}
	return slice
}
func (v UintVec) Slice() []Native {
	var slice = []Native{}
	for _, nat := range v {
		slice = append(slice, UintVal(nat))
	}
	return slice
}
func (v Uint8Vec) Slice() []Native {
	var slice = []Native{}
	for _, nat := range v {
		slice = append(slice, Uint8Val(nat))
	}
	return slice
}
func (v Uint16Vec) Slice() []Native {
	var slice = []Native{}
	for _, nat := range v {
		slice = append(slice, Uint16Val(nat))
	}
	return slice
}
func (v Uint32Vec) Slice() []Native {
	var slice = []Native{}
	for _, nat := range v {
		slice = append(slice, Uint32Val(nat))
	}
	return slice
}
func (v FltVec) Slice() []Native {
	var slice = []Native{}
	for _, nat := range v {
		slice = append(slice, FltVal(nat))
	}
	return slice
}
func (v Flt32Vec) Slice() []Native {
	var slice = []Native{}
	for _, nat := range v {
		slice = append(slice, Flt32Val(nat))
	}
	return slice
}
func (v ImagVec) Slice() []Native {
	var slice = []Native{}
	for _, nat := range v {
		slice = append(slice, ImagVal(nat))
	}
	return slice
}
func (v Imag64Vec) Slice() []Native {
	var slice = []Native{}
	for _, nat := range v {
		slice = append(slice, Imag64Val(nat))
	}
	return slice
}
func (v ByteVec) Slice() []Native {
	var slice = []Native{}
	for _, nat := range v {
		slice = append(slice, ByteVal(nat))
	}
	return slice
}
func (v RuneVec) Slice() []Native {
	var slice = []Native{}
	for _, nat := range v {
		slice = append(slice, RuneVal(nat))
	}
	return slice
}
func (v BytesVec) Slice() []Native {
	var slice = []Native{}
	for _, nat := range v {
		slice = append(slice, BytesVal(nat))
	}
	return slice
}
func (v StrVec) Slice() []Native {
	var slice = []Native{}
	for _, nat := range v {
		slice = append(slice, StrVal(nat))
	}
	return slice
}
func (v BigIntVec) Slice() []Native {
	var slice = []Native{}
	for _, nat := range v {
		slice = append(slice, (*BigIntVal)(nat))
	}
	return slice
}
func (v BigFltVec) Slice() []Native {
	var slice = []Native{}
	for _, nat := range v {
		slice = append(slice, (*BigFltVal)(nat))
	}
	return slice
}
func (v RatioVec) Slice() []Native {
	var slice = []Native{}
	for _, nat := range v {
		slice = append(slice, (*RatioVal)(nat))
	}
	return slice
}
func (v TimeVec) Slice() []Native {
	var slice = []Native{}
	for _, nat := range v {
		slice = append(slice, TimeVal(nat))
	}
	return slice
}
func (v DuraVec) Slice() []Native {
	var slice = []Native{}
	for _, nat := range v {
		slice = append(slice, DuraVal(nat))
	}
	return slice
}
func (v ErrorVec) Slice() []Native {
	var slice = []Native{}
	for _, nat := range v {
		slice = append(slice, NewError(nat))
	}
	return slice
}
func (v FlagSet) Slice() []Native {
	var slice = []Native{}
	for _, nat := range v {
		slice = append(slice, BitFlag(nat))
	}
	return slice
}

func (v NilVec) Interface() Sliceable    { return v }
func (v BoolVec) Interface() Sliceable   { return v }
func (v IntVec) Interface() Sliceable    { return v }
func (v Int8Vec) Interface() Sliceable   { return v }
func (v Int16Vec) Interface() Sliceable  { return v }
func (v Int32Vec) Interface() Sliceable  { return v }
func (v UintVec) Interface() Sliceable   { return v }
func (v Uint8Vec) Interface() Sliceable  { return v }
func (v Uint16Vec) Interface() Sliceable { return v }
func (v Uint32Vec) Interface() Sliceable { return v }
func (v FltVec) Interface() Sliceable    { return v }
func (v Flt32Vec) Interface() Sliceable  { return v }
func (v ImagVec) Interface() Sliceable   { return v }
func (v Imag64Vec) Interface() Sliceable { return v }
func (v ByteVec) Interface() Sliceable   { return v }
func (v RuneVec) Interface() Sliceable   { return v }
func (v BytesVec) Interface() Sliceable  { return v }
func (v StrVec) Interface() Sliceable    { return v }
func (v BigIntVec) Interface() Sliceable { return v }
func (v BigFltVec) Interface() Sliceable { return v }
func (v RatioVec) Interface() Sliceable  { return v }
func (v TimeVec) Interface() Sliceable   { return v }
func (v DuraVec) Interface() Sliceable   { return v }
func (v ErrorVec) Interface() Sliceable  { return v }
func (v FlagSet) Interface() Sliceable   { return v }

func (v NilVec) TypeName() string    { return "[" + Nil.TypeName() + "]" }
func (v BoolVec) TypeName() string   { return "[" + Bool.TypeName() + "]" }
func (v IntVec) TypeName() string    { return "[" + Int.TypeName() + "]" }
func (v Int8Vec) TypeName() string   { return "[" + Int8.TypeName() + "]" }
func (v Int16Vec) TypeName() string  { return "[" + Int16.TypeName() + "]" }
func (v Int32Vec) TypeName() string  { return "[" + Int32.TypeName() + "]" }
func (v UintVec) TypeName() string   { return "[" + Uint.TypeName() + "]" }
func (v Uint8Vec) TypeName() string  { return "[" + Uint8.TypeName() + "]" }
func (v Uint16Vec) TypeName() string { return "[" + Uint16.TypeName() + "]" }
func (v Uint32Vec) TypeName() string { return "[" + Uint32.TypeName() + "]" }
func (v FltVec) TypeName() string    { return "[" + Float.TypeName() + "]" }
func (v Flt32Vec) TypeName() string  { return "[" + Flt32.TypeName() + "]" }
func (v ImagVec) TypeName() string   { return "[" + Imag.TypeName() + "]" }
func (v Imag64Vec) TypeName() string { return "[" + Imag64.TypeName() + "]" }
func (v ByteVec) TypeName() string   { return "[" + Byte.TypeName() + "]" }
func (v RuneVec) TypeName() string   { return "[" + Rune.TypeName() + "]" }
func (v BytesVec) TypeName() string  { return "[" + Bytes.TypeName() + "]" }
func (v StrVec) TypeName() string    { return "[" + String.TypeName() + "]" }
func (v BigIntVec) TypeName() string { return "[" + BigInt.TypeName() + "]" }
func (v BigFltVec) TypeName() string { return "[" + BigFlt.TypeName() + "]" }
func (v RatioVec) TypeName() string  { return "[" + Ratio.TypeName() + "]" }
func (v TimeVec) TypeName() string   { return "[" + Time.TypeName() + "]" }
func (v DuraVec) TypeName() string   { return "[" + Duration.TypeName() + "]" }
func (v ErrorVec) TypeName() string  { return "[" + Error.TypeName() + "]" }
func (v FlagSet) TypeName() string   { return "[" + Type.TypeName() + "]" }

func (v NilVec) ElemType() TyNat    { return Nil.TypeNat() }
func (v BoolVec) ElemType() TyNat   { return Bool.TypeNat() }
func (v IntVec) ElemType() TyNat    { return Int.TypeNat() }
func (v Int8Vec) ElemType() TyNat   { return Int8.TypeNat() }
func (v Int16Vec) ElemType() TyNat  { return Int16.TypeNat() }
func (v Int32Vec) ElemType() TyNat  { return Int32.TypeNat() }
func (v UintVec) ElemType() TyNat   { return Uint.TypeNat() }
func (v Uint8Vec) ElemType() TyNat  { return Uint8.TypeNat() }
func (v Uint16Vec) ElemType() TyNat { return Uint16.TypeNat() }
func (v Uint32Vec) ElemType() TyNat { return Uint32.TypeNat() }
func (v FltVec) ElemType() TyNat    { return Float.TypeNat() }
func (v Flt32Vec) ElemType() TyNat  { return Flt32.TypeNat() }
func (v ImagVec) ElemType() TyNat   { return Imag.TypeNat() }
func (v Imag64Vec) ElemType() TyNat { return Imag64.TypeNat() }
func (v ByteVec) ElemType() TyNat   { return Byte.TypeNat() }
func (v RuneVec) ElemType() TyNat   { return Rune.TypeNat() }
func (v BytesVec) ElemType() TyNat  { return Bytes.TypeNat() }
func (v StrVec) ElemType() TyNat    { return String.TypeNat() }
func (v BigIntVec) ElemType() TyNat { return BigInt.TypeNat() }
func (v BigFltVec) ElemType() TyNat { return BigFlt.TypeNat() }
func (v RatioVec) ElemType() TyNat  { return Ratio.TypeNat() }
func (v TimeVec) ElemType() TyNat   { return Time.TypeNat() }
func (v DuraVec) ElemType() TyNat   { return Duration.TypeNat() }
func (v ErrorVec) ElemType() TyNat  { return Error.TypeNat() }
func (v FlagSet) ElemType() TyNat   { return Type.TypeNat() }

func (v NilVec) Null() Native    { return NilVec([]struct{}{}) }
func (v BoolVec) Null() Native   { return BoolVec([]bool{}) }
func (v IntVec) Null() Native    { return IntVec([]int{}) }
func (v Int8Vec) Null() Native   { return Int8Vec([]int8{}) }
func (v Int16Vec) Null() Native  { return Int16Vec([]int16{}) }
func (v Int32Vec) Null() Native  { return Int32Vec([]int32{}) }
func (v UintVec) Null() Native   { return UintVec([]uint{}) }
func (v Uint8Vec) Null() Native  { return Uint8Vec([]uint8{}) }
func (v Uint16Vec) Null() Native { return Uint16Vec([]uint16{}) }
func (v Uint32Vec) Null() Native { return Uint32Vec([]uint32{}) }
func (v FltVec) Null() Native    { return FltVec([]float64{}) }
func (v Flt32Vec) Null() Native  { return Flt32Vec([]float32{}) }
func (v ImagVec) Null() Native   { return ImagVec([]complex128{}) }
func (v Imag64Vec) Null() Native { return Imag64Vec([]complex64{}) }
func (v ByteVec) Null() Native   { return ByteVec([]byte{}) }
func (v RuneVec) Null() Native   { return RuneVec([]rune{}) }
func (v BytesVec) Null() Native  { return BytesVec([][]byte{}) }
func (v StrVec) Null() Native    { return StrVec([]string{}) }
func (v BigIntVec) Null() Native { return BigIntVec([]*big.Int{}) }
func (v BigFltVec) Null() Native { return BigFltVec([]*big.Float{}) }
func (v RatioVec) Null() Native  { return RatioVec([]*big.Rat{}) }
func (v TimeVec) Null() Native   { return TimeVec([]time.Time{}) }
func (v DuraVec) Null() Native   { return DuraVec([]time.Duration{}) }
func (v ErrorVec) Null() Native  { return ErrorVec([]error{}) }
func (v FlagSet) Null() Native   { return FlagSet([]BitFlag{}) }

func (v NilVec) Len() int    { return len(v) }
func (v BoolVec) Len() int   { return len(v) }
func (v IntVec) Len() int    { return len(v) }
func (v Int8Vec) Len() int   { return len(v) }
func (v Int16Vec) Len() int  { return len(v) }
func (v Int32Vec) Len() int  { return len(v) }
func (v UintVec) Len() int   { return len(v) }
func (v Uint8Vec) Len() int  { return len(v) }
func (v Uint16Vec) Len() int { return len(v) }
func (v Uint32Vec) Len() int { return len(v) }
func (v FltVec) Len() int    { return len(v) }
func (v Flt32Vec) Len() int  { return len(v) }
func (v ImagVec) Len() int   { return len(v) }
func (v Imag64Vec) Len() int { return len(v) }
func (v ByteVec) Len() int   { return len(v) }
func (v RuneVec) Len() int   { return len(v) }
func (v BytesVec) Len() int  { return len(v) }
func (v StrVec) Len() int    { return len(v) }
func (v BigIntVec) Len() int { return len(v) }
func (v BigFltVec) Len() int { return len(v) }
func (v RatioVec) Len() int  { return len(v) }
func (v TimeVec) Len() int   { return len(v) }
func (v DuraVec) Len() int   { return len(v) }
func (v ErrorVec) Len() int  { return len(v) }
func (v FlagSet) Len() int   { return len(v) }

func (v NilVec) Empty() bool {
	if v.Len() == 0 {
		return true
	}
	return false
}
func (v BoolVec) Empty() bool {
	if v.Len() == 0 {
		return true
	}
	return false
}
func (v IntVec) Empty() bool {
	if v.Len() == 0 {
		return true
	}
	return false
}
func (v Int8Vec) Empty() bool {
	if v.Len() == 0 {
		return true
	}
	return false
}
func (v Int16Vec) Empty() bool {
	if v.Len() == 0 {
		return true
	}
	return false
}
func (v Int32Vec) Empty() bool {
	if v.Len() == 0 {
		return true
	}
	return false
}
func (v UintVec) Empty() bool {
	if v.Len() == 0 {
		return true
	}
	return false
}
func (v Uint8Vec) Empty() bool {
	if v.Len() == 0 {
		return true
	}
	return false
}
func (v Uint16Vec) Empty() bool {
	if v.Len() == 0 {
		return true
	}
	return false
}
func (v Uint32Vec) Empty() bool {
	if v.Len() == 0 {
		return true
	}
	return false
}
func (v FltVec) Empty() bool {
	if v.Len() == 0 {
		return true
	}
	return false
}
func (v Flt32Vec) Empty() bool {
	if v.Len() == 0 {
		return true
	}
	return false
}
func (v ImagVec) Empty() bool {
	if v.Len() == 0 {
		return true
	}
	return false
}
func (v Imag64Vec) Empty() bool {
	if v.Len() == 0 {
		return true
	}
	return false
}
func (v ByteVec) Empty() bool {
	if v.Len() == 0 {
		return true
	}
	return false
}
func (v RuneVec) Empty() bool {
	if v.Len() == 0 {
		return true
	}
	return false
}
func (v BytesVec) Empty() bool {
	if v.Len() == 0 {
		return true
	}
	return false
}
func (v StrVec) Empty() bool {
	if v.Len() == 0 {
		return true
	}
	return false
}
func (v BigIntVec) Empty() bool {
	if v.Len() == 0 {
		return true
	}
	return false
}
func (v BigFltVec) Empty() bool {
	if v.Len() == 0 {
		return true
	}
	return false
}
func (v RatioVec) Empty() bool {
	if v.Len() == 0 {
		return true
	}
	return false
}
func (v TimeVec) Empty() bool {
	if v.Len() == 0 {
		return true
	}
	return false
}
func (v DuraVec) Empty() bool {
	if v.Len() == 0 {
		return true
	}
	return false
}
func (v ErrorVec) Empty() bool {
	if v.Len() == 0 {
		return true
	}
	return false
}
func (v FlagSet) Empty() bool {
	if v.Len() == 0 {
		return true
	}
	return false
}

func (v NilVec) Copy() Native {
	var d = NilVec{}
	for _, val := range v {
		d = append(d, val)
	}
	return d
}

func (v BoolVec) Copy() Native {
	var d = BoolVec{}
	for _, val := range v {
		d = append(d, val)
	}
	return d
}

func (v IntVec) Copy() Native {
	var d = IntVec{}
	for _, val := range v {
		d = append(d, val)
	}
	return d
}

func (v Int8Vec) Copy() Native {
	var d = Int8Vec{}
	for _, val := range v {
		d = append(d, val)
	}
	return d
}

func (v Int16Vec) Copy() Native {
	var d = Int16Vec{}
	for _, val := range v {
		d = append(d, val)
	}
	return d
}

func (v Int32Vec) Copy() Native {
	var d = Int32Vec{}
	for _, val := range v {
		d = append(d, val)
	}
	return d
}
func (v UintVec) Copy() Native {
	var d = UintVec{}
	for _, val := range v {
		d = append(d, val)
	}
	return d
}
func (v Uint8Vec) Copy() Native {
	var d = Uint8Vec{}
	for _, val := range v {
		d = append(d, val)
	}
	return d
}
func (v Uint16Vec) Copy() Native {
	var d = Uint16Vec{}
	for _, val := range v {
		d = append(d, val)
	}
	return d
}
func (v Uint32Vec) Copy() Native {
	var d = Uint32Vec{}
	for _, val := range v {
		d = append(d, val)
	}
	return d
}
func (v FltVec) Copy() Native {
	var d = FltVec{}
	for _, val := range v {
		d = append(d, val)
	}
	return d
}
func (v Flt32Vec) Copy() Native {
	var d = Flt32Vec{}
	for _, val := range v {
		d = append(d, val)
	}
	return d
}
func (v ImagVec) Copy() Native {
	var d = ImagVec{}
	for _, val := range v {
		d = append(d, val)
	}
	return d
}
func (v Imag64Vec) Copy() Native {
	var d = Imag64Vec{}
	for _, val := range v {
		d = append(d, val)
	}
	return d
}
func (v ByteVec) Copy() Native {
	var d = ByteVec{}
	for _, val := range v {
		d = append(d, val)
	}
	return d
}
func (v RuneVec) Copy() Native {
	var d = RuneVec{}
	for _, val := range v {
		d = append(d, val)
	}
	return d
}
func (v BytesVec) Copy() Native {
	var d = BytesVec{}
	for _, val := range v {
		d = append(d, val)
	}
	return d
}
func (v StrVec) Copy() Native {
	var d = StrVec{}
	for _, val := range v {
		d = append(d, val)
	}
	return d
}
func (v BigIntVec) Copy() Native {
	var d = BigIntVec{}
	for _, val := range v {
		d = append(d, val)
	}
	return d
}
func (v BigFltVec) Copy() Native {
	var d = BigFltVec{}
	for _, val := range v {
		d = append(d, val)
	}
	return d
}
func (v RatioVec) Copy() Native {
	var d = RatioVec{}
	for _, val := range v {
		d = append(d, val)
	}
	return d
}
func (v TimeVec) Copy() Native {
	var d = TimeVec{}
	for _, val := range v {
		d = append(d, val)
	}
	return d
}
func (v DuraVec) Copy() Native {
	var d = DuraVec{}
	for _, val := range v {
		d = append(d, val)
	}
	return d
}
func (v ErrorVec) Copy() Native {
	var d = ErrorVec{}
	for _, val := range v {
		d = append(d, val)
	}
	return d
}
func (v FlagSet) Copy() Native {
	var d = FlagSet{}
	for _, val := range v {
		d = append(d, val)
	}
	return d
}
