/*
  OBJECT CODE, STACK FRAMES & STATE

    struct to hold the info table flags common to all heap and stack onjects.
    thanks to golang memory layout, those flags are held perfectly tight (not a
    singele bit lost to alignment, or struct headers). the 'payload' of
    arguments, bytecode, or whatever (depends on frame type), is kept as value
    of type interface (which is a pointer) in a slice. all other info get's
    copyed to the stack. arguments are usually references, but may also be copys
    of (aliased, see data) values, to get them allocated on the 'real' stack
    during runtime. the fayout field marks, which arguments of type references
    with 'one'. frame-/ & object implementations, that utilze native type
    arguments are expected to embed a copy of the base info table, followed by
    additional struct fields, one field per argument, named according to the
    name of the corresponding layout constant.

    aiming for a fast implementation, the design trys to keep data from escaping
    to the heap, by using pointers as little as possible. Godeeps internal heap
    implements a directed cyclic graph and can't be implemented in a more
    performant way (that i know of), than using pointers. all all data belonging
    to one object, is held thightly in a struct. contained values and arguments
    may be enforced to be arranged sequentially in a consequtive piece of memory,
    by serializing them into a byte array using encoding/gob.

    the closure that needs to be evaluated to yield the value, will allways
    involve some sort of name lookup and a function call at least, which is why
    using a call to an interface, can't hurt that much either‥. TODO: that needs
    of course to be validated later.

    godeeps internal stack, is a slice of fixed size structs, which should help
    keep things & stuff from ecaping to the heap as much as possible, while
    still allowing for a stack of arbitrary size. since frames have to be of a
    fixed size to make the runtime use arrays of values instead of pointers. to
    make that possible, arguments and free variables need to go some place
    else‥. they are accessable by dereferrencing the pointer to the value
    closure, which needs to be dereferenced and evaluated anyway, so accessing
    it's data should not add extra cpu cycles. for performance critical tasks,
    serialized versions of the values will be accessable via object and can be
    memcopyed to the argument stack. each frame has a length field, so that the
    runtime can access and pop the argument stack synchronously to the execution
    stack. the argument stack may be accessed using the encoding/gob interfaces,
    but that would involve dynamicly dispatched method calls. also gobs data
    structures are pretty well documented, so it's been suggested, to rather
    access those values directly based on length field and offset. in the end,
    that's of course left to the particular object implementation.

    we'll see how good this design will hold up aginst reality‥. otherwise some
    reimplementation involving fixed size arrays and/or use of the unsafe
    package will most likely occur. will be evaluated later)

    for argument sets > 32 args, an array (mem-copy, fixed size, no slice, to
    get it on the 'real' stack during computation) per argument type is expected

    TODO: come up with a naming convention for arrays of memcopyed values.

    TODO: see if argument reference/copy tagging can also be used to implement
    currying, partial application and discrimination between strong-/ and weak
    normal form (neccessary to implement full laziness, aka 'knowing when to
    stop', or 'suspended evaluation') →  compare number of pointer type
    arguments that are left to evaluate, with the functions arity and copy the
    results of argument expansion as values, deleting their tags, so that it
    could be evaluate if expression is saturated, evaluated and atomic in a
*/
package parse

import (
	"fmt"

	d "github.com/JoergReinhardt/godeep/data"
	f "github.com/JoergReinhardt/godeep/functions"
)

type info struct {
	// o/ftype   uint16	      word 0 ⇒ since obj/frame embed info
	Arity     // uint8	      ‥.
	Propertys // Uint8	      ‥.
	Layout    // Uint32	      ‥.
	//
	// particular implementations of heap object and stack frame add
	// additional struct fields following the embedded info struct. their
	// layout can be inferred by object, or frame type and may include
	// arguments of type value (instead of reference)
}

func newInfo(a Arity, p Propertys, l Layout) info { return info{a, p, l} }

type Ftype uint8

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
	ReturnByteCode
	ReturnFunction
	ReturnValue
	ReturnArray
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
	info              // struct48
	Ftype             // uint16
	Closure f.NaryFnc // ptr32
	Return  *object   // ptr32
	numVals uint32    // uint32
}

func newFrame(
	a Arity,
	p Propertys,
	l Layout,
	f Ftype,
	c f.NaryFnc,
	r *object,
	n uint32,
) frame {
	return frame{newInfo(a, p, l), f, c, r, n}
}

type stack []frame

func newStack() stack { return stack([]frame{}) }

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
// type of heap object
type Otype uint16

//go:generate stringer -type=Otype
const (
	DataConstructorDynamic Otype = 1
	FunctionClosureDynamic Otype = 1 << iota
	DataConstructorStatic
	FunctionClosureStatic
	PartialApplication
	CallContinuation
	ByteCodeObject
	Indirection
	BlackHole
	Thunk
	Value
	Array
	IOSync   // blocking io reader/writer
	IOAsync  // non-blocking io buffer
	IOShared // shared value with mutex
)

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

