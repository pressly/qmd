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

	ID    string
	State JobState

	StartTime   time.Time
	EndTime     time.Time
	Duration    time.Duration
	StatusCode  int
	CallbackURL string

	StdoutFile        string
	StderrFile        string
	QmdOutFile        string
	ExtraWorkDirFiles map[string]string

	// Started channel blocks until the job is started.
	Started chan struct{}
	// Finished channel blocks until the job is finished.
	// Convinient with select{} statement with timeout.
	Finished chan struct{}

	WaitOnce sync.Once
	Err      error
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
		//Chroot: ,
	}

	if err := job.Cmd.Start(); err != nil {
		job.Err = err
		goto failedToStart
	}

	job.StartTime = time.Now()
	job.State = Running
	close(job.Started)
	return nil

failedToStart:
	job.StatusCode = -1
	job.State = Failed
	close(job.Started)
	close(job.Finished)
	log.Printf("Failed to start /jobs/%v: %v", job.ID, job.Err)
	return job.Err
}

// Wait waits for job to finish.
// It closes the Stdout and Stderr pipes.
func (job *Job) Wait() error {
	// Wait for Start(), if not already invoked.
	<-job.Started

	if job.State != Running {
		return job.Err
	}

	// Prevent running cmd.Wait() multiple times.
	job.WaitOnce.Do(func() {
		log.Printf("gonna wait")
		err := job.Cmd.Wait()
		log.Printf("wait done")
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

		log.Printf("wait done... but make sure all subprocesses are killed")
		// Make sure to kill all the whole process group,
		// so there are no subprocesses left.
		job.Kill()

		close(job.Finished)
		log.Printf("/jobs/%v finished\n", job.ID)
	})

	// Make sure the job.WaitOnce finished.
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

func (job *Job) Kill() error {
	switch job.State {
	case Running:
		job.State = Terminated
		pgid, err := syscall.Getpgid(job.Cmd.Process.Pid)
		if err == nil {
			log.Printf("gonna killall")
			// Kill the whole process group.
			syscall.Kill(-pgid, 15)
			log.Printf("killall done")
		} else {
			log.Printf("fall-back kill")
			// Fall-back on error. Kill the main process only.
			job.Cmd.Process.Kill()
			log.Printf("fall-back kill done")
		}

	case Finished:
		// Make sure to kill all the whole process group,
		// so there are no subprocesses left.
		pgid, err := syscall.Getpgid(job.Cmd.Process.Pid)
		if err != nil {
			break
		}
		log.Printf("gonna killall")
		// Kill the whole process group.
		syscall.Kill(-pgid, 15)
		log.Printf("killall done")

	case Initialized, Enqueued:
		job.State = Invalidated
		job.WaitOnce.Do(func() {
			close(job.Finished)
		})
		close(job.Started)
		job.StatusCode = -2
		job.Err = errors.New("invalidated")
		log.Printf("Invalidating /jobs/%v\n", job.ID)

	case Terminated:
		// NOP. The job might have been killed already
		// or it failed to start.
		log.Printf("Kill(): Job %v already Terminated, PID %v state: %v", job.ID, job.Cmd.Process.Pid, job.Cmd.ProcessState)
	default:
		// NOP.
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
		return "Terminated"
	case Invalidated:
		return "Terminated before start"
	case Failed:
		return "Failed to start"
	}
	panic("unreachable")
}
