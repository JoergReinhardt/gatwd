package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/joergreinhardt/gatwd/comonads"
	"github.com/joergreinhardt/gatwd/pkg/cli"
	"github.com/joergreinhardt/gatwd/visualizer"
)

// parseLifePattern converts an ASCII .life representation or coordinate string into a comonads.Grid[bool].
// This is registered directly into the custom edition via WithType("LifeGrid", parseLifePattern).
func parseLifePattern(raw string) any {
	lines := strings.Split(raw, "\n")
	var cellRows [][]bool
	maxW := 0

	for _, line := range lines {
		line = strings.TrimRight(line, "\r")
		t := strings.TrimSpace(line)
		if t == "" || strings.HasPrefix(t, "#") || strings.HasPrefix(t, "!") {
			continue
		}
		var row []bool
		for _, ch := range line {
			if ch == 'O' || ch == 'o' || ch == '*' || ch == '#' || ch == '1' {
				row = append(row, true)
			} else {
				row = append(row, false)
			}
		}
		if len(row) > maxW {
			maxW = len(row)
		}
		cellRows = append(cellRows, row)
	}

	targetW := maxW + 10
	if targetW < 20 {
		targetW = 20
	}
	targetH := len(cellRows) + 10
	if targetH < 15 {
		targetH = 15
	}

	offsetX := (targetW - maxW) / 2
	offsetY := (targetH - len(cellRows)) / 2

	return comonads.NewGrid(targetW, targetH, true, func(x, y int) bool {
		px := x - offsetX
		py := y - offsetY
		if py >= 0 && py < len(cellRows) {
			row := cellRows[py]
			if px >= 0 && px < len(row) {
				return row[px]
			}
		}
		return false
	})
}

// buildLifeGridFromDimensions creates a step to resize or embed an ASCII string into an explicitly dimensioned grid.
func stepParseLifeWithDims(arg string, val any) (any, error) {
	parts := strings.Fields(arg)
	targetW, targetH := 60, 30
	if len(parts) > 0 {
		if w, err := strconv.Atoi(parts[0]); err == nil && w > 0 {
			targetW = w
		}
	}
	if len(parts) > 1 {
		if h, err := strconv.Atoi(parts[1]); err == nil && h > 0 {
			targetH = h
		}
	}

	// If val is already a *comonads.Grid[bool], resize or re-center it into targetW, targetH
	if existingGrid, ok := val.(*comonads.Grid[bool]); ok {
		offsetX := (targetW - existingGrid.Width) / 2
		offsetY := (targetH - existingGrid.Height) / 2
		return comonads.NewGrid(targetW, targetH, true, func(x, y int) bool {
			px := x - offsetX
			py := y - offsetY
			if px >= 0 && px < existingGrid.Width && py >= 0 && py < existingGrid.Height {
				return existingGrid.Get(px, py)
			}
			return false
		}), nil
	}

	strVal := fmt.Sprintf("%v", val)
	lines := strings.Split(strVal, "\n")
	var cellRows [][]bool
	maxW := 0

	for _, line := range lines {
		line = strings.TrimRight(line, "\r")
		t := strings.TrimSpace(line)
		if t == "" || strings.HasPrefix(t, "#") || strings.HasPrefix(t, "!") {
			continue
		}
		var row []bool
		for _, ch := range line {
			if ch == 'O' || ch == 'o' || ch == '*' || ch == '#' || ch == '1' {
				row = append(row, true)
			} else {
				row = append(row, false)
			}
		}
		if len(row) > maxW {
			maxW = len(row)
		}
		cellRows = append(cellRows, row)
	}

	offsetX := (targetW - maxW) / 2
	offsetY := (targetH - len(cellRows)) / 2

	return comonads.NewGrid(targetW, targetH, true, func(x, y int) bool {
		px := x - offsetX
		py := y - offsetY
		if py >= 0 && py < len(cellRows) {
			row := cellRows[py]
			if px >= 0 && px < len(row) {
				return row[px]
			}
		}
		return false
	}), nil
}

// conwayRule evaluates Conway's B3/S23 transition comonadically for a cell.
func conwayRule(z *comonads.GridZipper[bool]) bool {
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

// stepConwayAdvances advances any comonads.Grid[bool] comonad by one discrete generation.
func stepConwayAdvance(val any) any {
	if grid, ok := val.(*comonads.Grid[bool]); ok {
		return comonads.Extend(grid.ZipperAt(0, 0), conwayRule)
	}
	return val
}

// stepConwayRender renders a comonadic Grid[bool] into high-resolution UTF-8 sub-pixels (braille, quadrant, halfblock).
func stepConwayRender(arg string, val any) (any, error) {
	grid, ok := val.(*comonads.Grid[bool])
	if !ok {
		return nil, fmt.Errorf("conway_render expects *comonads.Grid[bool], got %T", val)
	}

	mode := visualizer.ModeBraille
	color := "cyan"
	parts := strings.Fields(arg)
	if len(parts) > 0 {
		m := visualizer.SubpixelMode(parts[0])
		if m != "" {
			mode = m
		}
	}
	if len(parts) > 1 {
		color = parts[1]
	}

	return "\n" + visualizer.RenderGridSubpixel(grid.Width, grid.Height, grid.Get, mode, color), nil
}

func main() {
	// Mint "gatwd-life": A custom Gatwd Edition for cellular automata & comonadic physics simulations.
	// Notice that the core Gatwd language spec, core parser, and core built-ins remain 100% untouched!
	cli.NewEdition("gatwd-life").
		// 1. Register domain type factory
		WithType("LifeGrid", parseLifePattern).
		// 2. Register custom transducers using ergonomic helpers
		WithStep(cli.ArgStep("parse_life", stepParseLifeWithDims)).
		WithStep(cli.MapStep("conway_step", stepConwayAdvance)).
		WithStep(cli.ArgStep("conway_render", stepConwayRender)).
		// 3. Launch CLI runtime
		Run()
}
