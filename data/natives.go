package data

import (
	"math/big"
	"time"
)

func ConNativeSlice(flag BitFlag, data ...Native) Sliceable {

	var d Sliceable

	switch TyNative(flag) {
	case Nil:
		d = NilVec{}
		for _, _ = range data {
			d = append(d.(NilVec), struct{}{})
		}

	case Bool:
		d = BoolVec{}
		for _, dat := range data {
			d = append(d.(BoolVec), bool(dat.(BoolVal)))
		}

	case Int:
		d = IntVec{}
		for _, dat := range data {
			d = append(d.(IntVec), int(dat.(IntVal)))
		}

	case Int8:
		d = Int8Vec{}
		for _, dat := range data {
			d = append(d.(Int8Vec), int8(dat.(Int8Val)))
		}

	case Int16:
		d = Int16Vec{}
		for _, dat := range data {
			d = append(d.(Int16Vec), int16(dat.(Int16Val)))
		}

	case Int32:
		d = Int32Vec{}
		for _, dat := range data {
			d = append(d.(Int32Vec), int32(dat.(Int32Val)))
		}

	case Uint:
		d = UintVec{}
		for _, dat := range data {
			d = append(d.(UintVec), uint(dat.(UintVal)))
		}

	case Uint8:
		d = Uint8Vec{}
		for _, dat := range data {
			d = append(d.(Uint8Vec), uint8(dat.(Uint8Val)))
		}

	case Uint16:
		d = Uint16Vec{}
		for _, dat := range data {
			d = append(d.(Uint16Vec), uint16(dat.(Uint16Val)))
		}

	case Uint32:
		d = Uint32Vec{}
		for _, dat := range data {
			d = append(d.(Uint32Vec), uint32(dat.(Uint32Val)))
		}

	case Float:
		d = FltVec{}
		for _, dat := range data {
			d = append(d.(FltVec), float64(dat.(FltVal)))
		}

	case Flt32:
		d = Flt32Vec{}
		for _, dat := range data {
			d = append(d.(Flt32Vec), float32(dat.(Flt32Val)))
		}

	case Imag:
		d = ImagVec{}
		for _, dat := range data {
			d = append(d.(ImagVec), complex128(dat.(ImagVal)))
		}

	case Imag64:
		d = Imag64Vec{}
		for _, dat := range data {
			d = append(d.(Imag64Vec), complex64(dat.(Imag64Val)))
		}

	case Byte:
		d = ByteVec{}
		for _, dat := range data {
			d = append(d.(ByteVec), byte(dat.(ByteVal)))
		}

	case Rune:
		d = RuneVec{}
		for _, dat := range data {
			d = append(d.(RuneVec), rune(dat.(RuneVal)))
		}

	case Bytes:
		d = BytesVec{}
		for _, dat := range data {
			d = append(d.(BytesVec), []byte(dat.(BytesVal)))
		}

	case String:
		d = StrVec{}
		for _, dat := range data {
			d = append(d.(StrVec), string(dat.(StrVal)))
		}

	case BigInt:
		d = BigIntVec{}
		for _, dat := range data {
			bi := dat.(BigIntVal)
			d = append(d.(BigIntVec), big.NewInt(((*big.Int)(&bi)).Int64()))
		}

	case BigFlt:
		d = BigFltVec{}
		for _, dat := range data {
			bf := dat.(BigFltVal)
			f, _ := ((*big.Float)(&bf)).Float64()
			d = append(d.(BigFltVec), big.NewFloat(f))
		}

	case Ratio:
		d = RatioVec{}
		for _, dat := range data {
			rat := dat.(RatioVal)
			d = append(d.(RatioVec), big.NewRat(
				((*big.Rat)(&rat)).Num().Int64(),
				((*big.Rat)(&rat)).Denom().Int64(),
			))
		}

	case Time:
		d = TimeVec{}
		for _, dat := range data {
			d = append(d.(TimeVec), time.Time(dat.(TimeVal)))
		}

	case Duration:
		d = DuraVec{}
		for _, dat := range data {
			d = append(d.(DuraVec), time.Duration(dat.(DuraVal)))
		}

	case Error:
		d = ErrorVec{}
		for _, dat := range data {
			d = append(d.(ErrorVec), error(dat.(ErrorVal).E))
		}
	}
	return d
}

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
func (v Uint8Vec) GetInt(i int) Uint8Val          { return Uint8Val(v[i]) }
func (v Uint16Vec) GetInt(i int) Native           { return Uint16Val(v[i]) }
func (v Uint32Vec) GetInt(i int) Native           { return Uint32Val(v[i]) }
func (v FltVec) GetInt(i int) Native              { return FltVal(v[i]) }
func (v Flt32Vec) GetInt(i int) Flt32Val          { return Flt32Val(v[i]) }
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

