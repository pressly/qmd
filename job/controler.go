package job

import (
	"github.com/pressly/qmd/config"
)

type Controller struct {
	Jobs []*Job
}

func NewController(conf *config.Config) (*Controller, error) {
	ctl := &Controller{}

	return ctl, nil
}

func (c *Controller) Run() (*Job, error) {
	job := &Job{}

	return job, nil
}
