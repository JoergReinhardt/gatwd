package functions

import (
	"fmt"
	"testing"

	d "github.com/joergreinhardt/gatwd/data"
)

var isInteger = NewTruth(func(arg Callable) bool { return d.Integers.Flag().Match(arg.TypeNat()) })

var isFloat = NewTruth(func(arg Callable) bool { return d.Rationals.Flag().Match(arg.TypeNat()) })

func TestParseNativeType(t *testing.T) {
	var sig = parseNat((d.Map | d.Pair | d.String | d.Int).Flag())
	fmt.Printf("native sig: %s\n", sig)
}
func TestParseFunctionalType(t *testing.T) {
	var pair = NewPair(New("test"), New(0.999))
	var cons = NewTypeCons(pair, NewTypeSignatureFromExpr(pair))
	fmt.Printf("cons: %s\n", cons)
	fmt.Printf("expr: %s\n", cons.Expression())
	fmt.Printf("signature: %s\n", cons.Signature())
	fmt.Printf("sum: %s\n", cons.Signature().Sum())
	fmt.Printf("prod: %s\n", cons.Signature().Product())

	var list = NewList(
		New(0.1),
		New(0.2),
		New(0.2),
		New(0.3),
		New(0.4),
		New(0.5),
		New(0.5),
		New(0.6),
		New(0.7),
		New(0.8),
		New(0.9),
		New(0.10),
	)

	var lcon = NewTypeCons(list, NewTypeSignatureFromExpr(list))
	fmt.Printf("lcon: %s\n", lcon)
	fmt.Printf("expr: %s\n", lcon.Expression())
	fmt.Printf("signature: %s\n", lcon.Signature())
	fmt.Printf("sum: %s\n", lcon.Signature().Sum())
	fmt.Printf("prod: %s\n", lcon.Signature().Product())

}
