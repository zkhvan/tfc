package cmdutil

import (
	"github.com/zkhvan/tfc/internal/tfc"
	"github.com/zkhvan/tfc/pkg/iolib"
	"github.com/zkhvan/tfc/pkg/tfconfig"
)

type Factory struct {
	ExecutableName string
	AppVersion     string

	IOStreams *iolib.IOStreams
	Clock     *Clock

	Editor          func() *Editor
	TFEClient       func() (*tfc.Client, error)
	TerraformConfig func() *tfconfig.TerraformConfig
}
