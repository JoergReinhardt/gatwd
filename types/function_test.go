package types

import (
	"fmt"
	"strings"
	"testing"
)

func TestRecursiveCollection(t *testing.T) {
	strings := strings.Split("this is a public service announcement..."+
		" this is not a test!"+
		" and a hip, hop and'a hippe'd hop...", " ")
	var data = []Data{}
	for _, s := range strings {
		data = append(data, conData(s))
	}
	rec := conRecursive(data...)
	fmt.Printf("recursive method call head: %s\n", rec.Head())
	fmt.Printf("recursive method call tail: %s\n", rec.Tail())
	fmt.Printf("recursive method call Empty: %t\n", rec.Empty())
	fmt.Printf("recursive method call Flag: %s\n", rec.Flag().String())
	h, recu := rec.Head(), rec.Tail()
	for !recu.Empty() {
		//for rec != nil {
		fmt.Printf("recursive collection head: %s\n", h)
		h, recu = recu.Head(), recu.Tail()
	}
}
