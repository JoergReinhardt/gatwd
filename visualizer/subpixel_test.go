package visualizer

import (
	"strings"
	"testing"
)

func TestSubpixelCanvas_Braille(t *testing.T) {
	// A 2x4 pixel canvas should render to exactly 1 Braille character
	canvas := NewSubpixelCanvas(2, 4, ModeBraille)
	cols, rows := canvas.TerminalSize()
	if cols != 1 || rows != 1 {
		t.Fatalf("expected 1x1 terminal size, got %dx%d", cols, rows)
	}

	// Empty canvas
	out := canvas.Render("")
	if out != " " {
		t.Errorf("expected empty space, got %q", out)
	}

	// Dot 1: (0, 0) -> 0x2801
	canvas.Set(0, 0, true)
	out = canvas.Render("")
	expectedRune := rune(0x2800 + 0x01)
	if out != string(expectedRune) {
		t.Errorf("expected %q, got %q", string(expectedRune), out)
	}

	// Full 8 dots -> 0x28FF
	for x := 0; x < 2; x++ {
		for y := 0; y < 4; y++ {
			canvas.Set(x, y, true)
		}
	}
	out = canvas.Render("")
	if out != string(rune(0x28FF)) {
		t.Errorf("expected 0x28FF full braille character, got %q", out)
	}
}

func TestSubpixelCanvas_Quadrant(t *testing.T) {
	// A 2x2 pixel canvas should render to exactly 1 Quadrant character
	canvas := NewSubpixelCanvas(2, 2, ModeQuadrant)
	canvas.Set(0, 0, true) // TL
	canvas.Set(1, 1, true) // BR
	out := canvas.Render("")
	// TL (bit 0 = 1) + BR (bit 3 = 8) -> index 9 -> '▚'
	if out != "▚" {
		t.Errorf("expected diagonal quadrant '▚', got %q", out)
	}
}

func TestSubpixelCanvas_HalfBlock(t *testing.T) {
	canvas := NewSubpixelCanvas(1, 2, ModeHalfBlock)
	canvas.Set(0, 0, true) // top only
	out := canvas.Render("")
	if out != "▀" {
		t.Errorf("expected upper half block '▀', got %q", out)
	}

	canvas.Set(0, 1, true) // both
	out = canvas.Render("")
	if out != "█" {
		t.Errorf("expected full block '█', got %q", out)
	}
}

func TestSubpixelCanvas_MultiLine(t *testing.T) {
	canvas := NewSubpixelCanvas(4, 8, ModeBraille)
	cols, rows := canvas.TerminalSize()
	if cols != 2 || rows != 2 {
		t.Fatalf("expected 2x2 terminal size for 4x8 braille canvas, got %dx%d", cols, rows)
	}

	canvas.Set(0, 0, true)
	canvas.Set(2, 4, true)
	out := canvas.Render("")
	lines := strings.Split(out, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
}
