package data

import (
	"bytes"
	"io"
	"sync"
)

type (

	//// IO SYNCHRONOUS
	///
	// BYTES BUFFER
	BufferVal bytes.Buffer

	// READER/WRITER
	PipeReadVal  io.PipeReader
	PipeWriteVal io.PipeWriter
	ReadVal      struct {
		BytesVal
		io.ReadCloser
	}
	ReadWriteVal struct {
		BytesVal
		io.ReadWriteCloser
	}
	WriteVal struct {
		BytesVal
		io.WriteCloser
	}

	//// IO ASYNCHRONOUS
	///
	// THREADSAFE NATIVES
	TSNative struct {
		sync.Mutex
		Native
	}

	// THREADSAFE SLICE
	TSSlice struct {
		sync.Mutex
		DataSlice
	}

	// THREADSAFE BUFFER
	TSBuffer struct {
		sync.Mutex
		*BufferVal
	}

	// THREADSdFE READERS/WRITERS
	TSRead struct {
		sync.RWMutex
		ReadVal
	}
	TSWrite struct {
		sync.RWMutex
		WriteVal
	}
	TSReadWrite struct {
		sync.RWMutex
		ReadWriteVal
	}
)

// READER/WRITER IMPLEMENTATION
func (v ReadVal) Close() error               { return io.Closer(v).Close() }
func (v ReadVal) TypeNat() TyNative          { return Reader.TypeNat() }
func (v ReadVal) Read(p []byte) (int, error) { return io.Reader(v).Read(p) }
func (v ReadVal) Eval(n ...Native) Native {
	if len(n) > 0 {
		for _, val := range n {
			switch val.TypeNat() {
			case Bytes:
				io.Reader(v).Read(val.(BytesVal))
			case String:
				var p = []byte{}
				io.Reader(v).Read(p)
				return StrVal(string(p))
			default:
				var p = []byte{}
				io.Reader(v).Read(p)
				return BytesVal(p)
			}
		}
	}
	return v
}

func (v WriteVal) Close() error                { return io.Closer(v).Close() }
func (v WriteVal) TypeNat() TyNative           { return Writer.TypeNat() }
func (v WriteVal) Write(p []byte) (int, error) { return io.Writer(v).Write(p) }
func (v WriteVal) Eval(n ...Native) Native {
	if len(n) > 0 {
		for _, val := range n {
			switch val.TypeNat() {
			case Bytes:
				((io.Writer)(v)).Write(
					val.(BytesVal))
			case String:
				((io.Writer)(v)).Write(
					[]byte(string(val.(StrVal))))
			default:
				((io.Writer)(v)).Write(
					[]byte(string(StrVal(val.String()))))
			}
		}
	}
	return v
}

func (v ReadWriteVal) Close() error                { return io.Closer(v).Close() }
func (v ReadWriteVal) TypeNat() TyNative           { return Reader.TypeNat() | Writer.TypeNat() }
func (v ReadWriteVal) Read(p []byte) (int, error)  { return io.Reader(v).Read(p) }
func (v ReadWriteVal) Write(p []byte) (int, error) { return io.Writer(v).Write(p) }
func (v ReadWriteVal) Eval(n ...Native) Native {
	if len(n) > 0 {
		for _, val := range n {
			switch val.TypeNat() {
			case Bytes:
				io.ReadWriter(v).Write(
					val.(BytesVal))
			case String:
				io.ReadWriter(v).Write(
					[]byte(string(val.(StrVal))))
			default:
				io.ReadWriter(v).Write(
					[]byte(string(StrVal(val.String()))))
			}
		}
	}
	return v
}

// pipe endpoints come in pairs
func NewPipe() (*PipeReadVal, *PipeWriteVal) {
	var pr, pw = io.Pipe()
	return (*PipeReadVal)(pr), (*PipeWriteVal)(pw)
}
func (v PipeReadVal) Close() error                { return io.Closer(v).Close() }
func (v PipeReadVal) TypeNat() TyNative           { return Reader.TypeNat() | Pipe.TypeNat() }
func (v *PipeReadVal) Read(p []byte) (int, error) { return (*io.PipeWriter)(v).Write(p) }
func (v *PipeReadVal) Eval(n ...Native) Native {
	if len(n) > 0 {
		for _, val := range n {
			switch val.TypeNat() {
			case Bytes:
				((*io.PipeReader)(v)).Read(val.(BytesVal))
			case String:
				var p = []byte{}
				((*io.PipeReader)(v)).Read(p)
				return StrVal(string(p))
			default:
				var p = []byte{}
				((*io.PipeReader)(v)).Read(p)
				return BytesVal(p)
			}
		}
	}
	return v
}
func (v PipeWriteVal) Close() error                 { return io.Closer(v).Close() }
func (v PipeWriteVal) TypeNat() TyNative            { return Writer.TypeNat() | Pipe.TypeNat() }
func (v *PipeWriteVal) Write(p []byte) (int, error) { return (*io.PipeWriter)(v).Write(p) }
func (v *PipeWriteVal) Eval(n ...Native) Native {
	if len(n) > 0 {
		for _, val := range n {
			switch val.TypeNat() {
			case Bytes:
				((*io.PipeWriter)(v)).Write(
					val.(BytesVal))
			case String:
				((*io.PipeWriter)(v)).Write(
					[]byte(string(val.(StrVal))))
			default:
				((*io.PipeWriter)(v)).Write(
					[]byte(string(StrVal(val.String()))))
			}
		}
	}
	return v
}

