package parse

import (
	"fmt"
	"io"
	"strings"

	d "github.com/joergreinhardt/gatwd/data"
	f "github.com/joergreinhardt/gatwd/functions"
	l "github.com/joergreinhardt/gatwd/lex"
)

// lexer error returns an error, reflecting the rune that was lexed and the
// error that was returned.
type LexerError f.ErrorFnc

func (l LexerError) Error() string                 { return l().Error() }
func (l LexerError) String() string                { return l.native().String() }
func (l LexerError) TypeNat() d.TyNative           { return l.native().TypeNat() }
func (l LexerError) Eval(dat ...d.Native) d.Native { return l.native().Eval(dat...) }
func (l LexerError) native() d.ErrorVal            { return d.New(l()).(d.ErrorVal) }

// generates string to return by error function
func NewLexerError(err error, r rune) LexerError {
	if err == io.EOF {
		return LexerError(func() error { return err })
	}
	return LexerError(func() error {
		return fmt.Errorf(
			"lexer error lexing rune: %s\nerror:\n%s\n", string(r), err.Error())
	})
}

// new lexer returns a thread safe buffer, queue and a state function sharing
// those facilitys.
func NewLexer() (
	buffer *d.TSBuffer,
	queue *d.TSSlice,
	state f.StateFnc,
) {
	buffer = d.NewTSBuffer()
	queue = d.NewTSSlice()

	return buffer, queue, f.StateFnc(func() f.StateFnc { return lexer(buffer, queue) })
}

// the lexers main loop pops a rune from buffer & passes control, that rune,
// buffer & queue on to one of the parser functions to generate & emit the
// actual token. the return state function checks the bottom of the queue. if
// it turns out to be an error token, lexing will be haltet.
func lexer(buffer *d.TSBuffer, queue *d.TSSlice) f.StateFnc {

	// read a rune, write error to queue, in case something went wrong
	var r = pop(buffer, queue)

	// jump to parse functions, based on rune popped
	switch {
	case digit(r):
		parseDigit(buffer, queue, r)
	case letter(r):
		parseLetter(buffer, queue, r)
	case capital(r):
		parseCapital(buffer, queue, r)
	case syntax(r):
		parseSyntax(buffer, queue, r)
	}
	// return continuation state
	return returnState(buffer, queue)
}

func halt(queue *d.TSSlice) bool {
	return queue.DataSlice.Bottom().(Token).TypeTok() == Error_Token
}

// handle error generates an error token and appends it to the queue
func handleError(err error, r rune, queue *d.TSSlice) {
	// generate an error token and append it to the queue
	var tok = NewErrorToken(NewLexerError(err, r))
	push(tok, queue)
	return
}

// pops a rune from the buffer and handles the error in case one arises
func pop(buffer *d.TSBuffer, queue *d.TSSlice) rune {
	var r, n, err = buffer.ReadRune()
	if err != nil || n == 0 {
		handleError(err, r, queue)
	}
	return r
}
func peek(buffer *d.TSBuffer, queue *d.TSSlice) rune {
	var r = pop(buffer, queue)
	backup(buffer)
	return r
}

// push new token on to queue
func push(tok Token, queue *d.TSSlice) {
	(*queue).DataSlice = d.SliceAppend(queue.DataSlice, tok)
}

// unreads the last token
func backup(buffer *d.TSBuffer) {
	_ = buffer.UnreadRune()
}

// return state returns either nil to halt execution, should the queue contain
// an error, or wraps updated buffer and queue to become the returned next
// state function.
func returnState(buffer *d.TSBuffer, queue *d.TSSlice) f.StateFnc {
	if halt(queue) {
		return nil
	}
	// return new state function wrapping updated lexer & queue
	return f.StateFnc(func() f.StateFnc { return lexer(buffer, queue) })
}

// returns true, if rune turns out to be a digit
func digit(r rune) bool {
	return strings.ContainsAny(string(r), l.DigitString)
}

// returns true, if rune turns out to be a letter
func letter(r rune) bool {
	return strings.ContainsAny(string(r), l.LetterString)
}

// returns true, if rune turns out to be a capital
func capital(r rune) bool {
	return strings.ContainsAny(string(r), l.CapitalString)
}

// returns true on all syntax elememts.
func syntax(r rune) bool {
	_, ok := l.AllSyntaxRunes[r]
	return ok
}

// parse functions pop a rune to see if next rune turns out to be the same kind
// as the currently parsed. it then calls itself recursively, accumulating &
// passing on popped runes if that's the case. otherwise the buffer state will
// be backed up by unreading the popped rune. a token is then generated from
// the passed rune(s) & pushes it on the queue and returns control to the main
// lexer function.
func parseDigit(b *d.TSBuffer, q *d.TSSlice, r rune) {
	var runes = []rune{}
	for digit(r) {
		runes = append(runes, r)
		r = pop(b, q)
	}
	backup(b)
	var tok = NewDigitToken(string(runes))
	fmt.Printf("digit token generated: %s\n", tok)
	push(tok, q)
	return
}

func parseLetter(b *d.TSBuffer, q *d.TSSlice, r rune) {
	var runes = []rune{}
	for letter(r) {
		runes = append(runes, r)
		r = pop(b, q)
	}
	backup(b)
	var tok = NewDataValueToken(string(runes))
	fmt.Printf("letter token generated: %s\n", tok)
	push(tok, q)
	return
}

func parseCapital(b *d.TSBuffer, q *d.TSSlice, r rune) {
	var runes = []rune{}
	for capital(r) {
		runes = append(runes, r)
		r = pop(b, q)
	}
	backup(b)
	push(NewDataValueToken(string(runes)), q)
	return
}
func parseSyntax(b *d.TSBuffer, q *d.TSSlice, r rune) {
	// try to generate syntax item from combining ascii chars first
	var item, _ = asciiGreedy(b, q, nil, r)
	if item != nil { // if item got returned‥.
		// convert to syntax token and push
		var tok = NewSyntaxToken(item)
		fmt.Printf("ascii syntax token generated: %s\n", tok)
		push(tok, q)
		return
	}
	var ok bool
	// if no ascii di-/ or trigraph matched, try utf8 syntactic match
	item, ok = l.MatchUtf8(string(r))
	if ok { // if item got returned‥.
		// convert it to utf8 syntax token
		var tok = NewSyntaxToken(item)
		fmt.Printf("utf syntax token generated: %s\n", tok)
		push(tok, q)
		return
	}
	// otherwise just return
	return
}

// returns an item, if one or more runes match a di-, or trigraph.
func asciiGreedy(
	b *d.TSBuffer,
	q *d.TSSlice,
	item l.Item,
	runes ...rune,
) (l.Item, []rune) {
	// pop next rune‥.
	runes = append(runes, pop(b, q))
	//‥.and concat to string
	var str = string(runes)
	if l.Match(str) { // test if str matches ascisyntax‥.
		item = l.GetAsciiItem(str)
	}
	if l.Match(string(append(runes, peek(b, q)))) {
		runes = append(runes, pop(b, q))
		str = string(runes)
		item = l.GetAsciiItem(str)
	}
	if item == nil {
		// if popped rune doesn't result in a longer match, backup last rune
		// (should be the first rune popped, if no digraph matched)
		backup(b)
	}
	// return item and runes (should be nil and runes passed in initialy,
	// if nothing matched)
	return item, runes
}

// unknown runes generate new tokens
func parseUnknown(b *d.TSBuffer, q *d.TSSlice, r ...rune) {
	push(NewDataValueToken(string(r)), q)
	return
}
