package functions

import (
	"fmt"
	d "github.com/JoergReinhardt/gatwd/data"
	"testing"
)

var praed = NewPraedicate(func(scrut Value) bool {
	if scrut.TypeNat().Flag().Match(d.Int) {
		if scrut.Eval().(d.IntVal).Int() > 0 {
			return true
		}
	}
	return false
})

func TestPraedicate(t *testing.T) {

	fmt.Println(praed)

	fmt.Println(praed(New(2)))
	if !praed(New(2)) {
		t.Fail()
	}
	fmt.Println(praed(New(-2)))
	if praed(New(-2)) {
		t.Fail()
	}
}
