package main

import "fmt"

type Config struct {
	Topic        string
	ListenOnAddr string
	QueueAddr    string
	RedisAddr    string
	auth         authConfig    `toml:"auth"`
	logging      loggingConfig `toml:"logging"`
	Worker       workerConfig  `toml:"worker"`
}

type authConfig struct {
	Enabled    bool
	Username   string
	Password   string
	authString string
}

type loggingConfig struct {
	LogLevel    string
	LogBackends map[string]string
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

func (c *Config) setup() (err error) {
	// Setup auth
	c.auth.authString = fmt.Sprintf("%s:%s", config.auth.Username, config.auth.Password)

	// Setup logger
	log.Info("woooot %s", c.logging.LogLevel)

	return
}
