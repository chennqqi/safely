package safely

import (
	"strings"
	"testing"
	"time"
)

func TestDoesntPanic(t *testing.T) {
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

	Go(failer, r)
	time.Sleep(10 * time.Millisecond)

	lines := strings.Split(string(*r), "\n")

	if lines[0] != "safely caught a panic: failer" {
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
