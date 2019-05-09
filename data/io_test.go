package data

import (
	"fmt"
	"testing"
)

var buf = NewBuffer()

func TestBuffer(t *testing.T) {
	var n, err = buf.Write([]byte("this is a test string to be converted to a byte array\nand written to the buffer.\n"))
	fmt.Printf("bytes written %d, error: %s\n", n, err)
	var str string
	str, err = buf.ReadString([]byte("\n")[0])
	if err != nil {
		fmt.Printf("error reading first line from buffer: %s\n", err)
		t.Fail()
	}
	fmt.Printf("first line read from buffer: %s\n", str)
	str, err = buf.ReadString([]byte("\n")[0])
	if err != nil {
		fmt.Printf("error reading second line from buffer: %s\n", err)
		t.Fail()
	}
	fmt.Printf("second line read from buffer: %s\n", str)
}

var tsbuf = NewBuffer()
