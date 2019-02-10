package run

import (
	d "github.com/JoergReinhardt/godeep/data"
	f "github.com/JoergReinhardt/godeep/functions"
)

// INFO TABLE
//
// marks which arguments are references
type Layout d.Uint32Val

func (l Layout) IsValue(i int) bool { return l&Layout(1<<uint(i)) != 0 }

//go:generate stringer -type Layout
const (
	Arg0 Layout = 1
	Arg1 Layout = 1 << iota
	Arg2
	Arg3
	Arg4
	Arg5
	Arg6
	Arg7
	Arg8
	Arg9
	Arg10
	Arg11
	Arg12
	Arg13
	Arg14
	Arg15
	Arg16
	Arg17
	Arg18
	Arg19
	Arg20
	Arg21
	Arg22
	Arg23
	Arg24
	Arg25
	Arg26
	Arg27
	Arg28
	Arg29
	Arg30
	Arg31
)

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
func (a Arity) TypeHO() f.TyHigherOrder       { return f.Internal }

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
func (p Propertys) TypeHO() f.TyHigherOrder       { return f.Internal }
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

// FRAME
//
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

// frame info has a frame type flag, info field containing the closure code to
// run (we'll see, if the go compiler is able to inline the closure, or have it
// as stack value, while keeping the frame length fixed. the numArgs field
// tells the runtime/compiler how many memcopyed, gob-encoded values associated
// with this frame lay on the argument stack. the argument stack is a byte
// array of gob encoded values that all start with a length field. the runtime
// provides transparent access to arguments and values to the closure it calls,
// independent from beeing values or pointers. it also pushes & pop's the
// argument stack, whenever neccessary.
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
		obj.Value,
		obj,
	}
}

type stack []frame

func newStack() stack { return []frame{} }

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
	Value f.Value   // [ptr32,ptr32]
	Refs  []*object // slice/interface.
	// particular implementations of object append additional fields, and
	// embedd an instance of this type.

}

func (t Otype) TypeObj() Otype                { return t }
func (t Otype) TypeHO() f.TyHigherOrder       { return f.Internal }
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
	IOSync   // blocking io reader/writer
	IOAsync  // non-blocking io buffer
	IOShared // shared value with mutex
)

////
func (s *state) suspendCall() {}

////
func (s *state) partialCall()      {}
func (s *state) saturatedCall()    {}
func (s *state) callContinuation() {}

////
func (s *state) updateThunk() {}
func (s *state) updateCase()  {}

// STATE FUNCTION
//
type stateFnc func() stateFnc

// RUN STATE TRANSITION
//
// runs state function in a loop and block until final state is reached
func (s stateFnc) Run() {
	for state := s(); state != nil; {
		state = state()
	}
}

// SYMBOL TABLE
//
type symbols map[string]*object

func newSymbols() symbols { return make(map[string]*object) }

type state struct {
	tree *object
	stack
	symbols
}

// DECLARATION
//
func (s state) lookup(name string) *object    { return s.symbols[name] }
func (s *state) let(name string, obj *object) { (*s).symbols[name] = obj }

// STACK
//
// push pushes a new stack frame referencing a heap node
func (s *state) push(f frame) { (*s).stack = append(s.stack, f) }

// pop pops the topmost frame and replaces the heap root, with the object
// referenced as return address.
func (s *state) pull() (f frame) {
	var length = len(s.stack)
	switch length {
	case 0:
		(*s).stack, f = []frame{}, frame{}
	case 1:
		(*s).stack, f = []frame{}, s.stack[0]
	default:
		(*s).stack, f = s.stack[:length-2], s.stack[length-1]
	}
	return f
}

// FIND & REDUCE NEXT REDEX
func (s *state) next() stateFnc {
	var nextState stateFnc
	return nextState
}
func (s *state) fetchArgument(pos int, ref *object)  {}
func (s *state) fetchParameter(pos int, ref *object) {}
func (s *state) fetchUnbound(pos int, ref *object)   {}

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
