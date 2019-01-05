package functions

import (
	dat "github.com/JoergReinhardt/godeep/data"
	"math/big"
	"time"
)

type (
	////// TYPED NULL DATA CONSTRUCTORS ///////
	Nil    func(d dat.Data) dat.NilVal
	Bool   func(d dat.Data) dat.BoolVal
	Int    func(d dat.Data) dat.IntVal
	Int8   func(d dat.Data) dat.Int8Val
	Int16  func(d dat.Data) dat.Int16Val
	Int32  func(d dat.Data) dat.Int32Val
	Uint   func(d dat.Data) dat.UintVal
	Uint8  func(d dat.Data) dat.Uint8Val
	Uint16 func(d dat.Data) dat.Uint16Val
	Uint32 func(d dat.Data) dat.Uint32Val
	Flt    func(d dat.Data) dat.FltVal
	Flt32  func(d dat.Data) dat.Flt32Val
	Imag   func(d dat.Data) dat.ImagVal
	Imag64 func(d dat.Data) dat.Imag64Val
	Byte   func(d dat.Data) dat.ByteVal
	Rune   func(d dat.Data) dat.RuneVal
	Bytes  func(d dat.Data) dat.BytesVal
	Str    func(d dat.Data) dat.StrVal
	BigInt func(d dat.Data) dat.BigIntVal
	BigFlt func(d dat.Data) dat.BigFltVal
	Ratio  func(d dat.Data) dat.RatioVal
	Time   func(d dat.Data) dat.TimeVal
	Dura   func(d dat.Data) dat.DuraVal
	Error  func(d dat.Data) dat.ErrorVal
)

///
func (Nil) Flag() Flag      { return ComposeFlag(Flag(dat.Nil), Flag(dat.Function.Flag())) }
func (v Bool) Flag() Flag   { return ComposeFlag(Flag(dat.Bool), Flag(dat.Function.Flag())) }
func (v Int) Flag() Flag    { return ComposeFlag(Flag(dat.Int), Flag(dat.Function.Flag())) }
func (v Int8) Flag() Flag   { return ComposeFlag(Flag(dat.Int8), Flag(dat.Function.Flag())) }
func (v Int16) Flag() Flag  { return ComposeFlag(Flag(dat.Int16), Flag(dat.Function.Flag())) }
func (v Int32) Flag() Flag  { return ComposeFlag(Flag(dat.Int32), Flag(dat.Function.Flag())) }
func (v Uint) Flag() Flag   { return ComposeFlag(Flag(dat.Uint), Flag(dat.Function.Flag())) }
func (v Uint8) Flag() Flag  { return ComposeFlag(Flag(dat.Uint8), Flag(dat.Function.Flag())) }
func (v Uint16) Flag() Flag { return ComposeFlag(Flag(dat.Uint16), Flag(dat.Function.Flag())) }
func (v Uint32) Flag() Flag { return ComposeFlag(Flag(dat.Uint32), Flag(dat.Function.Flag())) }
func (v BigInt) Flag() Flag { return ComposeFlag(Flag(dat.BigInt), Flag(dat.Function.Flag())) }
func (v Flt) Flag() Flag    { return ComposeFlag(Flag(dat.Float), Flag(dat.Function.Flag())) }
func (v Flt32) Flag() Flag  { return ComposeFlag(Flag(dat.Flt32), Flag(dat.Function.Flag())) }
func (v BigFlt) Flag() Flag { return ComposeFlag(Flag(dat.BigFlt), Flag(dat.Function.Flag())) }
func (v Imag) Flag() Flag   { return ComposeFlag(Flag(dat.Imag), Flag(dat.Function.Flag())) }
func (v Imag64) Flag() Flag { return ComposeFlag(Flag(dat.Imag64), Flag(dat.Function.Flag())) }
func (v Ratio) Flag() Flag  { return ComposeFlag(Flag(dat.Ratio), Flag(dat.Function.Flag())) }
func (v Rune) Flag() Flag   { return ComposeFlag(Flag(dat.Rune), Flag(dat.Function.Flag())) }
func (v Byte) Flag() Flag   { return ComposeFlag(Flag(dat.Byte), Flag(dat.Function.Flag())) }
func (v Bytes) Flag() Flag  { return ComposeFlag(Flag(dat.Bytes), Flag(dat.Function.Flag())) }
func (v Str) Flag() Flag    { return ComposeFlag(Flag(dat.String), Flag(dat.Function.Flag())) }
func (v Time) Flag() Flag   { return ComposeFlag(Flag(dat.Time), Flag(dat.Function.Flag())) }
func (v Dura) Flag() Flag   { return ComposeFlag(Flag(dat.Duration), Flag(dat.Function.Flag())) }
func (v Error) Flag() Flag  { return ComposeFlag(Flag(dat.Error), Flag(dat.Function.Flag())) }

//// native nullable typed ///////
func (v Nil) Null() struct{}       { return struct{}{} }
func (v Bool) Null() bool          { return false }
func (v Int) Null() int            { return 0 }
func (v Int8) Null() int8          { return 0 }
func (v Int16) Null() int16        { return 0 }
func (v Int32) Null() int32        { return 0 }
func (v Uint) Null() uint          { return 0 }
func (v Uint8) Null() uint8        { return 0 }
func (v Uint16) Null() uint16      { return 0 }
func (v Uint32) Null() uint32      { return 0 }
func (v Flt) Null() float64        { return 0 }
func (v Flt32) Null() float32      { return 0 }
func (v Imag) Null() complex128    { return complex128(0.0) }
func (v Imag64) Null() complex64   { return complex64(0.0) }
func (v Byte) Null() byte          { return byte(0) }
func (v Rune) Null() rune          { return rune(' ') }
func (v Str) Null() string         { return string("") }
func (v BigInt) Null() *big.Int    { return big.NewInt(0) }
func (v BigFlt) Null() *big.Float  { return big.NewFloat(0) }
func (v Ratio) Null() *big.Rat     { return big.NewRat(1, 1) }
func (v Time) Null() time.Time     { return time.Now() }
func (v Dura) Null() time.Duration { return time.Duration(0) }
