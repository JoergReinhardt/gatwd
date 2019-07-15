package functions

import (
	"fmt"
	"testing"

	d "github.com/joergreinhardt/gatwd/data"
)

func TestData(t *testing.T) {
	var nat = NewData(d.UintVal(12))
	fmt.Printf("uint converted to native: %s, typeFnc: %s typeNat: %s \n",
		nat, nat.TypeFnc(), nat.TypeNat())

	var pair = NewData(d.NewPair(New("key"), New(42)))
	fmt.Printf("key pair converted to native: %s, typeFnc: %s typeNat: %s\n",
		pair, pair.TypeFnc(), pair.TypeNat())

	var nest = NewData(d.NewPair(d.StrVal("key"), d.NewPair(d.StrVal("inner key"), d.StrVal("Value"))))
	fmt.Printf("nested pair converted to native: %s, typeFnc: %s typeNat: %s\n",
		nest, nest.TypeFnc(), nest.TypeNat())
}
func TestDeclaration(t *testing.T) {
}
func TestFunction(t *testing.T) {
}
func TestNativeFunction(t *testing.T) {
}
func TestDataFunction(t *testing.T) {
}
