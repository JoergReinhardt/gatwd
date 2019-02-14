package run

import (
	"sync"

	d "github.com/JoergReinhardt/gatwd/data"
	f "github.com/JoergReinhardt/gatwd/functions"
)

// INFO TABLE
// how many arguments are expected (also see layout)
type Arity d.Uint8Val

//go:generate stringer -type Arity
const (
	Nullary Arity = 0 + iota
	Unary
	Binary
	Ternary
	Quaternary
	Quinary
	Senary
	Septenary
	Octonary
	Novenary
	Denary
)

func (a Arity) Eval(v ...d.Primary) d.Primary { return a }
func (a Arity) Flag() d.BitFlag               { return d.BitFlag(a) }
func (a Arity) TypePrime() d.TyPrime          { return d.Flag }
func (a Arity) TypeHO() f.TyFnc               { return f.HigherOrder }
func (a Arity) Match(arg Arity) bool          { return a == arg }

// properys relevant for application
type Propertys d.Uint8Val

//go:generate stringer -type Propertys
const (
	Default Propertys = 0
	PostFix Propertys = 1
	InFix   Propertys = 1 + iota
	// ⌐: PreFix
	Atomic
	// ⌐: Composit
	Eager
	// ⌐: Lazy
	Right
	// ⌐: Left_Bound
	Mutable
	// ⌐: Imutable
	SideEffect
	// ⌐: Pure
	Data
	// ⌐: Function
)

func (p Propertys) TypePrime() d.TyPrime          { return d.Flag }
func (p Propertys) TypeHO() f.TyFnc               { return f.HigherOrder }
func (p Propertys) Flag() d.BitFlag               { return p.TypeHO().Flag() }
func (p Propertys) Eval(a ...d.Primary) d.Primary { return p.Flag() }
func (p Propertys) Match(arg Propertys) bool {
	if p&arg != 0 {
		return true
	}
	return false
}

type Length d.Uint32Val

type Info struct {
	// o/ftype   uint16	      word 0 ⇒ since obj/frame embed info
	Length    // Uint32
	Arity     // uint8	      ‥.
	Propertys // Uint8	      ‥.
	//
	// particular implementations of heap object and stack frame add
	// additional struct fields following the embedded info struct. their
	// layout can be inferred by object, or frame type and may include
	// arguments of type value (instead of reference)
}

func newInfo(len Length, a Arity, p Propertys) Info { return Info{len, a, p} }

// OBJECT
//
// Object base info has an Object type flag and embeds a copy of the info
// struct. to implement objects that add fields to the info table, copys of the
// base info struct can be embedded as first field. again  no alignment, or
// header loss‥.  gotta love go for that alone.  native arguments and/or free
// variables intendet to be treated as memcopyed natives, will be serialized by
// encoding/gob and written to the values field. the embedded closure can be
// staticly linked function, as in declared pre compilation and part of the
// compiled Object code. top level (no non-constant free variables) callable by
// name; or a closure defined dynamicly during runtime as closure literal in a
// heap Object, that may or may not be known by name as either local, or
// top-level variablble, or constant.
type Object struct {
	Info            // struct64
	Otype           // uint16
	Expr  f.Value   // [ptr32,ptr32]
	Refs  []*Object // references
	// particular implementations of object append additional fields, and
	// embedd an instance of this type.

}

func newObject() *Object {
	return &Object{
		newInfo(Length(0), Arity(0), Propertys(0)),
		Otype(0),
		nil,
		[]*Object{},
	}
}

var objectPool = sync.Pool{New: func() interface{} { return newObject() }}

func allocateObject() *Object { return objectPool.Get().(*Object) }

func (o Otype) TypeObj() Otype                { return o }
func (o Otype) TypeHO() f.TyFnc               { return f.HigherOrder }
func (o Otype) TypePrime() d.TyPrime          { return d.Flag }
func (o Otype) Flag() d.BitFlag               { return d.BitFlag(o) }
func (o Otype) Eval(p ...d.Primary) d.Primary { return d.BitFlag(o) }
func (o Otype) Match(arg Otype) bool {
	if o&arg != 0 {
		return true
	}
	return false
}

