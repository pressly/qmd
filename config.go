package main

import "fmt"

type Config struct {
	Topic        string
	ListenOnAddr string
	QueueAddr    string
	RedisAddr    string
	Auth         authConfig    `toml:"auth"`
	Logging      loggingConfig `toml:"logging"`
	Worker       workerConfig  `toml:"worker"`
}

type authConfig struct {
	Enabled    bool
	Username   string
	Password   string
	AuthString string
}

type loggingConfig struct {
	LogLevel    string
	LogBackends []string
}

type workerConfig struct {
	Channel    string
	Throughput int
	ScriptDir  string
	WorkingDir string
	StoreDir   string
	WhiteList  string
	KeepTemp   bool
}

func (c *Config) Setup() (err error) {
	// Setup auth
	c.Auth.AuthString = fmt.Sprintf("%s:%s", config.Auth.Username, config.Auth.Password)

	// Setup logger
	log.Info("woooot %s", c.Logging.LogLevel)

	return
}
