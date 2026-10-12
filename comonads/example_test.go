package comonads_test

import (
	"fmt"

	"github.com/joergreinhardt/gatwd/comonads"
)

func ExampleGrid_Extend() {
	// 3x3 integer grid with a center impulse
	g := comonads.NewGrid(3, 3, false, func(x, y int) int {
		if x == 1 && y == 1 {
			return 9
		}
		return 0
	})

	// Apply a local 3x3 smoothing rule comonadically via Extend:
	// sum Moore neighbors + focused cell, divided by 9 (box blur filter)
	smoothed := g.Extend(func(z *comonads.GridZipper[int]) int {
		sum := z.Extract()
		for _, n := range z.MooreNeighbors() {
			sum += n
		}
		return sum / 9
	})

	fmt.Printf("Center before: %d\n", g.Get(1, 1))
	fmt.Printf("Center after:  %d\n", smoothed.Get(1, 1))
	fmt.Printf("Corner after:  %d\n", smoothed.Get(0, 0))

	// Output:
	// Center before: 9
	// Center after:  1
	// Corner after:  1
}
