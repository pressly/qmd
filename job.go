package qmd

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"

	"code.google.com/p/go-uuid/uuid"
)

type Job struct {
	*exec.Cmd

	ID          string
	State       JobState
	StartTime   time.Time
	EndTime     time.Time
	Duration    time.Duration
	StatusCode  int
	CallbackURL string
	Err         error

	StdoutFile        string
	StderrFile        string
	QmdOutFile        string
	ExtraWorkDirFiles map[string]string

	// Started channel block until the job is started.
	Started chan struct{}
	// Finished channel block until the job is finished/killed/invalidated.
	Finished chan struct{}

	// WaitOnce guards the Wait() logic, so it can be called multiple times.
	WaitOnce sync.Once
	// StartOnce guards the Start() logic, so it can be called multiple times.
	StartOnce sync.Once
}

type JobState int

const (
	Initialized JobState = iota
	Enqueued
	Running
	Finished
	Terminated
	Invalidated
	Failed
)

func (qmd *Qmd) Job(cmd *exec.Cmd) (*Job, error) {
	job := &Job{
		State:    Initialized,
		Cmd:      cmd,
		Started:  make(chan struct{}),
		Finished: make(chan struct{}),
	}

	// Assign an unique ID to the job.
	h := sha1.New()
	h.Write(uuid.NewRandom())
	job.ID = fmt.Sprintf("%x", h.Sum(nil))

	// Create working directory with /tmp subdirectory.
	job.Cmd.Dir = qmd.Config.WorkDir + "/" + job.ID
	err := os.MkdirAll(job.Cmd.Dir+"/tmp", 0777)
	if err != nil {
		return nil, err
	}

	job.StdoutFile = job.Cmd.Dir + "/stdout"
	job.StderrFile = job.Cmd.Dir + "/stderr"
	job.QmdOutFile = job.Cmd.Dir + "/QMD_OUT"

	cmd.Env = append(os.Environ(),
		"QMD_TMP="+job.Cmd.Dir+"/tmp",
		"QMD_STORE="+qmd.Config.StoreDir,
		"QMD_OUT="+job.QmdOutFile,
	)

	// Save this job to the QMD.
	//TODO: Check for possible ID colissions. Generate new ID & retry.
	qmd.MuJobs.Lock()
	defer qmd.MuJobs.Unlock()
	qmd.Jobs[job.ID] = job

	return job, nil
}

func (job *Job) Start() error {
	job.StartOnce.Do(job.startOnce)

	// Wait for job to start.
	<-job.Started

	return job.Err
}

func (job *Job) startOnce() {
	log.Printf("Starting /jobs/%v\n", job.ID)

	var err error
	job.Cmd.Stdout, err = os.Create(job.StdoutFile)
	if err != nil {
		job.Err = err
	}

	job.Cmd.Stderr, err = os.Create(job.StderrFile)
	if err != nil {
		job.Err = err
	}

	qmdOut, err := os.Create(job.QmdOutFile)
	if err != nil {
		job.Err = err
	}
	qmdOut.Close()

	// // FD 3
	// job.ExtraFiles = append(job.ExtraFiles, qmdOut)

	for file, data := range job.ExtraWorkDirFiles {
		// Must be a simple filename without slashes.
		if strings.Index(file, "/") != -1 {
			job.Err = errors.New("extra file contains slashes")
			goto failedToStart
		}
		err = ioutil.WriteFile(job.Cmd.Dir+"/tmp/"+file, []byte(data), 0644)
		if err != nil {
			job.Err = err
			goto failedToStart
		}
	}

	job.Cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
		//TODO: Chroot: job.Cmd.Dir,
	}

	if err := job.Cmd.Start(); err != nil {
		job.Err = err
		goto failedToStart
	}

	job.StartTime = time.Now()
	job.State = Running
	close(job.Started)
	job.Err = nil
	return

failedToStart:
	job.StatusCode = -1
	job.State = Failed
	job.WaitOnce.Do(func() {
		close(job.Finished)
	})
	close(job.Started)
	log.Printf("Failed to start /jobs/%v: %v", job.ID, job.Err)
}

// Wait waits for job to finish.
// It closes the Stdout and Stderr pipes.
func (job *Job) Wait() error {
	// Wait for Start(), if not already invoked.
	<-job.Started

	// Prevent running cmd.Wait() multiple times.
	job.WaitOnce.Do(job.waitOnce)

	// Wait for job to finish.
	<-job.Finished

	return job.Err
}

func (job *Job) waitOnce() {
	err := job.Cmd.Wait()
	job.Duration = time.Since(job.StartTime)
	job.EndTime = job.StartTime.Add(job.Duration)
	if job.State != Terminated {
		job.State = Finished
	}

	if err != nil {
		job.Err = err
		if e, ok := err.(*exec.ExitError); ok {
			if s, ok := e.Sys().(syscall.WaitStatus); ok {
				job.StatusCode = s.ExitStatus()
			}
		}
	}

	// Make sure to kill the whole process group,
	// so there are no subprocesses left.
	job.Kill()

	close(job.Finished)
	log.Printf("/jobs/%v finished\n", job.ID)
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

func (job *Job) Kill() error {
	switch job.State {
	case Running:
		job.State = Terminated
		pgid, err := syscall.Getpgid(job.Cmd.Process.Pid)
		if err != nil {
			// Fall-back on error. Kill the main process only.
			job.Cmd.Process.Kill()
			break
		}
		// Kill the whole process group.
		syscall.Kill(-pgid, 15)

	case Finished:
		pgid, err := syscall.Getpgid(job.Cmd.Process.Pid)
		if err != nil {
			break
		}
		// Make sure to kill the whole process group,
		// so there are no subprocesses left.
		syscall.Kill(-pgid, 15)

	case Initialized, Enqueued:
		// This one is tricky, as the job's Start() might have
		// been called and is already in progress, but the job's
		// state is not Running yet.
		usCallingStartOnce := false
		job.StartOnce.Do(func() {
			job.WaitOnce.Do(func() {
				job.State = Invalidated
				job.StatusCode = -2
				job.Err = errors.New("invalidated")
				log.Printf("Invalidating /jobs/%v\n", job.ID)
				close(job.Finished)
			})
			close(job.Started)
			usCallingStartOnce = true
		})
		if !usCallingStartOnce {
			// It was job.Start() that called StartOnce.Do(), not us,
			// thus we need to wait for Started and try to Kill again:
			<-job.Started
			job.Kill()
		}
	}

	return job.Err
}

func (s JobState) String() string {
	switch s {
	case Initialized:
		return "Initialized"
	case Enqueued:
		return "Enqueued"
	case Running:
		return "Running"
	case Finished:
		return "Finished"
	case Terminated:
		return "Terminated (killed by us)"
	case Invalidated:
		return "Invalidated before start"
	case Failed:
		return "Failed to start"
	}
	panic("unreachable")
}
