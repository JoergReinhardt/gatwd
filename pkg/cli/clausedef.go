package cli

import (
	"fmt"
	"strings"
)

// Clause-style definitions (Haskell layout): consecutive `def NAME <patterns> = body` lines
// with the same NAME form one function.
//
//	def fact 0 = 1
//	def fact n = n * (fact (n - 1))
//	def add x y if x > y = x
//
// The group is desugared, in the statement front end, into two ordinary defs: a
// `match` def over the tuple of the arguments (`NAME__clauses`, one arm per clause, in
// order, using the existing tuple patterns and guards) and the curried closure
// `def NAME = \a1 … aN -> NAME__clauses ([a1, …, aN])`, so partial application, `map (add 1)`
// and the function form all work. Clauses of different arity, a gap between the clauses of one
// name, and a later redefinition are loud errors.

type clauseGroup struct {
	arity int
	arms  []string
}

// clauseState is the interpreter's record of clause groups; open names the group a following
// clause may extend (any other statement closes it).
type clauseState struct {
	groups map[string]*clauseGroup
	open   string
}

// tryClauseDef handles line when it is a clause definition and reports whether it did.
// Every other statement closes the open group.
func (i *Interpreter) tryClauseDef(line string) bool {
	name, pats, guard, body, isClause := splitClause(line)
	if !isClause {
		if i.clauses != nil {
			i.clauses.open = ""
		}
		return false
	}
	if i.clauses == nil {
		i.clauses = &clauseState{groups: map[string]*clauseGroup{}}
	}
	cs := i.clauses
	fail := func(format string, a ...any) bool {
		fmt.Fprintf(i.out(), "∷ [ERROR] def %s: %s\n", name, fmt.Sprintf(format, a...))
		cs.open = ""
		return true
	}
	if body == "" {
		return fail("a clause needs a body after `=`")
	}
	g, exists := cs.groups[name]
	switch {
	case exists && cs.open != name:
		return fail("clauses of %q must be consecutive; a statement came between them or %q is redefined (join the clauses or use another name)", name, name)
	case exists && g.arity != len(pats):
		return fail("clause has %d parameter(s) but earlier clauses of %q have %d", len(pats), name, g.arity)
	case !exists:
		g = &clauseGroup{arity: len(pats)}
		cs.groups[name] = g
	}
	pat := pats[0]
	if len(pats) > 1 {
		pat = "[" + strings.Join(pats, ", ") + "]"
	}
	arm := pat
	if guard != "" {
		arm += " if " + guard
	}
	// A multi-token body is an expression: parenthesise it so the matched value is not
	// appended as an extra, data-last argument to a bare call (`add x y`).
	body = armBody(body)
	g.arms = append(g.arms, arm+" -> "+body)
	cs.open = name

	params := make([]string, g.arity)
	for k := range params {
		params[k] = fmt.Sprintf("a%d", k+1)
	}
	arg := params[0]
	if g.arity > 1 {
		arg = "([" + strings.Join(params, ", ") + "])"
	}
	hidden := name + "__clauses"
	i.Funcs[hidden] = "match { " + strings.Join(g.arms, " | ") + " }"
	i.Funcs[name] = "\\" + strings.Join(params, " ") + " -> " + hidden + " " + arg
	i.invalidateCompiledCache()
	return true
}

// splitClause splits `def NAME p1 … pN [if guard] = body` into its parts; isClause is false
// for every other statement (including `def NAME = body`, which has no parameters).
func splitClause(line string) (name string, pats []string, guard, body string, isClause bool) {
	if !strings.HasPrefix(line, "def ") {
		return
	}
	rest := line[4:]
	eq := findAssign(rest)
	if eq < 0 {
		return
	}
	toks := splitTopLevelSpaces(strings.TrimSpace(rest[:eq]))
	if len(toks) < 2 || !isDefName(toks[0]) || strings.HasPrefix(toks[1], ":") {
		return
	}
	body = strings.TrimSpace(rest[eq+1:])
	name = toks[0]
	toks = toks[1:]
	for k, t := range toks {
		if t == "if" {
			guard = strings.Join(toks[k+1:], " ")
			toks = toks[:k]
			break
		}
	}
	if len(toks) == 0 {
		return
	}
	return name, toks, guard, body, true
}

// findAssign returns the index of the first lone `=` (not part of ==, !=, <=, >=, =>) outside
// brackets and String/Char literals, or -1.
func findAssign(s string) int {
	runes := []rune(s)
	var sink strings.Builder
	depth := 0
	idx := func(r int) int { return len(string(runes[:r])) }
	for r := 0; r < len(runes); r++ {
		switch c := runes[r]; c {
		case '"', '\'':
			if end, ok := copyLiteral(runes, r, &sink); ok {
				r = end
			}
		case '(', '[', '{':
			depth++
		case ')', ']', '}':
			depth--
		case '=':
			if depth > 0 {
				continue
			}
			prev, next := rune(0), rune(0)
			if r > 0 {
				prev = runes[r-1]
			}
			if r+1 < len(runes) {
				next = runes[r+1]
			}
			if next == '=' {
				r++
				continue
			}
			if prev == '!' || prev == '<' || prev == '>' || next == '>' {
				continue
			}
			return idx(r)
		}
	}
	return -1
}

// splitTopLevelSpaces splits s on spaces outside brackets and literals.
func splitTopLevelSpaces(s string) []string {
	runes := []rune(s)
	var out []string
	var cur strings.Builder
	depth := 0
	flush := func() {
		if cur.Len() > 0 {
			out = append(out, cur.String())
			cur.Reset()
		}
	}
	for r := 0; r < len(runes); r++ {
		switch c := runes[r]; c {
		case '"', '\'':
			if end, ok := copyLiteral(runes, r, &cur); ok {
				r = end
				continue
			}
			cur.WriteRune(c)
		case '(', '[', '{':
			depth++
			cur.WriteRune(c)
		case ')', ']', '}':
			depth--
			cur.WriteRune(c)
		case ' ', '\t':
			if depth > 0 {
				cur.WriteRune(c)
			} else {
				flush()
			}
		default:
			cur.WriteRune(c)
		}
	}
	flush()
	return out
}

// armBody makes a clause or lambda body usable as a match arm body, which is a pipeline: a
// String or Char literal becomes `echo literal`, and a multi-token expression is parenthesised
// so the matched value is not appended as an extra, data-last argument to a bare call.
func armBody(body string) string {
	switch {
	case strings.Contains(body, "|>"):
		return body
	case isSingleLiteral(body):
		return "echo " + body
	case strings.Contains(body, " "):
		return "(" + body + ")"
	}
	return body
}

// isSingleLiteral reports whether s is exactly one String or Char literal.
func isSingleLiteral(s string) bool {
	runes := []rune(s)
	if len(runes) == 0 || (runes[0] != '"' && runes[0] != '\'') {
		return false
	}
	var sink strings.Builder
	end, ok := copyLiteral(runes, 0, &sink)
	return ok && end == len(runes)-1
}
