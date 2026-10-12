package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/pkg/cli"
	"github.com/joergreinhardt/gatwd/visualizer"
)

// Presets mapped to relative (dx, dy) coordinate pairs
var presets = map[string][][2]int{
	"glider": {
		{2, 1}, {3, 2}, {1, 3}, {2, 3}, {3, 3},
	},
	"blinker": {
		{0, -1}, {0, 0}, {0, 1},
	},
	"toad": {
		{1, 0}, {2, 0}, {3, 0}, {0, 1}, {1, 1}, {2, 1},
	},
	"beacon": {
		{0, 0}, {1, 0}, {0, 1}, {3, 2}, {2, 3}, {3, 3},
	},
	"pulsar": {
		// Top quadrants
		{2, 1}, {3, 1}, {4, 1}, {8, 1}, {9, 1}, {10, 1},
		{1, 2}, {6, 2}, {7, 2}, {12, 2},
		{1, 3}, {6, 3}, {7, 3}, {12, 3},
		{1, 4}, {6, 4}, {7, 4}, {12, 4},
		{2, 6}, {3, 6}, {4, 6}, {8, 6}, {9, 6}, {10, 6},
		// Bottom quadrants
		{2, 8}, {3, 8}, {4, 8}, {8, 8}, {9, 8}, {10, 8},
		{1, 10}, {6, 10}, {7, 10}, {12, 10},
		{1, 11}, {6, 11}, {7, 11}, {12, 11},
		{1, 12}, {6, 12}, {7, 12}, {12, 12},
		{2, 13}, {3, 13}, {4, 13}, {8, 13}, {9, 13}, {10, 13},
	},
	"acorn": {
		{1, -1}, {3, 0}, {0, 1}, {1, 1}, {4, 1}, {5, 1}, {6, 1},
	},
	"r-pentomino": {
		{1, 0}, {2, 0}, {0, 1}, {1, 1}, {1, 2},
	},
}

func toInt(v any) int {
	switch n := v.(type) {
	case int:
		return n
	case int8:
		return int(n)
	case int16:
		return int(n)
	case int32:
		return int(n)
	case int64:
		return int(n)
	case uint:
		return int(n)
	case uint8:
		return int(n)
	case uint16:
		return int(n)
	case uint32:
		return int(n)
	case uint64:
		return int(n)
	case float64:
		return int(n)
	case float32:
		return int(n)
	case string:
		var i int
		fmt.Sscanf(n, "%d", &i)
		return i
	default:
		return 0
	}
}

func parsePoint(elem any) (int, int, bool) {
	if ptList, ok := elem.(*collections.List[any]); ok {
		elems := ptList.Elements()
		if len(elems) >= 2 {
			return toInt(elems[0]), toInt(elems[1]), true
		}
	}
	if s, ok := elem.(string); ok {
		s = strings.Trim(s, "[] ")
		fields := strings.Fields(s)
		if len(fields) >= 2 {
			x, err1 := strconv.Atoi(fields[0])
			y, err2 := strconv.Atoi(fields[1])
			if err1 == nil && err2 == nil {
				return x, y, true
			}
		}
	}
	return 0, 0, false
}