func (v InterfaceSlice) Get(i Native) interface{} { return v[i.(IntVal).Int()] }
func (v NilVec) Get(i Native) Native              { return NilVal(v[i.(IntVal).Int()]) }
func (v BoolVec) Get(i Native) Native             { return BoolVal(v[i.(IntVal).Int()]) }
func (v IntVec) Get(i Native) Native              { return IntVal(v[i.(IntVal).Int()]) }
func (v Int8Vec) Get(i Native) Native             { return Int8Val(v[i.(IntVal).Int()]) }
func (v Int16Vec) Get(i Native) Native            { return Int16Val(v[i.(IntVal).Int()]) }
func (v Int32Vec) Get(i Native) Native            { return Int32Val(v[i.(IntVal).Int()]) }
func (v UintVec) Get(i Native) Native             { return UintVal(v[i.(IntVal).Int()]) }
func (v Uint8Vec) Get(i Native) Uint8Val          { return Uint8Val(v[i.(IntVal).Int()]) }
func (v Uint16Vec) Get(i Native) Native           { return Uint16Val(v[i.(IntVal).Int()]) }
func (v Uint32Vec) Get(i Native) Native           { return Uint32Val(v[i.(IntVal).Int()]) }
func (v FltVec) Get(i Native) Native              { return FltVal(v[i.(IntVal).Int()]) }
func (v Flt32Vec) Get(i Native) Flt32Val          { return Flt32Val(v[i.(IntVal).Int()]) }
func (v ImagVec) Get(i Native) Native             { return ImagVal(v[i.(IntVal).Int()]) }
func (v Imag64Vec) Get(i Native) Native           { return Imag64Val(v[i.(IntVal).Int()]) }
func (v ByteVec) Get(i Native) Native             { return ByteVal(v[i.(IntVal).Int()]) }
func (v RuneVec) Get(i Native) Native             { return RuneVal(v[i.(IntVal).Int()]) }
func (v BytesVec) Get(i Native) Native            { return BytesVal(v[i.(IntVal).Int()]) }
func (v StrVec) Get(i Native) Native              { return StrVal(v[i.(IntVal).Int()]) }
func (v BigIntVec) Get(i Native) Native           { return BigIntVal((*(*big.Int)(v[i.(IntVal).Int()]))) }
func (v BigFltVec) Get(i Native) Native           { return BigFltVal((*(*big.Float)(v[i.(IntVal).Int()]))) }
func (v RatioVec) Get(i Native) Native            { return RatioVal((*(*big.Rat)(v[i.(IntVal).Int()]))) }
func (v TimeVec) Get(i Native) Native             { return TimeVal(v[i.(IntVal).Int()]) }
func (v DuraVec) Get(i Native) Native             { return DuraVal(v[i.(IntVal).Int()]) }
func (v ErrorVec) Get(i Native) Native            { return ErrorVal{v[i.(IntVal).Int()]} }

