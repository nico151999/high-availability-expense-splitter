package cors

import (
	"context"

	"github.com/nico151999/high-availability-expense-splitter/internal/config"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/nico151999/high-availability-expense-splitter/pkg/param"
	"github.com/rotisserie/eris"
)

type Cors struct {
	UrlPatterns    []string `yaml:"urlPatterns"`
	AllowedHeaders []string `yaml:"allowedHeaders"`
	AllowedMethods []string `yaml:"allowedMethods"`
}

func LoadCorsFromParam(p *param.StringParam) (*Cors, error) {
	cfg, err := config.LoadConfigFromParam[Cors](p)
	if err != nil {
		return nil, eris.Wrap(
			err,
			"failed getting cors config file",
		)
	}
	return cfg, nil
}

func MustLoadCorsFromParam(ctx context.Context, p *param.StringParam) *Cors {
	cfg, err := LoadCorsFromParam(p)
	if err != nil {
		logging.FromContext(ctx).Panic("failed loading cors from parameter", logging.Error(err))
	}
	return cfg
}
