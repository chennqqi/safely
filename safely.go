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
		}, os.Stderr)
	}
*/
package safely

import (
	"fmt"
	"io"
	"runtime"
	"strings"
)

func getStack() string {
	l := 4096
	for {
		b := make([]byte, l)
		n := runtime.Stack(b, false)
		if n < l {
			st := string(b[:n])
			sp := strings.SplitAfter(st, "\n")
			return sp[0] + strings.Join(sp[5:len(sp)-5], "")
		}
		l *= 2
	}
}

func recoverer(stackTraceTo io.Writer) func() {
	if stackTraceTo == nil {
		return func() {
			recover()
		}
	}
	return func() {
		if r := recover(); r != nil {
			fmt.Fprintf(
				stackTraceTo,
				"safely caught panic: %s\n%s",
				r,
				getStack(),
			)
		}
	}
}

// Go runs its function argument in a separate goroutine, but recovers from any
// panics, optionally writing stack traces to an io.Writer.
func Go(f func(), stackTraceTo io.Writer) {
	r := recoverer(stackTraceTo)
	go func(r func()) {
		defer r()
		f()
	}(r)
}
