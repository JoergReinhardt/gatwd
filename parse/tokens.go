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

type TyToken uint16

func (t TyToken) Flag() d.BitFlag { return d.Flag.Flag() }

//go:generate stringer -type TyToken
const (
	Syntax_Token TyToken = 1 << iota
	TypeHO_Token
	TypePrim_Token
	Property_Token
	Data_Value_Token
	Pair_Value_Token
	Token_Collection
	Argument_Token  // like Return
	Parameter_Token // like Return
	Tree_Node_Token
)

func NewSyntaxToken(f l.SyntaxItemFlag) Token  { return newToken(Syntax_Token, f) }
func NewDataTypeToken(f d.TyPrimitive) Token   { return newToken(TypePrim_Token, f) }
func NewKindToken(flag f.TyHigherOrder) Token  { return newToken(TypeHO_Token, flag) }
func NewArgumentToken(dat f.Argumented) Token  { return newToken(Argument_Token, dat) }
func NewParameterToken(dat f.Parametric) Token { return newToken(Parameter_Token, dat) }
func NewDataValueToken(dat d.Primary) Token    { return newToken(Data_Value_Token, dat) }
func NewPairValueToken(dat f.Paired) Token     { return newToken(Pair_Value_Token, dat) }
func NewTokenCollection(dat ...Token) Token    { return newToken(Token_Collection, tokens(dat)) }
func NewKeyValToken(key, val d.Primary) Token {
	return newToken(
		Parameter_Token,
		f.NewKeyValueParm(f.NewFromData(key), f.NewFromData(val)),
	)
}

type TokVal struct {
	tok TyToken
	d.Primary
}

func (t TokVal) TypeTok() TyToken        { return t.tok }
func (t TokVal) TypePrim() d.TyPrimitive { return d.Flag }
func (t TokVal) Type() d.BitFlag         { return t.tok.Flag() }

type dataTok struct {
	TokVal
	d.Primary
}

func (t dataTok) TypeTok() TyToken        { return t.TokVal.TypeTok() }
func (d dataTok) TypePrim() d.TyPrimitive { return d.Primary.TypePrim() }
func newToken(t TyToken, dat d.Primary) Token {
	switch t {
	case Syntax_Token:
		return TokVal{Syntax_Token, dat.(l.SyntaxItemFlag)}
	case TypePrim_Token:
		return TokVal{TypePrim_Token, dat.(d.TyPrimitive)}
	case TypeHO_Token:
		return TokVal{TypeHO_Token, dat.(f.TyHigherOrder)}
	case Argument_Token:
		return dataTok{TokVal{Argument_Token, dat.TypePrim()}, dat.(f.Argumented)}
	case Parameter_Token:
		return dataTok{TokVal{Parameter_Token, dat.TypePrim()}, dat.(f.Parametric)}
	case Data_Value_Token:
		return dataTok{TokVal{Data_Value_Token, dat.TypePrim()}, dat.(d.Primary)}
	case Pair_Value_Token:
		return dataTok{TokVal{Pair_Value_Token, dat.TypePrim()}, dat.(f.Paired)}
	case Token_Collection:
		return dataTok{TokVal{Token_Collection, dat.TypePrim()}, dat.(tokens)}
	case Tree_Node_Token:
		return dataTok{TokVal{Tree_Node_Token, dat.TypePrim()}, dat.(f.Parametric)}
	}
	return nil
}

// slice of tokens
type tokens []Token

// implementing the sort-/ and search interfaces
func (t tokens) Len() int                { return len(t) }
func (t tokens) Swap(i, j int)           { t[i], t[j] = t[j], t[i] }
func (t tokens) Less(i, j int) bool      { return t[i].TypePrim() < t[j].TypePrim() }
func (t tokens) TypePrim() d.TyPrimitive { return d.Flag }
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
	return d.FlagMatch(t.TypePrim(), ts[sort.Search(
		len(ts),
		func(i int) bool {
			return ts[i].TypePrim().Flag().Uint() >= t.TypePrim().Flag().Uint()
		})].TypePrim())
}
