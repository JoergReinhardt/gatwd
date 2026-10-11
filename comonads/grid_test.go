package comonads

import (
	"testing"
)

func TestGridComonad_Extract(t *testing.T) {
	g := NewGrid(3, 3, false, func(x, y int) int {
		return y*3 + x
	})
	z := g.ZipperAt(1, 1)
	if z.Extract() != 4 {
		t.Fatalf("expected Extract() = 4 at (1, 1), got %d", z.Extract())
	}
}

func TestGridComonad_Laws(t *testing.T) {
	// Comonad Law 1: Extend Extract z == z.grid
	g := NewGrid(4, 4, true, func(x, y int) int {
		return x + y*10
	})
	z := g.ZipperAt(2, 2)
	idGrid := Extend(z, func(cursor *GridZipper[int]) int {
		return cursor.Extract()
	})
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			if idGrid.Get(x, y) != g.Get(x, y) {
				t.Errorf("Comonad law 1 violated at (%d, %d)", x, y)
			}
		}
	}

	// Comonad Law 2: Extract (Extend f z) == f z
	rule := func(cursor *GridZipper[int]) int {
		sum := 0
		for _, n := range cursor.MooreNeighbors() {
			sum += n
		}
		return sum + cursor.Extract()
	}
	extended := Extend(z, rule)
	if extended.Get(z.X, z.Y) != rule(z) {
		t.Errorf("Comonad law 2 violated: expected %d, got %d", rule(z), extended.Get(z.X, z.Y))
	}
}

func TestGridComonad_ConwayBlinker(t *testing.T) {
	// A 5x5 grid with a horizontal blinker at (2, 1), (2, 2), (2, 3)
	// After 1 step, should be vertical at (1, 2), (2, 2), (3, 2)
	// After 2 steps, should return to horizontal.
	g := NewGrid(5, 5, false, func(x, y int) bool {
		return y == 2 && (x == 1 || x == 2 || x == 3)
	})

	g1 := StepConway(g)

	// In g1, should be vertical at x=2, y=1, 2, 3
	if !g1.Get(2, 1) || !g1.Get(2, 2) || !g1.Get(2, 3) {
		t.Errorf("Blinker generation 1 failed to oscillate to vertical")
	}
	if g1.Get(1, 2) || g1.Get(3, 2) {
		t.Errorf("Blinker generation 1 old horizontal endpoints still alive")
	}

	g2 := StepConway(g1)

	// In g2, should return to horizontal
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			expected := (y == 2 && (x == 1 || x == 2 || x == 3))
			if g2.Get(x, y) != expected {
				t.Errorf("Blinker period 2 mismatch at (%d, %d): expected %v, got %v", x, y, expected, g2.Get(x, y))
			}
		}
	}
}

func TestGridComonad_ConwayGlider(t *testing.T) {
	// Glider at (1, 0), (2, 1), (0, 2), (1, 2), (2, 2)
	// In 4 generations, a glider translates by (1, 1)
	g := NewGrid(10, 10, true, func(x, y int) bool {
		glider := (x == 1 && y == 0) ||
			(x == 2 && y == 1) ||
			(x == 0 && y == 2) ||
			(x == 1 && y == 2) ||
			(x == 2 && y == 2)
		return glider
	})

	curr := g
	for i := 0; i < 4; i++ {
		curr = StepConway(curr)
	}

	// Should be translated by (1, 1)
	expectedGlider := func(x, y int) bool {
		return (x == 2 && y == 1) ||
			(x == 3 && y == 2) ||
			(x == 1 && y == 3) ||
			(x == 2 && y == 3) ||
			(x == 3 && y == 3)
	}

	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			if curr.Get(x, y) != expectedGlider(x, y) {
				t.Errorf("Glider position mismatch at (%d, %d) after 4 steps: expected %v, got %v", x, y, expectedGlider(x, y), curr.Get(x, y))
			}
		}
	}
}
