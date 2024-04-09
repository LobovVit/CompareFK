package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Мode      string `yaml:"mode"  env-required:"true"`
	Masterdsn string `yaml:"masterdsn" env-required:"true"`
	Slavedsn  string `yaml:"slavedsn" env-required:"true"`
	Attrs     string `yaml:"attrs"`
	LogLevel  string `yaml:"loglevel"`
}

func GetConfig() (*Config, error) {
	log.Print("read config")
	config := &Config{}
	if err := cleanenv.ReadConfig("config.yml", config); err != nil {
		return nil, err
	}
	return config, nil
}
