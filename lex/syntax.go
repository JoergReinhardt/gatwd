package lex

import (
	"strings"

	d "github.com/JoergReinhardt/godeep/data"
	"github.com/olekukonko/tablewriter"
)

///// SYNTAX DEFINITION /////
type SyntaxItemFlag d.BitFlag

func (t SyntaxItemFlag) Type() SyntaxItemFlag { return t }
func (t SyntaxItemFlag) Flag() d.BitFlag      { return d.BitFlag(t) }
func (t SyntaxItemFlag) Syntax() string       { return syntax[t] }
func (t SyntaxItemFlag) StringAlt() string    { return matchSyntax[syntax[SyntaxItemFlag(t.Flag())]] }

// all syntax items represented as string
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

// slice of all syntax items in there int constant form
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
	None  SyntaxItemFlag = 0
	Error SyntaxItemFlag = 1
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
	BackSlash
	Lambda
	Function
	Polymorph
	Number
	Text
)

var keywords = []d.StrVal{
	d.StrVal("in"),
	d.StrVal("con"),
	d.StrVal("let"),
	d.StrVal("mutable"),
	d.StrVal("where"),
	d.StrVal("otherwise"),
	d.StrVal("if"),
	d.StrVal("then"),
	d.StrVal("else"),
	d.StrVal("case"), d.StrVal("of"), d.StrVal("data"),
	d.StrVal("type"),
}

var matchSyntax = map[string]string{
	"":   "",
	"⊥":  "_|_",
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
	"‗":  "==",
	"≡":  "===",
	"→":  "->",
	"←":  "<-",
	"⇐":  "<=",
	"⇒":  "=>",
	"∷":  "::",
	`'`:  `'`,
	`"`:  `"`,
	`\`:  `\`,
	"λ":  `\y`,
	`ϝ`:  `\f`,
	`Ф`:  `\F`,
}

var syntax = map[SyntaxItemFlag]string{
	None:       "",
	Error:      "⊥",
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
	DoubEqual:  "‗",
	TripEqual:  "≡",
	RightArrow: "→",
	LeftArrow:  "←",
	FatLArrow:  "⇐",
	FatRArrow:  "⇒",
	DoubCol:    "∷",
	Sing_quote: `'`,
	Doub_quote: `"`,
	BackSlash:  `\`,
	Lambda:     "λ",
	Function:   "ϝ",
	Polymorph:  "Ф",
}

var match = map[string]SyntaxItemFlag{
	"":    None,
	"_|_": Error,
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
	`\`:   BackSlash,
	`\y`:  Lambda,
	`\f`:  Function,
	`\F`:  Polymorph,
}

// matches longest possible string
func MatchUtf8(str string) (Item, bool) {
	if item, ok := match[matchSyntax[str]]; ok {
		return SyntaxItemFlag(item), ok
	}
	return nil, false
}
func Match(str string) (Item, bool) {
	if item, ok := match[str]; ok {
		return SyntaxItemFlag(item), ok
	}
	return nil, false
}

// convert item string representation from editable to pretty
func ASCIIToUtf8(tos ...string) []SyntaxItemFlag {
	var ti = []SyntaxItemFlag{}
	for _, s := range tos {
		ti = append(ti, match[s])
	}
	return ti
}

// convert item string representation from pretty to editable
func Utf8ToASCII(tos ...string) string {
	var sto string
	for _, s := range tos {
		sto = sto + matchSyntax[s]
	}
	return sto
}

// item is a bitflag of course
type Item interface {
	Flag() d.BitFlag
	Type() SyntaxItemFlag
	String() string
	Syntax() string
}

type TextItem struct {
	SyntaxItemFlag
	Text string
}

func (t TextItem) Type() SyntaxItemFlag { return Text }

// pretty utf-8 version of syntax item
func (t TextItem) String() string { return t.Text }
func (t TextItem) Syntax() string { return Text.Syntax() }

// provides an alternative string representation that can be edited without
// having to produce utf-8 digraphs
func (t TextItem) StringAlt() string { return t.String() }
func (t TextItem) Flag() d.BitFlag   { return d.Flag.Flag() }