type Otype d.Uint16Val

//go:generate stringer -type=Otype
const (
	PartialApplication Otype = 1
	CallContinuation   Otype = 1 << iota
	CaseContinuation
	FunctionClosure
	DataConstructor
	Declaration
	Indirection
	BlackHole
	ByteCode
	Thunk
	///// a thread contains another instance of runtime, including heap,
	// stacks and state loop possibly running asynchronously in a go
	// routine. sheduling, synchronization and data sharing are dynamicly
	// defined and allocated as objects on the heap, referenced to from the
	// io systems list of references.
	System // static code, events, signals, flags, messages, indices
	Thread // instance of a, StateFn (possibly under parallel evaluation)
	Shared // shared flag marks value thread safe, set along Sync/Async.
	Async  // call substitution queue, applyed whenever value is yielded
	Sync   // blocks until read/write can be satisfied
)

// STACK
//
// FRAME
type Frame struct {
	Ftype // uint16
	Info  // struct64
	f.Value
	*Object
	Return int
}

func (fra Frame) Eval(...d.Primary) d.Primary { return fra }
func (fra Frame) Flag() d.BitFlag             { return fra.Otype.Flag() }
func (fra Frame) TypeFnc() f.TyFnc            { return f.Type }
func (fra Frame) TypePrime() d.TyPrime        { return d.Flag }
func (fra Frame) String() string              { return fra.Ftype.String() }
func (fra Frame) Segment(segment int) *Object {
	if segment < len(fra.Object.Refs) {
		return fra.Object.Refs[segment]
	}
	return nil
}

// ALLOCATION POOL
func allocateFrame(o *Object, ftype Ftype, ret int) Frame {
	return Frame{ftype, o.Info, o.Expr, o, ret}
}

type Ftype d.Uint8Val

func (t Ftype) Eval(...d.Primary) d.Primary { return t }
func (t Ftype) TypeFnc() f.TyFnc            { return f.Type }
func (t Ftype) TypePrime() d.TyPrime        { return d.Flag }
func (t Ftype) Flag() d.BitFlag             { return d.BitFlag(uint64(d.Uint8Val(t).Uint())) }
func (t Ftype) Match(f Ftype) bool {
	if t&f != 0 {
		return true
	}
	return false
}

//go:generate stringer -type=Ftype
const (
	Select Ftype = 1
	Update Ftype = 1 << iota
	Continuation
	ReturnByteCode
	ReturnFunction
	ReturnData
)

type Stack []Frame

func (s Stack) Eval(...d.Primary) d.Primary { return s }
func (s Stack) Flag() d.BitFlag             { return s.TypePrime().Flag() }
func (s Stack) TypeFnc() f.TyFnc            { return f.HigherOrder }
func (s Stack) TypePrime() d.TyPrime        { return d.HigherOrder }
func (s Stack) String() string              { return "stack" }

func newStack() Stack { return []Frame{} }

func (s Stack) newFrame(
	ftype Ftype,
	object *Object,
	caller int,
) Frame {
	return Frame{
		ftype,
		object.Info,
		object.Expr,
		object,
		caller,
	}
}

func stackPtr(s Stack, frame int) *Frame {
	var l = len(s)
	if frame < l {
		return &(s[l-1-frame])
	}
	return nil
}

// pop()
//
// pops the topmost frame and returns it. stack never runs out of frames and
// generates empty frames for calls to pop that preceed popping of the last
// frame.
func pop(s Stack) (Frame, Stack) {
	var f Frame
	var length = len(s)
	switch length {
	case 0:
		s, f = []Frame{}, Frame{}
	case 1:
		s, f = []Frame{}, s[0]
	default:
		s, f = s[:length-2], s[length-1]
	}
	return f, s
}

// push pushes a new stack frame referencing a heap node
func push(s Stack, f Frame) Stack { s = append(s, f); return s }

// SYMBOL TABLE
type Symbols map[string]*Object

func (s Symbols) Eval(...d.Primary) d.Primary { return s }
func (s Symbols) Flag() d.BitFlag             { return s.TypePrime().Flag() }
func (s Symbols) TypeFnc() f.TyFnc            { return f.HigherOrder }
func (s Symbols) TypePrime() d.TyPrime        { return d.HigherOrder }
func (s Symbols) String() string              { return "symbols" }

