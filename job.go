package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
	"time"
)

type Job struct {
	ID          int               `json:"id"`
	Script      string            `json:"script"`
	Args        []string          `json:"args"`
	Files       map[string]string `json:"files"`
	CallbackURL string            `json:"callback_url"`
	WorkingDir  string            `json:"working_dir"`
	ScriptDir   string            `json:"script_dir"`
	StoreDir    string            `json:"store_dir"`
	Output      string            `json:"output"`
	ExecLog     string            `json:"exec_log"`
	Status      string            `json:"status"`
	StartTime   time.Time         `json:"start_time"`
	FinishTime  time.Time         `json:"end_time"`
}

func (j *Job) CleanArgs() ([]string, error) {
	// TODO find a way to clean the arguments
	return j.Args, nil
}

func (j *Job) SaveFiles(dir string) error {
	var err error
	var file string
	for name, data := range j.Files {
		file = path.Join(dir, name)
		log.Printf("Writing %s to disk\n", file)
		err = ioutil.WriteFile(file, []byte(data), 0644)
		if err != nil {
			return err
		}
	}
	return nil
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
	fmt.Println(j.Files)
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

	err = j.SaveFiles(tmpPath)
	if err != nil {
		j.ExecLog = fmt.Sprintf("%s", err)
		return nil, err
	}

	outPath := path.Join(tmpPath, "qmd.out")
	os.Setenv("QMD_OUT", outPath)

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

	data, err := ioutil.ReadFile(outPath)
	if err != nil {
		log.Println(err)
	}
	j.Output = string(data)

	j.ExecLog = string(out)
	return out, err
}
