package functions

import (
	"fmt"
	"testing"

	d "github.com/JoergReinhardt/godeep/data"
)

func TestDataSorter(t *testing.T) {
	var dat = []d.Data{
		d.New("Aaron"),
		d.New("Aardvark"),
		d.New("Adam"),
		d.New("Victor"),
		d.New("Sylvest"),
		d.New("Stepen"),
		d.New("Sonja"),
		d.New("Tom"),
		d.New("Britta"),
		d.New("Peter"),
		d.New("Paul"),
		d.New("Mary"),
		d.New("Eve"),
		d.New("John"),
		d.New("Jill"),
	}

	ds := newDataSorter(dat...)
	ds.Sort(d.String)

	idx := ds.Search(d.New("Sonja"))
	if idx != 10 {
		t.Fail()
	}
	fmt.Println(idx)
	if ds[idx].String() != "Sonja" {
		t.Fail()
	}

	fdx := ds.Search(d.New("NotAName"))
	fmt.Printf("unfindable index supposed to be -1: %d\n", fdx)
	if fdx != -1 {
		t.Fail()
	}

	var flags = []d.Data{
		d.New(d.Nil),
		d.New(d.Bool),
		d.New(d.Int),
		d.New(d.Int8),
		d.New(d.Int16),
		d.New(d.Int32),
		d.New(d.BigInt),
	}

	fs := dataSorter(flags)
	fmt.Printf("unsorted flags: %s\n", fs)
	fs.Sort(d.Flag)
	fmt.Printf("sorted flags: %s\n", fs)

	var ints = []d.Data{
		d.New(int(11)),
		d.New(int(-12)),
		d.New(int(12321)),
		d.New(int(543)),
		d.New(int(8493)),
		d.New(int(-134)),
		d.New(int(381)),
	}

	is := dataSorter(ints)
	is.Sort(d.Integer)
	fmt.Printf("sorted ints: %s\n", is)
}
func TestDataSorterMixedType(t *testing.T) {

	// TODO: make this work
	var flags = []d.Data{
		d.New(int(11)),
		d.New(uint(134)),
		d.New("Peter"),
		d.New(int(-12)),
		d.New("Paul"),
		d.New(uint(12321)),
		d.New(int(12321)),
		d.New("Eve"),
		d.New(uint(543)),
		d.New(int(543)),
		d.New(uint(12)),
		d.New(int(8493)),
		d.New("John"),
		d.New(uint(8493)),
		d.New(int(-134)),
		d.New(uint(381)),
		d.New("Jill"),
		d.New(int(381)),
		d.New(uint(11)),
		d.New("Mary"),
	}

	ts := dataSorter(flags)
	ts.Sort(d.Flag)
	fmt.Printf("supposedly sorted by flag: %s\n", ts)
}
func TestPairSorterStrStr(t *testing.T) {
	var strPairs = []Paired{
		NewPair(d.New("Aaron"), d.New("val 0")),
		NewPair(d.New("Aardvark"), d.New("val 1")),
		NewPair(d.New("Adam"), d.New("val 2")),
		NewPair(d.New("Victor"), d.New("val 3")),
		NewPair(d.New("Sylvest"), d.New("val 4")),
		NewPair(d.New("Stepen"), d.New("val 5")),
		NewPair(d.New("Sonja"), d.New("val 6")),
		NewPair(d.New("Tom"), d.New("val 7")),
		NewPair(d.New("Britta"), d.New("val 8")),
		NewPair(d.New("Peter"), d.New("val 9")),
		NewPair(d.New("Paul"), d.New("val 10")),
		NewPair(d.New("Mary"), d.New("val 11")),
		NewPair(d.New("Eve"), d.New("val 12")),
		NewPair(d.New("John"), d.New("val 13")),
		NewPair(d.New("Jill"), d.New("val 14")),
	}

	ps := newPairSorter(strPairs...)
	fmt.Printf("unsorted string|string slice:\n %s\n\n", ps)
	ps.Sort(d.Symbolic)
	fmt.Printf("sorted string|string slice (sorted alphabeticly by key!) :\n %s\n\n", ps)
}
func TestPairSorterIntStr(t *testing.T) {
	var pairs = []Paired{
		NewPair(d.New(10), d.New("valeu ten")),
		NewPair(d.New(13), d.New("valeu thirteen")),
		NewPair(d.New(7), d.New("valeu seven")),
		NewPair(d.New(8), d.New("valeu eight")),
		NewPair(d.New(1), d.New("valeu one")),
		NewPair(d.New(2), d.New("valeu two")),
		NewPair(d.New(3), d.New("valeu three")),
		NewPair(d.New(4), d.New("valeu four")),
		NewPair(d.New(5), d.New("valeu five")),
		NewPair(d.New(6), d.New("valeu six")),
	}

	ps := newPairSorter(pairs...)
	ps.Sort(d.Integer)
	fmt.Printf("pairs sorted by int key:\n%s\n\n", ps)
}
func TestPairSorterUintStr(t *testing.T) {
	var pairs = []Paired{
		NewPair(d.New(uint(10)), d.New("valeu ten")),
		NewPair(d.New(uint(13)), d.New("valeu thirteen")),
		NewPair(d.New(uint(7)), d.New("valeu seven")),
		NewPair(d.New(uint(8)), d.New("valeu eight")),
		NewPair(d.New(uint(1)), d.New("valeu one")),
		NewPair(d.New(uint(2)), d.New("valeu two")),
		NewPair(d.New(uint(3)), d.New("valeu three")),
		NewPair(d.New(uint(4)), d.New("valeu four")),
		NewPair(d.New(uint(5)), d.New("valeu five")),
		NewPair(d.New(uint(6)), d.New("valeu six")),
	}

	ps := newPairSorter(pairs...)
	ps.Sort(d.Unsigned)
	fmt.Printf("pairs sorted by uint key:\n%s\n\n", ps)
}
func TestPairSorterIrrationalStr(t *testing.T) {
	var pairs = []Paired{
		NewPair(d.New(float64(10.21)), d.New("valeu ten")),
		NewPair(d.New(float64(13.23)), d.New("valeu thirteen")),
		NewPair(d.New(float64(7.72323)), d.New("valeu seven")),
		NewPair(d.New(float64(8.342)), d.New("valeu eight")),
		NewPair(d.New(float64(1.234)), d.New("valeu one")),
		NewPair(d.New(float64(2.25)), d.New("valeu two")),
		NewPair(d.New(float64(3.3333)), d.New("valeu three")),
		NewPair(d.New(float64(4)), d.New("valeu four")),
		NewPair(d.New(float64(5)), d.New("valeu five")),
		NewPair(d.New(float64(6)), d.New("valeu six")),
	}

	ps := newPairSorter(pairs...)
	ps.Sort(d.Irrational)
	fmt.Printf("pairs sorted by float key:\n%s\n\n", ps)
}
