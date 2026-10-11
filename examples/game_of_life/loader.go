package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/joergreinhardt/gatwd/comonads"
)

// Pattern holds the parsed ASCII cellular automaton pattern and metadata.
type Pattern struct {
	Name        string
	Description string
	Rule        string
	DeclaredW   int
	DeclaredH   int
	Width       int
	Height      int
	Cells       [][]bool
}

// LoadFile reads an ASCII .life or text file and parses its metadata and cell grid.
func LoadFile(path string) (*Pattern, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading pattern file %s: %w", path, err)
	}
	return ParseASCII(string(data))
}

// ParseASCII parses an ASCII formatted Game of Life string.
func ParseASCII(content string) (*Pattern, error) {
	pat := &Pattern{
		Rule: "B3/S23",
	}

	scanner := bufio.NewScanner(strings.NewReader(content))
	var rawRows []string

	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\r\n")
		trimmed := strings.TrimSpace(line)

		// Metadata & comments
		if strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "!") {
			upper := strings.ToUpper(trimmed)
			if strings.HasPrefix(upper, "#NAME ") {
				pat.Name = strings.TrimSpace(line[6:])
			} else if strings.HasPrefix(upper, "#DESCRIPTION ") {
				pat.Description = strings.TrimSpace(line[13:])
			} else if strings.HasPrefix(upper, "#RULE ") {
				pat.Rule = strings.TrimSpace(line[6:])
			} else if strings.HasPrefix(upper, "#SIZE ") {
				sizeStr := strings.TrimSpace(line[6:])
				var w, h int
				if strings.Contains(sizeStr, "x") {
					fmt.Sscanf(sizeStr, "%dx%d", &w, &h)
				} else {
					fmt.Sscanf(sizeStr, "%d %d", &w, &h)
				}
				pat.DeclaredW = w
				pat.DeclaredH = h
			}
			continue
		}

		if line == "" && len(rawRows) == 0 {
			continue // Skip leading empty lines
		}

		rawRows = append(rawRows, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Determine bounding box
	pat.Height = len(rawRows)
	for _, row := range rawRows {
		if len(row) > pat.Width {
			pat.Width = len(row)
		}
	}

	if pat.Width == 0 || pat.Height == 0 {
		return nil, fmt.Errorf("empty or invalid pattern grid")
	}

	pat.Cells = make([][]bool, pat.Height)
	for y, row := range rawRows {
		pat.Cells[y] = make([]bool, pat.Width)
		for x := 0; x < len(row); x++ {
			ch := row[x]
			// Alive chars: O, *, #, x, X, 1
			if ch == 'O' || ch == '*' || ch == '#' || ch == 'x' || ch == 'X' || ch == '1' {
				pat.Cells[y][x] = true
			}
		}
	}

	return pat, nil
}

// PlaceOnGrid constructs a comonadic Grid[bool] with the pattern placed at
// the center (or specified offset) of a target (width, height) universe.
func (p *Pattern) PlaceOnGrid(targetW, targetH int, center bool, torus bool) *comonads.Grid[bool] {
	if targetW <= 0 {
		if p.DeclaredW > 0 {
			targetW = p.DeclaredW
		} else {
			targetW = p.Width + 4
		}
	}
	if targetH <= 0 {
		if p.DeclaredH > 0 {
			targetH = p.DeclaredH
		} else {
			targetH = p.Height + 4
		}
	}

	offsetX := 0
	offsetY := 0
	if center {
		offsetX = (targetW - p.Width) / 2
		offsetY = (targetH - p.Height) / 2
		if offsetX < 0 {
			offsetX = 0
		}
		if offsetY < 0 {
			offsetY = 0
		}
	}

	return comonads.NewGrid(targetW, targetH, torus, func(x, y int) bool {
		px := x - offsetX
		py := y - offsetY
		if py >= 0 && py < p.Height && px >= 0 && px < p.Width {
			return p.Cells[py][px]
		}
		return false
	})
}

// Population counts the number of living cells in the pattern.
func (p *Pattern) Population() int {
	count := 0
	for y := 0; y < p.Height; y++ {
		for x := 0; x < p.Width; x++ {
			if p.Cells[y][x] {
				count++
			}
		}
	}
	return count
}
