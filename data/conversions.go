package data

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math/big"
	"math/bits"
	"strings"
	"time"
)

//
// 0  ← Nil
// 1  ← Bool
// 2  ← Int8
// 3  ← Int16
// 4  ← Int32
// 4  ← Int
// 5  ← BigInt
// 6  ← Uint8
// 7  ← Uint16
// 9  ← Uint32
// 10 ← Uint
// 11 ← Flt32
// 12 ← Float
// 13 ← BigFlt
// 14 ← Ratio
// 15 ← Imag64
// 16 ← Imag
// 17 ← Time
// 18 ← Duration
// 19 ← Byte
// 20 ← Rune
// 21 ← Flag
// 22 ← String
// 23 ← Bytes
// 24 ← Error
var TypeConversionTable = [][]func(arg Native) Native{

	[]func(arg Native) Native{ // 0  ← Nil
		func(Native) Native { return NilVal{} },                  // 0  ← Nil
		func(Native) Native { return BoolVal(false) },            // 1  ← Bool
		func(Native) Native { return Int8Val(0) },                // 2  ← Int8
		func(Native) Native { return Int16Val(0) },               // 3  ← Int16
		func(Native) Native { return Int32Val(0) },               // 4  ← Int32
		func(Native) Native { return IntVal(0) },                 // 4  ← Int
		func(Native) Native { return IntVal(0).BigInt() },        // 5  ← BigInt
		func(Native) Native { return Uint8Val(0) },               // 6  ← Uint8
		func(Native) Native { return Uint16Val(0) },              // 7  ← Uint16
		func(Native) Native { return Uint32Val(0) },              // 9  ← Uint32
		func(Native) Native { return UintVal(0) },                // 10 ← Uint
		func(Native) Native { return Flt32Val(0.0) },             // 12 ← Flt32
		func(Native) Native { return FltVal(0.0) },               // 13 ← Float
		func(Native) Native { return FltVal(0.0).BigFlt() },      // 14 ← BigFlt
		func(Native) Native { return IntVal(0).Ratio() },         // 15 ← Ratio
		func(Native) Native { return IntVal(0).Imag().Imag64() }, // 16 ← Imag64
		func(Native) Native { return IntVal(0).Imag() },          // 17 ← Imag
		func(Native) Native { return TimeVal{} },                 // 18 ← Time
		func(arg Native) Native { return DuraVal(0) },            // 19 ← Duration
		func(arg Native) Native { return ByteVal(0) },            // 20 ← Byte
		func(arg Native) Native { return RuneVal(0) },            // 21 ← Rune
		func(arg Native) Native { return BitFlag(0) },            // 22 ← Flag
		func(arg Native) Native { return StrVal("") },            // 23 ← String
		func(arg Native) Native { return BytesVal{} },            // 24 ← Bytes
		func(arg Native) Native { return ErrorVal{} },            // 25 ← Error
	},

	[]func(arg Native) Native{ // 1  ← Bool
		func(arg Native) Native { return NilVal{} },                             // 0  ← Nil
		func(arg Native) Native { return arg },                                  // 1  ← Bool
		func(arg Native) Native { return arg.(BoolVal).Int().Int8() },           // 2  ← Int8
		func(arg Native) Native { return arg.(BoolVal).Int().Int16() },          // 3  ← Int16
		func(arg Native) Native { return arg.(BoolVal).Int().Int32() },          // 4  ← Int32
		func(arg Native) Native { return arg.(BoolVal).Int() },                  // 4  ← Int
		func(arg Native) Native { return arg.(BoolVal).Int().BigInt() },         // 5  ← BigInt
		func(arg Native) Native { return arg.(BoolVal).Uint().Uint8() },         // 6  ← Uint8
		func(arg Native) Native { return arg.(BoolVal).Uint().Uint16() },        // 7  ← Uint16
		func(arg Native) Native { return arg.(BoolVal).Uint().Uint32() },        // 9  ← Uint32
		func(arg Native) Native { return arg.(BoolVal).Uint() },                 // 10 ← Uint
		func(arg Native) Native { return arg.(BoolVal).Int().Float().Flt32() },  // 12 ← Flt32
		func(arg Native) Native { return arg.(BoolVal).Int().Float() },          // 13 ← Float
		func(arg Native) Native { return arg.(BoolVal).Int().Float().BigFlt() }, // 14 ← BigFlt
		func(arg Native) Native { return arg.(BoolVal).Int().Ratio() },          // 15 ← Ratio
		func(arg Native) Native { return arg.(BoolVal).Int().Imag().Imag64() },  // 16 ← Imag64
		func(arg Native) Native { return arg.(BoolVal).Int().Imag() },           // 17 ← Imag
		func(arg Native) Native { // 18 ← Time
			if arg.(BoolVal).Bool() {
				return TimeVal(time.Now())
			}
			return TimeVal{}
		},
		func(arg Native) Native { // 19 ← Duration
			if arg.(BoolVal).Bool() {
				return DuraVal(1)
			}
			return DuraVal(0)
		}, // 18 ← Time
		func(arg Native) Native { return arg.(BoolVal).Int().Byte() },             // 20 ← Byte
		func(arg Native) Native { return arg.(BoolVal).Int().Rune() },             // 21 ← Rune
		func(arg Native) Native { return arg.(BoolVal).Int().Flag() },             // 22 ← Flag
		func(arg Native) Native { return StrVal(arg.(BoolVal).String()) },         // 23 ← String
		func(arg Native) Native { return StrVal(arg.(BoolVal).String()).Bytes() }, // 24 ← Bytes
		func(arg Native) Native {
			if arg.(BoolVal).Bool() {
				return NewError(
					fmt.Errorf(
						"error occured during convertion from bool"))
			}
			return ErrorVal{}
		}, // 25 ← Error
	},

	[]func(arg Native) Native{ // 2  ← Int8
		func(arg Native) Native { return NilVal{} }, // 0  ← Nil
		func(arg Native) Native {
			if arg.(Int8Val) > Int8Val(0) {
				return BoolVal(true)
			}
			return BoolVal(false)
		}, // 1  ← Bool
		func(arg Native) Native { return arg },                                   // 2  ← Int8
		func(arg Native) Native { return Int16Val(arg.(Int8Val)) },               // 3  ← Int16
		func(arg Native) Native { return Int32Val(arg.(Int8Val)) },               // 4  ← Int32
		func(arg Native) Native { return IntVal(arg.(Int8Val)) },                 // 4  ← Int
		func(arg Native) Native { return IntVal(arg.(Int8Val)).BigInt() },        // 5  ← BigInt
		func(arg Native) Native { return UintVal(arg.(Int8Val)) },                // 6  ← Uint8
		func(arg Native) Native { return Uint16Val(arg.(Int8Val)) },              // 7  ← Uint16
		func(arg Native) Native { return Uint32Val(arg.(Int8Val)) },              // 9  ← Uint32
		func(arg Native) Native { return UintVal(arg.(Int8Val)) },                // 10 ← Uint
		func(arg Native) Native { return Flt32Val(arg.(Uint8Val)) },              // 12 ← Flt32
		func(arg Native) Native { return FltVal(arg.(Int8Val)) },                 // 13 ← Float
		func(arg Native) Native { return FltVal(arg.(Int8Val)).BigFlt() },        // 14 ← BigFlt
		func(arg Native) Native { return IntVal(arg.(Int8Val)).Ratio() },         // 15 ← Ratio
		func(arg Native) Native { return IntVal(arg.(Int8Val)).Imag().Imag64() }, // 16 ← Imag64
		func(arg Native) Native { return IntVal(arg.(Int8Val)).Imag() },          // 17 ← Imag
		func(arg Native) Native {
			if arg.(Int8Val) > 0 {
				return TimeVal(time.Now())
			}
			return TimeVal{}
		}, // 18 ← Time
		func(arg Native) Native { return DuraVal(arg.(Int8Val)) },                 // 19 ← Duration
		func(arg Native) Native { return IntVal(arg.(Int8Val)).Byte() },           // 20 ← Byte
		func(arg Native) Native { return IntVal(arg.(Int8Val)).Rune() },           // 21 ← Rune
		func(arg Native) Native { return IntVal(arg.(Int8Val)).Flag() },           // 22 ← Flag
		func(arg Native) Native { return StrVal(arg.(Int8Val).String()) },         // 23 ← String
		func(arg Native) Native { return StrVal(arg.(Int8Val).String()).Bytes() }, // 24 ← Bytes
		func(arg Native) Native {
			if arg.(BoolVal).Bool() {
				return NewError(
					fmt.Errorf(
						"error occured during convertion from int8"))
			}
			return ErrorVal{}
		}, // 25 ← Error
	},

	[]func(arg Native) Native{ // 3  ← Int16
		func(arg Native) Native { return NilVal{} }, // 0  ← Nil
		func(arg Native) Native {
			if arg.(Int16Val) > Int16Val(0) {
				return BoolVal(true)
			}
			return BoolVal(false)
		}, // 1  ← Bool
		func(arg Native) Native { return arg },                                    // 2  ← Int16
		func(arg Native) Native { return Int16Val(arg.(Int16Val)) },               // 3  ← Int16
		func(arg Native) Native { return Int32Val(arg.(Int16Val)) },               // 4  ← Int32
		func(arg Native) Native { return IntVal(arg.(Int16Val)) },                 // 4  ← Int
		func(arg Native) Native { return IntVal(arg.(Int16Val)).BigInt() },        // 5  ← BigInt
		func(arg Native) Native { return UintVal(arg.(Int16Val)) },                // 6  ← Uint8
		func(arg Native) Native { return Uint16Val(arg.(Int16Val)) },              // 7  ← Uint16
		func(arg Native) Native { return Uint32Val(arg.(Int16Val)) },              // 9  ← Uint32
		func(arg Native) Native { return UintVal(arg.(Int16Val)) },                // 10 ← Uint
		func(arg Native) Native { return Flt32Val(arg.(Uint16Val)) },              // 12 ← Flt32
		func(arg Native) Native { return FltVal(arg.(Int16Val)) },                 // 13 ← Float
		func(arg Native) Native { return FltVal(arg.(Int16Val)).BigFlt() },        // 14 ← BigFlt
		func(arg Native) Native { return IntVal(arg.(Int16Val)).Ratio() },         // 15 ← Ratio
		func(arg Native) Native { return IntVal(arg.(Int16Val)).Imag().Imag64() }, // 16 ← Imag64
		func(arg Native) Native { return IntVal(arg.(Int16Val)).Imag() },          // 17 ← Imag
		func(arg Native) Native {
			if arg.(Int16Val) > 0 {
				return TimeVal(time.Now())
			}
			return TimeVal{}
		}, // 116 ← Time
		func(arg Native) Native { return DuraVal(arg.(Int16Val)) },                 // 19 ← Duration
		func(arg Native) Native { return IntVal(arg.(Int16Val)).Byte() },           // 20 ← Byte
		func(arg Native) Native { return IntVal(arg.(Int16Val)).Rune() },           // 21 ← Rune
		func(arg Native) Native { return IntVal(arg.(Int16Val)).Flag() },           // 22 ← Flag
		func(arg Native) Native { return StrVal(arg.(Int16Val).String()) },         // 23 ← String
		func(arg Native) Native { return StrVal(arg.(Int16Val).String()).Bytes() }, // 24 ← Bytes
		func(arg Native) Native {
			if arg.(BoolVal).Bool() {
				return NewError(
					fmt.Errorf(
						"error occured during convertion from int16"))
			}
			return ErrorVal{}
		}, // 25 ← Error
	},

	[]func(arg Native) Native{ // 4  ← Int32
		func(arg Native) Native { return NilVal{} }, // 0  ← Nil
		func(arg Native) Native {
			if arg.(Int32Val) > Int32Val(0) {
				return BoolVal(true)
			}
			return BoolVal(false)
		}, // 1  ← Bool
		func(arg Native) Native { return arg },                                    // 2  ← Int8
		func(arg Native) Native { return Int16Val(arg.(Int32Val)) },               // 3  ← Int16
		func(arg Native) Native { return Int32Val(arg.(Int32Val)) },               // 4  ← Int32
		func(arg Native) Native { return IntVal(arg.(Int32Val)) },                 // 4  ← Int
		func(arg Native) Native { return IntVal(arg.(Int32Val)).BigInt() },        // 5  ← BigInt
		func(arg Native) Native { return Uint8Val(arg.(Int32Val)) },               // 6  ← Uint8
		func(arg Native) Native { return Uint16Val(arg.(Int32Val)) },              // 7  ← Uint16
		func(arg Native) Native { return Uint32Val(arg.(Int32Val)) },              // 9  ← Uint32
		func(arg Native) Native { return UintVal(arg.(Int32Val)) },                // 10 ← Uint
		func(arg Native) Native { return Flt32Val(arg.(Uint32Val)) },              // 12 ← Flt32
		func(arg Native) Native { return FltVal(arg.(Int32Val)) },                 // 13 ← Float
		func(arg Native) Native { return FltVal(arg.(Int32Val)).BigFlt() },        // 14 ← BigFlt
		func(arg Native) Native { return IntVal(arg.(Int32Val)).Ratio() },         // 15 ← Ratio
		func(arg Native) Native { return IntVal(arg.(Int32Val)).Imag().Imag64() }, // 16 ← Imag64
		func(arg Native) Native { return IntVal(arg.(Int32Val)).Imag() },          // 17 ← Imag
		func(arg Native) Native {
			if arg.(Int32Val) > 0 {
				return TimeVal(time.Now())
			}
			return TimeVal{}
		}, // 132 ← Time
		func(arg Native) Native { return DuraVal(arg.(Int32Val)) },                 // 19 ← Duration
		func(arg Native) Native { return IntVal(arg.(Int32Val)).Byte() },           // 20 ← Byte
		func(arg Native) Native { return IntVal(arg.(Int32Val)).Rune() },           // 21 ← Rune
		func(arg Native) Native { return IntVal(arg.(Int32Val)).Flag() },           // 22 ← Flag
		func(arg Native) Native { return StrVal(arg.(Int32Val).String()) },         // 23 ← String
		func(arg Native) Native { return StrVal(arg.(Int32Val).String()).Bytes() }, // 24 ← Bytes
		func(arg Native) Native {
			if arg.(BoolVal).Bool() {
				return NewError(
					fmt.Errorf(
						"error occured during convertion from int32"))
			}
			return ErrorVal{}
		}, // 25 ← Error
	},

	[]func(arg Native) Native{ // 4  ← Int
		func(arg Native) Native { return NilVal{} }, // 0  ← Nil
		func(arg Native) Native {
			if arg.(IntVal) > IntVal(0) {
				return BoolVal(true)
			}
			return BoolVal(false)
		}, // 1  ← Bool
		func(arg Native) Native { return arg },                                  // 2  ← Int8
		func(arg Native) Native { return Int16Val(arg.(IntVal)) },               // 3  ← Int16
		func(arg Native) Native { return Int32Val(arg.(IntVal)) },               // 4  ← Int32
		func(arg Native) Native { return IntVal(arg.(IntVal)) },                 // 4  ← Int
		func(arg Native) Native { return IntVal(arg.(IntVal)).BigInt() },        // 5  ← BigInt
		func(arg Native) Native { return Uint8Val(arg.(IntVal)) },               // 6  ← Uint8
		func(arg Native) Native { return Uint16Val(arg.(IntVal)) },              // 7  ← Uint16
		func(arg Native) Native { return Uint32Val(arg.(IntVal)) },              // 9  ← Uint32
		func(arg Native) Native { return UintVal(arg.(IntVal)) },                // 10 ← Uint
		func(arg Native) Native { return FltVal(arg.(IntVal)) },                 // 12 ← Flt32
		func(arg Native) Native { return FltVal(arg.(IntVal)) },                 // 13 ← Float
		func(arg Native) Native { return FltVal(arg.(IntVal)).BigFlt() },        // 14 ← BigFlt
		func(arg Native) Native { return IntVal(arg.(IntVal)).Ratio() },         // 15 ← Ratio
		func(arg Native) Native { return IntVal(arg.(IntVal)).Imag().Imag64() }, // 16 ← Imag64
		func(arg Native) Native { return IntVal(arg.(IntVal)).Imag() },          // 17 ← Imag
		func(arg Native) Native {
			if arg.(IntVal) > 0 {
				return TimeVal(time.Now())
			}
			return TimeVal{}
		}, // 1 ← Time
		func(arg Native) Native { return DuraVal(arg.(IntVal)) },                 // 19 ← Duration
		func(arg Native) Native { return IntVal(arg.(IntVal)).Byte() },           // 20 ← Byte
		func(arg Native) Native { return IntVal(arg.(IntVal)).Rune() },           // 21 ← Rune
		func(arg Native) Native { return IntVal(arg.(IntVal)).Flag() },           // 22 ← Flag
		func(arg Native) Native { return StrVal(arg.(IntVal).String()) },         // 23 ← String
		func(arg Native) Native { return StrVal(arg.(IntVal).String()).Bytes() }, // 24 ← Bytes
		func(arg Native) Native {
			if arg.(BoolVal).Bool() {
				return NewError(
					fmt.Errorf(
						"error occured during convertion from int"))
			}
			return ErrorVal{}
		}, // 25 ← Error
	},

	[]func(arg Native) Native{ // 5  ← BigInt
		func(arg Native) Native { return NilVal{} }, // 0  ← Nil
		func(arg Native) Native {
			if arg.(BigIntVal).Int() > IntVal(0) {
				return BoolVal(true)
			}
			return BoolVal(false)
		}, // 1  ← Bool
		func(arg Native) Native { return arg },                                           // 2  ← Int8
		func(arg Native) Native { return Int16Val(arg.(BigIntVal).Int()) },               // 3  ← Int16
		func(arg Native) Native { return Int32Val(arg.(BigIntVal).Int()) },               // 4  ← Int32
		func(arg Native) Native { return IntVal(arg.(BigIntVal).Int()) },                 // 4  ← Int
		func(arg Native) Native { return IntVal(arg.(BigIntVal).Int()).BigInt() },        // 5  ← BigInt
		func(arg Native) Native { return Uint8Val(arg.(BigIntVal).Int()) },               // 6  ← Uint8
		func(arg Native) Native { return Uint16Val(arg.(BigIntVal).Int()) },              // 7  ← Uint16
		func(arg Native) Native { return Uint32Val(arg.(BigIntVal).Int()) },              // 9  ← Uint32
		func(arg Native) Native { return UintVal(arg.(BigIntVal).Int()) },                // 10 ← Uint
		func(arg Native) Native { return FltVal(arg.(BigIntVal).Int()) },                 // 12 ← Flt32
		func(arg Native) Native { return FltVal(arg.(BigIntVal).Int()) },                 // 13 ← Float
		func(arg Native) Native { return FltVal(arg.(BigIntVal).Int()).BigFlt() },        // 14 ← BigFlt
		func(arg Native) Native { return IntVal(arg.(BigIntVal).Int()).Ratio() },         // 15 ← Ratio
		func(arg Native) Native { return IntVal(arg.(BigIntVal).Int()).Imag().Imag64() }, // 16 ← Imag64
		func(arg Native) Native { return IntVal(arg.(BigIntVal).Int()).Imag() },          // 17 ← Imag
		func(arg Native) Native {
			if arg.(BigIntVal).Int() > 0 {
				return TimeVal(time.Now())
			}
			return TimeVal{}
		}, // 1 ← Time
		func(arg Native) Native { return DuraVal(arg.(BigIntVal).Int()) },                 // 19 ← Duration
		func(arg Native) Native { return IntVal(arg.(BigIntVal).Int()).Byte() },           // 20 ← Byte
		func(arg Native) Native { return IntVal(arg.(BigIntVal).Int()).Rune() },           // 21 ← Rune
		func(arg Native) Native { return IntVal(arg.(BigIntVal).Int()).Flag() },           // 22 ← Flag
		func(arg Native) Native { return StrVal(arg.(BigIntVal).Int().String()) },         // 23 ← String
		func(arg Native) Native { return StrVal(arg.(BigIntVal).Int().String()).Bytes() }, // 24 ← Bytes
		func(arg Native) Native {
			if arg.(BoolVal).Bool() {
				return NewError(
					fmt.Errorf(
						"error occured during convertion from bigint"))
			}
			return ErrorVal{}
		}, // 25 ← Error
	},

	[]func(arg Native) Native{ // 6  ← Uint8
		func(arg Native) Native { return NilVal{} }, // 0  ← Nil
		func(arg Native) Native {
			if arg.(Uint8Val) > Uint8Val(0) {
				return BoolVal(true)
			}
			return BoolVal(false)
		}, // 1  ← Bool
		func(arg Native) Native { return arg },                                     // 2  ← Int8
		func(arg Native) Native { return Uint16Val(arg.(Uint8Val)) },               // 3  ← Int16
		func(arg Native) Native { return Uint32Val(arg.(Uint8Val)) },               // 4  ← Int32
		func(arg Native) Native { return UintVal(arg.(Uint8Val)) },                 // 4  ← Int
		func(arg Native) Native { return UintVal(arg.(Uint8Val)).BigInt() },        // 5  ← BigInt
		func(arg Native) Native { return UintVal(arg.(Uint8Val)) },                 // 6  ← Uint8
		func(arg Native) Native { return Uint16Val(arg.(Uint8Val)) },               // 7  ← Uint16
		func(arg Native) Native { return Uint32Val(arg.(Uint8Val)) },               // 9  ← Uint32
		func(arg Native) Native { return UintVal(arg.(Uint8Val)) },                 // 10 ← Uint
		func(arg Native) Native { return Flt32Val(arg.(Uint8Val)) },                // 12 ← Flt32
		func(arg Native) Native { return FltVal(arg.(Uint8Val)) },                  // 13 ← Float
		func(arg Native) Native { return FltVal(arg.(Uint8Val)).BigFlt() },         // 14 ← BigFlt
		func(arg Native) Native { return UintVal(arg.(Uint8Val)).Ratio() },         // 15 ← Ratio
		func(arg Native) Native { return UintVal(arg.(Uint8Val)).Imag().Imag64() }, // 16 ← Imag64
		func(arg Native) Native { return UintVal(arg.(Uint8Val)).Imag() },          // 17 ← Imag
		func(arg Native) Native {
			if arg.(Uint8Val) > 0 {
				return TimeVal(time.Now())
			}
			return TimeVal{}
		}, // 18 ← Time
		func(arg Native) Native { return DuraVal(arg.(Uint8Val)) },                 // 19 ← Duration
		func(arg Native) Native { return ByteVal(arg.(Uint8Val)) },                 // 20 ← Byte
		func(arg Native) Native { return RuneVal(UintVal(arg.(Uint8Val))) },        // 21 ← Rune
		func(arg Native) Native { return BitFlag(UintVal(arg.(Uint8Val))) },        // 22 ← Flag
		func(arg Native) Native { return StrVal(arg.(Uint8Val).String()) },         // 23 ← String
		func(arg Native) Native { return StrVal(arg.(Uint8Val).String()).Bytes() }, // 24 ← Bytes
		func(arg Native) Native {
			if arg.(BoolVal).Bool() {
				return NewError(
					fmt.Errorf(
						"error occured during convertion from uint8"))
			}
			return ErrorVal{}
		}, // 25 ← Error
	},

	[]func(arg Native) Native{ // 7  ← Uint16
		func(arg Native) Native { return NilVal{} }, // 0  ← Nil
		func(arg Native) Native {
			if arg.(Uint16Val) > Uint16Val(0) {
				return BoolVal(true)
			}
			return BoolVal(false)
		}, // 1  ← Bool
		func(arg Native) Native { return Int8Val(arg.(Uint16Val)) },                 // 2  ← Int8
		func(arg Native) Native { return Int16Val(arg.(Uint16Val)) },                // 3  ← Int16
		func(arg Native) Native { return Int32Val(arg.(Uint16Val)) },                // 4  ← Int32
		func(arg Native) Native { return IntVal(arg.(Uint16Val)) },                  // 4  ← Int
		func(arg Native) Native { return IntVal(arg.(Uint16Val)).BigInt() },         // 5  ← BigInt
		func(arg Native) Native { return Uint8Val(arg.(Uint16Val)) },                // 6  ← Uint8
		func(arg Native) Native { return arg.(Uint16Val) },                          // 7  ← Uint16
		func(arg Native) Native { return Uint32Val(arg.(Uint16Val)) },               // 9  ← Uint32
		func(arg Native) Native { return UintVal(arg.(Uint16Val)) },                 // 10 ← Uint
		func(arg Native) Native { return Flt32Val(arg.(Uint16Val)) },                // 12 ← Flt32
		func(arg Native) Native { return FltVal(arg.(Uint16Val)) },                  // 13 ← Float
		func(arg Native) Native { return FltVal(arg.(Uint16Val)).BigFlt() },         // 14 ← BigFlt
		func(arg Native) Native { return UintVal(arg.(Uint16Val)).Ratio() },         // 15 ← Ratio
		func(arg Native) Native { return UintVal(arg.(Uint16Val)).Imag().Imag64() }, // 16 ← Imag64
		func(arg Native) Native { return UintVal(arg.(Uint16Val)).Imag() },          // 17 ← Imag
		func(arg Native) Native {
			if arg.(Uint16Val) > 0 {
				return TimeVal(time.Now())
			}
			return TimeVal{}
		}, // 18 ← Time
		func(arg Native) Native { return DuraVal(arg.(Uint16Val)) },                 // 19 ← Duration
		func(arg Native) Native { return ByteVal(arg.(Uint16Val)) },                 // 20 ← Byte
		func(arg Native) Native { return RuneVal(UintVal(arg.(Uint16Val))) },        // 21 ← Rune
		func(arg Native) Native { return BitFlag(UintVal(arg.(Uint16Val))) },        // 22 ← Flag
		func(arg Native) Native { return StrVal(arg.(Uint16Val).String()) },         // 23 ← String
		func(arg Native) Native { return StrVal(arg.(Uint16Val).String()).Bytes() }, // 24 ← Bytes
		func(arg Native) Native {
			if arg.(BoolVal).Bool() {
				return NewError(
					fmt.Errorf(
						"error occured during convertion from uint16"))
			}
			return ErrorVal{}
		}, // 25 ← Error
	},

	[]func(arg Native) Native{ // 9  ← Uint32
		func(arg Native) Native { return NilVal{} }, // 0  ← Nil
		func(arg Native) Native {
			if arg.(Uint32Val) > Uint32Val(0) {
				return BoolVal(true)
			}
			return BoolVal(false)
		}, // 1  ← Bool
		func(arg Native) Native { return IntVal(arg.(Uint32Val)) },                  // 2  ← Int8
		func(arg Native) Native { return IntVal(arg.(Uint32Val)) },                  // 3  ← Int16
		func(arg Native) Native { return IntVal(arg.(Uint32Val)) },                  // 4  ← Int32
		func(arg Native) Native { return IntVal(arg.(Uint32Val)) },                  // 4  ← Int
		func(arg Native) Native { return IntVal(arg.(Uint32Val)).BigInt() },         // 5  ← BigInt
		func(arg Native) Native { return Uint8Val(arg.(Uint32Val)) },                // 6  ← Uint8
		func(arg Native) Native { return Uint16Val(arg.(Uint32Val)) },               // 7  ← Uint16
		func(arg Native) Native { return Uint32Val(arg.(Uint32Val)) },               // 9  ← Uint32
		func(arg Native) Native { return UintVal(arg.(Uint32Val)) },                 // 10 ← Uint
		func(arg Native) Native { return Flt32Val(arg.(Uint32Val)) },                // 12 ← Flt32
		func(arg Native) Native { return FltVal(arg.(Uint32Val)) },                  // 13 ← Float
		func(arg Native) Native { return FltVal(arg.(Uint32Val)).BigFlt() },         // 14 ← BigFlt
		func(arg Native) Native { return UintVal(arg.(Uint32Val)).Ratio() },         // 15 ← Ratio
		func(arg Native) Native { return UintVal(arg.(Uint32Val)).Imag().Imag64() }, // 16 ← Imag64
		func(arg Native) Native { return UintVal(arg.(Uint32Val)).Imag() },          // 17 ← Imag
		func(arg Native) Native {
			if arg.(Uint32Val) > 0 {
				return TimeVal(time.Now())
			}
			return TimeVal{}
		}, // 18 ← Time
		func(arg Native) Native { return DuraVal(arg.(Uint32Val)) },                 // 19 ← Duration
		func(arg Native) Native { return ByteVal(arg.(Uint32Val)) },                 // 20 ← Byte
		func(arg Native) Native { return RuneVal(UintVal(arg.(Uint32Val))) },        // 21 ← Rune
		func(arg Native) Native { return BitFlag(UintVal(arg.(Uint32Val))) },        // 22 ← Flag
		func(arg Native) Native { return StrVal(arg.(Uint32Val).String()) },         // 23 ← String
		func(arg Native) Native { return StrVal(arg.(Uint32Val).String()).Bytes() }, // 24 ← Bytes
		func(arg Native) Native {
			if arg.(BoolVal).Bool() {
				return NewError(
					fmt.Errorf(
						"error occured during convertion from uint32"))
			}
			return ErrorVal{}
		}, // 25 ← Error
	},

	[]func(arg Native) Native{ // 10 ← Uint
		func(arg Native) Native { return NilVal{} }, // 0  ← Nil
		func(arg Native) Native {
			if arg.(UintVal) > UintVal(0) {
				return BoolVal(true)
			}
			return BoolVal(false)
		}, // 1  ← Bool
		func(arg Native) Native { return IntVal(arg.(UintVal)) },                  // 2  ← Int8
		func(arg Native) Native { return IntVal(arg.(UintVal)) },                  // 3  ← Int16
		func(arg Native) Native { return IntVal(arg.(UintVal)) },                  // 4  ← Int32
		func(arg Native) Native { return IntVal(arg.(UintVal)) },                  // 4  ← Int
		func(arg Native) Native { return IntVal(arg.(UintVal)).BigInt() },         // 5  ← BigInt
		func(arg Native) Native { return Uint8Val(arg.(UintVal)) },                // 6  ← Uint8
		func(arg Native) Native { return Uint16Val(arg.(UintVal)) },               // 7  ← Uint16
		func(arg Native) Native { return Uint32Val(arg.(UintVal)) },               // 9  ← Uint32
		func(arg Native) Native { return UintVal(arg.(UintVal)) },                 // 10 ← Uint
		func(arg Native) Native { return Flt32Val(arg.(UintVal)) },                // 12 ← Flt32
		func(arg Native) Native { return FltVal(arg.(UintVal)) },                  // 13 ← Float
		func(arg Native) Native { return FltVal(arg.(UintVal)).BigFlt() },         // 14 ← BigFlt
		func(arg Native) Native { return UintVal(arg.(UintVal)).Ratio() },         // 15 ← Ratio
		func(arg Native) Native { return UintVal(arg.(UintVal)).Imag().Imag64() }, // 16 ← Imag64
		func(arg Native) Native { return UintVal(arg.(UintVal)).Imag() },          // 17 ← Imag
		func(arg Native) Native {
			if arg.(UintVal) > 0 {
				return TimeVal(time.Now())
			}
			return TimeVal{}
		}, // 18 ← Time
		func(arg Native) Native { return DuraVal(arg.(UintVal)) },                 // 19 ← Duration
		func(arg Native) Native { return ByteVal(arg.(UintVal)) },                 // 20 ← Byte
		func(arg Native) Native { return RuneVal(UintVal(arg.(UintVal))) },        // 21 ← Rune
		func(arg Native) Native { return BitFlag(UintVal(arg.(UintVal))) },        // 22 ← Flag
		func(arg Native) Native { return StrVal(arg.(UintVal).String()) },         // 23 ← String
		func(arg Native) Native { return StrVal(arg.(UintVal).String()).Bytes() }, // 24 ← Bytes
		func(arg Native) Native {
			if arg.(BoolVal).Bool() {
				return NewError(
					fmt.Errorf(
						"error occured during convertion from uint"))
			}
			return ErrorVal{}
		}, // 25 ← Error
	},

	[]func(arg Native) Native{ // 12 ← Flt32
		func(arg Native) Native { return NilVal{} }, // 0  ← Nil
		func(arg Native) Native {
			if arg.(Flt32Val) > Flt32Val(0) {
				return BoolVal(true)
			}
			return BoolVal(false)
		}, // 1  ← Bool
		func(arg Native) Native { return IntVal(arg.(Flt32Val)) },                  // 2  ← Int8
		func(arg Native) Native { return IntVal(arg.(Flt32Val)) },                  // 3  ← Int16
		func(arg Native) Native { return IntVal(arg.(Flt32Val)) },                  // 4  ← Int32
		func(arg Native) Native { return IntVal(arg.(Flt32Val)) },                  // 4  ← Int
		func(arg Native) Native { return IntVal(arg.(Flt32Val)).BigInt() },         // 5  ← BigInt
		func(arg Native) Native { return Uint8Val(arg.(Flt32Val)) },                // 6  ← Uint8
		func(arg Native) Native { return Uint16Val(arg.(Flt32Val)) },               // 7  ← Uint16
		func(arg Native) Native { return Uint32Val(arg.(Flt32Val)) },               // 9  ← Uint32
		func(arg Native) Native { return UintVal(arg.(Flt32Val)) },                 // 10 ← Uint
		func(arg Native) Native { return Flt32Val(arg.(Flt32Val)) },                // 12 ← Flt32
		func(arg Native) Native { return arg.(Flt32Val) },                          // 13 ← Float
		func(arg Native) Native { return FltVal(arg.(Flt32Val)).BigFlt() },         // 14 ← BigFlt
		func(arg Native) Native { return UintVal(arg.(Flt32Val)).Ratio() },         // 15 ← Ratio
		func(arg Native) Native { return UintVal(arg.(Flt32Val)).Imag().Imag64() }, // 16 ← Imag64
		func(arg Native) Native { return UintVal(arg.(Flt32Val)).Imag() },          // 17 ← Imag
		func(arg Native) Native {
			if arg.(Flt32Val) > 0 {
				return TimeVal(time.Now())
			}
			return TimeVal{}
		}, // 18 ← Time
		func(arg Native) Native { return DuraVal(arg.(Flt32Val)) },                 // 19 ← Duration
		func(arg Native) Native { return ByteVal(arg.(Flt32Val)) },                 // 20 ← Byte
		func(arg Native) Native { return RuneVal(UintVal(arg.(Flt32Val))) },        // 21 ← Rune
		func(arg Native) Native { return BitFlag(UintVal(arg.(Flt32Val))) },        // 22 ← Flag
		func(arg Native) Native { return StrVal(arg.(Flt32Val).String()) },         // 23 ← String
		func(arg Native) Native { return StrVal(arg.(Flt32Val).String()).Bytes() }, // 24 ← Bytes
		func(arg Native) Native {
			if arg.(BoolVal).Bool() {
				return NewError(
					fmt.Errorf(
						"error occured during convertion from uint"))
			}
			return ErrorVal{}
		}, // 25 ← Error
	},

	[]func(arg Native) Native{ // 13 ← Float
		func(arg Native) Native { return NilVal{} }, // 0  ← Nil
		func(arg Native) Native {
			if arg.(FltVal) > FltVal(0) {
				return BoolVal(true)
			}
			return BoolVal(false)
		}, // 1  ← Bool
		func(arg Native) Native { return IntVal(arg.(FltVal)) },                  // 2  ← Int8
		func(arg Native) Native { return IntVal(arg.(FltVal)) },                  // 3  ← Int16
		func(arg Native) Native { return IntVal(arg.(FltVal)) },                  // 4  ← Int32
		func(arg Native) Native { return IntVal(arg.(FltVal)) },                  // 4  ← Int
		func(arg Native) Native { return IntVal(arg.(FltVal)).BigInt() },         // 5  ← BigInt
		func(arg Native) Native { return Uint8Val(arg.(FltVal)) },                // 6  ← Uint8
		func(arg Native) Native { return Uint16Val(arg.(FltVal)) },               // 7  ← Uint16
		func(arg Native) Native { return Uint32Val(arg.(FltVal)) },               // 9  ← Uint32
		func(arg Native) Native { return UintVal(arg.(FltVal)) },                 // 10 ← Uint
		func(arg Native) Native { return Flt32Val(arg.(FltVal)) },                // 12 ← Flt32
		func(arg Native) Native { return arg.(FltVal) },                          // 13 ← Float
		func(arg Native) Native { return FltVal(arg.(FltVal)).BigFlt() },         // 14 ← BigFlt
		func(arg Native) Native { return UintVal(arg.(FltVal)).Ratio() },         // 15 ← Ratio
		func(arg Native) Native { return UintVal(arg.(FltVal)).Imag().Imag64() }, // 16 ← Imag64
		func(arg Native) Native { return UintVal(arg.(FltVal)).Imag() },          // 17 ← Imag
		func(arg Native) Native {
			if arg.(FltVal) > 0 {
				return TimeVal(time.Now())
			}
			return TimeVal{}
		}, // 18 ← Time
		func(arg Native) Native { return DuraVal(arg.(FltVal)) },                 // 19 ← Duration
		func(arg Native) Native { return ByteVal(arg.(FltVal)) },                 // 20 ← Byte
		func(arg Native) Native { return RuneVal(UintVal(arg.(FltVal))) },        // 21 ← Rune
		func(arg Native) Native { return BitFlag(UintVal(arg.(FltVal))) },        // 22 ← Flag
		func(arg Native) Native { return StrVal(arg.(FltVal).String()) },         // 23 ← String
		func(arg Native) Native { return StrVal(arg.(FltVal).String()).Bytes() }, // 24 ← Bytes
		func(arg Native) Native {
			if arg.(BoolVal).Bool() {
				return NewError(
					fmt.Errorf(
						"error occured during convertion from uint"))
			}
			return ErrorVal{}
		}, // 25 ← Error
	},

	[]func(arg Native) Native{ // 14 ← BigFlt
		func(arg Native) Native { return NilVal{} }, // 0  ← Nil
		func(arg Native) Native {
			if arg.(BigFltVal).GoBigInt().Cmp(IntVal(0).BigInt().GoBigInt()) != 0 {
				return BoolVal(true)
			}
			return BoolVal(false)
		}, // 1  ← Bool
		func(arg Native) Native { return IntVal(arg.(BigFltVal).Int()) },           // 2  ← Int8
		func(arg Native) Native { return IntVal(arg.(BigFltVal).Int()) },           // 3  ← Int16
		func(arg Native) Native { return IntVal(arg.(BigFltVal).Int()) },           // 4  ← Int32
		func(arg Native) Native { return IntVal(arg.(BigFltVal).Int()) },           // 4  ← Int
		func(arg Native) Native { return IntVal(arg.(BigFltVal).Int()).BigInt() },  // 5  ← BigInt
		func(arg Native) Native { return Uint8Val(arg.(BigFltVal).Uint()) },        // 6  ← Uint8
		func(arg Native) Native { return Uint16Val(arg.(BigFltVal).Uint()) },       // 7  ← Uint16
		func(arg Native) Native { return Uint32Val(arg.(BigFltVal).Uint()) },       // 9  ← Uint32
		func(arg Native) Native { return UintVal(arg.(BigFltVal).Uint()) },         // 10 ← Uint
		func(arg Native) Native { return arg.(BigFltVal).Float().Flt32() },         // 12 ← Flt32
		func(arg Native) Native { return arg.(BigFltVal).Float() },                 // 13 ← Float
		func(arg Native) Native { return FltVal(arg.(BigFltVal).Int()).BigFlt() },  // 14 ← BigFlt
		func(arg Native) Native { return (*RatioVal)(arg.(BigFltVal).Ratio()) },    // 15 ← Ratio
		func(arg Native) Native { return arg.(BigFltVal).Float().Imag().Imag64() }, // 16 ← Imag64
		func(arg Native) Native { return arg.(BigFltVal).Float().Imag() },          // 17 ← Imag
		func(arg Native) Native {
			if arg.(BigFltVal).Int() > 0 {
				return TimeVal(time.Now())
			}
			return TimeVal{}
		}, // 18 ← Time
		func(arg Native) Native { return DuraVal(arg.(BigFltVal).Int()) },                 // 19 ← Duration
		func(arg Native) Native { return ByteVal(arg.(BigFltVal).Int()) },                 // 20 ← Byte
		func(arg Native) Native { return RuneVal(UintVal(arg.(BigFltVal).Int())) },        // 21 ← Rune
		func(arg Native) Native { return BitFlag(UintVal(arg.(BigFltVal).Int())) },        // 22 ← Flag
		func(arg Native) Native { return StrVal(arg.(BigFltVal).Int().String()) },         // 23 ← String
		func(arg Native) Native { return StrVal(arg.(BigFltVal).Int().String()).Bytes() }, // 24 ← Bytes
		func(arg Native) Native {
			if arg.(BoolVal).Bool() {
				return NewError(
					fmt.Errorf(
						"error occured during convertion from uint"))
			}
			return ErrorVal{}
		}, // 25 ← Error
	},

	[]func(arg Native) Native{ // 15 ← Ratio
		func(arg Native) Native { return NilVal{} }, // 0  ← Nil
		func(arg Native) Native {
			if arg.(RatioVal).GoBigInt().Cmp(IntVal(0).BigInt().GoBigInt()) != 0 {
				return BoolVal(true)
			}
			return BoolVal(false)
		}, // 1  ← Bool
		func(arg Native) Native { return IntVal(arg.(RatioVal).Int()) },           // 2  ← Int8
		func(arg Native) Native { return IntVal(arg.(RatioVal).Int()) },           // 3  ← Int16
		func(arg Native) Native { return IntVal(arg.(RatioVal).Int()) },           // 4  ← Int32
		func(arg Native) Native { return arg.(RatioVal).Int() },                   // 4  ← Int
		func(arg Native) Native { return IntVal(arg.(RatioVal).Int()).BigInt() },  // 5  ← BigInt
		func(arg Native) Native { return Uint8Val(arg.(RatioVal).Uint()) },        // 6  ← Uint8
		func(arg Native) Native { return Uint16Val(arg.(RatioVal).Uint()) },       // 7  ← Uint16
		func(arg Native) Native { return Uint32Val(arg.(RatioVal).Uint()) },       // 9  ← Uint32
		func(arg Native) Native { return UintVal(arg.(RatioVal).Uint()) },         // 10 ← Uint
		func(arg Native) Native { return arg.(RatioVal).Float().Flt32() },         // 12 ← Flt32
		func(arg Native) Native { return arg.(RatioVal).Float() },                 // 13 ← Float
		func(arg Native) Native { return FltVal(arg.(RatioVal).Int()).BigFlt() },  // 14 ← BigFlt
		func(arg Native) Native { return arg.(RatioVal) },                         // 15 ← Ratio
		func(arg Native) Native { return arg.(RatioVal).Float().Imag().Imag64() }, // 16 ← Imag64
		func(arg Native) Native { return arg.(RatioVal).Float().Imag() },          // 17 ← Imag
		func(arg Native) Native {
			if arg.(RatioVal).Int() > 0 {
				return TimeVal(time.Now())
			}
			return TimeVal{}
		}, // 18 ← Time
		func(arg Native) Native { return DuraVal(arg.(RatioVal).Int()) },                 // 19 ← Duration
		func(arg Native) Native { return ByteVal(arg.(RatioVal).Int()) },                 // 20 ← Byte
		func(arg Native) Native { return RuneVal(UintVal(arg.(RatioVal).Int())) },        // 21 ← Rune
		func(arg Native) Native { return BitFlag(UintVal(arg.(RatioVal).Int())) },        // 22 ← Flag
		func(arg Native) Native { return StrVal(arg.(RatioVal).Int().String()) },         // 23 ← String
		func(arg Native) Native { return StrVal(arg.(RatioVal).Int().String()).Bytes() }, // 24 ← Bytes
		func(arg Native) Native {
			if arg.(BoolVal).Bool() {
				return NewError(
					fmt.Errorf(
						"error occured during convertion from uint"))
			}
			return ErrorVal{}
		}, // 25 ← Error
	},

	[]func(arg Native) Native{ // 16 ← Imag64
		func(arg Native) Native { return NilVal{} }, // 0  ← Nil
		func(arg Native) Native {
			if arg.(Imag64Val) != Imag64Val(complex64(0)) {
				return BoolVal(true)
			}
			return BoolVal(false)
		}, // 1  ← Bool
		func(arg Native) Native { return ImagVal(arg.(Imag64Val)) },                          // 2  ← Int8
		func(arg Native) Native { return ImagVal(arg.(Imag64Val)) },                          // 3  ← Int16
		func(arg Native) Native { return ImagVal(arg.(Imag64Val)) },                          // 4  ← Int32
		func(arg Native) Native { return ImagVal(arg.(Imag64Val)) },                          // 4  ← Int
		func(arg Native) Native { return ImagVal(arg.(Imag64Val)).BigInt() },                 // 5  ← BigInt
		func(arg Native) Native { return Uint8Val(ImagVal(arg.(Imag64Val)).Uint()) },         // 6  ← Uint8
		func(arg Native) Native { return Uint16Val(ImagVal(arg.(Imag64Val)).Uint()) },        // 7  ← Uint16
		func(arg Native) Native { return Uint32Val(ImagVal(arg.(Imag64Val)).Uint()) },        // 9  ← Uint32
		func(arg Native) Native { return UintVal(ImagVal(arg.(Imag64Val)).Uint()) },          // 10 ← Uint
		func(arg Native) Native { return Flt32Val(ImagVal(arg.(Imag64Val)).Float()) },        // 12 ← Flt32
		func(arg Native) Native { return FltVal(ImagVal(arg.(Imag64Val)).Float()) },          // 13 ← Float
		func(arg Native) Native { return FltVal(ImagVal(arg.(Imag64Val)).Float()).BigFlt() }, // 14 ← BigFlt
		func(arg Native) Native { return ImagVal(arg.(Imag64Val)).Float().Ratio() },          // 15 ← Ratio
		func(arg Native) Native { return arg.(Imag64Val) },                                   // 16 ← Imag64
		func(arg Native) Native { return ImagVal(arg.(Imag64Val)) },                          // 17 ← Imag
		func(arg Native) Native {
			if ImagVal(arg.(Imag64Val)).Int() > 0 {
				return TimeVal(time.Now())
			}
			return TimeVal{}
		}, // 18 ← Time
		func(arg Native) Native { return DuraVal(ImagVal(arg.(Imag64Val)).Int()) },                 // 19 ← Duration
		func(arg Native) Native { return DuraVal(ImagVal(arg.(Imag64Val)).Int()) },                 // 20 ← Byte
		func(arg Native) Native { return RuneVal(ImagVal(arg.(Imag64Val)).Int()) },                 // 21 ← Rune
		func(arg Native) Native { return BitFlag(ImagVal(arg.(Imag64Val)).Int()) },                 // 22 ← Flag
		func(arg Native) Native { return StrVal(ImagVal(arg.(Imag64Val)).Int().String()) },         // 23 ← String
		func(arg Native) Native { return StrVal(ImagVal(arg.(Imag64Val)).Int().String()).Bytes() }, // 24 ← Bytes
		func(arg Native) Native {
			if arg.(BoolVal).Bool() {
				return NewError(
					fmt.Errorf(
						"error occured during convertion from uint"))
			}
			return ErrorVal{}
		}, // 25 ← Error
	},

	[]func(arg Native) Native{ // 17 ← Imag
		func(arg Native) Native { return NilVal{} }, // 0  ← Nil
		func(arg Native) Native {
			if arg.(ImagVal) != ImagVal(complex128(0)) {
				return BoolVal(true)
			}
			return BoolVal(false)
		}, // 1  ← Bool
		func(arg Native) Native { return IntVal(arg.(ImagVal).Int()) },           // 2  ← Int8
		func(arg Native) Native { return IntVal(arg.(ImagVal).Int()) },           // 3  ← Int16
		func(arg Native) Native { return IntVal(arg.(ImagVal).Int()) },           // 4  ← Int32
		func(arg Native) Native { return arg.(ImagVal).Int() },                   // 4  ← Int
		func(arg Native) Native { return IntVal(arg.(ImagVal).Int()).BigInt() },  // 5  ← BigInt
		func(arg Native) Native { return Uint8Val(arg.(ImagVal).Uint()) },        // 6  ← Uint8
		func(arg Native) Native { return Uint16Val(arg.(ImagVal).Uint()) },       // 7  ← Uint16
		func(arg Native) Native { return Uint32Val(arg.(ImagVal).Uint()) },       // 9  ← Uint32
		func(arg Native) Native { return UintVal(arg.(ImagVal).Uint()) },         // 10 ← Uint
		func(arg Native) Native { return arg.(ImagVal).Float().Flt32() },         // 12 ← Flt32
		func(arg Native) Native { return arg.(ImagVal).Float() },                 // 13 ← Float
		func(arg Native) Native { return FltVal(arg.(ImagVal).Int()).BigFlt() },  // 14 ← BigFlt
		func(arg Native) Native { return arg.(ImagVal) },                         // 15 ← Ratio
		func(arg Native) Native { return arg.(ImagVal).Float().Imag().Imag64() }, // 16 ← Imag64
		func(arg Native) Native { return arg.(ImagVal).Float().Imag() },          // 17 ← Imag
		func(arg Native) Native {
			if arg.(ImagVal).Int() > 0 {
				return TimeVal(time.Now())
			}
			return TimeVal{}
		}, // 18 ← Time
		func(arg Native) Native { return DuraVal(arg.(ImagVal).Int()) },                 // 19 ← Duration
		func(arg Native) Native { return ByteVal(arg.(ImagVal).Int()) },                 // 20 ← Byte
		func(arg Native) Native { return RuneVal(UintVal(arg.(ImagVal).Int())) },        // 21 ← Rune
		func(arg Native) Native { return BitFlag(UintVal(arg.(ImagVal).Int())) },        // 22 ← Flag
		func(arg Native) Native { return StrVal(arg.(ImagVal).Int().String()) },         // 23 ← String
		func(arg Native) Native { return StrVal(arg.(ImagVal).Int().String()).Bytes() }, // 24 ← Bytes
		func(arg Native) Native {
			if arg.(BoolVal).Bool() {
				return NewError(
					fmt.Errorf(
						"error occured during convertion from uint"))
			}
			return ErrorVal{}
		}, // 25 ← Error
	},

	[]func(arg Native) Native{ // 18 ← Time
		func(arg Native) Native { return NilVal{} }, // 0  ← Nil
		func(arg Native) Native {
			if arg.(TimeVal).Time().Before(time.Now()) {
				return BoolVal(true)
			}
			return BoolVal(false)
		}, // 1  ← Bool
		func(arg Native) Native { return IntVal(arg.(TimeVal).Int()) },          // 2  ← Int8
		func(arg Native) Native { return IntVal(arg.(TimeVal).Int()) },          // 3  ← Int16
		func(arg Native) Native { return IntVal(arg.(TimeVal).Int()) },          // 4  ← Int32
		func(arg Native) Native { return arg.(TimeVal).Int() },                  // 4  ← Int
		func(arg Native) Native { return IntVal(arg.(TimeVal).Int()).BigInt() }, // 5  ← BigInt
		func(arg Native) Native { return Uint8Val(arg.(TimeVal).Uint()) },       // 6  ← Uint8
		func(arg Native) Native { return Uint16Val(arg.(TimeVal).Uint()) },      // 7  ← Uint16
		func(arg Native) Native { return Uint32Val(arg.(TimeVal).Uint()) },      // 9  ← Uint32
		func(arg Native) Native { return UintVal(arg.(TimeVal).Uint()) },        // 10 ← Uint
		func(arg Native) Native { return arg.(TimeVal).Float().Flt32() },        // 12 ← Flt32
		func(arg Native) Native { return arg.(TimeVal).Float() },                // 13 ← Float
		func(arg Native) Native { return FltVal(arg.(TimeVal).Int()).BigFlt() }, // 14 ← BigFlt
		func(arg Native) Native { return arg.(TimeVal).Ratio() },                // 15 ← Ratio
		func(arg Native) Native { return Imag64Val(arg.(TimeVal).Imag()) },      // 16 ← Imag64
		func(arg Native) Native { return arg.(TimeVal).Imag() },                 // 17 ← Imag
		func(arg Native) Native { return arg.(TimeVal) },                        // 18 ← Time
		func(arg Native) Native {
			if time.Now().Before(arg.(TimeVal).Time()) {
				return DuraVal(arg.(TimeVal).Time().Sub(time.Now()))
			}
			return DuraVal(time.Now().Sub(arg.(TimeVal).Time()))
		}, // 19 ← Duration
		func(arg Native) Native { return ByteVal([]byte(arg.(TimeVal).String())[0]) }, // 20 ← Byte
		func(arg Native) Native { return RuneVal([]rune(arg.(TimeVal).String())[0]) }, // 21 ← Rune
		func(arg Native) Native { return BitFlag(UintVal(arg.(TimeVal).Int())) },      // 22 ← Flag
		func(arg Native) Native { return StrVal(arg.(TimeVal).String()) },             // 23 ← String
		func(arg Native) Native { return BytesVal([]byte(arg.(TimeVal).String())) },   // 24 ← Bytes
		func(arg Native) Native {
			if arg.(BoolVal).Bool() {
				return NewError(
					fmt.Errorf(
						"error occured during convertion from uint"))
			}
			return ErrorVal{}
		}, // 25 ← Error
	},

	[]func(arg Native) Native{ // 19 ← Duration
		func(arg Native) Native { return NilVal{} }, // 0  ← Nil
		func(arg Native) Native {
			if arg.(DuraVal) > 0 {
				return BoolVal(true)
			}
			return BoolVal(false)
		}, // 1  ← Bool
		func(arg Native) Native { return IntVal(arg.(DuraVal).Int()) },          // 2  ← Int8
		func(arg Native) Native { return IntVal(arg.(DuraVal).Int()) },          // 3  ← Int16
		func(arg Native) Native { return IntVal(arg.(DuraVal).Int()) },          // 4  ← Int32
		func(arg Native) Native { return arg.(DuraVal).Int() },                  // 4  ← Int
		func(arg Native) Native { return IntVal(arg.(DuraVal).Int()).BigInt() }, // 5  ← BigInt
		func(arg Native) Native { return Uint8Val(arg.(DuraVal).Uint()) },       // 6  ← Uint8
		func(arg Native) Native { return Uint16Val(arg.(DuraVal).Uint()) },      // 7  ← Uint16
		func(arg Native) Native { return Uint32Val(arg.(DuraVal).Uint()) },      // 9  ← Uint32
		func(arg Native) Native { return UintVal(arg.(DuraVal).Uint()) },        // 10 ← Uint
		func(arg Native) Native { return arg.(DuraVal).Float().Flt32() },        // 12 ← Flt32
		func(arg Native) Native { return arg.(DuraVal).Float() },                // 13 ← Float
		func(arg Native) Native { return FltVal(arg.(DuraVal).Int()).BigFlt() }, // 14 ← BigFlt
		func(arg Native) Native { return arg.(DuraVal).Ratio() },                // 15 ← Ratio
		func(arg Native) Native { return Imag64Val(arg.(DuraVal).Imag()) },      // 16 ← Imag64
		func(arg Native) Native { return arg.(DuraVal).Imag() },                 // 17 ← Imag
		func(arg Native) Native {
			if arg.(DuraVal).Int() > 0 {
				return TimeVal(time.Now())
			}
			return TimeVal{}
		}, // 18 ← Time
		func(arg Native) Native { return arg.(DuraVal) },                              // 19 ← Duration
		func(arg Native) Native { return ByteVal([]byte(arg.(DuraVal).String())[0]) }, // 20 ← Byte
		func(arg Native) Native { return RuneVal([]rune(arg.(DuraVal).String())[0]) }, // 21 ← Rune
		func(arg Native) Native { return BitFlag(UintVal(arg.(DuraVal).Int())) },      // 22 ← Flag
		func(arg Native) Native { return StrVal(arg.(DuraVal).String()) },             // 23 ← String
		func(arg Native) Native { return BytesVal([]byte(arg.(DuraVal).String())) },   // 24 ← Bytes
		func(arg Native) Native {
			if arg.(BoolVal).Bool() {
				return NewError(
					fmt.Errorf(
						"error occured during convertion from uint"))
			}
			return ErrorVal{}
		}, // 25 ← Error
	},

	[]func(arg Native) Native{ // 20 ← Byte
		func(arg Native) Native { return NilVal{} }, // 0  ← Nil
		func(arg Native) Native {
			if arg.(ByteVal).GoByte() > byte(0) {
				return BoolVal(true)
			}
			return BoolVal(false)
		}, // 1  ← Bool
		func(arg Native) Native { return IntVal(arg.(ByteVal).Int()) },          // 2  ← Int8
		func(arg Native) Native { return IntVal(arg.(ByteVal).Int()) },          // 3  ← Int16
		func(arg Native) Native { return IntVal(arg.(ByteVal).Int()) },          // 4  ← Int32
		func(arg Native) Native { return arg.(ByteVal).Int() },                  // 4  ← Int
		func(arg Native) Native { return IntVal(arg.(ByteVal).Int()).BigInt() }, // 5  ← BigInt
		func(arg Native) Native { return Uint8Val(arg.(ByteVal).Uint()) },       // 6  ← Uint8
		func(arg Native) Native { return Uint16Val(arg.(ByteVal).Uint()) },      // 7  ← Uint16
		func(arg Native) Native { return Uint32Val(arg.(ByteVal).Uint()) },      // 9  ← Uint32
		func(arg Native) Native { return UintVal(arg.(ByteVal).Uint()) },        // 10 ← Uint
		func(arg Native) Native { return arg.(ByteVal).Float().Flt32() },        // 12 ← Flt32
		func(arg Native) Native { return arg.(ByteVal).Float() },                // 13 ← Float
		func(arg Native) Native { return FltVal(arg.(ByteVal).Int()).BigFlt() }, // 14 ← BigFlt
		func(arg Native) Native { return arg.(ByteVal).Ratio() },                // 15 ← Ratio
		func(arg Native) Native { return Imag64Val(arg.(ByteVal).Imag()) },      // 16 ← Imag64
		func(arg Native) Native { return arg.(ByteVal).Imag() },                 // 17 ← Imag
		func(arg Native) Native {
			if arg.(ByteVal).Int() > 0 {
				return TimeVal(time.Now())
			}
			return TimeVal{}
		}, // 18 ← Time
		func(arg Native) Native { return arg.(ByteVal) },                            // 19 ← Duration
		func(arg Native) Native { return arg.(ByteVal) },                            // 20 ← Byte
		func(arg Native) Native { return RuneVal(rune(arg.(ByteVal))) },             // 21 ← Rune
		func(arg Native) Native { return BitFlag(UintVal(arg.(ByteVal))) },          // 22 ← Flag
		func(arg Native) Native { return StrVal(string(arg.(ByteVal))) },            // 23 ← String
		func(arg Native) Native { return BytesVal([]byte{arg.(ByteVal).GoByte()}) }, // 24 ← Bytes
		func(arg Native) Native {
			if arg.(BoolVal).Bool() {
				return NewError(
					fmt.Errorf(
						"error occured during convertion from uint"))
			}
			return ErrorVal{}
		}, // 25 ← Error
	},

	[]func(arg Native) Native{ // 21 ← Rune
		func(arg Native) Native { return NilVal{} }, // 0  ← Nil
		func(arg Native) Native {
			if rune(arg.(RuneVal)) > rune(0) {
				return BoolVal(true)
			}
			return BoolVal(false)
		}, // 1  ← Bool
		func(arg Native) Native { return IntVal(int(arg.(RuneVal))) },                   // 2  ← Int8
		func(arg Native) Native { return IntVal(int(arg.(RuneVal))) },                   // 3  ← Int16
		func(arg Native) Native { return IntVal(int(arg.(RuneVal))) },                   // 4  ← Int32
		func(arg Native) Native { return IntVal(int(arg.(RuneVal))) },                   // 4  ← Int
		func(arg Native) Native { return IntVal(int(arg.(RuneVal))).BigInt() },          // 5  ← BigInt
		func(arg Native) Native { return Uint8Val(uint(arg.(RuneVal))) },                // 6  ← Uint8
		func(arg Native) Native { return Uint16Val(uint(arg.(RuneVal))) },               // 7  ← Uint16
		func(arg Native) Native { return Uint32Val(uint(arg.(RuneVal))) },               // 9  ← Uint32
		func(arg Native) Native { return UintVal(uint(arg.(RuneVal))) },                 // 10 ← Uint
		func(arg Native) Native { return FltVal(float64(arg.(RuneVal))).Flt32() },       // 12 ← Flt32
		func(arg Native) Native { return FltVal(arg.(RuneVal)) },                        // 13 ← Float
		func(arg Native) Native { return FltVal(float64(arg.(RuneVal))).BigFlt() },      // 14 ← BigFlt
		func(arg Native) Native { return IntVal(int(arg.(RuneVal))).Ratio() },           // 15 ← Ratio
		func(arg Native) Native { return Imag64Val(IntVal(int(arg.(RuneVal))).Imag()) }, // 16 ← Imag64
		func(arg Native) Native { return IntVal(int(arg.(RuneVal))).Imag() },            // 17 ← Imag
		func(arg Native) Native {
			if int(arg.(RuneVal)) > 0 {
				return TimeVal(time.Now())
			}
			return TimeVal{}
		}, // 18 ← Time
		func(arg Native) Native { return arg.(RuneVal) },                               // 19 ← Duration
		func(arg Native) Native { return arg.(RuneVal) },                               // 20 ← Byte
		func(arg Native) Native { return RuneVal(rune(arg.(RuneVal))) },                // 21 ← Rune
		func(arg Native) Native { return BitFlag(UintVal(arg.(RuneVal))) },             // 22 ← Flag
		func(arg Native) Native { return StrVal(string(arg.(RuneVal))) },               // 23 ← String
		func(arg Native) Native { return BytesVal([]byte{byte(rune(arg.(RuneVal)))}) }, // 24 ← Bytes
		func(arg Native) Native {
			if arg.(BoolVal).Bool() {
				return NewError(
					fmt.Errorf(
						"error occured during convertion from uint"))
			}
			return ErrorVal{}
		}, // 25 ← Error
	},

	[]func(arg Native) Native{ // 22 ← Flag
		func(arg Native) Native { return NilVal{} }, // 0  ← Nil
		func(arg Native) Native {
			if rune(arg.(BitFlag)) > rune(0) {
				return BoolVal(true)
			}
			return BoolVal(false)
		}, // 1  ← Bool
		func(arg Native) Native { return IntVal(int(arg.(BitFlag))) },                   // 2  ← Int8
		func(arg Native) Native { return IntVal(int(arg.(BitFlag))) },                   // 3  ← Int16
		func(arg Native) Native { return IntVal(int(arg.(BitFlag))) },                   // 4  ← Int32
		func(arg Native) Native { return IntVal(int(arg.(BitFlag))) },                   // 4  ← Int
		func(arg Native) Native { return IntVal(int(arg.(BitFlag))).BigInt() },          // 5  ← BigInt
		func(arg Native) Native { return Uint8Val(uint(arg.(BitFlag))) },                // 6  ← Uint8
		func(arg Native) Native { return Uint16Val(uint(arg.(BitFlag))) },               // 7  ← Uint16
		func(arg Native) Native { return Uint32Val(uint(arg.(BitFlag))) },               // 9  ← Uint32
		func(arg Native) Native { return UintVal(uint(arg.(BitFlag))) },                 // 10 ← Uint
		func(arg Native) Native { return FltVal(float64(arg.(BitFlag))).Flt32() },       // 12 ← Flt32
		func(arg Native) Native { return FltVal(arg.(BitFlag)) },                        // 13 ← Float
		func(arg Native) Native { return FltVal(float64(arg.(BitFlag))).BigFlt() },      // 14 ← BigFlt
		func(arg Native) Native { return IntVal(int(arg.(BitFlag))).Ratio() },           // 15 ← Ratio
		func(arg Native) Native { return Imag64Val(IntVal(int(arg.(BitFlag))).Imag()) }, // 16 ← Imag64
		func(arg Native) Native { return IntVal(int(arg.(BitFlag))).Imag() },            // 17 ← Imag
		func(arg Native) Native {
			if int(arg.(BitFlag)) > 0 {
				return TimeVal(time.Now())
			}
			return TimeVal{}
		}, // 18 ← Time
		func(arg Native) Native { return arg.(BitFlag) },                               // 19 ← Duration
		func(arg Native) Native { return arg.(BitFlag) },                               // 20 ← Byte
		func(arg Native) Native { return RuneVal(rune(arg.(BitFlag))) },                // 21 ← Rune
		func(arg Native) Native { return BitFlag(UintVal(arg.(BitFlag))) },             // 22 ← Flag
		func(arg Native) Native { return StrVal(string(arg.(BitFlag))) },               // 23 ← String
		func(arg Native) Native { return BytesVal([]byte{byte(rune(arg.(BitFlag)))}) }, // 24 ← Bytes
		func(arg Native) Native {
			if arg.(BoolVal).Bool() {
				return NewError(
					fmt.Errorf(
						"error occured during convertion from uint"))
			}
			return ErrorVal{}
		}, // 25 ← Error
	},

	[]func(arg Native) Native{ // 23 ← String
		func(arg Native) Native { return NilVal{} },                   // 0  ← Nil
		func(arg Native) Native { return arg.(StrVal).ReadBoolVal() }, // 1  ← Bool
		func(arg Native) Native { // 2  ← Int8
			if i := arg.(StrVal).ReadIntVal(); i.Type().Match(Int) {
				return Int8Val(i.(IntVal).Int())
			}
			return NewNil()
		},
		func(arg Native) Native { // 2  ← Int16
			if i := arg.(StrVal).ReadIntVal(); i.Type().Match(Int) {
				return Int16Val(i.(IntVal).Int())
			}
			return NewNil()
		},
		func(arg Native) Native { // 2  ← Int32
			if i := arg.(StrVal).ReadIntVal(); i.Type().Match(Int) {
				return Int32Val(i.(IntVal).Int())
			}
			return NewNil()
		},
		func(arg Native) Native { // 2  ← Int
			if i := arg.(StrVal).ReadIntVal(); i.Type().Match(Int) {
				return IntVal(i.(IntVal).Int())
			}
			return NewNil()
		},
		func(arg Native) Native { // 2  ← Int32
			if i := arg.(StrVal).ReadIntVal(); i.Type().Match(Int) {
				return (*BigIntVal)(big.NewInt(int64(i.(IntVal))))
			}
			return NewNil()
		},
		func(arg Native) Native { // 2  ← Uint8
			if i := arg.(StrVal).ReadUintVal(); i.Type().Match(Uint) {
				return Uint8Val(i.(UintVal).Uint())
			}
			return NewNil()
		},
		func(arg Native) Native { // 2  ← Uint16
			if i := arg.(StrVal).ReadUintVal(); i.Type().Match(Uint) {
				return Uint16Val(i.(UintVal).Uint())
			}
			return NewNil()
		},
		func(arg Native) Native { // 2  ← Uint32
			if i := arg.(StrVal).ReadUintVal(); i.Type().Match(Uint) {
				return Uint32Val(i.(UintVal).Uint())
			}
			return NewNil()
		},
		func(arg Native) Native { // 11  ← Uint
			if i := arg.(StrVal).ReadUintVal(); i.Type().Match(Uint) {
				return UintVal(i.(UintVal).Uint())
			}
			return NewNil()
		},
		func(arg Native) Native { // 11  ← Flt32
			if i := arg.(StrVal).ReadFloatVal(); i.Type().Match(Float) {
				return Flt32Val(i.(FltVal).Float())
			}
			return NewNil()
		},
		func(arg Native) Native { // 12  ← Float
			if i := arg.(StrVal).ReadFloatVal(); i.Type().Match(Float) {
				return (*BigFltVal)(big.NewFloat(float64(i.(FltVal))))
			}
			return NewNil()
		},
		func(arg Native) Native { // 13  ← BigFlt
			if i := arg.(StrVal).ReadFloatVal(); i.Type().Match(Float) {
				return (*BigFltVal)(big.NewFloat(float64(i.(FltVal))))
			}
			return NewNil()
		},
		func(arg Native) Native { return IntVal(int(arg.(BitFlag))).Ratio() },           // 14 ← Ratio
		func(arg Native) Native { return Imag64Val(IntVal(int(arg.(BitFlag))).Imag()) }, // 15 ← Imag64
		func(arg Native) Native { return IntVal(int(arg.(BitFlag))).Imag() },            // 16 ← Imag
		func(arg Native) Native { // 18 ← Time
			if int(arg.(BitFlag)) > 0 {
				return TimeVal(time.Now())
			}
			return TimeVal{}
		},
		func(arg Native) Native { return arg.(BitFlag) },                            // 19 ← Duration
		func(arg Native) Native { return ByteVal([]byte(string(arg.(StrVal)))[0]) }, // 20 ← Byte
		func(arg Native) Native { return RuneVal([]rune(string(arg.(StrVal)))[0]) }, // 21 ← Rune
		func(arg Native) Native { return RuneVal([]rune(string(arg.(StrVal)))[0]) }, // 22 ← Flag
		func(arg Native) Native { return arg.(StrVal) },                             // 23 ← String
		func(arg Native) Native { return BytesVal([]byte(string(arg.(StrVal)))) },   // 24 ← Bytes
		func(arg Native) Native {
			return NewError(
				fmt.Errorf(
					"error occured during convertion from uint"))
			return ErrorVal{}
		}, // 25 ← Error
	},

	[]func(arg Native) Native{ // 24 ← Bytes
		func(Native) Native { return NilVal{} },                  // 0  ← Nil
		func(Native) Native { return BoolVal(false) },            // 1  ← Bool
		func(Native) Native { return Int8Val(0) },                // 2  ← Int8
		func(Native) Native { return Int16Val(0) },               // 3  ← Int16
		func(Native) Native { return Int32Val(0) },               // 4  ← Int32
		func(Native) Native { return IntVal(0) },                 // 4  ← Int
		func(Native) Native { return IntVal(0).BigInt() },        // 5  ← BigInt
		func(Native) Native { return Uint8Val(0) },               // 6  ← Uint8
		func(Native) Native { return Uint16Val(0) },              // 7  ← Uint16
		func(Native) Native { return Uint32Val(0) },              // 9  ← Uint32
		func(Native) Native { return UintVal(0) },                // 10 ← Uint
		func(Native) Native { return Flt32Val(0.0) },             // 12 ← Flt32
		func(Native) Native { return FltVal(0.0) },               // 13 ← Float
		func(Native) Native { return FltVal(0.0).BigFlt() },      // 14 ← BigFlt
		func(Native) Native { return IntVal(0).Ratio() },         // 15 ← Ratio
		func(Native) Native { return IntVal(0).Imag().Imag64() }, // 16 ← Imag64
		func(Native) Native { return IntVal(0).Imag() },          // 17 ← Imag
		func(Native) Native { return TimeVal{} },                 // 18 ← Time
		func(arg Native) Native { return DuraVal(0) },            // 19 ← Duration
		func(arg Native) Native { return ByteVal(0) },            // 20 ← Byte
		func(arg Native) Native { return RuneVal(0) },            // 21 ← Rune
		func(arg Native) Native { return BitFlag(0) },            // 22 ← Flag
		func(arg Native) Native { return StrVal("") },            // 23 ← String
		func(arg Native) Native { return BytesVal{} },            // 24 ← Bytes
		func(arg Native) Native { return ErrorVal{} },            // 25 ← Error
	},
}

