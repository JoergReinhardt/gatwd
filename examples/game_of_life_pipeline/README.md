# Conway's Game of Life — Evaluator Pipeline Variant

This directory contains Conway's Game of Life implemented using Gatwd's **Evaluator Pipeline Engine** ([`evaluator.Engine`](file:///home/j/src/rnd/gatwd/evaluator/engine.go)).

## Overview

Unlike the standalone loop appliance in [`examples/game_of_life`](file:///home/j/src/rnd/gatwd/examples/game_of_life), this variant:
1. Wraps the simulation state in an immutable comonadic grid ([`comonads.Grid[bool]`](file:///home/j/src/rnd/gatwd/comonads/grid.go)).
2. Lifts the grid into an `iter.Seq` stream of `dsltypes.ASTNode`.
3. Models generational advancement as a functional `evaluator.PipelineStep` transducer (`stepConway`).
4. Executes the step using `engine.EvaluatePipeline(inputStream, steps)`.
5. Renders high-resolution sub-pixel output using [`visualizer.SubpixelCanvas`](file:///home/j/src/rnd/gatwd/visualizer/subpixel.go) formatted with Gatwd's diagnostic card visualizer.

Notice that this variant does **not** modify or pollute the core `evaluator/pipeline.go` transducer set; it constructs domain-specific pipeline steps locally using standard Go function composition.

---

## Running the Example

### Interactive Mode (Animation Loop)
```bash
go run examples/game_of_life_pipeline/main.go -preset=glider -delay=80ms -mode=braille -color=cyan
```

### Batch Mode (Non-interactive, Step-limited)
```bash
go run examples/game_of_life_pipeline/main.go -preset=blinker -steps=3 -non-interactive
```

### Available Flags
- `-preset`: `glider`, `blinker`, `acorn`
- `-mode`: `braille` (2x4), `quadrant` (2x2), `halfblock` (1x2), `cell` (1x1)
- `-color`: `green`, `cyan`, `yellow`, `magenta`, `blue`, `white`
- `-delay`: Animation frame delay in milliseconds (default: 80ms)
- `-steps`: Number of generations to advance (0 = infinite loop)
- `-non-interactive`: Disable terminal screen clearing for test suites or CI pipes
