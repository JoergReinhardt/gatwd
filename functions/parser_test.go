package functions

import (
	"fmt"
	"testing"
)

func TestParseSignature(t *testing.T) {
	var freshmap = tvalm{}
	var tvmap = tvalm{}

	var sig, elems = splitSignature("int → int → int"), NewVector()
	fmt.Printf("test slice: %s\n", sig.String())
	fmt.Printf("test last: %s\n", sig.Last())
	if sig.TypeFnc().Match(Vector) {
		fmt.Printf("sig matched vector")
	}
	tvmap, sig, elems = parseSig(tvmap, sig, elems)
	fmt.Printf("test map: %s, expression slice: %s\n\n", tvmap, elems.String())

	sig = sig.Clear()
	elems = elems.Clear()
	tvmap = freshmap
	sig = splitSignature("a → a → b")
	tvmap, sig, elems = parseSig(tvmap, sig, elems)
	fmt.Printf("test map: %s, expression slice: %s\n\n", tvmap, elems.String())

	sig = sig.Clear()
	elems = elems.Clear()
	tvmap = freshmap
	sig = splitSignature("a → (b → c)")
	tvmap, sig, elems = parseSig(tvmap, sig, elems)
	fmt.Printf("test map: %s, expression slice: %s\n\n", tvmap, elems.String())
}