func newSymbols() Symbols { return make(map[string]*Object) }

// SYMBOL DEFINITION
func let(s Symbols, name string, obj *Object) Symbols { s[name] = obj; return s }
func lookup(s Symbols, name string) *Object           { return s[name] }

// STATE FUNCTION
type StateFnc func() StateFnc

func (s StateFnc) Eval(...d.Primary) d.Primary { return s }
func (s StateFnc) Flag() d.BitFlag             { return s.TypePrime().Flag() }
func (s StateFnc) TypeFnc() f.TyFnc            { return f.HigherOrder }
func (s StateFnc) TypePrime() d.TyPrime        { return d.HigherOrder }
func (s StateFnc) String() string              { return "state function" }

//// STATE
///
// state object composes reference to heap of linked objects, stack and a
// symbol table, to hold all state information of the running process
type State struct {
	Top *Object
	Stack
	Symbols
}

func (s State) Eval(...d.Primary) d.Primary { return s }
func (s State) Flag() d.BitFlag             { return s.TypePrime().Flag() }
func (s State) TypeFnc() f.TyFnc            { return f.HigherOrder }
func (s State) TypePrime() d.TyPrime        { return d.HigherOrder }
func (s State) String() string              { return "state" }

// SYMBOLS
func (s State) Lookup(name string) *Object { return lookup(s.Symbols, name) }
func (s State) Let(name string, o *Object) { s.Symbols = let(s.Symbols, name, o) }

// STACK
//
// the frame pointer returns a reference to another stack frame based on offset
// and can be chaned with Frames 'Segment()' method, to directly reference the
// list of objects referenced by the object the frame referes to. the object
// itself can be accessed directly based on struct field name on
func (s State) Frame(offset int) *Frame { return stackPtr(s.Stack, offset) }
func (s State) Push(f Frame)            { s.Stack = push(s.Stack, f) }
func (s State) Pop() (f Frame)          { f, s.Stack = pop(s.Stack); return f }

// HEAP
func (s State) Heap() *Object { return s.Top }
func (s State) Put(o *Object) { s.Top = o }

// FIND & REDUCE NEXT REDEX (APPLY REDUCTION RULES)
func (s State) next() StateFnc {
	// apply reduction rules
	// after state has been mutated, rinse and repeat‥.
	return func() StateFnc { return s.next() }
}

///////////////////////////////////////////////////////////////////////////////
// LOAD PRE-LINKED OBJETS
//
// initialize state with a slice of bytecode objects. slice contains references
// to preallocated heap objects of the declaration type. declaration objects
// declare either named top level variables, named localy variables, or
// anonymous localy defined variables. named declarations yield their name,
// when evaluated. top level declarations reference the single object, the
// declared name is supposed to be pointing to. local declarations reference
// the object that forms the top of the scope this variable is declared in and
// a reference to the values object as second reference.
//
// init state expects the slice of heap objects containing the program to run
// to be organized in a way, that the last object is the one (either main
// function, or system) supposed to start the execution with
func loadLinkedObjectCode(program ...*Object) StateFnc {
	// allocate symbol table
	var symbols = newSymbols()
	// safe reference of last object passed. which is expected to be the
	// initial object to start evaluation at.
	var init = program[len(program)-1]
	// push initial onbject on to stack
	var stack = push(newStack(), allocateFrame(init, Continuation, 0))
	// instanciate reference to state struct
	var state = State{init, stack, symbols}
	// wrap state reference in a stateFnc closure that calls states next
	// method to generate the next state function to be returned.
	return StateFnc(func() StateFnc { return (&state).next() })
}

// LOAD & RUN PROGRAM OBJECT CODE
// TODO: once lexer, parser, linker‥. are implemented, this should take a file
// descriptor (program file, os/stdin to run in interpreter mode, websocket‥.).
func Run(program ...*Object) {
	var stateFnc = loadLinkedObjectCode(program...)
	for stateFnc != nil {
		stateFnc = stateFnc()
	}
}
