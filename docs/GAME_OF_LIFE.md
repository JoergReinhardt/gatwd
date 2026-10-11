# Conway's Game of Life — Architecture & Implementation Variants

This document explains the mathematical foundations, architecture, and three distinct implementation approaches of **Conway's Game of Life** in the Gatwd ecosystem.

The three variants demonstrate how Gatwd achieves extensibility without bloating the core compiler, spec, or standard library:

1. **Standard Edition Appliance ([`examples/game_of_life`](file:///home/j/src/rnd/gatwd/examples/game_of_life))**: A high-performance native Go binary using low-level functional comonad types.
2. **Evaluator Pipeline Variant ([`examples/game_of_life_pipeline`](file:///home/j/src/rnd/gatwd/examples/game_of_life_pipeline))**: A stream-driven implementation where the state transition is modeled as a functional transducer step executed by the lazy `evaluator.Engine`.
3. **Minted Custom Edition ([`examples/game_of_life_edition`](file:///home/j/src/rnd/gatwd/examples/game_of_life_edition))**: A domain-specific distribution (`gatwd-life`) that extends the Gatwd DSL parser and runtime with custom domain types and transducers, enabling Conway's Game of Life to be authored and executed in pure Gatwd script syntax (`.gatwd`).

---

## 1. Mathematical Foundations: The 2D Grid Comonad

In pure functional programming, cellular automata are modeled using **Comonads**. While a Monad provides computation in an execution context ($\text{return} : A \to M(A)$, $\text{bind} : M(A) \to (A \to M(B)) \to M(B)$), a Comonad models computation *within a spatial or observational context*:

$$\text{extract} : W(A) \to A$$
$$\text{duplicate} : W(A) \to W(W(A))$$
$$\text{extend} : W(A) \to (W(A) \to B) \to W(B)$$

### Toroidal Space as a Comonad
In Gatwd, [`comonads.Grid[T]`](file:///home/j/src/rnd/gatwd/comonads/grid.go) represents a 2-dimensional toroidal plane:
- **`extract`**: Inspects the value of the focused cell at the current cursor $(x, y)$.
- **`duplicate`**: Produces a grid of sub-grids, where every cell contains a view of the universe centered on that cell.
- **`extend` (Cobind)**: Applies a local neighborhood rule $f : W(A) \to B$ across all coordinates simultaneously:

$$\text{NextGrid} = \text{Grid}.\text{Extend}(\text{MooreRule})$$

The Conway transition rule checks the 8 neighboring cells (Moore neighborhood) relative to the focused cursor:
- A live cell with 2 or 3 live neighbors survives.
- A dead cell with exactly 3 live neighbors becomes alive.
- All other cells die or remain dead.

---

## 2. High-Resolution UTF-8 Sub-Pixel Visualization

All variants utilize [`visualizer.SubpixelCanvas`](file:///home/j/src/rnd/gatwd/visualizer/subpixel.go) to render high-resolution simulation grids directly into terminal character cells using Unicode sub-character encodings:

| Mode | Aspect Ratio | Resolution Gain | Unicode Block | Description |
| :--- | :---: | :---: | :---: | :--- |
| **`braille`** | 2 × 4 dots/char | **8×** | `U+2800`–`U+28FF` | Standard high-density braille matrix (`⠁`, `⠃`, `⠓`, etc.) |
| **`quadrant`** | 2 × 2 blocks/char | **4×** | `U+2596`–`U+259F` | Quarter-character block elements (`▖`, `▗`, `▘`, `▙`, etc.) |
| **`halfblock`** | 1 × 2 blocks/char | **2×** | `U+2580` / `U+2584` | Top/Bottom half-block characters (`▀`, `▄`) |
| **`cell`** | 1 × 1 full char | 1× | `U+2588` | Full block character (`█`) |

---

## 3. Variant Comparison Matrix

| Feature | Variant 1: Native Appliance | Variant 2: Pipeline Engine | Variant 3: Minted Custom Edition | Variant 4: 100% Pure DSL |
| :--- | :--- | :--- | :--- | :--- |
| **Location** | [`examples/game_of_life/`](file:///home/j/src/rnd/gatwd/examples/game_of_life) | [`examples/game_of_life_pipeline/`](file:///home/j/src/rnd/gatwd/examples/game_of_life_pipeline) | [`examples/game_of_life_edition/`](file:///home/j/src/rnd/gatwd/examples/game_of_life_edition) | [`examples/game_of_life_pure_dsl/`](file:///home/j/src/rnd/gatwd/examples/game_of_life_pure_dsl) |
| **Language** | Go | Go + `evaluator.Engine` | Pure Gatwd DSL + Edition binary | **100% Pure Gatwd DSL** |
| **Extensibility Pattern** | Standalone Application | Embedded Functional Pipeline | Minted Edition (`cli.NewEdition`) | **Zero Go code (Standard Builtins)** |
| **Core Spec Touched?** | **No** | **No** | **No** | **No** |
| **Generational Transition** | Direct `comonads.StepConway` loop | `evaluator.PipelineStep` transducer stream | Gatwd pipe: `\|> conway_step` | Sparse `Set` operations & catamorphisms |
| **Runs With** | `go run ./examples/game_of_life` | `go run .../main.go` | `go run .../main.go script.gatwd` | **Standard `gatwd script.gatwd`** |
| **Use Case** | Maximum throughput / Standalone CLI tool | Composing physics steps with filters | Domain distribution with high-res Braille | Portable scripting on vanilla Gatwd |

---

## 4. Variant Walkthroughs

### Variant 1: Native Appliance (`examples/game_of_life`)
- **Philosophy**: An end-to-end interactive terminal tool with preset loading (`gosper-gun`, `pulsar`, `acorn`), file loaders (`.life` ASCII formats), dynamic terminal resizing, and color controls.
- **Run command**:
  ```bash
  go run ./examples/game_of_life -preset=gosper-gun -mode=braille -color=green
  ```

---

### Variant 2: Evaluator Pipeline (`examples/game_of_life_pipeline`)
- **Philosophy**: Shows how to drive the simulation purely through Gatwd's lazy `iter.Seq` pipeline evaluation engine without any imperative state mutation loops.
- **Code structure**:
  ```go
  // Define pipeline step transducer
  stepConway := func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
      return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
          for node := range stream {
              if val, ok := node.Value.UnwrapRight(); ok {
                  if grid, isGrid := val.(*comonads.Grid[bool]); isGrid {
                      nextGrid := comonads.StepConway(grid)
                      yield(dsltypes.NewASTNode(functions.Right[error, any](any(nextGrid)), node.DSLType))
                  }
              }
          }
      }
  }

  // Stream current generation through evaluator engine
  steps := []evaluator.PipelineStep[functions.Either[error, any]]{ stepConway }
  outStream := engine.EvaluatePipeline(inputStream, steps)
  ```
- **Run command**:
  ```bash
  go run examples/game_of_life_pipeline/main.go -preset=glider -steps=5 -non-interactive
  ```

---

### Variant 3: Minted Custom Edition (`examples/game_of_life_edition`)
- **Philosophy**: Demonstrates Gatwd's premier extensibility mechanism: **Minting an Edition**.
  Instead of polluting the core language spec, an author mints a customized distribution (`gatwd-life`) with domain types and transducers registered in user-space.

#### Step 1: Minting the Edition in Go (`main.go`)
```go
package main

import (
    "github.com/joergreinhardt/gatwd/comonads"
    "github.com/joergreinhardt/gatwd/pkg/cli"
    "github.com/joergreinhardt/gatwd/visualizer"
)

func main() {
    // Mint "gatwd-life": Core spec remains 100% clean!
    cli.NewEdition("gatwd-life").
        // 1. Register domain type constructor
        WithType("LifeGrid", parseLifePattern).
        // 2. Register custom transducers using ergonomic helpers
        WithStep(cli.ArgStep("parse_life", stepParseLifeWithDims)).
        WithStep(cli.MapStep("conway_step", stepConwayAdvance)).
        WithStep(cli.ArgStep("conway_render", stepConwayRender)).
        // 3. Launch CLI
        Run()
}
```

#### Step 2: Authoring the Simulation in Pure Gatwd Script (`game_of_life.gatwd`)
```gatwd
-- Declare domain type
type LifeGrid = Grid(Bool)

-- Define reusable macro pipelines
def step_and_render = conway_step |> conway_render braille cyan

-- Define ASCII pattern constants
const GLIDER_ASCII = "
#NAME Glider
.O.
..O
OOO
"

-- Execute Generation 0
spawn LifeGrid: GLIDER_ASCII |> parse_life 60 30 |> conway_render braille cyan

-- Advance Glider by 1 generation
spawn LifeGrid: GLIDER_ASCII |> parse_life 60 30 |> conway_step |> conway_render braille cyan

-- Advance Glider across an entire 4-generation cycle (diagonal translation)
spawn LifeGrid: GLIDER_ASCII |> parse_life 60 30 |> conway_step |> conway_step |> conway_step |> conway_step |> conway_render braille cyan
```

#### Step 3: Running the Edition
```bash
go run examples/game_of_life_edition/main.go examples/game_of_life_edition/game_of_life.gatwd
```

---

## 5. Architectural Takeaway: Extending Gatwd Cleanly

When extending Gatwd for domain-specific problems (cellular automata, quantum circuits, trading pipelines, machine learning):
1. **Do not modify `spec/spec.go` or core `evaluator/pipeline.go`**: Core Gatwd stays minimal, general, and mathematically pure.
2. **Use `cli.NewEdition(name)`**: Bind domain types via `.WithType(...)` and custom stream steps via `.WithStep(...)` using helpers like `cli.MapStep`, `cli.FilterStep`, and `cli.ArgStep`.
3. **Execute standard `.gatwd` scripts**: Users write familiar declarative pipelines while benefiting from domain transducers.
