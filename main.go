package main

import (
	"fmt"
	"github.com/JoergReinhardt/gatwd/run"
)

func main() {

	rl, linebuf := run.NewReadLine()

	rl.Run()

	for _, line := range string(linebuf.Bytes()) {
		fmt.Println(line)
	}
}
