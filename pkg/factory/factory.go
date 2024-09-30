package factory

import (
	"context"

	"github.com/sethvargo/go-envconfig"
	"github.com/zkhvan/tfc/pkg/cmdutil"
)

type Config struct {
	Hostname string `env:"TFE_HOSTNAME,default=app.terraform.io"`
	Address  string `env:"TFE_ADDRESS,default=https://$TFE_HOSTNAME"`
	Token    string `env:"TFE_TOKEN,required"`
}

func New() (*cmdutil.Factory, error) {
	f := &cmdutil.Factory{}

	var cfg Config
	if err := envconfig.Process(context.Background(), &cfg); err != nil {
		return nil, err
	}

	f.Hostname = cfg.Hostname
	f.Address = cfg.Address
	f.Token = cfg.Token

	return f, nil
}
