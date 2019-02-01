package run

import (
	"io"

	d "github.com/JoergReinhardt/godeep/data"
	f "github.com/JoergReinhardt/godeep/functions"
	l "github.com/JoergReinhardt/godeep/lex"
	p "github.com/JoergReinhardt/godeep/parse"
)

/// STATE BYTES READER|WRITER
//
// type State f.ParamSet
//
// the execution state is enclosed by the state function and implemented as
// generic set of parameters. it's entirely in the responsibility of the code
// contained by instances of the state function, on deciding how to structure
// and initioalize state, what to do with it, once it's initialized and so on‥.
// it's also in the state functions responsibility, to provide some mean of
// input/output, or other side effect, or alternatively contain all neccessary
// data and do go knows what with the result. it's be considered best paxis, to
// provid  a 'create'-/ or init function to instanciate a state function, that
// either already contains it's state generated from arguments passed, or
// looked up during initialization and/or either pass, or yield some kind of
// callback, reader/writer, channel, or the like, to connect the execution
// environment to side effects intended to be interacted with by the code
// running.
//
// this approach provides maximal flexibility. state functions run in a loop,
// can implement every possible automaton described by automate theory and
// everything that's computable can be computed in a, possibly endless series
// of calls to the state functions returned by former calls to state
// functions‥. and why that's great'nall it also leaks any structure or
// actual usability, right out of the box.
//
// to make it useful godeep among other things, implements a runtime to execute
// code that's syntacticly inspired by the haskell language. since lexers are
// tedious and painsrakingly workintensive to handcode, godeep implements the
// least neccessary base functionality of runtime and type system first to than
// münchhausen it's way to a full language runtime, parser and lexer, that use
// the functional concepts  and datastructures implemented in godeep, to
// implement any further lexing, parsing, linking and pattern matching.

// type StateFn func(State) StateFn
//
// state function closes over a State instance and progresses on it calling an
// enclosed function passing (parts of) the state as parameters and applying
// the results to yield new, altered state instances, when called. a new
// StateFn closing over the altered state and a new function to call, is then
// yielded in return, to be called next in a for loop, that's expected to run
// until nil is returned at some point (or execution get's abortet, paused
// haltet, heatdeath of universe occures‥. details may vary from implementation
// to implementation)
type StateFn func() StateFn

// runtime deals with declaration, instancialization, tree of instances and the
// currently evaluated node. provides the methods to either declare an
// instance, or (possibly partly) apply the arguments and/or replace the
// instance by the return value it yielded.
//
// the next method returns a node based on it's index position (== id!) casted
// as instance of the state function type, to return to the enclosing trampolin
// execution loop as next stateFn to run.
type Runtime interface {
	Next(uid int)
	Current() Declared
	Push(poly string) (uid int)
	Symbol(symbol string) Declared
	Declare(symbol string, instance Declared) (uid int)
}
type Declared interface {
	Uid() int
	Caller() int
	Poly() p.Polymorph
}

// runtime extensions, to connect to side effects via reader/writer pattern
type IOReaderWriter interface {
	Read(io.Writer) (int, error)
	Write(io.Reader) (int, error)
}
type Events interface {
	Subscribe(f.Function)
}

////////// STACK ////////////
//// LAST IN FIRST OUT //////
//
// data structure to implement push down automata
type Stacked interface {
	d.Collected
	Push(d.Data)
	Pop() d.Data
	Add(...d.Data)
}

///////// QUEUE /////////////
//// FIRST IN FIRST OUT /////
//
// data structure to feed data or instructions to the execution environment
type Queued interface {
	d.Collected
	Put(f.Functional)
	Pull() d.Data
	Append(...d.Data)
}

// to be go idiomatic, streams of data should be representable as
// readers/writers. underlying implementations may be concurrent by using
// locks, buffers, io.pipe reader, channels, callbacks‥.
//
// godeep provides reader/writer interfaces for data streams of different type.
// byte arrays as least common denominator can encode all other types, but need
// to be serialized and parsed explicitly . Token, Item, Data and Functional
// writer on the other hand, are convienient to use and optimized on the type
// of 'input' dealt with by either lexer, parser, or runtime environment
// implementations.  a protobuf implementation will most likely be following sooning soon.
//
// type StateReader interface{}
//
// reads len(t) bytes from the underlying state monad. returns number of bytes
// read and any errors occured while doing so
type StateReader interface {
	Read(io.Writer) (int, error)
}
type StateBytesReader interface {
	ReadBytes([]byte) (int, error)
}

// type StateWriter interface {}
//
// reads len(t) bytes from the underlying state monad. returns number of
// bytes read and any errors occured while doing so
type StateWriter interface {
	Write(io.Reader) (int, error)
}
type StateBytesWriter interface {
	WriteBytes([]byte) (int, error)
}

// TOKEN READER|WRITER
//
// type TokenWriter interface {
//
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
