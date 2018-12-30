package types

import (
	"fmt"
	"strings"
	"testing"
)

var str = strings.Split("this is a public service announcement..."+
	" this is not a test!"+
	" and a hip, hop and'a hippe'd hop...", " ")
var data = func() []Data {
	var data = []Data{}
	for _, d := range str {
		data = append(data, conData(d))
	}
	return data
}()

func TestLazyCollection(t *testing.T) {

	rec := conRecursiveLazy(data...)

	fmt.Printf("recursive lazy method call head: %s\n", rec.Head())
	fmt.Printf("recursive lazy method call tail: %s\n", rec.Tail())
	fmt.Printf("recursive lazy method call Empty: %t\n", rec.Empty())
	fmt.Printf("recursive lazy method call Flag: %s\n", rec.Flag())
	h, recu := rec.Head(), rec.Tail()
	for !recu.Empty() {
		//for rec != nil {
		h, recu = recu.Head(), recu.Tail()
		fmt.Printf("recursive lazy collection head: %s\n", h)
	}
}
func TestEagerRecursiveCollection(t *testing.T) {

	rec := conRecursiveEager(data...)

	fmt.Printf("recursive eager method call head: %s\n", rec.Head())
	fmt.Printf("recursive eager method call tail: %s\n", rec.Tail())
	fmt.Printf("recursive eager method call Empty: %t\n", rec.Empty())
	fmt.Printf("recursive eager method call Flag: %s\n", rec.Flag())
	h, recu := rec.Head(), rec.Tail()
	for !recu.Empty() {
		//for rec != nil {
		h, recu = recu.Head(), recu.Tail()
		fmt.Printf("recursive eager collection head: %s\n", h)
	}
}
func TestEagerFLatCollection(t *testing.T) {
	flat := conFlatColEager(data...)

	fmt.Printf("flat eager method call head: %s\n", flat.Head())
	fmt.Printf("flat eager method call tail: %s\n", flat.Tail())
	fmt.Printf("flat eager method call Empty: %t\n", flat.Empty())
	fmt.Printf("flat eager method call Flag: %s\n", flat.Flag())

	flat = conFlatColEager(flat.Slice()...)
	fmt.Printf("flat eager method call after decap head: %s\n", flat.Head())
	fmt.Printf("flat eager method call after decap tail: %s\n", flat.Tail())

	tup := conTuple(data...)

	fmt.Printf("flat eager method call head: %s\n", tup.Head())
	fmt.Printf("flat eager method call tail: %s\n", tup.Tail())
	fmt.Printf("flat eager method call Empty: %t\n", tup.Empty())
	fmt.Printf("flat eager method call Flag: %s\n", tup.Flag())

}
func TestNestCollectionEager(t *testing.T) {

	nest := conNestColEager(data...)

	fmt.Printf("nested col eager method call head: %s\n", nest.Head())
	fmt.Printf("nested col eager method call tail: %s\n", nest.Tail())
	fmt.Printf("nested col eager method call Empty: %t\n", nest.Empty())
	fmt.Printf("nested col eager method call Flag: %s\n", nest.Flag())

	nest = conNestColLazy(data...)()

	fmt.Printf("nested col lazy method call head: %s\n", nest.Head())
	fmt.Printf("nested col lazy method call tail: %s\n", nest.Tail())
	fmt.Printf("nested col lazy method call Empty: %t\n", nest.Empty())
	fmt.Printf("nested col lazy method call Flag: %s\n", nest.Flag())
}
