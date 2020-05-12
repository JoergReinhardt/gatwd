package gatw

import (
	"fmt"
	"testing"
)

func TestCategory(t *testing.T) {
	cons := initCat()
	fmt.Println(cons)

	var elem Elem
	elem, cons = cons()
	fmt.Println(elem)

	elem, cons = cons(N)
	fmt.Println(elem)

	elem, cons = cons(Type, Func, Symb)
	fmt.Println(elem)
}
