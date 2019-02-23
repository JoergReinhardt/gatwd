package lex

import (
	"fmt"
	"testing"
)

func TestMatchString(t *testing.T) {
	if item, ok := MatchItem("XOR"); ok {
		fmt.Println(item.Syntax())
	}
	if item, ok := MatchItem("::"); ok {
		fmt.Println(item.Syntax())
	}
	if item, ok := MatchItem(":"); ok {
		fmt.Println(item.Syntax())
	}
	if item, ok := MatchItem(`\F`); ok {
		fmt.Println(item.Syntax())
	}
	if item, ok := MatchItem(`\f`); ok {
		fmt.Println(item.Syntax())
	}
	if item, ok := MatchItem(`\x`); ok {
		fmt.Println(item.Syntax())
	}
	fmt.Println(MatchItem("a"))
}
func TestSyntaxMatchingAscii(t *testing.T) {
	fmt.Println(Digraphs())
	fmt.Println(AllSyntax())
}
