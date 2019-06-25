package functions

import (
	"fmt"
	"testing"

	d "github.com/joergreinhardt/gatwd/data"
)

func TestMaybe(t *testing.T) {

	var maybe = NewMaybe(NewCase(NewPredictArg(func(arg Callable) bool {
		return arg.TypeNat().Match(d.String)
	}).Nargs(), NewUnary(func(arg Callable) Callable { return arg })))

	fmt.Printf("maybe: %s\n", maybe)

	var str = maybe(New("string"))
	fmt.Printf("str: %s str type name: %s\n", str, str.TypeName())
	if str.String() != "string" {
		t.Fail()
	}

	var none = maybe(New(1))
	fmt.Printf("none: %s none type name: %s\n", none, none.TypeName())

	if none.TypeFnc() != None {
		t.Fail()
	}
}

func TestEither(t *testing.T) {
	var either = NewEither(NewCase(
		NewPredictArg(
			func(arg Callable) bool {
				return arg.TypeNat().Match(d.String)
			}).Nargs()),
		nil,
		NewUnary(func(arg Callable) Callable {
			return NewNative(d.ErrorVal{fmt.Errorf("error: " + arg.String())})
		}),
	)
	fmt.Printf("either: %s either type name: %s\n", either, either.TypeName())

	var str = either(New("string"))
	fmt.Printf("str: %s str type name: %s fnc type: %s nat type: %s\n",
		str, str.TypeName(), str.TypeFnc(), str.TypeNat().TypeName())
	if str.String() != "string" {
		t.Fail()
	}

	var err = either(New(1))
	fmt.Printf("err: %s err type name: %s type nat: %s\n", err, err.TypeName(), err.TypeNat())
	if err.Eval().(d.ErrorVal).E.Error() != "error: 1" {
		t.Fail()
	}
}
