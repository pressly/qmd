package main

import (
	"fmt"
	"os"
	"path/filepath"

	stdlog "log"

	"github.com/BurntSushi/toml"
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
	Channel         string
	QueueAddresses  []string
	LookupAddresses []string
	Throughput      int
	ScriptDir       string
	WorkingDir      string
	StoreDir        string
	WhiteList       string
	KeepTemp        bool
}

// Find and load config file by path
func (c *Config) Load(p string) error {
	var err error

	if p == "" {
		p = "./config.toml"
	}
	p, err = filepath.Abs(p)
	if err != nil {
		return err
	}

	if _, err = os.Stat(p); os.IsNotExist(err) {
		return fmt.Errorf("No config file found at %s", p)
	}

	_, err = toml.DecodeFile(p, &c)
	if err != nil {
		return err
	}
	return nil
}

func (c *Config) Setup() error {
	// Clean paths
	err := c.fixPaths()
	if err != nil {
		return err
	}

	// Setup logger
	logging.SetFormatter(logging.MustStringFormatter("%{level} %{message}"))

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

	logLevel, err := logging.LogLevel(config.Logging.LogLevel)
	if err != nil {
		return err
	}
	logging.SetLevel(logLevel, "qmd")

	// Redirect standard logger
	stdlog.SetOutput(&logProxyWriter{})

	// Setup auth
	c.Auth.AuthString = fmt.Sprintf("%s:%s", config.Auth.Username, config.Auth.Password)

	return nil
}

func (c *Config) fixPaths() error {
	var err error

	c.Worker.ScriptDir, err = filepath.Abs(c.Worker.ScriptDir)
	if err != nil {
		return err
	}
	c.Worker.WorkingDir, err = filepath.Abs(c.Worker.WorkingDir)
	if err != nil {
		return err
	}
	c.Worker.StoreDir, err = filepath.Abs(c.Worker.StoreDir)
	if err != nil {
		return err
	}
	return nil
}

type logProxyWriter struct{}

func (l *logProxyWriter) Write(p []byte) (n int, err error) {
	log.Info("%s", p)
	return len(p), nil
}