//
func Precedence(a, b Native) (x, y Native) {
	// if arguments types happend to be different‥.
	if at, bt := a.Type(), b.Type(); at != bt {
		var ati, bti = at.Flag().Index(), bt.Flag().Index()
		if ati > bti {
			// return unchanged a and cast b as type of a
			return a, TypeConversionTable[bti][ati](b)
		}
		// cast a as type of b and return unchanged b
		return TypeConversionTable[ati][bti](a), b
	}
	// ‥.otherwise just return both arguments unchagend
	return a, b
}

// BOOL VALUE
func (v BoolVal) Ratio() *RatioVal     { return v.IntVal().Ratio() }
func (v BoolVal) GoRat() *big.Rat      { return (*big.Rat)(v.Ratio()) }
func (v BoolVal) GoBigInt() *big.Int   { return big.NewInt(int64(v.Int())) }
func (v BoolVal) GoBigFlt() *big.Float { return big.NewFloat(v.GoFlt()) }
func (v BoolVal) BigInt() *BigIntVal   { return (*BigIntVal)(v.GoBigInt()) }
func (v BoolVal) BigFlt() *BigFltVal   { return (*BigFltVal)(v.GoBigFlt()) }
func (v BoolVal) Unit() Native         { return BoolVal(true) }
func (v BoolVal) Int() IntVal          { return IntVal(v.GoInt()) }
func (v BoolVal) Uint() UintVal        { return UintVal(uint(v.Int())) }
func (v BoolVal) Bool() bool           { return bool(v) }
func (v BoolVal) Idx() int             { return int(v.Int()) }
func (v BoolVal) GoFlt() float64       { return float64(v.Float()) }
func (v BoolVal) GoImag() complex128   { return complex128(v.Imag()) }
func (v BoolVal) GoUint() uint {
	if v {
		return 1
	}
	return 0
}
func (v BoolVal) GoInt() int {
	if v {
		return 1
	}
	return -1
}
func (v BoolVal) IntVal() IntVal {
	if v {
		return IntVal(1)
	}
	return IntVal(-1)
}
func (v BoolVal) Float() FltVal           { return v.IntVal().Float() }
func (v BoolVal) Imag() ImagVal           { return v.IntVal().Imag() }
func (v BoolVal) Or(arg BoolVal) BoolVal  { return v || arg }
func (v BoolVal) And(arg BoolVal) BoolVal { return v && arg }
func (v BoolVal) Not() BoolVal            { return !v }

