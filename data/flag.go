package data

import (
	"fmt"
	"math/bits"
)

//// BOUND TYPE FLAG METHODS ////
func (v BitFlag) FlagType() Uint8Val   { return 0 }
func (v BitFlag) Flag() BitFlag        { return v }
func (v BitFlag) Uint() uint           { return uint(v) }
func (v BitFlag) Int() int             { return int(v) }
func (v BitFlag) Len() int             { return FlagLength(v) }
func (v BitFlag) Count() int           { return FlagCount(v) }
func (v BitFlag) Least() int           { return FlagLeastSig(v) }
func (v BitFlag) Most() int            { return FlagMostSig(v) }
func (v BitFlag) TypeName() string     { return v.String() }
func (v BitFlag) Low(f Typed) Typed    { return FlagLow(f).Flag() }
func (v BitFlag) High(f Typed) Typed   { return FlagHigh(f).Flag() }
func (v BitFlag) Reverse() Typed       { return FlagReverse(v).Flag() }
func (v BitFlag) Rotate(n int) Typed   { return FlagRotate(v, n).Flag() }
func (v BitFlag) Toggle(f Typed) Typed { return FlagToggle(v, f).Flag() }
func (v BitFlag) Concat(f Typed) Typed { return FlagConcat(v, f).Flag() }
func (v BitFlag) Mask(f Typed) Typed   { return FlagMask(v, f).Flag() }
func (v BitFlag) Match(f Typed) bool   { return FlagMatch(v, f) }
func (v BitFlag) Decompose() []Typed   { return FlagDecompose(v) }

///// FREE TYPE FLAG METHOD IMPLEMENTATIONS /////
func flag(t Typed) BitFlag              { return t.Flag() }
func FlagLength(t Typed) int            { return bits.Len(t.Flag().Uint()) }
func FlagCount(t Typed) int             { return bits.OnesCount(t.Flag().Uint()) }
func FlagLeastSig(t Typed) int          { return bits.TrailingZeros(t.Flag().Uint()) + 1 }
func FlagMostSig(t Typed) int           { return bits.LeadingZeros(t.Flag().Uint()) - 1 }
func FlagReverse(t Typed) BitFlag       { return BitFlag(bits.Reverse(t.Flag().Uint())) }
func FlagRotate(t Typed, n int) BitFlag { return BitFlag(bits.RotateLeft(t.Flag().Uint(), n)) }

func FlagToggle(t Typed, v Typed) BitFlag {
	return BitFlag(t.Flag().Uint() ^ v.Flag().Uint())
}

func FlagConcat(t Typed, v Typed) BitFlag {
	return BitFlag(t.Flag().Uint() | v.Flag().Uint())
}

func FlagMask(t Typed, v Typed) BitFlag {
	return BitFlag(t.Flag().Uint() &^ v.Flag().Uint())
}

func FlagShow(f Typed) string { return fmt.Sprintf("%64b\n", f) }
func FlagLow(t Typed) Typed   { return FlagMask(t.Flag(), Typed(Type)) }
func FlagHigh(t Typed) BitFlag {
	len := FlagLength(BitFlag(Type))
	return FlagMask(FlagRotate(t.Flag(), len), FlagRotate(BitFlag(Type), len))
}

func FlagMatch(t Typed, v Typed) bool {
	if t.Flag().Uint()&^v.Flag().Uint() != 0 {
		return false
	}
	return true
}

// decomposes flag resulting from OR concatenation, to a slice of flags
func FlagDecompose(v Typed) []Typed {
	var slice = []Typed{}
	if bits.OnesCount(v.Flag().Uint()) == 1 {
		return append(slice, v.Flag())
	}
	var u = uint(1)
	var i = 0
	for i < 63 {
		if FlagMatch(BitFlag(u), v) {
			slice = append(slice, BitFlag(u))
		}
		i = i + 1
		u = uint(1) << uint(i)
	}
	return slice
}
