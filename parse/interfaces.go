package parse

import d "github.com/joergreinhardt/gatwd/data"
import f "github.com/joergreinhardt/gatwd/functions"

// data to parse
type Token interface {
	d.Native
	TypeTok() TyToken
	Data() d.Native
}

// Ident interface{}
//
// the ident interface is implemented by everything providing unique identification.
type Ident interface {
	f.Functional
	Ident() f.Functional // calls enclosed fnc, with enclosed parameters
}
