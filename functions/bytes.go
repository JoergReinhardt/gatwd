package functions

import (
	"encoding/hex"
	"fmt"
)

// Bytes is an immutable sequence of bytes: the value type of binary data in the DSL.
//
// # The concept
//
// Text (String) is a sequence of Unicode characters; binary data (a decoded Secret, a hash, a
// network payload) is a sequence of bytes that need not be valid text. Keeping them apart
// means a function that wants text never receives bytes by accident, and the conversion
// that can fail — bytes to text — is visible and partial (`utf8_decode` is a Maybe).
//
// # Why a named type
//
// Bytes is `type Bytes []byte`, distinct from a plain []byte, so that it can print as hex
// everywhere (`0x6869`) — a nested Bytes inside a Maybe or a list does not turn into a
// list of numbers — and so that a type switch can tell it from other slices. Every function
// treats it as immutable and copies before returning.
//
// # Example
//
//	b := functions.Bytes("hi")
//	b.String() // "0x6869"
type Bytes []byte

// String renders the bytes as 0x-prefixed lowercase hex (0x for empty), which is safe to
// print whatever the bytes hold.
func (b Bytes) String() string { return "0x" + hex.EncodeToString(b) }

// AsBytes reads v as Bytes: a Bytes, or a plain []byte (as the I/O adapters yield). It does
// not convert a String; that is the caller's explicit choice. The result is a copy.
func AsBytes(v any) (Bytes, bool) {
	switch b := v.(type) {
	case Bytes:
		return append(Bytes(nil), b...), true
	case []byte:
		return append(Bytes(nil), b...), true
	}
	return nil, false
}

// IsBytes reports whether v is a Bytes or a plain []byte.
func IsBytes(v any) bool {
	_, ok := AsBytes(v)
	return ok
}

// ByteOf reads an Int 0..255 as a byte; anything else is an error naming the value.
func ByteOf(v any) (byte, error) {
	switch n := v.(type) {
	case string, Char, bool, float32, float64, complex64, complex128:
		return 0, fmt.Errorf("%v (%s) is not a byte; want an Int 0..255", v, KindName(v))
	default:
		i, ok := toInt64(n)
		if !ok || i < 0 || i > 255 {
			return 0, fmt.Errorf("%v is not a byte; want an Int 0..255", v)
		}
		return byte(i), nil
	}
}
