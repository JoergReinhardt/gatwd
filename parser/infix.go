package parser

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/functions"
)

// TokenType represents lexical token types for infix expressions.
type TokenType int

// The Tok* constants enumerate the lexical token kinds produced by TokenizeInfix.
const (
	TokEOF       TokenType = iota // TokEOF marks the end of the token stream.
	TokNumber                     // TokNumber is an integer, float, or complex literal.
	TokString                     // TokString is a double-quoted string literal (quotes stripped).
	TokIdent                      // TokIdent is an identifier: variable, constant, or boolean literal.
	TokOp                         // TokOp is a binary operator such as +, *, ==, or &&.
	TokLParen                     // TokLParen is an opening parenthesis '('.
	TokRParen                     // TokRParen is a closing parenthesis ')'.
	TokLBracket                   // TokLBracket is an opening bracket '[' (list or comprehension).
	TokRBracket                   // TokRBracket is a closing bracket ']'.
	TokComma                      // TokComma is an element or clause separator ','.
	TokPipe                       // TokPipe is the list-comprehension bar '|'.
	TokArrow                      // TokArrow is the generator binding arrow "<-".
	TokLambda                     // TokLambda is the lambda introducer '\\'.
	TokThinArrow                  // TokThinArrow is the lambda body arrow "->".
	TokChar                       // TokChar is a single-quoted character literal (Text is the character itself).
)

// TokenItem is a single lexical token: its Type classifies it (one of the Tok*
// constants) and Text holds the matched source text, with any surrounding quotes
// already stripped for string literals.
type TokenItem struct {
	Type TokenType
	Text string
}

// TokenizeInfix splits an expression string into tokens.
func TokenizeInfix(input string) ([]TokenItem, error) {
	var tokens []TokenItem
	runes := []rune(input)
	n := len(runes)
	i := 0

	for i < n {
		r := runes[i]
		if unicode.IsSpace(r) {
			i++
			continue
		}

		if r == '(' {
			tokens = append(tokens, TokenItem{Type: TokLParen, Text: "("})
			i++
			continue
		}
		if r == ')' {
			tokens = append(tokens, TokenItem{Type: TokRParen, Text: ")"})
			i++
			continue
		}
		if r == '[' {
			tokens = append(tokens, TokenItem{Type: TokLBracket, Text: "["})
			i++
			continue
		}
		if r == ']' {
			tokens = append(tokens, TokenItem{Type: TokRBracket, Text: "]"})
			i++
			continue
		}
		if r == ',' {
			tokens = append(tokens, TokenItem{Type: TokComma, Text: ","})
			i++
			continue
		}

		// String literal: "..."
		if r == '"' {
			i++
			start := i
			for i < n && runes[i] != '"' {
				i++
			}
			if i >= n {
				return nil, fmt.Errorf("unterminated string literal")
			}
			tokens = append(tokens, TokenItem{Type: TokString, Text: string(runes[start:i])})
			i++ // skip closing quote
			continue
		}

		// Char literal: 'a', '\n', 'é'
		if r == '\'' {
			c, next, err := lexCharLiteral(runes, i)
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, TokenItem{Type: TokChar, Text: string(c)})
			i = next
			continue
		}

		// Multi-character or single-character operators
		// <-, ->, ++, ==, !=, <=, >=, &&, ||, //
		if i+1 < n {
			two := string(runes[i : i+2])
			if two == "->" {
				tokens = append(tokens, TokenItem{Type: TokThinArrow, Text: "->"})
				i += 2
				continue
			}
			if two == "<-" {
				tokens = append(tokens, TokenItem{Type: TokArrow, Text: "<-"})
				i += 2
				continue
			}
			if two == "++" || two == "==" || two == "!=" || two == "<=" || two == ">=" || two == "&&" || two == "||" || two == "//" {
				tokens = append(tokens, TokenItem{Type: TokOp, Text: two})
				i += 2
				continue
			}
		}

		if r == '\\' {
			tokens = append(tokens, TokenItem{Type: TokLambda, Text: "\\"})
			i++
			continue
		}

		if r == '|' {
			tokens = append(tokens, TokenItem{Type: TokPipe, Text: "|"})
			i++
			continue
		}

		// Single-character operators: +, -, *, /, %, <, >
		if r == '+' || r == '-' || r == '*' || r == '/' || r == '%' || r == '<' || r == '>' {
			isUnary := (len(tokens) == 0 || tokens[len(tokens)-1].Type == TokOp || tokens[len(tokens)-1].Type == TokLParen || tokens[len(tokens)-1].Type == TokLBracket || tokens[len(tokens)-1].Type == TokComma || tokens[len(tokens)-1].Type == TokPipe || tokens[len(tokens)-1].Type == TokArrow || tokens[len(tokens)-1].Type == TokThinArrow)
			if (r == '-' || r == '+') && isUnary && i+1 < n && (unicode.IsDigit(runes[i+1])) {
				// Parse signed number below
			} else {
				tokens = append(tokens, TokenItem{Type: TokOp, Text: string(r)})
				i++
				continue
			}
		}

		// Number (integer, float, or complex literal with 'i')
		if unicode.IsDigit(r) || ((r == '-' || r == '+') && i+1 < n && unicode.IsDigit(runes[i+1])) {
			start := i
			if r == '-' || r == '+' {
				i++
			}
			for i < n && (unicode.IsDigit(runes[i]) || runes[i] == '.' || runes[i] == 'i') {
				i++
			}
			tokens = append(tokens, TokenItem{Type: TokNumber, Text: string(runes[start:i])})
			continue
		}

		// Identifier (variables, constants, boolean literals)
		if unicode.IsLetter(r) || r == '_' {
			start := i
			for i < n && (unicode.IsLetter(runes[i]) || unicode.IsDigit(runes[i]) || runes[i] == '_') {
				i++
			}
			tokens = append(tokens, TokenItem{Type: TokIdent, Text: string(runes[start:i])})
			continue
		}

		return nil, fmt.Errorf("unexpected character in infix expression: %c", r)
	}

	tokens = append(tokens, TokenItem{Type: TokEOF, Text: ""})
	return tokens, nil
}

