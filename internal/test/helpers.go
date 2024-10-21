package test

import (
	"bytes"
	"slices"
	"testing"
)

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
