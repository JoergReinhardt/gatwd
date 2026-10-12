package comonads

// Grid represents an immutable 2D spatial plane or torus of values.
type Grid[T any] struct {
	Width  int
	Height int
	Torus  bool
	cells  []T
}

// NewGrid creates a new 2D grid initialized by a generator function.
func NewGrid[T any](width, height int, torus bool, init func(x, y int) T) *Grid[T] {
	cells := make([]T, width*height)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if init != nil {
				cells[y*width+x] = init(x, y)
			}
		}
	}
	return &Grid[T]{
		Width:  width,
		Height: height,
		Torus:  torus,
		cells:  cells,
	}
}

// Get returns the cell value at (x, y). If Torus is enabled, coordinates wrap around.
// If not Torus and out of bounds, returns the zero value of T.
func (g *Grid[T]) Get(x, y int) T {
	var zero T
	if g.Torus {
		x = (x%g.Width + g.Width) % g.Width
		y = (y%g.Height + g.Height) % g.Height
	} else {
		if x < 0 || x >= g.Width || y < 0 || y >= g.Height {
			return zero
		}
	}
	return g.cells[y*g.Width+x]
}

// SetImmutable returns a new Grid with the cell at (x, y) updated.
func (g *Grid[T]) SetImmutable(x, y int, val T) *Grid[T] {
	newGrid := &Grid[T]{
		Width:  g.Width,
		Height: g.Height,
		Torus:  g.Torus,
		cells:  make([]T, len(g.cells)),
	}
	copy(newGrid.cells, g.cells)
	if g.Torus {
		x = (x%g.Width + g.Width) % g.Width
		y = (y%g.Height + g.Height) % g.Height
	} else if x < 0 || x >= g.Width || y < 0 || y >= g.Height {
		return newGrid
	}
	newGrid.cells[y*g.Width+x] = val
	return newGrid
}

// GridZipper represents a Comonadic cursor focused at a specific (X, Y) coordinate
// within an immutable 2D Grid universe.
type GridZipper[T any] struct {
	grid *Grid[T]
	X    int
	Y    int
}

// ZipperAt focuses a Grid at the given (x, y) coordinate.
func (g *Grid[T]) ZipperAt(x, y int) *GridZipper[T] {
	return &GridZipper[T]{
		grid: g,
		X:    x,
		Y:    y,
	}
}

// Extract fulfills the Comonad counit (W[A] -> A).
// It returns the value at the current focus.
func (z *GridZipper[T]) Extract() T {
	return z.grid.Get(z.X, z.Y)
}

// Move creates a new zipper moved by (dx, dy) relative to current focus.
func (z *GridZipper[T]) Move(dx, dy int) *GridZipper[T] {
	return &GridZipper[T]{
		grid: z.grid,
		X:    z.X + dx,
		Y:    z.Y + dy,
	}
}

// MooreNeighbors returns the 8 surrounding values in the Moore neighborhood
// (N, NE, E, SE, S, SW, W, NW) in clockwise order starting from North.
func (z *GridZipper[T]) MooreNeighbors() [8]T {
	return [8]T{
		z.grid.Get(z.X, z.Y-1),   // N
		z.grid.Get(z.X+1, z.Y-1), // NE
		z.grid.Get(z.X+1, z.Y),   // E
		z.grid.Get(z.X+1, z.Y+1), // SE
		z.grid.Get(z.X, z.Y+1),   // S
		z.grid.Get(z.X-1, z.Y+1), // SW
		z.grid.Get(z.X-1, z.Y),   // W
		z.grid.Get(z.X-1, z.Y-1), // NW
	}
}

// Duplicate fulfills the Comonad cojoin (W[A] -> W[W[A]]).
// It lifts the current grid into a grid of zippers, each focused on its corresponding cell.
func Duplicate[T any](z *GridZipper[T]) *Grid[*GridZipper[T]] {
	return NewGrid(z.grid.Width, z.grid.Height, z.grid.Torus, func(x, y int) *GridZipper[T] {
		return &GridZipper[T]{
			grid: z.grid,
			X:    x,
			Y:    y,
		}
	})
}

// Extend fulfills the Comonad cobind ((W[A] -> B) -> W[A] -> W[B]).
// It applies a local neighborhood rule function over every position in the grid universe,
// producing the next entire universe.
func Extend[A, B any](z *GridZipper[A], rule func(*GridZipper[A]) B) *Grid[B] {
	return NewGrid(z.grid.Width, z.grid.Height, z.grid.Torus, func(x, y int) B {
		cursor := &GridZipper[A]{
			grid: z.grid,
			X:    x,
			Y:    y,
		}
		return rule(cursor)
	})
}

// VonNeumannNeighbors returns the 4 orthogonal neighboring values
// (North, East, South, West) relative to the focused cursor.
func (z *GridZipper[T]) VonNeumannNeighbors() [4]T {
	return [4]T{
		z.grid.Get(z.X, z.Y-1), // N
		z.grid.Get(z.X+1, z.Y),   // E
		z.grid.Get(z.X, z.Y+1), // S
		z.grid.Get(z.X-1, z.Y),   // W
	}
}

// Extend evaluates a local rule across the entire grid comonadically.
func (g *Grid[T]) Extend(rule func(*GridZipper[T]) T) *Grid[T] {
	return Extend(g.ZipperAt(0, 0), rule)
}

// Count returns the number of cells matching the given predicate.
func (g *Grid[T]) Count(predicate func(val T) bool) int {
	count := 0
	for _, cell := range g.cells {
		if predicate(cell) {
			count++
		}
	}
	return count
}

// CoordinatesWhere returns a slice of [x, y] coordinates where the predicate returns true.
func (g *Grid[T]) CoordinatesWhere(predicate func(val T) bool) [][2]int {
	var res [][2]int
	for y := 0; y < g.Height; y++ {
		for x := 0; x < g.Width; x++ {
			if predicate(g.Get(x, y)) {
				res = append(res, [2]int{x, y})
			}
		}
	}
	return res
}


