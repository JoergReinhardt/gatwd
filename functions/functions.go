package functions

import (
	"sort"

	d "github.com/JoergReinhardt/godeep/data"
	l "github.com/JoergReinhardt/godeep/lang"
)

///
//// Functional higher order types ////
// takes a state and advances it. returns the next state fn to run
type Flag d.BitFlag

func ComposeFlag(high, low Flag) Flag {
	return Flag(d.High(d.BitFlag(high)).Flag() | d.Low(d.BitFlag(low)).Flag())
}

func (t Flag) String() string  { return DataType(t).String() }
func (t Flag) Low() Flag       { return Flag(d.Low(d.BitFlag(t)).Flag()) }
func (t Flag) High() Flag      { return Flag(d.High(d.BitFlag(t)).Flag()) }
func (t Flag) Uint() uint      { return uint(t) }
func (t Flag) Flag() d.BitFlag { return d.BitFlag(t) }

type DataType Flag

func (t DataType) Flag() d.BitFlag { return d.BitFlag(t).Flag() }
func (t DataType) Uint() uint      { return d.BitFlag(t).Uint() }

//go:generate stringer -type=DataType
const (
	/// FUNCTIONAL ATTRIBUTES
	Parameter DataType = 1 << iota
	Argument
	Return
	/// FUNCTIONAL DATATYPES
	Tuple
	List
	Chain
	UniSet
	MuliSet
	AssocA
	Record
	Link
	DLink
	Node
	Tree

	Attributes = Parameter | Argument | Return

	Recursives = Tuple | List
	Sets       = UniSet | MuliSet | AssocA | Record
	Links      = Link | DLink | Node | Tree // Consumeables
)

type (
	// functional base types
	Data     func(d.Data) d.Data
	Vector   func() []Data
	Constant func() Data
	Unary    func(d Data) Data
	Binary   func(a, b Data) Data
	Nnary    func(...Data) Data
	// higher order function types
	Generator   func() (Data, Generator)
	Predicate   func(Data) bool
	Condition   func(d Data) bool // true if predicate(d) == true
	Conditional func(Data) Data   // returns either data, or not
)

type idGenerator func() (int, idGenerator)

func genCount() idGenerator {
	return func() (int, idGenerator) {
		var id int
		var gen idGenerator
		gen = func() (int, idGenerator) {
			id = id + 1
			return id, gen
		}
		return id, gen
	}
}

// TYPESPEC STATE
var (
	names = map[string]Polymorph{}
	sig   = []Signature{}
	iso   = []Isomorph{}  // sig & fnc
	poly  = []Polymorph{} // []sig & []fnc
	uid   = genCount()
)

func conId() int { var id int; id, uid = uid(); return id }

// PARTS OF TYPE SPEC
type (
	Signature func() (id int, sig []Token)                          // <- 1 : 1 type/data cons., ops‥. (tokens)
	Isomorph  func() (pid int, id int, sig Signature, fnc Function) // <- 1 : 1 implementation  (golang)
	Polymorph func() (id int, sig Signature, iso []Isomorph)        // 1 : n id/Isomorphisms (pattern matching)
	NamedDef  func() (id int, name string, p Polymorph)             // 1 : 1 name/Polymorphism
)

func conSignature(tok ...Token) Signature {
	return func() (id int, sig []Token) {
		return conId(), sig
	}
}
func conIsomorph(pid int, sig Signature, fnc Function) Isomorph {
	return func() (
		pid int,
		id int,
		sig Signature,
		fnc Function,
	) {
		id, _ = sig()
		return pid, id, sig, fnc
	}
}
func conPolymorph(sig Signature, iso ...Isomorph) Polymorph {
	return func() (
		id int,
		sig Signature,
		iso []Isomorph,
	) {
		id, _ = sig()
		return id, sig, iso
	}
}

// TOKEN GENERATION
type TokType uint8

//go:generate stringer -type TokType
const (
	Syntax TokType = 1 << iota
	Symbol
	Number
	Data_Type
	Data_Value
)

type token struct {
	flag d.BitFlag
	typ  TokType
}
type dataToken struct {
	token
	d d.Data
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
	}
	return str
}

func conToken(t TokType, dat d.Data) Token {
	switch t {
	case Syntax:
		return token{dat.Flag(), Syntax}
	case Data_Type:
		return token{dat.Flag(), Data_Type}
	case Data_Value:
		return dataToken{token{dat.Flag(), Data_Value}, dat}
	}
	return nil
}

type Tokens []Token

func (t Tokens) String() string {
	var str string
	for _, tok := range t {
		str = str + " " + tok.String()
	}
	return str
}

type TokSlice [][]Token

func (t TokSlice) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t TokSlice) Len() int           { return len(t) }
func (t TokSlice) Less(i, j int) bool { return t[i][0].Flag() < t[j][0].Flag() }
func sortTokens(t TokSlice) TokSlice {
	sort.Sort(t)
	return t
}
func decapTokSlice(t TokSlice) ([]Token, TokSlice) {
	if len(t) > 0 {
		if len(t) > 1 {
			return t[0], t[1:]
		}
		return t[0], nil
	}
	return nil, nil
}
func byToken(t TokSlice, match token) [][]Token {
	ret := [][]Token{}
	i := sort.Search(len(t), func(i int) bool { return t[i][0].Flag().Uint() >= match.Flag().Uint() })
	var j = i
	for j < len(t) && d.Match(t[j][0].Flag(), match.Flag()) {
		ret = append(ret, t[j])
		j++
	}
	return ret
}
func matchSigByTokSlice(sig []Token, matches TokSlice) bool {
	match, matches := decapTokSlice(matches)
	if len(sig) == 0 {
		return true
	}
	if !sigsMatch(sig, match) {
		return false
	}
	return matchSigByTokSlice(sig, matches)
}
func sigsMatch(sig, match []Token) bool {
	if len(sig) > len(match) {
		return smatch(sig, match)
	}
	return smatch(match, sig)
}
func smatch(long, short []Token) bool {
	if len(short) == 0 {
		if len(long) != 0 {
			return false
		}
		return true
	}
	l, s := long[0], short[0]
	if !d.Match(l.Flag(), s.Flag()) {
		return false
	}
	return smatch(long[1:], short[1:])
}

// SIGNATURE MATCHING
type SigSlice []Signature

func (s SigSlice) Len() int { return len(s) }
