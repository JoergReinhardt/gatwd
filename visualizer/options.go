package visualizer

import (
	"strconv"
	"strings"

	"github.com/joergreinhardt/gatwd/dsltypes"
)

type Style struct {
	Vertical string
	Cross    string
	Corner   string
	Empty    string
	BoxTop   string
	BoxBot   string
	BoxSide  string
}

var UTF8Style = Style{Vertical: "│   ", Cross: "├── ", Corner: "└── ", Empty: "    ", BoxTop: "╭───", BoxBot: "╰───", BoxSide: "│"}
var CurvedStyle = Style{Vertical: "│   ", Cross: "├─ ", Corner: "╰─ ", Empty: "    ", BoxTop: "╭───", BoxBot: "╰───", BoxSide: "│"}
var ASCIIStyle = Style{Vertical: "|   ", Cross: "+-- ", Corner: "`-- ", Empty: "    ", BoxTop: "+---", BoxBot: "+---", BoxSide: "|"}
var HeavyStyle = Style{Vertical: "┃   ", Cross: "┣━━ ", Corner: "┗━━ ", Empty: "    ", BoxTop: "┏━━━", BoxBot: "┗━━━", BoxSide: "┃"}

type Options struct {
	Style           Style
	CollapseScalars bool
	MaxDepth        int
	FormatDOT       bool
	Color           string
	BgColor         string
	Border          bool
	AutoBorder      int
	Head            int
	Tail            int
	TypeColors      map[string]string
	FlagColors      map[dsltypes.BitFlag]string
	ValueColors     map[string]string
	BoxTypes        map[string]bool
	BoxFlags        map[dsltypes.BitFlag]bool
}

func DefaultOptions() Options {
	return Options{
		Style:       UTF8Style,
		TypeColors:  make(map[string]string),
		FlagColors:  make(map[dsltypes.BitFlag]string),
		ValueColors: make(map[string]string),
		BoxTypes:    make(map[string]bool),
		BoxFlags:    make(map[dsltypes.BitFlag]bool),
	}
}

func ParseOptions(args string) Options {
	opts := DefaultOptions()
	parts := strings.Fields(args)

	for _, p := range parts {
		if p == "--ascii" {
			opts.Style = ASCIIStyle
		}
		if p == "--curved" {
			opts.Style = CurvedStyle
		}
		if p == "--heavy" {
			opts.Style = HeavyStyle
		}
		if p == "--collapse" {
			opts.CollapseScalars = true
		}
		if p == "--dot" {
			opts.FormatDOT = true
		}
		if p == "--border" {
			opts.Border = true
		}
		if p == "--box-all" {
			opts.BoxTypes["all"] = true
		}

		if strings.HasPrefix(p, "--depth=") {
			if d, err := strconv.Atoi(strings.TrimPrefix(p, "--depth=")); err == nil {
				opts.MaxDepth = d
			}
		}
		if strings.HasPrefix(p, "--auto-border=") {
			if d, err := strconv.Atoi(strings.TrimPrefix(p, "--auto-border=")); err == nil {
				opts.AutoBorder = d
			}
		}
		if strings.HasPrefix(p, "--head=") {
			if d, err := strconv.Atoi(strings.TrimPrefix(p, "--head=")); err == nil {
				opts.Head = d
			}
		}
		if strings.HasPrefix(p, "--tail=") {
			if d, err := strconv.Atoi(strings.TrimPrefix(p, "--tail=")); err == nil {
				opts.Tail = d
			}
		}
		if strings.HasPrefix(p, "--color=") {
			opts.Color = strings.TrimPrefix(p, "--color=")
		}
		if strings.HasPrefix(p, "--bg=") {
			opts.BgColor = strings.TrimPrefix(p, "--bg=")
		}

		if strings.HasPrefix(p, "--color-type=") {
			kv := strings.SplitN(strings.TrimPrefix(p, "--color-type="), ":", 2)
			if len(kv) == 2 {
				opts.TypeColors[kv[0]] = kv[1]
			}
		}
		if strings.HasPrefix(p, "--color-val=") {
			kv := strings.SplitN(strings.TrimPrefix(p, "--color-val="), ":", 2)
			if len(kv) == 2 {
				opts.ValueColors[kv[0]] = kv[1]
			}
		}
		if strings.HasPrefix(p, "--box=") {
			types := strings.Split(strings.TrimPrefix(p, "--box="), ",")
			for _, t := range types {
				opts.BoxTypes[t] = true
			}
		}
		if strings.HasPrefix(p, "--color-flag=") {
			kv := strings.SplitN(strings.TrimPrefix(p, "--color-flag="), ":", 2)
			if len(kv) == 2 {
				if f, err := strconv.ParseUint(kv[0], 10, 64); err == nil {
					opts.FlagColors[dsltypes.BitFlag(f)] = kv[1]
				}
			}
		}
		if strings.HasPrefix(p, "--box-flag=") {
			flags := strings.Split(strings.TrimPrefix(p, "--box-flag="), ",")
			for _, fStr := range flags {
				if f, err := strconv.ParseUint(fStr, 10, 64); err == nil {
					opts.BoxFlags[dsltypes.BitFlag(f)] = true
				}
			}
		}
	}
	return opts
}
