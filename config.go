package qmd

import (
	"fmt"
	"path/filepath"
	"time"
)

type baseConfig struct {
	Name    string         `toml:"name"`
	Queue   *QueueConfig   `toml:"queue"`
	Logging *LoggingConfig `toml:"logging"`
}

type ServerConfig struct {
	baseConfig

	// Basic Options
	ListenOnAddr string        `toml:"address"`
	TTL          time.Duration `toml:"ttl"`
	DBAddr       string        `toml:"database_address"`
	AdminAddr    string        `toml:"admin_address"`

	// Auth Options
	DisableAuth bool   `toml:"disable_auth"`
	Username    string `toml:"username"`
	Password    string `toml:"password"`
}

func (sc *ServerConfig) Clean() error {
	if sc.Name == "" {
		sc.Name = fmt.Sprintf("server-%s", NewID())
	}
	if sc.TTL <= 0 {
		sc.TTL = 5 * time.Minute
	}
	if !sc.DisableAuth {
		if sc.Username == "" || sc.Password == "" {
			return fmt.Errorf("Either username and password are missing")
		}
	}
	if err := sc.Queue.Clean(); err != nil {
		return err
	}
	sc.Logging.Clean()
	return nil
}

type WorkerConfig struct {
	baseConfig

	Throughput int    `toml:"throughput"`
	ScriptDir  string `toml:"script_dir"`
	WorkingDir string `toml:"working_dir"`
	StoreDir   string `toml:"store_dir"`
	Whitelist  string `toml:"whitelist"`
	KeepTemp   bool   `toml:"keep_temp"`
}

func (wc *WorkerConfig) Clean() error {
	var err error

	if wc.Name == "" {
		wc.Name = fmt.Sprintf("worker-%s", NewID())
	}

	// Fix paths
	wc.ScriptDir, err = filepath.Abs(wc.ScriptDir)
	if err != nil {
		return err
	}
	wc.WorkingDir, err = filepath.Abs(wc.WorkingDir)
	if err != nil {
		return err
	}
	wc.StoreDir, err = filepath.Abs(wc.StoreDir)
	if err != nil {
		return err
	}
	wc.Whitelist, err = filepath.Abs(wc.Whitelist)
	if err != nil {
		return err
	}

	if err := wc.Queue.Clean(); err != nil {
		return err
	}
	wc.Logging.Clean()
	return nil
}
