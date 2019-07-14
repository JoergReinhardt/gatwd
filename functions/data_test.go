package functions

import (
	"fmt"
	"testing"

	d "github.com/joergreinhardt/gatwd/data"
)

func TestData(t *testing.T) {
	var nat = NewData(d.UintVal(12))
	fmt.Printf("uint converted to native: %s, typeFnc: %s typeNat: %s FlagType %s\n",
		nat, nat.TypeFnc(), nat.TypeNat(), TyFlag(nat.FlagType()))

	var pair = NewData(d.NewPair(NewData(d.StrVal("key")), NewData()))
	fmt.Printf("key pair converted to native: %s, typeFnc: %s typeNat: %s FlagType %s\n",
		pair, pair.TypeFnc(), pair.TypeNat(), TyFlag(pair.FlagType()))

	var nest = NewData(d.NewPair(NewData(d.StrVal("key")), NewData(d.NewPair(d.StrVal("inner key"), d.StrVal("Value")))))
	fmt.Printf("key pair converted to native: %s, typeFnc: %s typeNat: %s FlagType %s\n",
		nest, nest.TypeFnc(), nest.TypeNat(), TyFlag(nest.FlagType()))
}
func TestDeclaration(t *testing.T) {
}
func TestFunction(t *testing.T) {
}
func TestNativeFunction(t *testing.T) {
}
func TestDataFunction(t *testing.T) {
}
