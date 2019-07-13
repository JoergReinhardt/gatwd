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
	var addIntNat = Declare(New(func(args ...d.Native) d.Native {
		if len(args) == 2 {
			return args[0].(d.Numeral).Int() + args[1].(d.Numeral).Int()
		}
		return d.IntVal(0)
	}), "add Int", d.Int, d.Int, d.Int)

	var add2 = addIntNat.Eval(d.IntVal(2)).(DeclaredExpr)
	fmt.Printf("addTwo ← addIntNat.Eval(2) ∷ %s\n", add2.Type())
	if Define("add2", d.Int, d.Int).Match(add2.Type()) {
		t.Fail()
	}
	var add22 = add2.Eval(d.IntVal(2))
	fmt.Printf("addTwo.Eval(2) = %s ∷ %s\n", add22, add22.TypeName())
	if !add22.Type().Match(d.Int) {
		t.Fail()
	}
	var add2Int = addIntNat.Eval(d.IntVal(2), d.IntVal(2))
	fmt.Printf("2 + 2 = %s ∷ %s\n", add2Int, add2Int.TypeName())
	if !add2Int.Type().Match(d.Int) {
		t.Fail()
	}
	var addNInt = addIntNat.Eval(d.IntVal(1), d.IntVal(2), d.IntVal(3), d.IntVal(4))
	fmt.Printf("1 + 2, 3 + 4 = %s ∷ %s\n", addNInt, addNInt.TypeName())
	fmt.Printf("TypeNat(): %s, TypeFnc(): %s Type(): %s, TypeName(): %s\n",
		addIntNat.TypeNat(), addIntNat.TypeFnc(), addIntNat.Type(), addIntNat.TypeName())
	if !addNInt.Type().Match(d.Int | d.Slice) {
		t.Fail()
	}
}
func TestDataFunction(t *testing.T) {
}
