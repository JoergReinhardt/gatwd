package data

import (
	"fmt"
	"math/bits"
)

//// BOUND TYPE FLAG METHODS ////
func (v BitFlag) Uint() uint               { return uint(v) }
func (v BitFlag) Len() int                 { return FlagLength(v) }
func (v BitFlag) Count() int               { return FlagCount(v) }
func (v BitFlag) Least() int               { return FlagLeastSig(v) }
func (v BitFlag) Most() int                { return FlagMostSig(v) }
func (v BitFlag) Low(f BitFlag) BitFlag    { return FlagLow(f).Flag() }
func (v BitFlag) High(f BitFlag) BitFlag   { return FlagHigh(f).Flag() }
func (v BitFlag) Reverse() BitFlag         { return FlagReverse(v).Flag() }
func (v BitFlag) Rotate(n int) BitFlag     { return FlagRotate(v, n).Flag() }
func (v BitFlag) Toggle(f BitFlag) BitFlag { return FlagToggle(v, f).Flag() }
func (v BitFlag) Concat(f BitFlag) BitFlag { return FlagConcat(v, f).Flag() }
func (v BitFlag) Mask(f BitFlag) BitFlag   { return FlagMask(v, f).Flag() }
func (v BitFlag) Match(f BitFlag) bool     { return FlagMatch(v, f) }
func (v BitFlag) Decompose() []BitFlag     { return FlagDecompose(v) }

///// FREE TYPE FLAG METHOD IMPLEMENTATIONS /////
func flag(t Typed) BitFlag                { return t.Flag() }
func FlagLength(t Typed) int              { return bits.Len(t.Flag().Uint()) }
func FlagCount(t Typed) int               { return bits.OnesCount(t.Flag().Uint()) }
func FlagLeastSig(t Typed) int            { return bits.TrailingZeros(t.Flag().Uint()) + 1 }
func FlagMostSig(t Typed) int             { return bits.LeadingZeros(t.Flag().Uint()) - 1 }
func FlagReverse(t Typed) BitFlag         { return BitFlag(bits.Reverse(t.Flag().Uint())) }
func FlagRotate(t Typed, n int) BitFlag   { return BitFlag(bits.RotateLeft(t.Flag().Uint(), n)) }
func FlagToggle(t Typed, v Typed) BitFlag { return BitFlag(t.Flag().Uint() ^ v.Flag().Uint()) }
func FlagConcat(t Typed, v Typed) BitFlag { return BitFlag(t.Flag().Uint() | v.Flag().Uint()) }
func FlagMask(t Typed, v Typed) BitFlag   { return BitFlag(t.Flag().Uint() &^ v.Flag().Uint()) }
func FlagShow(f Typed) string             { return fmt.Sprintf("%64b\n", f) }
func FlagLow(t Typed) Typed               { return FlagMask(t.Flag(), Typed(Mask)) }
func FlagHigh(t BitFlag) BitFlag {
	len := FlagLength(BitFlag(Flag))
	return FlagMask(FlagRotate(t.Flag(), len), FlagRotate(BitFlag(Definition), len))
}
func FlagMatch(t BitFlag, v BitFlag) bool {
	if t.Uint()&^v.Uint() != 0 {
		return false
	}
	return true
}

// decomposes flag resulting from OR concatenation, to a slice of flags
func FlagDecompose(v BitFlag) []BitFlag {
	var slice = []BitFlag{}
	if bits.OnesCount(v.Uint()) == 1 {
		return append(slice, v)
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
