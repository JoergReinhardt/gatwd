package functions_test

import (
	"testing"

	"github.com/joergreinhardt/gatwd/functions"
)

func TestBytesStringAndCopy(t *testing.T) {
	b := functions.Bytes("hi")
	if b.String() != "0x6869" || functions.Bytes(nil).String() != "0x" {
		t.Errorf("String() = %q, %q", b.String(), functions.Bytes(nil).String())
	}
	cp, ok := functions.AsBytes(b)
	if !ok {
		t.Fatal("AsBytes(Bytes) must succeed")
	}
	cp[0] = 'X'
	if b[0] != 'h' {
		t.Error("AsBytes must return a copy, not alias the argument")
	}
	if plain, ok := functions.AsBytes([]byte{1, 2}); !ok || len(plain) != 2 {
		t.Error("a plain []byte from an adapter is accepted")
	}
	for _, v := range []any{"text", 5, nil, []any{1}} {
		if _, ok := functions.AsBytes(v); ok {
			t.Errorf("AsBytes(%#v) must be false: a String is not silently Bytes", v)
		}
	}
	if functions.KindName(b) != "Bytes" || functions.KindName([]byte{1}) != "Bytes" {
		t.Error("KindName must say Bytes")
	}
}

func TestByteOf(t *testing.T) {
	for _, n := range []any{0, 255, 7} {
		if _, err := functions.ByteOf(n); err != nil {
			t.Errorf("ByteOf(%v): %v", n, err)
		}
	}
	for _, bad := range []any{-1, 256, 300, "a", functions.Char('a'), 1.5, true, nil} {
		if _, err := functions.ByteOf(bad); err == nil {
			t.Errorf("ByteOf(%#v) must fail", bad)
		}
	}
}

func TestBytesConcatAndEquality(t *testing.T) {
	a, b := functions.Bytes("ab"), functions.Bytes("cd")
	got, err := functions.EvalBinaryOp(functions.OpConcat, a, b)
	if err != nil {
		t.Fatal(err)
	}
	out, ok := got.(functions.Bytes)
	if !ok || string(out) != "abcd" {
		t.Fatalf("Bytes ++ Bytes = %#v, want Bytes abcd", got)
	}
	out[0] = 'Z'
	if a[0] != 'a' {
		t.Error("++ must not alias its operand")
	}
	if eq, _ := functions.EvalBinaryOp(functions.OpEq, a, functions.Bytes("ab")); eq != true {
		t.Error("equal Bytes must be ==")
	}
	if eq, _ := functions.EvalBinaryOp(functions.OpEq, a, b); eq != false {
		t.Error("different Bytes must not be ==")
	}
}
