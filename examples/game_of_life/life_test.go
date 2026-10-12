package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/joergreinhardt/gatwd/visualizer"
)

func TestLoader_ParseASCII(t *testing.T) {
	ascii := `#NAME Test Blinker
#RULE B3/S23
#SIZE 5x5
.....
..O..
..O..
..O..
.....
`
	pat, err := ParseASCII(ascii)
	if err != nil {
		t.Fatalf("failed to parse ASCII: %v", err)
	}
	if pat.Name != "Test Blinker" {
		t.Errorf("expected name 'Test Blinker', got %q", pat.Name)
	}
	if pat.Width != 5 || pat.Height != 5 {
		t.Errorf("expected 5x5 dimensions, got %dx%d", pat.Width, pat.Height)
	}
	if pat.Population() != 3 {
		t.Errorf("expected 3 living cells, got %d", pat.Population())
	}

	grid := pat.PlaceOnGrid(5, 5, false, false)
	if !grid.Get(2, 1) || !grid.Get(2, 2) || !grid.Get(2, 3) {
		t.Errorf("blinker cells not placed accurately")
	}

	// 1 comonadic step
	nextGrid := StepConway(grid)
	if !nextGrid.Get(1, 2) || !nextGrid.Get(2, 2) || !nextGrid.Get(3, 2) {
		t.Errorf("blinker did not oscillate horizontally")
	}
}

func TestPresets_AllAvailable(t *testing.T) {
	presetKeys := []string{"gun-and-eater", "gosper-gun", "pulsar", "acorn", "r-pentomino", "glider"}
	for _, key := range presetKeys {
		pat, err := GetPreset(key)
		if err != nil {
			t.Errorf("failed to load preset %s: %v", key, err)
			continue
		}
		if pat.Population() == 0 {
			t.Errorf("preset %s has 0 population", key)
		}
		grid := pat.PlaceOnGrid(60, 40, true, true)
		nextGrid := StepConway(grid)
		if nextGrid == nil {
			t.Errorf("StepConway failed on preset %s", key)
		}
	}
}

func TestLoader_FileLoading(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "custom.life")
	content := `#NAME Custom Oscillator
#SIZE 3x3
.O.
.O.
.O.
`
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	pat, err := LoadFile(filePath)
	if err != nil {
		t.Fatalf("LoadFile failed: %v", err)
	}
	if pat.Name != "Custom Oscillator" {
		t.Errorf("expected name 'Custom Oscillator', got %q", pat.Name)
	}
	if pat.Population() != 3 {
		t.Errorf("expected population 3, got %d", pat.Population())
	}
}

func TestSubpixel_BrailleRendering(t *testing.T) {
	pat, err := GetPreset("glider")
	if err != nil {
		t.Fatalf("failed to load glider: %v", err)
	}
	grid := pat.PlaceOnGrid(10, 10, false, false)

	canvas := visualizer.NewSubpixelCanvas(10, 10, visualizer.ModeBraille)
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			if grid.Get(x, y) {
				canvas.Set(x, y, true)
			}
		}
	}

	rendered := canvas.Render("")
	if len(strings.TrimSpace(rendered)) == 0 {
		t.Errorf("canvas rendered empty string for glider")
	}
}
