package parse

import d "github.com/JoergReinhardt/gatwd/data"
import f "github.com/JoergReinhardt/gatwd/functions"

// data to parse
type Token interface {
	d.Primary
	TypeTok() TyToken
}

// Ident interface{}
//
// the ident interface is implemented by everything providing unique identification.
type Ident interface {
	f.Value
	Ident() f.Callable // calls enclosed fnc, with enclosed parameters
}
