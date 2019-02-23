/*
TOKEN GENERATION

  the token type provides a way to serialize source code to be interpreted by
  gatwds runtime, as well as data to be computed on and all the data types the
  library itself consists of. that makes all gatwd compositions serializeable,
  including runtime state. that way running processes can be frozen for later
  execution and transferred for remote execution, including their current
  runtime state and possibly the dataset that's been worked on.

  Tokens come in different types to discriminate between the different bitflags
  used for different purpose by different parts of gatwd, as well as a token
  type to contain arbitrary instances of the data type. This makes gatwd
  entirely selfcontained.

  since the type system kind of 'needs to be there', at least at it's most
  basic form, for being able to define precedence types and further language
  features, a method to compare sequences of type decoding tokens is provided.
  this will be used during initialization to parse and compare the type
  definitions of precedence types that are neither recursive nor parametrized
  and don't define further types at the right hand side of their definition.
  any pattern matching more complicated will be implemented on top of that base
  comparision and get's defined in terms of gatwd itself.
*/

package parse

import (
	"fmt"
	"sort"
	"strconv"

	d "github.com/JoergReinhardt/gatwd/data"
	f "github.com/JoergReinhardt/gatwd/functions"
	l "github.com/JoergReinhardt/gatwd/lex"
)

type TyToken uint16

func (t TyToken) Eval(...d.Native) d.Native { return t }
func (t TyToken) Flag() d.BitFlag           { return d.Flag.Flag() }
func (t TyToken) TypeHO() f.TyFnc           { return f.HigherOrder }
func (t TyToken) TypeNat() d.TyNative       { return d.Flag }

//go:generate stringer -type TyToken
const (
	Syntax_Token TyToken = 1 << iota
	TypeFnc_Token
	TypeNat_Token
	Property_Token
	Data_Value_Token
	Error_Token
	Digit_Token
	Word_Token
	Name_Token
	Keyword_Token
	Pair_Token
	Token_Collection
	Argument_Token  // like Return
	Parameter_Token // like Return
	Tree_Node_Token
)

func NewSyntaxToken(pos int, f l.Item) Token      { return newToken(Syntax_Token, pos, f) }
func NewNatTypeToken(pos int, f d.TyNative) Token { return newToken(TypeNat_Token, pos, f) }
func NewFncTypeToken(pos int, flag f.TyFnc) Token { return newToken(TypeFnc_Token, pos, flag) }
func NewArgumentToken(pos int, dat string) Token {
	return newToken(Argument_Token, pos, f.NewArgument(f.New(dat)))
}
func NewParameterToken(pos int, acc, arg string) Token {
	return newToken(Parameter_Token, pos, f.NewParameter(f.NewPairFromData(d.New(acc), d.New(arg))))
}
func NewDataValueToken(pos int, dat string) Token { return newToken(Data_Value_Token, pos, d.New(dat)) }
func NewValueToken(pos int, dat string) Token     { return newToken(Name_Token, pos, d.New(dat)) }
func NewWordToken(pos int, dat string) Token {
	return newToken(Word_Token, pos, d.New(dat))
}
func NewNameToken(pos int, dat string) Token {
	return newToken(Name_Token, pos, d.New(dat))
}
func NewKeywordToken(pos int, dat string) Token {
	return newToken(Keyword_Token, pos, d.New(dat))
}
func NewErrorToken(pos int, dat string) Token {
	return newToken(Error_Token, pos, d.New(fmt.Errorf(dat)))
}
func NewDigitToken(pos int, dat string) Token {
	i, err := strconv.Atoi(dat)
	if err != nil {
		return newToken(Error_Token, pos, d.New(err))
	}
	return newToken(Digit_Token, pos, d.IntVal(i))
}
func NewPairToken(pos int, left, right string) Token {
	return newToken(Pair_Token, pos, f.NewPairFromData(d.New(left), d.New(right)))
}
func NewTokenCollection(pos int, dat ...Token) Token {
	return newToken(Token_Collection, pos, tokens(dat))
}
func NewTreeNodeToken(pos int, dat ...Token) Token { return newToken(Tree_Node_Token, pos, tokens(dat)) }
func NewKeyValToken(pos int, key, val d.Native) Token {
	return newToken(
		Parameter_Token, pos,
		f.NewKeyValueParm(f.NewFromData(key), f.NewFromData(val)),
	)
}

