package types

import (
	"fmt"
	"math/bits"
)

//// BOUND TYPE FLAG METHODS ////
func (t flag) Uint() uint         { return fuint(t) }
func (t flag) Len() int           { return flen(t) }
func (t flag) Count() int         { return fcount(t) }
func (t flag) Least() int         { return fleast(t) }
func (t flag) Most() int          { return fmost(t) }
func (t flag) Low(v Typed) Typed  { return flow(t) }
func (t flag) High(v Typed) Typed { return fhigh(t) }
func (t flag) Reverse() flag      { return frev(t) }
func (t flag) Rotate(n int) flag  { return frot(t, n) }
func (t flag) Toggle(v flag) flag { return ftog(t, v) }
func (t flag) Concat(v flag) flag { return fconc(t, v) }
func (t flag) Mask(v flag) flag   { return fmask(t, v) }
func (t flag) Match(v Typed) bool { return fmatch(t, v) }

///// FREE TYPE FLAG METHOD IMPLEMENTATIONS /////
func fuint(t flag) uint         { return uint(t) }
func flen(t flag) int           { return bits.Len(uint(t)) }
func fcount(t flag) int         { return bits.OnesCount(uint(t)) }
func fleast(t flag) int         { return bits.TrailingZeros(uint(t)) + 1 }
func fmost(t flag) int          { return bits.LeadingZeros(uint(t)) - 1 }
func frev(t flag) flag          { return flag(bits.Reverse(uint(t))) }
func frot(t flag, n int) flag   { return flag(bits.RotateLeft(uint(t), n)) }
func ftog(t flag, v flag) flag  { return flag(uint(t) ^ v.Type().Uint()) }
func fconc(t flag, v flag) flag { return flag(uint(t) | v.Type().Uint()) }
func fmask(t flag, v flag) flag { return flag(uint(t) &^ v.Type().Uint()) }
func fshow(f Typed) string      { return fmt.Sprintf("%64b\n", f) }
func flow(t Typed) Typed        { return fmask(t.Type(), flag(Mask)) }
func fhigh(t Typed) Typed {
	len := flen(flag(Natives))
	return fmask(frot(t.Type(), len), frot(flag(Natives), len))
}
func fmatch(t flag, v Typed) bool {
	if t.Uint()&v.Type().Uint() != 0 {
		return true
	}
	return false
}
