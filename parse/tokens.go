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
)

type TokType uint16

func (t TokType) Flag() d.BitFlag { return d.Flag.Flag() }

//go:generate stringer -type TokType
const (
	Syntax_Token TokType = 1 << iota
	Type_Token
	Kind_Token
	Data_Type_Token
	Data_Value_Token
	Pair_Value_Token
	Token_Collection
	Argument_Token  // like Return
	Parameter_Token // like Return
	Tree_Node_Token
)

func NewKindToken(dat d.Data) Token            { return newToken(Kind_Token, dat) }
func NewTypeToken(dat d.Data) Token            { return newToken(Type_Token, dat) }
func NewArgumentToken(dat f.Argumented) Token  { return newToken(Argument_Token, dat) }
func NewParameterToken(dat f.Parametric) Token { return newToken(Parameter_Token, dat) }
func NewDataTypeToken(dat d.Typed) Token       { return newToken(Data_Type_Token, dat.Flag()) }
func NewDataValueToken(dat d.Data) Token       { return newToken(Data_Value_Token, dat) }
func NewPairValueToken(dat f.Paired) Token     { return newToken(Pair_Value_Token, dat) }
func NewTokenCollection(dat ...Token) Token    { return newToken(Token_Collection, tokens(dat)) }
func NewKeyValToken(key, val d.Data) Token {
	return newToken(
		Parameter_Token,
		f.NewKeyValueParm(key, val),
	)
}

type TokVal struct {
	tok TokType
	d.BitFlag
}

func (t TokVal) TokType() TokType { return t.tok }
func (t TokVal) Flag() d.BitFlag  { return t.BitFlag.Flag() }
func (t TokVal) Type() d.BitFlag  { return t.tok.Flag() }

type dataTok struct {
	TokVal
	d.Data
}

func (t dataTok) TokType() TokType { return t.TokVal.TokType() }
func (d dataTok) Flag() d.BitFlag  { return d.Data.Flag() }
func newToken(t TokType, dat d.Data) Token {
	switch t {
	case Syntax_Token:
		return TokVal{Syntax_Token, dat.(d.BitFlag)}
	case Data_Type_Token:
		return TokVal{Data_Type_Token, dat.(d.BitFlag)}
	case Kind_Token:
		return TokVal{Kind_Token, dat.(d.BitFlag)}
	case Type_Token:
		return dataTok{TokVal{Type_Token, dat.Flag()}, dat}
	case Argument_Token:
		return dataTok{TokVal{Argument_Token, dat.Flag()}, dat.(f.Argumented)}
	case Parameter_Token:
		return dataTok{TokVal{Parameter_Token, dat.Flag()}, dat.(f.Parametric)}
	case Data_Value_Token:
		return dataTok{TokVal{Data_Value_Token, dat.Flag()}, dat.(d.Data)}
	case Pair_Value_Token:
		return dataTok{TokVal{Pair_Value_Token, dat.Flag()}, dat.(f.Paired)}
	case Token_Collection:
		return dataTok{TokVal{Token_Collection, dat.Flag()}, dat.(tokens)}
	case Tree_Node_Token:
		return dataTok{TokVal{Tree_Node_Token, dat.Flag()}, dat.(f.Parametric)}
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