func (v InterfaceSlice) Range(i, j int) interface{} { return v[i] }
func (v NilVec) Range(i, j int) NilVec              { return NilVec(v[i:j]) }
func (v BoolVec) Range(i, j int) BoolVec            { return BoolVec(v[i:j]) }
func (v IntVec) Range(i, j int) IntVec              { return IntVec(v[i:j]) }
func (v Int8Vec) Range(i, j int) Int8Vec            { return Int8Vec(v[i:j]) }
func (v Int16Vec) Range(i, j int) Int16Vec          { return Int16Vec(v[i:j]) }
func (v Int32Vec) Range(i, j int) Int32Vec          { return Int32Vec(v[i:j]) }
func (v UintVec) Range(i, j int) UintVec            { return UintVec(v[i:j]) }
func (v Uint8Vec) Range(i, j int) Uint8Vec          { return Uint8Vec(v[i:j]) }
func (v Uint16Vec) Range(i, j int) Uint16Vec        { return Uint16Vec(v[i:j]) }
func (v Uint32Vec) Range(i, j int) Uint32Vec        { return Uint32Vec(v[i:j]) }
func (v FltVec) Range(i, j int) FltVec              { return FltVec(v[i:j]) }
func (v Flt32Vec) Range(i, j int) Flt32Vec          { return Flt32Vec(v[i:j]) }
func (v ImagVec) Range(i, j int) ImagVec            { return ImagVec(v[i:j]) }
func (v Imag64Vec) Range(i, j int) Imag64Vec        { return Imag64Vec(v[i:j]) }
func (v ByteVec) Range(i, j int) ByteVec            { return ByteVec(v[i:j]) }
func (v RuneVec) Range(i, j int) RuneVec            { return RuneVec(v[i:j]) }
func (v BytesVec) Range(i, j int) BytesVec          { return BytesVec(v[i:j]) }
func (v StrVec) Range(i, j int) StrVec              { return StrVec(v[i:j]) }
func (v BigIntVec) Range(i, j int) BigIntVec        { return BigIntVec(v[i:j]) }
func (v BigFltVec) Range(i, j int) BigFltVec        { return BigFltVec(v[i:j]) }
func (v RatioVec) Range(i, j int) RatioVec          { return RatioVec(v[i:j]) }
func (v TimeVec) Range(i, j int) TimeVec            { return TimeVec(v[i:j]) }
func (v DuraVec) Range(i, j int) DuraVec            { return DuraVec(v[i:j]) }
func (v ErrorVec) Range(i, j int) ErrorVec          { return ErrorVec(v[i:j]) }

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
func (v ErrorVec) Native(i int) struct{ e error } { return struct{ e error }{v[i]} }
func (v FlagSet) Native(i int) BitFlag            { return v[i] }

func (v NilVec) intf(i int) interface{}    { return v[i] }
func (v BoolVec) intf(i int) interface{}   { return v[i] }
func (v IntVec) intf(i int) interface{}    { return v[i] }
func (v Int8Vec) intf(i int) interface{}   { return v[i] }
func (v Int16Vec) intf(i int) interface{}  { return v[i] }
func (v Int32Vec) intf(i int) interface{}  { return v[i] }
func (v UintVec) intf(i int) interface{}   { return v[i] }
func (v Uint8Vec) intf(i int) interface{}  { return v[i] }
func (v Uint16Vec) intf(i int) interface{} { return v[i] }
func (v Uint32Vec) intf(i int) interface{} { return v[i] }
func (v FltVec) intf(i int) interface{}    { return v[i] }
func (v Flt32Vec) intf(i int) interface{}  { return v[i] }
func (v ImagVec) intf(i int) interface{}   { return v[i] }
func (v Imag64Vec) intf(i int) interface{} { return v[i] }
func (v ByteVec) intf(i int) interface{}   { return v[i] }
func (v RuneVec) intf(i int) interface{}   { return v[i] }
func (v BytesVec) intf(i int) interface{}  { return v[i] }
func (v StrVec) intf(i int) interface{}    { return v[i] }
func (v BigIntVec) intf(i int) interface{} { return v[i] }
func (v BigFltVec) intf(i int) interface{} { return v[i] }
func (v RatioVec) intf(i int) interface{}  { return v[i] }
func (v TimeVec) intf(i int) interface{}   { return v[i] }
func (v DuraVec) intf(i int) interface{}   { return v[i] }
func (v ErrorVec) intf(i int) interface{}  { return v[i] }
func (v FlagSet) intf(i int) interface{}   { return v[i] }

