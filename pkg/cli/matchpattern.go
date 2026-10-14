package cli

import (
	"fmt"
	"strings"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/functions"
	"github.com/joergreinhardt/gatwd/parser"
)

// This file is the arm matcher of `match { pat -> body | … }`.
//
// A pattern is compiled once, when the step is built (compileArmPattern), so a pattern that
// means nothing is an error naming it at build time instead of an arm that silently never
// matches. A compiled pattern is then tried against each value by matchArm. The vocabulary:
//
//	_                   the wildcard: matches anything, binds nothing
//	name                a constructor of a declared sum type (routes on the tag), otherwise
//	                    a variable that matches anything and is bound to the value
//	[]                  the empty List
//	[h, ...t]           a non-empty List: head element pattern h, tail bound to t
//	[p1, p2] / (p1, p2) a fixed-length list or tuple whose elements are element patterns
//	LITERAL             a number, "string", 'c', true/false, a Duration, Time or Date
//
// An element pattern (inside brackets, or the head of [h, ...t]) is `_`, a variable name, or
// a LITERAL. A LITERAL matches when the value is == the literal under the language's own
// equality (functions.StructuralEqual): the pattern `0` is the Int 0, so it matches 0 and
// 0.0 and never "0", '0', 10 or 20; `"a"` is the String a; `'a'` the Char a (a one-character
// String equals a Char, D1); `5s` the Duration. There is no substring matching: the text of
// a value never decides which arm runs.

// armKind classifies a compiled arm pattern.
type armKind int

const (
	armWild     armKind = iota + 1 // _
	armIdent                       // a constructor of a sum type, or a variable (decided per run)
	armLit                         // a literal compared with ==
	armEmpty                       // []
	armHeadTail                    // [h, ...t]
	armFixed                       // [p1, p2, …] or (p1, p2, …)
)

// elemKind classifies one element pattern.
type elemKind int

const (
	elemWild elemKind = iota + 1 // _
	elemVar                      // a name that binds the element
	elemLit                      // a literal compared with ==
)

// elemPat is a pattern for one element of a list or tuple.
type elemPat struct {
	kind elemKind
	name string // elemVar: the name bound
	lit  any    // elemLit: the value compared with ==
}

// armPattern is a compiled arm pattern.
type armPattern struct {
	kind  armKind
	text  string    // the pattern as written, for messages
	name  string    // armIdent: the name; armHeadTail: the tail variable ("_" binds nothing)
	lit   any       // armLit: the value compared with ==
	elems []elemPat // armFixed: one per element; armHeadTail: the single head pattern
}

// parsePatternLiteral reports whether text is a literal pattern and returns its value. It
// reads the text as the expression lexer does, so a pattern is typed exactly like the same
// token anywhere else in a script: `0` is an Int, `-1` and `2.5` numbers, `"a"` a String,
// `'a'` a Char, `5s` a Duration, `@2026-10-13` a Date, and `true`/`false` Bools. An
// expression (`1 + 2`), a name or a collection is not a literal.
func parsePatternLiteral(text string) (any, bool) {
	switch text {
	case "true":
		return true, true
	case "false":
		return false, true
	}
	expr, err := parser.ParseInfixExpr(text)
	if err != nil {
		return nil, false
	}
	switch e := expr.(type) {
	case parser.LiteralExpr:
		return e.Value, true
	case parser.UnaryExpr:
		// A negated non-number literal (`-5s`); a negated number is already one literal.
		if _, ok := e.Operand.(parser.LiteralExpr); ok && e.Op == "-" {
			if v, err := e.Eval(nil); err == nil {
				return v, true
			}
		}
	}
	return nil, false
}

// splitPatternList splits the interior of `[…]` or `(…)` at its top-level commas; a comma
// inside a String or Char literal or a nested bracket belongs to the element.
func splitPatternList(inner string) []string {
	var parts []string
	var cur strings.Builder
	depth := 0
	runes := []rune(inner)
	for k := 0; k < len(runes); k++ {
		if last, ok := copyLiteral(runes, k, &cur); ok {
			k = last
			continue
		}
		r := runes[k]
		switch r {
		case '[', '(':
			depth++
		case ']', ')':
			depth--
		case ',':
			if depth == 0 {
				parts = append(parts, strings.TrimSpace(cur.String()))
				cur.Reset()
				continue
			}
		}
		cur.WriteRune(r)
	}
	return append(parts, strings.TrimSpace(cur.String()))
}

