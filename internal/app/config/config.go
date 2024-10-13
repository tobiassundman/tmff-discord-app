package config

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	QueryTimeout   string `yaml:"queryTimeout"`
	DBFile         string `yaml:"dbFile"`
	MaxGameAgeDays int    `yaml:"maxGameAgeDays"`
	CurrentSeason  string `yaml:"currentSeason"`
	EloKFactor     int    `yaml:"eloKFactor"`
	AppID          string `yaml:"appID"`
	DiscordToken   string `yaml:"discordToken"`
	PublicKey      string `yaml:"publicKey"`
}

func ReadConfig(configFile string) (*Config, error) {
	conf := &Config{}
	yamlFile, err := os.ReadFile(configFile)
	if err != nil {
		return nil, errors.Wrap(err, "could not open config.yaml")
	}
	err = yaml.Unmarshal(yamlFile, conf)
	if err != nil {
		return nil, errors.Wrap(err, "could not read config.yaml")
	}

	return conf, err
}
