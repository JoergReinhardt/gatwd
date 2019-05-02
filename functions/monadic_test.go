package functions

import (
	"fmt"
	"testing"

	d "github.com/joergreinhardt/gatwd/data"
)

var isInteger = NewTruth(func(arg Callable) bool { return d.Integers.Flag().Match(arg.TypeNat()) })

var isFloat = NewTruth(func(arg Callable) bool { return d.Rationals.Flag().Match(arg.TypeNat()) })

func TestParseNativeType(t *testing.T) {
	var sig = parseNat(d.Map | d.String | d.Int)
	fmt.Printf("flag: %s\n", sig.Flag())
	fmt.Printf("sum: %s\n", sig.Sum())
	fmt.Printf("product: %s\n", sig.Product())
	var head, list = sig.Product()()
	fmt.Printf("product 0: %s\n", head)
	fmt.Printf("product 0 flag: %s\n", head.(TypeSig).Flag())
	fmt.Printf("product 0 sum: %s\n", head.(TypeSig).Sum())
	fmt.Printf("product 0 prod: %s\n", head.(TypeSig).Product())
	head, list = list()
	fmt.Printf("product 1: %s\n", head)
	fmt.Printf("product 1 flag: %s\n", head.(TypeSig).Flag())
	fmt.Printf("product 1 sum: %s\n", head.(TypeSig).Sum())
	fmt.Printf("product 1 prod: %s\n", head.(TypeSig).Product())
}
func TestParseFunctionalType(t *testing.T) {
	var sig = parseFnc(List | Data | Monad | Maybe | Number)
	fmt.Printf("flag: %s\n", sig.Flag())
	fmt.Printf("sum: %s\n", sig.Sum())
	fmt.Printf("product: %s\n", sig.Product())
	var head, list = sig.Product()()
	fmt.Printf("product 0: %s\n", head)
	fmt.Printf("product 0 flag: %s\n", head.(TypeSig).Flag())
	fmt.Printf("product 0 sum: %s\n", head.(TypeSig).Sum())
	fmt.Printf("product 0 prod: %s\n", head.(TypeSig).Product())
	head, list = list()
	fmt.Printf("product 1: %s\n", head)
	fmt.Printf("product 1 flag: %s\n", head.(TypeSig).Flag())
	fmt.Printf("product 1 sum: %s\n", head.(TypeSig).Sum())
	fmt.Printf("product 1 prod: %s\n", head.(TypeSig).Product())
}
