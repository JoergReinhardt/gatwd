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

func testConwayRule(z *GridZipper[bool]) bool {
	aliveNeighbors := 0
	for _, neighbor := range z.MooreNeighbors() {
		if neighbor {
			aliveNeighbors++
		}
	}
	if z.Extract() {
		return aliveNeighbors == 2 || aliveNeighbors == 3
	}
	return aliveNeighbors == 3
}

func testStepConway(g *Grid[bool]) *Grid[bool] {
	return Extend(g.ZipperAt(0, 0), testConwayRule)
}

func TestGridComonad_CellularAutomaton(t *testing.T) {
	// A 5x5 grid with a horizontal blinker at (2, 1), (2, 2), (2, 3)
	// After 1 step, should be vertical at (1, 2), (2, 2), (3, 2)
	// After 2 steps, should return to horizontal.
	g := NewGrid(5, 5, false, func(x, y int) bool {
		return y == 2 && (x == 1 || x == 2 || x == 3)
	})

	if g.Count(func(v bool) bool { return v }) != 3 {
		t.Fatalf("expected 3 living cells initially")
	}

	g1 := testStepConway(g)

	// In g1, should be vertical at x=2, y=1, 2, 3
	if !g1.Get(2, 1) || !g1.Get(2, 2) || !g1.Get(2, 3) {
		t.Errorf("Blinker generation 1 failed to oscillate to vertical")
	}
	if g1.Get(1, 2) || g1.Get(3, 2) {
		t.Errorf("Blinker generation 1 old horizontal endpoints still alive")
	}

	g2 := testStepConway(g1)

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

func TestGridComonad_VonNeumannAndCoordinates(t *testing.T) {
	g := NewGrid(3, 3, false, func(x, y int) int {
		return y*3 + x
	})
	z := g.ZipperAt(1, 1)
	vn := z.VonNeumannNeighbors() // N (1,0)=1, E (2,1)=5, S (1,2)=7, W (0,1)=3
	expected := [4]int{1, 5, 7, 3}
	if vn != expected {
		t.Errorf("expected Von Neumann neighbors %v, got %v", expected, vn)
	}

	coords := g.CoordinatesWhere(func(val int) bool {
		return val%2 != 0 // odd numbers: 1, 3, 5, 7
	})
	if len(coords) != 4 {
		t.Errorf("expected 4 odd coordinates, got %d", len(coords))
	}
}

