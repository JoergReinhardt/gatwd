package main

import "github.com/joergreinhardt/gatwd/comonads"

// ConwayRule is the canonical comonadic local rule for Conway's Game of Life.
//   - Any live cell with 2 or 3 live neighbors survives.
//   - Any dead cell with exactly 3 live neighbors becomes a live cell.
//   - All other cells die or stay dead.
func ConwayRule(z *comonads.GridZipper[bool]) bool {
	aliveNeighbors := 0
	for _, neighbor := range z.MooreNeighbors() {
		if neighbor {
			aliveNeighbors++
		}
	}

	currentlyAlive := z.Extract()
	if currentlyAlive {
		return aliveNeighbors == 2 || aliveNeighbors == 3
	}
	return aliveNeighbors == 3
}

// StepConway evaluates one discrete generation of Conway's Game of Life
// across the entire grid comonadically using comonads.Extend.
func StepConway(g *comonads.Grid[bool]) *comonads.Grid[bool] {
	return comonads.Extend(g.ZipperAt(0, 0), ConwayRule)
}
