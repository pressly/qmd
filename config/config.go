package config

import (
	"errors"
	"os"

	"github.com/BurntSushi/toml"
)

var ErrNoConfFile = errors.New("no configuration file specified")

// Config holds configuration read from config file.
type Config struct {
	Bind        string      `toml:"bind"`
	URL         string      `toml:"url"`
	ScriptDir   string      `toml:"script_dir"`
	WorkDir     string      `toml:"work_dir"`
	StoreDir    string      `toml:"store_dir"`
	MaxJobs     int         `toml:"max_jobs"`
	MaxExecTime int         `toml:"max_exec_time"`
	DB          DBConfig    `toml:"db"`
	Queue       QueueConfig `toml:"queue"`
	Slack       SlackConfig `toml:"slack"`
}

type DBConfig struct {
	RedisURI string `toml:"redis_uri"`
}

type QueueConfig struct {
	DisqueURI string `toml:"disque_uri"`
}

type SlackConfig struct {
	Enabled    bool   `toml:"enabled"`
	WebhookURL string `toml:"webhook_url"`
	Channel    string `toml:"channel"`
}

// New reads configuration from a specified file and creates new Config object.
func New(file string) (*Config, error) {
	if file == "" {
		return nil, ErrNoConfFile
	}
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return nil, err
	}

	conf := &Config{}
	if _, err := toml.DecodeFile(file, &conf); err != nil {
		return nil, err
	}

	return conf, nil
}
