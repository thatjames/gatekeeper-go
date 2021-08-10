package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

var Config *ConfigInstance

type ConfigInstance struct {
	DNS *DNS `yaml:"dns"`
}

type DNS struct {
	ListenAddr string `yaml:""`
}

func LoadConfig(filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}

	defer f.Close()
	Config = new(ConfigInstance)
	return yaml.NewDecoder(f).Decode(Config)
}