// BYTE BUFFER IMPLEMENTATION
func NewBuffer(...byte) *BufferVal {
	return (*BufferVal)(bytes.NewBuffer([]byte{}))
}
func (v *BufferVal) Write(p []byte) (int, error) {
	return ((*bytes.Buffer)(v)).Write(p)
}
func (v *BufferVal) WriteBytes(b []byte) (int, error) {
	return ((*bytes.Buffer)(v)).Write(b)
}
func (v *BufferVal) WriteByte(b byte) error {
	return ((*bytes.Buffer)(v)).WriteByte(b)
}
func (v *BufferVal) WriteString(s string) (int, error) {
	return ((*bytes.Buffer)(v)).WriteString(s)
}
func (v *BufferVal) WriteRune(r rune) (int, error) {
	return ((*bytes.Buffer)(v)).WriteRune(r)
}
func (v *BufferVal) WriteTo(w io.Writer) (int64, error) {
	return ((*bytes.Buffer)(v)).WriteTo(w)
}
func (v *BufferVal) Read(p []byte) (int, error) {
	return ((*bytes.Buffer)(v)).Read(p)
}
func (v *BufferVal) ReadFrom(r io.Reader) (int64, error) {
	return ((*bytes.Buffer)(v)).ReadFrom(r)
}
func (v *BufferVal) ReadByte() (byte, error) {
	var b, err = ((*bytes.Buffer)(v)).ReadByte()
	return b, err
}
func (v *BufferVal) ReadBytes(delim byte) ([]byte, error) {
	var bs, err = ((*bytes.Buffer)(v)).ReadBytes(delim)
	return bs, err
}
func (v *BufferVal) ReadString(delim byte) (string, error) {
	var bs, err = ((*bytes.Buffer)(v)).ReadString(delim)
	return bs, err
}
func (v *BufferVal) ReadRune() (rune, int, error) {
	var r, n, err = ((*bytes.Buffer)(v)).ReadRune()
	return r, n, err
}
func (v *BufferVal) TypeNat() TyNative { return Buffer.TypeNat() }
func (v *BufferVal) Len() int          { return ((*bytes.Buffer)(v).Len()) }
func (v *BufferVal) String() string    { return ((*bytes.Buffer)(v)).String() }
func (v *BufferVal) Bytes() []byte     { return ((*bytes.Buffer)(v)).Bytes() }
func (v *BufferVal) Next(n int) []byte { return ((*bytes.Buffer)(v)).Next(n) }
func (v *BufferVal) UnreadByte() error { return ((*bytes.Buffer)(v)).UnreadByte() }
func (v *BufferVal) UnreadRune() error { return ((*bytes.Buffer)(v)).UnreadRune() }
func (v *BufferVal) Truncate(n int)    { ((*bytes.Buffer)(v)).Truncate(n) }
func (v *BufferVal) Reset()            { ((*bytes.Buffer)(v)).Reset() }
func (v *BufferVal) Eval(n ...Native) Native {
	if len(n) > 0 {
		for _, val := range n {
			switch val.TypeNat() {
			case Bytes:
				((*bytes.Buffer)(v)).Write(val.(BytesVal))
			case String:
				((*bytes.Buffer)(v)).WriteString(string(val.(StrVal)))
			default:
				((*bytes.Buffer)(v)).WriteString(string(StrVal(val.String())))
			}
		}
	}
	return v
}

// THREADSAFE NATIVE VALUE
func NewTSNative(n Native) TSNative { return TSNative{sync.Mutex{}, n} }
func (v TSNative) Lock()            { v.Mutex.Lock() }
func (v TSNative) Unlock()          { v.Mutex.Unlock() }
func (v TSNative) TypeNat() TyNative {
	return Reader.TypeNat() | Writer.TypeNat()
}
func (v TSNative) Get() Native {
	v.Mutex.Lock()
	defer v.Mutex.Unlock()
	return v.Native
}
func (v TSNative) Set(n Native) Native {
	v.Mutex.Lock()
	defer v.Mutex.Unlock()
	v.Native = n
	return v
}
func (v TSNative) Eval(n ...Native) Native {
	if len(n) > 0 {
		if len(n) > 1 {
			return DataSlice(n)
		}
		return n[0]
	}
	return v
}

