package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"time"
)

type Job struct {
	ID         int      `json:"id"`
	Script     string   `json:"script"`
	Args       []string `json:"args"`
	Dir        string
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

	log.Printf("Adding log for %s to Redis\n", j.ID)
	_, err = conn.Do("ZADD", j.Script, j.ID, data)
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
	s := fmt.Sprintf("./%s", j.Script)
	cmd := exec.Command(s, args...)
	cmd.Dir = j.Dir
	log.Printf("Executing command: %s in %s\n", cmd.Args, cmd.Dir)
	out, err := cmd.Output()
	fmt.Println(string(out))
	j.FinishTime = time.Now()
	if err != nil {
		j.Output = fmt.Sprintf("%s", err)
		return nil, err
	}
	j.Output = string(out)
	return out, err
}
