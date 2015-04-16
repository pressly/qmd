package qmd

import (
	"crypto/sha1"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"

	"code.google.com/p/go-uuid/uuid"
)

type Job struct {
	*exec.Cmd

	ID        string
	Stdin     io.WriteCloser
	Stdout    io.Reader
	Stderr    io.Reader
	Running   bool
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration

	CallbackURL string
	Started     chan struct{}
	WaitOnce    sync.Once
	Finished    chan struct{}
	Err         error
}

func (qmd *Qmd) Job(cmd *exec.Cmd) (*Job, error) {
	var err error

	cmd.Dir = qmd.Config.WorkDir
	cmd.Env = append(cmd.Env,
		"QMD_TMP="+qmd.Config.WorkDir,
		"QMD_STORE="+qmd.Config.StoreDir,
		"QMD_OUT="+qmd.Config.WorkDir+"/out",
	)

	job := &Job{
		Cmd:      cmd,
		Started:  make(chan struct{}, 0),
		Finished: make(chan struct{}, 0),
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

	h := sha1.New()
	h.Write(uuid.NewRandom())
	job.ID = fmt.Sprintf("%x", h.Sum(nil))

	qmd.muJobs.Lock()
	defer qmd.muJobs.Unlock()
	//TODO: Check for possible ID colissions. Generate new ID & retry.
	qmd.Jobs[job.ID] = job

	return job, nil
}

func (job *Job) Start() error {
	if err := job.Cmd.Start(); err != nil {
		return err
	}

	job.StartTime = time.Now()
	job.Running = true
	close(job.Started)
	return nil
}

// Wait waits for job to finish.
// It closes the Stdout and Stderr pipes.
func (job *Job) Wait() error {
	<-job.Started

	// Prevent running cmd.Wait() multiple times.
	job.WaitOnce.Do(func() {
		if err := job.Cmd.Wait(); err != nil {
			job.Err = err
		}
		job.Duration = time.Since(job.StartTime)
		job.EndTime = job.StartTime.Add(job.Duration)
		job.Running = false
		close(job.Finished)
	})

	<-job.Finished

	return job.Err
}

func (job *Job) Run() error {
	err := job.Start()
	if err != nil {
		return err
	}

	err = job.Wait()
	if err != nil {
		return err
	}

	return nil
}
