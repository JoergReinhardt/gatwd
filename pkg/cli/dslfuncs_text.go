package cli

import (
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/functions"
)

// registerCharFuncs adds the Char function forms (hardening W2, D1). All are data-last
// and curried; each argument that must be a Char also accepts a one-character String
// (the conversion table in valueops.go).
//
//	char_code c            Char → Int (the Unicode code point)
//	char_from_code n       Int → Char (an error for an invalid code point)
//	char_to_str c          Char → String
//	char_of_str s          String → Maybe Char (Nothing unless s is exactly one character)
//	char_is_digit/alpha/space/upper/lower c   Char → Bool
//	char_upper/lower c     Char → Char
func registerCharFuncs(m map[string]any) {
	defn(m, "char_code", "c", func(a []any) (any, error) {
		c, err := asChar(a[0])
		if err != nil {
			return nil, err
		}
		return int(c), nil
	})
	defn(m, "char_from_code", "n", func(a []any) (any, error) {
		n, err := asInt(a[0])
		if err != nil {
			return nil, err
		}
		c, ok := functions.CharFromCode(int64(n))
		if !ok {
			return nil, fmt.Errorf("%d is not a valid Unicode code point", n)
		}
		return c, nil
	})
	defn(m, "char_to_str", "c", func(a []any) (any, error) {
		c, err := asChar(a[0])
		if err != nil {
			return nil, err
		}
		return c.String(), nil
	})
	defn(m, "char_of_str", "s", func(a []any) (any, error) {
		s, ok := a[0].(string)
		if !ok {
			return nil, fmt.Errorf("want a String, got %T", a[0])
		}
		if c, ok := functions.CharOfString(s); ok {
			return functions.Just[any](c), nil
		}
		return functions.Nothing[any](), nil
	})

	predicates := map[string]func(rune) bool{
		"char_is_digit": unicode.IsDigit,
		"char_is_alpha": unicode.IsLetter,
		"char_is_space": unicode.IsSpace,
		"char_is_upper": unicode.IsUpper,
		"char_is_lower": unicode.IsLower,
	}
	for name, pred := range predicates {
		defn(m, name, "c", func(a []any) (any, error) {
			c, err := asChar(a[0])
			if err != nil {
				return nil, err
			}
			return pred(rune(c)), nil
		})
	}
	defn(m, "char_upper", "c", func(a []any) (any, error) {
		c, err := asChar(a[0])
		if err != nil {
			return nil, err
		}
		return functions.Char(unicode.ToUpper(rune(c))), nil
	})
	defn(m, "char_lower", "c", func(a []any) (any, error) {
		c, err := asChar(a[0])
		if err != nil {
			return nil, err
		}
		return functions.Char(unicode.ToLower(rune(c))), nil
	})
}

// ---- the str_* library (hardening W2) ---------------------------------------------
//
// Every function is data-last and curried (`s |> str_split ","` is `str_split "," s`), counts
// positions in CHARACTERS (code points), not bytes, and takes a String wherever the
// conversion table (valueops.go) allows one: a collection of Chars is accepted as a String.
// Malformed text is a Maybe (`str_to_int`, `str_index_of`), an impossible request an error.
//
//	str_split sep s · str_join sep xs · str_lines s · str_words s · str_chars s · str_from_chars cs
//	str_trim s · str_upper s · str_lower s · str_repeat n s
//	str_contains sub s · str_starts_with p s · str_ends_with p s · str_index_of sub s (Maybe Int)
//	str_replace old new s · str_slice from to s · str_pad_left n ch s · str_pad_right n ch s
//	str_to_int s · str_to_float s (Maybe) · str_from x · show x · str_format fmt args

// strictInt reads an Int argument; unlike asInt it does not accept a numeric String.
func strictInt(v any) (int, error) {
	if _, isString := v.(string); isString {
		return 0, fmt.Errorf("want an Int, got the String %q", v)
	}
	return asInt(v)
}

// maxStrLen bounds what str_repeat and the padders will build, so a typo cannot exhaust memory.
const maxStrLen = 1 << 26

