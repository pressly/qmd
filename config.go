package main

import (
	"fmt"
	"os"

	stdlog "log"

	"github.com/op/go-logging"
)

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

func (c *Config) Setup() error {
	// Setup logger
	logLevel, err := logging.LogLevel(config.Logging.LogLevel)
	if err != nil {
		return err
	}
	logging.SetLevel(logLevel, "qmd")

	var logBackends []logging.Backend
	for _, lb := range config.Logging.LogBackends {
		// TODO: test for starting with / or ./ and treat it
		// as a file logger
		// TODO: case insensitive stdout / syslog
		switch lb {
		case "STDOUT":
			logBackend := logging.NewLogBackend(os.Stdout, "", stdlog.LstdFlags)
			logBackends = append(logBackends, logBackend)
		case "syslog":
			logBackend, err := logging.NewSyslogBackend("qmd")
			if err != nil {
				return err
			}
			logBackends = append(logBackends, logBackend)
		}
	}
	if len(logBackends) > 0 {
		logging.SetBackend(logBackends...)
	}

	// Setup auth
	c.Auth.AuthString = fmt.Sprintf("%s:%s", config.Auth.Username, config.Auth.Password)

	return nil
}
