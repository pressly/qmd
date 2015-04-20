package qmd

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"code.google.com/p/go-uuid/uuid"
)

type Job struct {
	*exec.Cmd

	ID        string
	Running   bool
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration

	StdoutFile   string
	StderrFile   string
	ExtraOutFile string

	CallbackURL string
	Started     chan struct{}
	WaitOnce    sync.Once
	Finished    chan struct{}
	Err         error
}

func (qmd *Qmd) Job(cmd *exec.Cmd) (*Job, error) {
	job := &Job{
		Cmd:      cmd,
		Started:  make(chan struct{}, 0),
		Finished: make(chan struct{}, 0),
	}

	// Assign an unique ID to the job.
	h := sha1.New()
	h.Write(uuid.NewRandom())
	job.ID = fmt.Sprintf("%x", h.Sum(nil))

	// Create working directory.
	job.Cmd.Dir = qmd.Config.WorkDir + "/" + job.ID + "/tmp"
	err := os.MkdirAll(job.Cmd.Dir, 0777)
	if err != nil {
		return nil, err
	}

	job.StdoutFile = qmd.Config.WorkDir + "/" + job.ID + "/stdout"
	job.StderrFile = qmd.Config.WorkDir + "/" + job.ID + "/stderr"
	job.ExtraOutFile = qmd.Config.WorkDir + "/" + job.ID + "/QMD_OUT"

	cmd.Env = append(cmd.Env,
		"QMD_TMP="+job.Cmd.Dir,
		"QMD_STORE="+qmd.Config.StoreDir,
		"QMD_OUT="+job.ExtraOutFile,
	)

	// Save this job to the QMD.
	//TODO: Check for possible ID colissions. Generate new ID & retry.
	qmd.muJobs.Lock()
	defer qmd.muJobs.Unlock()
	qmd.Jobs[job.ID] = job

	return job, nil
}

func (job *Job) Start() error {
	var err error

	if job.Running {
		return errors.New(fmt.Sprintf("job #%v already running", job.ID))
	}

	job.Cmd.Stdout, err = os.Create(job.StdoutFile)
	if err != nil {
		return err
	}

	job.Cmd.Stderr, err = os.Create(job.StderrFile)
	if err != nil {
		return err
	}

	if err := job.Cmd.Start(); err != nil {
		return err
	}

	job.StartTime = time.Now()
	job.Running = true
	close(job.Started)
	return nil
}

func (job *Job) WaitForStart() {
	<-job.Started
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