func (v NilVec) ints(i, j int) interface{}    { return v[i:j] }
func (v BoolVec) ints(i, j int) interface{}   { return v[i:j] }
func (v IntVec) ints(i, j int) interface{}    { return v[i:j] }
func (v Int8Vec) ints(i, j int) interface{}   { return v[i:j] }
func (v Int16Vec) ints(i, j int) interface{}  { return v[i:j] }
func (v Int32Vec) ints(i, j int) interface{}  { return v[i:j] }
func (v UintVec) ints(i, j int) interface{}   { return v[i:j] }
func (v Uint8Vec) ints(i, j int) interface{}  { return v[i:j] }
func (v Uint16Vec) ints(i, j int) interface{} { return v[i:j] }
func (v Uint32Vec) ints(i, j int) interface{} { return v[i:j] }
func (v FltVec) ints(i, j int) interface{}    { return v[i:j] }
func (v Flt32Vec) ints(i, j int) interface{}  { return v[i:j] }
func (v ImagVec) ints(i, j int) interface{}   { return v[i:j] }
func (v Imag64Vec) ints(i, j int) interface{} { return v[i:j] }
func (v ByteVec) ints(i, j int) interface{}   { return v[i:j] }
func (v RuneVec) ints(i, j int) interface{}   { return v[i:j] }
func (v BytesVec) ints(i, j int) interface{}  { return v[i:j] }
func (v StrVec) ints(i, j int) interface{}    { return v[i:j] }
func (v BigIntVec) ints(i, j int) interface{} { return v[i:j] }
func (v BigFltVec) ints(i, j int) interface{} { return v[i:j] }
func (v RatioVec) ints(i, j int) interface{}  { return v[i:j] }
func (v TimeVec) ints(i, j int) interface{}   { return v[i:j] }
func (v DuraVec) ints(i, j int) interface{}   { return v[i:j] }
func (v ErrorVec) ints(i, j int) interface{}  { return v[i:j] }
func (v FlagSet) ints(i, j int) interface{}   { return v[i:j] }

func (v NilVec) NativesRange(i, j int) []struct{}       { return NilVec(v[i:j]) }
func (v BoolVec) NativesRange(i, j int) []bool          { return BoolVec(v[i:j]) }
func (v IntVec) NativesRange(i, j int) []int            { return IntVec(v[i:j]) }
func (v Int8Vec) NativesRange(i, j int) []int8          { return Int8Vec(v[i:j]) }
func (v Int16Vec) NativesRange(i, j int) []int16        { return Int16Vec(v[i:j]) }
func (v Int32Vec) NativesRange(i, j int) []int32        { return Int32Vec(v[i:j]) }
func (v UintVec) NativesRange(i, j int) []uint          { return UintVec(v[i:j]) }
func (v Uint8Vec) NativesRange(i, j int) []uint8        { return Uint8Vec(v[i:j]) }
func (v Uint16Vec) NativesRange(i, j int) []uint16      { return Uint16Vec(v[i:j]) }
func (v Uint32Vec) NativesRange(i, j int) []uint32      { return Uint32Vec(v[i:j]) }
func (v FltVec) NativesRange(i, j int) []float64        { return FltVec(v[i:j]) }
func (v Flt32Vec) NativesRange(i, j int) []float32      { return Flt32Vec(v[i:j]) }
func (v ImagVec) NativesRange(i, j int) []complex128    { return ImagVec(v[i:j]) }
func (v Imag64Vec) NativesRange(i, j int) []complex64   { return Imag64Vec(v[i:j]) }
func (v ByteVec) NativesRange(i, j int) []byte          { return ByteVec(v[i:j]) }
func (v RuneVec) NativesRange(i, j int) []rune          { return RuneVec(v[i:j]) }
func (v BytesVec) NativesRange(i, j int) [][]byte       { return BytesVec(v[i:j]) }
func (v StrVec) NativesRange(i, j int) []string         { return StrVec(v[i:j]) }
func (v BigIntVec) NativesRange(i, j int) []*big.Int    { return BigIntVec(v[i:j]) }
func (v BigFltVec) NativesRange(i, j int) []*big.Float  { return BigFltVec(v[i:j]) }
func (v RatioVec) NativesRange(i, j int) []*big.Rat     { return RatioVec(v[i:j]) }
func (v TimeVec) NativesRange(i, j int) []time.Time     { return TimeVec(v[i:j]) }
func (v DuraVec) NativesRange(i, j int) []time.Duration { return DuraVec(v[i:j]) }
func (v ErrorVec) NativesRange(i, j int) []error        { return ErrorVec(v[i:j]) }
func (v FlagSet) NativesRange(i, j int) []BitFlag       { return FlagSet(v[i:j]) }

