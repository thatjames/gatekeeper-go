package config

import (
	"bytes"
	"fmt"
	"os"
	"reflect"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

var loadedFilePath *string

var Config *ConfigInstance

type Auth interface {
	Type() string
}

type BaseAuth struct {
	AuthType string `yaml:"AuthType"`
}

func (b BaseAuth) Type() string {
	return b.AuthType
}

type ConfigInstance struct {
	DHCP *DHCP `yaml:"DHCP"`
	Web  *Web  `yaml:"Web"`
	DNS  *DNS  `yaml:"DNS"`
	Auth Auth  `yaml:"Auth"`
}

func (c *ConfigInstance) UnmarshalYAML(value *yaml.Node) error {
	// Alias that excludes Auth so inline doesn't conflict with our yaml.Node capture
	type rawConfig struct {
		DHCP *DHCP `yaml:"DHCP"`
		Web  *Web  `yaml:"Web"`
		DNS  *DNS  `yaml:"DNS"`
	}

	var base struct {
		rawConfig `yaml:",inline"`
		Auth      yaml.Node `yaml:"Auth"`
	}

	if err := value.Decode(&base); err != nil {
		return err
	}

	c.DHCP = base.DHCP
	c.Web = base.Web
	c.DNS = base.DNS

	if base.Auth.IsZero() {
		return nil
	}

	var authBase BaseAuth
	if err := base.Auth.Decode(&authBase); err != nil {
		return err
	}

	switch authBase.AuthType {
	case "oidc":
		var oidcAuth OIDCAuth
		if err := base.Auth.Decode(&oidcAuth); err != nil {
			return err
		}
		c.Auth = &oidcAuth
	default:
		return fmt.Errorf("unknown AuthType: %q", authBase.AuthType)
	}

	return nil
}

func (c ConfigInstance) String() string {
	return printStruct("", c)
}

func printStruct(prefix string, iface interface{}) string {
	buff := new(bytes.Buffer)

	refVal := reflect.ValueOf(iface)
	for refVal.Kind() == reflect.Ptr || refVal.Kind() == reflect.Interface {
		if refVal.IsNil() {
			return ""
		}
		refVal = refVal.Elem()
	}

	if refVal.Kind() != reflect.Struct {
		fmt.Fprintf(buff, "%s - %v\n", prefix, refVal.Interface())
		return buff.String()
	}

	refType := refVal.Type()

	for i := 0; i < refType.NumField(); i++ {
		fieldType := refType.Field(i)
		field := refVal.Field(i)

		// Inline anonymous/embedded fields instead of nesting them
		if fieldType.Anonymous {
			for field.Kind() == reflect.Ptr || field.Kind() == reflect.Interface {
				if field.IsNil() {
					break
				}
				field = field.Elem()
			}
			if field.Kind() == reflect.Struct {
				fmt.Fprint(buff, printStruct(prefix, field.Interface()))
			}
			continue
		}

		for field.Kind() == reflect.Ptr || field.Kind() == reflect.Interface {
			if field.IsNil() {
				break
			}
			field = field.Elem()
		}

		switch field.Kind() {
		case reflect.Ptr, reflect.Interface:
			continue
		case reflect.Struct:
			fmt.Fprintln(buff, prefix+fieldType.Name)
			fmt.Fprintln(buff, printStruct(prefix+" ", field.Interface()))
		case reflect.Map:
			fmt.Fprintln(buff, prefix, "-", fieldType.Name)
			for _, mapKey := range field.MapKeys() {
				fmt.Fprintln(buff, prefix, prefix, "-", mapKey.Interface(), ":", field.MapIndex(mapKey).Interface())
			}
		case reflect.Slice:
			fmt.Fprintln(buff, prefix, "-", fieldType.Name)
			for j := 0; j < field.Len(); j++ {
				fmt.Fprintln(buff, prefix, prefix, "-", field.Index(j))
			}
		default:
			fmt.Fprintf(buff, "%s - %s: %v\n", prefix, fieldType.Name, field.Interface())
		}
	}
	return buff.String()
}

type OIDCAuth struct {
	BaseAuth        `yaml:",inline"`
	IssuerURL       string   `yaml:"IssuerURL"`
	ClientID        string   `yaml:"ClientID"`
	ClientSecretVar string   `yaml:"ClientSecretVar"`
	RedirectURL     string   `yaml:"RedirectURL"`
	Scopes          []string `yaml:"Scopes"`
}

type DefaultAuth struct {
	BaseAuth
}

type DHCP struct {
	Interface         string            `yaml:"Interface"`
	StartAddr         string            `yaml:"StartAddr"`
	EndAddr           string            `yaml:"EndAddr"`
	DomainName        string            `yaml:"DomainName"`
	NameServers       []string          `yaml:"NameServers"`
	LeaseTTL          int               `yaml:"LeaseTTL"`
	SubnetMask        string            `yaml:"SubnetMask"`
	Gateway           string            `yaml:"Gateway"`
	ReservedAddresses map[string]string `yaml:"ReservedAddresses"`
	LeaseFile         string            `yaml:"LeaseFile"`
}

type Web struct {
	Address      string     `yaml:"Address"`
	TLS          *TLSConfig `yaml:"TLS"`
	HTPasswdFile string     `yaml:"HTPasswdFile"`
	Prometheus   bool       `yaml:"Prometheus"`
}

type TLSConfig struct {
	PublicKey  string `yaml:"PublicKey"`
	PrivateKey string `yaml:"PrivateKey"`
}

type DNS struct {
	UpstreamServers []string          `yaml:"UpstreamServers"`
	Interface       string            `yaml:"Interface"`
	LocalDomains    map[string]string `yaml:"LocalDomains"`
	Port            int               `yaml:"Port"`
	BlockLists      []string          `yaml:"BlockLists"`
	BlockedDomains  []string          `yaml:"BlockedDomains"`
}

func LoadConfig(filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	loadedFilePath = &filePath
	defer f.Close()

	if err := yaml.NewDecoder(f).Decode(&Config); err != nil {
		return err
	}

	return nil
}

func UpdateConfig() error {
	if loadedFilePath == nil {
		return fmt.Errorf("no config file path loaded")
	}

	log.Debugf("Updating config to %s", *loadedFilePath)

	// Open file for writing (this will truncate/overwrite the existing file)
	f, err := os.Create(*loadedFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Create encoder that writes to the file
	encoder := yaml.NewEncoder(f)
	encoder.SetIndent(2)
	defer encoder.Close()

	return encoder.Encode(Config)
}
