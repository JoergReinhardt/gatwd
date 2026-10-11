package visualizer

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/joergreinhardt/gatwd/dsltypes"
)

type Graphable interface {
	Graph() any
}

type Visualizer interface {
	Visualize(opts Options, prefix string, isLast bool, currentDepth int) string
}

func applyColor(text, fg, bg string) string {
	fgCodes := map[string]string{"red": "31", "green": "32", "yellow": "33", "blue": "34", "magenta": "35", "cyan": "36", "white": "37"}
	bgCodes := map[string]string{"red": "41", "green": "42", "yellow": "43", "blue": "44", "magenta": "45", "cyan": "46", "white": "47"}

	code := ""
	if c, ok := fgCodes[fg]; ok {
		code += "\033[" + c + "m"
	}
	if c, ok := bgCodes[bg]; ok {
		code += "\033[" + c + "m"
	}

	if code != "" {
		return code + text + "\033[0m"
	}
	return text
}

// FormatAST formats an AST node value using its BitFlag DSL type rather than Go reflection.
func FormatAST(flag dsltypes.BitFlag, val any, opts Options) string {
	if opts.FormatDOT {
		dot := "digraph G {\n  node [shape=box, style=filled, color=black, fillcolor=white];\n"
		_, dotLines := formatDotNode(val, "root", "root", opts, 0, &[]string{}, flag, true)
		dot += strings.Join(*dotLines, "\n") + "\n}"
		return dot
	}

	res := formatNode(val, "", "", true, opts, 0, flag, true)

	lineCount := len(strings.Split(strings.TrimSpace(res), "\n"))
	if opts.Border || (opts.AutoBorder > 0 && lineCount > opts.AutoBorder) {
		res = addBorder(res)
	}
	return applyColor(res, opts.Color, opts.BgColor)
}

// Format renders any value into a pretty-printed UTF-8 string graph.
// If val is a dsltypes.TypedNode (such as dsltypes.ASTNode), its type is determined
// strictly from the node's BitFlag type tag rather than Go reflection.
func Format(val any, opts Options) string {
	if tn, ok := val.(dsltypes.TypedNode); ok {
		flag := tn.TypeFlag()
		inner := tn.NodeValue()
		if either, ok := inner.(interface {
			IsLeft() bool
			UnwrapLeft() (error, bool)
			UnwrapRight() (any, bool)
		}); ok {
			if either.IsLeft() {
				err, _ := either.UnwrapLeft()
				return FormatAST(flag, err, opts)
			}
			rightVal, _ := either.UnwrapRight()
			return FormatAST(flag, rightVal, opts)
		}
		return FormatAST(flag, inner, opts)
	}

	if opts.FormatDOT {
		dot := "digraph G {\n  node [shape=box, style=filled, color=black, fillcolor=white];\n"
		_, dotLines := formatDotNode(val, "root", "root", opts, 0, &[]string{}, 0, false)
		dot += strings.Join(*dotLines, "\n") + "\n}"
		return dot
	}

	res := formatNode(val, "", "", true, opts, 0, 0, false)

	lineCount := len(strings.Split(strings.TrimSpace(res), "\n"))
	if opts.Border || (opts.AutoBorder > 0 && lineCount > opts.AutoBorder) {
		res = addBorder(res)
	}
	return applyColor(res, opts.Color, opts.BgColor)
}

func extractNodeColor(v reflect.Value) string {
	if v.Kind() == reflect.Struct {
		f := v.FieldByName("NodeColor")
		if f.IsValid() && f.Kind() == reflect.String {
			return f.String()
		}
	}
	return ""
}

