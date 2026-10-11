# Gatwd Cellular Automaton: Conway's Game of Life

A high-performance standalone edition of **Conway's Game of Life** implemented in Go on top of the **Gatwd** functional ecosystem. It features a mathematically rigorous **Comonadic 2D Torus Grid Engine**, ultra-high-resolution **UTF-8 Sub-Character Rasterization** (using 2×4 Braille and 2×2 Quadrants), and full configuration via ASCII pattern files and presets.

```
╭─────────────────────────────────────────────────────────────────────────────╮
│ Gatwd Game of Life | Preset: gun-and-eater | Gen: 64 | Pop: 48 | Mode: braille │
╰─────────────────────────────────────────────────────────────────────────────╯
      ⣀   ⣀⠄⠆    ⣀                    
⠶    ⡎ ⠌⡦ ⠛⠄⡄    ⠉                    
     ⠈⠒⠈                              
                                      
                         ⣀            
                         ⠁⠧           
∷ [Gatwd Life] Gen 64 | Pop 48 | Mode: braille (2x4) | Press Ctrl+C to Exit ∷
```

---

## 1. Mathematical & Architectural Foundation: The Grid Comonad

In Category Theory and Functional Programming, Cellular Automata are the canonical example of a **Comonad**. While Monads model computations that produce effects, Comonads model computations that depend on **local spatial context**.

A 2D cellular automaton space is formalized as a comonad $W$:

$$\begin{aligned}
\text{Extract} &: W[A] \to A && \text{(counit: inspect the focused cell)} \\
\text{Extend} &: (W[A] \to B) \to W[A] \to W[B] && \text{(cobind: evaluate neighborhood across all space)}
\end{aligned}$$

### The Local Rule as a Co-Kleisli Arrow

In standard imperative implementations, two 2D arrays are mutated in a loop with manual index bounds checking. In Gatwd, the step from generation $t$ to $t+1$ is a pure co-Kleisli extension:

```go
// 1. Local rule receives a Comonadic cursor (GridZipper) and yields next state
func ConwayRule(z *comonads.GridZipper[bool]) bool {
    aliveNeighbors := 0
    for _, neighbor := range z.MooreNeighbors() {
        if neighbor {
            aliveNeighbors++
        }
    }
    currentlyAlive := z.Extract()
    if currentlyAlive {
        return aliveNeighbors == 2 || aliveNeighbors == 3
    }
    return aliveNeighbors == 3
}

// 2. Extend slides the local rule across the entire universe
nextGrid := comonads.Extend(grid.ZipperAt(0, 0), ConwayRule)
```

The comonad laws guarantee mathematical consistency:
1. **Left Identity**: $\text{Extend}(\text{Extract}, z) = z$
2. **Right Identity**: $\text{Extract}(\text{Extend}(f, z)) = f(z)$
3. **Associativity**: $\text{Extend}(f, \text{Extend}(g, z)) = \text{Extend}(\lambda w'. f(\text{Extend}(g, w')), z)$

---

## 2. UTF-8 Sub-Character Resolution

Standard terminal applications render 1 cell per character position (`█` or ` `), severely limiting the visible simulation size to ~80×24 cells on standard terminals.

