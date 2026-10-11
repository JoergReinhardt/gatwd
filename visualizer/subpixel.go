package visualizer

import (
	"strings"
)

// SubpixelMode specifies the UTF-8 sub-character rendering resolution technique.
type SubpixelMode string

const (
	// ModeBraille uses Unicode Braille Patterns (U+2800..U+28FF) offering
	// 2x4 sub-pixel resolution (8 sub-pixels per terminal character cell).
	ModeBraille SubpixelMode = "braille"

	// ModeQuadrant uses Unicode Block Elements Quadrants (U+2596..U+259F) offering
	// 2x2 sub-pixel resolution (4 sub-pixels per terminal character cell).
	ModeQuadrant SubpixelMode = "quadrant"

	// ModeHalfBlock uses Unicode Upper/Lower Half Blocks (U+2580, U+2584) offering
	// 1x2 sub-pixel resolution (2 sub-pixels per terminal character cell).
	ModeHalfBlock SubpixelMode = "halfblock"

	// ModeCell uses full character blocks (1x1 resolution per character cell).
	ModeCell SubpixelMode = "cell"
)

// BrailleDotWeights maps (dx, dy) within a 2x4 cell to Unicode Braille bit offsets.
// Standard Unicode Braille ordering:
// Dot 1: (0,0) = 0x01,  Dot 4: (1,0) = 0x08
// Dot 2: (0,1) = 0x02,  Dot 5: (1,1) = 0x10
// Dot 3: (0,2) = 0x04,  Dot 6: (1,2) = 0x20
// Dot 7: (0,3) = 0x40,  Dot 8: (1,3) = 0x80
var BrailleDotWeights = [2][4]int{
	{0x01, 0x02, 0x04, 0x40},
	{0x08, 0x10, 0x20, 0x80},
}

// QuadrantRunes maps the 4-bit index (TL, TR, BL, BR) to Unicode Quadrant block characters.
var QuadrantRunes = [16]rune{
	' ', // 0000: empty
	'▘', // 0001: TL
	'▝', // 0010: TR
	'▀', // 0011: TL + TR (upper half)
	'▖', // 0100: BL
	'▌', // 0101: TL + BL (left half)
	'▞', // 0110: TR + BL (diagonal)
	'▛', // 0111: TL + TR + BL
	'▗', // 1000: BR
	'▚', // 1001: TL + BR (diagonal)
	'▐', // 1010: TR + BR (right half)
	'▜', // 1011: TL + TR + BR
	'▄', // 1100: BL + BR (lower half)
	'▙', // 1101: TL + BL + BR
	'▟', // 1110: TR + BL + BR
	'█', // 1111: full block
}

// SubpixelCanvas represents a 2D bitmap that renders to terminal strings
// using sub-character resolution techniques.
type SubpixelCanvas struct {
	Width  int
	Height int
	Mode   SubpixelMode
	grid   []bool
}

// NewSubpixelCanvas creates a new canvas with the specified pixel dimensions and sub-pixel mode.
func NewSubpixelCanvas(width, height int, mode SubpixelMode) *SubpixelCanvas {
	if mode == "" {
		mode = ModeBraille
	}
	return &SubpixelCanvas{
		Width:  width,
		Height: height,
		Mode:   mode,
		grid:   make([]bool, width*height),
	}
}

// Set sets the sub-pixel state at (x, y). Out of bounds writes are safely ignored.
func (c *SubpixelCanvas) Set(x, y int, alive bool) {
	if x >= 0 && x < c.Width && y >= 0 && y < c.Height {
		c.grid[y*c.Width+x] = alive
	}
}

// Get returns the sub-pixel state at (x, y). Out of bounds reads return false.
func (c *SubpixelCanvas) Get(x, y int) bool {
	if x >= 0 && x < c.Width && y >= 0 && y < c.Height {
		return c.grid[y*c.Width+x]
	}
	return false
}

