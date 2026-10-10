package visualizer

import (
	"strconv"
	"strings"
)

type Style struct {
	Vertical   string
	Cross      string
	Corner     string
	Empty      string
}

var UTF8Style = Style{ Vertical: "│   ", Cross: "├── ", Corner: "└── ", Empty: "    " }
var ASCIIStyle = Style{ Vertical: "|   ", Cross: "+-- ", Corner: "`-- ", Empty: "    " }

type Options struct {
	Style           Style
	CollapseScalars bool
	MaxDepth        int
	FormatDOT       bool
	Color           string
	BgColor         string
	Border          bool
}

func DefaultOptions() Options {
	return Options{ Style: UTF8Style }
}

func ParseOptions(args string) Options {
	opts := DefaultOptions()
	parts := strings.Fields(args)
	
	for _, p := range parts {
		if p == "--ascii" { opts.Style = ASCIIStyle }
		if p == "--collapse" { opts.CollapseScalars = true }
		if p == "--dot" { opts.FormatDOT = true }
		if p == "--border" { opts.Border = true }
		if strings.HasPrefix(p, "--depth=") {
			if d, err := strconv.Atoi(strings.TrimPrefix(p, "--depth=")); err == nil { opts.MaxDepth = d }
		}
		if strings.HasPrefix(p, "--color=") { opts.Color = strings.TrimPrefix(p, "--color=") }
		if strings.HasPrefix(p, "--bg=") { opts.BgColor = strings.TrimPrefix(p, "--bg=") }
	}
	return opts
}
