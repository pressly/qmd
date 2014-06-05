package main

type Config struct {
	Topic        string
	ListenOnAddr string
	QueueAddr    string
	RedisAddr    string
	Worker       worker
}

type worker struct {
	Channel    string
	Throughput int
	ScriptDir  string
	WorkingDir string
}
