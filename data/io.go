package data

import (
	"bytes"
	"io"
	"sync"
	"time"
)

type (
	// BYTES BUFFER
	BufferVal bytes.Buffer

	// READER/WRITER
	PipeWriteVal io.PipeWriter
	PipeReadVal  io.PipeReader
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

	//// IO SYNCHRONOUS
	///
	// CONDITION
	SyncCondition sync.Cond

	// WAIT GROUP
	WaitGroup sync.WaitGroup

	// CHANNELS
	Chan        chan Native
	ChanRcv     <-chan Native
	ChanTrx     chan<- Native
	ChanCtrl    chan struct{}
	ChanRcvCtrl <-chan struct{}
	ChanTrxCtrl chan<- struct{}
	ChanTime    chan time.Time
	ChanRcvTime <-chan time.Time
	ChanTrxTime chan<- time.Time

	//// IO ASYNCHRONOUS
	///
	// NATIVES
	TSNative struct {
		*sync.Mutex
		Native
	}

	// SLICE
	TSSlice struct {
		*sync.Mutex
		DataSlice
	}

	// BUFFER
	TSBuffer struct {
		*sync.Mutex
		*BufferVal
	}

	// READERS/WRITERS
	TSRead struct {
		*sync.Mutex
		ReadVal
	}

	TSWrite struct {
		*sync.RWMutex
		WriteVal
	}

	TSReadWrite struct {
		*sync.RWMutex
		ReadWriteVal
	}
)

func NewSyncCondition(locker sync.Locker) *sync.Cond { return sync.NewCond(locker) }
func NewSyncWaitGroup() *sync.WaitGroup              { return &sync.WaitGroup{} }

func NewChan() Chan               { return make(chan Native) }
func NewChanRcv() ChanRcv         { return make(<-chan Native) }
func NewChanTrx() ChanTrx         { return make(chan<- Native) }
func NewChanCtrl() ChanCtrl       { return make(chan struct{}) }
func NewChanRcvCtrl() ChanRcvCtrl { return make(<-chan struct{}) }
func NewChanTrxCtrl() ChanTrxCtrl { return make(chan<- struct{}) }
func NewChanTime() ChanTime       { return make(chan time.Time) }
func NewChanRcvTime() ChanRcvTime { return make(<-chan time.Time) }
func NewChanTrxTime() ChanTrxTime { return make(chan<- time.Time) }

func (c WaitGroup) Eval(...Native) Native { return c }
func (c *WaitGroup) Done()                { c.Done() }

func (c *SyncCondition) Eval(...Native) Native { return c }
func (c *SyncCondition) Broadcast()            { c.Broadcast() }
func (c *SyncCondition) Signal()               { c.Signal() }
func (c *SyncCondition) Wait()                 { c.Wait() }

func (c Chan) Eval(args ...Native) Native {
	if len(args) == 0 {
		return <-c
	}
	for _, arg := range args {
		c <- arg
	}
	return BoolVal(true)
}

func (c ChanRcv) Eval(...Native) Native { return <-c }

func (c ChanTrx) Eval(args ...Native) Native {
	for _, arg := range args {
		c <- arg
	}
	return BoolVal(true)
}

func (c ChanCtrl) Eval(args ...Native) Native {
	if len(args) == 0 {
		<-c
		return BoolVal(true)
	}
	for _, _ = range args {
		c <- struct{}{}
	}
	return BoolVal(true)
}

func (c ChanRcvCtrl) Eval(...Native) Native {
	<-c
	return BoolVal(true)
}

func (c ChanTrxCtrl) Eval(args ...Native) Native {
	for _, _ = range args {
		c <- struct{}{}
	}
	return BoolVal(true)
}

func (c ChanTime) Eval(args ...Native) Native {
	if len(args) > 0 {
		c <- time.Time(args[0].(TimeVal))
		return BoolVal(true)
	}
	return TimeVal(<-c)
}
func (c ChanRcvTime) Eval(...Native) Native { return TimeVal(<-c) }

func (c ChanTrxTime) Eval(args ...Native) Native {
	for _, arg := range args {
		if t, ok := arg.(TimeVal); ok {
			c <- time.Time(t)
		}
	}
	return BoolVal(true)
}

