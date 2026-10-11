package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	"iter"

	"github.com/joergreinhardt/gatwd/comonads"
	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/evaluator"
	"github.com/joergreinhardt/gatwd/functions"
	"github.com/joergreinhardt/gatwd/visualizer"
)

// BlinkerPattern generates an initial blinking oscillator on a 2D Torus grid.
func CreateBlinkerGrid(width, height int) *comonads.Grid[bool] {
	return comonads.NewGrid(width, height, true, func(x, y int) bool {
		midX := width / 2
		midY := height / 2
		// Vertical blinker: 3 cells centered
		return x == midX && (y == midY-1 || y == midY || y == midY+1)
	})
}

// CreateGliderGrid places a classic Conway glider traveling southeast.
func CreateGliderGrid(width, height int) *comonads.Grid[bool] {
	return comonads.NewGrid(width, height, true, func(x, y int) bool {
		// Glider coordinates at top-left:
		// . O .
		// . . O
		// O O O
		return (x == 2 && y == 1) ||
			(x == 3 && y == 2) ||
			(x == 1 && y == 3) ||
			(x == 2 && y == 3) ||
			(x == 3 && y == 3)
	})
}

// CreateAcornGrid places the legendary Acorn methuselah (generates 5,206 steps).
func CreateAcornGrid(width, height int) *comonads.Grid[bool] {
	return comonads.NewGrid(width, height, true, func(x, y int) bool {
		midX := width / 2
		midY := height / 2
		// Acorn:
		// . O . . . . .
		// . . . O . . .
		// O O . . O O O
		dx := x - midX
		dy := y - midY
		return (dx == 1 && dy == -1) ||
			(dx == 3 && dy == 0) ||
			(dx == 0 && dy == 1) ||
			(dx == 1 && dy == 1) ||
			(dx == 4 && dy == 1) ||
			(dx == 5 && dy == 1) ||
			(dx == 6 && dy == 1)
	})
}

func main() {
	preset := flag.String("preset", "glider", "Pattern preset: glider, blinker, acorn")
	modeStr := flag.String("mode", "braille", "Sub-character rendering mode: braille (2x4), quadrant (2x2), halfblock (1x2), cell (1x1)")
	width := flag.Int("width", 80, "Grid width in sub-pixels")
	height := flag.Int("height", 40, "Grid height in sub-pixels")
	delayMs := flag.Int("delay", 80, "Frame delay in milliseconds")
	maxSteps := flag.Int("steps", 0, "Max generations (0 = infinite)")
	color := flag.String("color", "cyan", "Living cell color (green, cyan, yellow, magenta)")
	nonInteractive := flag.Bool("non-interactive", false, "Disable screen clear (batch mode)")
	flag.Parse()

	mode := visualizer.SubpixelMode(*modeStr)
	var initialGrid *comonads.Grid[bool]
	switch *preset {
	case "blinker":
		initialGrid = CreateBlinkerGrid(*width, *height)
	case "acorn":
		initialGrid = CreateAcornGrid(*width, *height)
	default:
		initialGrid = CreateGliderGrid(*width, *height)
	}

	// 1. Initialize Functional Evaluator Engine & Grid Type Flag
	engine := evaluator.NewEngine[functions.Either[error, any]]()
	gridFlag := dsltypes.BitFlag(1 << 8)
	dsltypes.RegisterTypeFlag(gridFlag, "Grid[bool]")

	// 2. Setup Evaluator Pipeline
	// The entire generational cycle is handed to the Evaluator Pipeline:
	// Initial Grid -> StepConway (Extend) -> Render via Visualizer
	stepConway := func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			for node := range stream {
				if node.Value.IsLeft() {
					if !yield(node) {
						return
					}
					continue
				}
				val, _ := node.Value.UnwrapRight()
				if grid, ok := val.(*comonads.Grid[bool]); ok {
					nextGrid := comonads.StepConway(grid)
					newNode := dsltypes.NewASTNode(functions.Right[error, any](any(nextGrid)), node.DSLType)
					if !yield(newNode) {
						return
					}
				}
			}
		}
	}
	steps := []evaluator.PipelineStep[functions.Either[error, any]]{
		stepConway,
	}

	// Graceful terminal exit
	if !*nonInteractive {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-sig
			fmt.Print("\033[?25h\n")
			os.Exit(0)
		}()
		fmt.Print("\033[?25l")
		defer fmt.Print("\033[?25h")
	}

	canvas := visualizer.NewSubpixelCanvas(*width, *height, mode)
	xFactor, yFactor := canvas.SubpixelDimensions()
	termCols, termRows := canvas.TerminalSize()

	generation := 0
	currentGrid := initialGrid
	stepDuration := time.Duration(*delayMs) * time.Millisecond

	for {
		startTime := time.Now()

		// Populate canvas from functional grid
		canvas.Clear()
		for _, pt := range currentGrid.LiveCoordinates() {
			canvas.Set(pt[0], pt[1], true)
		}
		renderedStr := canvas.Render(*color)

		if !*nonInteractive {
			fmt.Print("\033[H\033[2J")
		}

		// Status card rendered via Gatwd Visualizer format
		statusInfo := map[string]any{
			"Variant":    "Evaluator Pipeline (Functional Comonadic Stream)",
			"Preset":     *preset,
			"Generation": generation,
			"Population": currentGrid.Population(),
			"Resolution": fmt.Sprintf("%dx%d sub-pixels (%dx%d chars, mode: %s [%dx%d])", *width, *height, termCols, termRows, mode, xFactor, yFactor),
		}
		bannerOpts := visualizer.ParseOptions("--border --curved --border-color=blue --color=white")
		fmt.Println(visualizer.Format(statusInfo, bannerOpts))
		fmt.Println(renderedStr)
		fmt.Printf("∷ [Pipeline Life] Generation %d | Population %d | Delay: %dms ∷\n", generation, currentGrid.Population(), *delayMs)

		generation++
		if *maxSteps > 0 && generation >= *maxSteps {
			break
		}

		// 3. EXECUTION VIA EVALUATOR PIPELINE
		// We lift currentGrid into an iter.Seq stream of ASTNode
		inputStream := func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			node := dsltypes.NewASTNode(functions.Right[error, any](any(currentGrid)), gridFlag)
			yield(node)
		}

		// Evaluate using pure functional Evaluator engine
		outStream := engine.EvaluatePipeline(inputStream, steps)
		for node := range outStream {
			if node.Value.IsRight() {
				res, _ := node.Value.UnwrapRight()
				if nextG, ok := res.(*comonads.Grid[bool]); ok {
					currentGrid = nextG
				}
			}
		}

		// Rate limiter
		elapsed := time.Since(startTime)
		if elapsed < stepDuration {
			time.Sleep(stepDuration - elapsed)
		}
	}
	_ = context.Background()
}
