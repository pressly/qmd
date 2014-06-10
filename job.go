package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"path"
	"time"
)

type Job struct {
	ID         int      `json:"id"`
	Script     string   `json:"script"`
	Args       []string `json:"args"`
	WorkingDir string   `json:"dir"`
	ScriptDir  string
	Output     string
	Status     string
	StartTime  time.Time
	FinishTime time.Time
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

func (j *Job) Execute() ([]byte, error) {
	args, err := j.CleanArgs()
	if err != nil {
		return nil, err
	}

	j.StartTime = time.Now()

	// Intialize command
	s := path.Join(j.ScriptDir, j.Script)
	cmd := exec.Command(s, args...)
	cmd.Dir = path.Clean(j.WorkingDir)

	log.Printf("Executing command: %s\n", s)
	out, err := cmd.Output()
	j.FinishTime = time.Now()

	if err != nil {
		j.Output = fmt.Sprintf("%s", err)
		return nil, err
	}

	j.Output = string(out)
	return out, err
}
