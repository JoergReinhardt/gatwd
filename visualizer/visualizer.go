package visualizer

import (
	"fmt"
	"reflect"
)

type Style struct {
	Vertical   string // "│   "
	Cross      string // "├── "
	Corner     string // "└── "
	Empty      string // "    "
}

var UTF8Style = Style{
	Vertical: "│   ",
	Cross:    "├── ",
	Corner:   "└── ",
	Empty:    "    ",
}

var ASCIIStyle = Style{
	Vertical: "|   ",
	Cross:    "+-- ",
	Corner:   "`-- ",
	Empty:    "    ",
}

type Options struct {
	Style           Style
	CollapseScalars bool // If true, short scalar lists are printed on one line
	MaxDepth        int  // 0 means infinite
}

func DefaultOptions() Options {
	return Options{
		Style:           UTF8Style,
		CollapseScalars: false,
		MaxDepth:        0,
	}
}

// Visualizer allows custom types to define their own graph traversal for printing
type Visualizer interface {
	Visualize(opts Options, prefix string, isLast bool, currentDepth int) string
}

// Format takes any arbitrary data structure and returns a UTF-8 graph string.
func Format(val any, opts Options) string {
	return formatNode(val, "", "", true, opts, 0)
}

func formatNode(val any, name string, prefix string, isLast bool, opts Options, depth int) string {
	if opts.MaxDepth > 0 && depth > opts.MaxDepth {
		return prefix + branch(isLast, opts.Style) + name + " [...]\n"
	}

	// 1. Check if it explicitly implements Graphable
	if g, ok := val.(Graphable); ok {
		val = g.Graph()
	}

	// 1. If it implements Visualizer
	if v, ok := val.(Visualizer); ok {
		return v.Visualize(opts, prefix, isLast, depth)
	}

	// 2. Reflect on the value to handle maps, slices, and pointers generically
	v := reflect.ValueOf(val)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return prefix + branch(isLast, opts.Style) + name + ": <nil>\n"
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Map:
		res := prefix + branch(isLast, opts.Style) + formatName(name, "Record") + "\n"
		newPrefix := prefix + extend(isLast, opts.Style)
		keys := v.MapKeys()
		for i, k := range keys {
			childIsLast := (i == len(keys)-1)
			keyName := fmt.Sprintf("%v", k.Interface())
			res += formatNode(v.MapIndex(k).Interface(), keyName, newPrefix, childIsLast, opts, depth+1)
		}
		if len(keys) == 0 {
			res += newPrefix + opts.Style.Corner + "(empty)\n"
		}
		return res

	case reflect.Slice, reflect.Array:
		if opts.CollapseScalars && isScalarSlice(v) {
			res := prefix + branch(isLast, opts.Style) + formatName(name, fmt.Sprintf("%v", v.Interface())) + "\n"
			return res
		}
		
		res := prefix + branch(isLast, opts.Style) + formatName(name, "List") + "\n"
		newPrefix := prefix + extend(isLast, opts.Style)
		for i := 0; i < v.Len(); i++ {
			childIsLast := (i == v.Len()-1)
			res += formatNode(v.Index(i).Interface(), fmt.Sprintf("[%d]", i), newPrefix, childIsLast, opts, depth+1)
		}
		if v.Len() == 0 {
			res += newPrefix + opts.Style.Corner + "(empty)\n"
		}
		return res

	case reflect.Struct:
		// We treat unrecognized structs like Records
		res := prefix + branch(isLast, opts.Style) + formatName(name, v.Type().Name()) + "\n"
		newPrefix := prefix + extend(isLast, opts.Style)
		for i := 0; i < v.NumField(); i++ {
			childIsLast := (i == v.NumField()-1)
			fieldName := v.Type().Field(i).Name
			if v.Field(i).CanInterface() {
				res += formatNode(v.Field(i).Interface(), fieldName, newPrefix, childIsLast, opts, depth+1)
			} else {
				res += newPrefix + branch(childIsLast, opts.Style) + formatName(fieldName, "<unexported>") + "\n"
			}
		}
		if v.NumField() == 0 {
			res += newPrefix + opts.Style.Corner + "(empty)\n"
		}
		return res

	default:
		// Scalar value
		valStr := fmt.Sprintf("%v", val)
		return prefix + branch(isLast, opts.Style) + formatName(name, valStr) + "\n"
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

func isScalarSlice(v reflect.Value) bool {
	if v.Len() > 10 { return false } // Too big to collapse
	for i := 0; i < v.Len(); i++ {
		k := v.Index(i).Kind()
		if k == reflect.Map || k == reflect.Slice || k == reflect.Struct || k == reflect.Ptr {
			return false
		}
	}
	return true
}

// Graphable allows structs with unexported fields (like our Collections) 
// to expose a generic structurally-equivalent copy for graph visualization.
type Graphable interface {
	Graph() any
}