// lexCharLiteral reads the character literal that opens at runes[start] (a quote) and
// returns the character and the index just past the closing quote. Escapes: \n \t \r \0
// \\ \' \" \xHH \uHHHH. Anything else — an empty literal, two characters, an unknown
// escape, a missing closing quote — is an error naming the problem.
func lexCharLiteral(runes []rune, start int) (rune, int, error) {
	i := start + 1
	n := len(runes)
	if i >= n {
		return 0, 0, fmt.Errorf("unterminated character literal")
	}
	var c rune
	switch runes[i] {
	case '\'':
		return 0, 0, fmt.Errorf("empty character literal ''")
	case '\\':
		i++
		if i >= n {
			return 0, 0, fmt.Errorf("unterminated character literal")
		}
		switch e := runes[i]; e {
		case 'n':
			c = '\n'
		case 't':
			c = '\t'
		case 'r':
			c = '\r'
		case '0':
			c = 0
		case '\\', '\'', '"':
			c = e
		case 'x', 'u':
			width := 2
			if e == 'u' {
				width = 4
			}
			if i+width >= n {
				return 0, 0, fmt.Errorf("unterminated character literal")
			}
			code, err := strconv.ParseUint(string(runes[i+1:i+1+width]), 16, 32)
			if err != nil {
				return 0, 0, fmt.Errorf("invalid \\%c escape in character literal", e)
			}
			if _, ok := functions.CharFromCode(int64(code)); !ok {
				return 0, 0, fmt.Errorf("invalid code point U+%04X in character literal", code)
			}
			c = rune(code)
			i += width
		default:
			return 0, 0, fmt.Errorf("unknown escape \\%c in character literal", e)
		}
	default:
		c = runes[i]
	}
	i++
	if i >= n || runes[i] != '\'' {
		return 0, 0, fmt.Errorf("character literal must hold exactly one character")
	}
	return c, i + 1, nil
}

// Expr represents an abstract syntax tree for infix expressions.
type Expr interface {
	Eval(lookup func(string) (any, bool)) (any, error)
}

// LiteralExpr is a leaf Expr holding an already-parsed constant value, such as an
// int, float, complex, string, or *big.Int.
type LiteralExpr struct {
	Value any
}

