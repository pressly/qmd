package job

import (
	"errors"
	"os"

	"github.com/pressly/qmd/config"
)

var Ctl *Controller

type Controller struct {
	WorkDir string
	Waiting chan *Job
	Running chan *Job
}

// NewController creates new Controller instance.
func NewController(conf *config.Config) (*Controller, error) {
	info, err := os.Stat(conf.WorkDir)
	if err != nil {
		return nil, errors.New("work_dir=\"" + conf.WorkDir + "\": " + err.Error())
	}
	if !info.IsDir() {
		return nil, errors.New("work_dir=\"" + conf.WorkDir + "\": not a directory")
	}

	//TODO: Check the actual path. Create directory etc.
	Ctl = &Controller{
		WorkDir: conf.WorkDir,
		Waiting: make(chan *Job),
		Running: make(chan *Job, conf.MaxJobs),
	}

	return Ctl, nil
}

// Run runs the Controller loop.
func (c *Controller) Run() {
	for {
		select {
		//case job, ok := <-c.Waiting:
		}
	}
}

// func (c *Controller) Add(job *Job) error {
// 	c.Waiting <- job
// }

// func (c *Controller) RunCmd(cmd *Cmd) *Job {
// 	job := Job{
// 		Cmd: cmd,
// 	}

// 	job.Run()
// }