func isBoolWord(s string) bool { return s == "true" || s == "false" }

// compileElemPattern compiles one element pattern: `_`, a name, or a literal.
func compileElemPattern(text string) (elemPat, error) {
	switch {
	case text == "_":
		return elemPat{kind: elemWild}, nil
	case isIdent(text) && !isBoolWord(text):
		return elemPat{kind: elemVar, name: text}, nil
	}
	if lit, ok := parsePatternLiteral(text); ok {
		return elemPat{kind: elemLit, lit: lit}, nil
	}
	if strings.HasPrefix(text, "[") || strings.HasPrefix(text, "(") {
		return elemPat{}, fmt.Errorf("nested patterns are not supported yet (element %q)", text)
	}
	return elemPat{}, fmt.Errorf("element %q is not _, a name or a literal", text)
}

// badPattern is the loud error for a pattern that is none of the supported forms.
func badPattern(text, why string) error {
	if strings.Contains(" "+text+" ", " if ") {
		return fmt.Errorf("match: guards are not supported yet: pattern %q", text)
	}
	if why == "" {
		why = "a pattern is _, a name, a literal, [], [h, ...t], or a list or tuple of names and literals"
	}
	return fmt.Errorf("match: invalid pattern %q: %s", text, why)
}

// compileArmPattern compiles the text left of `->` into an armPattern, or returns the error
// that names the pattern. It does not consult the interpreter: whether a name is a sum-type
// constructor is decided when a value arrives (matchArm), as a `def` is re-read at its use.
func compileArmPattern(text string) (armPattern, error) {
	text = strings.TrimSpace(text)
	p := armPattern{text: text}
	if text == "" {
		return p, fmt.Errorf("match: an arm has an empty pattern")
	}
	switch {
	case text == "_":
		p.kind = armWild
		return p, nil
	case strings.HasPrefix(text, "[") && strings.HasSuffix(text, "]") || strings.HasPrefix(text, "(") && strings.HasSuffix(text, ")"):
		inner := strings.TrimSpace(text[1 : len(text)-1])
		if text[0] == '[' && inner == "" {
			p.kind = armEmpty
			return p, nil
		}
		parts := splitPatternList(inner)
		if text[0] == '[' {
			if last := parts[len(parts)-1]; strings.HasPrefix(last, "...") {
				return compileHeadTail(p, parts)
			}
		}
		for _, part := range parts {
			if strings.HasPrefix(part, "...") {
				return p, badPattern(text, "a tail ...t must be the last element of a [h, ...t] pattern")
			}
			e, err := compileElemPattern(part)
			if err != nil {
				return p, badPattern(text, err.Error())
			}
			p.elems = append(p.elems, e)
		}
		p.kind = armFixed
		return p, nil
	case isIdent(text) && !isBoolWord(text):
		p.kind, p.name = armIdent, text
		return p, nil
	}
	if lit, ok := parsePatternLiteral(text); ok {
		p.kind, p.lit = armLit, lit
		return p, nil
	}
	return p, badPattern(text, "")
}

// compileHeadTail compiles `[h, ...t]`: one head element pattern and a tail variable.
func compileHeadTail(p armPattern, parts []string) (armPattern, error) {
	if len(parts) != 2 {
		return p, badPattern(p.text, "a head-tail pattern has exactly one head: [h, ...t]")
	}
	tail := strings.TrimSpace(strings.TrimPrefix(parts[1], "..."))
	if tail != "_" && !(isIdent(tail) && !isBoolWord(tail)) {
		return p, badPattern(p.text, fmt.Sprintf("the tail %q is not a name or _", parts[1]))
	}
	head, err := compileElemPattern(parts[0])
	if err != nil {
		return p, badPattern(p.text, err.Error())
	}
	p.kind, p.name, p.elems = armHeadTail, tail, []elemPat{head}
	return p, nil
}

