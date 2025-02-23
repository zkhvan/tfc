package test

import (
	"bytes"
	"net/http"
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
		t.Errorf("string slice got: %v, want %v", got, want)
	}
}

func PathValue(t *testing.T, r *http.Request, path, want string) {
	t.Helper()

	got := r.PathValue(path)
	if got != want {
		t.Errorf("request path value got: %v, want %v", got, want)
	}
}