// NATURAL VALUE
func (v UintVal) Idx() int                      { return int(v.Int()) }
func (v UintVal) GoInt() int                    { return int(v.Int()) }
func (v UintVal) GoUint() uint                  { return uint(v.Uint()) }
func (v UintVal) GoFlt() float64                { return float64(v.Float()) }
func (v UintVal) GoImag() complex128            { return complex128(v.Imag()) }
func (v UintVal) GoRat() *big.Rat               { return (*big.Rat)(v.Ratio()) }
func (v UintVal) GoBigInt() *big.Int            { return big.NewInt(int64(v.Int())) }
func (v UintVal) GoBigFlt() *big.Float          { return big.NewFloat(v.GoFlt()) }
func (v UintVal) BigInt() *BigIntVal            { return (*BigIntVal)(v.GoBigInt()) }
func (v UintVal) BigFlt() *BigFltVal            { return (*BigFltVal)(v.GoBigFlt()) }
func (v UintVal) Unit() Native                  { return UintVal(1) }
func (v UintVal) Uint() UintVal                 { return v }
func (v UintVal) Uint8() Uint8Val               { return Uint8Val(uint8(v)) }
func (v UintVal) Uint16() Uint16Val             { return Uint16Val(uint16(v)) }
func (v UintVal) Uint32() Uint32Val             { return Uint32Val(uint32(v)) }
func (v UintVal) Int() IntVal                   { return IntVal(int(v)) }
func (v UintVal) IntVal() IntVal                { return IntVal(int(v)) }
func (v UintVal) Bool() BoolVal                 { return v.BoolVal().(BoolVal) }
func (v UintVal) Float() FltVal                 { return FltVal(float64(v)) }
func (v UintVal) Imag() ImagVal                 { return v.IntVal().Imag() }
func (v UintVal) And(arg UintVal) UintVal       { return v & arg }
func (v UintVal) Xor(arg UintVal) UintVal       { return v ^ arg }
func (v UintVal) Or(arg UintVal) UintVal        { return v | arg }
func (v UintVal) AndNot(arg UintVal) UintVal    { return v &^ arg }
func (v UintVal) Not() UintVal                  { return ^v }
func (v UintVal) Negate() UintVal               { return -v }
func (v UintVal) Add(arg UintVal) UintVal       { return v + arg }
func (v UintVal) Substract(arg UintVal) UintVal { return v - arg }
func (v UintVal) Multipy(arg UintVal) UintVal   { return v * arg }
func (v UintVal) Power(arg UintVal) UintVal     { return v ^ arg }
func (v UintVal) Divide(arg UintVal) UintVal    { return v / arg }
func (v UintVal) DivideRatio(arg UintVal) *RatioVal {
	return (*RatioVal)(big.NewRat(int64(v), int64(arg)))
}
func (v UintVal) Ratio() *RatioVal {
	var rat = big.NewRat(int64(v), 1)
	return (*RatioVal)(rat)
}
func (v UintVal) BoolVal() Native {
	if v > 0 {
		return BoolVal(true)
	}
	return BoolVal(false)
}

