package svc

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DataSource  string            `yaml:"DataSource"`
	MeiliSearch MeiliSearchConfig `yaml:"MeiliSearch"`
	TikaHost    string            `yaml:"TikaHost"`
	SearchLog   SearchLogConfig   `yaml:"SearchLog"`
	Proxy       ProxyConfig       `yaml:"Proxy"`
	DJIMetaKey  string            `yaml:"DJIMetaKey"`
}

type MeiliSearchConfig struct {
	Address string `yaml:"Address"`
	APIKey  string `yaml:"APIKey"`
}

type SearchLogConfig struct {
	Enabled              bool `yaml:"Enabled"`
	DisabledRequestBody  bool `yaml:"DisabledRequestBody"`
	DisabledResponseBody bool `yaml:"DisabledResponseBody"`
}

type ProxyConfig struct {
	Enabled bool   `yaml:"Enabled"`
	URL     string `yaml:"URL"`
}

func ReadConfig(path string) Config {
	conf, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer conf.Close()
	var c Config
	err = yaml.NewDecoder(conf).Decode(&c)
	if err != nil {
		panic(err)
	}

	return c
}
