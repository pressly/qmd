package qmd

import (
	"bytes"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
)

type Job struct {
	*exec.Cmd `json:"cmd"`

	ID          string
	State       JobState      `json:"state"`
	StartTime   time.Time     `json:"start_time,omitempty"`
	EndTime     time.Time     `json:"end_time,omitempty"`
	Duration    time.Duration `json:"duration,omitempty"`
	StatusCode  int           `json:"status_code,omitempty"`
	CallbackURL string        `json:"callback_url"`
	Err         error         `json:"err,omitempty"`
	Priority    Priority      `json:"priority"`

	CmdOut bytes.Buffer `json:"-"`
	//QmdOut bytes.Buffer `json:"-"`
	QmdOutFile string `json:"qmdoutfile"`

	StoreDir          string            `json:"storedir"`
	ExtraWorkDirFiles map[string]string `json:"extraworkdirfiles"`

	// Started channel block until the job is started.
	Started chan struct{} `json:"-"`
	// Finished channel block until the job is finished/killed/invalidated.
	Finished chan struct{} `json:"-"`

	// WaitOnce guards the Wait() logic, so it can be called multiple times.
	WaitOnce sync.Once `json:"-"`
	// StartOnce guards the Start() logic, so it can be called multiple times.
	StartOnce sync.Once `json:"-"`
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

type Priority int

const (
	PriorityLow Priority = iota
	PriorityHigh
	PriorityUrgent
)

func (s Priority) String() string {
	switch s {
	case PriorityLow:
		return "low"
	case PriorityHigh:
		return "high"
	case PriorityUrgent:
		return "urgent"
	}
	panic("unreachable")
}

func (qmd *Qmd) Job(cmd *exec.Cmd) (*Job, error) {
	job := &Job{
		Cmd:      cmd,
		State:    Initialized,
		Started:  make(chan struct{}),
		Finished: make(chan struct{}),
		StoreDir: qmd.Config.StoreDir,
	}
	//TODO: Create random temp dir instead.
	job.Cmd.Dir = qmd.Config.WorkDir + "/" + job.ID

	return job, nil
}

func (job *Job) Start() error {
	job.StartOnce.Do(job.startOnce)

	// Wait for job to start.
	<-job.Started

	return job.Err
}

func (job *Job) startOnce() {
	log.Printf("Job: Starting job %v", job.ID)

	job.QmdOutFile = job.Cmd.Dir + "/QMD_OUT"
	job.Cmd.Env = append(os.Environ(),
		"QMD_TMP="+job.Cmd.Dir,
		"QMD_STORE="+job.StoreDir,
		"QMD_OUT="+job.QmdOutFile,
	)

	job.Cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
		//TODO: Chroot: job.Cmd.Dir,
	}

	job.Cmd.Stdout = &job.CmdOut
	job.Cmd.Stderr = &job.CmdOut

	// r, w, err := os.Pipe()
	// if err != nil {
	// 	job.Err = err
	// 	goto failedToStart
	// }
	// job.Cmd.ExtraFiles = []*os.File{w}
	// go job.QmdOut.ReadFrom(r)

	// Create working directory.
	err := os.MkdirAll(job.Cmd.Dir, 0777)
	if err != nil {
		job.Err = err
	}

	// Create QMD_OUT file.
	// TODO: Change this to pipe?
	qmdOut, err := os.Create(job.QmdOutFile)
	if err != nil {
		job.Err = err
	}
	qmdOut.Close()

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
	log.Printf("Job: Failed to start job %v: %v", job.ID, job.Err)
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
	log.Printf("Job: Job %v finished", job.ID)
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
				log.Printf("Job: Invalidating job %v\n", job.ID)
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
