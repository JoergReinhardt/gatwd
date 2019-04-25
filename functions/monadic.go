/*
 */
package functions

import (
	"math/big"
	"time"

	d "github.com/joergreinhardt/gatwd/data"
)

type (
	//// NOTHING
	NoOp func()

	//// MONADIC PRAEDICATE EXPRESSIONS
	///
	// monadic expressions return branch-, order-, compare-, lookup-,
	// count-, filter-‥. predicates as results/types at runtime by
	// returning hidden bool-, uint-, int-, string-, []byte-, and type flag
	// values as second argument to decide which continuation to return
	// according to passed arguments
	//
	BranchMonad   func(...Callable) (Callable, bool)       // true/left; false/right
	OrderMonad    func(...Callable) (Callable, uint)       // n == 0 = false/empty/head; n > 0 = true/len/tail
	EqualMonad    func(...Callable) (Callable, int)        // - n → false/lesser; 0 → undecided/equal; + n → true/greater
	ContinMonad   func(...Callable) (Callable, float64)    // continuous sequence
	ComplxMonad   func(...Callable) (Callable, complex128) // complex plain
	SymbolMonad   func(...Callable) (Callable, string)     // grester/lesser/equal
	BytesMonad    func(...Callable) (Callable, []byte)     // grester/lesser/equal
	FlagsMonad    func(...Callable) (Callable, Typed)      // match-flag/match-flagset
	TimeMonad     func(...Callable) (Callable, time.Time)
	DurationMonad func(...Callable) (Callable, time.Duration)
	RatioMonad    func(...Callable) (Callable, *big.Rat)   // continuous sequence
	BigIntMonad   func(...Callable) (Callable, *big.Int)   // continuous sequence
	BigFloatMonad func(...Callable) (Callable, *big.Float) // continuous sequence

	//// SWITCH EXPRESSION
	///
	// switch expressions branch by returning a particular continuation
	// chosed from a set of continuations passed during switch creation,
	// according to hidden values returned by monadic praedicate
	// expressions passed at switch creation as well. switch returns either
	// contained expression, when called without arguments, or arguments
	// are applyed to the enclosed expression and final return value is
	// composed according to hidden return value yielded by praedicate.
	//
	// results in either another switch expression, in which case all
	// remaining expressions, excluding the current one, are returned in a
	// consumeable as second field of return type pair, or if anything but
	// a SwitchExpr is returned, conumeable is nil & first return value is
	// the result instance boxed according to hidden return value.
	SwitchExpr func(...Callable) (Callable, Consumeable)
)

/// NOOP
func NewNoOp() NoOp                      { return func() {} }
func (n NoOp) Ident() Callable           { return n }
func (n NoOp) Maybe() bool               { return false }
func (n NoOp) Empty() bool               { return true }
func (n NoOp) Eval(...d.Native) d.Native { return nil }
func (n NoOp) Value() Callable           { return nil }
func (n NoOp) Call(...Callable) Callable { return nil }
func (n NoOp) String() string            { return "⊥" }
func (n NoOp) Len() int                  { return 0 }
func (n NoOp) TypeFnc() TyFnc            { return None }
func (n NoOp) TypeNat() d.TyNative       { return d.Nil }