func (c WaitGroup) TypeNat() TyNat     { return SyncWait }
func (c SyncCondition) TypeNat() TyNat { return SyncCon }

func (c Chan) TypeNat() TyNat        { return Channel }
func (c ChanRcv) TypeNat() TyNat     { return Channel }
func (c ChanTrx) TypeNat() TyNat     { return Channel }
func (c ChanCtrl) TypeNat() TyNat    { return Channel }
func (c ChanRcvCtrl) TypeNat() TyNat { return Channel }
func (c ChanTrxCtrl) TypeNat() TyNat { return Channel }
func (c ChanRcvTime) TypeNat() TyNat { return Channel }
func (c ChanTrxTime) TypeNat() TyNat { return Channel }

func (c SyncCondition) String() string { return "sync condition" }
func (c WaitGroup) String() string     { return "wait group" }

func (c Chan) String() string        { return "channel" }
func (c ChanRcv) String() string     { return "Channel" }
func (c ChanTrx) String() string     { return "Channel" }
func (c ChanCtrl) String() string    { return "Channel" }
func (c ChanRcvCtrl) String() string { return "Channel" }
func (c ChanTrxCtrl) String() string { return "Channel" }
func (c ChanRcvTime) String() string { return "Channel" }
func (c ChanTrxTime) String() string { return "Channel" }

// READER/WRITER IMPLEMENTATION
func (v ReadVal) Close() error               { return io.Closer(v).Close() }
func (v ReadVal) TypeNat() TyNat             { return Reader.TypeNat() }
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
func (v WriteVal) TypeNat() TyNat              { return Writer.TypeNat() }
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
func (v ReadWriteVal) TypeNat() TyNat              { return Reader.TypeNat() | Writer.TypeNat() }
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
func (v PipeReadVal) TypeNat() TyNat              { return Reader.TypeNat() | Pipe.TypeNat() }
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
func (v PipeWriteVal) TypeNat() TyNat               { return Writer.TypeNat() | Pipe.TypeNat() }
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
func (v *BufferVal) TypeNat() TyNat    { return Buffer.TypeNat() }
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
func NewTSNative(n Native) TSNative { return TSNative{&sync.Mutex{}, n} }
func (v TSNative) Lock()            { v.Mutex.Lock() }
func (v TSNative) Unlock()          { v.Mutex.Unlock() }
func (v TSNative) TypeNat() TyNat {
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
		&sync.Mutex{},
		DataSlice{},
	}
}
func (v TSSlice) Lock()   { v.Mutex.Lock() }
func (v TSSlice) Unlock() { v.Mutex.Unlock() }
func (v TSSlice) TypeNat() TyNat {
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
		&sync.Mutex{},
		NewBuffer(),
	}
}
func (v TSBuffer) TypeNat() TyNat {
	return Reader.TypeNat() | Writer.TypeNat()
}
func (v TSBuffer) Lock()   { v.Mutex.Lock() }
func (v TSBuffer) Unlock() { v.Mutex.Unlock() }

// THREADSAFE READER
func (v TSRead) Lock()          { v.Mutex.Lock() }
func (v TSRead) Unlock()        { v.Mutex.Unlock() }
func (v TSRead) Close() error   { return io.Closer(v).Close() }
func (v TSRead) TypeNat() TyNat { return Reader.TypeNat() }
func (v TSRead) Read(p []byte) (int, error) {
	v.Mutex.Lock()
	defer v.Mutex.Unlock()
	return io.Reader(v).Read(p)
}
func (v TSRead) Eval(n ...Native) Native {
	v.Mutex.Lock()
	defer v.Mutex.Unlock()
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
func (v TSWrite) TypeNat() TyNat { return Writer.TypeNat() }
func (v TSWrite) Lock()          { v.RWMutex.Lock() }
func (v TSWrite) Unlock()        { v.RWMutex.Unlock() }
func (v TSWrite) Close() error   { return io.Closer(v).Close() }
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
func (v TSReadWrite) TypeNat() TyNat { return Reader.TypeNat() | Writer.TypeNat() }
func (v TSReadWrite) Lock()          { v.RWMutex.Lock() }
func (v TSReadWrite) Unlock()        { v.RWMutex.Unlock() }
func (v TSReadWrite) Close() error   { return io.Closer(v).Close() }
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
