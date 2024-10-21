package tfetest

import "bytes"

type CmdOut struct {
	OutBuf *bytes.Buffer
	ErrBuf *bytes.Buffer
}