// INTEGER VALUE
func (v IntVal) GoInt() int                  { return int(v) }
func (v IntVal) GoFlt() float64              { return float64(v) }
func (v IntVal) GoUint() uint                { return uint(v) }
func (v IntVal) GoImag() complex128          { return complex128(v.Imag()) }
func (v IntVal) GoRat() *big.Rat             { return (*big.Rat)(v.Ratio()) }
func (v IntVal) GoBigInt() *big.Int          { return big.NewInt(int64(v)) }
func (v IntVal) GoBigFlt() *big.Float        { return big.NewFloat(float64(v)) }
func (v IntVal) BigInt() *BigIntVal          { return (*BigIntVal)(v.GoBigInt()) }
func (v IntVal) BigFlt() *BigFltVal          { return (*BigFltVal)(v.GoBigFlt()) }
func (v IntVal) Unit() Native                { return UintVal(uint(v)) }
func (v IntVal) Byte() Native                { return IntVal(byte(v.Int8())) }
func (v IntVal) Rune() Native                { return RuneVal(rune(v.Int32())) }
func (v IntVal) Flag() Native                { return BitFlag(uint(v)) }
func (v IntVal) IntVal() IntVal              { return v }
func (v IntVal) Int8() Int8Val               { return Int8Val(int8(v)) }
func (v IntVal) Int16() Int16Val             { return Int16Val(int16(v)) }
func (v IntVal) Int32() Int32Val             { return Int32Val(int32(v)) }
func (v IntVal) Int() IntVal                 { return v }
func (v IntVal) Float() FltVal               { return FltVal(float64(v)) }
func (v IntVal) Imag() ImagVal               { return ImagVal(complex(v.Float(), 1.0)) }
func (v IntVal) Idx() int                    { return int(v) }
func (v IntVal) And(arg IntVal) IntVal       { return v & arg }
func (v IntVal) Xor(arg IntVal) IntVal       { return v ^ arg }
func (v IntVal) Or(arg IntVal) IntVal        { return v | arg }
func (v IntVal) AndNot(arg IntVal) IntVal    { return v &^ arg }
func (v IntVal) Not() IntVal                 { return ^v }
func (v IntVal) Negate() IntVal              { return -v }
func (v IntVal) Add(arg IntVal) IntVal       { return v + arg }
func (v IntVal) Substract(arg IntVal) IntVal { return v - arg }
func (v IntVal) Multipy(arg IntVal) IntVal   { return v * arg }
func (v IntVal) Power(arg IntVal) IntVal     { return v ^ arg }
func (v IntVal) Divide(arg IntVal) IntVal    { return v / arg }
func (v IntVal) Ratio() *RatioVal {
	var rat = big.NewRat(1, int64(v))
	return (*RatioVal)(rat)
}
func (v IntVal) Bool() BoolVal {
	if v > 0 {
		return BoolVal(true)
	}
	return BoolVal(false)
}
func (v IntVal) Truth() Native {
	if v < 0 {
		return BoolVal(false)
	}
	if v > 0 {
		return BoolVal(true)
	}
	return NilVal{}
}
func (v IntVal) Uint() UintVal {
	if v < 0 {
		return UintVal(v * -1)
	}
	return UintVal(v)
}

