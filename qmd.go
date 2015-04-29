package qmd

import (
	"log"
	"sync"

	"github.com/pressly/qmd/config"
)

type Qmd struct {
	Config *config.Config
	DB     *DB

	Scripts Scripts
	Queue   chan *Job

	Workers chan Worker

	MuJobs sync.Mutex
	Jobs   map[string]*Job

	Closing chan struct{}
}

func New(conf *config.Config) (*Qmd, error) {
	db, err := NewDB(conf.DB.RedisURI)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &Qmd{
		Config:  conf,
		Queue:   make(chan *Job, 4096),
		Jobs:    make(map[string]*Job),
		Closing: make(chan struct{}, 1),
		DB:      db,
	}, nil
}

func (qmd *Qmd) Close() {
	log.Printf("qmd.Close()")
	qmd.Closing <- struct{}{}
	log.Fatalf("exit")
}

func (qmd *Qmd) GetScript(file string) (string, error) {
	return qmd.Scripts.Get(file)
}

func (qmd *Qmd) WatchScripts() {
	qmd.Scripts.Watch(qmd.Config.ScriptDir)
}
