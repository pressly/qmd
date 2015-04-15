package qmd

type queue struct {
	WorkDir string
	Waiting chan *Job
	Running chan *Job
	//Done chan
}

// NewController creates new Controller instance.
func (qmd *Qmd) ListenQueue() {
	select {}
	// info, err := os.Stat(conf.WorkDir)
	// if err != nil {
	// 	return nil, errors.New("work_dir=\"" + conf.WorkDir + "\": " + err.Error())
	// }
	// if !info.IsDir() {
	// 	return nil, errors.New("work_dir=\"" + conf.WorkDir + "\": not a directory")
	// }

	// //TODO: Check the actual path. Create directory etc.
	// Ctl = &Controller{
	// 	WorkDir: conf.WorkDir,
	// 	Waiting: make(chan *Job),
	// 	Running: make(chan *Job, conf.MaxJobs),
	// }

	// for {
	// 	select {
	// 	//case job, ok := <-c.Waiting:
	// 	}
	// }
}

// func (qmd *Qmd) Enqueue() {
// 	//if qmd.Queue == nil {err}
// 	return
// }

// func (c *Controller) Add(job *Job) error {
// 	c.Waiting <- job
// }

// func (c *Controller) RunCmd(cmd *Cmd) *Job {
// 	job := Job{
// 		Cmd: cmd,
// 	}

// 	job.Run()
// }
