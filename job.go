package qmd

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"

	"code.google.com/p/go-uuid/uuid"
)

type Job struct {
	*exec.Cmd

	ID        string
	State     JobState
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration

	StdoutFile string
	StderrFile string
	QmdOutFile string

	CallbackURL string
	Started     chan struct{}
	WaitOnce    sync.Once
	Finished    chan struct{}
	Err         error
}

type JobState int

const (
	Initialized JobState = iota
	Enqueued
	Running
	Finished
)

func (qmd *Qmd) Job(cmd *exec.Cmd) (*Job, error) {
	job := &Job{
		State:    Initialized,
		Cmd:      cmd,
		Started:  make(chan struct{}, 0),
		Finished: make(chan struct{}, 0),
	}

	// Assign an unique ID to the job.
	h := sha1.New()
	h.Write(uuid.NewRandom())
	job.ID = fmt.Sprintf("%x", h.Sum(nil))

	// Create working directory.
	job.Cmd.Dir = qmd.Config.WorkDir + "/" + job.ID
	err := os.MkdirAll(job.Cmd.Dir, 0777)
	if err != nil {
		return nil, err
	}

	job.StdoutFile = job.Cmd.Dir + "/stdout"
	job.StderrFile = job.Cmd.Dir + "/stderr"
	job.QmdOutFile = job.Cmd.Dir + "/QMD_OUT"

	cmd.Env = append(cmd.Env,
		"QMD_TMP="+job.Cmd.Dir+"/tmp",
		"QMD_STORE="+qmd.Config.StoreDir,
		"QMD_OUT="+job.QmdOutFile,
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

	if job.State >= Running {
		return errors.New(fmt.Sprintf("/jobs/%v already running", job.ID))
	}

	log.Printf("Starting /jobs/%v\n", job.ID)

	job.Cmd.Stdout, err = os.Create(job.StdoutFile)
	if err != nil {
		return err
	}

	job.Cmd.Stderr, err = os.Create(job.StderrFile)
	if err != nil {
		return err
	}

	qmdOut, err := os.Create(job.QmdOutFile)
	if err != nil {
		return err
	}
	qmdOut.Close()

	// // FD 3
	// job.ExtraFiles = append(job.ExtraFiles, qmdOut)

	if err := job.Cmd.Start(); err != nil {
		return err
	}

	job.StartTime = time.Now()
	job.State = Running
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
		job.State = Finished
		close(job.Finished)
		log.Printf("/jobs/%v finished\n", job.ID)
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

func (s JobState) String() string {
	switch s {
	case Enqueued:
		return "Enqueued"
	case Running:
		return "Running"
	case Finished:
		return "Finished"
	default:
		return "Initialized"
	}
}
