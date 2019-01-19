/*
TOKEN GENERATION

  tokens are closures over a token data structure. the purpose of a token depends
  on the context. the enclosed data can range from a single bitflag, to kb's of
  unparsed sourcecode. tokens can recursively contain, or reference other tokens,
  to form linked lists, or graphs (in which case they also implement the
  'linked', 'node' & 'tree' interfaces). streams, and more so trees of tokens can
  express program source, parsed code in different levels of abstraction,
  typespec-, runtime information and last but not least references to executable
  golang code generated elsewhere in the program.

  tokens are implementet as data structure, to leaverage golang slices. loops and
  index operations for serialization of internal structures, whenever that seems
  oportune.
*/
package functions

import (
	"sort"

	d "github.com/JoergReinhardt/godeep/data"
	l "github.com/JoergReinhardt/godeep/lang"
)

type TokType uint16

func (t TokType) Flag() d.BitFlag { return d.Flag.Flag() }

//go:generate stringer -type TokType
const (
	Flat_Token TokType = 1 << iota
	Branch_Token
	Collection_Token
	Syntax_Token
	Symbolic_Token
	Kind_Token
	Number_Token
	Return_Token   // contains a data type-/ & value pair
	Argument_Token // like Return
	Data_Type_Token
	Data_Value_Token
)

type token struct {
	typ  TokType
	flag d.Typed
}

func (t token) Type() d.BitFlag { return t.typ.Flag() }

type dataToken struct {
	token
	d Data
}

func (t dataToken) Type() d.BitFlag { return Data_Value_Token.Flag() }

// syntax, symbol, number and data-type nodes all fit the bitflag. all other
// existing and later defined tokens, are considered data tokens and keep
// their content in the additional field
func newToken(t TokType, dat Data) Token {
	switch t {
	case Syntax_Token:
		return token{t, dat.(l.TokType)}
	case Data_Type_Token:
		return token{t, dat.(d.Type)}
	case Kind_Token:
		return token{t, dat.(Kind)}
	case Return_Token:
		return dataToken{token{t, dat.Flag()}, dat}
	case Argument_Token:
		return dataToken{token{t, dat.Flag()}, dat}
	case Data_Value_Token:
		return dataToken{token{t, dat.Flag()}, dat}
	}
	return nil
}

func (t token) Flag() d.BitFlag { return t.flag.Flag() }
func (t token) String() string {
	var str string
	switch t.typ {
	case Syntax_Token:
		str = t.flag.(l.TokType).Syntax()
	case Data_Type_Token:
		str = d.StringBitFlag(t.flag.(d.Type).Flag())
	case Kind_Token:
		str = d.StringBitFlag(t.flag.(Kind).Flag())
	default:
		str = "Don't know how to print this token"
	}
	return str
}
func (t dataToken) String() string {
	var str string
	switch t.typ {
	case Data_Value_Token:
		str = t.d.(d.Data).String()
	case Argument_Token:
		str = "Arg: " + d.Type(t.Flag()).String()
	case Return_Token:
		str = "Ret: " + d.Type(t.Flag()).String()
	}
	return str
}

// slice of tokens
type tokens []Token

// implementing the sort-/ and search interfaces
func (t tokens) Len() int           { return len(t) }
func (t tokens) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t tokens) Less(i, j int) bool { return t[i].Flag() < t[j].Flag() }
func sortTokens(t tokens) tokens {
	sort.Sort(t)
	return t
}
func (t tokens) Flag() BitFlag { return HigherOrder.Flag() }
func (t tokens) String() string {
	var str string
	for _, tok := range t {
		str = str + " " + tok.String()
	}
	return str
}

// consume the first token
func decapTokens(t tokens) (Token, []Token) {
	if len(t) > 0 {
		if len(t) > 1 {
			return t[0], t[1:]
		}
		return t[0], nil
	}
	return nil, nil
}

func sliceContainsToken(ts tokens, t Token) bool {
	return d.FlagMatch(t.Flag(), ts[sort.Search(
		len(ts),
		func(i int) bool {
			return ts[i].Flag().Uint() >= t.Flag().Uint()
		})].Flag())
}

// slice of token slices
type tokenSlice [][]Token

// implementing the sort-/ and search interfaces
func (t tokenSlice) Flag() BitFlag { return HigherOrder.Flag() }
func (t tokenSlice) String() string {
	var str string
	for _, s := range t {
		str = str + tokens(s).String() + "\n"
	}
	return str
}
func (t tokenSlice) Len() int           { return len(t) }
func (t tokenSlice) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t tokenSlice) Less(i, j int) bool { return t[i][0].Flag() < t[j][0].Flag() }
func sortTokenSlice(t tokenSlice) tokenSlice {
	sort.Sort(t)
	return t
}
func decapTokSlice(t tokenSlice) ([]Token, tokenSlice) {
	if len(t) > 0 {
		if len(t) > 1 {
			return t[0], t[1:]
		}
		return t[0], nil
	}
	return nil, nil
}

// match and filter tokens based on flags
func pickSliceByFirstToken(t tokenSlice, match token) [][]Token {
	ret := [][]Token{}
	i := sort.Search(len(t), func(i int) bool { return t[i][0].Flag().Uint() >= match.Flag().Uint() })
	var j = i
	for j < len(t) && d.FlagMatch(t[j][0].Flag(), match.Flag()) {
		ret = append(ret, t[j])
		j++
	}
	return ret
}
func sliceContainsSignature(sig []Token, matches tokenSlice) bool {
	match, matches := decapTokSlice(matches)
	if len(sig) == 0 {
		return false
	}
	if sortSlicePairByLength(sig, match) {
		return true
	}
	return sliceContainsSignature(sig, matches)
}
func sortSlicePairByLength(sig, match []Token) bool {
	if len(sig) > 0 {
		if len(sig) > len(match) {
			return compareTokenSequence(sig, match)
		}
		return compareTokenSequence(match, sig)
	}
	return false
}
func compareTokenSequence(long, short []Token) bool {
	// return when done with slice
	if len(short) == 0 {
		return true
	}
	l, s := long[0], short[0]
	// if either token type or flag value mismatches, return false
	if (s.Type() != l.Type()) || (!d.FlagMatch(l.Flag(), s.Flag())) {
		return false
	}
	// recurse over tails of slices
	return compareTokenSequence(long[1:], short[1:])
}
