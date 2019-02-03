package lex

import (
	"fmt"
	"testing"
)

func TestMatchString(t *testing.T) {
	if item, ok := Match("XOR"); ok {
		fmt.Println(item.Syntax())
	}
	if item, ok := Match("::"); ok {
		fmt.Println(item.Syntax())
	}
	if item, ok := Match(":"); ok {
		fmt.Println(item.Syntax())
	}
	if item, ok := Match(`\F`); ok {
		fmt.Println(item.Syntax())
	}
	if item, ok := Match(`\f`); ok {
		fmt.Println(item.Syntax())
	}
	if item, ok := Match(`\x`); ok {
		fmt.Println(item.Syntax())
	}
	fmt.Println(Match("a"))
}
