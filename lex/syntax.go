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
func (t SyntaxItemFlag) Syntax() string            { return itemToString[t] }
func (t SyntaxItemFlag) StringAlt() string         { return utfToAscii[t.Syntax()] }

// all syntax items represented as string
func AllSyntax() string {
	str := &strings.Builder{}
	tab := tablewriter.NewWriter(str)
	for asc, utf := range asciiToUtf {
		if asc == `\n` {
			asc = `⏎`
			utf = asc
		}
		var is = stringToItem[utf].String()
		if asc == `\t` {
			asc = `␉`
			utf = asc
		}
		utf = "  " + utf + "  "
		row := []string{
			is, utf, asc,
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
	Blank SyntaxItemFlag = 1
	Tab   SyntaxItemFlag = 1 << iota
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
	DoubleEqual
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
	Monad
	Parameter
	Sequence
	SequenceRev
	Integral
	IsMember
	EmptySet
	Number
	Text
	Eta
	Epsilon
)

var itemToString = map[SyntaxItemFlag]string{
	None:  "⊥",
	Blank: " ",
	Tab: "	",
	NewLine:      "\n",
	Underscore:   "_",
	SquareRoot:   "√",
	Asterisk:     "∗",
	Fullstop:     ".",
	Comma:        ",",
	Colon:        ":",
	Semicolon:    ";",
	Substraction: "-",
	Addition:     "+",
	Dot:          "∘",
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
	DoubleEqual:  "⇔",
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
	Monad:        "Ω",
	Parameter:    "Π",
	Sequence:     "»",
	SequenceRev:  "«",
	Integral:     "∑",
	IsMember:     "∈",
	EmptySet:     "∅",
}
var stringToItem = func() map[string]SyntaxItemFlag {
	var m = make(map[string]SyntaxItemFlag, len(itemToString))
	for item, str := range itemToString {
		m[str] = item
	}
	return m
}()

var utfToAscii = map[string]string{
	"⊥": "",
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
	"∘": `\dot`,
	"⨉": `\prod`,
	"⊙": `\dProd`,
	"⊗": `\cProd`,
	"÷": `\div`,
	"∞": `\inf`,
	"∨": `\or`,
	"⊻": `\xor`,
	"∧": `\and`,
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
	"⇔": "==",
	"≡": "⇔=",
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
	`Ω`: `\M`,
	`Π`: `\P`,
	"»": "≫>",
	"«": "≪<",
	`π`: `\p`,
	"∑": `\E`,
	"∈": `\is`,
	"∅": `\emp`,
	"η": `\eta`,
	"ε": `\eps`,
}

var asciiToUtf = func() map[string]string {
	var m = make(map[string]string, len(utfToAscii))
	for utf, asc := range utfToAscii {
		m[asc] = utf
	}
	return m
}()
var asciiToItem = func() map[string]SyntaxItemFlag {
	var m = make(map[string]SyntaxItemFlag, len(stringToItem))
	for utf, asc := range utfToAscii {
		if item, ok := stringToItem[utf]; ok {
			m[asc] = item
		}
	}
	return m
}()

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

type SortedDigraphs []string

func (s SortedDigraphs) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s SortedDigraphs) Less(i, j int) bool { return len(s[i]) > len(s[j]) }
func (s SortedDigraphs) Len() int           { return len(s) }
func (s SortedDigraphs) Sort()              { sort.Sort(s) }

// returns a slice of strings sorted by length, each am ascii alternative
// syntax matching a bit of defined syntax
func Digraphs() []string {
	var str = SortedDigraphs{}
	for u, key := range utfToAscii {
		if u != "⊥" {
			str = append(str, key)
		}
	}
	str.Sort()
	return str
}

func UniRunes() []rune {
	var runes = []rune{}
	for _, str := range UniChars() {
		runes = append(runes, []rune(str)[0])
	}
	return runes
}
func UniChars() []string {
	var str = []string{}
	for item, s := range itemToString {
		if item != None {
			str = append(str, s)
		}
	}
	return str
}

// matches longest possible string
func MatchUtf8(str string) (Item, bool) {
	if item, ok := stringToItem[str]; ok {
		return item, ok
	}
	return nil, false
}
func Match(str string) bool {
	if _, ok := asciiToItem[str]; ok {
		return ok
	}
	return false
}
func GetItem(str string) Item {
	if item, ok := asciiToItem[str]; ok {
		return item
	}
	return nil
}
func MatchItem(str string) (Item, bool) {
	if item, ok := asciiToItem[str]; ok {
		return item, true
	}
	return nil, false
}

// convert item string representation from editable to pretty
func AsciiToUnicode(ascii string) string {
	return asciiToUtf[ascii]
}

// convert item string representation from pretty to editable
func UnicodeToASCII(tos ...string) string {
	var sto string
	for _, s := range tos {
		sto = sto + utfToAscii[s]
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
