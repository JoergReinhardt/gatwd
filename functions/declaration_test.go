package functions

import (
	"fmt"
	"testing"

	d "github.com/joergreinhardt/gatwd/data"
)

func TestArgType(t *testing.T) {
	var at = DeclareArguments(Def(Data, d.Int), Def(Data, d.Int), Def(Data, d.Int))
	fmt.Printf("declared arguments: %s\n", at)
	if !at.Type().Match(Def(Def(Data, d.Int), Def(Data, d.Int), Def(Data, d.Int))) {
		t.Fail()
	}

	var result = at.Call(NewNative(1))
	fmt.Printf("match pass int: %s result type: %s\n", result, result.Type())
	if !result.Type().Match(Def(Data, d.Int)) {
		t.Fail()
	}

	result = at.Call(NewNative(1), NewNative(1), NewNative(1))
	fmt.Printf("match pass three ints: %s result type: %s\n", result, result.Type())
	if !result.Type().Match(Def(Vector, Def(Data, d.Int))) {
		t.Fail()
	}

	result = at.Call(NewNative(1.0))
	fmt.Printf("match pass float: %s result type: %s\n", result, result.Type())
	if !result.Type().Match(d.Float) {
		t.Fail()
	}
}