type TokVal struct {
	tok TyToken
	pos int
	d.Native
}

func (t TokVal) Pos() int            { return t.pos }
func (t TokVal) Data() d.Native      { return t.Native }
func (t TokVal) TypeTok() TyToken    { return t.tok }
func (t TokVal) TypeNat() d.TyNative { return d.Flag }
func (t TokVal) Type() d.BitFlag     { return t.tok.Flag() }

type dataTok struct {
	TokVal
	d.Native
}

func (t dataTok) Data() d.Native { return t.Native }

func (t dataTok) TypeTok() TyToken    { return t.TokVal.TypeTok() }
func (d dataTok) TypeNat() d.TyNative { return d.Native.TypeNat() }
func newToken(t TyToken, pos int, dat d.Native) Token {
	switch t {
	case Syntax_Token:
		return TokVal{Syntax_Token, pos, dat.(l.SyntaxItemFlag)}
	case TypeNat_Token:
		return TokVal{TypeNat_Token, pos, dat.(d.TyNative)}
	case TypeFnc_Token:
		return TokVal{TypeFnc_Token, pos, dat.(f.TyFnc)}
	case Argument_Token:
		return dataTok{TokVal{Argument_Token, pos, dat.TypeNat()}, dat.(f.Argumented)}
	case Parameter_Token:
		return dataTok{TokVal{Parameter_Token, pos, dat.TypeNat()}, dat.(f.Parametric)}
	case Data_Value_Token:
		return dataTok{TokVal{Data_Value_Token, pos, dat.TypeNat()}, dat.(d.Native)}
	case Pair_Token:
		return dataTok{TokVal{Pair_Token, pos, dat.TypeNat()}, dat.(f.Paired)}
	case Token_Collection:
		return dataTok{TokVal{Token_Collection, pos, dat.TypeNat()}, dat.(tokens)}
	case Tree_Node_Token:
		return dataTok{TokVal{Tree_Node_Token, pos, dat.TypeNat()}, dat.(f.Parametric)}
	}
	return nil
}

// slice of tokens
type tokens []Token

// implementing the sort-/ and search interfaces
func (t tokens) Len() int                  { return len(t) }
func (t tokens) Swap(i, j int)             { t[i], t[j] = t[j], t[i] }
func (t tokens) Less(i, j int) bool        { return t[i].TypeNat() < t[j].TypeNat() }
func (t tokens) Eval(...d.Native) d.Native { return t }
func (t tokens) TypeNat() d.TyNative       { return d.Flag }
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
	return d.FlagMatch(t.TypeNat(), ts[sort.Search(
		len(ts),
		func(i int) bool {
			return ts[i].TypeNat().Flag().Uint() >= t.TypeNat().Flag().Uint()
		})].TypeNat())
}

//// DATA TOKEN MANGLING
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

//// TOKEN SLICE
//
type tokenSlice [][]Token

// implementing the sort-/ and search interfaces
func (t tokenSlice) Flag() d.BitFlag    { return d.Flag.Flag() }
func (t tokenSlice) Len() int           { return len(t) }
func (t tokenSlice) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t tokenSlice) Less(i, j int) bool { return t[i][0].TypeNat() < t[j][0].TypeNat() }
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
		return t[i][0].TypeNat().Flag().Uint() >= match.TypeNat().Flag().Uint()
	})
	var j = i
	for j < len(t) && d.FlagMatch(t[j][0].TypeNat(), match.TypeNat()) {
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
	if (s.TypeTok() != l.TypeTok()) || (!d.FlagMatch(l.TypeNat(), s.TypeNat())) {
		return false
	}
	// recurse over tails of slices
	return compareTokenSequence(long[1:], short[1:])
}
