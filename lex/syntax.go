package lex

import (
	"sort"
	"strings"

	d "github.com/JoergReinhardt/gatwd/data"
	"github.com/olekukonko/tablewriter"
)

///// SYNTAX DEFINITION /////
type SyntaxItemFlag d.BitFlag

func (t SyntaxItemFlag) Type() SyntaxItemFlag      { return t }
func (t SyntaxItemFlag) Eval(...d.Native) d.Native { return t }
func (t SyntaxItemFlag) TypeNat() d.TyNative       { return d.Flag }
func (t SyntaxItemFlag) Syntax() string            { return utfSyntax[t] }
func (t SyntaxItemFlag) StringAlt() string         { return asciiSyntax[utfSyntax[SyntaxItemFlag(t.TypeNat())]] }

// all syntax items represented as string
func AllSyntax() string {
	str := &strings.Builder{}
	tab := tablewriter.NewWriter(str)
	for _, t := range AllItems() {
		row := []string{
			t.String(), utfSyntax[t], asciiSyntax[utfSyntax[t]],
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
	Tab
	NewLine
	Underscore
	SquareRoot
	Asterisk
	Fullstop
	Comma
	Colon
	Semicolon
	Substraction
	Addition
	Dot
	Times
	DotProduct
	CrossProduct
	Division
	Infinite
	Or
	Xor
	And
	Equal
	Lesser
	Greater
	LesserEq
	GreaterEq
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
	Decrement
	Increment
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

var Keywords = []string{
	"in",
	"con",
	"let",
	"if",
	"then",
	"else",
	"case",
	"of",
	"where",
	"otherwise",
	"data",
	"type",
	"mutable",
	"dot",
	"times",
	"dotProduct",
	"crossProduct",
	"div",
	"infinite",
	"or",
	"xor",
	"and",
}
var Digits = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}

var asciiSyntax = map[string]string{
	"":  "",
	"⊥": "_|_",
	" ": " ",
	"	": `\t`,
	"": `\n`,
	"_": "_",
	"∗": "*",
	".": ".",
	",": ",",
	":": ":",
	";": ";",
	"-": "-",
	"+": "+",
	"·": "dot",
	"⨉": "times",
	"⊙": "dotProduct",
	"⊗": "crossProduct",
	"÷": "div",
	"∞": "infinite",
	"∨": "or",
	"⊻": "xor",
	"∧": "and",
	"=": "=",
	"≪": "<<",
	"≫": ">>",
	"≤": "=<",
	"≥": ">=",
	"(": "(",
	")": ")",
	"[": "[",
	"]": "]",
	"{": "{",
	"}": "}",
	"/": "/",
	"¬": "!",
	"≠": "!=",
	"∇": "--",
	"∆": "++",
	"‗": "==",
	"≡": "===",
	"→": "->",
	"←": "<-",
	"⇐": "<=",
	"⇒": "=>",
	"∷": "::",
	`'`: `'`,
	`"`: `"`,
	`\`: `\`,
	"λ": `\y`,
	`ϝ`: `\f`,
	`Ф`: `\F`,
}

var utfSyntax = map[SyntaxItemFlag]string{
	None:  "",
	Error: "⊥",
	Blank: " ",
	Tab: "	",
	NewLine:      "",
	Underscore:   "_",
	SquareRoot:   "√",
	Asterisk:     "∗",
	Fullstop:     ".",
	Comma:        ",",
	Colon:        ":",
	Semicolon:    ";",
	Substraction: "-",
	Addition:     "+",
	Dot:          "·",
	Times:        "⨉",
	DotProduct:   "⊙",
	CrossProduct: "⊗",
	Division:     "÷",
	Infinite:     "∞",
	Or:           "∨",
	Xor:          "⊻",
	And:          "∧",
	Equal:        "=",
	Lesser:       "≪",
	Greater:      "≫",
	LesserEq:     "≤",
	GreaterEq:    "≥",
	LeftPar:      "(",
	RightPar:     ")",
	LeftBra:      "[",
	RightBra:     "]",
	LeftCur:      "{",
	RightCur:     "}",
	Slash:        "/",
	Pipe:         "|",
	Not:          "¬",
	Unequal:      "≠",
	Decrement:    "∇",
	Increment:    "∆",
	DoubEqual:    "‗",
	TripEqual:    "≡",
	RightArrow:   "→",
	LeftArrow:    "←",
	FatLArrow:    "⇐",
	FatRArrow:    "⇒",
	DoubCol:      "∷",
	Sing_quote:   `'`,
	Doub_quote:   `"`,
	BackSlash:    `\`,
	Lambda:       "λ",
	Function:     "ϝ",
	Polymorph:    "Ф",
}

func UniChars() []string {
	var str = []string{}
	for _, s := range utfSyntax {
		str = append(str, s)
	}
	return str
}

var matchAscii = map[string]SyntaxItemFlag{
	"":    None,
	"_|_": Error,
	" ":   Blank,
	`\t`:  Tab,
	`\n`:  NewLine,
	"_":   Underscore,
	"*":   Asterisk,
	".":   Fullstop,
	",":   Comma,
	":":   Colon,
	";":   Semicolon,
	"-":   Substraction,
	"+":   Addition,
	"=":   Equal,
	"<<":  Lesser,
	">>":  Greater,
	"=<":  LesserEq,
	">=":  GreaterEq,
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
	"--":  Decrement,
	"++":  Increment,
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

type SortedDigraphs []string

func (s SortedDigraphs) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s SortedDigraphs) Less(i, j int) bool { return len(s[i]) > len(s[j]) }
func (s SortedDigraphs) Len() int           { return len(s) }
func (s SortedDigraphs) Sort()              { sort.Sort(s) }

// returns a slice of strings sorted by length, each am ascii alternative
// syntax matching a bit of defined syntax
func Digraphs() []string {
	var str = SortedDigraphs{}
	for key, _ := range matchAscii {
		str = append(str, key)
	}
	str.Sort()
	return str
}

// matches longest possible string
func MatchUtf8(str string) (Item, bool) {
	if item, ok := matchAscii[asciiSyntax[str]]; ok {
		return SyntaxItemFlag(item), ok
	}
	return nil, false
}
func Match(str string) bool {
	if _, ok := matchAscii[str]; ok {
		return ok
	}
	return false
}
func GetItem(str string) Item {
	if item, ok := matchAscii[str]; ok {
		return item
	}
	return nil
}
func MatchItem(str string) (Item, bool) {
	if item, ok := matchAscii[str]; ok {
		return SyntaxItemFlag(item), ok
	}
	return nil, false
}

// convert item string representation from editable to pretty
func AsciiToUnicode(ascii string) string {
	return matchAscii[ascii].Syntax()
}

// convert item string representation from pretty to editable
func UnicodeToASCII(tos ...string) string {
	var sto string
	for _, s := range tos {
		sto = sto + asciiSyntax[s]
	}
	return sto
}

// item is a bitflag of course
type Item interface {
	d.Native
	Type() SyntaxItemFlag
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
func (t TextItem) Flag() d.BitFlag   { return d.Flag.TypeNat().Flag() }