// REAL VALUE
func (v FltVal) Unit() Native                { return FltVal(1.0) }
func (v FltVal) Idx() int                    { return int(v) }
func (v FltVal) GoInt() int                  { return int(v) }
func (v FltVal) GoUint() uint                { return uint(v) }
func (v FltVal) GoFlt() float64              { return float64(v) }
func (v FltVal) GoImag() complex128          { return complex128(v.Imag()) }
func (v FltVal) GoRat() *big.Rat             { return (*big.Rat)(v.Ratio()) }
func (v FltVal) GoBigInt() *big.Int          { return big.NewInt(int64(v.GoInt())) }
func (v FltVal) GoBigFlt() *big.Float        { return big.NewFloat(v.GoFlt()) }
func (v FltVal) BigInt() *BigIntVal          { return (*BigIntVal)(v.GoBigInt()) }
func (v FltVal) BigFlt() *BigFltVal          { return (*BigFltVal)(v.GoBigFlt()) }
func (v FltVal) Uint() UintVal               { return UintVal(uint(v)) }
func (v FltVal) Int() IntVal                 { return IntVal(int(v)) }
func (v FltVal) Float() FltVal               { return v }
func (v FltVal) Flt32() Flt32Val             { return Flt32Val(float32(v)) }
func (v FltVal) Imag() ImagVal               { return ImagVal(complex(v, 1.0)) }
func (v FltVal) Negate() FltVal              { return -v }
func (v FltVal) Add(arg FltVal) FltVal       { return v + arg }
func (v FltVal) Substract(arg FltVal) FltVal { return v - arg }
func (v FltVal) Multipy(arg FltVal) FltVal   { return v * arg }
func (v FltVal) Divide(arg FltVal) FltVal    { return v / arg }
func (v FltVal) Ratio() *RatioVal {
	var rat = big.NewRat(int64(1), int64(1)).SetFloat64(v.GoFlt())
	return (*RatioVal)(rat)
}
func (v FltVal) Bool() BoolVal {
	if v > 0.0 {
		return BoolVal(true)
	}
	return BoolVal(false)
}
func (v FltVal) Truth() Native {
	if v < 0.0 {
		return BoolVal(false)
	}
	if v > 0.0 {
		return BoolVal(true)
	}
	return NilVal{}
}

