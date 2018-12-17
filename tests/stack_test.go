package main

import (
	"fmt"
	"testing"

	"git.lesara.de/k8xtract/json"
	"git.lesara.de/k8xtract/types"
)

func TestStack(t *testing.T) {
	q, err := json.NewQueue("./testfile.json")
	if err != nil {
		t.Fail()
	}
	s := types.NewStack()

	var n int
	for {
		n = n + 1
		item := (*q).Next()
		if item.Type() == json.ItemError {
			break
		}
		(*s).PushFrame(fmt.Sprintf("func-%d", n), item)
	}
	fmt.Println(s)
	fmt.Println(s.Pop())
	fmt.Println(s.Pop())
	fmt.Println(s.Pop())
	fmt.Println(s.Pop())
	fmt.Println(s)
}
