package types

import (
	"math/big"
	"math/bits"
	"strconv"
	"time"
)

type (
	////// TYPED NULL DATA CONSTRUCTORS ///////
	nilFnc    func(d Data) nilVal
	boolFnc   func(d Data) boolVal
	intFnc    func(d Data) intVal
	int8Fnc   func(d Data) int8Val
	int16Fnc  func(d Data) int16Val
	int32Fnc  func(d Data) int32Val
	uintFnc   func(d Data) uintVal
	uint8Fnc  func(d Data) uint8Val
	uint16Fnc func(d Data) uint16Val
	uint32Fnc func(d Data) uint32Val
	fltFnc    func(d Data) fltVal
	flt32Fnc  func(d Data) flt32Val
	imagFnc   func(d Data) imagVal
	imag64Fnc func(d Data) imag64Val
	byteFnc   func(d Data) byteVal
	runeFnc   func(d Data) runeVal
	bytesFnc  func(d Data) bytesVal
	strFnc    func(d Data) strVal
	bigIntFnc func(d Data) bigIntVal
	bigFltFnc func(d Data) bigFltVal
	ratioFnc  func(d Data) ratioVal
	timeFnc   func(d Data) timeVal
	duraFnc   func(d Data) duraVal
	errorFnc  func(d Data) errorVal
	///// TYPED NULL DATA CONSTRUCTORS ENUMERATED /////
	ColnilFnc    func(d Data) []nilVal
	ColboolFnc   func(d Data) []boolVal
	ColintFnc    func(d Data) []intVal
	Colint8Fnc   func(d Data) []int8Val
	Colint16Fnc  func(d Data) []int16Val
	Colint32Fnc  func(d Data) []int32Val
	ColuintFnc   func(d Data) []uintVal
	Coluint8Fnc  func(d Data) []uint8Val
	Coluint16Fnc func(d Data) []uint16Val
	Coluint32Fnc func(d Data) []uint32Val
	ColfltFnc    func(d Data) []fltVal
	Colflt32Fnc  func(d Data) []flt32Val
	ColimagFnc   func(d Data) []imagVal
	Colimag64Fnc func(d Data) []imag64Val
	ColbyteFnc   func(d Data) []byteVal
	ColruneFnc   func(d Data) []runeVal
	ColbytesFnc  func(d Data) []bytesVal
	ColstrFnc    func(d Data) []strVal
	ColbigIntFnc func(d Data) []bigIntVal
	ColbigFltFnc func(d Data) []bigFltVal
	ColratioFnc  func(d Data) []ratioVal
	ColtimeFnc   func(d Data) []timeVal
	ColduraFnc   func(d Data) []duraVal
	ColerrorFnc  func(d Data) []errorVal
)

///// ALIASES FOR ENUMERATED NATIVES  /////
func (v nilVal) NativeEnumNull() []struct{}       { return []struct{}{} }
func (v boolVal) NativeEnumNull() []bool          { return []bool{} }
func (v intVal) NativeEnumNull() []int            { return []int{} }
func (v int8Val) NativeEnumNull() []int8          { return []int8{} }
func (v int16Val) NativeEnumNull() []int16        { return []int16{} }
func (v int32Val) NativeEnumNull() []int32        { return []int32{} }
func (v uintVal) NativeEnumNull() []uint          { return []uint{} }
func (v uint8Val) NativeEnumNull() []uint8        { return []uint8{} }
func (v uint16Val) NativeEnumNull() []uint16      { return []uint16{} }
func (v uint32Val) NativeEnumNull() []uint32      { return []uint32{} }
func (v fltVal) NativeEnumNull() []float64        { return []float64{} }
func (v flt32Val) NativeEnumNull() []float32      { return []float32{} }
func (v imagVal) NativeEnumNull() []complex128    { return []complex128{} }
func (v imag64Val) NativeEnumNull() []complex64   { return []complex64{} }
func (v byteVal) NativeEnumNull() []byte          { return []byte{} }
func (v runeVal) NativeEnumNull() []rune          { return []rune{} }
func (v strVal) NativeEnumNull() []string         { return []string{} }
func (v bigIntVal) NativeEnumNull() []big.Int     { return []big.Int{} }
func (v bigFltVal) NativeEnumNull() []big.Float   { return []big.Float{} }
func (v ratioVal) NativeEnumNull() []big.Rat      { return []big.Rat{} }
func (v timeVal) NativeEnumNull() []time.Time     { return []time.Time{} }
func (v duraVal) NativeEnumNull() []time.Duration { return []time.Duration{} }

