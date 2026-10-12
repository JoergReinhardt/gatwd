package dsltypes_test

import (
	"fmt"

	"github.com/joergreinhardt/gatwd/dsltypes"
)

func ExampleVariant_Map() {
	// A tagged union is a Functor over its payload: Map transforms the carried
	// value while leaving the branch identity (Tag/Flag) untouched.
	v := dsltypes.NewVariant("Just", dsltypes.BitFlag(1), 21)
	doubled := v.Map(func(x any) any { return x.(int) * 2 })
	fmt.Println(doubled)

	// Output:
	// Just(42)
}

func ExampleVariant_Fold() {
	// Fold eliminates the variant by applying f to its tag and payload,
	// collapsing the structure to a single result (its catamorphism).
	v := dsltypes.NewVariant("Error", dsltypes.BitFlag(2), "boom")
	msg := v.Fold(func(tag string, payload any) any {
		return fmt.Sprintf("%s: %v", tag, payload)
	})
	fmt.Println(msg)

	// Output:
	// Error: boom
}

func ExampleVariant_Bind() {
	// Bind (monadic flatMap) applies f to the payload and returns the Variant f
	// produces, allowing a branch to be re-tagged based on its carried value.
	v := dsltypes.NewVariant("Input", dsltypes.BitFlag(1), 0)
	classify := func(payload any) dsltypes.Variant {
		if payload.(int) == 0 {
			return dsltypes.NewVariant("Nothing", dsltypes.BitFlag(0), nil)
		}
		return dsltypes.NewVariant("Just", dsltypes.BitFlag(1), payload)
	}
	fmt.Println(v.Bind(classify))

	// Output:
	// Nothing
}