func (v NilVec) TypeNat() TyNative    { return Vector.TypeNat() | Nil.TypeNat() }
func (v BoolVec) TypeNat() TyNative   { return Vector.TypeNat() | Bool.TypeNat() }
func (v IntVec) TypeNat() TyNative    { return Vector.TypeNat() | Int.TypeNat() }
func (v Int8Vec) TypeNat() TyNative   { return Vector.TypeNat() | Int8.TypeNat() }
func (v Int16Vec) TypeNat() TyNative  { return Vector.TypeNat() | Int16.TypeNat() }
func (v Int32Vec) TypeNat() TyNative  { return Vector.TypeNat() | Int32.TypeNat() }
func (v UintVec) TypeNat() TyNative   { return Vector.TypeNat() | Uint.TypeNat() }
func (v Uint8Vec) TypeNat() TyNative  { return Vector.TypeNat() | Uint8.TypeNat() }
func (v Uint16Vec) TypeNat() TyNative { return Vector.TypeNat() | Uint16.TypeNat() }
func (v Uint32Vec) TypeNat() TyNative { return Vector.TypeNat() | Uint32.TypeNat() }
func (v FltVec) TypeNat() TyNative    { return Vector.TypeNat() | Float.TypeNat() }
func (v Flt32Vec) TypeNat() TyNative  { return Vector.TypeNat() | Flt32.TypeNat() }
func (v ImagVec) TypeNat() TyNative   { return Vector.TypeNat() | Imag.TypeNat() }
func (v Imag64Vec) TypeNat() TyNative { return Vector.TypeNat() | Imag64.TypeNat() }
func (v ByteVec) TypeNat() TyNative   { return Vector.TypeNat() | Byte.TypeNat() }
func (v RuneVec) TypeNat() TyNative   { return Vector.TypeNat() | Rune.TypeNat() }
func (v BytesVec) TypeNat() TyNative  { return Vector.TypeNat() | Bytes.TypeNat() }
func (v StrVec) TypeNat() TyNative    { return Vector.TypeNat() | String.TypeNat() }
func (v BigIntVec) TypeNat() TyNative { return Vector.TypeNat() | BigInt.TypeNat() }
func (v BigFltVec) TypeNat() TyNative { return Vector.TypeNat() | BigFlt.TypeNat() }
func (v RatioVec) TypeNat() TyNative  { return Vector.TypeNat() | Ratio.TypeNat() }
func (v TimeVec) TypeNat() TyNative   { return Vector.TypeNat() | Time.TypeNat() }
func (v DuraVec) TypeNat() TyNative   { return Vector.TypeNat() | Duration.TypeNat() }
func (v ErrorVec) TypeNat() TyNative  { return Vector.TypeNat() | Error.TypeNat() }
func (v FlagSet) TypeNat() TyNative   { return Vector.TypeNat() | Flag.TypeNat() }

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

