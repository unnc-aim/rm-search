package service

type Config struct {
	DataSource    string        `yaml:"DataSource"`
	ElasticConfig ElasticConfig `yaml:"ElasticConfig"`
}

type ElasticConfig struct {
	Addresses []string `yaml:"Addresses"`
	APIKey    string   `yaml:"APIKey"`
}
