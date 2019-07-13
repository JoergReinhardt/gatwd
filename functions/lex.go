package functions

import (
	"sort"
	"strings"

	d "github.com/joergreinhardt/gatwd/data"
)

///// SYNTAX DEFINITION /////
type TyLex d.BitFlag

func (t TyLex) Type() d.Typed                 { return t }
func (t TyLex) FlagType() d.Uint8Val          { return Flag_Lex.U() }
func (t TyLex) TypeFnc() TyFnc                { return Type }
func (t TyLex) TypeNat() d.TyNat              { return d.Type }
func (t TyLex) Flag() d.BitFlag               { return d.BitFlag(t) }
func (t TyLex) Utf8() string                  { return mapUtf8[t] }
func (t TyLex) Ascii() string                 { return mapAscii[t] }
func (t TyLex) MatchUtf8(arg string) bool     { return t.Utf8() == arg }
func (t TyLex) MatchAscii(arg string) bool    { return t.Ascii() == arg }
func (t TyLex) Match(arg d.Typed) bool        { return t.Flag().Match(arg) }
func (t TyLex) TypeName() string              { return mapUtf8[t] }
func (t TyLex) Call(...Expression) Expression { return t }
func (t TyLex) FindUtf8(arg string) (TyLex, bool) {
	var lex, ok = mapUtf8Text[arg]
	return lex, ok
}
func (t TyLex) FindAscii(arg string) (TyLex, bool) {
	var lex, ok = mapAsciiText[arg]
	return lex, ok
}

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
)

