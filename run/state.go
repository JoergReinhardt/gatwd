package run

import (
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
func (a Arity) TypePrim() d.TyPrimitive       { return d.Flag }
func (a Arity) TypeHO() f.TyHigherOrder       { return f.HigherOrder }

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
	Effected
	// ⌐: Pure
	Data
	// ⌐: Function
)

func (p Propertys) TypePrim() d.TyPrimitive       { return d.Flag }
func (p Propertys) TypeHO() f.TyHigherOrder       { return f.HigherOrder }
func (p Propertys) Flag() d.BitFlag               { return p.TypeHO().Flag() }
func (p Propertys) Eval(a ...d.Primary) d.Primary { return p.Flag() }
func (p Propertys) Match(a Propertys) bool {
	if p&a != 0 {
		return true
	}
	return false
}

type Length d.Uint32Val

type info struct {
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

func newInfo(len Length, a Arity, p Propertys) info { return info{len, a, p} }

// OBJECT
//
// object base info has an object type flag and embeds a copy of the info
// struct. to implement objects that add fields to the info table, copys of the
// base info struct can be embedded as first field. again  no alignment, or
// header loss‥.  gotta love go for that alone.  native arguments and/or free
// variables intendet to be treated as memcopyed natives, will be serialized by
// encoding/gob and written to the values field. the embedded closure can be
// staticly linked function, as in declared pre compilation and part of the
// compiled object code. top level (no non-constant free variables) callable by
// name; or a closure defined dynamicly during runtime as closure literal in a
// heap object, that may or may not be known by name as either local, or
// top-level variablble, or constant.
type object struct {
	info            // struct64
	Otype           // uint16
	Expr  f.Value   // [ptr32,ptr32]
	Refs  []*object // references
	// particular implementations of object append additional fields, and
	// embedd an instance of this type.

}

func (t Otype) TypeObj() Otype                { return t }
func (t Otype) TypeHO() f.TyHigherOrder       { return f.HigherOrder }
func (t Otype) TypePrim() d.TyPrimitive       { return d.Flag }
func (t Otype) Flag() d.BitFlag               { return d.BitFlag(t) }
func (t Otype) Eval(p ...d.Primary) d.Primary { return d.BitFlag(t) }
func (t Otype) Match(o Otype) bool {
	if t&o != 0 {
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
	DataConstructor
	FunctionClosure
	Declaration
	Indirection
	BlackHole
	ByteCode
	Thunk
	IOEvents // event subscritptions
	IOReader // blocking io reader
	IOWriter // blocking io writer
	IOShared // shared value with mutex
)

// STATE FUNCTION
type stateFnc func() stateFnc

// RUN STATE TRANSITION
// runs state function in a loop and block until final state is reached
func (s stateFnc) Run() {
	var sf = s()
	for sf != nil {
		sf = sf()
	}
}

// STACK
// frame info has a frame type flag, info field containing the closure code to
// run (we'll see, if the go compiler is able to inline the closure, or have it
// as stack value, while keeping the frame length fixed. the numArgs field
// tells the runtime/compiler how many memcopyed, gob-encoded values associated
// with this frame lay on the argument stack. the argument stack is a byte
// array of gob encoded values that all start with a length field. the runtime
// provides transparent access to arguments and values to the closure it calls,
// independent from beeing values or pointers. it also pushes & pop's the
// argument stack, whenever neccessary.
//
// FRAME
type frame struct {
	Ftype // uint16
	info  // struct64
	f.Value
	*object
}

func newFrame(
	ftype Ftype,
	obj *object,
) frame {

	return frame{
		ftype,
		obj.info,
		obj.Expr,
		obj,
	}
}

type Ftype d.Uint8Val

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

// SYMBOL TABLE
type symbols map[string]*object

func newSymbols() symbols { return make(map[string]*object) }

// SYMBOL DEFINITION
func let(s symbols, name string, obj *object) symbols { s[name] = obj; return s }
func lookup(s symbols, name string) *object           { return s[name] }

type stack []frame

func newStack() stack { return []frame{} }

// pop()
//
// pops the topmost frame and returns it. stack never runs out of frames and
// generates empty frames for calls to pop that preceed popping of the last
// frame.
func pop(s stack) (frame, stack) {
	var f frame
	var length = len(s)
	switch length {
	case 0:
		s, f = []frame{}, frame{}
	case 1:
		s, f = []frame{}, s[0]
	default:
		s, f = s[:length-2], s[length-1]
	}
	return f, s
}

// push pushes a new stack frame referencing a heap node
func push(s stack, f frame) stack { s = append(s, f); return s }

type state struct {
	top *object
	stack
	symbols
}

// manipulate stack, heap & symbol table of the statestructure
func (s state) lookup(name string) *object { return lookup(s.symbols, name) }
func (s state) let(name string, o *object) { s.symbols = let(s.symbols, name, o) }

// pointer. stack index reversed to make stack appear to grow from the top →
// offset addressed (stable) stack pointers, without crashing the performance,
// since frames can be appended without having to offset the slices tail.
func (s state) frame(off int) frame { return s.stack[len(s.stack)-1-off] }
func (s state) push(f frame)        { s.stack = push(s.stack, f) }
func (s state) pop() (f frame)      { f, s.stack = pop(s.stack); return f }

// state maintains a reference to the current top of the heap.
func (s state) heap() *object { return s.top }
func (s state) put(o *object) { s.top = o }

// FIND & REDUCE NEXT REDEX
func (s state) next() stateFnc {
	var nextState stateFnc
	return nextState
}
func (s state) suspendCall() {}

////
func (s state) partialCall()      {}
func (s state) saturatedCall()    {}
func (s state) callContinuation() {}

////
func (s state) updateThunk() {}
func (s state) updateCase()  {}

// INITIALIZATION
//
// initialize state with a slice of bytecode objects. slice contains references
// to preallocated heap objects of the declaration type. declaration objects
// declare either named top level variables, named localy variables, or
// anonymous localy defined variables. named declarations yield their name,
// when evaluated. top level declarations reference the single object, the
// declared name is supposed to be pointing to. local declarations reference
// the object that forms the top of the scope this variable is declared in and
// a reference to the values object as second reference.
func initState(program []*object) stateFnc {
	var tree = new(object)
	var stack = newStack()
	var symbols = newSymbols()
	var state = &state{tree, stack, symbols}

	return stateFnc(func() stateFnc { return state.next() })
}
