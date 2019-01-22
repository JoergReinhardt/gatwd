package lex

import (
	"strings"

	d "github.com/JoergReinhardt/godeep/data"
	"github.com/olekukonko/tablewriter"
)

///// SYNTAX DEFINITION /////
type SyntaxItemFlag d.BitFlag

func (t SyntaxItemFlag) Flag() d.BitFlag { return d.BitFlag(t) }
func (t SyntaxItemFlag) Syntax() string  { return syntax[t] }

func AllSyntax() string {
	str := &strings.Builder{}
	tab := tablewriter.NewWriter(str)
	for _, t := range AllItems() {
		row := []string{
			t.String(), syntax[t], matchSyntax[syntax[t]],
		}
		tab.Append(row)
	}
	tab.Render()
	return str.String()
}
func AllItems() []SyntaxItemFlag {
	var tt = []SyntaxItemFlag{}
	var i uint
	var t SyntaxItemFlag = 0
	for i < 63 {
		t = 1 << i
		i = i + 1
		tt = append(tt, SyntaxItemFlag(t))
	}
	return tt
}

//go:generate stringer -type=SyntaxItemFlag
const (
	None  SyntaxItemFlag = 1
	Blank SyntaxItemFlag = 1 << iota
	Underscore
	Asterisk
	Dot
	Comma
	Colon
	Semicolon
	Minus
	Plus
	Or
	Xor
	And
	Equal
	Lesser
	Greater
	Lesseq
	Greaterq
	LeftPar
	RightPar
	LeftBra
	RightBra
	LeftCur
	RightCur
	Slash
	Pipe
	Not
	Unequal
	Dec
	Inc
	DoubEqual
	TripEqual
	RightArrow
	LeftArrow
	FatLArrow
	FatRArrow
	DoubCol
	Sing_quote
	Doub_quote
	BckSla
	Lambda
	HeadWord
	TailWord
)

var match = map[string]SyntaxItemFlag{
	"":    None,
	" ":   Blank,
	"_":   Underscore,
	"*":   Asterisk,
	".":   Dot,
	",":   Comma,
	":":   Colon,
	";":   Semicolon,
	"-":   Minus,
	"+":   Plus,
	"OR":  Or,
	"XOR": Xor,
	"AND": And,
	"=":   Equal,
	"<<":  Lesser,
	">>":  Greater,
	"=<":  Lesseq,
	">=":  Greaterq,
	"(":   LeftPar,
	")":   RightPar,
	"[":   LeftBra,
	"]":   RightBra,
	"{":   LeftCur,
	"}":   RightCur,
	"/":   Slash,
	"|":   Pipe,
	"!":   Not,
	"!=":  Unequal,
	"--":  Dec,
	"++":  Inc,
	"==":  DoubEqual,
	"===": TripEqual,
	"->":  RightArrow,
	"<-":  LeftArrow,
	"<=":  FatLArrow,
	"=>":  FatRArrow,
	"::":  DoubCol,
	`'`:   Sing_quote,
	`"`:   Doub_quote,
	`\`:   BckSla,
	"\\x": Lambda,
	"x":   HeadWord,
	"xs":  TailWord,
}
var keywords = map[string]string{
	"InWord":        "in",
	"ConWord":       "con",
	"LetWord":       "let",
	"MutableWord":   "mutable",
	"WhereWord":     "where",
	"OtherwiseWord": "otherwise",
	"IfWord":        "if",
	"ThenWord":      "then",
	"ElseWord":      "else",
	"CaseWord":      "case",
	"OfWord":        "of",
	"DataWord":      "data",
	"TypeWord":      "type",
	"Number":        "[Number]",
	"Letter":        "[Letter]",
	"Capital":       "[Capital]",
	"GenType":       "[letter]",
	"FuncIdent":     "[letter]*",
	"TypeIdent":     "[Capital][letter]*",
}
var matchSyntax = map[string]string{
	"⊥":  "",
	" ":  " ",
	"_":  "_",
	"∗":  "*",
	".":  ".",
	",":  ",",
	":":  ":",
	";":  ";",
	"-":  "-",
	"+":  "+",
	"∨":  "OR",
	"※":  "XOR",
	"∧":  "AND",
	"=":  "=",
	"≪":  "<<",
	"≫":  ">>",
	"≤":  "=<",
	"≥":  ">=",
	"(":  "(",
	")":  ")",
	"[":  "[",
	"]":  "]",
	"{":  "{",
	"}":  "}",
	"/":  "/",
	"¬":  "!",
	"≠":  "!=",
	"--": "--",
	"++": "++",
	"==": "==",
	"≡":  "===",
	"→":  "->",
	"←":  "<-",
	"⇐":  "<=",
	"⇒":  "=>",
	"∷":  "::",
	`'`:  `'`,
	`"`:  `"`,
	`\`:  `\`,
	"λ":  "\\x",
	"x":  "x",
	"xs": "xs",
}
var syntax = map[SyntaxItemFlag]string{
	None:       "⊥",
	Blank:      " ",
	Underscore: "_",
	Asterisk:   "∗",
	Dot:        ".",
	Comma:      ",",
	Colon:      ":",
	Semicolon:  ";",
	Minus:      "-",
	Plus:       "+",
	Or:         "∨",
	Xor:        "※",
	And:        "∧",
	Equal:      "=",
	Lesser:     "≪",
	Greater:    "≫",
	Lesseq:     "≤",
	Greaterq:   "≥",
	LeftPar:    "(",
	RightPar:   ")",
	LeftBra:    "[",
	RightBra:   "]",
	LeftCur:    "{",
	RightCur:   "}",
	Slash:      "/",
	Pipe:       "|",
	Not:        "¬",
	Unequal:    "≠",
	Dec:        "--",
	Inc:        "++",
	DoubEqual:  "==",
	TripEqual:  "≡",
	RightArrow: "→",
	LeftArrow:  "←",
	FatLArrow:  "⇐",
	FatRArrow:  "⇒",
	DoubCol:    "∷",
	Sing_quote: `'`,
	Doub_quote: `"`,
	BckSla:     `\`,
	Lambda:     "λ",
	HeadWord:   "x",
	TailWord:   "xs",
}

func MatchString(tos string) Item { return Item(match[tos]) }
func ASCIIToUtf8(tos ...string) []SyntaxItemFlag {
	var ti = []SyntaxItemFlag{}
	for _, s := range tos {
		ti = append(ti, match[s])
	}
	return ti
}
func Utf8ToASCII(tos ...string) string {
	var sto string
	for _, s := range tos {
		sto = sto + matchSyntax[s]
	}
	return sto
}

type Item d.BitFlag

func (t Item) Type() d.BitFlag   { return SyntaxItemFlag(t).Flag() }
func (t Item) String() string    { return SyntaxItemFlag(t).Syntax() }
func (t Item) StringAlt() string { return matchSyntax[syntax[SyntaxItemFlag(t)]] }
func (t Item) Flag() d.BitFlag   { return d.Flag.Flag() }
