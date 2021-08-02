package config

type ConfigInstance struct {
	DNS *DNS `yaml:"dns"`
}

type DNS struct {
	ListenAddr string `yaml:""`
}
