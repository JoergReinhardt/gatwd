package functions

import (
	"fmt"
	"strings"
	"testing"

	d "github.com/joergreinhardt/gatwd/data"
)

func TestDataSorter(t *testing.T) {
	var dat = []Callable{
		New("Aaron"),
		New("Aardvark"),
		New("Adam"),
		New("Victor"),
		New("Sylvest"),
		New("Stepen"),
		New("Sonja"),
		New("Tom"),
		New("Britta"),
		New("Peter"),
		New("Paul"),
		New("Mary"),
		New("Eve"),
		New("John"),
		New("Jill"),
	}

	ds := newDataSorter(dat...)
	ds.Sort(d.String)
	fmt.Printf("list after sorterd by string: %s\n", ds)

	idx := ds.Search(New("Sonja"))
	fmt.Printf("Sonja idx: %d\n", idx)
	if idx != 10 {
		fmt.Printf("why fail!?: %d\n", idx)
		t.Fail()
	}
	fmt.Printf("access Sonja by idx: %s\n", ds[idx].Eval().String())
	if strings.Compare(ds[idx].Eval().String(), `"Sonja"`) != 0 {
		fmt.Printf("why fail: access Sonja by idx: %s\n", ds[idx].String())
		t.Fail()
	}

	fdx := ds.Search(New("NotAName"))
	fmt.Printf("unfindable index supposed to be -1: %d\n", fdx)
	if fdx != -1 {
		fmt.Printf("why fail: unfindable index supposed to be -1: %d\n", fdx)
		t.Fail()
	}

}
func TestDataSorterFlags(t *testing.T) {
	var flags = []Callable{
		New(d.Nil),
		New(d.Bool),
		New(d.Int),
		New(d.Int8),
		New(d.Int16),
		New(d.Int32),
		New(d.BigInt),
	}

	fs := dataSorter(flags)
	fmt.Printf("unsorted flags: %s\n", fs)
	fs.Sort(d.Flag)
	fmt.Printf("sorted flags: %s\n", fs)

	var ints = []Callable{
		New(int(11)),
		New(int(-12)),
		New(int(12321)),
		New(int(543)),
		New(int(8493)),
		New(int(-134)),
		New(int(381)),
	}

	is := dataSorter(ints)
	is.Sort(d.Integers)
	fmt.Printf("sorted ints: %s\n", is)
}
func TestDataSorterMixedType(t *testing.T) {

	// TODO: make this work
	var flags = []Callable{
		New(int(11)),
		New(uint(134)),
		New("Peter"),
		New(int(-12)),
		New("Paul"),
		New(uint(12321)),
		New(int(12321)),
		New("Eve"),
		New(uint(543)),
		New(int(543)),
		New(uint(12)),
		New(int(8493)),
		New("John"),
		New(uint(8493)),
		New(int(-134)),
		New(uint(381)),
		New("Jill"),
		New(int(381)),
		New(uint(11)),
		New("Mary"),
	}

	ts := dataSorter(flags)
	ts.Sort(d.Flag)
	fmt.Printf("supposedly sorted by flag: %s\n", ts)
}
func TestPairSorterStrStr(t *testing.T) {
	var strPairs = []PairVal{
		NewPair(New("Aaron"), New("val 0")),
		NewPair(New("Aardvark"), New("val 1")),
		NewPair(New("Adam"), New("val 2")),
		NewPair(New("Victor"), New("val 3")),
		NewPair(New("Sylvest"), New("val 4")),
		NewPair(New("Stepen"), New("val 5")),
		NewPair(New("Sonja"), New("val 6")),
		NewPair(New("Tom"), New("val 7")),
		NewPair(New("Britta"), New("val 8")),
		NewPair(New("Peter"), New("val 9")),
		NewPair(New("Paul"), New("val 10")),
		NewPair(New("Mary"), New("val 11")),
		NewPair(New("Eve"), New("val 12")),
		NewPair(New("John"), New("val 13")),
		NewPair(New("Jill"), New("val 14")),
	}

	ps := newPairSorter(strPairs...)
	fmt.Printf("unsorted string|string slice:\n %s\n\n", ps)
	ps.Sort(d.Letters)
	fmt.Printf("sorted string|string slice (sorted alphabeticly by key!) :\n %s\n\n", ps)
}
func TestPairSorterIntStr(t *testing.T) {
	var pairs = []PairVal{
		NewPair(New(10), New("valeu ten")),
		NewPair(New(13), New("valeu thirteen")),
		NewPair(New(7), New("valeu seven")),
		NewPair(New(8), New("valeu eight")),
		NewPair(New(1), New("valeu one")),
		NewPair(New(2), New("valeu two")),
		NewPair(New(3), New("valeu three")),
		NewPair(New(4), New("valeu four")),
		NewPair(New(5), New("valeu five")),
		NewPair(New(6), New("valeu six")),
	}

	ps := newPairSorter(pairs...)
	ps.Sort(d.Integers)
	fmt.Printf("pairs sorted by int key:\n%s\n\n", ps)
}
func TestPairSorterUintStr(t *testing.T) {
	var pairs = []PairVal{
		NewPair(New(uint(10)), New("valeu ten")),
		NewPair(New(uint(13)), New("valeu thirteen")),
		NewPair(New(uint(7)), New("valeu seven")),
		NewPair(New(uint(8)), New("valeu eight")),
		NewPair(New(uint(1)), New("valeu one")),
		NewPair(New(uint(2)), New("valeu two")),
		NewPair(New(uint(3)), New("valeu three")),
		NewPair(New(uint(4)), New("valeu four")),
		NewPair(New(uint(5)), New("valeu five")),
		NewPair(New(uint(6)), New("valeu six")),
	}

	ps := newPairSorter(pairs...)
	ps.Sort(d.Naturals)
	fmt.Printf("pairs sorted by uint key:\n%s\n\n", ps)
}
func TestPairSorterIrrationalStr(t *testing.T) {
	var pairs = []PairVal{
		NewPair(New(float64(10.21)), New("valeu ten")),
		NewPair(New(float64(13.23)), New("valeu thirteen")),
		NewPair(New(float64(7.72323)), New("valeu seven")),
		NewPair(New(float64(8.342)), New("valeu eight")),
		NewPair(New(float64(1.234)), New("valeu one")),
		NewPair(New(float64(2.25)), New("valeu two")),
		NewPair(New(float64(3.3333)), New("valeu three")),
		NewPair(New(float64(4)), New("valeu four")),
		NewPair(New(float64(5)), New("valeu five")),
		NewPair(New(float64(6)), New("valeu six")),
	}

	ps := newPairSorter(pairs...)
	ps.Sort(d.Reals)
	fmt.Printf("pairs sorted by float key:\n%s\n\n", ps)
}