// Eval returns the literal's stored Value unchanged. The lookup function is ignored
// because a literal does not depend on any variable binding, and it never fails.
func (e LiteralExpr) Eval(lookup func(string) (any, bool)) (any, error) {
	return e.Value, nil
}

// VarExpr is a leaf Expr that references an identifier by Name, to be resolved at
// evaluation time.
type VarExpr struct {
	Name string
}

// Eval resolves the variable's Name using the lookup function, returning the bound
// value if present. The boolean literals "true" and "false" are recognized directly
// even when lookup is nil or has no binding for them. It returns an error for any
// other undefined identifier.
func (e VarExpr) Eval(lookup func(string) (any, bool)) (any, error) {
	if lookup != nil {
		if val, ok := lookup(e.Name); ok {
			return val, nil
		}
	}
	if e.Name == "true" {
		return true, nil
	}
	if e.Name == "false" {
		return false, nil
	}
	return nil, &UndefinedError{Name: e.Name}
}

// UndefinedError reports an identifier that is neither bound by the lookup nor a
// boolean literal. Callers use errors.As to tell a mere name-resolution miss (the
// word may be a pipeline step) from an operation that was made and failed.
type UndefinedError struct{ Name string }

// Error returns "undefined identifier: <name>".
func (e *UndefinedError) Error() string { return "undefined identifier: " + e.Name }

// BinaryExpr is an interior Expr node applying operator Op to the results of its Left
// and Right sub-expressions. Precedence and associativity are already resolved during
// parsing, so the tree shape alone determines evaluation order.
type BinaryExpr struct {
	Op    functions.OpKind
	Left  Expr
	Right Expr
}

// Eval evaluates the Left and Right operands (short-circuiting on the first error),
// then applies Op via functions.EvalBinaryOp. For the concatenation operator OpConcat
// it first tries the native binary op (for strings and monoids) and, failing that,
// falls back to treating both operands as sequences: it extracts their items and
// rebuilds a single immutable cons list, which gives `++` list-append semantics. The
// lookup function is threaded down to both operands for variable resolution.
func (e BinaryExpr) Eval(lookup func(string) (any, bool)) (any, error) {
	l, err := e.Left.Eval(lookup)
	if err != nil {
		return nil, err
	}
	r, err := e.Right.Eval(lookup)
	if err != nil {
		return nil, err
	}

	var res any
	if e.Op == functions.OpConcat {
		res, err = ConcatValues(l, r)
	} else {
		res, err = functions.EvalBinaryOp(e.Op, l, r)
	}
	if err != nil {
		return nil, &ApplyError{Err: err}
	}
	return res, nil
}

// ConcatValues is the `++` Monoid operation as a Go function: it tries the native
// binary operator first (strings, Sets, HAMTs, Graphs, Trees and the other monoids)
// and, failing that, treats both operands as sequences and appends them into one
// immutable cons list. It is shared by the infix `++` operator and the `concat`
// expression function so the two are behaviorally identical.
func ConcatValues(l, r any) (any, error) {
	if res, err := functions.EvalBinaryOp(functions.OpConcat, l, r); err == nil {
		return res, nil
	}
	leftItems, errL := extractItems(l)
	rightItems, errR := extractItems(r)
	if errL == nil && errR == nil {
		combined := append(append([]any(nil), leftItems...), rightItems...)
		res := collections.Nil[any]()
		for idx := len(combined) - 1; idx >= 0; idx-- {
			res = collections.Cons(combined[idx], res)
		}
		return res, nil
	}
	return functions.EvalBinaryOp(functions.OpConcat, l, r)
}

// ListLiteralExpr represents a list literal: [1, 2, 3]
type ListLiteralExpr struct {
	Elements []Expr
}

// Eval evaluates each element expression and builds an immutable cons list preserving
// source order. Because collections.Cons prepends, the elements are evaluated and
// consed from last to first. It returns the first element evaluation error, if any,
// threading lookup through for variable resolution.
func (e ListLiteralExpr) Eval(lookup func(string) (any, bool)) (any, error) {
	res := collections.Nil[any]()
	// Cons prepends, so reverse to preserve order
	for idx := len(e.Elements) - 1; idx >= 0; idx-- {
		val, err := e.Elements[idx].Eval(lookup)
		if err != nil {
			return nil, err
		}
		res = collections.Cons(val, res)
	}
	return res, nil
}

