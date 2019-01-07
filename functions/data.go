package functions

import (
	dat "github.com/JoergReinhardt/godeep/data"
	"math/big"
	"time"
)

type (
	////// TYPED NULL DATA CONSTRUCTORS ///////
	Nil      func() dat.NilVal
	Bool     func() dat.BoolVal
	Int      func() dat.IntVal
	Int8     func() dat.Int8Val
	Int16    func() dat.Int16Val
	Int32    func() dat.Int32Val
	Uint     func() dat.UintVal
	Uint8    func() dat.Uint8Val
	Uint16   func() dat.Uint16Val
	Uint32   func() dat.Uint32Val
	Float    func() dat.FltVal
	Flt32    func() dat.Flt32Val
	Imag     func() dat.ImagVal
	Imag64   func() dat.Imag64Val
	Byte     func() dat.ByteVal
	Rune     func() dat.RuneVal
	Bytes    func() dat.BytesVal
	String   func() dat.StrVal
	BigInt   func() dat.BigIntVal
	BigFlt   func() dat.BigFltVal
	Ratio    func() dat.RatioVal
	Time     func() dat.TimeVal
	Duration func() dat.DuraVal
	Error    func() dat.ErrorVal
)

func conData(data dat.Data) DataValue {
	var dd DataValue
	switch dat.Type(data.Flag()) {
	case dat.Nil:
		dd = Nil(func() dat.NilVal { return data.(dat.NilVal) })
	case dat.Bool:
		dd = Bool(func() dat.BoolVal { return data.(dat.BoolVal) })
	case dat.Int:
		dd = Int(func() dat.IntVal { return data.(dat.IntVal) })
	case dat.Int8:
		dd = Int8(func() dat.Int8Val { return data.(dat.Int8Val) })
	case dat.Int16:
		dd = Int16(func() dat.Int16Val { return data.(dat.Int16Val) })
	case dat.Int32:
		dd = Int32(func() dat.Int32Val { return data.(dat.Int32Val) })
	case dat.Uint:
		dd = Uint(func() dat.UintVal { return data.(dat.UintVal) })
	case dat.Uint8:
		dd = Uint8(func() dat.Uint8Val { return data.(dat.Uint8Val) })
	case dat.Uint16:
		dd = Uint16(func() dat.Uint16Val { return data.(dat.Uint16Val) })
	case dat.Uint32:
		dd = Uint32(func() dat.Uint32Val { return data.(dat.Uint32Val) })
	case dat.Float:
		dd = Float(func() dat.FltVal { return data.(dat.FltVal) })
	case dat.Flt32:
		dd = Flt32(func() dat.Flt32Val { return data.(dat.Flt32Val) })
	case dat.Imag:
		dd = Imag(func() dat.ImagVal { return data.(dat.ImagVal) })
	case dat.Imag64:
		dd = Imag64(func() dat.Imag64Val { return data.(dat.Imag64Val) })
	case dat.Byte:
		dd = Byte(func() dat.ByteVal { return data.(dat.ByteVal) })
	case dat.Rune:
		dd = Rune(func() dat.RuneVal { return data.(dat.RuneVal) })
	case dat.Bytes:
		dd = Bytes(func() dat.BytesVal { return data.(dat.BytesVal) })
	case dat.String:
		dd = String(func() dat.StrVal { return data.(dat.StrVal) })
	case dat.BigInt:
		dd = BigInt(func() dat.BigIntVal { return data.(dat.BigIntVal) })
	case dat.BigFlt:
		dd = BigFlt(func() dat.BigFltVal { return data.(dat.BigFltVal) })
	case dat.Ratio:
		dd = Ratio(func() dat.RatioVal { return data.(dat.RatioVal) })
	case dat.Time:
		dd = Time(func() dat.TimeVal { return data.(dat.TimeVal) })
	case dat.Duration:
		dd = Duration(func() dat.DuraVal { return data.(dat.DuraVal) })
	case dat.Error:
		dd = Error(func() dat.ErrorVal { return data.(dat.ErrorVal) })
	}
	return dd
}

