package parser

import (
	"fmt"
	"iter"
	"strings"
	"unicode"

	"github.com/joergreinhardt/gatwd/functions"
)

// EBNFNode represents a parsed node in a grammar syntax tree.
type EBNFNode struct {
	Rule     string     `json:"rule"`
	Text     string     `json:"text,omitempty"`
	Children []EBNFNode `json:"children,omitempty"`
}

// EBNFExpr represents a production rule expression.
type EBNFExpr interface {
	Compile(rules map[string]EBNFExpr) Combinator[EBNFNode]
}

// EBNFLiteral matches an exact terminal string.
type EBNFLiteral struct {
	Value string
}

// Compile turns the terminal literal into a Combinator that matches its exact text
// after leading whitespace (via Token) and yields a single "terminal" EBNFNode whose
// Text is the matched string. The rules map is unused because a literal has no
// references to resolve. It satisfies the EBNFExpr interface.
func (l EBNFLiteral) Compile(rules map[string]EBNFExpr) Combinator[EBNFNode] {
	return Map(Token(l.Value), func(matched string) EBNFNode {
		return EBNFNode{Rule: "terminal", Text: matched}
	})
}

// EBNFNonTerminal references another production rule by name.
type EBNFNonTerminal struct {
	Name string
}

// Compile resolves the named production in rules and compiles it on demand, wrapping
// each resulting child in an EBNFNode whose Rule is the non-terminal's Name. The
// lookup happens lazily inside the returned Combinator (not at compile time), which is
// what lets mutually recursive grammars terminate: a rule only expands its references
// as input is consumed. If the referenced rule is absent the Combinator fails cleanly
// by yielding nothing. It satisfies the EBNFExpr interface.
func (nt EBNFNonTerminal) Compile(rules map[string]EBNFExpr) Combinator[EBNFNode] {
	return func(input string) iter.Seq[functions.Pair[EBNFNode, string]] {
		target, ok := rules[nt.Name]
		if !ok {
			// Rule not found: fail cleanly
			return func(yield func(functions.Pair[EBNFNode, string]) bool) {}
		}
		comb := target.Compile(rules)
		return Map(comb, func(child EBNFNode) EBNFNode {
			return EBNFNode{Rule: nt.Name, Children: []EBNFNode{child}}
		})(input)
	}
}

// EBNFSeq matches a sequence of expressions (Expr1 Expr2 ...).
type EBNFSeq struct {
	Items []EBNFExpr
}

// Compile produces a Combinator that matches every item in order and collects the
// results into a single "sequence" EBNFNode whose Children are the sub-matches. An
// empty sequence matches the empty string and yields an "empty" node; a single item
// is compiled directly without an enclosing sequence wrapper. For two or more items it
// threads the remaining input through each item recursively, committing to the first
// success of each (deterministic, no backtracking across items). It satisfies the
// EBNFExpr interface.
func (s EBNFSeq) Compile(rules map[string]EBNFExpr) Combinator[EBNFNode] {
	if len(s.Items) == 0 {
		return func(input string) iter.Seq[functions.Pair[EBNFNode, string]] {
			return func(yield func(functions.Pair[EBNFNode, string]) bool) {
				yield(functions.Pair[EBNFNode, string]{Fst: EBNFNode{Rule: "empty"}, Snd: input})
			}
		}
	}
	if len(s.Items) == 1 {
		return s.Items[0].Compile(rules)
	}

	return func(input string) iter.Seq[functions.Pair[EBNFNode, string]] {
		return func(yield func(functions.Pair[EBNFNode, string]) bool) {
			var parseRemaining func(items []EBNFExpr, currInput string, acc []EBNFNode)
			parseRemaining = func(items []EBNFExpr, currInput string, acc []EBNFNode) {
				if len(items) == 0 {
					yield(functions.Pair[EBNFNode, string]{
						Fst: EBNFNode{Rule: "sequence", Children: acc},
						Snd: currInput,
					})
					return
				}
				headComb := items[0].Compile(rules)
				for res := range headComb(currInput) {
					newAcc := append(append([]EBNFNode{}, acc...), res.Fst)
					parseRemaining(items[1:], res.Snd, newAcc)
					break // Deterministic first success
				}
			}
			parseRemaining(s.Items, input, nil)
		}
	}
}

// EBNFChoice matches alternative expressions (Expr1 | Expr2).
type EBNFChoice struct {
	Alternatives []EBNFExpr
}

