package cmdutil

import (
	"github.com/hashicorp/go-tfe"

	"github.com/zkhvan/tfc/pkg/iolib"
)

type Factory struct {
	ExecutableName string
	AppVersion     string

	IOStreams *iolib.IOStreams
	Clock     *Clock

	TFEClient func() (*tfe.Client, error)
}
