# Conway's Game of Life — Minted Custom Edition (`gatwd-life`)

This directory contains a complete example of **extending Gatwd by minting a custom Edition** rather than modifying the core Gatwd codebase.

## Overview

Instead of adding domain-specific keywords or built-ins (`conway_step`, `conway_render`, `parse_life`) to the core compiler, this example uses Gatwd's fluent **Edition Builder** ([`cli.NewEdition`](file:///home/j/src/rnd/gatwd/pkg/cli/edition.go)) to mint a domain-specific runtime named **`gatwd-life`**.

The resulting runtime:
1. Keeps the core Gatwd language spec (`spec/spec.go`), parser, and standard library 100% clean and unpolluted.
2. Exposes domain types (`LifeGrid`) and domain transducers (`conway_step`, `conway_render`, `parse_life`) to DSL scripts.
3. Allows users to write simulations entirely in pure Gatwd DSL scripts ([`game_of_life.gatwd`](file:///home/j/src/rnd/gatwd/examples/game_of_life_edition/game_of_life.gatwd)).

---

## File Structure

- [`main.go`](file:///home/j/src/rnd/gatwd/examples/game_of_life_edition/main.go): The Go entrypoint that mints the `gatwd-life` Edition and registers custom types and step handlers.
- [`game_of_life.gatwd`](file:///home/j/src/rnd/gatwd/examples/game_of_life_edition/game_of_life.gatwd): The pure Gatwd DSL script that defines patterns, macro pipelines, and executes simulation generations.

---

## Running the Example

### 1. Execute the DSL script with the custom Edition
```bash
go run examples/game_of_life_edition/main.go examples/game_of_life_edition/game_of_life.gatwd
```

### 2. Launch the Interactive REPL for `gatwd-life`
```bash
go run examples/game_of_life_edition/main.go
```
Inside the interactive REPL:
```gatwd
gatwd-life ⊳ type LifeGrid = Grid(Bool)
gatwd-life ⊳ spawn LifeGrid: "3x3\n.O.\n..O\nOOO" |> parse_life 40 20 |> conway_render braille cyan
gatwd-life ⊳ spawn LifeGrid: "3x3\n.O.\n..O\nOOO" |> parse_life 40 20 |> conway_step |> conway_render braille cyan
```

---

## How It Works

### Step 1: Minting the Edition in Go (`main.go`)
```go
cli.NewEdition("gatwd-life").
    // 1. Register domain type factory
    WithType("LifeGrid", parseLifePattern).
    // 2. Register custom transducers using ergonomic helpers
    WithStep(cli.ArgStep("parse_life", stepParseLifeWithDims)).
    WithStep(cli.MapStep("conway_step", stepConwayAdvance)).
    WithStep(cli.ArgStep("conway_render", stepConwayRender)).
    // 3. Launch CLI runtime
    Run()
```

### Step 2: Using the Minted Transducers in the DSL (`game_of_life.gatwd`)
```gatwd
type LifeGrid = Grid(Bool)

-- Reusable macro pipeline
def step_and_render = conway_step |> conway_render braille cyan

const GLIDER_ASCII = "
#NAME Glider
.O.
..O
OOO
"

-- Run Generation 0
spawn LifeGrid: GLIDER_ASCII |> parse_life 60 30 |> conway_render braille cyan

-- Advance by 1 Generation
spawn LifeGrid: GLIDER_ASCII |> parse_life 60 30 |> conway_step |> conway_render braille cyan

-- Advance 4 full generations (one diagonal cycle)
spawn LifeGrid: GLIDER_ASCII |> parse_life 60 30 |> conway_step |> conway_step |> conway_step |> conway_step |> conway_render braille cyan
```
