package functions

import (
	"sort"

	d "github.com/JoergReinhardt/godeep/data"
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

//// FUNCTION TYPES ////
type (
	// propertys to implement the function interface
	Signature   func() (id int, sig []Token)
	Implement   func() (id int, sig []Token, fnc Function)
	IdGenerator func() (int, IdGenerator)
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

func matchSig(short, long []Token) bool {
	for i, s := range short {
		if s != long[i] {
			return false
		}
	}
	return true
}
func MatchSig(siga, sigb Signature) bool {
	_, a := siga()
	_, b := sigb()
	if len(a) < len(b) {
		return matchSig(a, b)
	}
	return matchSig(b, a)
}

func genCount() IdGenerator {
	return func() (int, IdGenerator) {
		var id int
		var gen IdGenerator
		gen = func() (int, IdGenerator) {
			id = id + 1
			return id, gen
		}
		return id, gen
	}
}

var (
	names      = map[string]Signature{}
	signatures = []Signature{}
	implements = []Implement{}
	initSig    = genCount()
	initImp    = genCount()
)

type token struct {
	flag d.BitFlag
}

func (t token) Flag() d.BitFlag { return t.flag }

func (t token) String() string { return t.flag.String() }

type TokSlice [][]token

func (t TokSlice) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t TokSlice) Len() int           { return len(t) }
func (t TokSlice) Less(i, j int) bool { return t[i][0].Flag() < t[j][0].Flag() }
func sortSlice(t TokSlice) TokSlice {
	sort.Sort(t)
	return t
}
func byToken(t TokSlice, match token) [][]token {
	ret := [][]token{}
	i := sort.Search(len(t), func(i int) bool { return t[i][0].flag.Uint() >= match.flag.Uint() })
	var j = i
	for j < len(t) && t[j][0].flag.Uint() == match.flag.Uint() {
		ret = append(ret, t[j])
		j++
	}
	return ret
}
func decapTokSlice(t TokSlice) TokSlice {
	var rs = [][]token{}
	for _, t := range t {
		if len(t) > 0 {
			rs = append(rs, t[1:])
		}
	}
	return rs
}
