package functions

import (
	"fmt"
	"testing"

	d "github.com/JoergReinhardt/godeep/data"
)

func TestDataSorter(t *testing.T) {
	var dat = []Data{
		newData(d.New("Aaron")),
		newData(d.New("Aardvark")),
		newData(d.New("Adam")),
		newData(d.New("Victor")),
		newData(d.New("Sylvest")),
		newData(d.New("Stepen")),
		newData(d.New("Sonja")),
		newData(d.New("Tom")),
		newData(d.New("Britta")),
		newData(d.New("Peter")),
		newData(d.New("Paul")),
		newData(d.New("Mary")),
		newData(d.New("Eve")),
		newData(d.New("John")),
		newData(d.New("Jill")),
	}

	ds := dataSorter(dat)
	ds.Sort(d.String)

	fmt.Printf("names sorted by string: %s\n", ds)
	idx := ds.Search(newData(d.New("Sonja")))
	fmt.Printf("name found via search: %s\n", ds[idx].String())
	if idx != 10 {
		t.Fail()
	}
	if ds[idx].String() != "Sonja" {
		t.Fail()
	}

	fdx := ds.Search(newData(d.New("NotAName")))
	fmt.Printf("unfindable index supposed to be -1: %d\n", fdx)
	if fdx != -1 {
		t.Fail()
	}

	var flags = []Data{
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

	var ints = []Data{
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
	var typs = []Data{
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

	ts := dataSorter(typs)
	ts.Sort(d.Flag)
	fmt.Printf("supposedly sorted by flag: %s\n", ts)

func TestPairSorterStrStr(t *testing.T) {
	var strPairs = []Paired{
		newPair(d.New("Aaron"), d.New("val 0")),
		newPair(d.New("Aardvark"), d.New("val 1")),
		newPair(d.New("Adam"), d.New("val 2")),
		newPair(d.New("Victor"), d.New("val 3")),
		newPair(d.New("Sylvest"), d.New("val 4")),
		newPair(d.New("Stepen"), d.New("val 5")),
		newPair(d.New("Sonja"), d.New("val 6")),
		newPair(d.New("Tom"), d.New("val 7")),
		newPair(d.New("Britta"), d.New("val 8")),
		newPair(d.New("Peter"), d.New("val 9")),
		newPair(d.New("Paul"), d.New("val 10")),
		newPair(d.New("Mary"), d.New("val 11")),
		newPair(d.New("Eve"), d.New("val 12")),
		newPair(d.New("John"), d.New("val 13")),
		newPair(d.New("Jill"), d.New("val 14")),
	}

	ps := newPairSorter(strPairs...)
	fmt.Printf("unsorted string|string slice:\n %s\n\n", ps)
	ps.Sort(d.Symbolic)
	fmt.Printf("sorted string|string slice (sorted alphabeticly by key!) :\n %s\n\n", ps)
}
func TestPairSorterIntStr(t *testing.T) {
	var pairs = []Paired{
		newPair(d.New(10), d.New("valeu ten")),
		newPair(d.New(13), d.New("valeu thirteen")),
		newPair(d.New(7), d.New("valeu seven")),
		newPair(d.New(8), d.New("valeu eight")),
		newPair(d.New(1), d.New("valeu one")),
		newPair(d.New(2), d.New("valeu two")),
		newPair(d.New(3), d.New("valeu three")),
		newPair(d.New(4), d.New("valeu four")),
		newPair(d.New(5), d.New("valeu five")),
		newPair(d.New(6), d.New("valeu six")),
	}

	ps := newPairSorter(pairs...)
	ps.Sort(d.Integer)
	fmt.Printf("pairs sorted by int key:\n%s\n\n", ps)
}
func TestPairSorterUintStr(t *testing.T) {
	var pairs = []Paired{
		newPair(d.New(uint(10)), d.New("valeu ten")),
		newPair(d.New(uint(13)), d.New("valeu thirteen")),
		newPair(d.New(uint(7)), d.New("valeu seven")),
		newPair(d.New(uint(8)), d.New("valeu eight")),
		newPair(d.New(uint(1)), d.New("valeu one")),
		newPair(d.New(uint(2)), d.New("valeu two")),
		newPair(d.New(uint(3)), d.New("valeu three")),
		newPair(d.New(uint(4)), d.New("valeu four")),
		newPair(d.New(uint(5)), d.New("valeu five")),
		newPair(d.New(uint(6)), d.New("valeu six")),
	}

	ps := newPairSorter(pairs...)
	ps.Sort(d.Unsigned)
	fmt.Printf("pairs sorted by uint key:\n%s\n\n", ps)
}
func TestPairSorterIrrationalStr(t *testing.T) {
	var pairs = []Paired{
		newPair(d.New(float64(10.21)), d.New("valeu ten")),
		newPair(d.New(float64(13.23)), d.New("valeu thirteen")),
		newPair(d.New(float64(7.72323)), d.New("valeu seven")),
		newPair(d.New(float64(8.342)), d.New("valeu eight")),
		newPair(d.New(float64(1.234)), d.New("valeu one")),
		newPair(d.New(float64(2.25)), d.New("valeu two")),
		newPair(d.New(float64(3.3333)), d.New("valeu three")),
		newPair(d.New(float64(4)), d.New("valeu four")),
		newPair(d.New(float64(5)), d.New("valeu five")),
		newPair(d.New(float64(6)), d.New("valeu six")),
	}

	ps := newPairSorter(pairs...)
	ps.Sort(d.Irrational)
	fmt.Printf("pairs sorted by float key:\n%s\n\n", ps)
}
