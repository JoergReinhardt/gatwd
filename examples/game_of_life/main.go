package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joergreinhardt/gatwd/comonads"
	"github.com/joergreinhardt/gatwd/visualizer"
)

func main() {
	filePath := flag.String("file", "", "Path to custom ASCII .life or text pattern file")
	presetName := flag.String("preset", "gun-and-eater", "Built-in preset (gun-and-eater, gosper-gun, pulsar, acorn, r-pentomino, glider)")
	modeStr := flag.String("mode", "braille", "Sub-character rendering mode: braille (2x4), quadrant (2x2), halfblock (1x2), cell (1x1)")
	widthFlag := flag.Int("width", 0, "Grid width in sub-pixels (default: auto from terminal / preset)")
	heightFlag := flag.Int("height", 0, "Grid height in sub-pixels (default: auto from terminal / preset)")
	delayMs := flag.Int("delay", 60, "Animation delay between generations in milliseconds")
	maxSteps := flag.Int("steps", 0, "Number of generations to run (0 = infinite interactive loop)")
	color := flag.String("color", "green", "Color of living cells: green, cyan, yellow, magenta, blue, white")
	torus := flag.Bool("torus", true, "Enable toroidal universe (edges wrap around)")
	center := flag.Bool("center", true, "Center pattern in the simulation grid")
	border := flag.Bool("border", true, "Render decorative border around simulation")
	nonInteractive := flag.Bool("non-interactive", false, "Disable interactive screen clearing (batch mode)")
	listPresets := flag.Bool("list-presets", false, "List all available built-in presets and exit")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Gatwd Conway's Game of Life — High-Resolution UTF-8 Comonadic Appliance\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [flags]\n\nFlags:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nSub-Character Modes:\n")
		fmt.Fprintf(os.Stderr, "  braille   : 2x4 sub-pixels per character (8 dots per cell, U+2800..U+28FF) [default]\n")
		fmt.Fprintf(os.Stderr, "  quadrant  : 2x2 sub-pixels per character (4 quadrants per cell, U+2596..U+259F)\n")
		fmt.Fprintf(os.Stderr, "  halfblock : 1x2 sub-pixels per character (upper/lower half blocks, U+2580/U+2584)\n")
		fmt.Fprintf(os.Stderr, "  cell      : 1x1 full character cell (█)\n\n")
		fmt.Fprintf(os.Stderr, "Available Presets:\n")
		for _, p := range AvailablePresets {
			fmt.Fprintf(os.Stderr, "  • %s\n", p)
		}
	}
	flag.Parse()

	if *listPresets {
		fmt.Println("Available Built-in Presets:")
		for _, p := range AvailablePresets {
			fmt.Printf("  • %s\n", p)
		}
		return
	}

	// 1. Load pattern from file or preset
	var pattern *Pattern
	var err error
	var sourceLabel string

	if *filePath != "" {
		pattern, err = LoadFile(*filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading file %s: %v\n", *filePath, err)
			os.Exit(1)
		}
		sourceLabel = fmt.Sprintf("File: %s", *filePath)
	} else {
		pattern, err = GetPreset(*presetName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading preset %s: %v\n", *presetName, err)
			os.Exit(1)
		}
		sourceLabel = fmt.Sprintf("Preset: %s", *presetName)
	}

	mode := visualizer.SubpixelMode(*modeStr)
	switch mode {
	case visualizer.ModeBraille, visualizer.ModeQuadrant, visualizer.ModeHalfBlock, visualizer.ModeCell:
	default:
		mode = visualizer.ModeBraille
	}

	// Determine grid dimensions
	width := *widthFlag
	height := *heightFlag
	if width <= 0 || height <= 0 {
		switch mode {
		case visualizer.ModeBraille:
			// 2x4 sub-pixels per char -> 80 chars x 20 chars = 160 x 80 sub-pixels
			if pattern.DeclaredW > 0 && pattern.DeclaredH > 0 {
				width = pattern.DeclaredW
				height = pattern.DeclaredH
			} else {
				width = pattern.Width + 20
				height = pattern.Height + 16
				if width < 80 {
					width = 80
				}
				if height < 40 {
					height = 40
				}
			}
		case visualizer.ModeQuadrant:
			width = pattern.Width + 10
			height = pattern.Height + 10
			if width < 60 {
				width = 60
			}
			if height < 30 {
				height = 30
			}
		default:
			width = pattern.Width + 6
			height = pattern.Height + 6
		}
	}

	// 2. Initialize Comonadic Grid
	grid := pattern.PlaceOnGrid(width, height, *center, *torus)

	// Clean exit handling for terminal cursor
	if !*nonInteractive {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-sigChan
			fmt.Print("\033[?25h\n") // restore cursor
			os.Exit(0)
		}()
		fmt.Print("\033[?25l") // hide cursor
		defer fmt.Print("\033[?25h")
	}

	// 3. Main Simulation Loop
	canvas := visualizer.NewSubpixelCanvas(width, height, mode)
	xFactor, yFactor := canvas.SubpixelDimensions()
	termCols, termRows := canvas.TerminalSize()

	generation := 0
	stepDuration := time.Duration(*delayMs) * time.Millisecond

	for {
		startTime := time.Now()

		// Count population and populate subpixel canvas
		population := 0
		canvas.Clear()
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				alive := grid.Get(x, y)
				if alive {
					population++
					canvas.Set(x, y, true)
				}
			}
		}

		renderStr := canvas.Render(*color)

		// Render frame
		if !*nonInteractive {
			// Clear screen and home cursor
			fmt.Print("\033[H\033[2J")
		}

		// Status banner
		statusMap := map[string]any{
			"Engine":     "Comonadic 2D Torus Automaton",
			"Source":     sourceLabel,
			"Generation": generation,
			"Population": population,
			"Resolution": fmt.Sprintf("%dx%d sub-pixels (%dx%d chars, mode: %s [%dx%d])", width, height, termCols, termRows, mode, xFactor, yFactor),
		}
		if pattern.Name != "" {
			statusMap["Pattern"] = pattern.Name
		}
		if *border {
			bannerOpts := visualizer.ParseOptions("--border --curved --border-color=cyan --color=white")
			fmt.Println(visualizer.Format(statusMap, bannerOpts))
			fmt.Println(renderStr)
			fmt.Printf("∷ [Gatwd Life] Gen %d | Pop %d | Delay: %dms | Press Ctrl+C to Exit ∷\n", generation, population, *delayMs)
		} else {
			fmt.Printf("--- Gatwd Life: %s | Gen: %d | Pop: %d | Res: %dx%d (%s) ---\n", sourceLabel, generation, population, width, height, mode)
			fmt.Println(renderStr)
		}

		generation++
		if *maxSteps > 0 && generation >= *maxSteps {
			break
		}

		// Evaluate next generation comonadically via StepConway (Extend with ConwayRule)
		calcStart := time.Now()
		grid = comonads.StepConway(grid)
		calcElapsed := time.Since(calcStart)

		// Frame timing
		elapsed := time.Since(startTime)
		if elapsed < stepDuration {
			time.Sleep(stepDuration - elapsed)
		} else if calcElapsed > stepDuration {
			// Don't starve CPU if compute is heavy
			time.Sleep(5 * time.Millisecond)
		}
	}
}
