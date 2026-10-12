# Conway's Game of Life — 100% Pure Gatwd DSL Variant

Conway's Game of Life with the **rule and the generational loop both written in pure Gatwd DSL**. The companion Go program (`main.go`) is only a *viewer*: it loads the rule from the `.gatwd` file, seeds a starting grid, and renders frames. No Conway logic lives in Go.

## Key Distinction

| Variant | Go domain logic | How it runs |
| :--- | :---: | :--- |
| **Pure DSL (`game_of_life_pure_dsl`)** | **none** — rule + loop are DSL | `gatwd examples/game_of_life_pure_dsl/game_of_life.gatwd`, or `go run ./examples/game_of_life_pure_dsl` to animate |
| **Minted Edition (`game_of_life_edition`)** | ~60 lines (Go `conway_step`) | `go run ./examples/game_of_life_edition ...` |
| **Native Appliance (`game_of_life`)** | Standalone Go tool | `go run ./examples/game_of_life ...` |

The rule is defined **once**, in `game_of_life.gatwd`. The viewer loads that same file, so there is no second copy of the physics anywhere.

---

## The pure-DSL implementation

Two definitions in [`game_of_life.gatwd`](game_of_life.gatwd):

```gatwd
-- One B3/S23 generation over a sparse set of live [x, y] cells.
def conway_step = match {
    live -> let nbrs = [[x + dx, y + dy] | [x, y] <- live, dx <- [-1, 0, 1], dy <- [-1, 0, 1], (dx != 0 || dy != 0)]
            in let pairs = nbrs |> frequencies |> hamt_pairs |> collect_list
            in let births = [pt | [pt, cnt] <- pairs, cnt == 3]
            in let survivors = [p | [pt, cnt] <- pairs, p <- live, cnt == 2, p == pt]
            in births ++ survivors
}

-- Iterate conway_step n times, purely by recursion (same shape as ackermann).
def evolve = match {
    [state, 0] -> state
    | [state, n] -> let next = state |> to_chunk |> conway_step in [next, n - 1] |> evolve
}
```

A board is a sparse list of live coordinate pairs on an unbounded integer plane: births are empty cells with exactly 3 live neighbours; survivors are live cells with exactly 2 (cells with 3 already appear as births, giving standard B3/S23).

**Why `to_chunk`:** a value bound by a `match` pattern flows as a *stream* of its elements, so piping the board straight into `conway_step` would apply the rule cell-by-cell. `to_chunk` re-collapses the streamed state into a single board before each step. (At the top level, `let live = …` already binds a whole board, which is why single-step use needs no `to_chunk`.)

---

## Running

### Animate (default: Edna, a methuselah, in a 64×42 frame)
```bash
go run ./examples/game_of_life_pure_dsl
```
Infinite real-time animation rendered in UTF-8 braille sub-pixels, each generation computed by the DSL (`let live = live |> conway_step`).

### Define the startup grid by file
The startup grid is a pure-gatwd file that binds `live`. The shipped default is [`patterns/edna.gatwd`](patterns/edna.gatwd) — the **Edna** methuselah (149 cells, lifespan 31,192 generations; from [gameoflifelab.com/patterns](https://www.gameoflifelab.com/patterns)), decoded from its canonical RLE and **pre-centered in the 64×42 frame so it has room to expand** (patterns seeded at the origin can collapse).
```bash
# Use a different startup-grid file (any .gatwd that binds `live`)
go run ./examples/game_of_life_pure_dsl -pattern path/to/grid.gatwd

# Fall back to a built-in preset instead of a file
go run ./examples/game_of_life_pure_dsl -pattern= -preset pulsar
```
A startup-grid file is just:
```gatwd
let live = [[22, 11], [25, 11], ...]
```

### Presets, rendering modes, frame size
```bash
go run ./examples/game_of_life_pure_dsl -pattern= -preset glider
go run ./examples/game_of_life_pure_dsl -mode cell -color yellow      # 1x1 blocks
go run ./examples/game_of_life_pure_dsl -width 80 -height 50          # bigger frame
go run ./examples/game_of_life_pure_dsl -steps 40 -non-interactive    # batch, no clear
```
Flags: `-pattern`, `-preset`, `-mode` (braille/quadrant/halfblock/cell), `-color`, `-width`, `-height`, `-torus`, `-delay`, `-steps`, `-non-interactive`.

### Verify the pure DSL (no Go)
```bash
gatwd examples/game_of_life_pure_dsl/game_of_life.gatwd    # asserts blinker period-2, glider survives
```

---

## Notes

- **Sparse, unbounded plane.** The rule uses integer coordinates with no array/grid; the `-width`/`-height` "frame" is a rendering window (toroidal by default), not a limit on the computation.
- **Room to expand.** Seed patterns away from the origin (the shipped files are centered) so an expanding pattern has space; a pattern crowded at the origin can die prematurely.
- **Single source of truth.** `main.go` loads `conway_step` from `game_of_life.gatwd`; the physics is never duplicated in Go.
