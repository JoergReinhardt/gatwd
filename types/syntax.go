package types

import "text/scanner"

///// SYNTAX DEFINITION /////
type TokenType flag

func (t TokenType) Type() flag     { return flag(t) }
func (t TokenType) Syntax() string { return syntax[t] }

//go:generate stringer -type=TokenType
const (
	tok_none  TokenType = 1
	tok_blank TokenType = 1 << iota
	tok_underscore
	tok_asterisk
	tok_dot
	tok_comma
	tok_colon
	tok_semicolon
	tok_minus
	tok_plus
	tok_or
	tok_xor
	tok_and
	tok_equal
	tok_lesser
	tok_greater
	tok_leftPar
	tok_rightPar
	tok_leftBra
	tok_rightBra
	tok_leftCur
	tok_rightCur
	tok_slash
	tok_not
	tok_dec
	tok_inc
	tok_doubEqual
	tok_rightArrow
	tok_leftArrow
	tok_fatLArrow
	tok_fatRArrow
	tok_doubCol
	tok_sing_quote
	tok_doub_quote
	tok_bckSla
	tok_lambda
	tok_number
	tok_letter
	tok_capital
	tok_genType
	tok_headWord
	tok_tailWord
	tok_inWord
	tok_conWord
	tok_letWord
	tok_whereWord
	tok_otherwiseWord
	tok_ifWord
	tok_thenWord
	tok_elseWord
	tok_caseWord
	tok_ofWord
	tok_dataWord
	tok_typeWord
	tok_typeIdent
	tok_funcIdent
)

var syntax = map[TokenType]string{
	tok_none:          "",
	tok_blank:         " ",
	tok_underscore:    "_",
	tok_asterisk:      "*",
	tok_dot:           ".",
	tok_comma:         ",",
	tok_colon:         ":",
	tok_semicolon:     ";",
	tok_minus:         "-",
	tok_plus:          "+",
	tok_or:            "|",
	tok_xor:           "^",
	tok_and:           "&",
	tok_equal:         "=",
	tok_lesser:        "<",
	tok_greater:       ">",
	tok_leftPar:       "(",
	tok_rightPar:      ")",
	tok_leftBra:       "[",
	tok_rightBra:      "]",
	tok_leftCur:       "{",
	tok_rightCur:      "}",
	tok_slash:         "/",
	tok_not:           "&^",
	tok_dec:           "--",
	tok_inc:           "++",
	tok_doubEqual:     "==",
	tok_rightArrow:    "->",
	tok_leftArrow:     "<-",
	tok_fatLArrow:     "<=",
	tok_fatRArrow:     "=>",
	tok_doubCol:       "::",
	tok_sing_quote:    `'`,
	tok_doub_quote:    `"`,
	tok_bckSla:        `\`,
	tok_lambda:        `\x`,
	tok_number:        "[0-9]",
	tok_letter:        "[a-z]",
	tok_capital:       "[A-Z]",
	tok_genType:       "[[a-w]|y|z]",
	tok_headWord:      "x",
	tok_tailWord:      "xs",
	tok_inWord:        "in",
	tok_conWord:       "con",
	tok_letWord:       "let",
	tok_whereWord:     "where",
	tok_otherwiseWord: "otherwise",
	tok_ifWord:        "if",
	tok_thenWord:      "then",
	tok_elseWord:      "else",
	tok_caseWord:      "case",
	tok_ofWord:        "of",
	tok_dataWord:      "data",
	tok_typeWord:      "type",
	tok_typeIdent:     "[A-z][a-z]*",
	tok_funcIdent:     "([a-w|y|z][a-z])|(x[a-r|t-z])",
}

type item struct {
	typ  rune
	text string
	pos  scanner.Position
}

func newItem(typ rune, text string, pos scanner.Position) item {
	return item{typ, text, pos}
}
func (t item) Text() string           { return t.text }
func (t item) ItemType() rune         { return t.typ }
func (t item) ItemTypeString() string { return scanner.TokenString(t.typ) }
