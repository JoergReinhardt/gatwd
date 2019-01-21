package lang

import (
	"strings"

	d "github.com/JoergReinhardt/godeep/data"
	"github.com/olekukonko/tablewriter"
)

///// SYNTAX DEFINITION /////
type TypeItem d.BitFlag

func (t TypeItem) Type() TypeItem  { return TypeIdent }
func (t TypeItem) Flag() d.BitFlag { return d.BitFlag(t) }
func (t TypeItem) Syntax() string  { return syntax[t] }

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
func AllItems() []TypeItem {
	var tt = []TypeItem{}
	var i uint
	var t TypeItem = 0
	for i < 63 {
		t = 1 << i
		i = i + 1
		tt = append(tt, TypeItem(t))
	}
	return tt
}

//go:generate stringer -type=TypeItem
const (
	None  TypeItem = 1
	Blank TypeItem = 1 << iota
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
	GenType
	HeadWord
	TailWord
	InWord
	ConWord
	LetWord
	MutableWord
	WhereWord
	OtherwiseWord
	IfWord
	ThenWord
	ElseWord
	CaseWord
	OfWord
	DataWord
	TypeWord
	Number
	Letter
	Capital
	FuncIdent
	TypeIdent
)

var match = map[string]TypeItem{
	"":                   None,
	" ":                  Blank,
	"_":                  Underscore,
	"*":                  Asterisk,
	".":                  Dot,
	",":                  Comma,
	":":                  Colon,
	";":                  Semicolon,
	"-":                  Minus,
	"+":                  Plus,
	"OR":                 Or,
	"XOR":                Xor,
	"AND":                And,
	"=":                  Equal,
	"<<":                 Lesser,
	">>":                 Greater,
	"=<":                 Lesseq,
	">=":                 Greaterq,
	"(":                  LeftPar,
	")":                  RightPar,
	"[":                  LeftBra,
	"]":                  RightBra,
	"{":                  LeftCur,
	"}":                  RightCur,
	"/":                  Slash,
	"|":                  Pipe,
	"!":                  Not,
	"!=":                 Unequal,
	"--":                 Dec,
	"++":                 Inc,
	"==":                 DoubEqual,
	"===":                TripEqual,
	"->":                 RightArrow,
	"<-":                 LeftArrow,
	"<=":                 FatLArrow,
	"=>":                 FatRArrow,
	"::":                 DoubCol,
	`'`:                  Sing_quote,
	`"`:                  Doub_quote,
	`\`:                  BckSla,
	"\\x":                Lambda,
	"x":                  HeadWord,
	"xs":                 TailWord,
	"in":                 InWord,
	"con":                ConWord,
	"let":                LetWord,
	"mutable":            MutableWord,
	"where":              WhereWord,
	"otherwise":          OtherwiseWord,
	"if":                 IfWord,
	"then":               ThenWord,
	"else":               ElseWord,
	"case":               CaseWord,
	"of":                 OfWord,
	"data":               DataWord,
	"type":               TypeWord,
	"[Number]":           Number,
	"[Letter]":           Letter,
	"[Capital]":          Capital,
	"[letter]":           GenType,
	"[letter]*":          FuncIdent,
	"[Capital][letter]*": TypeIdent,
}
var matchSyntax = map[string]string{
	"⊥":                             "",
	" ":                             " ",
	"_":                             "_",
	"∗":                             "*",
	".":                             ".",
	",":                             ",",
	":":                             ":",
	";":                             ";",
	"-":                             "-",
	"+":                             "+",
	"∨":                             "OR",
	"※":                             "XOR",
	"∧":                             "AND",
	"=":                             "=",
	"≪":                             "<<",
	"≫":                             ">>",
	"≤":                             "=<",
	"≥":                             ">=",
	"(":                             "(",
	")":                             ")",
	"[":                             "[",
	"]":                             "]",
	"{":                             "{",
	"}":                             "}",
	"/":                             "/",
	"¬":                             "!",
	"≠":                             "!=",
	"--":                            "--",
	"++":                            "++",
	"==":                            "==",
	"≡":                             "===",
	"→":                             "->",
	"←":                             "<-",
	"⇐":                             "<=",
	"⇒":                             "=>",
	"∷":                             "::",
	`'`:                             `'`,
	`"`:                             `"`,
	`\`:                             `\`,
	"λ":                             "\\x",
	"x":                             "x",
	"xs":                            "xs",
	"in":                            "in",
	"con":                           "con",
	"let":                           "let",
	"mutable":                       "mutable",
	"where":                         "where",
	"otherwise":                     "otherwise",
	"if":                            "if",
	"then":                          "then",
	"else":                          "else",
	"case":                          "case",
	"of":                            "of",
	"data":                          "data",
	"type":                          "type",
	"[0-9]":                         "[Number]",
	"[a-z]":                         "[Letter]",
	"[A-Z]":                         "[Capital]",
	"[[a-w]|y|z]":                   "[letter]",
	"([a-w|y|z][a-z])|(x[a-r|t-z])": "[letter]*",
	"[A-z][a-z]*":                   "[Capital][letter]*",
}
var syntax = map[TypeItem]string{
	None:          "⊥",
	Blank:         " ",
	Underscore:    "_",
	Asterisk:      "∗",
	Dot:           ".",
	Comma:         ",",
	Colon:         ":",
	Semicolon:     ";",
	Minus:         "-",
	Plus:          "+",
	Or:            "∨",
	Xor:           "※",
	And:           "∧",
	Equal:         "=",
	Lesser:        "≪",
	Greater:       "≫",
	Lesseq:        "≤",
	Greaterq:      "≥",
	LeftPar:       "(",
	RightPar:      ")",
	LeftBra:       "[",
	RightBra:      "]",
	LeftCur:       "{",
	RightCur:      "}",
	Slash:         "/",
	Pipe:          "|",
	Not:           "¬",
	Unequal:       "≠",
	Dec:           "--",
	Inc:           "++",
	DoubEqual:     "==",
	TripEqual:     "≡",
	RightArrow:    "→",
	LeftArrow:     "←",
	FatLArrow:     "⇐",
	FatRArrow:     "⇒",
	DoubCol:       "∷",
	Sing_quote:    `'`,
	Doub_quote:    `"`,
	BckSla:        `\`,
	Lambda:        "λ",
	HeadWord:      "x",
	TailWord:      "xs",
	InWord:        "in",
	ConWord:       "con",
	LetWord:       "let",
	MutableWord:   "mutable",
	WhereWord:     "where",
	OtherwiseWord: "otherwise",
	IfWord:        "if",
	ThenWord:      "then",
	ElseWord:      "else",
	CaseWord:      "case",
	OfWord:        "of",
	DataWord:      "data",
	TypeWord:      "type",
	Number:        "[0-9]",
	Letter:        "[a-z]",
	Capital:       "[A-Z]",
	GenType:       "[[a-w]|y|z]",
	FuncIdent:     "([a-w|y|z][a-z])|(x[a-r|t-z])",
	TypeIdent:     "[A-z][a-z]*",
}

func ParseToken(tos ...string) string {
	var sto string
	for _, s := range tos {
		sto = sto + matchSyntax[s]
	}
	return sto
}

type item d.BitFlag

func (t item) Type() d.BitFlag   { return TypeItem(t).Flag() }
func (t item) String() string    { return TypeItem(t).Syntax() }
func (t item) StringAlt() string { return matchSyntax[syntax[TypeItem(t)]] }
func (t item) Flag() d.BitFlag   { return d.Flag.Flag() }
