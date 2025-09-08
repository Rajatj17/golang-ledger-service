package config

type MongoConfig struct {
	URI      string `yaml:"uri"`
	Timeout  int    `yaml:"timeout"`
	Database string `yaml:"database"`
}
