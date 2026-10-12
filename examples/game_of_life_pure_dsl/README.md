# Conway's Game of Life — 100% Pure Gatwd DSL Variant

This directory contains Conway's Game of Life implemented in **100% Pure Gatwd DSL Script**.

## Key Distinction

| Variant | Go Code Added | How It Runs |
| :--- | :---: | :--- |
| **Pure DSL (`game_of_life_pure_dsl`)** | **0 lines domain logic** | `go run ./examples/game_of_life_pure_dsl` or `gatwd examples/game_of_life_pure_dsl/game_of_life.gatwd` |
| **Minted Edition (`game_of_life_edition`)** | ~60 lines | `go run ./examples/game_of_life_edition ...` |
| **Native Appliance (`game_of_life`)** | Standalone tool | `go run ./examples/game_of_life ...` |

This variant requires **NO custom compiler editions**, **NO Go automata physics extensions**, and **NO domain-specific Go builtins**. All generational physics (B3/S23) run directly in the standard Gatwd DSL.

---

## Running the Simulation

Just like the other editions, `game_of_life_pure_dsl` provides a full, interactive terminal runner:

### 1. Interactive Terminal Animation (Default)
Runs an infinite real-time animation loop rendered in high-resolution UTF-8 Braille sub-pixels, computed generation-by-generation entirely by the Gatwd DSL engine:
```bash
go run ./examples/game_of_life_pure_dsl
```

### 2. Built-in Pattern Presets
```bash
# Classic spaceship gliding across the universe
go run ./examples/game_of_life_pure_dsl -preset glider

# Period-2 oscillators
go run ./examples/game_of_life_pure_dsl -preset blinker
go run ./examples/game_of_life_pure_dsl -preset toad
go run ./examples/game_of_life_pure_dsl -preset beacon

# Period-3 geometric pulsar
go run ./examples/game_of_life_pure_dsl -preset pulsar

# Methuselahs (long-running active chaotic evolutions)
go run ./examples/game_of_life_pure_dsl -preset acorn
go run ./examples/game_of_life_pure_dsl -preset r-pentomino
```

### 3. Rendering Modes & Custom Flags
```bash
# Quadrant blocks (2x2) with yellow coloring
go run ./examples/game_of_life_pure_dsl -preset pulsar -mode quadrant -color yellow

# Full character blocks (1x1) with 100ms frame delay
go run ./examples/game_of_life_pure_dsl -preset glider -mode cell -delay 100

# Batch mode: run 10 generations without clearing screen
go run ./examples/game_of_life_pure_dsl -preset glider -steps 10 -non-interactive
```

### 4. Mathematical Verification Suite
Execute the full mathematical verification suite asserting discrete spatial invariants and multi-generational lifecycles:
```bash
go run ./examples/game_of_life_pure_dsl -verify
```
or via the standard Gatwd CLI binary:
```bash
gatwd examples/game_of_life_pure_dsl/game_of_life.gatwd
```

---

## Architectural Principles

### 1. Dynamic Generational Transitions in Pure DSL
Rather than hardcoding static coordinate transitions, the entire simulation dynamically computes Generation $t+1$ from Generation $t$ directly in Gatwd syntax using Moore neighborhood expansion, multiset frequencies, and B3/S23 rule filtering:

```gatwd
-- 1. Moore Neighborhood Expansion via Multi-Generator List Comprehension:
let live = [[2, 1], [2, 2], [2, 3]]
let nbrs = [[x + dx, y + dy] | [x, y] <- live, dx <- [-1, 0, 1], dy <- [-1, 0, 1], (dx != 0 || dy != 0)]

-- 2. Candidate Cell Neighbor Counts via Persistent HAMT Frequencies:
let pairs = nbrs |> unchunk |> frequencies |> hamt_pairs |> collect_list

-- 3. Conway Rules Application (Birth: count == 3, Survival: count == 2 & was alive):
let births = [pt | [pt, cnt] <- pairs, cnt == 3]
let survivors = [p | [pt, cnt] <- pairs, p <- live, cnt == 2, p == pt]
let nextGen = births ++ survivors
```

### 2. Multi-Generational Lifecycles
The verification script verifies two fundamental Game of Life patterns dynamically across multiple steps:

1. **Blinker Oscillator (Period-2)**:
   - **Generation 0**: Horizontal `[[2, 1], [2, 2], [2, 3]]`
   - **Generation 1**: Dynamically computed vertical `[[1, 2], [3, 2], [2, 2]]`
   - **Generation 2**: Dynamically computed horizontal `[[2, 1], [2, 3], [2, 2]]`
2. **Glider Spaceship (Period-4 Diagonal Shift)**:
   - **Generation 0**: 5-cell glider at `[[2, 1], [3, 2], [1, 3], [2, 3], [3, 3]]`
   - **Generations 1, 2, 3**: Iteratively evolved step-by-step
   - **Generation 4**: Evolved to `[[3, 2], [4, 3], [2, 4], [3, 4], [4, 4]]` (+1, +1 diagonal shift) with population 5 verified

### 3. Sparse Coordinate Sets & Invariant Testing
Living cells are checked for membership and spatial invariants using persistent sets:

```gatwd
-- Verify cell membership in sparse set
blinker1 |> unchunk |> to_set |> set_contains [1 2] |> assert true
blinker1 |> unchunk |> to_set |> set_contains [2 1] |> assert false

-- Population count invariant
glider4 |> unchunk |> fold count |> assert 5
```
