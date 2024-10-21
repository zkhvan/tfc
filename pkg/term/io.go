package term

import (
	"io"
)

type fileReader interface {
	io.ReadCloser
	Fd() uintptr
}

type fileWriter interface {
	io.Writer
	Fd() uintptr
}

type fdReader struct {
	io.ReadCloser
	fd uintptr
}

func (r *fdReader) Fd() uintptr {
	return r.fd
}

type fdWriter struct {
	io.Writer
	fd uintptr
}

func (r *fdWriter) Fd() uintptr {
	return r.fd
}
