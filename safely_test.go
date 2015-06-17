package safely

import (
	"runtime"
	"strings"
	"testing"
)

func TestDoesntPanic(t *testing.T) {
	DefaultPanicHandler = nil
	Go(func() {
		panic("omgOmgOMG")
	}, nil)
}

type recorder []byte

func (r *recorder) Write(b []byte) (int, error) {
	*r = append(*r, b...)
	return len(b), nil
}

func failer() {
	panic("failer")
}

func TestPrintsStack(t *testing.T) {
	r := &recorder{}
	DefaultPanicHandler = StackWriter(r)

	Go(failer, nil)
	runtime.Gosched()

	lines := strings.Split(string(*r), "\n")

	if lines[0] != "safely caught panic: failer" {
		t.Fatalf("wrong first line: '%s'", lines[0])
	}
	if !strings.HasPrefix(lines[1], "goroutine ") ||
		!strings.HasSuffix(lines[1], "[running]:") {
		t.Fatalf("wrong second line: '%s'", lines[1])
	}

	if !strings.HasPrefix(lines[2], "github.com/teepark/safely.failer") {
		t.Fatalf("wrong third line: '%s'", lines[2])
	}
}

func TestHandlerDoesntRunInAbsenseOfPanic(t *testing.T) {
	ran := false
	handler := func(obj interface{}) {
		ran = true
	}

	Go(func() {}, handler)
	runtime.Gosched()

	if ran {
		t.Fatal("panic handler ran even though main func never paniced?")
	}
}