// Clear sets all sub-pixels to false.
func (c *SubpixelCanvas) Clear() {
	for i := range c.grid {
		c.grid[i] = false
	}
}

// SubpixelDimensions returns the (xFactor, yFactor) sub-pixel resolution for the current mode.
func (c *SubpixelCanvas) SubpixelDimensions() (int, int) {
	switch c.Mode {
	case ModeBraille:
		return 2, 4
	case ModeQuadrant:
		return 2, 2
	case ModeHalfBlock:
		return 1, 2
	default:
		return 1, 1
	}
}

// TerminalSize returns the expected character columns and rows required to display this canvas.
func (c *SubpixelCanvas) TerminalSize() (cols int, rows int) {
	xFactor, yFactor := c.SubpixelDimensions()
	cols = (c.Width + xFactor - 1) / xFactor
	rows = (c.Height + yFactor - 1) / yFactor
	return cols, rows
}

// Render converts the 2D bitmap into a formatted multi-line UTF-8 string with optional ANSI coloring.
func (c *SubpixelCanvas) Render(color string) string {
	var sb strings.Builder
	xFactor, yFactor := c.SubpixelDimensions()
	charWidth := (c.Width + xFactor - 1) / xFactor
	charHeight := (c.Height + yFactor - 1) / yFactor

	for cy := 0; cy < charHeight; cy++ {
		for cx := 0; cx < charWidth; cx++ {
			var r rune
			switch c.Mode {
			case ModeBraille:
				mask := 0
				for dx := 0; dx < 2; dx++ {
					for dy := 0; dy < 4; dy++ {
						px := cx*2 + dx
						py := cy*4 + dy
						if c.Get(px, py) {
							mask |= BrailleDotWeights[dx][dy]
						}
					}
				}
				if mask == 0 {
					r = ' '
				} else {
					r = rune(0x2800 + mask)
				}

			case ModeQuadrant:
				idx := 0
				if c.Get(cx*2, cy*2) { // TL
					idx |= 1
				}
				if c.Get(cx*2+1, cy*2) { // TR
					idx |= 2
				}
				if c.Get(cx*2, cy*2+1) { // BL
					idx |= 4
				}
				if c.Get(cx*2+1, cy*2+1) { // BR
					idx |= 8
				}
				r = QuadrantRunes[idx]

			case ModeHalfBlock:
				top := c.Get(cx, cy*2)
				bottom := c.Get(cx, cy*2+1)
				if top && bottom {
					r = '█'
				} else if top {
					r = '▀'
				} else if bottom {
					r = '▄'
				} else {
					r = ' '
				}

			default: // ModeCell
				if c.Get(cx, cy) {
					r = '█'
				} else {
					r = ' '
				}
			}

			if color != "" && r != ' ' {
				sb.WriteString(applyColor(string(r), color, ""))
			} else {
				sb.WriteRune(r)
			}
		}
		if cy < charHeight-1 {
			sb.WriteRune('\n')
		}
	}

	return sb.String()
}

// Visualize implements the Visualizer interface so SubpixelCanvas can be
// formatted directly by visualizer.Format / StepVisualize.
func (c *SubpixelCanvas) Visualize(opts Options, prefix string, isLast bool, currentDepth int) string {
	rendered := c.Render(opts.Color)
	lines := strings.Split(rendered, "\n")
	var sb strings.Builder
	for i, line := range lines {
		if i == 0 {
			sb.WriteString(prefix + branch(isLast, opts.Style) + line + "\n")
		} else {
			sb.WriteString(prefix + extend(isLast, opts.Style) + line + "\n")
		}
	}
	return sb.String()
}

// RenderGridSubpixel renders a 2D boolean grid or slice of coordinates directly into UTF-8 sub-pixels.
func RenderGridSubpixel(width, height int, isLive func(x, y int) bool, mode SubpixelMode, color string) string {
	canvas := NewSubpixelCanvas(width, height, mode)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if isLive(x, y) {
				canvas.Set(x, y, true)
			}
		}
	}
	return canvas.Render(color)
}


