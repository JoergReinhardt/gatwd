package main

import (
	"fmt"
	"testing"

	"git.lesara.de/k8xtract/json"
	"git.lesara.de/k8xtract/types"
)

func TestList(t *testing.T) {
	r := types.NewTree("test-tree")
	l := types.NewList()
	(*l).Add(json.NewStringNode(types.Idx(0), "str-1", "str-1-value", r))
	(*l).Add(json.NewStringNode(types.Idx(1), "str-2", "str-2-value", r))
	(*l).Add(json.NewStringNode(types.Idx(2), "str-3", "str-3-value", r))
	(*l).Add(json.NewStringNode(types.Idx(3), "str-4", "str-4-value", r))
	(*l).Push(json.NewStringNode(types.Idx(4), "str-5", "str-5-value", r))
	(*l).Push(json.NewStringNode(types.Idx(5), "str-6", "str-6-value", r))
	(*l).Push(json.NewStringNode(types.Idx(6), "str-7", "str-7-value", r))
	(*l).Push(json.NewStringNode(types.Idx(7), "str-8", "str-8-value", r))
	fmt.Println(l)
	fmt.Println((*l).Pop())
	fmt.Println((*l).Pop())
	fmt.Println((*l).Pop())
	fmt.Println(l)
	fmt.Println((*l).Pull())
	fmt.Println((*l).Pull())
	fmt.Println((*l).Pull())
	fmt.Println(l)
}
