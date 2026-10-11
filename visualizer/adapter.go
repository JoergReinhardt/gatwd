package visualizer

import (
	"fmt"
	"iter"

	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/evaluator"
	"github.com/joergreinhardt/gatwd/functions"
)

// ParseOptions converts a string like "--ascii --collapse --depth=2" into Options.

// StepVisualize is a pipeline step that transforms any complex structural Node
// into a pretty-printed UTF-8 string graph.
func StepVisualize(args string) evaluator.PipelineStep[functions.Either[error, any]] {
	opts := ParseOptions(args)

	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			for node := range stream {
				if node.Value.IsLeft() {
					err, _ := node.Value.UnwrapLeft()
					errCard := "\n" + FormatAST(node.DSLType, err, opts)
					errNode := dsltypes.NewASTNode(functions.Left[error, any](fmt.Errorf("%s", errCard)), node.DSLType)
					if !yield(errNode) {
						return
					}
					continue
				}

				val, _ := node.Value.UnwrapRight()
				var graphStr string
				if grid, ok := val.(interface {
					Width() int
					Height() int
					Get(x, y int) bool
				}); ok {
					subMode := ModeBraille
					if opts.Style.Corner == "+" {
						subMode = ModeCell
					}
					graphStr = "\n" + RenderGridSubpixel(grid.Width(), grid.Height(), grid.Get, subMode, opts.Color)
				} else if gridPtr, ok := val.(interface {
					LiveCoordinates() [][2]int
					Population() int
				}); ok {
					// Fallback for *comonads.Grid[bool]
					pts := gridPtr.LiveCoordinates()
					w, h := 80, 40
					for _, pt := range pts {
						if pt[0] >= w {
							w = pt[0] + 10
						}
						if pt[1] >= h {
							h = pt[1] + 10
						}
					}
					isLive := func(x, y int) bool {
						for _, pt := range pts {
							if pt[0] == x && pt[1] == y {
								return true
							}
						}
						return false
					}
					graphStr = "\n" + RenderGridSubpixel(w, h, isLive, ModeBraille, opts.Color)
				} else {
					graphStr = "\n" + FormatAST(node.DSLType, val, opts)
				}

				newNode := dsltypes.NewASTNode(functions.Right[error, any](any(graphStr)), node.DSLType)
				if !yield(newNode) {
					return
				}
			}
		}
	}
}

// StepConwayRender converts a comonadic Grid[bool] into high-resolution UTF-8 sub-pixels (braille, quadrant, halfblock, cell).
func StepConwayRender(modeStr, color string) evaluator.PipelineStep[functions.Either[error, any]] {
	mode := SubpixelMode(modeStr)
	if mode == "" {
		mode = ModeBraille
	}
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			for node := range stream {
				if node.Value.IsLeft() {
					if !yield(node) {
						return
					}
					continue
				}

				val, _ := node.Value.UnwrapRight()
				if grid, ok := val.(interface {
					Width() int
					Height() int
					Get(x, y int) bool
				}); ok {
					rendered := RenderGridSubpixel(grid.Width(), grid.Height(), grid.Get, mode, color)
					newNode := dsltypes.NewASTNode(functions.Right[error, any](any(rendered)), node.DSLType)
					if !yield(newNode) {
						return
					}
				} else if gridPtr, ok := val.(interface {
					LiveCoordinates() [][2]int
					Population() int
				}); ok {
					pts := gridPtr.LiveCoordinates()
					w, h := 80, 40
					for _, pt := range pts {
						if pt[0] >= w {
							w = pt[0] + 10
						}
						if pt[1] >= h {
							h = pt[1] + 10
						}
					}
					isLive := func(x, y int) bool {
						for _, pt := range pts {
							if pt[0] == x && pt[1] == y {
								return true
							}
						}
						return false
					}
					rendered := RenderGridSubpixel(w, h, isLive, mode, color)
					newNode := dsltypes.NewASTNode(functions.Right[error, any](any(rendered)), node.DSLType)
					if !yield(newNode) {
						return
					}
				} else {
					errNode := evaluator.InjectError[any](fmt.Errorf("step conway_render expected Grid, got %T", val), node.DSLType)
					if !yield(errNode) {
						return
					}
				}
			}
		}
	}
}

