package qmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

const (
	StatusQUEUED  string = "QUEUED"
	StatusTIMEOUT string = "TIMED OUT"
	StatusOK      string = "OK"
	StatusERR     string = "ERR"
)

type Job struct {
	*Request
	Output   string `json:"output,omitempty"`
	ExecLog  string `json:"exec_log,omitempty"`
	killChan chan int
}

func NewJob(data []byte) (Job, error) {
	var job Job
	if err := json.Unmarshal(data, &job); err != nil {
		return job, err
	}
	job.killChan = make(chan int, 1)
	return job, nil
}

func (j *Job) Execute(scriptDir, workingDir, storeDir string, keep bool) error {
	defer close(j.killChan)
	var err error

	// Intialize script path and arguments
	s := path.Join(scriptDir, j.Script)
	args, err := j.cleanArgs()
	if err != nil {
		j.ExecLog = err.Error()
		return err
	}

	// Set environment variables
	env := os.Environ()
	env = append(env, fmt.Sprintf("%s=%s", "QMD_STORE", storeDir))

	tmpPath := path.Join(workingDir, j.ID)
	env = append(env, fmt.Sprintf("%s=%s", "QMD_TMP", tmpPath))
	os.MkdirAll(tmpPath, 0775)
	if !keep {
		defer j.removeTmpDir(tmpPath)
	}

	err = j.SaveFiles(tmpPath)
	if err != nil {
		j.ExecLog = err.Error()
		return err
	}

	outPath := path.Join(tmpPath, "qmd.out")
	env = append(env, fmt.Sprintf("%s=%s", "QMD_OUT", outPath))

	// Setup and execute cmd
	cmd := exec.Command(s, args...)
	cmd.Dir = workingDir
	cmd.Env = env
	cmdOut := bytes.NewBuffer(nil)
	cmd.Stdout = cmdOut
	cmd.Stderr = cmdOut

	lg.Info("Executing command: %s %s", s, args)
	doneChan := make(chan error)
	defer close(doneChan)
	go func() {
		e := cmd.Run()
		doneChan <- e
	}()

	select {
	case <-j.killChan:
		cmd.Process.Kill()
		<-doneChan
		j.Status = StatusTIMEOUT
		return fmt.Errorf("Killed job %s early", j.ID)
	case err = <-doneChan:
		j.FinishTime = time.Now()
		j.Duration = fmt.Sprintf("%f", j.FinishTime.Sub(j.StartTime).Seconds())
		j.ExecLog = fmt.Sprintf("%s", string(cmdOut.Bytes()))

		data, er := ioutil.ReadFile(outPath)
		if !os.IsNotExist(er) {
			j.Output = string(data)
		} else {
			lg.Error(er.Error())
		}
		if err != nil {
			j.ExecLog = fmt.Sprintf("%s\n%s", j.ExecLog, err.Error())
			j.Status = StatusERR
			return err
		}
	}
	j.Status = StatusOK
	return nil
}

func (j *Job) SaveFiles(dir string) error {
	var err error
	var file string
	for name, data := range j.Files {

		// Clean bad input
		name = strings.Replace(name, "..", "", -1)
		name = strings.Replace(name, "/", "", -1)

		file = path.Join(dir, name)
		lg.Debug("Writing %s to disk", file)
		err = ioutil.WriteFile(file, []byte(data), 0644)
		if err != nil {
			return err
		}
	}
	return nil
}

func (j *Job) kill() {
	j.killChan <- 1
}

func (j *Job) cleanArgs() ([]string, error) {
	// TODO find a better way to clean the arguments
	return j.Args, nil
}

func (j *Job) removeTmpDir(tmpPath string) {
	lg.Debug("Deleting all files and dirs in %s", tmpPath)
	err := os.RemoveAll(tmpPath)
	if err != nil {
		lg.Error("Failed to delete all files and dirs in %s - %s", tmpPath, err)
	}
}