// matchElem tries one element pattern against an element, recording what it binds.
func matchElem(e elemPat, elem any, binds map[string]any) bool {
	switch e.kind {
	case elemVar:
		binds[e.name] = elem
		return true
	case elemLit:
		return functions.StructuralEqual(elem, e.lit)
	}
	return true // elemWild
}

// matchArm tries a compiled pattern against a value. On a match it returns the interpreter
// the arm's body must run in — i itself, or (for a pattern that binds names) a fork in which
// they are bound as runtime nodes of the given DSL type — and true.
func (i *Interpreter) matchArm(p armPattern, val any, flag dsltypes.BitFlag) (*Interpreter, bool) {
	binds := map[string]any{}
	scoped := false
	switch p.kind {
	case armWild:
	case armLit:
		if !functions.StructuralEqual(val, p.lit) {
			return nil, false
		}
	case armEmpty:
		if list, ok := val.(*collections.List[any]); !ok || !list.IsEmpty() {
			return nil, false
		}
	case armHeadTail:
		list, ok := val.(*collections.List[any])
		if !ok || list.IsEmpty() {
			return nil, false
		}
		head, _ := list.Head()
		tail, _ := list.Tail()
		if !matchElem(p.elems[0], head, binds) {
			return nil, false
		}
		if p.name != "_" {
			binds[p.name] = any(tail)
		}
		scoped = true
	case armFixed:
		elems, ok := toListElements(val)
		if !ok || len(elems) != len(p.elems) {
			return nil, false
		}
		for k, e := range p.elems {
			if !matchElem(e, elems[k], binds) {
				return nil, false
			}
		}
		scoped = true
	case armIdent:
		if i.Parser.IsConstructor(p.name) {
			// A registered constructor discriminates on the value's variant tag; one whose tag
			// does not match this value falls through to the next arm. A value that flows as
			// its bare tag string (the interpreter spawn path) is compared by exact tag.
			if variant, ok := val.(dsltypes.Variant); ok {
				if !variant.Is(p.name) {
					return nil, false
				}
			} else if fmt.Sprintf("%v", val) != p.name {
				return nil, false
			}
			break
		}
		binds[p.name] = val // a fresh variable: matches any value
		scoped = true
	}
	if !scoped {
		return i, true
	}
	local := i.fork()
	for name, v := range binds {
		local.RuntimeNodes[name] = dsltypes.NewASTNode(functions.Right[error, any](v), flag)
	}
	return local, true
}

// splitArmGuard splits `pattern if condition` at the first top-level ` if ` (outside String
// and Char literals) into the pattern and its guard. hasGuard is false for a plain pattern.
func splitArmGuard(arm string) (pattern, guard string, hasGuard bool) {
	runes := []rune(arm)
	var sink strings.Builder
	for k := 0; k < len(runes); k++ {
		if last, ok := copyLiteral(runes, k, &sink); ok {
			k = last
			continue
		}
		if runes[k] == ' ' && strings.HasPrefix(string(runes[k:]), " if ") {
			return strings.TrimSpace(string(runes[:k])), strings.TrimSpace(string(runes[k+4:])), true
		}
	}
	return strings.TrimSpace(arm), "", false
}

// evalGuard runs a guard condition in the scope of an arm's bindings and requires a Bool: a
// guard that errors, yields nothing or yields a non-Bool is an error naming it, never a
// silent fall-through.
func (i *Interpreter) evalGuard(cond string, node dsltypes.ASTNode[functions.Either[error, any]]) (bool, error) {
	for res := range i.pipelineToAdapter(cond)(node) {
		if res.Value.IsLeft() {
			err, _ := res.Value.UnwrapLeft()
			return false, err
		}
		v, _ := res.Value.UnwrapRight()
		b, ok := v.(bool)
		if !ok {
			return false, fmt.Errorf("guard %q must be a Bool, got %v (%T)", cond, v, v)
		}
		return b, nil
	}
	return false, fmt.Errorf("guard %q yielded no value", cond)
}
