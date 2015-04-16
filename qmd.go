package qmd

import (
	"sync"

	"github.com/pressly/qmd/config"
)

type Qmd struct {
	Config *config.Config

	Scripts Scripts
	Queue   chan *Job

	muJobs sync.Mutex // guards Jobs
	Jobs   map[string]*Job

	Closing chan struct{}
}

func New(conf *config.Config) *Qmd {
	return &Qmd{
		Config:  conf,
		Queue:   make(chan *Job),
		Jobs:    make(map[string]*Job),
		Closing: make(chan struct{}, 1),
	}
}

func (qmd *Qmd) Close() {
	qmd.Closing <- struct{}{}
}

func (qmd *Qmd) GetScript(file string) (string, error) {
	return qmd.Scripts.Get(file)
}

func (qmd *Qmd) WatchScripts() {
	qmd.Scripts.Watch(qmd.Config.ScriptDir)
}