// RATIONAL VALUE
func (v RatioVal) Idx() int               { return int(v.Int()) }
func (v RatioVal) GoInt() int             { return int(v.Int()) }
func (v RatioVal) GoUint() uint           { return uint(v.Uint()) }
func (v RatioVal) GoFlt() float64         { return float64(v.Float()) }
func (v RatioVal) GoImag() complex128     { return complex128(v.Imag()) }
func (v RatioVal) GoRat() *big.Rat        { return (*big.Rat)(&v) }
func (v RatioVal) GoBigInt() *big.Int     { return big.NewInt(int64(v.GoInt())) }
func (v RatioVal) GoBigFlt() *big.Float   { return big.NewFloat(v.GoFlt()) }
func (v RatioVal) BigInt() *BigIntVal     { return (*BigIntVal)(v.GoBigInt()) }
func (v RatioVal) BigFlt() *BigFltVal     { return (*BigFltVal)(v.GoBigFlt()) }
func (v RatioVal) Unit() Native           { return RatioVal(*big.NewRat(1, 1)) }
func (v RatioVal) Uint() UintVal          { return UintVal(uint(v.Int())) }
func (v RatioVal) Int() IntVal            { var num, _ = v.Rat().Float64(); return IntVal(int(num)) }
func (v RatioVal) Float() FltVal          { var flt, _ = v.Rat().Float64(); return FltVal(flt) }
func (v RatioVal) Rat() *big.Rat          { return (*big.Rat)(&v) }
func (v RatioVal) Imag() ImagVal          { return ImagVal(complex(v.Float(), 1.0)) }
func (v RatioVal) Numerator() IntVal      { return IntVal(int(v.Rat().Num().Int64())) }
func (v RatioVal) Denominator() IntVal    { return IntVal(int(v.Rat().Denom().Int64())) }
func (v RatioVal) Both() (Native, Native) { return IntVal(v.Numerator()), IntVal(v.Denominator()) }
func (v RatioVal) Left() Native           { return IntVal(v.Numerator()) }
func (v RatioVal) Right() Native          { return IntVal(v.Denominator()) }
func (v RatioVal) BothInt() (IntVal, IntVal) {
	return IntVal(int(v.Rat().Num().Int64())), IntVal(int(v.Rat().Denom().Int64()))
}
func (v RatioVal) Bool() BoolVal {
	if v.Int() > 0 {
		return BoolVal(true)
	}
	return BoolVal(false)
}
func (v RatioVal) Truth() Native {
	if v.Int() > 0 {
		return BoolVal(true)
	}
	if v.Int() < 0 {
		return BoolVal(true)
	}
	return NilVal{}
}