// Generator represents a single 'var <- source' or '[x, y] <- source' clause in a comprehension.
type Generator struct {
	PatternNames []string
	Source       Expr
}

// ListComprehensionExpr represents [expr | gen1, gen2..., filterExpr...]
type ListComprehensionExpr struct {
	YieldExpr  Expr
	Generators []Generator
	Filters    []Expr
	// FilterAfter[i] is the number of generators written before Filters[i]. Like a
	// Haskell guard, Filters[i] is applied as soon as those generators are bound, so
	// it can protect a later generator that depends on it
	// (`[… | x <- xs, has "k" x, y <- (get "k" x)]`). A nil FilterAfter means every
	// filter runs after all generators are bound.
	FilterAfter []int
}

func extractItems(srcVal any) ([]any, error) {
	switch s := srcVal.(type) {
	case *collections.List[any]:
		var items []any
		for item := range s.Seq() {
			items = append(items, item)
		}
		return items, nil
	case *collections.List[int]:
		var items []any
		for item := range s.Seq() {
			items = append(items, item)
		}
		return items, nil
	case *collections.List[string]:
		var items []any
		for item := range s.Seq() {
			items = append(items, item)
		}
		return items, nil
	case interface{ Elements() []string }:
		var items []any
		for _, x := range s.Elements() {
			items = append(items, x)
		}
		return items, nil
	case interface{ Elements() []any }:
		return s.Elements(), nil
	case functions.Pair[any, any]:
		return []any{s.Fst, s.Snd}, nil
	case interface{ PairElements() (any, any) }:
		f, snd := s.PairElements()
		return []any{f, snd}, nil
	case []any:
		return s, nil
	case []int:
		var items []any
		for _, x := range s {
			items = append(items, x)
		}
		return items, nil
	case []string:
		var items []any
		for _, x := range s {
			items = append(items, x)
		}
		return items, nil
	case string:
		// If string represents a list literal or formatted list, parse it into items
		trimmed := strings.TrimSpace(s)
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			// Convert spaces or commas to standard comma-separated format and evaluate
			inner := strings.TrimSpace(trimmed[1 : len(trimmed)-1])
			if inner == "" {
				return nil, nil
			}
			// Attempt parsing as Infix expression
			if parsedList, err := EvalInfixString(trimmed, nil); err == nil && parsedList != nil {
				return extractItems(parsedList)
			}
		}
		return nil, fmt.Errorf("list comprehension source must be a collection, got string: %s", s)
	case int:
		var items []any
		for i := 0; i < s; i++ {
			items = append(items, i)
		}
		return items, nil
	default:
		return nil, fmt.Errorf("list comprehension source must be a collection, got %T", srcVal)
	}
}