var mapUtf8 = map[TyLex]string{
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
var mapUtf8Text = map[string]TyLex{
	"⊥":  Null,
	" ":  Blank,
	"  ": Tab,
	`\n`: NewLine,
	"_":  Underscore,
	"∗":  Asterisk,
	"‥.": Ellipsis,
	"-":  Substraction,
	"+":  Addition,
	"√":  SquareRoot,
	"∘":  Dot,
	"⨉":  Times,
	"⊙":  DotProduct,
	"⊗":  CrossProduct,
	"÷":  Division,
	"∞":  Infinite,
	"∧":  And,
	"∨":  Or,
	"⊻":  Xor,
	"¬":  Not,
	"＝":  Equal,
	"≠":  Unequal,
	"≪":  Lesser,
	"≫":  Greater,
	"≤":  LesserEq,
	"≥":  GreaterEq,
	"(":  LeftPar,
	")":  RightPar,
	"[":  LeftBra,
	"]":  RightBra,
	"{":  LeftCur,
	"}":  RightCur,
	"<":  LeftLace,
	">":  RightLace,
	`'`:  SingQuote,
	`"`:  DoubQuote,
	"`":  BackTick,
	`\`:  BackSlash,
	"/":  Slash,
	"|":  Pipe,
	"∇":  Decrement,
	"∆":  Increment,
	"≡":  TripEqual,
	"→":  RightArrow,
	"←":  LeftArrow,
	"⇐":  LeftFatArrow,
	"⇒":  RightFatArrow,
	"⇔":  DoubleFatArrow,
	"»":  Sequence,
	"«":  SequenceRev,
	"∷":  DoubCol,
	"$":  Application,
	"λ":  Lambda,
	"ϝ":  Function,
	"Ф":  Polymorph,
	"Ω":  Monad,
	"Π":  Parameter,
	"∑":  Integral,
	"⊆":  SubSet,
	"∅":  EmptySet,
	`π`:  Pi,
}

var mapAscii = map[TyLex]string{
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

var mapAsciiText = map[string]TyLex{
	"":        Null,
	" ":       Blank,
	`\t`:      Tab,
	`\n`:      NewLine,
	"_":       Underscore,
	"*":       Asterisk,
	"...":     Ellipsis,
	"-":       Substraction,
	"+":       Addition,
	`\sqrt`:   SquareRoot,
	`\dot`:    Dot,
	`\mul`:    Times,
	`\dotprd`: DotProduct,
	`\crxprd`: CrossProduct,
	`\div`:    Division,
	`\inf`:    Infinite,
	`\and`:    And,
	`\or`:     Or,
	`\xor`:    Xor,
	"!-":      Not,
	"=":       Equal,
	"!=":      Unequal,
	"<<":      Lesser,
	">>":      Greater,
	"=<":      LesserEq,
	">=":      GreaterEq,
	"(":       LeftPar,
	")":       RightPar,
	"[":       LeftBra,
	"]":       RightBra,
	"{":       LeftCur,
	"}":       RightCur,
	"<":       LeftLace,
	">":       RightLace,
	`'`:       SingQuote,
	`"`:       DoubQuote,
	"`":       BackTick,
	`\`:       BackSlash,
	`/`:       Slash,
	"|":       Pipe,
	"--":      Decrement,
	"++":      Increment,
	"===":     TripEqual,
	"->":      RightArrow,
	"<-":      LeftArrow,
	"<=":      LeftFatArrow,
	"=>":      RightFatArrow,
	"<=>":     DoubleFatArrow,
	">>>":     Sequence,
	"<<<":     SequenceRev,
	"::":      DoubCol,
	"$":       Application,
	`\y`:      Lambda,
	`\f`:      Function,
	`\F`:      Polymorph,
	`\M`:      Monad,
	`\P`:      Parameter,
	`\integ`:  Integral,
	`\subset`: SubSet,
	`\empty`:  EmptySet,
	`\pi`:     Pi,
}

var AsciiKeysSortedByLength = func() [][]rune {
	var runes = [][]rune{}
	for _, key := range mapAscii {
		runes = append(runes, []rune(key))
	}
	sort.Sort(keyLengthSorter(runes))
	return runes
}()

type keyLengthSorter [][]rune

func (k keyLengthSorter) Len() int           { return len(k) }
func (k keyLengthSorter) Less(i, j int) bool { return len(k[i]) <= len(k[j]) }
func (k keyLengthSorter) Swap(i, j int)      { k[i], k[j] = k[j], k[i] }

type TyKeyWord d.BitFlag

func (t TyKeyWord) Type() d.Typed                 { return t }
func (t TyKeyWord) FlagType() d.Uint8Val          { return Flag_KeyWord.U() }
func (t TyKeyWord) TypeFnc() TyFnc                { return Type }
func (t TyKeyWord) TypeNat() d.TyNat              { return d.Type }
func (t TyKeyWord) Flag() d.BitFlag               { return d.BitFlag(t) }
func (t TyKeyWord) KeyWord() string               { return mapKeyWords[t] }
func (t TyKeyWord) MatchKeyWord(arg string) bool  { return t.KeyWord() == arg }
func (t TyKeyWord) Match(arg d.Typed) bool        { return t == arg }
func (t TyKeyWord) TypeName() string              { return mapKeyWords[t] }
func (t TyKeyWord) Call(...Expression) Expression { return t }
func (t TyKeyWord) Find(arg string) (TyKeyWord, bool) {
	var kw, ok = mapKeyWordsText[arg]
	return kw, ok
}

//go:generate stringer -type=TyKeyWord
const (
	Word_Do TyKeyWord = 0 + iota
	Word_In
	Word_Of
	Word_Con
	Word_Let
	Word_If
	Word_Then
	Word_Else
	Word_Case
	Word_Where
	Word_Data
	Word_Type
	Word_Mutable
	Word_Otherwise
)

var mapKeyWords = map[TyKeyWord]string{
	Word_Do:        "do",
	Word_In:        "in",
	Word_Of:        "of",
	Word_Con:       "con",
	Word_Let:       "let",
	Word_If:        "if",
	Word_Then:      "then",
	Word_Else:      "else",
	Word_Case:      "case",
	Word_Where:     "where",
	Word_Data:      "data",
	Word_Type:      "type",
	Word_Mutable:   "mutable",
	Word_Otherwise: "otherwise",
}
var mapKeyWordsText = map[string]TyKeyWord{
	"do":        Word_Do,
	"in":        Word_In,
	"of":        Word_Of,
	"con":       Word_Con,
	"let":       Word_Let,
	"if":        Word_If,
	"then":      Word_Then,
	"else":      Word_Else,
	"case":      Word_Case,
	"where":     Word_Where,
	"data":      Word_Data,
	"type":      Word_Type,
	"mutable":   Word_Mutable,
	"otherwise": Word_Otherwise,
}

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

var Punktation = []string{".", ",", ";", ":", "?", "!"}
var PunktationString = strings.Join(Punktation, "")

type asciiSorter []string

func (s asciiSorter) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s asciiSorter) Less(i, j int) bool { return len(s[i]) > len(s[j]) }
func (s asciiSorter) Len() int           { return len(s) }
func (s asciiSorter) Sort()              { sort.Sort(s) }

// item is a bitflag of course
type Item interface {
	d.Native
	Syntax() string
}
