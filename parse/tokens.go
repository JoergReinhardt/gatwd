/*
TOKEN GENERATION

  the token type provides a way to serialize source code to be interpreted by
  godeeps runtime, as well as data to be computed on and all the data types the
  library itself consists of. that makes all godeep compositions serializeable,
  including runtime state. that way running processes can be frozen for later
  execution and transferred for remote execution, including their current
  runtime state and possibly the dataset that's been worked on.

  Tokens come in different types to discriminate between the different bitflags
  used for different purpose by different parts of godeep, as well as a token
  type to contain arbitrary instances of the data type. This makes godeep
  entirely selfcontained.

  since the type system kind of 'needs to be there', at least at it's most
  basic form, for being able to define precedence types and further language
  features, a method to compare sequences of type decoding tokens is provided.
  this will be used during initialization to parse and compare the type
  definitions of precedence types that are neither recursive nor parametrized
  and don't define further types at the right hand side of their definition.
  any pattern matching more complicated will be implemented on top of that base
  comparision and get's defined in terms of godeep itself.
*/

package parse

import (
	"sort"

	d "github.com/JoergReinhardt/godeep/data"
	f "github.com/JoergReinhardt/godeep/functions"
	l "github.com/JoergReinhardt/godeep/lex"
)

type TokType uint16

func (t TokType) Flag() d.BitFlag { return d.Flag.Flag() }

//go:generate stringer -type TokType
const (
	Syntax_Token TokType = 1 << iota
	Kind_Token
	Argument_Token  // like Return
	Parameter_Token // like Return
	Data_Type_Token
	Data_Value_Token
	Pair_Value_Token
	Token_Collection
)

func NewKindToken(dat d.Data) Token            { return newToken(Kind_Token, dat) }
func NewArgumentToken(dat f.Argumented) Token  { return newToken(Argument_Token, dat) }
func NewParameterToken(dat f.Parametric) Token { return newToken(Parameter_Token, dat) }
func NewKeyValToken(key, val d.Data) Token {
	return newToken(
		Parameter_Token,
		f.NewKeyValueParm(key, val),
	)
}
func NewDataTypeToken(dat d.Typed) Token    { return newToken(Data_Type_Token, dat.Flag()) }
func NewDataValueToken(dat d.Data) Token    { return newToken(Data_Value_Token, dat) }
func NewPairValueToken(dat f.Paired) Token  { return newToken(Pair_Value_Token, dat) }
func NewTokenCollection(dat ...Token) Token { return newToken(Token_Collection, tokens(dat)) }

type TokVal struct {
	tok TokType
	d.Typed
}

func (t TokVal) TokType() TokType { return t.tok }
func (t TokVal) Flag() d.BitFlag  { return t.Typed.Flag() }

type dataTok struct {
	TokVal
	d.Data
}

func (t dataTok) TokType() TokType { return t.TokVal.TokType() }
func (d dataTok) Flag() d.BitFlag  { return d.Data.Flag() }
func newToken(t TokType, dat d.Data) Token {
	switch t {
	case Syntax_Token:
		return TokVal{t, dat}
	case Data_Type_Token:
		return TokVal{t, dat}
	case Argument_Token:
		return dataTok{TokVal{t, d.Type(dat.Flag())}, dat.(f.Argumented)}
	case Parameter_Token:
		return dataTok{TokVal{t, dat.Flag()}, dat.(f.Parametric)}
	case Data_Value_Token:
		return dataTok{TokVal{t, dat.Flag()}, dat.(d.Data)}
	case Pair_Value_Token:
		return dataTok{TokVal{t, dat.Flag()}, dat.(f.Paired)}
	case Token_Collection:
		return dataTok{TokVal{t, dat.Flag()}, dat.(tokens)}
	}
	return nil
}

// slice of tokens
type tokens []Token

// implementing the sort-/ and search interfaces
func (t tokens) Len() int           { return len(t) }
func (t tokens) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t tokens) Less(i, j int) bool { return t[i].Flag() < t[j].Flag() }
func (t tokens) Flag() d.BitFlag    { return d.Flag.Flag() }
func sortTokens(t tokens) tokens {
	sort.Sort(t)
	return t
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
func pickSliceByFirstToken(t tokenSlice, match TokVal) [][]Token {
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
	if (s.TokType() != l.TokType()) || (!d.FlagMatch(l.Flag(), s.Flag())) {
		return false
	}
	// recurse over tails of slices
	return compareTokenSequence(long[1:], short[1:])
}

///////////////////////////////////////////////////////////////////////////////
// SIGNATURES
type signature func() (uid int, name string, signature string)

// token mangling
func NewSyntaxToken(f l.SyntaxItemFlag) Token {
	return newToken(Syntax_Token, f)
}
func NewSyntaxTokens(f ...l.SyntaxItemFlag) []Token {
	var t = make([]Token, 0, len(f))
	for _, flag := range f {
		t = append(t, newToken(Syntax_Token, flag))
	}
	return t
}
func NewDataToken(f d.Type) Token {
	return newToken(Data_Type_Token, f)
}
func NewDataTokens(f ...d.Type) []Token {
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
