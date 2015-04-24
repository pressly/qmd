package qmd

import (
	"log"
	"sync"

	"github.com/pressly/qmd/config"
)

type Qmd struct {
	Config *config.Config

	Scripts Scripts
	Queue   chan *Job

	MuJobs sync.Mutex
	Jobs   map[string]*Job

	Closing chan struct{}
}

func New(conf *config.Config) *Qmd {
	return &Qmd{
		Config:  conf,
		Queue:   make(chan *Job, 4096),
		Jobs:    make(map[string]*Job),
		Closing: make(chan struct{}, 1),
	}
}

func (qmd *Qmd) Close() {
	log.Printf("qmd.Close()")
	qmd.Closing <- struct{}{}
}

func (qmd *Qmd) GetScript(file string) (string, error) {
	return qmd.Scripts.Get(file)
}

func (qmd *Qmd) WatchScripts() {
	qmd.Scripts.Watch(qmd.Config.ScriptDir)
}