//// native nullable typed ///////
func (v nilVal) NativeNull() struct{}       { return struct{}{} }
func (v boolVal) NativeNull() bool          { return false }
func (v intVal) NativeNull() int            { return 0 }
func (v int8Val) NativeNull() int8          { return 0 }
func (v int16Val) NativeNull() int16        { return 0 }
func (v int32Val) NativeNull() int32        { return 0 }
func (v uintVal) NativeNull() uint          { return 0 }
func (v uint8Val) NativeNull() uint8        { return 0 }
func (v uint16Val) NativeNull() uint16      { return 0 }
func (v uint32Val) NativeNull() uint32      { return 0 }
func (v fltVal) NativeNull() float64        { return 0 }
func (v flt32Val) NativeNull() float32      { return 0 }
func (v imagVal) NativeNull() complex128    { return complex128(0.0) }
func (v imag64Val) NativeNull() complex64   { return complex64(0.0) }
func (v byteVal) NativeNull() byte          { return byte(0) }
func (v runeVal) NativeNull() rune          { return rune(' ') }
func (v strVal) NativeNull() string         { return string("") }
func (v bigIntVal) NativeNull() *big.Int    { return big.NewInt(0) }
func (v bigFltVal) NativeNull() *big.Float  { return big.NewFloat(0) }
func (v ratioVal) NativeNull() *big.Rat     { return big.NewRat(1, 1) }
func (v timeVal) NativeNull() time.Time     { return time.Now() }
func (v duraVal) NativeNull() time.Duration { return time.Duration(0) }

///// TYPE CONVERSION //////
// BOOL -> VALUE
func (v boolVal) Int() intVal {
	if v {
		return intVal(1)
	}
	return intVal(-1)
}
func (v boolVal) IntNat() int {
	if v {
		return 1
	}
	return -1
}
func (v boolVal) UintNat() uint {
	if v {
		return 1
	}
	return 0
}

// VALUE -> BOOL
func (v intVal) Bool() boolVal {
	if v == 1 {
		return boolVal(true)
	}
	return boolVal(false)
}
func (v strVal) Bool() boolVal {
	s, err := strconv.ParseBool(string(v))
	if err != nil {
		return false
	}
	return boolVal(s)
}

// INT -> VALUE
func (v intVal) Integer() int { return int(v) } // implements Idx Attribut
func (v intVal) Idx() intVal  { return v }      // implements Idx Attribut
//func (v intVal) Key() strVal    { return v.String() } // implements Key Attribut
func (v intVal) FltNat() fltVal { return fltVal(v) }
func (v intVal) IntNat() intVal { return v }
func (v intVal) UintNat() uintVal {
	if v < 0 {
		return uintVal(v * -1)
	}
	return uintVal(v)
}

// VALUE -> INT
func (v int8Val) Int() intVal   { return intVal(int(v)) }
func (v int16Val) Int() intVal  { return intVal(int(v)) }
func (v int32Val) Int() intVal  { return intVal(int(v)) }
func (v uintVal) Int() intVal   { return intVal(int(v)) }
func (v uint16Val) Int() intVal { return intVal(int(v)) }
func (v uint32Val) Int() intVal { return intVal(int(v)) }
func (v fltVal) Int() intVal    { return intVal(int(v)) }
func (v flt32Val) Int() intVal  { return intVal(int(v)) }
func (v byteVal) Int() intVal   { return intVal(int(v)) }
func (v imagVal) Real() intVal  { return intVal(real(v)) }
func (v imagVal) Imag() intVal  { return intVal(imag(v)) }
func (v strVal) Int() intVal {
	s, err := strconv.Atoi(string(v))
	if err != nil {
		return -1
	}
	return intVal(s)
}

// VALUE -> FLOAT
func (v uintVal) Float() fltVal { return fltVal(v.Int().Float()) }
func (v intVal) Float() fltVal  { return fltVal(v.FltNat()) }
func (v strVal) Float() fltVal {
	s, err := strconv.ParseFloat(string(v), 64)
	if err != nil {
		return -1
	}
	return fltVal(s)
}

// VALUE -> UINT
func (v uintVal) Uint() uintVal { return v }
func (v uintVal) UintNat() uint { return uint(v) }
func (v intVal) Uint() uintVal  { return uintVal(v.UintNat()) }
func (v strVal) Uint() uintVal {
	u, err := strconv.ParseUint(string(v), 10, 64)
	if err != nil {
		return 0
	}
	return uintVal(u)
}
func (v boolVal) Uint() uintVal {
	if v {
		return uintVal(1)
	}
	return uintVal(0)
}

// INTEGERS FOR DEDICATED PURPOSES
func (v uintVal) Len() intVal  { return intVal(bits.Len64(uint64(v))) }
func (v byteVal) Len() intVal  { return intVal(bits.Len64(uint64(v))) }
func (v bytesVal) Len() intVal { return intVal(len(v)) }
func (v strVal) Len() intVal   { return intVal(len(string(v))) }

// SLICE ->
func (v chain) Slice() []Data { return v }
func (v chain) Len() int      { return len(v) }
