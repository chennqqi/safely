/*
Package safely wraps "go" goroutine-creation syntax, adding recovery from
any panics.

Particularly in HTTP request handlers, spawning separate goroutines can be
especially scary because the panic protection enjoyed in the request handler
itself isn't available in other goroutines. So panics in that context can
bring down the entire server process:

	func (handler *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
		go func() {
			// do something in parallel,
			// but if it panics the whole server exits!
		}()
	}

But using safely to spawn goroutines adds protection for them:

	func (handler *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
		safely.Go(func() {
			// do something in parallel, panics will be recovered.
		})
	}
*/
package safely

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
)

// DefaultSender is used by the global Go function to spawn goroutines.
var DefaultSender = &Sender{&stackWriter{os.Stderr}}

// PanicHandler is an object that can deal appropriately
// with a panic in a spawned goroutine.
type PanicHandler interface {
	Handle(interface{})
}

// Sender is a struct that can be used to customize the
// safe spawning of goroutines.
type Sender struct {
	handler PanicHandler
}

// NewSender creates a new sender with a specific
// panic handler (which can be nil).
func NewSender(handler PanicHandler) *Sender {
	return &Sender{handler}
}

// SetPanicHandler overwrites the Sender's panic handler.
func (sender *Sender) SetPanicHandler(handler PanicHandler) {
	sender.handler = handler
}

// SetStackWriter sets a panic handler that formats
// and writes the stack trace to an io.Writer.
func (sender *Sender) SetStackWriter(writer io.Writer) {
	sender.SetPanicHandler(&stackWriter{writer})
}

// Go runs the provided function in a new goroutine
// with recovery and panic handling.
func (sender *Sender) Go(function func()) {
	r := sender.recoverer()
	go func(r func()) {
		defer r()
		function()
	}(r)
}

// Go runs its function argument in a separate goroutine, but recovers from any
// panics, optionally writing stack traces to an io.Writer.
func Go(f func()) {
	DefaultSender.Go(f)
}

func (sender *Sender) recoverer() func() {
	if sender.handler == nil {
		return func() {
			recover()
		}
	}
	return func() {
		sender.handler.Handle(recover())
	}
}

type stackWriter struct {
	io.Writer
}

func (sw *stackWriter) Handle(msg interface{}) {
	stack := getStack()
	fmt.Fprintf(sw, "safely caught panic: %s\n%s", msg, stack)
}

func getStack() string {
	l := 4096
	for {
		b := make([]byte, l)
		n := runtime.Stack(b, false)
		if n < l {
			st := string(b[:n])
			//return st
			sp := strings.SplitAfter(st, "\n")
			return sp[0] + strings.Join(sp[7:len(sp)-5], "")
		}
		l *= 2
	}
}