// IMAGINARY VALUE
func (v ImagVal) Idx() int                    { return int(v.Int()) }
func (v ImagVal) GoInt() int                  { return int(v.Int()) }
func (v ImagVal) GoUint() uint                { return uint(v.Uint()) }
func (v ImagVal) GoFlt() float64              { return float64(v.Float()) }
func (v ImagVal) GoImag() complex128          { return complex128(v.Imag()) }
func (v ImagVal) GoRat() *big.Rat             { return (*big.Rat)(v.Ratio()) }
func (v ImagVal) GoBigInt() *big.Int          { return big.NewInt(int64(v.GoInt())) }
func (v ImagVal) GoBigFlt() *big.Float        { return big.NewFloat(v.GoFlt()) }
func (v ImagVal) BigInt() *BigIntVal          { return (*BigIntVal)(v.GoBigInt()) }
func (v ImagVal) BigFlt() *BigFltVal          { return (*BigFltVal)(v.GoBigFlt()) }
func (v ImagVal) Unit() Native                { return ImagVal(complex(0, 0)) }
func (v ImagVal) Uint() UintVal               { return UintVal(uint(real(v))) }
func (v ImagVal) Int() IntVal                 { return IntVal(int(real(v))) }
func (v ImagVal) IntVal() IntVal              { return IntVal(real(v)) }
func (v ImagVal) Float() FltVal               { return FltVal(float64(real(v))) }
func (v ImagVal) Ratio() *big.Rat             { return big.NewRat(int64(real(v)), int64(imag(v))) }
func (v ImagVal) Imag() ImagVal               { return v }
func (v ImagVal) Imag64() Imag64Val           { return Imag64Val(complex64(v)) }
func (v ImagVal) Imaginary() FltVal           { return FltVal(imag(v)) }
func (v ImagVal) Real() FltVal                { return FltVal(real(v)) }
func (v ImagVal) BothFloat() (FltVal, FltVal) { return FltVal(real(v)), FltVal(imag(v)) }
func (v ImagVal) Both() (Native, Native)      { return FltVal(real(v)), FltVal(imag(v)) }
func (v ImagVal) Left() Native                { return FltVal(v.Real()) }
func (v ImagVal) Right() Native               { return FltVal(v.Imaginary()) }
func (v ImagVal) Bool() BoolVal {
	if real(v) > 0 {
		return BoolVal(true)
	}
	return BoolVal(false)
}

/// BIG INT VALUE
func (v *BigIntVal) Int64() int64                     { return v.Int64() }
func (v BigIntVal) Idx() int                          { return int(v.Int()) }
func (v BigIntVal) GoInt() int                        { return int(v.Int()) }
func (v BigIntVal) GoUint() uint                      { return uint(v.Uint()) }
func (v BigIntVal) GoFlt() float64                    { return float64(v.Int64()) }
func (v BigIntVal) GoImag() complex128                { return complex128(v.Imag()) }
func (v BigIntVal) GoRat() *big.Rat                   { return (*big.Rat)(v.Ratio()) }
func (v BigIntVal) GoBigInt() *big.Int                { return (*big.Int)(&v) }
func (v BigIntVal) GoBigFlt() *big.Float              { return big.NewFloat(v.GoFlt()) }
func (v BigIntVal) BigInt() *BigIntVal                { return (*BigIntVal)(v.GoBigInt()) }
func (v BigIntVal) BigFlt() *BigFltVal                { return (*BigFltVal)(v.GoBigFlt()) }
func (v BigIntVal) Bool() BoolVal                     { return IntVal(v.Int()).Bool() }
func (v BigIntVal) Int() IntVal                       { return IntVal(int(v.Int64())) }
func (v BigIntVal) Uint() UintVal                     { return UintVal(uint(v.GoBigInt().Uint64())) }
func (v BigIntVal) Float() FltVal                     { return FltVal(float64(v.GoFlt())) }
func (v BigIntVal) Ratio() *RatioVal                  { return IntVal(v.Int()).Ratio() }
func (v BigIntVal) Imag() ImagVal                     { return IntVal(v.Int()).Imag() }
func (v BigIntVal) Add(arg BigIntVal) BigIntVal       { return v.Add(arg) }
func (v BigIntVal) Substract(arg BigIntVal) BigIntVal { return v.Substract(arg) }
func (v BigIntVal) Multipy(arg BigIntVal) BigIntVal   { return v.Multipy(arg) }
func (v BigIntVal) Divide(arg BigIntVal) BigIntVal    { return v.Divide(arg) }
func (v BigIntVal) Power(arg BigIntVal) BigIntVal     { return v.Power(arg) }
func (v BigIntVal) And(arg BigIntVal) BigIntVal       { return v.And(arg) }
func (v BigIntVal) Xor(arg BigIntVal) BigIntVal       { return v.Xor(arg) }
func (v BigIntVal) Or(arg BigIntVal) BigIntVal        { return v.Or(arg) }
func (v BigIntVal) AndNot(arg BigIntVal) BigIntVal    { return v.AndNot(arg) }
func (v BigIntVal) Not() BigIntVal                    { return v.Not() }

