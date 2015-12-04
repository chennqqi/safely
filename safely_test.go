package safely

import (
	"strings"
	"testing"
	"time"

	"gopkg.in/stack.v1"
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
	time.Sleep(time.Millisecond)

	lines := strings.Split(string(*r), "\n")

	if lines[0] != "safely caught panic: failer" {
		t.Fatalf("wrong first line: '%s'", lines[0])
	}
}

func TestHandlerDoesntRunInAbsenseOfPanic(t *testing.T) {
	ran := false
	handler := func(obj interface{}, _ stack.CallStack) {
		ran = true
	}

	Go(func() {}, handler)
	time.Sleep(time.Millisecond)

	if ran {
		t.Fatal("panic handler ran even though main func never paniced?")
	}
}
