package config

import (
	"os"

	"github.com/nico151999/high-availability-expense-splitter/pkg/param"
	"github.com/rotisserie/eris"
	"gopkg.in/yaml.v3"
)

var ErrMissingSrvCfgPath = eris.New("the path to the server config was not passed")

// LoadConfig loads a yaml file from the path passed as parameter
func LoadConfig[CONFIG any](file string) (*CONFIG, error) {
	// Viper is too much for what we need, so we only read the file and unmarshal the contents
	var data []byte
	{
		var err error
		if data, err = os.ReadFile(file); err != nil {
			return nil, eris.Wrapf(err, "failed reading config file from %s", file)
		}
	}
	var cfg CONFIG
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, eris.Wrap(err, "unable to decode into config struct")
	}
	return &cfg, nil
}

func LoadConfigFromParam[CONFIG any](cfgFlag *param.StringParam) (*CONFIG, error) {
	if !cfgFlag.IsSet() {
		return nil, ErrMissingSrvCfgPath
	}
	cfgPath := cfgFlag.String()

	cfg, err := LoadConfig[CONFIG](cfgPath)
	if err != nil {
		return nil, eris.Wrap(
			err,
			"failed getting config file",
		)
	}
	return cfg, nil
}
