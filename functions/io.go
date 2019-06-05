package functions

import (
	"bytes"

	d "github.com/joergreinhardt/gatwd/data"
)

type (
	// BYTES BUFFER
	BufferVal func(...d.Native) *bytes.Buffer
)
