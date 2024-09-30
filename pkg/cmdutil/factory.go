package cmdutil

import (
	"github.com/hashicorp/go-tfe"
	"github.com/zkhvan/tfc/pkg/iolib"
)

type Factory struct {
	ExecutableName string

	IOStreams *iolib.IOStreams

	TFEClient func() (*tfe.Client, error)
}
