package config

import (
	"errors"
	"flag"
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

var Cfg *Config

type Config struct {
	Мode       string `yaml:"mode"`
	Masterdsn  string `yaml:"masterdsn"`
	Slavedsn   string `yaml:"slavedsn"`
	LogLevel   string `yaml:"loglevel"`
	Limit      int    `yaml:"limit"`
	RateLimit  int    `yaml:"ratelimit"`
	MasterSQL  string `yaml:"mastersql"`
	SlaveSQL   string `yaml:"slavesql"`
	ConfigFile string
}

func Initialize() error {
	cfg, err := getConfig()
	Cfg = cfg
	return err
}

func getConfig() (*Config, error) {
	log.Print("read config")
	cfgFile := flag.String("c", "config.yml", "путь к конфиг файлу")
	flag.Parse()
	config := &Config{}
	if err := cleanenv.ReadConfig(*cfgFile, config); err != nil {
		return nil, err
	}

	config.ConfigFile = *cfgFile

	if !(config.Мode == "difference" || config.Мode == "intersection") {
		return nil, errors.New("укажите mode: \"difference\" или \"intersection\"")
	}

	if config.Masterdsn == "" {
		return nil, errors.New("укажите Masterdsn: \"postgresql://<login>:<password>@<host>:<port>/<sid>\"")
	}

	if config.Slavedsn == "" {
		return nil, errors.New("укажите Slavedsn: \"postgresql://<login>:<password>@<host>:<port>/<sid>\"")
	}

	if config.LogLevel == "" {
		config.LogLevel = "info"
	}

	if config.Limit == 0 {
		config.Limit = 10000
	}

	if config.RateLimit == 0 {
		config.RateLimit = 5
	}

	if config.MasterSQL == "" {
		config.MasterSQL = "./Master/"
	}

	if config.SlaveSQL == "" {
		config.SlaveSQL = "./Slave.sql"
	}

	return config, nil
}
