package main

import (
	"fmt"
	"github.com/joergreinhardt/gatwd/visualizer"
)

func main() {
	opts := visualizer.ParseOptions("--box-all")
	
	val := struct{ Value string }{ Value: "Root" }
	res := visualizer.Format(val, opts)
	fmt.Println(res)
}
