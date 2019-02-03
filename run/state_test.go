package run

import (
	"fmt"
	"testing"

	d "github.com/JoergReinhardt/godeep/data"
	p "github.com/JoergReinhardt/godeep/parse"
)

// generate a pattern for testFnc ∷ Integer → String → String
var pat = p.NewPattern("testFnc",
	p.NewDataTypeToken(d.String),
	p.NewDataTypeToken(d.Integer),
	p.NewDataTypeToken(d.String))

// implement a monoid for testFnc ∷ Integer → String → String
//var mon = p.NewMonoid(
//	pat,
//	f.NewValue(
//		d.StrVal(
//			dat[0].String()+
//				dat[1].String())))
//
//// define a polymorph based on the monoid
//var pol = p.NewPolymorph(pat, mon)
//
func TestState(t *testing.T) {
	//fmt.Println(pat.Signature())
	//	fmt.Println(mon)
	//	fmt.Println(pol)
}
