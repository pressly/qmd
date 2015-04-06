package job

import (
	"io"
	"os/exec"
	"time"
)

type Job struct {
	*exec.Cmd

	Stdin   io.WriteCloser
	Stdout  io.Reader
	Stderr  io.Reader
	Running bool
	Start   time.Time
	End     time.Time
}

func New(cmd *exec.Cmd) (*Job, error) {
	var err error

	job := &Job{
		Cmd: cmd,
	}

	job.Stdout, err = cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	job.Stderr, err = cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	job.Stdin, err = cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	return job, nil
}

func (job *Job) Run() error {
	if err := job.Cmd.Start(); err != nil {
		return err
	}

	job.Running = true
	return nil
}

func (job *Job) Wait() error {
	if err := job.Cmd.Wait(); err != nil {
		return err
	}

	job.Running = false
	return nil
}
