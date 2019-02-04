package data

import (
	"fmt"
	"math/bits"
)

//// BOUND TYPE FLAG METHODS ////
func (v BitFlag) Flag() BitFlag            { return v }
func (v BitFlag) Uint() uint               { return uint(v) }
func (v BitFlag) Len() int                 { return FlagLength(v) }
func (v BitFlag) Count() int               { return FlagCount(v) }
func (v BitFlag) Least() int               { return FlagLeastSig(v) }
func (v BitFlag) Most() int                { return FlagMostSig(v) }
func (v BitFlag) Low(f BitFlag) BitFlag    { return FlagLow(f).TypePrim() }
func (v BitFlag) High(f BitFlag) BitFlag   { return FlagHigh(f).TypePrim() }
func (v BitFlag) Reverse() BitFlag         { return FlagReverse(v).TypePrim() }
func (v BitFlag) Rotate(n int) BitFlag     { return FlagRotate(v, n).TypePrim() }
func (v BitFlag) Toggle(f BitFlag) BitFlag { return FlagToggle(v, f).TypePrim() }
func (v BitFlag) Concat(f BitFlag) BitFlag { return FlagConcat(v, f).TypePrim() }
func (v BitFlag) Mask(f BitFlag) BitFlag   { return FlagMask(v, f).TypePrim() }
func (v BitFlag) Match(f BitFlag) bool     { return FlagMatch(v, f) }
func (v BitFlag) Decompose() []BitFlag     { return FlagDecompose(v) }

///// FREE TYPE FLAG METHOD IMPLEMENTATIONS /////
func flag(t Primary) BitFlag              { return t.TypePrim() }
func FlagLength(t Primary) int            { return bits.Len(t.TypePrim().Uint()) }
func FlagCount(t Primary) int             { return bits.OnesCount(t.TypePrim().Uint()) }
func FlagLeastSig(t Primary) int          { return bits.TrailingZeros(t.TypePrim().Uint()) + 1 }
func FlagMostSig(t Primary) int           { return bits.LeadingZeros(t.TypePrim().Uint()) - 1 }
func FlagReverse(t Primary) BitFlag       { return BitFlag(bits.Reverse(t.TypePrim().Uint())) }
func FlagRotate(t Primary, n int) BitFlag { return BitFlag(bits.RotateLeft(t.TypePrim().Uint(), n)) }
func FlagToggle(t Primary, v Primary) BitFlag {
	return BitFlag(t.TypePrim().Uint() ^ v.TypePrim().Uint())
}
func FlagConcat(t Primary, v Primary) BitFlag {
	return BitFlag(t.TypePrim().Uint() | v.TypePrim().Uint())
}
func FlagMask(t Primary, v Primary) BitFlag {
	return BitFlag(t.TypePrim().Uint() &^ v.TypePrim().Uint())
}
func FlagShow(f Primary) string { return fmt.Sprintf("%64b\n", f) }
func FlagLow(t Primary) Primary { return FlagMask(t.TypePrim(), Primary(Mask)) }
func FlagHigh(t BitFlag) BitFlag {
	len := FlagLength(BitFlag(Flag))
	return FlagMask(FlagRotate(t.TypePrim(), len), FlagRotate(BitFlag(HigherOrder), len))
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
