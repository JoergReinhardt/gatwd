package run

import (
	"fmt"
	"testing"
)

func TestArity(t *testing.T) {
	fmt.Println(Arity(0))
	if Arity(0).String() != "Nullary" {
		t.Fail()
	}
	fmt.Println(Arity(5))
	if Arity(5).String() != "Quinary" {
		t.Fail()
	}
	fmt.Println(Effected)
}
