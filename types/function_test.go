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
	fmt.Printf("recursive stringer: %s\n", rec.String())
	h, rec := rec()
	for !rec.Empty() {
		//for rec != nil {
		fmt.Printf("recursive collection head: %s\n", h)
		h, rec = rec()
	}
}
