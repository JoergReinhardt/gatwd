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
	PipeReaderVal io.PipeReader
	PipeWriterVal io.PipeWriter
	ReadVal       struct {
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

	// THREADSAFE BUFFER
	TSBuffer struct {
		sync.Mutex
		*BufferVal
	}

	// THREADSAFE READERS/WRITERS
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

func (v WriteVal) Close() error                { return io.Closer(v).Close() }
func (v WriteVal) TypeNat() TyNative           { return Writer.TypeNat() }
func (v WriteVal) Write(p []byte) (int, error) { return io.Writer(v).Write(p) }

func (v ReadWriteVal) Close() error                { return io.Closer(v).Close() }
func (v ReadWriteVal) TypeNat() TyNative           { return Reader.TypeNat() | Writer.TypeNat() }
func (v ReadWriteVal) Read(p []byte) (int, error)  { return io.Reader(v).Read(p) }
func (v ReadWriteVal) Write(p []byte) (int, error) { return io.Writer(v).Write(p) }

func (v PipeReaderVal) Eval(n ...Native) Native { return v }
func (v PipeReaderVal) TypeNat() TyNative       { return Reader.TypeNat() | Pipe.TypeNat() }
func (v PipeReaderVal) Close() error            { return io.Closer(v).Close() }

func (v PipeWriterVal) Eval(n ...Native) Native     { return v }
func (v PipeWriterVal) TypeNat() TyNative           { return Writer.TypeNat() | Pipe.TypeNat() }
func (v PipeWriterVal) Write(p []byte) (int, error) { return io.Writer(v).Write(p) }
func (v PipeWriterVal) Close() error                { return io.Closer(v).Close() }

// BYTE BUFFER IMPLEMENTATION
func NewBuffer(...BytesVal) *BufferVal {
	return (*BufferVal)(bytes.NewBuffer([]byte{}))
}
func (v *BufferVal) TypeNat() TyNative { return Buffer.TypeNat() }

func (v *BufferVal) Write(p BytesVal) (int, error) {
	return ((*bytes.Buffer)(v)).Write(BytesVal(p))
}
func (v *BufferVal) WriteBytes(b BytesVal) (int, error) {
	return ((*bytes.Buffer)(v)).Write([]byte(b))
}
func (v *BufferVal) WriteByte(b ByteVal) error {
	return ((*bytes.Buffer)(v)).WriteByte(byte(b))
}
func (v *BufferVal) WriteString(s StrVal) (int, error) {
	return ((*bytes.Buffer)(v)).WriteString(string(s))
}
func (v *BufferVal) WriteRune(r RuneVal) (int, error) {
	return ((*bytes.Buffer)(v)).WriteRune(rune(r))
}
func (v *BufferVal) WriteTo(w WriteVal) (int64, error) {
	return ((*bytes.Buffer)(v)).WriteTo(io.Writer(w))
}
func (v *BufferVal) Read(p BytesVal) (int, error) {
	return ((*bytes.Buffer)(v)).Read([]byte(p))
}
func (v *BufferVal) ReadFrom(r ReadVal) (int64, error) {
	return ((*bytes.Buffer)(v)).ReadFrom(io.Reader(r))
}
func (v *BufferVal) ReadByte() (ByteVal, error) {
	var b, err = ((*bytes.Buffer)(v)).ReadByte()
	return ByteVal(b), err
}
func (v *BufferVal) ReadBytes(delim ByteVal) (BytesVal, error) {
	var bs, err = ((*bytes.Buffer)(v)).ReadBytes(byte(delim))
	return BytesVal(bs), err
}
func (v *BufferVal) ReadString(delim ByteVal) (StrVal, error) {
	var bs, err = ((*bytes.Buffer)(v)).ReadString(byte(delim))
	return StrVal(bs), err
}
func (v *BufferVal) ReadRune() (RuneVal, int, error) {
	var r, n, err = ((*bytes.Buffer)(v)).ReadRune()
	return RuneVal(r), n, err
}
func (v *BufferVal) Len() int            { return ((*bytes.Buffer)(v).Len()) }
func (v *BufferVal) String() string      { return ((*bytes.Buffer)(v)).String() }
func (v *BufferVal) Bytes() BytesVal     { return BytesVal(((*bytes.Buffer)(v)).Bytes()) }
func (v *BufferVal) Next(n int) BytesVal { return ((*bytes.Buffer)(v)).Next(n) }
func (v *BufferVal) UnreadByte() error   { return ((*bytes.Buffer)(v)).UnreadByte() }
func (v *BufferVal) UnreadRune() error   { return ((*bytes.Buffer)(v)).UnreadRune() }
func (v *BufferVal) Truncate(n int)      { ((*bytes.Buffer)(v)).Truncate(n) }
func (v *BufferVal) Reset()              { ((*bytes.Buffer)(v)).Reset() }
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

func (v TSRead) Close() error               { return io.Closer(v).Close() }
func (v TSRead) TypeNat() TyNative          { return Reader.TypeNat() }
func (v TSRead) Read(p []byte) (int, error) { return io.Reader(v).Read(p) }
func (v TSRead) Lock()                      { v.RWMutex.Lock() }
func (v TSRead) Unlock()                    { v.RWMutex.Unlock() }

func (v TSWrite) Close() error                { return io.Closer(v).Close() }
func (v TSWrite) TypeNat() TyNative           { return Writer.TypeNat() }
func (v TSWrite) Write(p []byte) (int, error) { return io.Writer(v).Write(p) }
func (v TSWrite) Lock()                       { v.RWMutex.Lock() }
func (v TSWrite) Unlock()                     { v.RWMutex.Unlock() }

func (v TSReadWrite) Close() error                { return io.Closer(v).Close() }
func (v TSReadWrite) TypeNat() TyNative           { return Reader.TypeNat() | Writer.TypeNat() }
func (v TSReadWrite) Read(p []byte) (int, error)  { return io.Reader(v).Read(p) }
func (v TSReadWrite) Write(p []byte) (int, error) { return io.Writer(v).Write(p) }
func (v TSReadWrite) Lock()                       { v.RWMutex.Lock() }
func (v TSReadWrite) Unlock()                     { v.RWMutex.Unlock() }