func (v NilVec) Eval(...Native) Native    { return v }
func (v BoolVec) Eval(...Native) Native   { return v }
func (v IntVec) Eval(...Native) Native    { return v }
func (v Int8Vec) Eval(...Native) Native   { return v }
func (v Int16Vec) Eval(...Native) Native  { return v }
func (v Int32Vec) Eval(...Native) Native  { return v }
func (v UintVec) Eval(...Native) Native   { return v }
func (v Uint8Vec) Eval(...Native) Native  { return v }
func (v Uint16Vec) Eval(...Native) Native { return v }
func (v Uint32Vec) Eval(...Native) Native { return v }
func (v FltVec) Eval(...Native) Native    { return v }
func (v Flt32Vec) Eval(...Native) Native  { return v }
func (v ImagVec) Eval(...Native) Native   { return v }
func (v Imag64Vec) Eval(...Native) Native { return v }
func (v ByteVec) Eval(...Native) Native   { return v }
func (v RuneVec) Eval(...Native) Native   { return v }
func (v BytesVec) Eval(...Native) Native  { return v }
func (v StrVec) Eval(...Native) Native    { return v }
func (v BigIntVec) Eval(...Native) Native { return v }
func (v BigFltVec) Eval(...Native) Native { return v }
func (v RatioVec) Eval(...Native) Native  { return v }
func (v TimeVec) Eval(...Native) Native   { return v }
func (v DuraVec) Eval(...Native) Native   { return v }
func (v ErrorVec) Eval(...Native) Native  { return v }
func (v FlagSet) Eval(...Native) Native   { return v }

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

func (v NilVec) Slice() []Native {
	var d = []Native{}
	for _, val := range v {
		d = append(d, New(val))
	}
	return d
}
func (v BoolVec) Slice() []Native {
	var d = []Native{}
	for _, val := range v {
		d = append(d, New(val))
	}
	return d
}
func (v IntVec) Slice() []Native {
	var d = []Native{}
	for _, val := range v {
		d = append(d, New(val))
	}
	return d
}
func (v Int8Vec) Slice() []Native {
	var d = []Native{}
	for _, val := range v {
		d = append(d, New(val))
	}
	return d
}
func (v Int16Vec) Slice() []Native {
	var d = []Native{}
	for _, val := range v {
		d = append(d, New(val))
	}
	return d
}
func (v Int32Vec) Slice() []Native {
	var d = []Native{}
	for _, val := range v {
		d = append(d, New(val))
	}
	return d
}
func (v UintVec) Slice() []Native {
	var d = []Native{}
	for _, val := range v {
		d = append(d, New(val))
	}
	return d
}
func (v Uint8Vec) Slice() []Native {
	var d = []Native{}
	for _, val := range v {
		d = append(d, New(val))
	}
	return d
}
func (v Uint16Vec) Slice() []Native {
	var d = []Native{}
	for _, val := range v {
		d = append(d, New(val))
	}
	return d
}
func (v Uint32Vec) Slice() []Native {
	var d = []Native{}
	for _, val := range v {
		d = append(d, New(val))
	}
	return d
}
func (v FltVec) Slice() []Native {
	var d = []Native{}
	for _, val := range v {
		d = append(d, New(val))
	}
	return d
}
func (v Flt32Vec) Slice() []Native {
	var d = []Native{}
	for _, val := range v {
		d = append(d, New(val))
	}
	return d
}
func (v ImagVec) Slice() []Native {
	var d = []Native{}
	for _, val := range v {
		d = append(d, New(val))
	}
	return d
}
func (v Imag64Vec) Slice() []Native {
	var d = []Native{}
	for _, val := range v {
		d = append(d, New(val))
	}
	return d
}
func (v ByteVec) Slice() []Native {
	var d = []Native{}
	for _, val := range v {
		d = append(d, New(val))
	}
	return d
}
func (v RuneVec) Slice() []Native {
	var d = []Native{}
	for _, val := range v {
		d = append(d, New(val))
	}
	return d
}
func (v BytesVec) Slice() []Native {
	var d = []Native{}
	for _, val := range v {
		d = append(d, New(val))
	}
	return d
}
func (v StrVec) Slice() []Native {
	var d = []Native{}
	for _, val := range v {
		d = append(d, New(val))
	}
	return d
}
func (v BigIntVec) Slice() []Native {
	var d = []Native{}
	for _, val := range v {
		d = append(d, New(val))
	}
	return d
}
func (v BigFltVec) Slice() []Native {
	var d = []Native{}
	for _, val := range v {
		d = append(d, New(val))
	}
	return d
}
func (v RatioVec) Slice() []Native {
	var d = []Native{}
	for _, val := range v {
		d = append(d, New(val))
	}
	return d
}
func (v TimeVec) Slice() []Native {
	var d = []Native{}
	for _, val := range v {
		d = append(d, New(val))
	}
	return d
}
func (v DuraVec) Slice() []Native {
	var d = []Native{}
	for _, val := range v {
		d = append(d, New(val))
	}
	return d
}
func (v ErrorVec) Slice() []Native {
	var d = []Native{}
	for _, val := range v {
		d = append(d, New(val))
	}
	return d
}
func (v FlagSet) Slice() []Native {
	var d = []Native{}
	for _, val := range v {
		d = append(d, New(val))
	}
	return d
}

