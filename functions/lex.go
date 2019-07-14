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
func (t TyLex) TypeName() string              { return mapUtf8[t] }
func (t TyLex) Call(...Expression) Expression { return t }
func (t TyLex) Match(arg d.Typed) bool        { return t.Flag().Match(arg) }
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
	Lex_Null  TyLex = 0
	Lex_Blank TyLex = 1
	Lex_Tab   TyLex = 1 << iota
	Lex_NewLine
	Lex_Underscore
	Lex_Asterisk
	Lex_Fullstop
	Lex_Ellipsis
	Lex_Substraction
	Lex_Addition
	Lex_SquareRoot
	Lex_Dot
	Lex_Times
	Lex_DotProduct
	Lex_CrossProduct
	Lex_Division
	Lex_Infinite
	Lex_And
	Lex_Or
	Lex_Xor
	Lex_Equal
	Lex_Unequal
	Lex_Lesser
	Lex_Greater
	Lex_LesserEq
	Lex_GreaterEq
	Lex_LeftPar
	Lex_RightPar
	Lex_LeftBra
	Lex_RightBra
	Lex_LeftCur
	Lex_RightCur
	Lex_LeftLace
	Lex_RightLace
	Lex_SingQuote
	Lex_DoubQuote
	Lex_BackTick
	Lex_BackSlash
	Lex_Slash
	Lex_Pipe
	Lex_Not
	Lex_Decrement
	Lex_Increment
	Lex_TripEqual
	Lex_RightArrow
	Lex_LeftArrow
	Lex_LeftFatArrow
	Lex_RightFatArrow
	Lex_DoubleFatArrow
	Lex_Sequence
	Lex_SequenceRev
	Lex_DoubCol
	Lex_Application
	Lex_Lambda
	Lex_Function
	Lex_Polymorph
	Lex_Monad
	Lex_Parameter
	Lex_Integral
	Lex_SubSet
	Lex_EmptySet
	Lex_Pi
)

