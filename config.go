package main

type Config struct {
	Topic        string
	ListenOnAddr string
	QueueAddr    string
	RedisAddr    string
	Username     string
	Password     string
	logging      loggingConfig `toml:"logging"`
	Worker       workerConfig  `toml:"worker"`
}

type loggingConfig struct {
	LogLevel    string
	LogBackends map[string]string
}

func (l *loggingConfig) setup() {
	log.Info("yes %s", l.LogLevel)
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
