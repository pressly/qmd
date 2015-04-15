package qmd

import "github.com/pressly/qmd/config"

type Qmd struct {
	Config *config.Config

	scripts Scripts

	//	queue
}

func New(conf *config.Config) *Qmd {
	return &Qmd{
		Config: conf,
	}
}

func (qmd *Qmd) Close() {
	return
}

func (qmd *Qmd) GetScript(file string) (string, error) {
	return qmd.scripts.Get(file)
}

func (qmd *Qmd) WatchScripts() {
	qmd.scripts.Watch(qmd.Config.ScriptDir)
}