// THREADSAFE SLICE OF NATIVE VALUES
func NewTSSlice() *TSSlice {
	return &TSSlice{
		sync.Mutex{},
		DataSlice{},
	}
}
func (v TSSlice) Lock()   { v.Mutex.Lock() }
func (v TSSlice) Unlock() { v.Mutex.Unlock() }
func (v TSSlice) TypeNat() TyNative {
	return Reader.TypeNat() | Writer.TypeNat()
}
func (v TSSlice) Eval(n ...Native) Native {
	if len(n) > 0 {
		if len(n) > 1 {
			return DataSlice(n)
		}
		return n[0]
	}
	return v
}

// THREADSAFE BUFFER
func NewTSBuffer() *TSBuffer {
	return &TSBuffer{
		sync.Mutex{},
		NewBuffer(),
	}
}
func (v TSBuffer) TypeNat() TyNative {
	return Reader.TypeNat() | Writer.TypeNat()
}
func (v TSBuffer) Lock()   { v.Mutex.Lock() }
func (v TSBuffer) Unlock() { v.Mutex.Unlock() }

// THREADSAFE READER
func (v TSRead) Lock()             { v.RWMutex.Lock() }
func (v TSRead) Unlock()           { v.RWMutex.Unlock() }
func (v TSRead) Close() error      { return io.Closer(v).Close() }
func (v TSRead) TypeNat() TyNative { return Reader.TypeNat() }
func (v TSRead) Read(p []byte) (int, error) {
	v.RWMutex.Lock()
	defer v.RWMutex.Unlock()
	return io.Reader(v).Read(p)
}
func (v TSRead) Eval(n ...Native) Native {
	v.RWMutex.Lock()
	defer v.RWMutex.Unlock()
	if len(n) > 0 {
		for _, val := range n {
			switch val.TypeNat() {
			case Bytes:
				io.Reader(v).Read(val.(BytesVal))
			case String:
				var p = []byte{}
				io.Reader(v).Read(p)
				return StrVal(string(p))
			default:
				var p = []byte{}
				io.Reader(v).Read(p)
				return BytesVal(p)
			}
		}
	}
	return v
}

// THREAD SAFE WRITER
func (v TSWrite) TypeNat() TyNative { return Writer.TypeNat() }
func (v TSWrite) Lock()             { v.RWMutex.Lock() }
func (v TSWrite) Unlock()           { v.RWMutex.Unlock() }
func (v TSWrite) Close() error      { return io.Closer(v).Close() }
func (v TSWrite) Write(p []byte) (int, error) {
	v.RWMutex.Lock()
	defer v.RWMutex.Unlock()
	return io.Writer(v).Write(p)
}
func (v TSWrite) Eval(n ...Native) Native {
	v.RWMutex.Lock()
	defer v.RWMutex.Unlock()
	if len(n) > 0 {
		for _, val := range n {
			switch val.TypeNat() {
			case Bytes:
				((io.Writer)(v)).Write(
					val.(BytesVal))
			case String:
				((io.Writer)(v)).Write(
					[]byte(string(val.(StrVal))))
			default:
				((io.Writer)(v)).Write(
					[]byte(string(StrVal(val.String()))))
			}
		}
	}
	return v
}

// THREAD SAFE READER/WRITER
func (v TSReadWrite) TypeNat() TyNative { return Reader.TypeNat() | Writer.TypeNat() }
func (v TSReadWrite) Lock()             { v.RWMutex.Lock() }
func (v TSReadWrite) Unlock()           { v.RWMutex.Unlock() }
func (v TSReadWrite) Close() error      { return io.Closer(v).Close() }
func (v TSReadWrite) Read(p []byte) (int, error) {
	return io.Reader(v).Read(p)
}
func (v TSReadWrite) Write(p []byte) (int, error) {
	v.RWMutex.Lock()
	defer v.RWMutex.Unlock()
	return io.Writer(v).Write(p)
}
func (v TSReadWrite) Eval(n ...Native) Native {
	v.RWMutex.Lock()
	defer v.RWMutex.Unlock()
	if len(n) > 0 {
		for _, val := range n {
			switch val.TypeNat() {
			case Bytes:
				((io.Writer)(v)).Write(
					val.(BytesVal))
			case String:
				((io.Writer)(v)).Write(
					[]byte(string(val.(StrVal))))
			default:
				((io.Writer)(v)).Write(
					[]byte(string(StrVal(val.String()))))
			}
		}
	}
	return v
}