// Eval evaluates the comprehension into an immutable cons list. It runs the
// Generators as nested loops (the first generator is outermost), extending the
// variable scope at each level with that generator's pattern bindings; single-name
// patterns bind the whole element while multi-name patterns destructure it. Once all
// generators are bound, every Filter expression must evaluate to boolean true for the
// YieldExpr to be evaluated and its result collected. A comprehension with no
// generators yields the empty list. Errors from any source, filter, or yield
// evaluation abort and propagate. lookup supplies the enclosing scope's bindings.
func (e ListComprehensionExpr) Eval(lookup func(string) (any, bool)) (any, error) {
	if len(e.Generators) == 0 {
		return collections.Nil[any](), nil
	}

	var collected []any

	// filtersPass applies the guards that become applicable once `bound` generators
	// are bound (all of them when FilterAfter is nil and bound == len(Generators)).
	filtersPass := func(bound int, lookup func(string) (any, bool)) (bool, error) {
		for i, filt := range e.Filters {
			at := len(e.Generators)
			if e.FilterAfter != nil {
				at = e.FilterAfter[i]
			}
			if at != bound {
				continue
			}
			fRes, err := filt.Eval(lookup)
			if err != nil {
				return false, err
			}
			if b, ok := fRes.(bool); !ok || !b {
				return false, nil
			}
		}
		return true, nil
	}

	var runGenerators func(genIdx int, currentLookup func(string) (any, bool)) error
	runGenerators = func(genIdx int, currentLookup func(string) (any, bool)) error {
		passed, err := filtersPass(genIdx, currentLookup)
		if err != nil || !passed {
			return err
		}
		if genIdx >= len(e.Generators) {
			yVal, err := e.YieldExpr.Eval(currentLookup)
			if err != nil {
				return err
			}
			collected = append(collected, yVal)
			return nil
		}

		gen := e.Generators[genIdx]
		srcVal, err := gen.Source.Eval(currentLookup)
		if err != nil {
			return err
		}
		items, err := extractItems(srcVal)
		if err != nil {
			return err
		}

		for _, item := range items {
			boundItem := item
			scopedLookup := func(k string) (any, bool) {
				if len(gen.PatternNames) == 1 && gen.PatternNames[0] == k {
					return boundItem, true
				}
				// Pattern tuple destructuring: [x, y] <- [[1, 2], [3, 4]]
				if len(gen.PatternNames) > 1 {
					subElems, ok := extractItems(boundItem)
					if ok == nil && len(subElems) == len(gen.PatternNames) {
						for pIdx, pName := range gen.PatternNames {
							if pName == k {
								return subElems[pIdx], true
							}
						}
					}
				}
				if currentLookup != nil {
					return currentLookup(k)
				}
				return nil, false
			}
			if err := runGenerators(genIdx+1, scopedLookup); err != nil {
				return err
			}
		}
		return nil
	}

	if err := runGenerators(0, lookup); err != nil {
		return nil, err
	}

	resList := collections.Nil[any]()
	for idx := len(collected) - 1; idx >= 0; idx-- {
		resList = collections.Cons(collected[idx], resList)
	}
	return resList, nil
}

// ApplyExpr is an interior Expr node representing function application by
// juxtaposition: a Head expression applied to one or more Args, as in
// `graph_add_node color id payload g` or `hamt_path "metadata.name" r`. Application
// binds tighter than every binary operator, so `f x + g y` parses as
// `(f x) + (g y)`; nested application requires parentheses — `f (g x)` — exactly as
// in Haskell, which is what lets juxtaposition be unambiguous.
type ApplyExpr struct {
	Head Expr
	Args []Expr
}

// Eval evaluates the Head to a callable, the Args left-to-right, then applies the
// callable to the argument values. The Head must resolve to a functions.DSLFunc
// (the DSL's type-erased callable); applying anything else is an error rather than a
// panic, keeping a mis-application inside the Either[error, any] discipline. Any
// error from the head, an argument, or the call itself short-circuits and
// propagates. lookup is threaded through for variable/function resolution.
func (e ApplyExpr) Eval(lookup func(string) (any, bool)) (any, error) {
	hv, err := e.Head.Eval(lookup)
	if err != nil {
		return nil, err
	}
	if !functions.IsCallable(hv) {
		return nil, &ApplyError{Err: fmt.Errorf("cannot apply non-function value of type %T", hv)}
	}
	args := make([]any, len(e.Args))
	for idx, a := range e.Args {
		av, err := a.Eval(lookup)
		if err != nil {
			return nil, err
		}
		args[idx] = av
	}
	res, err := functions.Apply(hv, args)
	if err != nil {
		var already *ApplyError
		if errors.As(err, &already) {
			return nil, err
		}
		return nil, &ApplyError{Err: err}
	}
	return res, nil
}

// ApplyError marks an error raised by an operation that was actually performed — a
// non-callable head, a callee that returned an error, or an operator that rejected its
// operands (`1 / 0`, `true + 1`) — as opposed to a lexical or name-resolution
// failure (`undefined identifier`). The distinction lets the interpreter fall back
// to other parses when an expression merely does not resolve, yet report loudly when a
// real call was made and failed. Its message is the underlying error's.
type ApplyError struct{ Err error }

// Error returns the underlying error's message unchanged.
func (e *ApplyError) Error() string { return e.Err.Error() }

// Unwrap exposes the underlying error to errors.Is/As.
func (e *ApplyError) Unwrap() error { return e.Err }

