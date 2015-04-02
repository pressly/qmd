package job

import (
	"github.com/pressly/qmd/config"
)

type Controller struct {
	WorkDir string
	Jobs    []*Job
}

func NewController(conf *config.Config) (*Controller, error) {
	ctl := &Controller{}

	//TODO: Check the actual path. Create directory etc.
	ctl.WorkDir = conf.WorkDir

	return ctl, nil
}

func (c *Controller) Run() (*Job, error) {
	select {}
}
