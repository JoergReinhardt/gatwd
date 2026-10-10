package visualizer

import (
	"fmt"
	"reflect"
	"strings"
)

// Graphable allows structs with unexported fields to expose a generic copy.
type Graphable interface {
	Graph() any
}

// Visualizer allows custom types to define their own graph traversal.
type Visualizer interface {
	Visualize(opts Options, prefix string, isLast bool, currentDepth int) string
}

func applyColor(text, fg, bg string) string {
	fgCodes := map[string]string{"red": "31", "green": "32", "yellow": "33", "blue": "34", "magenta": "35", "cyan": "36", "white": "37"}
	bgCodes := map[string]string{"red": "41", "green": "42", "yellow": "43", "blue": "44", "magenta": "45", "cyan": "46", "white": "47"}
	
	code := ""
	if c, ok := fgCodes[fg]; ok { code += "\033[" + c + "m" }
	if c, ok := bgCodes[bg]; ok { code += "\033[" + c + "m" }
	
	if code != "" { return code + text + "\033[0m" }
	return text
}

func Format(val any, opts Options) string {
	if opts.FormatDOT {
		dot := "digraph G {\n  node [shape=box, style=filled, color=black, fillcolor=white];\n"
		_, dotLines := formatDotNode(val, "root", "root", opts, 0, &[]string{})
		dot += strings.Join(*dotLines, "\n") + "\n}"
		return dot
	}

	res := formatNode(val, "", "", true, opts, 0)
	
	if opts.Border {
		lines := strings.Split(strings.TrimSpace(res), "\n")
		maxLen := 0
		for _, l := range lines {
			// naive length without color stripping
			if len(l) > maxLen { maxLen = len(l) }
		}
		borderLine := "┌" + strings.Repeat("─", maxLen+2) + "┐\n"
		bottomLine := "└" + strings.Repeat("─", maxLen+2) + "┘"
		framed := borderLine
		for _, l := range lines {
			framed += fmt.Sprintf("│ %-*s │\n", maxLen, l)
		}
		res = framed + bottomLine
	}
	
	return applyColor(res, opts.Color, opts.BgColor)
}

func formatDotNode(val any, name string, id string, opts Options, depth int, lines *[]string) (string, *[]string) {
	if g, ok := val.(Graphable); ok { val = g.Graph() }
	
	nodeLabel := name
	if nodeLabel == "" { nodeLabel = "Node" }
	
	v := reflect.ValueOf(val)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			*lines = append(*lines, fmt.Sprintf("  \"%s\" [label=\"<nil>\"];", id))
			return id, lines
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Map:
		*lines = append(*lines, fmt.Sprintf("  \"%s\" [label=\"%s (Record)\"];", id, nodeLabel))
		for i, k := range v.MapKeys() {
			childId := fmt.Sprintf("%s_%d", id, i)
			formatDotNode(v.MapIndex(k).Interface(), fmt.Sprintf("%v", k.Interface()), childId, opts, depth+1, lines)
			*lines = append(*lines, fmt.Sprintf("  \"%s\" -> \"%s\";", id, childId))
		}
	case reflect.Slice, reflect.Array:
		*lines = append(*lines, fmt.Sprintf("  \"%s\" [label=\"%s (List)\"];", id, nodeLabel))
		for i := 0; i < v.Len(); i++ {
			childId := fmt.Sprintf("%s_%d", id, i)
			formatDotNode(v.Index(i).Interface(), fmt.Sprintf("[%d]", i), childId, opts, depth+1, lines)
			*lines = append(*lines, fmt.Sprintf("  \"%s\" -> \"%s\";", id, childId))
		}
	case reflect.Struct:
		*lines = append(*lines, fmt.Sprintf("  \"%s\" [label=\"%s (%s)\"];", id, nodeLabel, v.Type().Name()))
		for i := 0; i < v.NumField(); i++ {
			if v.Field(i).CanInterface() {
				childId := fmt.Sprintf("%s_%d", id, i)
				formatDotNode(v.Field(i).Interface(), v.Type().Field(i).Name, childId, opts, depth+1, lines)
				*lines = append(*lines, fmt.Sprintf("  \"%s\" -> \"%s\";", id, childId))
			}
		}
	default:
		valStr := strings.ReplaceAll(fmt.Sprintf("%v", val), "\"", "\\\"")
		*lines = append(*lines, fmt.Sprintf("  \"%s\" [label=\"%s: %s\"];", id, nodeLabel, valStr))
	}
	return id, lines
}

func formatNode(val any, name string, prefix string, isLast bool, opts Options, depth int) string {
	if opts.MaxDepth > 0 && depth > opts.MaxDepth {
		return prefix + branch(isLast, opts.Style) + name + " [...]\n"
	}
	if g, ok := val.(Graphable); ok { val = g.Graph() }
	if v, ok := val.(Visualizer); ok { return v.Visualize(opts, prefix, isLast, depth) }

	v := reflect.ValueOf(val)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() { return prefix + branch(isLast, opts.Style) + name + ": <nil>\n" }
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Map:
		res := prefix + branch(isLast, opts.Style) + formatName(name, "Record") + "\n"
		newPrefix := prefix + extend(isLast, opts.Style)
		keys := v.MapKeys()
		for i, k := range keys {
			res += formatNode(v.MapIndex(k).Interface(), fmt.Sprintf("%v", k.Interface()), newPrefix, i == len(keys)-1, opts, depth+1)
		}
		if len(keys) == 0 { res += newPrefix + opts.Style.Corner + "(empty)\n" }
		return res
	case reflect.Slice, reflect.Array:
		res := prefix + branch(isLast, opts.Style) + formatName(name, "List") + "\n"
		newPrefix := prefix + extend(isLast, opts.Style)
		for i := 0; i < v.Len(); i++ {
			res += formatNode(v.Index(i).Interface(), fmt.Sprintf("[%d]", i), newPrefix, i == v.Len()-1, opts, depth+1)
		}
		if v.Len() == 0 { res += newPrefix + opts.Style.Corner + "(empty)\n" }
		return res
	case reflect.Struct:
		res := prefix + branch(isLast, opts.Style) + formatName(name, v.Type().Name()) + "\n"
		newPrefix := prefix + extend(isLast, opts.Style)
		for i := 0; i < v.NumField(); i++ {
			if v.Field(i).CanInterface() {
				res += formatNode(v.Field(i).Interface(), v.Type().Field(i).Name, newPrefix, i == v.NumField()-1, opts, depth+1)
			} else {
				res += newPrefix + branch(i == v.NumField()-1, opts.Style) + formatName(v.Type().Field(i).Name, "<unexported>") + "\n"
			}
		}
		if v.NumField() == 0 { res += newPrefix + opts.Style.Corner + "(empty)\n" }
		return res
	default:
		return prefix + branch(isLast, opts.Style) + formatName(name, fmt.Sprintf("%v", val)) + "\n"
	}
}

func branch(isLast bool, style Style) string {
	if isLast { return style.Corner }
	return style.Cross
}
func extend(isLast bool, style Style) string {
	if isLast { return style.Empty }
	return style.Vertical
}
func formatName(name, val string) string {
	if name == "" { return val }
	return name + ": " + val
}
