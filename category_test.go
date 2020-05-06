package main

import (
	"fmt"
	"testing"
)

func TestFlag(t *testing.T) {
	var i = rank(Flg(Truth))
	fmt.Printf("rank of 'Truth': %d\n", i)
	i = rank(Flg(Uint))
	fmt.Printf("rank of 'Uint': %d\n", i)
	i = rank(Flg(List))
	fmt.Printf("rank of 'List': %d\n", i)
}
