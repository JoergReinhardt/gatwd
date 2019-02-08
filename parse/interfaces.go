package parse

import d "github.com/JoergReinhardt/godeep/data"
import f "github.com/JoergReinhardt/godeep/functions"

// data to parse
type Token interface {
	TypeTok() TyToken
	TypePrim() d.TyPrimitive
	String() string
}

// Ident interface{}
//
// the ident interface is implemented by everything providing unique identification.
type Ident interface {
	f.Value
	Ident() f.Callable // calls enclosed fnc, with enclosed parameters
}
