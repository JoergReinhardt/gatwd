package types

import (
	"fmt"
	"testing"
)

func TestFunctionComposition(t *testing.T) {

	helloWorld := func(...Data) Data { return conData("hello, world!") }
	helloData := conData("hello World data")

	lam := composeLambda(helloWorld, String, ConFix, List)
	clo := enclsoseLambda(lam)
	dtc := encloseData(helloData)

	fmt.Printf("calling helloWorld function directly: %s\n", helloWorld())
	fmt.Printf("calling closure:%v\n"+
		"type: %s\nenclosed: %s\n"+
		"call closure directy, assert as lambda: %s\n"+
		"call Enclosed %s\n",
		clo, clo.Type().String(),
		clo.Enclosed(), clo().(lambda).Call(),
		clo.Enclosed().(lambda).Call())

	fmt.Printf("calling lamba:\n"+
		"type: %s\targs: %v\n"+
		"arity: %s\tfixity: %s\n"+
		"calling call method: %s\n",
		lam.Flag().String(),
		lam.ArgTypes(),
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
		fenc.ArgTypes(),
		fenc.Arity(),
		fenc.Fixity(),
		fenc.Call(),
	)
}
