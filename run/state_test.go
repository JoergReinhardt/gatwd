package run

import (
	"fmt"
	"testing"

	f "github.com/JoergReinhardt/gatwd/functions"
)

func TestStateFncProgress(t *testing.T) {
	var count int
	var sf f.StateFnc
	sf = func() f.StateFnc {
		count = count + 1
		if count == 10 {
			return nil
		}
		return sf
	}
	sf.Run()
	fmt.Println(count)
	if count != 10 {
		t.Fail()
	}
}

var line = []rune("\\y => -> === :: \n ab\tcd 123 12 data")
var otherline = []rune("\\y => -> === :: abcd 123 12 data")

func TestUnicodeReplacement(t *testing.T) {
	fmt.Printf("ascii line: %s\n", string(line))
	fmt.Printf("ascii byte length %d\n", len([]byte(string(line))))
	fmt.Printf("projected unicode length in byte of ascii line as calculated %d\n\n",
		unilen(uni(line)))

	if len([]byte(string(line))) != unilen(uni(line)) {
		t.Fail()
	}

	fmt.Printf("unicode line: %s\n", string(uni(line)))
	fmt.Printf("unicode byte length: %d\n", len([]byte(string(uni(line)))))
	fmt.Printf("projected ascii length in byte of unicode line as calculated %d\n",
		asclen(uni(line)))

	if len([]byte(string(uni(line)))) != asclen(uni(line)) {
		t.Fail()
	}

	fmt.Printf("unicode other line: %s\n", string(uni(otherline)))
	fmt.Printf("unicode byte length other line: %d\n", len([]byte(string(uni(otherline)))))
	fmt.Printf("projected ascii length in byte of unicode other line as calculated %d\n",
		asclen(uni(otherline)))
	if len([]byte(string(uni(otherline)))) != asclen(uni(otherline)) {
		t.Fail()
	}
}
func TestThreadsafeSource(t *testing.T) {
	source := NewSource()

	source.Write([]byte(string(line)))
	fmt.Printf("fresh written source:\n %s\n\n", source)

	source.Delete(3)
	fmt.Printf("source after Delete(3):\n %s\n\n", source)

	source.InsertSlice(8, 10, []byte(string(line)))
	fmt.Printf("source after InsertSlice(8,10,[]byte(string(line))):\n %s\n\n", source)

	source.Cut(5, 30)
	fmt.Printf("source after Cut(5,30):\n %s\n\n", source)
}
func TestLineBufferReadLine(t *testing.T) {
	buf := NewSource()
	buf.WriteRunes(line)
	fmt.Printf("prepared buffer:\n %s\n\n", buf)

	var p = []byte{}
	i, err := buf.ReadLine(&p)
	if err != nil {
		fmt.Printf("bytes read: %d, error encountered:\n %s\n\n", i, err)
	}
	fmt.Printf("bytes read: %d, line read:\n %s\n\n", i, p)
	fmt.Printf("buffer left:\n %s\n\n", buf)
}
func TestLineBufferReadPresized(t *testing.T) {

	buf := NewSource()
	buf.WriteRunes(line)
	fmt.Printf("prepared buffer:\n %s\n\n", buf)

	var p = make([]byte, 0, 10)

	i, err := buf.Read(&p)
	if err != nil {
		fmt.Printf("bytes read: %d, error encountered:\n %s\n\n", i, err)
	}
	fmt.Printf("bytes read: %d, line read:\n %s\n\n", i, p)
	fmt.Printf("buffer left:\n %s\n\n", buf)
}
func TestLineBufferUpdateTrailing(t *testing.T) {
	buf := NewSource()
	buf.WriteRunes(line)

	buf.UpdateTrailing([]rune("####"))
	fmt.Printf("buffer after update:\n %s\n\n", buf)

	buf.UpdateTrailing([]rune("####-----####"))
	fmt.Printf("buffer after update:\n %s\n\n", buf)
}
