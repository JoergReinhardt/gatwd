package parse

import (
	"sort"

	d "github.com/JoergReinhardt/godeep/data"
)

//// DATA TOKEN HANDLING
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
