package functions_test

import (
	"fmt"

	"github.com/joergreinhardt/gatwd/functions"
)

func ExampleEither_Bind() {
	// Bind chains computations on the Right branch and short-circuits on Left.
	half := func(x int) functions.Either[string, int] {
		if x%2 == 0 {
			return functions.Right[string, int](x / 2)
		}
		return functions.Left[string, int]("odd input")
	}

	// Right chains: 8 -> 4 -> 2.
	chained := functions.Right[string, int](8).Bind(half).Bind(half)
	fmt.Println(chained)

	// Left short-circuits: 5 is odd, so the first Bind yields Left and the
	// second Bind is skipped.
	shortCircuit := functions.Right[string, int](5).Bind(half).Bind(half)
	fmt.Println(shortCircuit)

	// Output:
	// Right(2)
	// Left(odd input)
}
