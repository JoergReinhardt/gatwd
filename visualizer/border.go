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

func addBorder(res string, borderColor string) string {
	lines := strings.Split(strings.TrimSpace(res), "\n")
	maxLen := 0
	for _, l := range lines {
		visibleLen := len([]rune(stripANSI(l)))
		if visibleLen > maxLen {
			maxLen = visibleLen
		}
	}

	topBorder := "╭" + strings.Repeat("─", maxLen+2) + "╮"
	botBorder := "╰" + strings.Repeat("─", maxLen+2) + "╯"

	if borderColor != "" {
		topBorder = applyColor(topBorder, borderColor, "")
		botBorder = applyColor(botBorder, borderColor, "")
	}

	borderLine := topBorder + "\n"
	bottomLine := botBorder

	framed := borderLine
	for _, l := range lines {
		visibleLen := len([]rune(stripANSI(l)))
		padding := maxLen - visibleLen
		leftSide := "│"
		rightSide := "│"
		if borderColor != "" {
			leftSide = applyColor(leftSide, borderColor, "")
			rightSide = applyColor(rightSide, borderColor, "")
		}
		framed += fmt.Sprintf("%s %s%s %s\n", leftSide, l, strings.Repeat(" ", padding), rightSide)
	}
	return framed + bottomLine
}
