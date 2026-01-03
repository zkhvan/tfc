package factory

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-tfe"
	"github.com/sethvargo/go-envconfig"

	"github.com/zkhvan/tfc/internal/tfc"
	"github.com/zkhvan/tfc/pkg/cmdutil"
	"github.com/zkhvan/tfc/pkg/credentials"
	"github.com/zkhvan/tfc/pkg/iolib"
)

type Config struct {
	Hostname string `env:"TFE_HOSTNAME,default=app.terraform.io"`
	Address  string `env:"TFE_ADDRESS"`
	Token    string `env:"TFE_TOKEN"`
}

func (c *Config) GetAddress() string {
	if len(c.Address) > 0 {
		return c.Address
	}

	return fmt.Sprintf("https://%s", c.Hostname)
}

func New(appVersion string) (*cmdutil.Factory, error) {
	f := &cmdutil.Factory{
		ExecutableName: "tfc",
		AppVersion:     appVersion,
		Clock:          cmdutil.NewClock(nil),
	}

	f.IOStreams = ioStreams(f)
	f.Editor = editorFunc(f)
	f.TFEClient = tfeClientFunc(f)

	return f, nil
}

func ioStreams(_ *cmdutil.Factory) *iolib.IOStreams {
	return iolib.System()
}

func editorFunc(f *cmdutil.Factory) func() *cmdutil.Editor {
	return func() *cmdutil.Editor {
		return cmdutil.NewEditor(f.IOStreams)
	}
}

func tfeClientFunc(_ *cmdutil.Factory) func() (*tfc.Client, error) {
	return func() (*tfc.Client, error) {
		var cfg Config
		if err := envconfig.Process(context.Background(), &cfg); err != nil {
			return nil, err
		}

		tfeCfg := tfe.DefaultConfig()
		tfeCfg.Address = cfg.GetAddress()
		tfeCfg.Token = cfg.Token

		if len(tfeCfg.Token) == 0 {
			token, err := credentials.GetTokenForHost(cfg.Hostname)
			if err != nil {
				return nil, fmt.Errorf("error getting token from credentials file: %w", err)
			}
			tfeCfg.Token = token
		}

		if len(tfeCfg.Token) == 0 {
			return nil, fmt.Errorf("no tfe token found")
		}

		client, err := tfe.NewClient(tfeCfg)
		if err != nil {
			return nil, fmt.Errorf("error creating tfe client: %w", err)
		}

		return tfc.NewClient(client), nil
	}
}
