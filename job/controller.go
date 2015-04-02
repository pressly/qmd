package job

import (
	"github.com/pressly/qmd/config"
)

var Ctl *Controller

type Controller struct {
	WorkDir string
	Jobs    []*Job
}

func NewController(conf *config.Config) (*Controller, error) {
	Ctl = &Controller{}

	//TODO: Check the actual path. Create directory etc.
	Ctl.WorkDir = conf.WorkDir

	return Ctl, nil
}

func (c *Controller) Run() (*Job, error) {
	select {}
}