func main() {
	presetFlag := flag.String("preset", "glider", "Initial pattern preset: glider, blinker, toad, beacon, pulsar, acorn, r-pentomino")
	stepsFlag := flag.Int("steps", 0, "Number of generations to run (0 = infinite interactive loop)")
	delayMsFlag := flag.Int("delay", 80, "Animation delay between generations in milliseconds")
	modeFlag := flag.String("mode", "braille", "Sub-character rendering mode: braille (2x4), quadrant (2x2), halfblock (1x2), cell (1x1)")
	widthFlag := flag.Int("width", 64, "Grid width in sub-pixels (the border the pattern evolves within)")
	heightFlag := flag.Int("height", 42, "Grid height in sub-pixels (the border the pattern evolves within)")
	colorFlag := flag.String("color", "cyan", "Living cell color: cyan, green, yellow, magenta, white")
	torusFlag := flag.Bool("torus", true, "Wrap coordinates toroidally around canvas boundaries")
	nonInteractiveFlag := flag.Bool("non-interactive", false, "Disable screen clearing (batch output mode)")
	patternFlag := flag.String("pattern", "examples/game_of_life_pure_dsl/patterns/edna.gatwd", "Pure-gatwd startup-grid file (binds `live`); empty to use -preset")
	scriptFlag := flag.String("script", "", "Path to custom .gatwd script to run directly")
	verifyFlag := flag.Bool("verify", false, "Execute the full dynamic verification test suite script")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Gatwd Conway's Game of Life — 100%% Pure DSL Generational Simulation\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] [script.gatwd]\n\nFlags:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nAvailable Presets:\n")
		for p := range presets {
			fmt.Fprintf(os.Stderr, "  • %s\n", p)
		}
	}
	flag.Parse()

	// 1. Script execution mode (if script is supplied or -verify is passed)
	targetScript := *scriptFlag
	if *verifyFlag {
		targetScript = "examples/game_of_life_pure_dsl/game_of_life.gatwd"
	} else if flag.NArg() > 0 {
		targetScript = flag.Arg(0)
	}

	if targetScript != "" {
		content, err := os.ReadFile(targetScript)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading script file %s: %v\n", targetScript, err)
			os.Exit(1)
		}
		interp := cli.NewInterpreter()
		interp.EvaluateScript(string(content))
		return
	}

	// 2. Interactive Running Simulation Mode
	mode := visualizer.SubpixelMode(*modeFlag)
	switch mode {
	case visualizer.ModeBraille, visualizer.ModeQuadrant, visualizer.ModeHalfBlock, visualizer.ModeCell:
	default:
		mode = visualizer.ModeBraille
	}

	width := *widthFlag
	height := *heightFlag

	presetCoords, ok := presets[*presetFlag]
	if !ok {
		presetCoords = presets["glider"]
	}

	// Center the pattern within the grid
	midX := width / 2
	midY := height / 2
	if *presetFlag == "glider" {
		midX = 4
		midY = 2
	}

	var coordStrings []string
	for _, pt := range presetCoords {
		coordStrings = append(coordStrings, fmt.Sprintf("[%d, %d]", midX+pt[0], midY+pt[1]))
	}
	initialListExpr := "[" + strings.Join(coordStrings, ", ") + "]"

	// Clean exit handling for terminal cursor
	if !*nonInteractiveFlag {
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

	// Initialize the interpreter and load the pure-DSL rule (conway_step) from its
	// single source of truth. No Conway logic lives in Go — the wrapper renders.
	interp := cli.NewInterpreter()
	interp.LoadScript("examples/game_of_life_pure_dsl/game_of_life.gatwd")

	// Seed the board. A pattern file binds `live` directly (pre-centered); if none
	// is given or it is missing, fall back to the selected preset.
	seedName := *presetFlag
	if *patternFlag != "" {
		if _, err := os.Stat(*patternFlag); err != nil {
			fmt.Fprintf(os.Stderr, "pattern %s not found (%v); using preset %q\n", *patternFlag, err, *presetFlag)
			interp.EvaluateLine(fmt.Sprintf("let live = %s", initialListExpr))
		} else {
			interp.LoadScript(*patternFlag)
			seedName = *patternFlag
		}
	} else {
		interp.EvaluateLine(fmt.Sprintf("let live = %s", initialListExpr))
	}

	canvas := visualizer.NewSubpixelCanvas(width, height, mode)
	xFactor, yFactor := canvas.SubpixelDimensions()
	termCols, termRows := canvas.TerminalSize()

	generation := 0
	stepDuration := time.Duration(*delayMsFlag) * time.Millisecond

	for {
		startTime := time.Now()

		// Read current live cells from DSL environment
		var currentCells [][2]int
		if rtNode, exists := interp.RuntimeNodes["live"]; exists {
			if val, ok := rtNode.Value.UnwrapRight(); ok {
				if list, ok := val.(*collections.List[any]); ok {
					for elem := range list.Seq() {
						if x, y, ok := parsePoint(elem); ok {
							currentCells = append(currentCells, [2]int{x, y})
						}
					}
				}
			}
		}

		population := len(currentCells)

		// Populate subpixel canvas
		canvas.Clear()
		for _, pt := range currentCells {
			x, y := pt[0], pt[1]
			if *torusFlag {
				x = (x%width + width) % width
				y = (y%height + height) % height
			}
			if x >= 0 && x < width && y >= 0 && y < height {
				canvas.Set(x, y, true)
			}
		}
		renderedStr := canvas.Render(*colorFlag)

		if !*nonInteractiveFlag {
			fmt.Print("\033[H\033[2J")
		}

		// Status card formatted using standard Gatwd visualizer
		statusInfo := map[string]any{
			"Engine":     "100% Pure Gatwd DSL (Reusable def conway_step Macro)",
			"Seed":       seedName,
			"Generation": generation,
			"Population": population,
			"Resolution": fmt.Sprintf("%dx%d sub-pixels (%dx%d chars, mode: %s [%dx%d])", width, height, termCols, termRows, mode, xFactor, yFactor),
		}
		bannerOpts := visualizer.ParseOptions("--border --curved --border-color=cyan --color=white")
		fmt.Println(visualizer.Format(statusInfo, bannerOpts))
		fmt.Println(renderedStr)
		fmt.Printf("∷ [Pure DSL Life] Generation %d | Population %d | Delay: %dms | Press Ctrl+C to Exit ∷\n", generation, population, *delayMsFlag)

		generation++
		if *stepsFlag > 0 && generation >= *stepsFlag {
			break
		}

		// Advance one generation purely via Gatwd DSL conway_step invocation, on a
		// width x height torus so the board stays bounded and keeps wrapping.
		interp.EvaluateLine(fmt.Sprintf("let live = [live, %d, %d] |> conway_step", width, height))

		elapsed := time.Since(startTime)
		if elapsed < stepDuration {
			time.Sleep(stepDuration - elapsed)
		}
	}
}
