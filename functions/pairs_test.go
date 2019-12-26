package functions

import (
	"fmt"
	"testing"
)

func TestPairVal(t *testing.T) {
	var pair = NewPair(NewNone(), NewNone())
	fmt.Printf("name of empty pair: %s\n", pair.Type())

	pair = NewPair(Dat(12), Dat("string"))
	fmt.Printf("name of (int,string) pair: %s\n",
		pair.Type())
	fmt.Printf("name of (int,string) pair args: %s\n",
		pair.Type().TypeArgs())
	fmt.Printf("name of (int,string) pair return: %s\n",
		pair.Type().TypeRet())
}
