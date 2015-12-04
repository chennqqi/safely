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
		}, nil)
	}
*/
package safely

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/stack.v1"
)

// DefaultPanicHandler is used by Go when the second argument is nil.
var DefaultPanicHandler = StackWriter(os.Stderr)

// PanicHandler is a func that can deal appropriately
// with panics from spawned goroutine.
type PanicHandler func(interface{}, stack.CallStack)

// Go runs its first argument in a separate goroutine, but recovers from any
// panics with the provided PanicHandler (using DefaultPanicHandler if nil).
func Go(f func(), h PanicHandler) {
	if h == nil {
		h = DefaultPanicHandler
	}

	go func() {
		defer func() {
			r := recover()
			if r != nil && h != nil {
				h(r, stack.Trace().TrimRuntime()[2:])
			}
		}()

		f()
	}()
}

// StackWriter creates a PanicHandler that dumps a stack trace to the provided
// io.Writer in the event of a panic.
func StackWriter(out io.Writer) PanicHandler {
	return func(obj interface{}, callstack stack.CallStack) {
		fmt.Fprintf(out, "safely caught panic: %s\n%+v", obj, callstack)
	}
}
