package dsltypes

import (
	"fmt"
	"math/bits"
	"strings"
	"sync"
)

var (
	flagRegistryMu sync.RWMutex
	flagNames      = make(map[BitFlag]string)
)

// RegisterTypeFlag registers a human-readable type name for a BitFlag.
func RegisterTypeFlag(flag BitFlag, name string) {
	flagRegistryMu.Lock()
	defer flagRegistryMu.Unlock()
	flagNames[flag] = name
}

// LookupTypeName returns the registered type name for a BitFlag if found.
func LookupTypeName(flag BitFlag) (string, bool) {
	flagRegistryMu.RLock()
	defer flagRegistryMu.RUnlock()
	name, ok := flagNames[flag]
	return name, ok
}

// TypeName returns the registered type name for this BitFlag,
// or falls back to its decomposed string representation.
func (v BitFlag) TypeName() string {
	if name, ok := LookupTypeName(v); ok {
		return name
	}
	return v.String()
}

// BitFlag acts as the core of the dynamic type registry for the DSL.
// It allows representing Algebraic Data Types without relying on Go's
// native static type system (reflect).
type BitFlag uint64

// Flag returns the receiver itself, letting a BitFlag satisfy any interface
// that requires a Flag() BitFlag accessor. It is the identity projection used
// so plain flags and richer flag-bearing types can be treated uniformly.
func (v BitFlag) Flag() BitFlag { return v }

// Len returns the bit length of v: the position of its highest set bit, i.e.
// the smallest number of bits needed to represent the value. It is 0 for the
// empty flag.
func (v BitFlag) Len() int { return bits.Len64(uint64(v)) }

// Count returns the population count of v — the number of set bits — which is
// how many distinct single-bit types a composite flag is composed of.
func (v BitFlag) Count() int { return bits.OnesCount64(uint64(v)) }

// Toggle returns v with the bits in f flipped (exclusive-or), adding each bit
// of f that is absent and removing each bit that is present. It is its own
// inverse: toggling by the same f twice restores v.
func (v BitFlag) Toggle(f BitFlag) BitFlag { return v ^ f }

// Concat returns the union of v and f (bitwise or), combining their type bits
// into one composite flag. This is the monoidal "++" for BitFlag, with the
// empty flag 0 as its identity.
func (v BitFlag) Concat(f BitFlag) BitFlag { return v | f }

// Mask returns v with every bit present in f cleared (bit-clear, v AND NOT f),
// the set-difference that removes f's type bits from v.
func (v BitFlag) Mask(f BitFlag) BitFlag { return v &^ f }

// Match checks if v contains all the bits set in f (or vice versa depending on logic).
// If v is a composite type (multiple bits), and f is a single bit, we check if f is in v.
// In this modernized version, we check if f is a subset of v, or v is a subset of f.
func (v BitFlag) Match(f BitFlag) bool {
	if v == f {
		return true
	}
	// If f has exactly 1 bit, check if it's set in v
	if f.Count() == 1 {
		return (v & f) != 0
	}
	// If v has exactly 1 bit, check if it's set in f
	if v.Count() == 1 {
		return (v & f) != 0
	}
	// Otherwise, check for strict subset/intersection
	return (v&f) == f || (v&f) == v
}

// Decompose breaks down a composite BitFlag into its individual bit components.
func (v BitFlag) Decompose() []BitFlag {
	var slice = []BitFlag{}
	if v.Count() == 1 {
		return append(slice, v)
	}
	for i := 0; i < 64; i++ {
		mask := BitFlag(1 << i)
		if (v & mask) != 0 {
			slice = append(slice, mask)
		}
	}
	return slice
}

// String renders v for display, satisfying the fmt.Stringer interface. A
// registered name is returned when one exists; a single-bit flag becomes
// "Type(n)" where n is its bit index; and a composite flag is decomposed into
// its constituent bits joined by " | " (for example "Int | String"). The empty
// flag renders as "Type(0)".
func (v BitFlag) String() string {
	if name, ok := LookupTypeName(v); ok {
		return name
	}
	if v.Count() == 1 {
		return fmt.Sprintf("Type(%d)", bits.TrailingZeros64(uint64(v)))
	}
	var names []string
	for _, flag := range v.Decompose() {
		if name, ok := LookupTypeName(flag); ok {
			names = append(names, name)
		} else {
			names = append(names, fmt.Sprintf("Type(%d)", bits.TrailingZeros64(uint64(flag))))
		}
	}
	if len(names) == 0 {
		return "Type(0)"
	}
	return strings.Join(names, " | ")
}