var mapUtf8 = map[TyLex]string{
	Lex_Null:  "⊥",
	Lex_Blank: " ",
	Lex_Tab: "	",
	Lex_NewLine:        `\n`,
	Lex_Underscore:     "_",
	Lex_Asterisk:       "∗",
	Lex_Ellipsis:       "‥.",
	Lex_Substraction:   "-",
	Lex_Addition:       "+",
	Lex_SquareRoot:     "√",
	Lex_Dot:            "∘",
	Lex_Times:          "⨉",
	Lex_DotProduct:     "⊙",
	Lex_CrossProduct:   "⊗",
	Lex_Division:       "÷",
	Lex_Infinite:       "∞",
	Lex_And:            "∧",
	Lex_Or:             "∨",
	Lex_Xor:            "⊻",
	Lex_Not:            "¬",
	Lex_Equal:          "＝",
	Lex_Unequal:        "≠",
	Lex_Lesser:         "≪",
	Lex_Greater:        "≫",
	Lex_LesserEq:       "≤",
	Lex_GreaterEq:      "≥",
	Lex_LeftPar:        "(",
	Lex_RightPar:       ")",
	Lex_LeftBra:        "[",
	Lex_RightBra:       "]",
	Lex_LeftCur:        "{",
	Lex_RightCur:       "}",
	Lex_LeftLace:       "<",
	Lex_RightLace:      ">",
	Lex_SingQuote:      `'`,
	Lex_DoubQuote:      `"`,
	Lex_BackTick:       "`",
	Lex_BackSlash:      `\`,
	Lex_Slash:          "/",
	Lex_Pipe:           "|",
	Lex_Decrement:      "∇",
	Lex_Increment:      "∆",
	Lex_TripEqual:      "≡",
	Lex_RightArrow:     "→",
	Lex_LeftArrow:      "←",
	Lex_LeftFatArrow:   "⇐",
	Lex_RightFatArrow:  "⇒",
	Lex_DoubleFatArrow: "⇔",
	Lex_Sequence:       "»",
	Lex_SequenceRev:    "«",
	Lex_DoubCol:        "∷",
	Lex_Application:    "$",
	Lex_Lambda:         "λ",
	Lex_Function:       "ϝ",
	Lex_Polymorph:      "Ф",
	Lex_Monad:          "Ω",
	Lex_Parameter:      "Π",
	Lex_Integral:       "∑",
	Lex_SubSet:         "⊆",
	Lex_EmptySet:       "∅",
	Lex_Pi:             `π`,
}
var mapUtf8Text = map[string]TyLex{
	"⊥":  Lex_Null,
	" ":  Lex_Blank,
	"  ": Lex_Tab,
	`\n`: Lex_NewLine,
	"_":  Lex_Underscore,
	"∗":  Lex_Asterisk,
	"‥.": Lex_Ellipsis,
	"-":  Lex_Substraction,
	"+":  Lex_Addition,
	"√":  Lex_SquareRoot,
	"∘":  Lex_Dot,
	"⨉":  Lex_Times,
	"⊙":  Lex_DotProduct,
	"⊗":  Lex_CrossProduct,
	"÷":  Lex_Division,
	"∞":  Lex_Infinite,
	"∧":  Lex_And,
	"∨":  Lex_Or,
	"⊻":  Lex_Xor,
	"¬":  Lex_Not,
	"＝":  Lex_Equal,
	"≠":  Lex_Unequal,
	"≪":  Lex_Lesser,
	"≫":  Lex_Greater,
	"≤":  Lex_LesserEq,
	"≥":  Lex_GreaterEq,
	"(":  Lex_LeftPar,
	")":  Lex_RightPar,
	"[":  Lex_LeftBra,
	"]":  Lex_RightBra,
	"{":  Lex_LeftCur,
	"}":  Lex_RightCur,
	"<":  Lex_LeftLace,
	">":  Lex_RightLace,
	`'`:  Lex_SingQuote,
	`"`:  Lex_DoubQuote,
	"`":  Lex_BackTick,
	`\`:  Lex_BackSlash,
	"/":  Lex_Slash,
	"|":  Lex_Pipe,
	"∇":  Lex_Decrement,
	"∆":  Lex_Increment,
	"≡":  Lex_TripEqual,
	"→":  Lex_RightArrow,
	"←":  Lex_LeftArrow,
	"⇐":  Lex_LeftFatArrow,
	"⇒":  Lex_RightFatArrow,
	"⇔":  Lex_DoubleFatArrow,
	"»":  Lex_Sequence,
	"«":  Lex_SequenceRev,
	"∷":  Lex_DoubCol,
	"$":  Lex_Application,
	"λ":  Lex_Lambda,
	"ϝ":  Lex_Function,
	"Ф":  Lex_Polymorph,
	"Ω":  Lex_Monad,
	"Π":  Lex_Parameter,
	"∑":  Lex_Integral,
	"⊆":  Lex_SubSet,
	"∅":  Lex_EmptySet,
	`π`:  Lex_Pi,
}

var mapAscii = map[TyLex]string{
	Lex_Null:           "",
	Lex_Blank:          " ",
	Lex_Tab:            `\t`,
	Lex_NewLine:        `\n`,
	Lex_Underscore:     "_",
	Lex_Asterisk:       "*",
	Lex_Ellipsis:       "...",
	Lex_Substraction:   "-",
	Lex_Addition:       "+",
	Lex_SquareRoot:     `\sqrt`,
	Lex_Dot:            `\dot`,
	Lex_Times:          `\mul`,
	Lex_DotProduct:     `\dotprd`,
	Lex_CrossProduct:   `\crxprd`,
	Lex_Division:       `\div`,
	Lex_Infinite:       `\inf`,
	Lex_And:            `\and`,
	Lex_Or:             `\or`,
	Lex_Xor:            `\xor`,
	Lex_Not:            "!-",
	Lex_Equal:          "=",
	Lex_Unequal:        "!=",
	Lex_Lesser:         "<<",
	Lex_Greater:        ">>",
	Lex_LesserEq:       "=<",
	Lex_GreaterEq:      ">=",
	Lex_LeftPar:        "(",
	Lex_RightPar:       ")",
	Lex_LeftBra:        "[",
	Lex_RightBra:       "]",
	Lex_LeftCur:        "{",
	Lex_RightCur:       "}",
	Lex_LeftLace:       "<",
	Lex_RightLace:      ">",
	Lex_SingQuote:      `'`,
	Lex_DoubQuote:      `"`,
	Lex_BackTick:       "`",
	Lex_BackSlash:      `\`,
	Lex_Slash:          `/`,
	Lex_Pipe:           "|",
	Lex_Decrement:      "--",
	Lex_Increment:      "++",
	Lex_TripEqual:      "===",
	Lex_RightArrow:     "->",
	Lex_LeftArrow:      "<-",
	Lex_LeftFatArrow:   "<=",
	Lex_RightFatArrow:  "=>",
	Lex_DoubleFatArrow: "<=>",
	Lex_Sequence:       ">>>",
	Lex_SequenceRev:    "<<<",
	Lex_DoubCol:        "::",
	Lex_Application:    "$",
	Lex_Lambda:         `\y`,
	Lex_Function:       `\f`,
	Lex_Polymorph:      `\F`,
	Lex_Monad:          `\M`,
	Lex_Parameter:      `\P`,
	Lex_Integral:       `\integ`,
	Lex_SubSet:         `\subset`,
	Lex_EmptySet:       `\empty`,
	Lex_Pi:             `\pi`,
}

var mapAsciiText = map[string]TyLex{
	"":        Lex_Null,
	" ":       Lex_Blank,
	`\t`:      Lex_Tab,
	`\n`:      Lex_NewLine,
	"_":       Lex_Underscore,
	"*":       Lex_Asterisk,
	"...":     Lex_Ellipsis,
	"-":       Lex_Substraction,
	"+":       Lex_Addition,
	`\sqrt`:   Lex_SquareRoot,
	`\dot`:    Lex_Dot,
	`\mul`:    Lex_Times,
	`\dotprd`: Lex_DotProduct,
	`\crxprd`: Lex_CrossProduct,
	`\div`:    Lex_Division,
	`\inf`:    Lex_Infinite,
	`\and`:    Lex_And,
	`\or`:     Lex_Or,
	`\xor`:    Lex_Xor,
	"!-":      Lex_Not,
	"=":       Lex_Equal,
	"!=":      Lex_Unequal,
	"<<":      Lex_Lesser,
	">>":      Lex_Greater,
	"=<":      Lex_LesserEq,
	">=":      Lex_GreaterEq,
	"(":       Lex_LeftPar,
	")":       Lex_RightPar,
	"[":       Lex_LeftBra,
	"]":       Lex_RightBra,
	"{":       Lex_LeftCur,
	"}":       Lex_RightCur,
	"<":       Lex_LeftLace,
	">":       Lex_RightLace,
	`'`:       Lex_SingQuote,
	`"`:       Lex_DoubQuote,
	"`":       Lex_BackTick,
	`\`:       Lex_BackSlash,
	`/`:       Lex_Slash,
	"|":       Lex_Pipe,
	"--":      Lex_Decrement,
	"++":      Lex_Increment,
	"===":     Lex_TripEqual,
	"->":      Lex_RightArrow,
	"<-":      Lex_LeftArrow,
	"<=":      Lex_LeftFatArrow,
	"=>":      Lex_RightFatArrow,
	"<=>":     Lex_DoubleFatArrow,
	">>>":     Lex_Sequence,
	"<<<":     Lex_SequenceRev,
	"::":      Lex_DoubCol,
	"$":       Lex_Application,
	`\y`:      Lex_Lambda,
	`\f`:      Lex_Function,
	`\F`:      Lex_Polymorph,
	`\M`:      Lex_Monad,
	`\P`:      Lex_Parameter,
	`\integ`:  Lex_Integral,
	`\subset`: Lex_SubSet,
	`\empty`:  Lex_EmptySet,
	`\pi`:     Lex_Pi,
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
