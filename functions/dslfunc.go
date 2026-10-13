package functions

// DSLFunc is a first-class, dynamically-typed function value — the runtime
// representation of a callable that the guest DSL's infix expression language can
// apply by juxtaposition, as in `graph_add_node color id payload g` or
// `hamt_path "metadata.name" r`.
//
// # Why this type exists
//
// The host engine is statically generic (`func[T any]`), but the guest DSL is
// dynamically typed: every value flowing through a pipeline is `any` (wrapped in
// Either[error, any]). To let a DSL lambda body *invoke* a builtin — rather than
// only reference a bound value — the expression evaluator needs a uniform,
// type-erased notion of "a function you can call with a list of already-evaluated
// arguments." DSLFunc is that notion.
//
// A builtin is registered into the evaluation environment as a value of this type;
// when the parser sees an application `f a b`, it resolves `f` to a DSLFunc and
// calls it with the evaluated arguments. Returning an `error` (rather than
// panicking) keeps failures inside the DSL's Either[error, any] discipline — a bad
// call surfaces as a Left, not a crash.
//
// # Arity and currying
//
// DSLFunc takes all arguments at once as a slice, leaving arity checking to the
// function body (which returns a descriptive error on a mismatch). A function may
// implement partial application by returning another DSLFunc when under-applied;
// nested applications `(f a) b` then compose naturally, because the inner result is
// itself a DSLFunc the evaluator can apply again.
//
// # Example
//
//	inc := DSLFunc(func(args []any) (any, error) {
//	    n, ok := args[0].(int)
//	    if !ok { return nil, fmt.Errorf("inc: want int, got %T", args[0]) }
//	    return n + 1, nil
//	})
//	// in the DSL: `inc 41` evaluates to 42
type DSLFunc func(args []any) (any, error)
