package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

var Config ConfigInstance

type ConfigInstance struct {
	DHCP DHCP `yaml:"DHCP"`
}

type DHCP struct {
	Interface         string            `yaml:"Interface"`
	StartAddr         string            `yaml:"StartAddr"`
	EndAddr           string            `yaml:"EndAddr"`
	DomainName        string            `yaml:"DomainName"`
	NameServers       []string          `yaml:"NameServers"`
	LeaseTTL          int               `yaml:"LeaseTTL"`
	SubnetMask        string            `yaml:"SubnetMask"`
	Router            string            `yaml:"Router"`
	ReservedAddresses map[string]string `yaml:"ReservedAddresses"`
}

func LoadConfig(filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}

	defer f.Close()
	return yaml.NewDecoder(f).Decode(&Config)
}
