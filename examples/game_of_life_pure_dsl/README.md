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

The rule and loop are defined in [`game_of_life.gatwd`](game_of_life.gatwd):

```gatwd
-- Wrap a coordinate onto the torus (also for -1 and size).
def wrap = \n size -> (n % size + size) % size

-- The 8 offsets of a Moore neighbourhood (everything but the centre).
let offsets = [[dx, dy] | dx <- [-1, 0, 1], dy <- [-1, 0, 1], (dx != 0 || dy != 0)]

-- One entry per (live cell, neighbour) pair: the wrapped neighbours of every live cell.
def neighbours = \w h live ->
    [[(wrap (x + dx) w), (wrap (y + dy) h)] | [x, y] <- live, [dx, dy] <- offsets]

-- Count occurrences: a HAMT cell -> number of live neighbours.
def tally = \cells -> fold (\acc c -> hamt_set c ((hamt_get_or 0 c acc) + 1) acc) hamt_empty cells

def births = \pairs -> [pt | [pt, n] <- pairs, n == 3]
def survivors = \live pairs -> [p | [pt, n] <- pairs, p <- live, n == 2, p == pt]

-- B3/S23: births plus survivors, from the neighbour counts.
def next_from = \live pairs -> (births pairs) ++ (survivors live pairs)
def next_gen = \w h live -> next_from live (hamt_pairs (tally (neighbours w h live)))

def conway_step = match { [live, w, h] -> (next_gen w h live) }

-- Iterate conway_step n times on a w x h torus, purely by recursion (ackermann-shaped).
def evolve = match {
    [state, 0, w, h] -> state
    | [state, n, w, h] ->
        let next = [state, w, h] |> conway_step
        in [next, n - 1, w, h] |> evolve
}
```

A board is a sparse list of live coordinate pairs; births are empty cells with exactly 3 live neighbours, survivors are live cells with exactly 2 (cells with 3 already appear as births, giving standard B3/S23).

**Bounded torus.** Neighbour coordinates are reduced modulo `w`/`h`, so the world is a finite `w x h` torus: the board stays within `[0,w) x [0,h)` and wraps at the edges instead of drifting outward forever. This keeps it lively indefinitely — an earlier unbounded version let expanding patterns drift into a region where the sparse rule misbehaved, and they collapsed after ~130 generations.

---

## Running

### Animate (default: Edna, a methuselah, in a 64×42 frame)
```bash
go run ./examples/game_of_life_pure_dsl
```
Infinite real-time animation rendered in UTF-8 braille sub-pixels, each generation computed by the DSL (`let live = [live, w, h] |> conway_step` on the 64×42 torus).

### Define the startup grid by file
The startup grid is a pure-gatwd file that binds `live`. The shipped default is [`patterns/edna.gatwd`](patterns/edna.gatwd) — the **Edna** methuselah (149 cells, lifespan 31,192 generations; from [gameoflifelab.com/patterns](https://www.gameoflifelab.com/patterns)), decoded from its canonical RLE and **centered in the 64×42 torus**.
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

- **Bounded torus.** The rule is a sparse coordinate set wrapped modulo the grid, so the world is a finite `w x h` torus with no array allocation. `-width`/`-height` set that torus (default 64×42); the viewer injects them into the rule, so the simulation and the rendered frame always match.
- **Why toroidal.** An earlier unbounded-plane version let expanding patterns drift into a region where the sparse rule misbehaved, collapsing a methuselah like Edna to zero after ~130 generations. Wrapping keeps coordinates bounded and the board lively indefinitely.
- **Single source of truth.** `main.go` loads `conway_step` from `game_of_life.gatwd`; the physics is never duplicated in Go.