func formatDotNode(val any, name string, id string, opts Options, depth int, lines *[]string, flag dsltypes.BitFlag, hasFlag bool) (string, *[]string) {
	if tn, ok := val.(dsltypes.TypedNode); ok {
		flag = tn.TypeFlag()
		hasFlag = true
		val = tn.NodeValue()
		if either, ok := val.(interface {
			IsLeft() bool
			UnwrapLeft() (error, bool)
			UnwrapRight() (any, bool)
		}); ok {
			if either.IsLeft() {
				err, _ := either.UnwrapLeft()
				val = err
			} else {
				rightVal, _ := either.UnwrapRight()
				val = rightVal
			}
		}
	}

	if g, ok := val.(Graphable); ok {
		val = g.Graph()
	}
	nodeLabel := name
	if nodeLabel == "" {
		nodeLabel = "Node"
	}
	v := reflect.ValueOf(val)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			*lines = append(*lines, fmt.Sprintf("  \"%s\" [label=\"<nil>\"];", id))
			return id, lines
		}
		v = v.Elem()
	}
	nodeColor := extractNodeColor(v)

	var typeName string
	if hasFlag {
		typeName = flag.TypeName()
		if c, ok := opts.FlagColors[flag]; ok {
			nodeColor = c
		}
	} else {
		typeName = v.Type().Name()
	}

	if nodeColor == "" && typeName != "" {
		if c, ok := opts.TypeColors[typeName]; ok {
			nodeColor = c
		}
	}

	if err, ok := val.(error); ok {
		val = err.Error()
		v = reflect.ValueOf(val)
		if nodeColor == "" {
			nodeColor = "red"
		}
	} else if str, ok := val.(fmt.Stringer); ok {
		val = str.String()
		v = reflect.ValueOf(val)
	}
	colorAttr := ""
	if nodeColor != "" {
		colorAttr = fmt.Sprintf(", fillcolor=\"%s\"", nodeColor)
	}

	switch v.Kind() {
	case reflect.Map:
		label := typeName
		if label == "" {
			label = "Record"
		}
		*lines = append(*lines, fmt.Sprintf("  \"%s\" [label=\"%s (%s)\"%s];", id, nodeLabel, label, colorAttr))
		for i, k := range v.MapKeys() {
			childId := fmt.Sprintf("%s_%d", id, i)
			formatDotNode(v.MapIndex(k).Interface(), fmt.Sprintf("%v", k.Interface()), childId, opts, depth+1, lines, 0, false)
			*lines = append(*lines, fmt.Sprintf("  \"%s\" -> \"%s\";", id, childId))
		}
	case reflect.Slice, reflect.Array:
		label := typeName
		if label == "" {
			label = "List"
		}
		*lines = append(*lines, fmt.Sprintf("  \"%s\" [label=\"%s (%s)\"%s];", id, nodeLabel, label, colorAttr))
		for i := 0; i < v.Len(); i++ {
			childId := fmt.Sprintf("%s_%d", id, i)
			formatDotNode(v.Index(i).Interface(), fmt.Sprintf("[%d]", i), childId, opts, depth+1, lines, 0, false)
			*lines = append(*lines, fmt.Sprintf("  \"%s\" -> \"%s\";", id, childId))
		}
	case reflect.Struct:
		label := typeName
		if label == "" {
			label = "Struct"
		}
		*lines = append(*lines, fmt.Sprintf("  \"%s\" [label=\"%s (%s)\"%s];", id, nodeLabel, label, colorAttr))
		for i := 0; i < v.NumField(); i++ {
			if v.Type().Field(i).Name == "NodeColor" {
				continue
			}
			if v.Field(i).CanInterface() {
				childId := fmt.Sprintf("%s_%d", id, i)
				formatDotNode(v.Field(i).Interface(), v.Type().Field(i).Name, childId, opts, depth+1, lines, 0, false)
				*lines = append(*lines, fmt.Sprintf("  \"%s\" -> \"%s\";", id, childId))
			}
		}
	default:
		valStr := strings.ReplaceAll(fmt.Sprintf("%v", val), "\"", "\\\"")
		if c, ok := opts.ValueColors[valStr]; ok {
			colorAttr = fmt.Sprintf(", fillcolor=\"%s\"", c)
		}
		*lines = append(*lines, fmt.Sprintf("  \"%s\" [label=\"%s: %s\"%s];", id, nodeLabel, valStr, colorAttr))
	}
	return id, lines
}

