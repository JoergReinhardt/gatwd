package cli

import (
	"fmt"
	"hash/fnv"
	"strings"
)

// Pattern parameters in lambdas: `\0 y -> y`, `\(a, b) -> a + b`, `\[h, ...t] -> h`.
//
// A lambda whose parameters are all names (or `_`) is the ordinary closure. When a parameter is
// a pattern, the statement front end rewrites the lambda — the same desugaring clause
// definitions use — into an ordinary closure over a hidden `match` def:
//
//	\0 y -> y   ⇒   (\_p1 _p2 -> __lam1a2b3c__ ([_p1, _p2]))      match { [0, y] -> y }
//
// One pattern parameter may be any pattern; several parameters are matched as the tuple of
// the arguments, so each must then be `_`, a name or a literal (the tuple-pattern rule). The
// hidden def is global, like every def, so the body sees the lambda's own parameters but not
// the locals of an enclosing lambda; a pattern lambda that needs one takes it as a parameter.
// A value no pattern matches is the loud error of any `match`.

// rewritePatternLambdas rewrites every pattern lambda in line and registers the hidden match
// defs through define. A line without pattern lambdas is returned unchanged. A pattern that is
// not valid, or a tuple/list pattern among several parameters, is an error naming it.
func rewritePatternLambdas(line string, isCtor func(string) bool, define func(name, def string)) (string, error) {
	if !strings.Contains(line, "\\") {
		return line, nil
	}
	runes := []rune(line)
	var out strings.Builder
	for r := 0; r < len(runes); r++ {
		c := runes[r]
		if c == '"' || c == '\'' {
			if end, ok := copyLiteral(runes, r, &out); ok {
				r = end
				continue
			}
		}
		if c != '\\' {
			out.WriteRune(c)
			continue
		}
		arrow := lambdaArrow(runes, r+1)
		if arrow < 0 {
			out.WriteRune(c)
			continue
		}
		params := splitTopLevelSpaces(string(runes[r+1 : arrow]))
		if len(params) == 0 || !hasPatternParam(params, isCtor) {
			out.WriteRune(c)
			continue
		}
		bodyEnd := lambdaBodyEnd(runes, arrow+2)
		body, err := rewritePatternLambdas(strings.TrimSpace(string(runes[arrow+2:bodyEnd])), isCtor, define)
		if err != nil {
			return "", err
		}
		text, err := hoistPatternLambda(params, body, define)
		if err != nil {
			return "", err
		}
		out.WriteString(text)
		r = bodyEnd - 1
	}
	return out.String(), nil
}

// hasPatternParam reports whether any parameter is a pattern rather than a name.
func hasPatternParam(params []string, isCtor func(string) bool) bool {
	for _, p := range params {
		if !isIdent(p) || isBoolWord(p) || (isCtor != nil && isCtor(p)) {
			return true
		}
	}
	return false
}

// hoistPatternLambda builds the closure text for one pattern lambda and registers its match.
func hoistPatternLambda(params []string, body string, define func(name, def string)) (string, error) {
	if len(params) > 1 {
		for _, p := range params {
			if strings.HasPrefix(p, "[") || strings.HasPrefix(p, "(") {
				return "", fmt.Errorf("lambda: pattern %q among several parameters must be _, a name or a literal (a list or tuple pattern needs a lambda of its own)", p)
			}
		}
	}
	pat := params[0]
	if len(params) > 1 {
		pat = "[" + strings.Join(params, ", ") + "]"
	}
	if _, err := compileArmPattern(pat); err != nil {
		return "", fmt.Errorf("lambda: %w", err)
	}
	if body == "" {
		return "", fmt.Errorf("lambda: pattern %q has no body after `->`", strings.Join(params, " "))
	}
	body = armBody(body)
	arm := pat + " -> " + body
	h := fnv.New32a()
	h.Write([]byte(arm))
	hidden := fmt.Sprintf("__lam%08x", h.Sum32())
	define(hidden, "match { "+arm+" }")

	names := make([]string, len(params))
	for k := range names {
		names[k] = fmt.Sprintf("_p%d", k+1)
	}
	arg := names[0]
	if len(names) > 1 {
		arg = "([" + strings.Join(names, ", ") + "])"
	}
	return "(\\" + strings.Join(names, " ") + " -> " + hidden + " " + arg + ")", nil
}

// lambdaArrow returns the index of the `->` that ends the parameters of the lambda whose
// parameters start at from (outside brackets and literals), or -1.
func lambdaArrow(runes []rune, from int) int {
	depth := 0
	var sink strings.Builder
	for r := from; r < len(runes); r++ {
		switch c := runes[r]; c {
		case '"', '\'':
			if end, ok := copyLiteral(runes, r, &sink); ok {
				r = end
			}
		case '(', '[', '{':
			depth++
		case ')', ']', '}':
			depth--
			if depth < 0 {
				return -1
			}
		case '-':
			if depth == 0 && r+1 < len(runes) && runes[r+1] == '>' && r > from {
				return r
			}
		}
	}
	return -1
}

// lambdaBodyEnd returns the index just past the lambda body that starts at from: the body runs
// to the end of the line, or to the closer or comma that ends the group the lambda sits in.
func lambdaBodyEnd(runes []rune, from int) int {
	depth := 0
	var sink strings.Builder
	for r := from; r < len(runes); r++ {
		switch c := runes[r]; c {
		case '"', '\'':
			if end, ok := copyLiteral(runes, r, &sink); ok {
				r = end
			}
		case '(', '[', '{':
			depth++
		case ')', ']', '}':
			depth--
			if depth < 0 {
				return r
			}
		case ',':
			if depth == 0 {
				return r
			}
		}
	}
	return len(runes)
}

// rewriteLine applies the statement-level desugarings that need the interpreter's defs.
func (i *Interpreter) rewriteLine(line string) (string, error) {
	var isCtor func(string) bool
	if i.Parser != nil {
		isCtor = i.Parser.IsConstructor
	}
	return rewritePatternLambdas(line, isCtor, func(name, def string) {
		if i.Funcs[name] != def {
			i.Funcs[name] = def
			i.invalidateCompiledCache()
		}
	})
}