// Compile folds the alternatives into a single Combinator using Or, so the choice
// tries each alternative in declaration order and yields the first that succeeds.
// An empty set of alternatives compiles to a parser that always fails (yields
// nothing). It satisfies the EBNFExpr interface.
func (c EBNFChoice) Compile(rules map[string]EBNFExpr) Combinator[EBNFNode] {
	if len(c.Alternatives) == 0 {
		return func(input string) iter.Seq[functions.Pair[EBNFNode, string]] {
			return func(yield func(functions.Pair[EBNFNode, string]) bool) {}
		}
	}
	current := c.Alternatives[0].Compile(rules)
	for i := 1; i < len(c.Alternatives); i++ {
		nextComb := c.Alternatives[i].Compile(rules)
		current = Or(current, nextComb)
	}
	return current
}

// EBNFRepeat matches 0 or more occurrences { Expr }.
type EBNFRepeat struct {
	Item EBNFExpr
}

// Compile builds a Combinator that greedily matches zero or more occurrences of the
// inner item (via Many) and gathers them into a single "repeat" EBNFNode whose
// Children are the matches in order. Because Many always succeeds (possibly with zero
// matches), a repetition never fails. It satisfies the EBNFExpr interface.
func (r EBNFRepeat) Compile(rules map[string]EBNFExpr) Combinator[EBNFNode] {
	return func(input string) iter.Seq[functions.Pair[EBNFNode, string]] {
		return func(yield func(functions.Pair[EBNFNode, string]) bool) {
			childComb := r.Item.Compile(rules)
			manyComb := Many(childComb)
			for res := range manyComb(input) {
				yield(functions.Pair[EBNFNode, string]{
					Fst: EBNFNode{Rule: "repeat", Children: res.Fst},
					Snd: res.Snd,
				})
			}
		}
	}
}

// EBNFOptional matches 0 or 1 occurrence [ Expr ].
type EBNFOptional struct {
	Item EBNFExpr
}

// Compile builds a Combinator that matches zero or one occurrence of the inner item.
// On a match it yields an "optional" EBNFNode wrapping the single child and the
// remaining input; when the item does not match it still succeeds, yielding an empty
// "optional" node and the input unchanged. It satisfies the EBNFExpr interface.
func (o EBNFOptional) Compile(rules map[string]EBNFExpr) Combinator[EBNFNode] {
	return func(input string) iter.Seq[functions.Pair[EBNFNode, string]] {
		return func(yield func(functions.Pair[EBNFNode, string]) bool) {
			childComb := o.Item.Compile(rules)
			matched := false
			for res := range childComb(input) {
				matched = true
				if !yield(functions.Pair[EBNFNode, string]{
					Fst: EBNFNode{Rule: "optional", Children: []EBNFNode{res.Fst}},
					Snd: res.Snd,
				}) {
					return
				}
				break
			}
			if !matched {
				yield(functions.Pair[EBNFNode, string]{
					Fst: EBNFNode{Rule: "optional"},
					Snd: input,
				})
			}
		}
	}
}

// EBNFGrammar represents a complete, compiled EBNF grammar specification.
type EBNFGrammar struct {
	StartRule string
	Rules     map[string]EBNFExpr
}

// Parse parses an input string using the grammar's start rule.
func (g *EBNFGrammar) Parse(input string) (*EBNFNode, string, error) {
	rule, ok := g.Rules[g.StartRule]
	if !ok {
		return nil, "", fmt.Errorf("start rule %q not found in grammar", g.StartRule)
	}
	comb := rule.Compile(g.Rules)
	for res := range comb(input) {
		return &res.Fst, res.Snd, nil
	}
	return nil, "", fmt.Errorf("failed to parse input according to rule %q", g.StartRule)
}

// ParseEBNFGrammar parses standard W3C/ISO EBNF text into an executable EBNFGrammar.
// Supported syntax:
//
//	Rule = Expr ;
//	"literal"
//	NonTerminal
//	Expr1 | Expr2
//	Expr1 Expr2
//	[ Optional ]
//	{ Repetition }
func ParseEBNFGrammar(startRule, ebnfText string) (*EBNFGrammar, error) {
	rules := make(map[string]EBNFExpr)
	// Replace escaped quotes if passed as JSON/string literal
	ebnfText = strings.ReplaceAll(ebnfText, `\"`, `"`)

	// Split by newline or semicolon
	rawStatements := strings.FieldsFunc(ebnfText, func(r rune) bool {
		return r == '\n' || r == ';'
	})

	for _, rawLine := range rawStatements {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "--") || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		ruleName := strings.TrimSpace(parts[0])
		exprStr := strings.TrimSpace(parts[1])

		expr, err := parseEBNFExpression(exprStr)
		if err != nil {
			return nil, fmt.Errorf("rule %q: %w", ruleName, err)
		}
		rules[ruleName] = expr
	}

	if _, ok := rules[startRule]; !ok {
		return nil, fmt.Errorf("start rule %q not defined in grammar", startRule)
	}

	return &EBNFGrammar{
		StartRule: startRule,
		Rules:     rules,
	}, nil
}

