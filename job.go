package main

import (
	"fmt"
	"os/exec"
	"time"
)

type Job struct {
	ID         string   `json:"id"`
	Script     string   `json:"script"`
	Args       []string `json:"args"`
	StartTime  time.Time
	FinishTime time.Time
}

func (j *Job) CleanArgs() ([]string, error) {
	// TODO find a way to clean the arguments
	return j.Args, nil
}

func (j *Job) Execute() ([]byte, error) {
	name := fmt.Sprintf("./%s", j.Script)
	args, err := j.CleanArgs()
	if err != nil {
		return nil, err
	}

	j.StartTime = time.Now()
	cmd := exec.Command(name, args...)
	out, err := cmd.Output()
	j.FinishTime = time.Now()
	if err != nil {
		return nil, err
	}
	return out, err
}
