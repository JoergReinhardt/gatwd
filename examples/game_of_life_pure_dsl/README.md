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
-- One B3/S23 generation on a w x h torus (neighbour coords taken modulo the grid).
def conway_step = match {
    [live, w, h] -> let nbrs = [[((x + dx) % w + w) % w, ((y + dy) % h + h) % h] | [x, y] <- live, dx <- [-1, 0, 1], dy <- [-1, 0, 1], (dx != 0 || dy != 0)]
            in let pairs = nbrs |> frequencies |> hamt_pairs |> collect_list
            in let births = [pt | [pt, cnt] <- pairs, cnt == 3]
            in let survivors = [p | [pt, cnt] <- pairs, p <- live, cnt == 2, p == pt]
            in births ++ survivors
}

-- Iterate conway_step n times on a w x h torus, purely by recursion (ackermann-shaped).
def evolve = match {
    [state, 0, w, h] -> state
    | [state, n, w, h] -> let next = [state, w, h] |> conway_step in [next, n - 1, w, h] |> evolve
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
