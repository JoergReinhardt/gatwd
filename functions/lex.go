package functions

import (
	"sort"
	"strings"

	d "github.com/joergreinhardt/gatwd/data"
)

///// SYNTAX DEFINITION /////
type TyLex d.BitFlag

func (t TyLex) FlagType() d.Uint8Val          { return 4 }
func (t TyLex) Type() Typed                   { return t }
func (t TyLex) TypeName() string              { return t.String() }
func (t TyLex) TypeNat() d.TyNat              { return d.Type }
func (t TyLex) TypeFnc() TyFnc                { return Type }
func (t TyLex) Flag() d.BitFlag               { return d.BitFlag(t) }
func (t TyLex) Match(arg d.Typed) bool        { return t.Flag().Match(arg) }
func (t TyLex) Utf8() string                  { return MapUtf[t] }
func (t TyLex) Ascii() string                 { return MapAscii[t] }
func (t TyLex) Call(...Expression) Expression { return t }

// slice of all syntax items in there int constant form
var AllItems = func() []TyLex {
	var tt = []TyLex{}
	var i uint
	var t TyLex = 0
	for i < 63 {
		t = 1 << i
		i = i + 1
		tt = append(tt, TyLex(t))
	}
	return tt
}()

//go:generate stringer -type=TyLex
const (
	Null  TyLex = 0
	Blank TyLex = 1
	Tab   TyLex = 1 << iota
	NewLine
	Underscore
	Asterisk
	Fullstop
	Ellipsis
	Substraction
	Addition
	SquareRoot
	Dot
	Times
	DotProduct
	CrossProduct
	Division
	Infinite
	And
	Or
	Xor
	Equal
	Unequal
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
	LeftLace
	RightLace
	SingQuote
	DoubQuote
	BackTick
	BackSlash
	Slash
	Pipe
	Not
	Decrement
	Increment
	TripEqual
	RightArrow
	LeftArrow
	LeftFatArrow
	RightFatArrow
	DoubleFatArrow
	Sequence
	SequenceRev
	DoubCol
	Application
	Lambda
	Function
	Polymorph
	Monad
	Parameter
	Integral
	SubSet
	EmptySet
	Pi
	Number
	keyword
)