Gatwd introduces **Sub-Character Rasterization** in [`visualizer.SubpixelCanvas`](file:///home/j/src/rnd/gatwd/visualizer/subpixel.go):

| Mode | Sub-Pixel Ratio | Sub-Pixels per Char | Unicode Block | Resolution Multiplier |
| :--- | :---: | :---: | :--- | :---: |
| **`braille`** (default) | **2 × 4** | **8 dots** | `U+2800..U+28FF` (Braille Patterns) | **8× higher resolution** |
| **`quadrant`** | **2 × 2** | **4 blocks** | `U+2596..U+259F` (Block Elements) | **4× higher resolution** |
| **`halfblock`** | **1 × 2** | **2 blocks** | `U+2580, U+2584` (Half Blocks) | **2× higher resolution** |
| **`cell`** | **1 × 1** | **1 block** | Standard character cells | **1× baseline** |

### Braille 2×4 Dot Encoding Math

Unicode Braille patterns occupy code points `U+2800` through `U+28FF`. Each terminal character represents an array of 2 columns by 4 rows of sub-pixels:

```
Sub-pixel (0,0): Dot 1 (0x01)    Sub-pixel (1,0): Dot 4 (0x08)
Sub-pixel (0,1): Dot 2 (0x02)    Sub-pixel (1,1): Dot 5 (0x10)
Sub-pixel (0,2): Dot 3 (0x04)    Sub-pixel (1,2): Dot 6 (0x20)
Sub-pixel (0,3): Dot 7 (0x40)    Sub-pixel (1,3): Dot 8 (0x80)
```

Given pixel coordinates $(x, y)$, the target terminal character cell $(cx, cy)$ is:
$$cx = \lfloor x / 2 \rfloor, \quad cy = \lfloor y / 4 \rfloor$$
The intra-cell dot offset is calculated via bitwise OR:
$$\text{Rune} = 0x2800 + \sum_{dx=0}^{1} \sum_{dy=0}^{3} \text{Pixel}(cx \cdot 2 + dx, cy \cdot 4 + dy) \times \text{Weight}(dx, dy)$$

An 80×25 terminal window can thus simulate a **160×100 cellular universe** in crisp, high-density resolution!

---

## 3. Presets & The Default Emitter-Sink Engine

### Default: Gosper Glider Gun + Eater 1 Sink (`gun-and-eater`)

By default, the appliance runs a perpetual **Emitter-Sink oscillator**:
1. **Gosper Glider Gun** (discovered by Bill Gosper in 1970): The left engine produces a new Glider every 30 generations, shooting it diagonally down-right.
2. **Eater 1 Sink** (fishhook still life): Positioned in the path of the incoming gliders. When a glider strikes the tip of the Eater, the Eater temporarily deforms, consumes the glider, and returns to its pristine still-life state in 4 generations.
3. The result is a perpetual, stable cosmic machine that runs infinitely without filling the screen with debris!

### Additional Presets

| Preset Name | Description | Characteristics |
| :--- | :--- | :--- |
| `gun-and-eater` | Gosper Glider Gun + Eater 1 sink (default) | Infinite emitter-sink loop |
| `gosper-gun` | Standalone Gosper Glider Gun | Emits gliders indefinitely |
| `pulsar` | Classic period-3 oscillator | Highly symmetric breathing pattern |
| `acorn` | Famous 7-cell methuselah | Stabilizes after 5,206 generations, creating 13 gliders |
| `r-pentomino` | 5-cell active methuselah | Active for 1,103 generations |
| `glider` | Elementary diagonal spaceship | Translates (1, 1) every 4 generations |

---

## 4. ASCII Pattern File Format

Custom patterns can be loaded via the `-file` flag using standard `.life` ASCII format.

### Directives
- `#NAME <string>`: Name displayed in the UI banner.
- `#DESCRIPTION <string>`: Description of the automaton.
- `#SIZE <width>x<height>`: Recommended simulation dimensions.
- `#RULE <rule>`: Cellular automaton rule (default: `B3/S23`).
- `#` or `!`: Comment lines.

### Cell Representation
- Dead cells: `.` or ` ` (space) or `-` or `_`
- Alive cells: `O`, `*`, `#`, `x`, `X`, or `1`

### Example File (`my_pulsar.life`):
```
#NAME My Pulsar Oscillator
#SIZE 17x17
.................
...OOO...OOO.....
.................
.O....O.O....O...
.O....O.O....O...
.O....O.O....O...
...OOO...OOO.....
.................
...OOO...OOO.....
.O....O.O....O...
.O....O.O....O...
.O....O.O....O...
.................
...OOO...OOO.....
.................
```

---

## 5. Usage & CLI Reference

### Running the Default Emitter-Sink Appliance
```bash
go run ./examples/game_of_life
```

### Selecting Built-In Presets
```bash
# Run the classic Pulsar period-3 oscillator
go run ./examples/game_of_life -preset pulsar

# Run the 5,206-generation Acorn methuselah
go run ./examples/game_of_life -preset acorn -delay 25
```

### Loading a Custom ASCII File
```bash
go run ./examples/game_of_life -file examples/game_of_life/patterns/gun_and_eater.life
```

### Changing Sub-Character Resolution Modes
```bash
# High-density Braille (2x4 dots per cell) [Default]
go run ./examples/game_of_life -mode braille

# Quadrant blocks (2x2 sub-pixels per cell)
go run ./examples/game_of_life -preset pulsar -mode quadrant

# Half blocks (1x2 sub-pixels per cell)
go run ./examples/game_of_life -mode halfblock

# Full blocks (1x1 standard cell)
go run ./examples/game_of_life -mode cell
```

### All Command-Line Flags

| Flag | Default | Description |
| :--- | :--- | :--- |
| `-file <path>` | `""` | Path to custom ASCII pattern file (`.life` or `.txt`) |
| `-preset <name>` | `"gun-and-eater"` | Preset to load (`gun-and-eater`, `gosper-gun`, `pulsar`, `acorn`, `r-pentomino`, `glider`) |
| `-mode <mode>` | `"braille"` | Sub-character mode: `braille` (2x4), `quadrant` (2x2), `halfblock` (1x2), `cell` (1x1) |
| `-width <int>` | `auto` | Simulation grid width in sub-pixels |
| `-height <int>` | `auto` | Simulation grid height in sub-pixels |
| `-delay <int>` | `60` | Delay between generations in milliseconds |
| `-steps <int>` | `0` | Number of generations to run (`0` = infinite interactive loop) |
| `-color <string>` | `"green"` | Living cell ANSI color: `green`, `cyan`, `yellow`, `magenta`, `blue`, `white` |
| `-torus` | `true` | Enable toroidal wrapping (edges wrap around to opposite side) |
| `-center` | `true` | Center pattern inside the grid canvas |
| `-border` | `true` | Render decorative UTF-8 box border |
| `-non-interactive` | `false` | Batch mode: disable screen clear and cursor manipulation (useful for CI/tests) |
| `-list-presets` | `false` | Print all available built-in presets and exit |

---

## 6. Testing

The Game of Life appliance includes comprehensive unit tests verifying the Comonad laws, Moore neighborhood traversal, oscillator and spaceship kinematics, sub-character Braille rasterization, and ASCII file parsing:

```bash
go test -v ./examples/game_of_life
```
