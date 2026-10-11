package visualizer

import (
	"fmt"
	"regexp"
	"strings"
)

var ansiRegex = regexp.MustCompile("\033\\[[0-9;]*m")

func stripANSI(str string) string {
	return ansiRegex.ReplaceAllString(str, "")
}

func addBorder(res string) string {
	lines := strings.Split(strings.TrimSpace(res), "\n")
	maxLen := 0
	for _, l := range lines {
		visibleLen := len([]rune(stripANSI(l)))
		if visibleLen > maxLen {
			maxLen = visibleLen
		}
	}

	borderLine := "╭" + strings.Repeat("─", maxLen+2) + "╮\n"
	bottomLine := "╰" + strings.Repeat("─", maxLen+2) + "╯"

	framed := borderLine
	for _, l := range lines {
		visibleLen := len([]rune(stripANSI(l)))
		padding := maxLen - visibleLen
		framed += fmt.Sprintf("│ %s%s │\n", l, strings.Repeat(" ", padding))
	}
	return framed + bottomLine
}