///
func (Nil) Flag() dat.BitFlag        { return dat.Nil.Flag() }
func (v Bool) Flag() dat.BitFlag     { return v.Type().Flag() }
func (v Int) Flag() dat.BitFlag      { return v.Type().Flag() }
func (v Int8) Flag() dat.BitFlag     { return v.Type().Flag() }
func (v Int16) Flag() dat.BitFlag    { return v.Type().Flag() }
func (v Int32) Flag() dat.BitFlag    { return v.Type().Flag() }
func (v Uint) Flag() dat.BitFlag     { return v.Type().Flag() }
func (v Uint8) Flag() dat.BitFlag    { return v.Type().Flag() }
func (v Uint16) Flag() dat.BitFlag   { return v.Type().Flag() }
func (v Uint32) Flag() dat.BitFlag   { return v.Type().Flag() }
func (v BigInt) Flag() dat.BitFlag   { return v.Type().Flag() }
func (v Float) Flag() dat.BitFlag    { return v.Type().Flag() }
func (v Flt32) Flag() dat.BitFlag    { return v.Type().Flag() }
func (v BigFlt) Flag() dat.BitFlag   { return v.Type().Flag() }
func (v Imag) Flag() dat.BitFlag     { return v.Type().Flag() }
func (v Imag64) Flag() dat.BitFlag   { return v.Type().Flag() }
func (v Ratio) Flag() dat.BitFlag    { return v.Type().Flag() }
func (v Rune) Flag() dat.BitFlag     { return v.Type().Flag() }
func (v Byte) Flag() dat.BitFlag     { return v.Type().Flag() }
func (v Bytes) Flag() dat.BitFlag    { return v.Type().Flag() }
func (v String) Flag() dat.BitFlag   { return v.Type().Flag() }
func (v Time) Flag() dat.BitFlag     { return v.Type().Flag() }
func (v Duration) Flag() dat.BitFlag { return v.Type().Flag() }
func (v Error) Flag() dat.BitFlag    { return v.Type().Flag() }

///
func (Nil) Type() Flag        { return conFlag(Constant.Flag(), dat.Nil.Flag()) }
func (v Bool) Type() Flag     { return conFlag(Constant.Flag(), dat.Bool.Flag()) }
func (v Int) Type() Flag      { return conFlag(Constant.Flag(), dat.Int.Flag()) }
func (v Int8) Type() Flag     { return conFlag(Constant.Flag(), dat.Int8.Flag()) }
func (v Int16) Type() Flag    { return conFlag(Constant.Flag(), dat.Int16.Flag()) }
func (v Int32) Type() Flag    { return conFlag(Constant.Flag(), dat.Int32.Flag()) }
func (v Uint) Type() Flag     { return conFlag(Constant.Flag(), dat.Uint.Flag()) }
func (v Uint8) Type() Flag    { return conFlag(Constant.Flag(), dat.Uint8.Flag()) }
func (v Uint16) Type() Flag   { return conFlag(Constant.Flag(), dat.Uint16.Flag()) }
func (v Uint32) Type() Flag   { return conFlag(Constant.Flag(), dat.Uint32.Flag()) }
func (v BigInt) Type() Flag   { return conFlag(Constant.Flag(), dat.BigInt.Flag()) }
func (v Float) Type() Flag    { return conFlag(Constant.Flag(), dat.Float.Flag()) }
func (v Flt32) Type() Flag    { return conFlag(Constant.Flag(), dat.Flt32.Flag()) }
func (v BigFlt) Type() Flag   { return conFlag(Constant.Flag(), dat.BigFlt.Flag()) }
func (v Imag) Type() Flag     { return conFlag(Constant.Flag(), dat.Imag.Flag()) }
func (v Imag64) Type() Flag   { return conFlag(Constant.Flag(), dat.Imag64.Flag()) }
func (v Ratio) Type() Flag    { return conFlag(Constant.Flag(), dat.Ratio.Flag()) }
func (v Rune) Type() Flag     { return conFlag(Constant.Flag(), dat.Rune.Flag()) }
func (v Byte) Type() Flag     { return conFlag(Constant.Flag(), dat.Byte.Flag()) }
func (v Bytes) Type() Flag    { return conFlag(Constant.Flag(), dat.Bytes.Flag()) }
func (v String) Type() Flag   { return conFlag(Constant.Flag(), dat.String.Flag()) }
func (v Time) Type() Flag     { return conFlag(Constant.Flag(), dat.Time.Flag()) }
func (v Duration) Type() Flag { return conFlag(Constant.Flag(), dat.Duration.Flag()) }
func (v Error) Type() Flag    { return conFlag(Constant.Flag(), dat.Error.Flag()) }

//// native nullable typed ///////
func (v Nil) Null() struct{}           { return struct{}{} }
func (v Bool) Null() bool              { return false }
func (v Int) Null() int                { return 0 }
func (v Int8) Null() int8              { return 0 }
func (v Int16) Null() int16            { return 0 }
func (v Int32) Null() int32            { return 0 }
func (v Uint) Null() uint              { return 0 }
func (v Uint8) Null() uint8            { return 0 }
func (v Uint16) Null() uint16          { return 0 }
func (v Uint32) Null() uint32          { return 0 }
func (v Float) Null() float64          { return 0 }
func (v Flt32) Null() float32          { return 0 }
func (v Imag) Null() complex128        { return complex128(0.0) }
func (v Imag64) Null() complex64       { return complex64(0.0) }
func (v Byte) Null() byte              { return byte(0) }
func (v Rune) Null() rune              { return rune(' ') }
func (v String) Null() string          { return string("") }
func (v BigInt) Null() *big.Int        { return big.NewInt(0) }
func (v BigFlt) Null() *big.Float      { return big.NewFloat(0) }
func (v Ratio) Null() *big.Rat         { return big.NewRat(1, 1) }
func (v Time) Null() time.Time         { return time.Now() }
func (v Duration) Null() time.Duration { return time.Duration(0) }
