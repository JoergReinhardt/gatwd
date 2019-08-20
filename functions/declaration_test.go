package functions

import (
	"fmt"
	"testing"
)

func TestTuple(t *testing.T) {
	var tup = ConstructTupleType(
		DecNative(0).Type(),
		DecNative(0.0).Type(),
		DecNative("").Type(),
	)
	fmt.Printf("tuple: %s\n", tup)
	var tupval = tup(DecNative(23), DecNative(42.23), DecNative("string"))
	fmt.Printf("tuple value: %s type: %s\n", tupval, tupval.Type())

}

func TestNamedTuple(t *testing.T) {
	var ntup = ConstructTupleType(
		DefSym("Named Tuple"),
		DecNative(0).Type(),
		DecNative(0.0).Type(),
		DecNative("").Type(),
	)
	fmt.Printf("named tuple: %s name: %s\n", ntup, ntup.Symbol())
	var tupval = ntup(DecNative(23), DecNative(42.23), DecNative("string"))
	fmt.Printf("tuple value: %s type: %s\n", tupval, tupval.Type())
}

func TestDeclaredTuple(t *testing.T) {
	var tup = ConstructTupleType(
		DecNative(0).Type(),
		DecNative(0.0).Type(),
		DecNative("").Type(),
	).Declare()
	fmt.Printf("tuple: %s\n", tup)
	var partial = tup.Call(DecNative(23))
	fmt.Printf("partial: %s\n", partial)
	partial = partial.Call(DecNative(42.23))
	fmt.Printf("partial: %s\n", partial)
	partial = partial.Call(DecNative("string"))
	fmt.Printf("partial: %s\n", partial)
}
