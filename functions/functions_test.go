package functions

import (
	"fmt"
	d "github.com/JoergReinhardt/godeep/data"
	"testing"
)

func TestIdGenerator(t *testing.T) {
	var id int
	id, initSig = initSig()
	fmt.Println(id)
	id, initSig = initSig()
	fmt.Println(id)
	if id != 1 {
		t.Fail()
	}
	id, initSig = initSig()
	id, initSig = initSig()
	id, initSig = initSig()
	id, initSig = initSig()
	id, initSig = initSig()
	fmt.Println(id)
	if id != 6 {
		t.Fail()
	}
}
func TestTokenSlice(t *testing.T) {
	ts1 := [][]token{
		[]token{token{Argument.Flag()}},
		[]token{token{Parameter.Flag()}},
		[]token{token{Return.Flag()}},
		[]token{token{Argument.Flag()}},
		[]token{token{Return.Flag()}},
		[]token{token{Return.Flag()}},
		[]token{token{Argument.Flag()}},
	}
	ts2 := [][]token{
		[]token{token{d.Bool.Flag()}},
		[]token{token{Parameter.Flag()}},
		[]token{token{Return.Flag()}},
		[]token{token{Argument.Flag()}},
		[]token{token{Return.Flag()}},
		[]token{token{Return.Flag()}},
		[]token{token{Argument.Flag()}},
	}

	sortSlice(ts1)
	sortSlice(ts2)
	tss1 := byToken(ts1, token{Argument.Flag()})
	tss2 := byToken(ts2, token{Parameter.Flag()})
	tss3 := byToken(ts2, token{Return.Flag()})
	tss4 := byToken(tss2, token{Return.Flag()})
	fmt.Println(tss1)
	fmt.Println(tss2)
	fmt.Println(tss3)
	fmt.Println(tss4)
}