// LambdaExpr is an interior Expr node for a lambda abstraction `\x y -> body`. Its
// Eval does not run the body: it returns a *functions.Closure that captures the
// lookup function of the defining environment (lexical scope) and evaluates Body
// later, once all Params are supplied, with the parameters shadowing that
// environment. Because Closure implements currying, `\x y -> …` applied to one
// argument yields a function awaiting y.
type LambdaExpr struct {
	Params []string
	Body   Expr
}

// Eval returns the Closure for this lambda over the given defining environment.
func (e LambdaExpr) Eval(lookup func(string) (any, bool)) (any, error) {
	params, body := e.Params, e.Body
	return functions.NewClosure(params, func(args []any) (any, error) {
		return body.Eval(func(name string) (any, bool) {
			for idx := len(params) - 1; idx >= 0; idx-- {
				if params[idx] == name {
					return args[idx], true
				}
			}
			if lookup != nil {
				return lookup(name)
			}
			return nil, false
		})
	}), nil
}

// InfixParser builds an Expr using precedence climbing.
type InfixParser struct {
	tokens []TokenItem
	pos    int
	// appDisabled suppresses juxtaposition function application while parsing the
	// elements of a PLAIN list literal. gatwd serializes lists space-separated
	// (`[2 1]`, `[nginx web]`) and re-parses them, so inside a bare `[...]` adjacent
	// atoms must stay SEPARATE elements, not fold into `apply 2 to 1`. Parentheses
	// and comprehensions reset it to false, so application still works everywhere a
	// function call is actually intended (lambda bodies, args, comprehension
	// sources/yields). See parseApplication and the TokLBracket branch.
	appDisabled bool
}

// NewInfixParser returns an InfixParser positioned at the start of tokens, ready to
// build an Expr via ParseExpression. The caller typically obtains tokens from
// TokenizeInfix.
func NewInfixParser(tokens []TokenItem) *InfixParser {
	return &InfixParser{tokens: tokens, pos: 0}
}

func (p *InfixParser) peek() TokenItem {
	if p.pos >= len(p.tokens) {
		return TokenItem{Type: TokEOF, Text: ""}
	}
	return p.tokens[p.pos]
}

func (p *InfixParser) next() TokenItem {
	t := p.peek()
	p.pos++
	return t
}

func getPrecedence(op string) int {
	if meta, exists := CanonicalOperators[op]; exists {
		return meta.Precedence
	}
	return 0
}

// ParseExpression parses a full infix expression from the current position and
// returns its Expr AST. It drives precedence-climbing starting at minimum precedence
// 0, so operators bind according to the precedences in CanonicalOperators and lower
// ones form the outer (later-evaluated) nodes of the tree. It returns an error on
// malformed input, such as an unexpected token or unbalanced brackets.
func (p *InfixParser) ParseExpression() (Expr, error) {
	return p.parsePrecedence(0)
}

func (p *InfixParser) parsePrecedence(minPrec int) (Expr, error) {
	left, err := p.parseApplication()
	if err != nil {
		return nil, err
	}

	for {
		tok := p.peek()
		if tok.Type != TokOp {
			break
		}
		prec := getPrecedence(tok.Text)
		if prec <= minPrec {
			break
		}

		p.next() // consume op
		right, err := p.parsePrecedence(prec)
		if err != nil {
			return nil, err
		}

		left = BinaryExpr{
			Op:    functions.OpKind(tok.Text),
			Left:  left,
			Right: right,
		}
	}

	return left, nil
}

// atomStarters are the token types that can begin an atom (a primary expression):
// a number, string, identifier, parenthesised group, or list/comprehension. Two
// adjacent atoms with no operator between them form a function application.
func (p *InfixParser) isAtomStart() bool {
	switch p.peek().Type {
	case TokNumber, TokString, TokChar, TokIdent, TokLParen, TokLBracket:
		return true
	default:
		return false
	}
}