func parseEBNFExpression(input string) (EBNFExpr, error) {
	// 1. Check for Choice '|' (outside quotes)
	parts := splitEBNFChoice(input)
	if len(parts) > 1 {
		var alts []EBNFExpr
		for _, part := range parts {
			alt, err := parseEBNFSeq(part)
			if err != nil {
				return nil, err
			}
			alts = append(alts, alt)
		}
		return EBNFChoice{Alternatives: alts}, nil
	}

	return parseEBNFSeq(input)
}

func splitEBNFChoice(input string) []string {
	var parts []string
	var curr strings.Builder
	inQuotes := false

	for _, r := range input {
		if r == '"' {
			inQuotes = !inQuotes
		}
		if r == '|' && !inQuotes {
			parts = append(parts, strings.TrimSpace(curr.String()))
			curr.Reset()
			continue
		}
		curr.WriteRune(r)
	}
	if curr.Len() > 0 {
		parts = append(parts, strings.TrimSpace(curr.String()))
	}
	return parts
}

func parseEBNFSeq(input string) (EBNFExpr, error) {
	tokens := tokenizeEBNFLine(input)
	if len(tokens) == 0 {
		return EBNFSeq{}, nil
	}

	var items []EBNFExpr
	for _, tok := range tokens {
		tok = strings.TrimSpace(tok)
		if strings.HasPrefix(tok, "\"") && strings.HasSuffix(tok, "\"") && len(tok) >= 2 {
			// Literal
			lit := tok[1 : len(tok)-1]
			items = append(items, EBNFLiteral{Value: lit})
		} else if strings.HasPrefix(tok, "[") && strings.HasSuffix(tok, "]") {
			// Optional
			inner := strings.TrimSpace(tok[1 : len(tok)-1])
			innerExpr, err := parseEBNFExpression(inner)
			if err != nil {
				return nil, err
			}
			items = append(items, EBNFOptional{Item: innerExpr})
		} else if strings.HasPrefix(tok, "{") && strings.HasSuffix(tok, "}") {
			// Repeat
			inner := strings.TrimSpace(tok[1 : len(tok)-1])
			innerExpr, err := parseEBNFExpression(inner)
			if err != nil {
				return nil, err
			}
			items = append(items, EBNFRepeat{Item: innerExpr})
		} else {
			// Non-terminal identifier
			items = append(items, EBNFNonTerminal{Name: tok})
		}
	}

	if len(items) == 1 {
		return items[0], nil
	}
	return EBNFSeq{Items: items}, nil
}

func tokenizeEBNFLine(input string) []string {
	var tokens []string
	runes := []rune(strings.TrimSpace(input))
	n := len(runes)
	i := 0

	for i < n {
		r := runes[i]
		if unicode.IsSpace(r) {
			i++
			continue
		}

		if r == '"' {
			start := i
			i++
			for i < n && runes[i] != '"' {
				i++
			}
			if i < n {
				i++ // include closing quote
			}
			tokens = append(tokens, string(runes[start:i]))
			continue
		}

		if r == '[' || r == '{' {
			closing := ']'
			if r == '{' {
				closing = '}'
			}
			start := i
			i++
			depth := 1
			for i < n && depth > 0 {
				if runes[i] == r {
					depth++
				} else if runes[i] == closing {
					depth--
				}
				i++
			}
			tokens = append(tokens, string(runes[start:i]))
			continue
		}

		// Non-terminal word
		start := i
		for i < n && !unicode.IsSpace(runes[i]) && runes[i] != '|' && runes[i] != '[' && runes[i] != ']' && runes[i] != '{' && runes[i] != '}' {
			i++
		}
		if start < i {
			tokens = append(tokens, string(runes[start:i]))
		}
	}

	return tokens
}
