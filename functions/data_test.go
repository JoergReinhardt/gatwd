package functions

import (
	"fmt"
	"testing"

	d "github.com/joergreinhardt/gatwd/data"
)

func TestData(t *testing.T) {
	var nat = Box(d.UintVal(12))
	fmt.Printf("uint converted to native: %s, type: %s, typeFnc: %s, typeNat: %s \n",
		nat, nat.Type(), nat.TypeFnc(), nat.Type())

	var pair = Box(d.NewPair(d.New("key"), d.New(42)))
	fmt.Printf("key pair converted to native: %s, type: %s, len type: %d,"+
		" typeFnc: %s, typeNat: %s\n",
		pair, pair.Type(), pair.Type().Len(), pair.TypeFnc(), pair.Type())

	var nest = Box(d.NewPair(d.StrVal("key"),
		d.NewPair(d.StrVal("inner key"), d.StrVal("Value"))))
	fmt.Printf("nested pair converted to native: %s, type: %s, typeFnc: %s typeNat: %s\n",
		nest, nest.Type(), nest.TypeFnc(), nest.Type())
}