func (v NilVec) Empty() bool {
	if len(v) == 0 {
		return true
	}
	return false
}

func (v BoolVec) Empty() bool {
	if len(v) == 0 {
		return true
	}
	return false
}

func (v IntVec) Empty() bool {
	if len(v) == 0 {
		return true
	}
	return false
}

func (v Int8Vec) Empty() bool {
	if len(v) == 0 {
		return true
	}
	return false
}

func (v Int16Vec) Empty() bool {
	if len(v) == 0 {
		return true
	}
	return false
}

func (v Int32Vec) Empty() bool {
	if len(v) == 0 {
		return true
	}
	return false
}

func (v UintVec) Empty() bool {
	if len(v) == 0 {
		return true
	}
	return false
}

func (v Uint8Vec) Empty() bool {
	if len(v) == 0 {
		return true
	}
	return false
}

func (v Uint16Vec) Empty() bool {
	if len(v) == 0 {
		return true
	}
	return false
}

func (v Uint32Vec) Empty() bool {
	if len(v) == 0 {
		return true
	}
	return false
}

func (v FltVec) Empty() bool {
	if len(v) == 0 {
		return true
	}
	return false
}

func (v Flt32Vec) Empty() bool {
	if len(v) == 0 {
		return true
	}
	return false
}

func (v ImagVec) Empty() bool {
	if len(v) == 0 {
		return true
	}
	return false
}

func (v Imag64Vec) Empty() bool {
	if len(v) == 0 {
		return true
	}
	return false
}

func (v ByteVec) Empty() bool {
	if len(v) == 0 {
		return true
	}
	return false
}

func (v RuneVec) Empty() bool {
	if len(v) == 0 {
		return true
	}
	return false
}

func (v BytesVec) Empty() bool {
	if len(v) == 0 {
		return true
	}
	return false
}

func (v StrVec) Empty() bool {
	if len(v) == 0 {
		return true
	}
	return false
}

func (v BigIntVec) Empty() bool {
	if len(v) == 0 {
		return true
	}
	return false
}

func (v BigFltVec) Empty() bool {
	if len(v) == 0 {
		return true
	}
	return false
}

func (v RatioVec) Empty() bool {
	if len(v) == 0 {
		return true
	}
	return false
}

func (v TimeVec) Empty() bool {
	if len(v) == 0 {
		return true
	}
	return false
}

func (v DuraVec) Empty() bool {
	if len(v) == 0 {
		return true
	}
	return false
}

func (v ErrorVec) Empty() bool {
	if len(v) == 0 {
		return true
	}
	return false
}

func (v FlagSet) Empty() bool {
	if len(v) == 0 {
		return true
	}
	return false
}
