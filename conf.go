package qmd

import (
	"flag"
	"os"
	"runtime"
	"strconv"

	"github.com/BurntSushi/toml"
)

func Boot(cf *Conf) {
	conf = cf // conf is a package variable
	conf.Setup()

	var err error
	queue, err = NewQueue(cf.Queue) // maybe move this somewhere later...?
	if err != nil {
		lg.Fatal("broken...", err)
	}
}

// func Configure(cf *Conf) {
// 	conf = cf // conf is a package variable
// 	conf.Setup()
// }

func NewConf(confFile string, confEnv string, flags *flag.FlagSet) (*Conf, error) {
	cf := &Conf{}

	// Load conf file
	if confFile == "" {
		confFile = confEnv
		if _, err := os.Stat(confFile); os.IsNotExist(err) {
			confFile = ""
		}
	}
	if confFile != "" {
		if _, err := toml.DecodeFile(confFile, &cf); err != nil {
			return nil, err
		}
	}

	// TODO: .. these flag settings should only be set if they are from the user..
	// not just the default flag value.... ***

	// Flag settings
	cf.Bind = flags.Lookup("bind").Value.String()
	cf.MaxProcs, _ = strconv.Atoi((flags.Lookup("max-procs").Value.String()))

	logLevel := flags.Lookup("log-level").Value.String()
	if cf.Logging == nil {
		cf.Logging = &LoggingConf{LogLevel: logLevel}
	} else {
		cf.Logging.LogLevel = logLevel
	}

	return cf, nil
}

type Conf struct {
	Bind     string       `toml:"bind"`
	MaxProcs int          `toml:"max_procs"`
	Logging  *LoggingConf `toml:"logging"`

	Queue *QueueConf `toml:"queue"`
}

func (cf *Conf) Setup() {
	// var err error
	lg.Info("** QMD Server v%s **", VERSION)

	// Defaults
	if cf.MaxProcs == 0 {
		cf.MaxProcs = runtime.NumCPU()
	}
	runtime.GOMAXPROCS(cf.MaxProcs)

	// Logging
	if cf.Logging == nil {
		cf.Logging = &LoggingConf{}
	}
	cf.Logging.Setup()

	// Queue
	cf.Queue.Setup()

	// Announce settings
	lg.Info(" - maxProcs: %d", cf.MaxProcs)
	lg.Info(" - bind: %s", cf.Bind)
	lg.Info(" - log level: %s", cf.Logging.LogLevel)
}