// parseApplication parses a run of juxtaposed atoms as function application:
// the first atom is the callable, each subsequent atom is an argument. A lone atom
// (no trailing atoms) is returned unwrapped, so non-application expressions are
// unaffected. Because an atom never starts with an operator token, this greedily
// consumes only argument atoms and stops at the first operator, giving application
// a higher binding than any infix operator.
func (p *InfixParser) parseApplication() (Expr, error) {
	if p.appDisabled {
		// Inside a plain list literal: juxtaposed atoms are separate elements, so
		// do not fold trailing atoms into an application.
		return p.parsePrimary()
	}
	head, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}
	var args []Expr
	for p.isAtomStart() {
		arg, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
	}
	if len(args) == 0 {
		return head, nil
	}
	return ApplyExpr{Head: head, Args: args}, nil
}

func (p *InfixParser) parsePrimary() (Expr, error) {
	tok := p.next()
	switch tok.Type {
	case TokNumber:
		if strings.HasSuffix(tok.Text, "i") {
			c, err := strconv.ParseComplex(tok.Text, 128)
			if err != nil {
				return nil, err
			}
			return LiteralExpr{Value: c}, nil
		}
		if strings.Contains(tok.Text, ".") {
			f, err := strconv.ParseFloat(tok.Text, 64)
			if err != nil {
				return nil, err
			}
			return LiteralExpr{Value: f}, nil
		}
		i, err := strconv.ParseInt(tok.Text, 10, 64)
		if err != nil {
			// If number exceeds int64, promote directly to big.Int
			n, ok := functions.ToBigInt(tok.Text)
			if ok {
				return LiteralExpr{Value: n}, nil
			}
			return nil, err
		}
		return LiteralExpr{Value: functions.TryNarrowInteger(i)}, nil

	case TokString:
		return LiteralExpr{Value: tok.Text}, nil

	case TokChar:
		r, _ := utf8.DecodeRuneInString(tok.Text)
		return LiteralExpr{Value: functions.Char(r)}, nil

	case TokIdent:
		return VarExpr{Name: tok.Text}, nil

	case TokLambda:
		// `\x y -> body`: the body extends as far right as possible and is an
		// ordinary application context, even inside a plain list literal.
		var params []string
		for p.peek().Type == TokIdent {
			params = append(params, p.next().Text)
		}
		if len(params) == 0 || p.next().Type != TokThinArrow {
			return nil, fmt.Errorf("malformed lambda: expected `\\x y -> body`")
		}
		savedApp := p.appDisabled
		p.appDisabled = false
		body, err := p.parsePrecedence(0)
		p.appDisabled = savedApp
		if err != nil {
			return nil, err
		}
		return LambdaExpr{Params: params, Body: body}, nil

	case TokLParen:
		// A parenthesised group is a fresh application context — `(f x)` applies even
		// when it sits inside a plain list literal — so re-enable application here.
		savedApp := p.appDisabled
		p.appDisabled = false
		expr, err := p.parsePrecedence(0)
		if err != nil {
			return nil, err
		}
		p.appDisabled = savedApp
		if p.next().Type != TokRParen {
			return nil, fmt.Errorf("expected closing parenthesis ')'")
		}
		return expr, nil

	case TokLBracket:
		// Check for empty list []
		if p.peek().Type == TokRBracket {
			p.next()
			return ListLiteralExpr{Elements: nil}, nil
		}

		// Decide comprehension vs. plain list BEFORE parsing the first expression: a
		// comprehension (`[e | …]`) wants application in its yield/sources, a plain
		// list (`[2 1]`, `[a, b]`) wants adjacent atoms kept as separate elements.
		savedApp := p.appDisabled
		p.appDisabled = !p.bracketHasTopLevelPipe()

		firstExpr, err := p.parsePrecedence(0)
		if err != nil {
			return nil, err
		}

		// Check if this is a List Comprehension: [expr | gen1, gen2..., filter1, filter2...]
		if p.peek().Type == TokPipe {
			p.next() // consume '|'

			var generators []Generator
			var filters []Expr
			var filterAfter []int

			for {
				if p.peek().Type == TokRBracket || p.peek().Type == TokEOF {
					break
				}

				// Check if the next segment is a generator 'var <- src' or '[x, y] <- src'
				// Look ahead for TokArrow ("<-") before any comma or closing bracket at top level
				arrowIdx := -1
				bracketDepth := 0
				parenDepth := 0
				for scan := p.pos; scan < len(p.tokens); scan++ {
					tt := p.tokens[scan].Type
					if tt == TokLBracket {
						bracketDepth++
					} else if tt == TokRBracket {
						if bracketDepth == 0 {
							break
						}
						bracketDepth--
					} else if tt == TokLParen {
						parenDepth++
					} else if tt == TokRParen {
						parenDepth--
					} else if tt == TokComma {
						if bracketDepth == 0 && parenDepth == 0 {
							break
						}
					} else if tt == TokArrow && p.tokens[scan].Text == "<-" {
						if bracketDepth <= 1 && parenDepth == 0 {
							arrowIdx = scan
							break
						}
					}
				}

				if arrowIdx != -1 {
					// It's a generator! Collect pattern tokens up to arrowIdx
					var patNames []string
					for p.pos < arrowIdx {
						tok := p.next()
						if tok.Type == TokIdent {
							patNames = append(patNames, tok.Text)
						}
					}
					p.next() // consume '<-'
					srcExpr, err := p.parsePrecedence(0)
					if err != nil {
						return nil, err
					}
					generators = append(generators, Generator{
						PatternNames: patNames,
						Source:       srcExpr,
					})
				} else {
					// Otherwise, it's a filter boolean expression
					filt, err := p.parsePrecedence(0)
					if err != nil {
						return nil, err
					}
					filters = append(filters, filt)
					filterAfter = append(filterAfter, len(generators))
				}

				if p.peek().Type == TokComma {
					p.next() // consume ','
					continue
				}
				break
			}

			if p.next().Type != TokRBracket {
				return nil, fmt.Errorf("expected closing bracket ']' after list comprehension")
			}

			p.appDisabled = savedApp
			return ListComprehensionExpr{
				YieldExpr:   firstExpr,
				Generators:  generators,
				Filters:     filters,
				FilterAfter: filterAfter,
			}, nil
		}

		// Standard List Literal: [elem1, elem2, ...] or [elem1 elem2 ...]
		elements := []Expr{firstExpr}
		for {
			if p.peek().Type == TokRBracket {
				p.next()
				break
			}
			if p.peek().Type == TokComma {
				p.next()
			}
			elem, err := p.parsePrecedence(0)
			if err != nil {
				return nil, err
			}
			elements = append(elements, elem)
		}
		p.appDisabled = savedApp
		return ListLiteralExpr{Elements: elements}, nil

	default:
		return nil, fmt.Errorf("unexpected token in expression: %s", tok.Text)
	}
}