func formatNode(val any, name string, prefix string, isLast bool, opts Options, depth int, flag dsltypes.BitFlag, hasFlag bool) string {
	if opts.MaxDepth > 0 && depth > opts.MaxDepth {
		return prefix + branch(isLast, opts.Style) + name + " [...]\n"
	}

	// Unwrap AST node if val is a TypedNode
	if tn, ok := val.(dsltypes.TypedNode); ok {
		flag = tn.TypeFlag()
		hasFlag = true
		val = tn.NodeValue()
		if either, ok := val.(interface {
			IsLeft() bool
			UnwrapLeft() (error, bool)
			UnwrapRight() (any, bool)
		}); ok {
			if either.IsLeft() {
				err, _ := either.UnwrapLeft()
				val = err
			} else {
				rightVal, _ := either.UnwrapRight()
				val = rightVal
			}
		}
	}

	if g, ok := val.(Graphable); ok {
		val = g.Graph()
	}
	if v, ok := val.(Visualizer); ok {
		return v.Visualize(opts, prefix, isLast, depth)
	}

	v := reflect.ValueOf(val)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			nilColor := opts.ValueColors["<nil>"]
			out := name + ": <nil>"
			if nilColor != "" {
				out = applyColor(out, nilColor, "")
			}
			return prefix + branch(isLast, opts.Style) + out + "\n"
		}
		v = v.Elem()
	}

	nodeColor := extractNodeColor(v)

	// Determine type name: Prioritize the AST node's BitFlag type over Go reflection
	var typeName string
	if hasFlag {
		typeName = flag.TypeName()
	} else {
		typeName = v.Type().Name()
		if typeName == "" {
			if v.Kind() == reflect.Struct {
				typeName = "Struct"
			}
			if v.Kind() == reflect.Map {
				typeName = "Record"
			}
			if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
				typeName = "List"
			}
		}
	}

	if nodeColor == "" && hasFlag {
		if c, ok := opts.FlagColors[flag]; ok {
			nodeColor = c
		}
	}
	if nodeColor == "" && typeName != "" {
		if c, ok := opts.TypeColors[typeName]; ok {
			nodeColor = c
		}
	}

	// If it implements error, format it with error styling
	if err, ok := val.(error); ok {
		val = err.Error()
		v = reflect.ValueOf(val)
		if nodeColor == "" {
			nodeColor = "red"
		}
	} else if str, ok := val.(fmt.Stringer); ok {
		val = str.String()
		v = reflect.ValueOf(val)
	}

	valStr := fmt.Sprintf("%v", val)
	if nodeColor == "" {
		if c, ok := opts.ValueColors[valStr]; ok {
			nodeColor = c
		}
	}

	formatColoredName := func(name, val string) string {
		out := name
		if out != "" && val != "" {
			out += ": " + val
		} else if out == "" {
			out = val
		}
		if nodeColor != "" {
			return applyColor(out, nodeColor, "")
		}
		return out
	}

	shouldBox := opts.BoxTypes["all"] || (hasFlag && opts.BoxFlags[flag]) || (typeName != "" && opts.BoxTypes[typeName])
	var boxPrefix, boxTop, boxBot string
	if shouldBox {
		boxPrefix = opts.Style.BoxSide + "   "
		boxTop = prefix + extend(isLast, opts.Style) + opts.Style.BoxTop + "\n"
		boxBot = prefix + extend(isLast, opts.Style) + opts.Style.BoxBot + "\n"
	}

	switch v.Kind() {
	case reflect.Map:
		res := prefix + branch(isLast, opts.Style) + formatColoredName(name, typeName) + "\n"
		newPrefix := prefix + extend(isLast, opts.Style)
		if shouldBox {
			res += boxTop
			newPrefix = prefix + extend(isLast, opts.Style) + boxPrefix
		}

		keys := v.MapKeys()
		for i, k := range keys {
			res += formatNode(v.MapIndex(k).Interface(), fmt.Sprintf("%v", k.Interface()), newPrefix, i == len(keys)-1, opts, depth+1, 0, false)
		}
		if len(keys) == 0 {
			res += newPrefix + opts.Style.Corner + "(empty)\n"
		}
		if shouldBox {
			res += boxBot
		}
		return res

	case reflect.Slice, reflect.Array:
		res := prefix + branch(isLast, opts.Style) + formatColoredName(name, typeName) + "\n"
		newPrefix := prefix + extend(isLast, opts.Style)
		if shouldBox {
			res += boxTop
			newPrefix = prefix + extend(isLast, opts.Style) + boxPrefix
		}

		length := v.Len()
		var iterList []int
		if opts.Head > 0 && opts.Tail > 0 && length > (opts.Head+opts.Tail) {
			for i := 0; i < opts.Head; i++ {
				iterList = append(iterList, i)
			}
			iterList = append(iterList, -1)
			for i := length - opts.Tail; i < length; i++ {
				iterList = append(iterList, i)
			}
		} else {
			for i := 0; i < length; i++ {
				iterList = append(iterList, i)
			}
		}

		for idx, i := range iterList {
			if i == -1 {
				omitted := length - (opts.Head + opts.Tail)
				msg := fmt.Sprintf("... (%d items omitted) ...", omitted)
				res += newPrefix + branch(idx == len(iterList)-1, opts.Style) + applyColor(msg, "yellow", "") + "\n"
			} else {
				res += formatNode(v.Index(i).Interface(), fmt.Sprintf("[%d]", i), newPrefix, idx == len(iterList)-1, opts, depth+1, 0, false)
			}
		}
		if v.Len() == 0 {
			res += newPrefix + opts.Style.Corner + "(empty)\n"
		}
		if shouldBox {
			res += boxBot
		}
		return res

	case reflect.Struct:
		res := prefix + branch(isLast, opts.Style) + formatColoredName(name, typeName) + "\n"
		newPrefix := prefix + extend(isLast, opts.Style)
		if shouldBox {
			res += boxTop
			newPrefix = prefix + extend(isLast, opts.Style) + boxPrefix
		}

		fieldCount := v.NumField()
		for i := 0; i < fieldCount; i++ {
			if v.Type().Field(i).Name == "NodeColor" {
				continue
			}
			if v.Field(i).CanInterface() {
				res += formatNode(v.Field(i).Interface(), v.Type().Field(i).Name, newPrefix, i == fieldCount-1, opts, depth+1, 0, false)
			} else {
				res += newPrefix + branch(i == fieldCount-1, opts.Style) + formatColoredName(v.Type().Field(i).Name, "<unexported>") + "\n"
			}
		}
		if fieldCount == 0 {
			res += newPrefix + opts.Style.Corner + "(empty)\n"
		}
		if shouldBox {
			res += boxBot
		}
		return res

	default:
		displayName := name
		if displayName == "" && hasFlag {
			displayName = typeName
		}
		return prefix + branch(isLast, opts.Style) + formatColoredName(displayName, valStr) + "\n"
	}
}

func branch(isLast bool, style Style) string {
	if isLast {
		return style.Corner
	}
	return style.Cross
}

func extend(isLast bool, style Style) string {
	if isLast {
		return style.Empty
	}
	return style.Vertical
}