/// BIG FLOAT VALUE
func (v *BigFltVal) Float64() int64                   { return v.Float64() }
func (v BigFltVal) Idx() int                          { return int(v.Int()) }
func (v BigFltVal) GoInt() int                        { return int(v.Int()) }
func (v BigFltVal) GoUint() uint                      { return uint(v.Uint()) }
func (v BigFltVal) GoFlt() float64                    { return float64(v.BigFlt().Float64()) }
func (v BigFltVal) GoImag() complex128                { return complex128(v.Imag()) }
func (v BigFltVal) GoRat() *big.Rat                   { return (*big.Rat)(v.Ratio()) }
func (v BigFltVal) GoBigInt() *big.Int                { return big.NewInt(int64(v.Int())) }
func (v BigFltVal) GoBigFlt() *big.Float              { return (*big.Float)(&v) }
func (v BigFltVal) BigInt() *BigIntVal                { return (*BigIntVal)(v.GoBigInt()) }
func (v BigFltVal) BigFlt() *BigFltVal                { return (*BigFltVal)(v.GoBigFlt()) }
func (v BigFltVal) Bool() BoolVal                     { return IntVal(v.Int()).Bool() }
func (v BigFltVal) Int() IntVal                       { return IntVal(int(v.GoBigInt().Int64())) }
func (v BigFltVal) Uint() UintVal                     { return UintVal(uint(v.GoBigInt().Uint64())) }
func (v BigFltVal) Float() FltVal                     { return FltVal(float64(v.Float64())) }
func (v BigFltVal) Ratio() *RatioVal                  { return IntVal(v.Int()).Ratio() }
func (v BigFltVal) Imag() ImagVal                     { return IntVal(v.Int()).Imag() }
func (v BigFltVal) Add(arg BigFltVal) BigFltVal       { return v.Add(arg) }
func (v BigFltVal) Substract(arg BigFltVal) BigFltVal { return v.Substract(arg) }
func (v BigFltVal) Multipy(arg BigFltVal) BigFltVal   { return v.Multipy(arg) }
func (v BigFltVal) Divide(arg BigFltVal) BigFltVal    { return v.Divide(arg) }

/// TIME VALUE
func (v TimeVal) Idx() int                      { return int(v.Int()) }
func (v TimeVal) GoInt() int                    { return int(v.Int()) }
func (v TimeVal) GoUint() uint                  { return uint(v.Uint()) }
func (v TimeVal) GoFlt() float64                { return float64(v.Float()) }
func (v TimeVal) GoImag() complex128            { return complex128(v.Imag()) }
func (v TimeVal) GoRat() *big.Rat               { return (*big.Rat)(v.Ratio()) }
func (v TimeVal) GoBigInt() *big.Int            { return big.NewInt(int64(v.GoInt())) }
func (v TimeVal) GoBigFlt() *big.Float          { return big.NewFloat(v.GoFlt()) }
func (v TimeVal) BigInt() *BigIntVal            { return (*BigIntVal)(v.GoBigInt()) }
func (v TimeVal) BigFlt() *BigFltVal            { return (*BigFltVal)(v.GoBigFlt()) }
func (v TimeVal) Time() time.Time               { return time.Time(v) }
func (v TimeVal) Uint() UintVal                 { return UintVal(uint(time.Time(v).Unix())) }
func (v TimeVal) UintVal() UintVal              { return UintVal(uint(time.Time(v).Unix())) }
func (v TimeVal) Int() IntVal                   { return IntVal(int(time.Time(v).Unix())) }
func (v TimeVal) IntVal() IntVal                { return IntVal(time.Time(v).Unix()) }
func (v TimeVal) Bool() BoolVal                 { return IntVal(v.Int()).Bool() }
func (v TimeVal) Ratio() *RatioVal              { return IntVal(v.Int()).Ratio() }
func (v TimeVal) Float() FltVal                 { return IntVal(v.Int()).Float() }
func (v TimeVal) Imag() ImagVal                 { return IntVal(v.Int()).Imag() }
func (v TimeVal) ANSIC() StrVal                 { return StrVal(time.ANSIC) }
func (v TimeVal) Add(arg DuraVal) TimeVal       { return v.Add(arg) }
func (v TimeVal) Substract(arg DuraVal) TimeVal { return v.Substract(arg) }

/// DURATION VALUE
func (v DuraVal) Idx() int                { return int(v.Int()) }
func (v DuraVal) GoInt() int              { return int(v.Int()) }
func (v DuraVal) GoUint() uint            { return uint(v.Uint()) }
func (v DuraVal) GoFlt() float64          { return float64(v.Float()) }
func (v DuraVal) GoImag() complex128      { return complex128(v.Imag()) }
func (v DuraVal) GoRat() *big.Rat         { return (*big.Rat)(v.Ratio()) }
func (v DuraVal) GoBigInt() *big.Int      { return big.NewInt(int64(v.GoInt())) }
func (v DuraVal) GoBigFlt() *big.Float    { return big.NewFloat(v.GoFlt()) }
func (v DuraVal) BigInt() *BigIntVal      { return (*BigIntVal)(v.GoBigInt()) }
func (v DuraVal) BigFlt() *BigFltVal      { return (*BigFltVal)(v.GoBigFlt()) }
func (v DuraVal) Duration() time.Duration { return time.Duration(v) }
func (v DuraVal) Uint() UintVal           { return UintVal(uint(v)) }
func (v DuraVal) UintVal() UintVal        { return UintVal(v.Uint()) }
func (v DuraVal) Int() IntVal             { return IntVal(int(v)) }
func (v DuraVal) IntVal() IntVal          { return IntVal(v.Int()) }
func (v DuraVal) Bool() BoolVal           { return IntVal(v.Int()).Bool() }
func (v DuraVal) Ratio() *RatioVal        { return IntVal(v.Int()).Ratio() }
func (v DuraVal) Float() FltVal           { return IntVal(v.Int()).Float() }
func (v DuraVal) Imag() ImagVal           { return IntVal(v.Int()).Imag() }

/// BYTE VALUE
func (v ByteVal) Bool() bool {
	if v > ByteVal(0) {
		return true
	}
	return false
}
func (v ByteVal) Idx() int                    { return int(v.Int()) }
func (v ByteVal) String() string              { return string(v.Bytes()) }
func (v ByteVal) GoByte() byte                { return byte(v) }
func (v ByteVal) GoInt() int                  { return int(v.Int()) }
func (v ByteVal) GoUint() uint                { return uint(v.Uint()) }
func (v ByteVal) GoFlt() float64              { return float64(v.Float()) }
func (v ByteVal) GoImag() complex128          { return complex128(v.Imag()) }
func (v ByteVal) GoRat() *big.Rat             { return (*big.Rat)(v.Ratio()) }
func (v ByteVal) GoBigInt() *big.Int          { return big.NewInt(int64(v.GoInt())) }
func (v ByteVal) GoBigFlt() *big.Float        { return big.NewFloat(v.GoFlt()) }
func (v ByteVal) BigInt() *BigIntVal          { return (*BigIntVal)(v.GoBigInt()) }
func (v ByteVal) BigFlt() *BigFltVal          { return (*BigFltVal)(v.GoBigFlt()) }
func (v ByteVal) Bytes() BytesVal             { return BytesVal([]byte{v.GoByte()}) }
func (v ByteVal) Unit() Native                { return ByteVal(byte(0)) }
func (v ByteVal) Uint() UintVal               { return UintVal(uint(v)) }
func (v ByteVal) Int() IntVal                 { return IntVal(int(v)) }
func (v ByteVal) Ratio() *RatioVal            { return IntVal(int(v)).Ratio() }
func (v ByteVal) Float() FltVal               { return FltVal(float64(v)) }
func (v ByteVal) Imag() ImagVal               { return IntVal(int(v)).Imag() }
func (v ByteVal) Byte() ByteVal               { return ByteVal(byte(v)) }
func (v ByteVal) Rune() RuneVal               { return RuneVal(rune(v.Byte())) }
func (v ByteVal) Len() IntVal                 { return IntVal(bits.Len8(uint8(v.Uint()))) }
func (v ByteVal) Concat(arg ByteVal) BytesVal { return BytesVal([]byte{byte(v), byte(arg)}) }
func (v ByteVal) And(arg ByteVal) ByteVal     { return v & arg }
func (v ByteVal) Xor(arg ByteVal) ByteVal     { return v ^ arg }
func (v ByteVal) Or(arg ByteVal) ByteVal      { return v | arg }
func (v ByteVal) AndNot(arg ByteVal) ByteVal  { return v &^ arg }
func (v ByteVal) Not() ByteVal                { return ^v }

/// BYTE SLICE VALUE
func (v BytesVal) String() string            { return string(v) }
func (v BytesVal) GoBytes() []byte           { return []byte(v) }
func (v BytesVal) GoRunes() []rune           { return []rune(v.String()) }
func (v BytesVal) ByteBuffer() *bytes.Buffer { return bytes.NewBuffer(v) }
func (v BytesVal) ByteReader() io.ByteReader { return bytes.NewReader(v) }
func (v BytesVal) StrVal() StrVal            { return StrVal(v.String()) }
func (v BytesVal) Unit() BytesVal            { return BytesVal([]byte{byte(0)}) }
func (v BytesVal) Bytes() ByteVec            { return ByteVec(v) }
func (v BytesVal) RuneVec() RuneVec          { return RuneVec(v.GoRunes()) }
func (v BytesVal) Len() IntVal               { return IntVal(len(v.Bytes())) }
func (v BytesVal) Concat(args ...ByteVal) BytesVal {
	for _, arg := range args {
		v = append(v, byte(arg))
	}
	return v
}
func (v BytesVal) UintNative() Native {
	u, err := binary.ReadUvarint(v.ByteReader())
	if err != nil {
		return NewError(err)
	}
	return UintVal(u)
}
func (v BytesVal) IntNative() Native {
	i, err := binary.ReadVarint(v.ByteReader())
	if err != nil {
		return NewError(err)
	}
	return IntVal(i)
}
func (v BytesVal) Bool() BoolVal {
	for _, b := range v {
		if b > byte(0) {
			return BoolVal(true)
		}
	}
	return BoolVal(false)
}

/// STRING VALUE
func (v StrVal) String() string                  { return string(v) }
func (v StrVal) StringBuffer() *strings.Reader   { return strings.NewReader(v.String()) }
func (v StrVal) Unit() Native                    { return StrVal(" ") }
func (v StrVal) Runes() RuneVec                  { return RuneVec([]rune(string(v))) }
func (v StrVal) Len() IntVal                     { return IntVal(int(len(string(v)))) }
func (v StrVal) Bytes() BytesVal                 { return []byte(string(v)) }
func (v StrVal) DurationNative() Native          { return v.ReadDuraVal() }
func (v StrVal) TimeNative(layout string) Native { return v.ReadTimeVal(layout) }
func (v StrVal) Concat(arg StrVal) StrVal        { return StrVal(string(v) + string(arg)) }
func (v StrVal) Split(arg StrVal) (val []StrVal) {
	var slice = strings.Split(string(v), string(arg))
	val = make([]StrVal, 0, len(slice))
	for _, str := range slice {
		val = append(val, StrVal(str))
	}
	return val
}
func (v StrVal) NumberNative() Native {
	if _, err := v.ReadBool(); err == nil {
		return v.ReadBoolVal()
	}
	if _, err := v.ReadUint(); err == nil {
		return v.ReadUintVal()
	}
	if _, err := v.ReadInt(); err == nil {
		return v.ReadIntVal()
	}
	if _, err := v.ReadFloat(); err == nil {
		return v.ReadFloatVal()
	}
	return NewNil()
}
func (v StrVec) Join(arg StrVal) (val []StrVal) {
	val = make([]StrVal, 0, len(v))
	for _, str := range v {
		val = append(val, StrVal(str))
	}
	return val
}
