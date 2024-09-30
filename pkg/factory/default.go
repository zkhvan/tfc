package factory

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/go-tfe"
	"github.com/sethvargo/go-envconfig"

	"github.com/zkhvan/tfc/pkg/cmdutil"
	"github.com/zkhvan/tfc/pkg/iolib"
)

type Config struct {
	Hostname string `env:"TFE_HOSTNAME,default=app.terraform.io"`
	Address  string `env:"TFE_ADDRESS,default=https://$TFE_HOSTNAME"`
	Token    string `env:"TFE_TOKEN,required"`
}

func New() (*cmdutil.Factory, error) {
	f := &cmdutil.Factory{
		ExecutableName: "tfc",
	}

	f.IOStreams = ioStreams(f)
	f.TFEClient = tfeClientFunc(f)

	return f, nil
}

func ioStreams(_ *cmdutil.Factory) *iolib.IOStreams {
	return &iolib.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}
}

func tfeClientFunc(_ *cmdutil.Factory) func() (*tfe.Client, error) {
	return func() (*tfe.Client, error) {
		var cfg Config
		if err := envconfig.Process(context.Background(), &cfg); err != nil {
			return nil, err
		}

		tfeCfg := tfe.DefaultConfig()
		tfeCfg.Address = cfg.Address
		tfeCfg.Token = cfg.Token

		client, err := tfe.NewClient(tfeCfg)
		if err != nil {
			return nil, fmt.Errorf("error creating tfe client: %w", err)
		}

		return client, nil
	}
}