// bracketHasTopLevelPipe reports whether a '|' appears at the current bracket's top
// level before its matching ']', i.e. whether the '[' just consumed opens a list
// comprehension (`[e | …]`) rather than a plain list literal. It scans without
// advancing the parser, tracking nested bracket and paren depth so a pipe inside a
// nested collection or group does not count.
func (p *InfixParser) bracketHasTopLevelPipe() bool {
	depthB, depthP := 0, 0
	for scan := p.pos; scan < len(p.tokens); scan++ {
		switch p.tokens[scan].Type {
		case TokLBracket:
			depthB++
		case TokRBracket:
			if depthB == 0 {
				return false // reached our closing ']' with no top-level pipe
			}
			depthB--
		case TokLParen:
			depthP++
		case TokRParen:
			if depthP > 0 {
				depthP--
			}
		case TokPipe:
			if depthB == 0 && depthP == 0 {
				return true
			}
		case TokEOF:
			return false
		}
	}
	return false
}

// ParseInfixExpr is the public helper to parse and return an Expr AST.
func ParseInfixExpr(input string) (Expr, error) {
	tokens, err := TokenizeInfix(input)
	if err != nil {
		return nil, err
	}
	p := NewInfixParser(tokens)
	expr, err := p.ParseExpression()
	if err != nil {
		return nil, err
	}
	if p.peek().Type != TokEOF {
		return nil, fmt.Errorf("unexpected trailing tokens: %s", p.peek().Text)
	}
	return expr, nil
}

// EvalInfixString parses and evaluates an infix expression with variable/constant lookup.
func EvalInfixString(input string, lookup func(string) (any, bool)) (any, error) {
	expr, err := ParseInfixExpr(input)
	if err != nil {
		return nil, err
	}
	return expr.Eval(lookup)
}
