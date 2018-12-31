package types

import (
	"fmt"
	"testing"
)

var fn = func(d ...Data) Data { return conData(string(d[0].(strVal)) + string(d[1].(strVal))) }

func TestPartial(t *testing.T) {
	s1 := partial(fn, conData("one"))
	s2 := s1(conData(" two"))
	fmt.Printf("final result: %s\n", s2)
	if fmt.Sprintf("%s", s2) != "one two" {
		t.Fail()
	}
}
func TestReverseArgs(t *testing.T) {
	fnr := reverseArgs(fn)
	s3 := fnr(conData(" one").(strVal), conData("two").(strVal))
	fmt.Printf("final result: %s\n", s3)
	if fmt.Sprintf("%s", s3) != "two one" {
		t.Fail()
	}
}
func TestCount(t *testing.T) {

	c := conCount()
	fmt.Printf("count initial: %v\n", c)
	var i int

	for u := 0; u < 100; u++ {
		i, c = c()
		fmt.Printf(": %v\n", i)
	}
	if i != 99 {
		t.Fail()
	}

	c = conCount(33, 7)
	for u := 0; u < 100; u++ {
		i, c = c()
		fmt.Printf(": %v\n", i)
	}
	if i != 693 {
		t.Fail()
	}
}

var camaeleon = conData("comma", "comma", "comma", "comma", "comma", "k", 8, "maeleon")

func TestAcceptTypes(t *testing.T) {
	fmt.Println(camaeleon)
	fmt.Printf("testflag: %s\n", camaeleon.(chain)[2].Flag())
	fmt.Println(acceptDataTypes(String.Flag(), camaeleon.(chain)...))

	for _, v := range acceptDataTypes(String.Flag(), camaeleon.(chain)...) {
		if !fmatch(v.Flag(), String.Flag()) {
			t.Fail()
		}
	}
}

func TestArityGuard(t *testing.T) {
	camstring := arity(5, stringer)
	fmt.Printf("the lesser of the camelii: %s\n", binary(stringer)(camaeleon.(chain)...))
	if !(camstring(camaeleon.(chain)...).String() == "commacommacommacommacomma") {
		t.Fail()

	}
}
func testFn(d ...Data) Data {
	return strVal(fmt.Sprintf("%s", d))
}

func TestCurry(t *testing.T) {
	ca := curry(testFn, 4)
	//fmt.Printf("freshly curryd: %s\n", ca)
	one := ca(conData("one")).(fnc)
	two := one(conData("two")).(fnc)
	three := two(conData("three")).(fnc)
	str := three(conData("four")).(strVal)
	fmt.Printf("first parameter applyed curryd: %v\n", str)
}
