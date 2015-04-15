package qmd

import (
	"io"
	"os/exec"
	"time"
)

type Job struct {
	*exec.Cmd

	Stdin    io.WriteCloser
	Stdout   io.Reader
	Stderr   io.Reader
	Running  bool
	start    time.Time
	Duration time.Duration
}

// func New(cmd *exec.Cmd) (*Job, error) {
// 	var err error

// 	job := &Job{
// 		Cmd: cmd,
// 	}

// 	job.Stdout, err = cmd.StdoutPipe()
// 	if err != nil {
// 		return nil, err
// 	}

// 	job.Stderr, err = cmd.StderrPipe()
// 	if err != nil {
// 		return nil, err
// 	}

// 	job.Stdin, err = cmd.StdinPipe()
// 	if err != nil {
// 		return nil, err
// 	}

// 	return job, nil
// }

// func (job *Job) Start() error {
// 	if err := job.Cmd.Start(); err != nil {
// 		return err
// 	}

// 	job.start = time.Now()
// 	job.Running = true
// 	return nil
// }

// func (job *Job) Wait() error {
// 	if err := job.Cmd.Wait(); err != nil {
// 		return err
// 	}

// 	job.Duration = time.Since(job.start)
// 	job.Running = false
// 	return nil
// }

// func (job *Job) Run() error {
// 	err := job.Start()
// 	if err != nil {
// 		return err
// 	}

// 	err = job.Wait()
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }
