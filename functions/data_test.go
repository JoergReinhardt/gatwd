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
	var addInt = Declare(New(func(args ...d.Native) d.Native {
		if len(args) == 2 {
			return args[0].(d.Numeral).Int() + args[1].(d.Numeral).Int()
		}
		return d.IntVal(0)
	}).(DataExpr), "add Int", d.Int, d.Int, d.Int)

	fmt.Printf("2 + 2 = %s\n", addInt.Eval(d.IntVal(2), d.IntVal(2)))
	fmt.Printf("TypeNat(): %s, TypeFnc(): %s Type(): %s, TypeName(): %s\n",
		addInt.TypeNat(), addInt.TypeFnc(), addInt.Type(), addInt.TypeName())
}
