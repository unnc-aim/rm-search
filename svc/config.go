package svc

import (
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Driver        string          `yaml:"Driver"`
	DataSource    string          `yaml:"DataSource"`
	ElasticConfig ElasticConfig   `yaml:"ElasticConfig"`
	TikaHost      string          `yaml:"TikaHost"`
	AdminToken    string          `yaml:"AdminToken"`
	SearchLog     SearchLogConfig `yaml:"SearchLog"`
	Proxy         ProxyConfig     `yaml:"Proxy"`
	DJIMetaKey    string          `yaml:"DJIMetaKey"`
}

type ElasticConfig struct {
	Addresses []string `yaml:"Addresses"`
	APIKey    string   `yaml:"APIKey"`
	Username  string   `yaml:"Username"`
	Password  string   `yaml:"Password"`
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
	confBytes, err := io.ReadAll(conf)
	if err != nil {
		panic(err)
	}
	var c Config
	err = yaml.Unmarshal(confBytes, &c)
	if err != nil {
		panic(err)
	}

	return c
}
