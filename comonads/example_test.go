package comonads_test

import (
	"fmt"

	"github.com/joergreinhardt/gatwd/comonads"
)

func ExampleGrid_Extend() {
	// 5x5 grid with a horizontal blinker at y=2, x in {1, 2, 3}
	g := comonads.NewGrid(5, 5, false, func(x, y int) bool {
		return y == 2 && (x >= 1 && x <= 3)
	})

	fmt.Printf("Gen 0 population: %d\n", g.Population())

	// Advance one step comonadically using StepConway (which calls Extend)
	g1 := comonads.StepConway(g)
	fmt.Printf("Gen 1 population: %d\n", g1.Population())
	fmt.Printf("Gen 1 vertical center (2, 2): %t\n", g1.Get(2, 2))
	fmt.Printf("Gen 1 vertical top (2, 1): %t\n", g1.Get(2, 1))
	fmt.Printf("Gen 1 vertical bottom (2, 3): %t\n", g1.Get(2, 3))
	fmt.Printf("Gen 1 old left endpoint (1, 2): %t\n", g1.Get(1, 2))

	// Output:
	// Gen 0 population: 3
	// Gen 1 population: 3
	// Gen 1 vertical center (2, 2): true
	// Gen 1 vertical top (2, 1): true
	// Gen 1 vertical bottom (2, 3): true
	// Gen 1 old left endpoint (1, 2): false
}
