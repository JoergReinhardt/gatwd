package functions

import (
	"fmt"
	"testing"
)

var intslice = []Callable{
	New(0), New(1), New(2), New(3), New(4), New(5), New(6), New(7), New(8), New(9), New(0), New(134), New(8566735),
	New(4534), New(3445), New(76575), New(2234), New(45), New(7646), New(64), New(3), New(314),
}

var intkeys = []Callable{New("zero"), New("one"), New("two"), New("three"),
	New("four"), New("five"), New("six"), New("seven"), New("eight"), New("nine"),
	New("ten"), New("eleven"), New("twelve"), New("thirteen"), New("fourteen"),
	New("fifteen"), New("sixteen"), New("seventeen"), New("eighteen"), New("nineteen"), New("twenty"),
	New("twentyone"),
}

func TestSignature(t *testing.T) {
	var zipped = ZipF(NewVector(intkeys...), NewVector(intslice...), func(l, r Callable) Paired { return NewPair(l, r) })
	var vec = NewVector()
	var head, tail = zipped.DeCap()
	for head != nil {
		vec = vec.Cons(head)
		fmt.Printf("head: %s\n", head)
		head, tail = tail.DeCap()
	}
	fmt.Printf("vector: %s\n", vec())
	//	var sig = NewSignature(vec()...)
	//	fmt.Printf("signature: %s\n", sig())
}
