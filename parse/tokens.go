/*
TOKEN GENERATION
*/
package parse

import (
	"sort"

	d "github.com/JoergReinhardt/godeep/data"
	l "github.com/JoergReinhardt/godeep/lex"
)

type TokType uint16

func (t TokType) Flag() d.BitFlag { return d.Flag.Flag() }

//go:generate stringer -type TokType
const (
	Syntax_Token TokType = 1 << iota
	Kind_Token
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
	d d.Data
}

func (t dataToken) Type() d.BitFlag { return Data_Value_Token.Flag() }

func newToken(t TokType, dat d.Data) Token {
	switch t {
	case Syntax_Token:
		return token{t, dat.(l.SyntaxItemFlag)}
	case Data_Type_Token:
		return token{t, dat.(d.Type)}
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
func (t tokens) Flag() d.BitFlag { return d.Flag.Flag() }

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
func (t tokenSlice) Flag() d.BitFlag    { return d.Flag.Flag() }
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
	i := sort.Search(len(t), func(i int) bool {
		return t[i][0].Flag().Uint() >= match.Flag().Uint()
	})
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

///////////////////////////////////////////////////////////////////////////////
// SIGNATURES
type signature func() (uid int, name string, signature string)

// token mangling
func newSyntaxToken(f l.SyntaxItemFlag) Token {
	return newToken(Syntax_Token, f)
}
func newSyntaxTokens(f ...l.SyntaxItemFlag) []Token {
	var t = make([]Token, 0, len(f))
	for _, flag := range f {
		t = append(t, newToken(Syntax_Token, flag))
	}
	return t
}
func newDataToken(f d.Type) Token {
	return newToken(Data_Type_Token, f)
}
func newDataTokens(f ...d.Type) []Token {
	var t = make([]Token, 0, len(f))
	for _, flag := range f {
		t = append(t, newToken(Data_Type_Token, flag))
	}
	return t
}
func tokPutAppend(last Token, tok []Token) []Token {
	return append(tok, last)
}
func tokPutUpFront(first Token, tok []Token) []Token {
	return append([]Token{first}, tok...)
}
func tokJoin(sep Token, tok []Token) []Token {
	var args = tokens{}
	for i, t := range tok {
		args = append(args, t)
		if i < len(tok)-1 {
			args = append(args, sep)
		}
	}
	return args
}
func tokEnclose(left, right Token, tok []Token) []Token {
	var args = tokens{left}
	for _, t := range tok {
		args = append(args, t)
	}
	args = append(args, right)
	return args
}
func tokEmbed(left, tok, right []Token) []Token {
	var args = left
	args = append(args, tok...)
	args = append(args, right...)
	return args
}

// concatenate typeflags with right arrows as seperators, to generate a chain
// of curryed arguments
func newToksFromArguments(f ...d.Type) []Token {
	return tokJoin(newSyntaxToken(l.LeftArrow), newDataTokens(f...))
}
func newTokFromRetVal(f d.BitFlag) Token {
	return newToken(Return_Token, f)
}
