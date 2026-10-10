package main

import (
	"github.com/joergreinhardt/gatwd/pkg/cli"
	
	// SEAMLESS PLUGIN REGISTRATION!
	_ "github.com/joergreinhardt/gatwd/examples/14_plugins/my_plugin"
)

func main() {
	cli.Run()
}
