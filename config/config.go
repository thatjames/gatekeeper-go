package config

import (
	"bytes"
	"fmt"
	"os"
	"reflect"

	"gopkg.in/yaml.v2"
)

var Config ConfigInstance

type ConfigInstance struct {
	DHCP *DHCP `yaml:"DHCP"`
	Web  *Web  `yaml:"Web"`
}

func (c ConfigInstance) String() string {
	return printStruct("", c)
}

func printStruct(prefix string, iface interface{}) string {
	buff := new(bytes.Buffer)
	refType := reflect.TypeOf(iface)
	refVal := reflect.ValueOf(iface)
	for i := 0; i < refType.NumField(); i++ {
		switch refVal.Field(i).Kind() {
		case reflect.Ptr, reflect.Interface, reflect.Struct:
			fmt.Fprintln(buff, refType.Field(i).Name)
			val := reflect.Indirect(refVal.Field(i))
			fmt.Fprintln(buff, printStruct(prefix+" ", val.Interface()))

		case reflect.Map:
			fmt.Fprintln(buff, prefix, "-", refType.Field(i).Name)
			for _, mapVal := range refVal.Field(i).MapKeys() {
				v := refVal.Field(i).MapIndex(mapVal)
				fmt.Fprintln(buff, prefix, prefix, "-", mapVal.Interface(), ":", v.Interface())
			}

		case reflect.Slice:
			fmt.Fprintln(buff, prefix, "-", refType.Field(i).Name)
			for j := 0; j < refVal.Field(i).Len(); j++ {
				fmt.Fprintln(buff, prefix, prefix, "-", refVal.Field(i).Index(j))
			}

		default:
			fmt.Fprintf(buff, "%s - %s: %v\n", prefix, refType.Field(i).Name, refVal.Field(i).Interface())
		}
	}

	return buff.String()
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

type Web struct {
	Address string `yaml:"Address"`
}

func LoadConfig(filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}

	defer f.Close()
	return yaml.NewDecoder(f).Decode(&Config)
}
