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
func TestNativeFunction(t *testing.T) {
	var addInt = New(func(args ...d.Native) d.Native {
		var a, b = args[0].(d.Numeral), args[1].(d.Numeral)
		return a.Int() + b.Int()
	})
	fmt.Printf("2 + 2 = %s\n", addInt.Eval(d.IntVal(2), d.IntVal(2)))
}
