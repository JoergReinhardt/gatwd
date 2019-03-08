package lex

import (
	"sort"
	"strings"

	d "github.com/joergreinhardt/gatwd/data"
	"github.com/olekukonko/tablewriter"
)

///// SYNTAX DEFINITION /////
type SyntaxItemFlag d.BitFlag

func (t SyntaxItemFlag) Type() SyntaxItemFlag      { return t }
func (t SyntaxItemFlag) Eval(...d.Native) d.Native { return t }
func (t SyntaxItemFlag) TypeNat() d.TyNative       { return d.Flag }
func (t SyntaxItemFlag) Syntax() string            { return MapItemString[t] }
func (t SyntaxItemFlag) StringAlt() string         { return MapUtfAscii[t.Syntax()] }

// all syntax items represented as string
var AllSyntax = func() string {
	str := &strings.Builder{}
	tab := tablewriter.NewWriter(str)
	for asc, utf := range MapAsciiUtf {
		if asc == `\n` {
			asc = `⏎`
			utf = asc
		}
		var is = MapStringItem[utf].String()
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
}()

// slice of all syntax items in there int constant form
var AllItems = func() []SyntaxItemFlag {
	var tt = []SyntaxItemFlag{}
	var i uint
	var t SyntaxItemFlag = 0
	for i < 63 {
		t = 1 << i
		i = i + 1
		tt = append(tt, SyntaxItemFlag(t))
	}
	return tt
}()

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
	Ellipsis
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

var MapItemString = map[SyntaxItemFlag]string{
	None:  "⊥",
	Blank: " ",
	Tab: "	",
	NewLine:      "\n",
	Underscore:   "_",
	SquareRoot:   "√",
	Asterisk:     "∗",
	Fullstop:     ".",
	Ellipsis:     "‥.",
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
var MapStringItem = func() map[string]SyntaxItemFlag {
	var m = make(map[string]SyntaxItemFlag, len(MapItemString))
	for item, str := range MapItemString {
		m[str] = item
	}
	return m
}()
var Utf8String = func() string {
	var str string
	for _, val := range MapItemString {
		str = str + val
	}
	return str
}()

var MapUtfAscii = map[string]string{
	"⊥": "",
	"	": `\t`,
	"\n": "\n",
	"∗":  "*",
	"‥.": "...",
	"+":  "+",
	"∘":  `\do`,
	"⨉":  `\pr`,
	"⊙":  `\dp`,
	"⊗":  `\cp`,
	"÷":  `\di`,
	"∞":  `\in`,
	"∨":  `\or`,
	"⊻":  `\xo`,
	"∧":  `\an`,
	"≪":  "<<",
	"≫":  ">>",
	"≤":  "=<",
	"≥":  ">=",
	"¬":  "!",
	"≠":  "!=",
	"∇":  "--",
	"∆":  "++",
	"⇔":  "==",
	"≡":  "===",
	"→":  "->",
	"←":  "<-",
	"⇐":  "<=",
	"⇒":  "=>",
	"∷":  "::",
	"λ":  `\y`,
	`ϝ`:  `\f`,
	`Ф`:  `\F`,
	`Ω`:  `\M`,
	`Π`:  `\P`,
	"»":  ">>>",
	"«":  "<<<",
	`π`:  `\p`,
	"∑":  `\E`,
	"∈":  `\is`,
	"∅":  `\em`,
	"η":  `\et`,
	"ε":  `\ep`,
}
var AllSyntaxRunes = func() map[rune]struct{} {
	var m = map[rune]struct{}{}
	for utf, ascii := range MapUtfAscii {
		for _, r := range []rune(utf) {
			m[r] = struct{}{}
		}
		for _, r := range []rune(ascii) {
			m[r] = struct{}{}
		}
	}
	return m
}()
var AsciiFirstCharsString = func() string {
	var str string
	for _, val := range MapUtfAscii {
		if len(val) > 0 {
			str = str + string([]rune(val)[0])
		}
	}
	return str
}()

var AsciiKeysSortedByLength = func() [][]rune {
	var runes = [][]rune{}
	for _, key := range MapUtfAscii {
		runes = append(runes, []rune(key))
	}
	sort.Sort(keyLengthSorter(runes))
	return runes
}()

type keyLengthSorter [][]rune

func (k keyLengthSorter) Len() int           { return len(k) }
func (k keyLengthSorter) Less(i, j int) bool { return len(k[i]) <= len(k[j]) }
func (k keyLengthSorter) Swap(i, j int)      { k[i], k[j] = k[j], k[i] }

var MapAsciiUtf = func() map[string]string {
	var m = make(map[string]string, len(MapUtfAscii))
	for utf, asc := range MapUtfAscii {
		m[asc] = utf
	}
	return m
}()
var MapAsciiItem = func() map[string]SyntaxItemFlag {
	var m = make(map[string]SyntaxItemFlag, len(MapStringItem))
	for utf, asc := range MapUtfAscii {
		if item, ok := MapStringItem[utf]; ok {
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
var KeyWordString = strings.Join(Keywords, "")

var Digits = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
var DigitString = strings.Join(Digits, "")

var Letters = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k",
	"l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z", "ä",
	"ö", "ü", "ß"}
var LetterString = strings.Join(Letters, "")

var Capitals = []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K",
	"L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z", "Ä",
	"Ö", "Ü"}
var CapitalString = strings.Join(Capitals, "")

type asciiSorter []string

func (s asciiSorter) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s asciiSorter) Less(i, j int) bool { return len(s[i]) > len(s[j]) }
func (s asciiSorter) Len() int           { return len(s) }
func (s asciiSorter) Sort()              { sort.Sort(s) }

// returns a slice of strings sorted by length, each am ascii alternative
// syntax matching a bit of defined syntax
var Ascii = func() []string {
	var str = asciiSorter{}
	for u, key := range MapUtfAscii {
		if u != "⊥" {
			str = append(str, key)
		}
	}
	str.Sort()
	return str
}()
var AsciiString = strings.Join(Ascii, "")

var UniRunes = func() []rune {
	var runes = []rune{}
	for _, str := range UniChars {
		runes = append(runes, []rune(str)[0])
	}
	return runes
}()
var UniChars = func() []string {
	var str = []string{}
	for item, s := range MapItemString {
		if item != None {
			str = append(str, s)
		}
	}
	return str
}()
var UniCharString = strings.Join(UniChars, "")

// matches longest possible string
func MatchUtf8(str string) (Item, bool) {
	if item, ok := MapStringItem[str]; ok {
		return item, ok
	}
	return nil, false
}
func Match(str string) bool {
	if _, ok := MapAsciiItem[str]; ok {
		return ok
	}
	return false
}
func GetUtf8Item(str string) Item {
	if item, ok := MapStringItem[str]; ok {
		return item
	}
	return nil
}
func GetAsciiItem(str string) Item {
	if item, ok := MapAsciiItem[str]; ok {
		return item
	}
	return nil
}
func MatchItem(str string) (Item, bool) {
	if item, ok := MapAsciiItem[str]; ok {
		return item, true
	}
	return nil, false
}

// convert item string representation from editable to pretty
func AsciiToUnicode(ascii string) string {
	return MapAsciiUtf[ascii]
}

// convert item string representation from pretty to editable
func UnicodeToASCII(tos ...string) string {
	var sto string
	for _, s := range tos {
		sto = sto + MapUtfAscii[s]
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

// STRING REPLACER & REPLACEMENT LISTS
func NewUnicodeReplacer() *strings.Replacer {
	return strings.NewReplacer(UnicodeReplacementList()...)
}

func UnicodeReplacementList() []string {
	var ucrl = []string{}
	for _, unc := range UniChars {
		ucrl = append(ucrl, unc)
		ucrl = append(ucrl, UnicodeToASCII(unc))
	}
	return ucrl
}

func NewAsciiReplacer() *strings.Replacer {
	return strings.NewReplacer(AsciiReplacementList()...)
}

func AsciiReplacementList() []string {
	var acrl = []string{}
	for _, dig := range Ascii {
		acrl = append(acrl, dig)
		acrl = append(acrl, AsciiToUnicode(dig))
	}
	return acrl
}

func ContainsUtf(str string) bool {
	return strings.ContainsAny(str, strings.Join(UniChars, ""))
}
func ContainsAscii(str string) bool {
	return strings.ContainsAny(str, strings.Join(Ascii, ""))
}
func ContainsDigit(str string) bool {
	return strings.ContainsAny(str, strings.Join(Digits, ""))
}
func ContainsKeyword(str string) bool {
	for _, keyword := range Keywords {
		strings.ContainsAny(str, keyword)
	}
	return false
}