var MapUtf = map[TyLex]string{
	Null:  "⊥",
	Blank: " ",
	Tab: "	",
	NewLine:        `\n`,
	Underscore:     "_",
	Asterisk:       "∗",
	Ellipsis:       "‥.",
	Substraction:   "-",
	Addition:       "+",
	SquareRoot:     "√",
	Dot:            "∘",
	Times:          "⨉",
	DotProduct:     "⊙",
	CrossProduct:   "⊗",
	Division:       "÷",
	Infinite:       "∞",
	And:            "∧",
	Or:             "∨",
	Xor:            "⊻",
	Not:            "¬",
	Equal:          "＝",
	Unequal:        "≠",
	Lesser:         "≪",
	Greater:        "≫",
	LesserEq:       "≤",
	GreaterEq:      "≥",
	LeftPar:        "(",
	RightPar:       ")",
	LeftBra:        "[",
	RightBra:       "]",
	LeftCur:        "{",
	RightCur:       "}",
	LeftLace:       "<",
	RightLace:      ">",
	SingQuote:      `'`,
	DoubQuote:      `"`,
	BackTick:       "`",
	BackSlash:      `\`,
	Slash:          "/",
	Pipe:           "|",
	Decrement:      "∇",
	Increment:      "∆",
	TripEqual:      "≡",
	RightArrow:     "→",
	LeftArrow:      "←",
	LeftFatArrow:   "⇐",
	RightFatArrow:  "⇒",
	DoubleFatArrow: "⇔",
	Sequence:       "»",
	SequenceRev:    "«",
	DoubCol:        "∷",
	Application:    "$",
	Lambda:         "λ",
	Function:       "ϝ",
	Polymorph:      "Ф",
	Monad:          "Ω",
	Parameter:      "Π",
	Integral:       "∑",
	SubSet:         "⊆",
	EmptySet:       "∅",
	Pi:             `π`,
}
var MapStringItem = func() map[string]TyLex {
	var m = make(map[string]TyLex, len(MapUtf))
	for item, str := range MapUtf {
		m[str] = item
	}
	return m
}()
var Utf8String = func() string {
	var str string
	for _, val := range MapUtf {
		str = str + val
	}
	return str
}()

var MapAscii = map[TyLex]string{
	Null:           "",
	Blank:          " ",
	Tab:            `\t`,
	NewLine:        `\n`,
	Underscore:     "_",
	Asterisk:       "*",
	Ellipsis:       "...",
	Substraction:   "-",
	Addition:       "+",
	SquareRoot:     `\sqrt`,
	Dot:            `\dot`,
	Times:          `\mul`,
	DotProduct:     `\dotprd`,
	CrossProduct:   `\crxprd`,
	Division:       `\div`,
	Infinite:       `\inf`,
	And:            `\and`,
	Or:             `\or`,
	Xor:            `\xor`,
	Not:            "!-",
	Equal:          "=",
	Unequal:        "!=",
	Lesser:         "<<",
	Greater:        ">>",
	LesserEq:       "=<",
	GreaterEq:      ">=",
	LeftPar:        "(",
	RightPar:       ")",
	LeftBra:        "[",
	RightBra:       "]",
	LeftCur:        "{",
	RightCur:       "}",
	LeftLace:       "<",
	RightLace:      ">",
	SingQuote:      `'`,
	DoubQuote:      `"`,
	BackTick:       "`",
	BackSlash:      `\`,
	Slash:          `/`,
	Pipe:           "|",
	Decrement:      "--",
	Increment:      "++",
	TripEqual:      "===",
	RightArrow:     "->",
	LeftArrow:      "<-",
	LeftFatArrow:   "<=",
	RightFatArrow:  "=>",
	DoubleFatArrow: "<=>",
	Sequence:       ">>>",
	SequenceRev:    "<<<",
	DoubCol:        "::",
	Application:    "$",
	Lambda:         `\y`,
	Function:       `\f`,
	Polymorph:      `\F`,
	Monad:          `\M`,
	Parameter:      `\P`,
	Integral:       `\integ`,
	SubSet:         `\subset`,
	EmptySet:       `\empty`,
	Pi:             `\pi`,
}

var AsciiKeysSortedByLength = func() [][]rune {
	var runes = [][]rune{}
	for _, key := range MapAscii {
		runes = append(runes, []rune(key))
	}
	sort.Sort(keyLengthSorter(runes))
	return runes
}()

type keyLengthSorter [][]rune

func (k keyLengthSorter) Len() int           { return len(k) }
func (k keyLengthSorter) Less(i, j int) bool { return len(k[i]) <= len(k[j]) }
func (k keyLengthSorter) Swap(i, j int)      { k[i], k[j] = k[j], k[i] }

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
}
var KeyWordString = strings.Join(Keywords, "")

var Digits = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
var DigitString = strings.Join(Digits, "")

var letters = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k",
	"l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z", "ä",
	"ö", "ü", "ß"}
var LetterString = strings.Join(letters, "")

var Capitals = []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K",
	"L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z", "Ä",
	"Ö", "Ü"}
var CapitalString = strings.Join(Capitals, "")

var ScentenceMarks = []string{".", ",", ";", ":", "?", "!"}
var ScentenceMarkString = strings.Join(ScentenceMarks, "")

type asciiSorter []string

func (s asciiSorter) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s asciiSorter) Less(i, j int) bool { return len(s[i]) > len(s[j]) }
func (s asciiSorter) Len() int           { return len(s) }
func (s asciiSorter) Sort()              { sort.Sort(s) }

// item is a bitflag of course
type Item interface {
	d.Native
	Type() TyLex
	Syntax() string
}

type TextItem struct {
	TyLex
	Text string
}

func (t TextItem) Type() TyLex { return keyword }

// pretty utf-8 version of syntax item
func (t TextItem) String() string { return t.Text }
func (t TextItem) Syntax() string { return keyword.Utf8() }

// provides an alternative string representation that can be edited without
// having to produce utf-8 digraphs
func (t TextItem) StringAlt() string { return t.String() }
func (t TextItem) Flag() d.BitFlag   { return d.Type.TypeNat().Flag() }
