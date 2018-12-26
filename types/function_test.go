package types

import (
	"fmt"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestApplyPartial(t *testing.T) {
	data := conData(1, 2.0, "three", uint(4), (uint(1) << 5)).(slice)

	lambds := []lambda{}

	for _, dat := range data.Slice() {
		l := composeLambda(encloseData(dat), dat.Flag(), PostFix)
		lambds = append(lambds, l)
	}

	spew.Dump(lambds)

	argtypes := args{}

	for i, l := range lambds {
		fmt.Printf("lambda %d:\ntype: %s\nargs: %s\n\n",
			i, l.Flag().String(), l.Args())
		argtypes = append(argtypes, l.Flag())
	}

	part := enclsoseLambda(
		composeLambda(func(...Data) Data { return strVal("partial") },
			String.Flag(), PostFix, argtypes...))

	fmt.Printf("partial fn args: %s", part.Args())

	var d1 = []Data{}
	part, d1 = applyPartial(part.Enclosed(), argtypes, data.Slice()[:2]...)

	fmt.Printf("args after partial arg application: %s\nresult s1: %s\n\n",
		part.Args(), d1)

	var d2 = []Data{}
	part, d2 = applyPartial(part.Enclosed(), argtypes, data.Slice()[2:]...)

	fmt.Printf("args after full arg application: %s\nresult: %s\n\n",
		part.Args(), d2)
}

func TestFunctionComposition(t *testing.T) {

	helloWorld := func(...Data) Data { return conData("hello, world!") }
	helloData := conData("hello World data")

	lam := composeLambda(helloWorld, String, ConFix, List.Flag())
	clo := enclsoseLambda(lam)
	dtc := encloseData(helloData)

	fmt.Printf("calling helloWorld function directly: %s\n", helloWorld())
	fmt.Printf("calling closure:%v\n"+
		"type: %s\n"+
		"call closure directy, assert as lambda: %s\n"+
		"call Enclosed %s\n",
		clo, clo.Flag().String(),
		clo().(lambda).Call(),
		clo.Enclosed().Call())

	fmt.Printf("calling lamba:\n"+
		"type: %s\targs: %v\n"+
		"arity: %s\tfixity: %s\n"+
		"calling call method: %s\n",
		lam.Flag().String(),
		lam.Args(),
		lam.Arity().String(),
		lam.Fixity().String(),
		lam.Call())

	str1 := dtc()
	fmt.Printf("call to data enclosure: %v\n", str1)

	dat := clo()
	fmt.Printf("closure data without parameters: %v\n", dat)
	dat = clo(nilVal{})
	fmt.Printf("closure data with nil value: %s\n", dat)
	dat = clo(helloData, str1)
	fmt.Printf("closure data with passed arguments value: %s\n", dat)

	fenc := composeFunction("Test Function", lam)

	fmt.Printf("test function: %v\n"+
		"name: %s\n"+
		"type: %s\n"+
		"params: %v\n"+
		"arity: %s\n"+
		"fixity: %s\n"+
		"returns from call: %s\n",
		fenc,
		fenc.Name(),
		fenc.Flag().String(),
		fenc.Args(),
		fenc.Arity(),
		fenc.Fixity(),
		fenc.Call(),
	)
}
