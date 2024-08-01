package config

import (
	"errors"
	"flag"
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Мode      string `yaml:"mode"`
	Masterdsn string `yaml:"masterdsn"`
	Slavedsn  string `yaml:"slavedsn"`
	LogLevel  string `yaml:"loglevel"`
	Limit     int    `yaml:"limit"`
	RateLimit int    `yaml:"ratelimit"`
	MasterSQL string `yaml:"mastersql"`
	SlaveSQL  string `yaml:"slavesql"`
	LogFile   string `yaml:"logfile"`
	ResFile   string `yaml:"resfile"`
}

func GetConfig() (*Config, error) {
	log.Print("read config")
	cfgFile := flag.String("mode", "config.yml", "путь к конфиг файлу")
	flag.Parse()
	config := &Config{}
	if err := cleanenv.ReadConfig(*cfgFile, config); err != nil {
		return nil, err
	}

	if config.Мode == "" {
		return nil, errors.New("Укажите mode: \"difference\" или \"intersection\"")
	}

	if config.Masterdsn == "" {
		return nil, errors.New("Укажите Masterdsn: \"postgresql://<login>:<password>@<host>:<port>/<sid>\"")
	}

	if config.Slavedsn == "" {
		return nil, errors.New("Укажите Slavedsn: \"postgresql://<login>:<password>@<host>:<port>/<sid>\"")
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
		config.MasterSQL = "./Master1.sql"
	}

	if config.SlaveSQL == "" {
		config.SlaveSQL = "./Slave1.sql"
	}

	if config.LogFile == "" {
		config.LogFile = "./Slave1.sql"
	}

	if config.ResFile == "" {
		config.ResFile = "./Slave1.sql"
	}

	return config, nil
}