var (
	intText   = regexp.MustCompile(`^[+-]?[0-9]+$`)
	floatText = regexp.MustCompile(`^[+-]?([0-9]+\.?[0-9]*|\.[0-9]+)([eE][+-]?[0-9]+)?$`)
)

// displayText is the printed form of a value: what the REPL shows for it.
func displayText(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case []byte:
		return string(x)
	}
	return fmt.Sprintf("%+v", v)
}

// quoteString renders s as a double-quoted literal that lexes back to s: the escapes of
// the String literal (\n \t \r \0 \\ \") and \xHH for any other control character.
func quoteString(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch {
		case r == '"':
			b.WriteString(`\"`)
		case r == '\\':
			b.WriteString(`\\`)
		case r == '\n':
			b.WriteString(`\n`)
		case r == '\t':
			b.WriteString(`\t`)
		case r == '\r':
			b.WriteString(`\r`)
		case r == 0:
			b.WriteString(`\0`)
		case unicode.IsControl(r) && r < 0x100:
			fmt.Fprintf(&b, `\x%02x`, r)
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}

// showText renders v as source-like text: Strings in double quotes, Chars in single quotes,
// sequences of such elements in brackets, all with the literals' escapes so the text lexes
// back to the same value.
func showText(v any) string {
	switch x := v.(type) {
	case string:
		return quoteString(x)
	case functions.Char:
		switch x {
		case '\n':
			return `'\n'`
		case '\t':
			return `'\t'`
		case '\r':
			return `'\r'`
		case 0:
			return `'\0'`
		case '\\':
			return `'\\'`
		case '\'':
			return `'\''`
		}
		return "'" + x.String() + "'"
	case *collections.List[any]:
		items, _ := functions.ToSliceAny(x)
		parts := make([]string, len(items))
		for i, it := range items {
			parts[i] = showText(it)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	}
	return displayText(v)
}

func registerStrFuncs(m map[string]any) {
	str := func(v any) (string, error) { return asString(v) }

	// two-argument String → Bool/other helpers share this shape: (needle, haystack).
	pair := func(a []any) (string, string, error) {
		x, err := str(a[0])
		if err != nil {
			return "", "", fmt.Errorf("first argument: %w", err)
		}
		y, err := str(a[1])
		if err != nil {
			return "", "", fmt.Errorf("last argument: %w", err)
		}
		return x, y, nil
	}

	defn(m, "str_split", "sep s", func(a []any) (any, error) {
		sep, s, err := pair(a)
		if err != nil {
			return nil, err
		}
		if sep == "" {
			return nil, fmt.Errorf("empty separator (use str_chars to split into characters)")
		}
		parts := strings.Split(s, sep)
		out := make([]any, len(parts))
		for i, p := range parts {
			out[i] = p
		}
		return listOf(out), nil
	})
	defn(m, "str_join", "sep xs", func(a []any) (any, error) {
		sep, err := str(a[0])
		if err != nil {
			return nil, fmt.Errorf("separator: %w", err)
		}
		items, ok := elementsOf(a[1])
		if !ok {
			return nil, fmt.Errorf("want a collection of Strings, got %T", a[1])
		}
		parts := make([]string, len(items))
		for i, it := range items {
			switch x := it.(type) {
			case string:
				parts[i] = x
			case functions.Char:
				parts[i] = x.String()
			default:
				return nil, fmt.Errorf("element %d is %v (%T), want a String (use str_from to convert)", i, it, it)
			}
		}
		return strings.Join(parts, sep), nil
	})
	defn(m, "str_lines", "s", func(a []any) (any, error) {
		s, err := str(a[0])
		if err != nil {
			return nil, err
		}
		var out []any
		if s != "" {
			for _, l := range strings.Split(strings.TrimSuffix(s, "\n"), "\n") {
				out = append(out, strings.TrimSuffix(l, "\r"))
			}
		}
		return listOf(out), nil
	})
	defn(m, "str_words", "s", func(a []any) (any, error) {
		s, err := str(a[0])
		if err != nil {
			return nil, err
		}
		fields := strings.Fields(s)
		out := make([]any, len(fields))
		for i, f := range fields {
			out[i] = f
		}
		return listOf(out), nil
	})
	defn(m, "str_chars", "s", func(a []any) (any, error) {
		s, err := str(a[0])
		if err != nil {
			return nil, err
		}
		return charList(s)
	})
	defn(m, "str_from_chars", "cs", func(a []any) (any, error) { return asString(a[0]) })

	for name, f := range map[string]func(string) string{
		"str_trim": strings.TrimSpace, "str_upper": strings.ToUpper, "str_lower": strings.ToLower,
	} {
		defn(m, name, "s", func(a []any) (any, error) {
			s, err := str(a[0])
			if err != nil {
				return nil, err
			}
			return f(s), nil
		})
	}
	defn(m, "str_repeat", "n s", func(a []any) (any, error) {
		n, err := strictInt(a[0])
		if err != nil {
			return nil, err
		}
		s, err := str(a[1])
		if err != nil {
			return nil, err
		}
		if n < 0 {
			return nil, fmt.Errorf("negative count %d", n)
		}
		if len(s)*n > maxStrLen {
			return nil, fmt.Errorf("result longer than %d bytes", maxStrLen)
		}
		return strings.Repeat(s, n), nil
	})

	for name, f := range map[string]func(s, sub string) bool{
		"str_contains": strings.Contains, "str_starts_with": strings.HasPrefix, "str_ends_with": strings.HasSuffix,
	} {
		defn(m, name, "sub s", func(a []any) (any, error) {
			sub, s, err := pair(a)
			if err != nil {
				return nil, err
			}
			return f(s, sub), nil
		})
	}
	defn(m, "str_index_of", "sub s", func(a []any) (any, error) {
		sub, s, err := pair(a)
		if err != nil {
			return nil, err
		}
		byteIdx := strings.Index(s, sub)
		if byteIdx < 0 {
			return functions.Nothing[any](), nil
		}
		return functions.Just[any](utf8.RuneCountInString(s[:byteIdx])), nil
	})
	defn(m, "str_replace", "old new s", func(a []any) (any, error) {
		old, err := str(a[0])
		if err != nil {
			return nil, fmt.Errorf("first argument: %w", err)
		}
		repl, err := str(a[1])
		if err != nil {
			return nil, fmt.Errorf("second argument: %w", err)
		}
		s, err := str(a[2])
		if err != nil {
			return nil, fmt.Errorf("last argument: %w", err)
		}
		if old == "" {
			return nil, fmt.Errorf("empty pattern")
		}
		return strings.ReplaceAll(s, old, repl), nil
	})
	defn(m, "str_slice", "from to s", func(a []any) (any, error) {
		from, err := strictInt(a[0])
		if err != nil {
			return nil, err
		}
		to, err := strictInt(a[1])
		if err != nil {
			return nil, err
		}
		s, err := str(a[2])
		if err != nil {
			return nil, err
		}
		rs := []rune(s)
		if from < 0 || to < from || to > len(rs) {
			return nil, fmt.Errorf("slice [%d,%d) out of range for a String of %d characters", from, to, len(rs))
		}
		return string(rs[from:to]), nil
	})
	for _, left := range []bool{true, false} {
		name := "str_pad_right"
		if left {
			name = "str_pad_left"
		}
		defn(m, name, "n ch s", func(a []any) (any, error) {
			n, err := strictInt(a[0])
			if err != nil {
				return nil, err
			}
			ch, err := asChar(a[1])
			if err != nil {
				return nil, err
			}
			s, err := str(a[2])
			if err != nil {
				return nil, err
			}
			if n > maxStrLen {
				return nil, fmt.Errorf("width %d exceeds %d", n, maxStrLen)
			}
			fill := n - utf8.RuneCountInString(s)
			if fill <= 0 {
				return s, nil
			}
			pad := strings.Repeat(ch.String(), fill)
			if left {
				return pad + s, nil
			}
			return s + pad, nil
		})
	}

	defn(m, "str_to_int", "s", func(a []any) (any, error) {
		s, err := str(a[0])
		if err != nil {
			return nil, err
		}
		if !intText.MatchString(s) {
			return functions.Nothing[any](), nil
		}
		if n, perr := strconv.ParseInt(s, 10, 64); perr == nil {
			return functions.Just[any](functions.TryNarrowInteger(n)), nil
		}
		big, _ := new(big.Int).SetString(strings.TrimPrefix(s, "+"), 10)
		return functions.Just[any](big), nil
	})
	defn(m, "str_to_float", "s", func(a []any) (any, error) {
		s, err := str(a[0])
		if err != nil {
			return nil, err
		}
		if !floatText.MatchString(s) {
			return functions.Nothing[any](), nil
		}
		f, perr := strconv.ParseFloat(s, 64)
		if perr != nil {
			return functions.Nothing[any](), nil
		}
		return functions.Just[any](f), nil
	})

	defn(m, "str_from", "x", func(a []any) (any, error) { return displayText(a[0]), nil })
	defn(m, "show", "x", func(a []any) (any, error) { return showText(a[0]), nil })
	defn(m, "str_format", "fmt args", func(a []any) (any, error) {
		f, err := str(a[0])
		if err != nil {
			return nil, fmt.Errorf("format: %w", err)
		}
		args, ok := elementsOf(a[1])
		if !ok {
			return nil, fmt.Errorf("arguments must be a list, got %T", a[1])
		}
		return formatText(f, args)
	})
}

// formatText is a checked printf subset: %s %v %d %f %g %e %x %q %c %% with the usual flags,
// width and precision. Each verb names the type it needs, and a mismatch, a missing or a
// surplus argument is an error — printf's `%!d(string=…)` placeholders never appear.
func formatText(f string, args []any) (any, error) {
	var out strings.Builder
	next := 0
	for i := 0; i < len(f); {
		c := f[i]
		if c != '%' {
			out.WriteByte(c)
			i++
			continue
		}
		j := i + 1
		for j < len(f) && strings.IndexByte("+-# 0123456789.", f[j]) >= 0 {
			j++
		}
		if j >= len(f) {
			return nil, fmt.Errorf("format ends inside a %%-verb")
		}
		verb := f[j]
		spec := f[i : j+1]
		i = j + 1
		if verb == '%' {
			out.WriteByte('%')
			continue
		}
		if next >= len(args) {
			return nil, fmt.Errorf("%s needs an argument but only %d were given", spec, len(args))
		}
		arg := args[next]
		next++
		var goArg any
		switch verb {
		case 's':
			switch x := arg.(type) {
			case string:
				goArg = x
			case functions.Char:
				goArg = x.String()
			default:
				return nil, fmt.Errorf("%s wants a String, got %v (%T); use %%v or str_from", spec, arg, arg)
			}
		case 'v':
			goArg = displayText(arg)
			spec = spec[:len(spec)-1] + "s"
		case 'q':
			switch x := arg.(type) {
			case string:
				goArg = x
			case functions.Char:
				goArg = x.String()
			default:
				return nil, fmt.Errorf("%s wants a String or Char, got %v (%T)", spec, arg, arg)
			}
		case 'c':
			ch, err := asChar(arg)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", spec, err)
			}
			goArg = rune(ch)
		case 'd', 'x':
			switch x := arg.(type) {
			case *big.Int:
				goArg = x
			case string, functions.Char:
				if verb == 'x' {
					if sx, ok := x.(string); ok {
						goArg = sx
						break
					}
				}
				return nil, fmt.Errorf("%s wants an Int, got %v (%T)", spec, arg, arg)
			default:
				n, err := strictInt(arg)
				if err != nil {
					return nil, fmt.Errorf("%s wants an Int, got %v (%T)", spec, arg, arg)
				}
				goArg = n
			}
		case 'f', 'g', 'e':
			if _, isString := arg.(string); isString || functions.IsChar(arg) {
				return nil, fmt.Errorf("%s wants a number, got %v (%T)", spec, arg, arg)
			}
			fl, ok := functions.ToFloat64(arg)
			if !ok {
				return nil, fmt.Errorf("%s wants a number, got %v (%T)", spec, arg, arg)
			}
			goArg = fl
		default:
			return nil, fmt.Errorf("unsupported verb %q in %s", verb, spec)
		}
		out.WriteString(fmt.Sprintf(spec, goArg))
	}
	if next < len(args) {
		return nil, fmt.Errorf("%d argument(s) given but the format uses %d", len(args), next)
	}
	return out.String(), nil
}