type object struct {
	info              // struct48
	Otype             // uint16
	Closure f.Value   // [ptr32,ptr32]
	Refs    []*object // slice/interface.
	// particular implementations of object append additional fields, and
	// embedd an instance of this type.

}

func newObject(
	o Otype,
	a Arity,
	p Propertys,
	l Layout,
	i info,
	c f.Value,
	r ...*object,
) object {
	return object{newInfo(a, p, l), o, c, r}
}

// how many arguments are expected (also see layout)
type Arity uint8

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
type Propertys uint8

//go:generate stringer -type Propertys
const (
	Default Propertys = 0
	PostFix Propertys = 1
	InFix   Propertys = 1 + iota
	// ⌐: PreFix
	Complex
	// ⌐: Atomic
	Eager
	// ⌐: Lazy
	Right
	// ⌐: Left_Bound
	Mutable
	// ⌐: Imutable
	Effected
	// ⌐: Pure
	Atomic
	// ⌐: Composit
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

// marks which arguments are references
type Layout uint32

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

// STATE
type StateFn func() StateFn

// runs state function in a loop and block until final state is reached
func (s StateFn) Run() {
	var state = s()
	for state != nil {
		state = state()
	}
}

type symbols map[string]*object

func newSymbols() symbols { return make(map[string]*object) }

type state struct {
	tree *object
	stack
	symbols
}

func (s state) lookup(name string) *object        { return s.symbols[name] }
func (s *state) declare(name string, obj *object) { (*s).symbols[name] = obj }

// push pushes a new stack frame referencing a heap node
func (s *state) push(f frame) { (*s).stack = append(s.stack, f) }

// pop pops the topmost frame and replaces the heap root, with the object
// referenced as return address.
func (s *state) pop() {
	if len(s.stack) > 0 {
		var f frame
		(*s).stack, f = s.stack[:len(s.stack)-1], s.stack[len(s.stack)-1]
		// make return address to be the current tree node
		(*s).tree = f.Return
		// fetch arguments neccessary to evaluate closure
		(*s).fetch()
	}
}

// fetches arguments by accessing the heap objects references. checks for
// amoticity and either evaluates values that are atomic, or initiates
// evaluation of arguments, that are not, while suspending the current function
// call. when all accessable arguments are fetched, depending on if they over-,
// under- or just saturate the call the according type of function application
// will be evaluated next.
func (s *state) fetch() {
	var argcount = 0
	for i, ref := range s.tree.Refs {
		// is this a reference to an argument?
		switch {
		case ref.TypeHO().Flag().Match(f.Argument):
			(*s).fetchArgument(i, ref)
		case ref.TypeHO().Flag().Match(f.Parameter):
			(*s).fetchParameter(i, ref)
		case ref.TypeHO().Flag().Match(f.Unbound):
			(*s).fetchUnbound(i, ref)
		}
	}
	// the argument set can now be either over-, under-, or exactly
	// saturated →  decide which type of closure application is demanded
	var arity = int(s.tree.Arity)
	switch {
	case argcount > arity:
		// if there are more arguments than expected by the closure,
		// push call continuation frame, saving the surplus arguments
		// and evaluate closure.
		(*s).callContinuation()
	case argcount == arity:
		// if there is exactly the right number of arguments, just
		// apply the closure
		(*s).saturatedCall()
	case argcount < arity:
		// if there are some argument, but not enough to saturate the
		// call, allocate partial application
		(*s).partialCall()
	}
}

// next 'tye the knot' by returning a new instance of the state function type,
// wrapping the mutated state.
func next(s *state) StateFn { return func() StateFn { return next(s) } }

func (s *state) fetchArgument(pos int, ref *object)  {}
func (s *state) fetchParameter(pos int, ref *object) {}
func (s *state) fetchUnbound(pos int, ref *object)   {}
func (s *state) suspendCall(refs *object) {
	// build suspension
}
func (s *state) partialCall()                     {}
func (s *state) saturatedCall()                   {}
func (s *state) callContinuation(refs ...*object) {}

func (s *state) updateThunk()  {}
func (s *state) evaluateCase() {}

// initialize state with a slice of bytecode objects. slice contains program
// code that has been precompiled (or written by hand) and concatenated to a
// slice of byte code objects. all references within bytecode either
// identifable by name and eventually subsequently called by name lookup
// utilizing the symbol table, or included in the bytecode slice as data
// constructor object with all it's arguments saturated, by either being
// directly defined as literal expression or, as slice index based reference to
// an atomic value object, that must be declared by a literal contained in the
// same slice.
func initState(bytecode []object) (StateFn, error) {
	var tree = new(object)
	var stack = newStack()
	var symbols = newSymbols()
	var state = &state{tree, stack, symbols}
	err := load(state, bytecode) //"*",8,1 :)
	if err != nil {
		return nil, err
	}
	return StateFn(func() StateFn { return next(state) }), nil
}
func load(*state, []object) error {
	var count int
	var obj object

	return fmt.Errorf(
		"this is not a valid program!\n"+
			"failed to link object code at index position %d\n"+
			" object:\n%s", count, obj)
}
