package config

import "github.com/kelseyhightower/envconfig"

type Config struct {
	DB Db
}

type Db struct {
	DbUserName     string `envconfig:"MYSQL_USER"`
	DbUserPassword string `envconfig:"MYSQL_PASSWORD"`
	DbHost         string `envconfig:"MYSQL_HOST"`
	DbPort         string `envconfig:"MYSQL_PORT"`
	DbName         string `envconfig:"MYSQL_DATABASE"`
}

func New() (*Config, error) {
	cfg := &Config{}
	if err := envconfig.Process("", cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
