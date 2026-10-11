# Conway's Game of Life — 100% Pure Gatwd DSL Variant

This directory contains Conway's Game of Life implemented in **100% Pure Gatwd DSL Script**.

## Key Distinction

| Variant | Go Code Added | How It Runs |
| :--- | :---: | :--- |
| **Pure DSL (`game_of_life_pure_dsl`)** | **0 lines** | `gatwd examples/game_of_life_pure_dsl/game_of_life.gatwd` |
| **Minted Edition (`game_of_life_edition`)** | ~60 lines | `go run examples/game_of_life_edition/main.go ...` |
| **Native Appliance (`game_of_life`)** | Standalone tool | `go run ./examples/game_of_life ...` |

This variant requires **NO Go compiler**, **NO Go runtime code additions**, and **NO minted Edition binary**. It runs on the plain, out-of-the-box `gatwd` CLI binary.

---

## Architectural Principles

### 1. Sparse Coordinate Sets
Instead of maintaining a dense 2D matrix array, the simulation models the unbounded universe as an immutable mathematical set of coordinate strings:

```gatwd
type CellSet = Set(String)
```

Living cells are stored as `"x:y"` keys in an immutable Hash Array Mapped Trie (HAMT) backing the `Set`.

### 2. Population Measurement via Catamorphisms
Population counts across discrete generational steps are computed using Gatwd's catamorphic stream folds:

```gatwd
spawn CellSet: Start |> emit 2:1, 2:2, 2:3 |> drop 1 |> to_set |> set_elements |> fold count |> assert 3
```

### 3. Spatial Membership & Invariant Testing
Cell existence, births, and deaths are validated using persistent set membership checks and assertions:

```gatwd
-- Check that live center cell survived
spawn CellSet: LiveGrid |> set_contains 2:2 |> assert true

-- Check that extinguished cell died from underpopulation
spawn CellSet: LiveGrid |> set_contains 2:1 |> assert false
```

### 4. Diagnostic Card Visualization
Simulation generation cards and invariant summaries are rendered using the standard Gatwd card visualizer (`visualize`):

```gatwd
spawn Status: Start |> echo "All Pure DSL Game of Life Invariants Verified!" |> visualize --curved --border --border-color=green
```

---

## Running the Script

Execute the script with the standard Gatwd CLI:

```bash
gatwd examples/game_of_life_pure_dsl/game_of_life.gatwd
```

or via `go run cmd/gatwd/main.go`:

```bash
go run cmd/gatwd/main.go examples/game_of_life_pure_dsl/game_of_life.gatwd
```
