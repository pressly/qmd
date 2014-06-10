package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
	"time"
)

type Job struct {
	ID          int      `json:"id"`
	Script      string   `json:"script"`
	Args        []string `json:"args"`
	CallbackURL string   `json:"callback_url"`
	WorkingDir  string
	ScriptDir   string
	StoreDir    string
	Output      string
	ExecLog     string
	Status      string
	StartTime   time.Time
	FinishTime  time.Time
}

func (j *Job) CleanArgs() ([]string, error) {
	// TODO find a way to clean the arguments
	return j.Args, nil
}

func (j *Job) Log() error {
	conn := redisDB.Get()
	defer conn.Close()

	data, err := json.Marshal(j)
	if err != nil {
		log.Println(err)
		return err
	}

	log.Printf("Adding job %d to log for %s to Redis\n", j.ID, j.Script)
	_, err = conn.Do("ZADD", j.Script, j.ID, data)
	if err != nil {
		log.Println(err)
		return err
	}

	log.Printf("Trimming log for %s to the %d most recent\n", j.Script, LOGLIMIT)
	_, err = conn.Do("ZREMRANGEBYRANK", j.Script, 0, -LOGLIMIT-1)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (j *Job) Callback() error {
	data, err := json.Marshal(j)
	if err != nil {
		log.Println(err)
		return err
	}

	log.Printf("Sending status back to %s\n", j.CallbackURL)
	buf := bytes.NewBuffer(data)
	_, err = http.Post(j.CallbackURL, "application/json", buf)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (j *Job) Execute() ([]byte, error) {
	j.StartTime = time.Now()

	// Intialize script path and arguments
	s := path.Join(j.ScriptDir, j.Script)
	args, err := j.CleanArgs()
	if err != nil {
		j.ExecLog = fmt.Sprintf("%s", err)
		return nil, err
	}

	// Set environment variables
	os.Setenv("QMD_STORE", path.Clean(j.StoreDir))
	tmpPath := path.Join(j.WorkingDir, "tmp", strconv.Itoa(j.ID))
	os.MkdirAll(tmpPath, 0777)
	os.Setenv("QMD_TMP", tmpPath)
	os.Setenv("QMD_OUT", path.Join(j.WorkingDir, "qmd.out"))

	// Setup and execute cmd
	cmd := exec.Command(s, args...)
	cmd.Dir = path.Clean(j.WorkingDir)

	log.Printf("Executing command: %s\n", s)
	out, err := cmd.Output()
	j.FinishTime = time.Now()

	if err != nil {
		j.ExecLog = fmt.Sprintf("%s", err)
		return nil, err
	}

	j.ExecLog = string(out)
	return out, err
}
