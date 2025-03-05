package svc

import (
	"gopkg.in/yaml.v3"
	"io"
	"os"
)

type Config struct {
	DataSource    string        `yaml:"DataSource"`
	ElasticConfig ElasticConfig `yaml:"ElasticConfig"`
	TikaHost      string        `yaml:"TikaHost"`
}

type ElasticConfig struct {
	Addresses []string `yaml:"Addresses"`
	APIKey    string   `yaml:"APIKey"`
	Username  string   `yaml:"Username"`
	Password  string   `yaml:"Password"`
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
