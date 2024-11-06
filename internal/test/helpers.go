package test

import (
	"bytes"
	"slices"
	"testing"
)

func BufferEmpty(t *testing.T, got *bytes.Buffer) {
	t.Helper()

	if got := got.String(); len(got) > 0 {
		t.Errorf("buffer not empty, got:\n%s", got)
	}
}

func Buffer(t *testing.T, got *bytes.Buffer, want string) {
	t.Helper()

	if got := got.String(); got != want {
		t.Errorf("buffer got:\n%swant:\n%s", got, want)
	}
}

func StringSlice(t *testing.T, got, want []string) {
	t.Helper()

	if !slices.Equal(got, want) {
		t.Errorf("got: %v, want %v", got, want)
	}
}
