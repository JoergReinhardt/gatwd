package run

import (
	d "github.com/JoergReinhardt/godeep/data"
	f "github.com/JoergReinhardt/godeep/functions"
	l "github.com/JoergReinhardt/godeep/lex"
	p "github.com/JoergReinhardt/godeep/parse"
)

////////// STACK ////////////
//// LAST IN FIRST OUT //////
type Stacked interface {
	Collected
	Push(f.Functional)
	Pop() f.Functional
	Add(...f.Functional)
}

///////// QUEUE /////////////
//// FIRST IN FIRST OUT /////
type Queued interface {
	Collected
	Put(f.Functional)
	Pull() f.Functional
	Append(...f.Functional)
}

type StateFn func() StateFn

// STATE BYTES READER|WRITER
// reads len(t) bytes from the underlying state monad. returns number of
// bytes read and any errors occured while doing so
type StateReader interface {
	Read(p []byte) (int, error)
}

// reads len(t) bytes from the underlying state monad. returns number of
// bytes read and any errors occured while doing so
type StateWriter interface {
	Write(p []byte) (int, error)
}

// TOKEN READER|WRITER
// writes len(t) tokens to the underlying state monad. returns the number of
// tokens read and any errors occured during so.
type TokenWriter interface {
	Write(t []p.Token) (int, error)
}

// reads len(t) tokens from the underlying state monad. returns number of
// tokens read and any errors occured while doing so
type TokenReader interface {
	Read(t []p.Token) (int, error)
}

// ITEM READER|WRITER
// items, as to be expected from a lexer output
type ItemWriter interface {
	Write(t []l.Item) (int, error)
}
type ItemReader interface {
	Read(t []l.Item) (int, error)
}

// data reader|writer
type DataReader interface {
	Read(d d.Data) (int, error)
}
type DataWriter interface {
	Write(d d.Data) (int, error)
}

// data reader|writer
type FunctionalReader interface {
	Read(d f.Functional) (int, error)
}
type FunctionalWriter interface {
	Write(d f.Functional) (int, error)
}
