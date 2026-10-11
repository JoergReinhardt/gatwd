package parser

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/functions"
)

// TokenType represents lexical token types for infix expressions.
type TokenType int

const (
	TokEOF TokenType = iota
	TokNumber
	TokString
	TokIdent
	TokOp
	TokLParen
	TokRParen
	TokLBracket
	TokRBracket
	TokComma
	TokPipe
	TokArrow
)

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

		// Multi-character or single-character operators
		// <-, ->, ++, ==, !=, <=, >=, &&, ||, //
		if i+1 < n {
			two := string(runes[i : i+2])
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

		if r == '|' {
			tokens = append(tokens, TokenItem{Type: TokPipe, Text: "|"})
			i++
			continue
		}

		// Single-character operators: +, -, *, /, %, <, >
		if r == '+' || r == '-' || r == '*' || r == '/' || r == '%' || r == '<' || r == '>' {
			isUnary := (len(tokens) == 0 || tokens[len(tokens)-1].Type == TokOp || tokens[len(tokens)-1].Type == TokLParen || tokens[len(tokens)-1].Type == TokLBracket || tokens[len(tokens)-1].Type == TokComma || tokens[len(tokens)-1].Type == TokPipe || tokens[len(tokens)-1].Type == TokArrow)
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

// Expr represents an abstract syntax tree for infix expressions.
type Expr interface {
	Eval(lookup func(string) (any, bool)) (any, error)
}

type LiteralExpr struct {
	Value any
}

func (e LiteralExpr) Eval(lookup func(string) (any, bool)) (any, error) {
	return e.Value, nil
}

type VarExpr struct {
	Name string
}

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
	return nil, fmt.Errorf("undefined identifier: %s", e.Name)
}

type BinaryExpr struct {
	Op    functions.OpKind
	Left  Expr
	Right Expr
}

func (e BinaryExpr) Eval(lookup func(string) (any, bool)) (any, error) {
	l, err := e.Left.Eval(lookup)
	if err != nil {
		return nil, err
	}
	r, err := e.Right.Eval(lookup)
	if err != nil {
		return nil, err
	}
	return functions.EvalBinaryOp(e.Op, l, r)
}

// ListLiteralExpr represents a list literal: [1, 2, 3]
type ListLiteralExpr struct {
	Elements []Expr
}

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

// ListComprehensionExpr represents [expr | var <- source, filterExpr...]
type ListComprehensionExpr struct {
	YieldExpr Expr
	VarName   string
	Source    Expr
	Filters   []Expr
}

func (e ListComprehensionExpr) Eval(lookup func(string) (any, bool)) (any, error) {
	srcVal, err := e.Source.Eval(lookup)
	if err != nil {
		return nil, err
	}

	var items []any
	switch s := srcVal.(type) {
	case *collections.List[any]:
		for item := range s.Seq() {
			items = append(items, item)
		}
	case *collections.List[int]:
		for item := range s.Seq() {
			items = append(items, item)
		}
	case *collections.List[string]:
		for item := range s.Seq() {
			items = append(items, item)
		}
	case []any:
		items = append(items, s...)
	case []int:
		for _, x := range s {
			items = append(items, x)
		}
	case []string:
		for _, x := range s {
			items = append(items, x)
		}
	case int:
		// Range 0 .. s-1
		for i := 0; i < s; i++ {
			items = append(items, i)
		}
	default:
		return nil, fmt.Errorf("list comprehension source must be a collection, got %T", srcVal)
	}

	var collected []any
	for _, item := range items {
		// Scoped lookup binding VarName
		scopedLookup := func(k string) (any, bool) {
			if k == e.VarName {
				return item, true
			}
			if lookup != nil {
				return lookup(k)
			}
			return nil, false
		}

		passed := true
		for _, filt := range e.Filters {
			fRes, err := filt.Eval(scopedLookup)
			if err != nil {
				return nil, err
			}
			b, ok := fRes.(bool)
			if !ok || !b {
				passed = false
				break
			}
		}

		if passed {
			yVal, err := e.YieldExpr.Eval(scopedLookup)
			if err != nil {
				return nil, err
			}
			collected = append(collected, yVal)
		}
	}

	resList := collections.Nil[any]()
	for idx := len(collected) - 1; idx >= 0; idx-- {
		resList = collections.Cons(collected[idx], resList)
	}
	return resList, nil
}

// InfixParser builds an Expr using precedence climbing.
type InfixParser struct {
	tokens []TokenItem
	pos    int
}

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

func (p *InfixParser) ParseExpression() (Expr, error) {
	return p.parsePrecedence(0)
}

func (p *InfixParser) parsePrecedence(minPrec int) (Expr, error) {
	left, err := p.parsePrimary()
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

	case TokIdent:
		return VarExpr{Name: tok.Text}, nil

	case TokLParen:
		expr, err := p.parsePrecedence(0)
		if err != nil {
			return nil, err
		}
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

		firstExpr, err := p.parsePrecedence(0)
		if err != nil {
			return nil, err
		}

		// Check if this is a List Comprehension: [expr | var <- src, filter...]
		if p.peek().Type == TokPipe {
			p.next() // consume '|'
			varTok := p.next()
			if varTok.Type != TokIdent {
				return nil, fmt.Errorf("expected generator variable name after '|'")
			}
			arrowTok := p.next()
			if arrowTok.Type != TokArrow || arrowTok.Text != "<-" {
				return nil, fmt.Errorf("expected '<-' in generator, got %s", arrowTok.Text)
			}
			sourceExpr, err := p.parsePrecedence(0)
			if err != nil {
				return nil, err
			}

			var filters []Expr
			for p.peek().Type == TokComma {
				p.next() // consume ','
				filt, err := p.parsePrecedence(0)
				if err != nil {
					return nil, err
				}
				filters = append(filters, filt)
			}

			if p.next().Type != TokRBracket {
				return nil, fmt.Errorf("expected closing bracket ']' after list comprehension")
			}

			return ListComprehensionExpr{
				YieldExpr: firstExpr,
				VarName:   varTok.Text,
				Source:    sourceExpr,
				Filters:   filters,
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
		return ListLiteralExpr{Elements: elements}, nil

	default:
		return nil, fmt.Errorf("unexpected token in expression: %s", tok.Text)
	}
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
