package config

import (
	"errors"
	"os"
	"runtime"

	"github.com/BurntSushi/toml"
)

var ErrNoConfFile = errors.New("no configuration file specified")

// Config holds configuration read from config file.
type Config struct {
	Environment string `toml:"environment"`
	Bind        string `toml:"bind"`
	MaxProcs    int    `toml:"max_procs"`
	DebugMode   bool   `toml:"debug_mode"`
	WorkDir     string `toml:"work_dir"`
	StoreDir    string `toml:"store_dir"`
	MaxJobs     int    `toml:"max_jobs"`
	MaxExecTime int    `toml:"max_exec_time"`
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

	if conf.MaxProcs <= 0 {
		conf.MaxProcs = runtime.NumCPU()
	}

	return conf, nil
}
