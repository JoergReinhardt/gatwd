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

	fmt.Println(ds)
	fmt.Println(ds.Search(newData(d.New("Sonja"))))
	fmt.Println(ds.Get(newData(d.New("Sonja"))))
	fmt.Println(ds.Get(newData(d.New("Peter"))))

	var flags = []Data{
		newData(d.New(d.Nil)),
		newData(d.New(d.Bool)),
		newData(d.New(d.Int)),
		newData(d.New(d.Int8)),
		newData(d.New(d.Int16)),
		newData(d.New(d.Int32)),
		newData(d.New(d.BigInt)),
	}

	fs := dataSorter(flags)
	fmt.Println(fs)
	fs.Sort(d.Flag)
	fmt.Println(fs)

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
	fmt.Println(is)

	// TODO: make this work
	var typs = []Data{
		d.New(int(11)),
		d.New(uint(134)),
		newData(d.New("Peter")),
		d.New(int(-12)),
		newData(d.New("Paul")),
		d.New(uint(12321)),
		d.New(int(12321)),
		newData(d.New("Eve")),
		d.New(uint(543)),
		d.New(int(543)),
		d.New(uint(12)),
		d.New(int(8493)),
		newData(d.New("John")),
		d.New(uint(8493)),
		d.New(int(-134)),
		d.New(uint(381)),
		newData(d.New("Jill")),
		d.New(int(381)),
		d.New(uint(11)),
		newData(d.New("Mary")),
	}

	ts := dataSorter(typs)
	ts.Sort(d.Flag)
	fmt.Println(ts)
}
func TestPairSorter(t *testing.T) {
	var strPairs = []Paired{
		newPair(d.New("key 0"), d.New("Aaron")),
		newPair(d.New("key 1"), d.New("Aardvark")),
		newPair(d.New("key 2"), d.New("Adam")),
		newPair(d.New("key 3"), d.New("Victor")),
		newPair(d.New("key 4"), d.New("Sylvest")),
		newPair(d.New("key 5"), d.New("Stepen")),
		newPair(d.New("key 6"), d.New("Sonja")),
		newPair(d.New("key 7"), d.New("Tom")),
		newPair(d.New("key 8"), d.New("Britta")),
		newPair(d.New("key 9"), d.New("Peter")),
		newPair(d.New("key 10"), d.New("Paul")),
		newPair(d.New("key 11"), d.New("Mary")),
		newPair(d.New("key 12"), d.New("Eve")),
		newPair(d.New("key 13"), d.New("John")),
		newPair(d.New("key 14"), d.New("Jill")),
	}

	ps := pairSorter(strPairs)
	ps.Sort(d.String)
	fmt.Println(ps)
	fmt.Println(ps.Search(d.New("key 10")))
	pair := ps.Get(d.New("key 10"))

	fmt.Println(pair)

	var uintPairs = []Paired{
		newPair(d.UintVal(uint(894512)), d.New("Aaron")),
		newPair(d.UintVal(uint(48512)), d.New("Aardvark")),
		newPair(d.UintVal(uint(2489512)), d.New("Adam")),
		newPair(d.UintVal(uint(4895212)), d.New("Victor")),
		newPair(d.UintVal(uint(982512)), d.New("Sylvest")),
		newPair(d.UintVal(uint(25712)), d.New("Stepen")),
		newPair(d.UintVal(uint(2412)), d.New("Sonja")),
		newPair(d.UintVal(uint(8973412)), d.New("Tom")),
		newPair(d.UintVal(uint(8032112)), d.New("Britta")),
		newPair(d.UintVal(uint(1389812)), d.New("Peter")),
		newPair(d.UintVal(uint(832412)), d.New("Paul")),
		newPair(d.UintVal(uint(1331812)), d.New("Mary")),
		newPair(d.UintVal(uint(234412)), d.New("Eve")),
		newPair(d.UintVal(uint(459012)), d.New("John")),
		newPair(d.UintVal(uint(234212)), d.New("Jill")),
	}

	us := pairSorter(uintPairs)
	fmt.Println(us)
	us.Sort(d.Unsigned)
	fmt.Println(us)

	var intPairs = []Paired{
		newPair(d.IntVal(int(894512)), d.New("Aaron")),
		newPair(d.IntVal(int(48512)), d.New("Aardvark")),
		newPair(d.IntVal(int(2489512)), d.New("Adam")),
		newPair(d.IntVal(int(4895212)), d.New("Victor")),
		newPair(d.IntVal(int(982512)), d.New("Sylvest")),
		newPair(d.IntVal(int(25712)), d.New("Stepen")),
		newPair(d.IntVal(int(2412)), d.New("Sonja")),
		newPair(d.IntVal(int(8973412)), d.New("Tom")),
		newPair(d.IntVal(int(8032112)), d.New("Britta")),
		newPair(d.IntVal(int(1389812)), d.New("Peter")),
		newPair(d.IntVal(int(832412)), d.New("Paul")),
		newPair(d.IntVal(int(1331812)), d.New("Mary")),
		newPair(d.IntVal(int(234412)), d.New("Eve")),
		newPair(d.IntVal(int(459012)), d.New("John")),
		newPair(d.IntVal(int(234212)), d.New("Jill")),
	}

	is := pairSorter(intPairs)
	fmt.Println(is)
	is.Sort(d.Integer)
	fmt.Println(is)
}
