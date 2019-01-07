package functions

import (
	d "github.com/JoergReinhardt/godeep/data"
	l "github.com/JoergReinhardt/godeep/lang"
	"sort"
)

// TOKEN GENERATION
//
// tokens are closures over a token data structure. the purpose of a token
// depends on the context. the enclosed data can range from a single bitflag,
// to kb's of unparsed sourcecode. tokens can recursively contain, or reference
// other tokens, to form linked lists, or graphs (in which case they also
// implement the 'linked', 'node' & 'tree' interfaces). streams, and more so
// trees of tokens can express program source, parsed code in different levels
// of abstraction, typespec-, runtime information and last but not least
// references to executable golang code generated elsewhere in the program.
//
// tokens are implementet as data structure, to leaverage golang slices. loops
// and index operations for serialization of internal structures, whenever that
// seems oportune.
type TokType uint8

//go:generate stringer -type TokType
const (
	Syntax TokType = 1 << iota
	Symbol
	Number
	Return   // contains a data type-/ & value pair
	Argument // like Return
	Data_Type
	Data_Value
)

type token struct {
	typ  TokType
	flag d.BitFlag
}
type dataToken struct {
	token
	d d.Data
}

// syntax, symbol, number and data-type nodes all fit the bitflag. all other
// existing and later defined tokens, are considered data tokens and keep
// their content in the additional field
func conToken(t TokType, dat d.Data) Token {
	switch t {
	case Syntax:
		return token{t, dat.Flag()}
	case Data_Type:
		return token{t, dat.Flag()}
	case Return:
		return dataToken{token{t, dat.Flag()}, dat}
	case Argument:
		return dataToken{token{t, dat.Flag()}, dat}
	case Data_Value:
		return dataToken{token{t, dat.Flag()}, dat}
	}
	return nil
}

func (t token) Type() TokType   { return t.typ }
func (t token) Flag() d.BitFlag { return t.flag }
func (t token) String() string {
	var str string
	switch t.typ {
	case Syntax:
		str = l.Token(t.flag).Text()
	case Data_Type:
		str = d.Type(t.flag).Flag().String()
	default:
		str = "Don't know how to print this token"
	}
	return str
}
func (t dataToken) String() string {
	var str string
	switch t.typ {
	case Data_Value:
		str = t.d.String()
	case Argument:
		str = "Arg [" + d.Type(t.Flag()).String() + "] " + t.d.String()
	case Return:
		str = "Ret [" + d.Type(t.Flag()).String() + "] " + t.d.String()
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
	return d.Match(t.Flag(), ts[sort.Search(
		len(ts),
		func(i int) bool {
			return ts[i].Flag().Uint() >= t.Flag().Uint()
		})].Flag())
}

// slice of token slices
type tokenSlice [][]Token

// implementing the sort-/ and search interfaces
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
	for j < len(t) && d.Match(t[j][0].Flag(), match.Flag()) {
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
	if len(short) == 0 {
		return true
	}
	l, s := long[0], short[0]
	if !d.Match(l.Flag(), s.Flag()) {
		return false
	}
	return compareTokenSequence(long[1:], short[1:])
}
