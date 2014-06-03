package main

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/bitly/go-nsq"
	"github.com/nulayer/qmd/common"
)

type Worker struct {
	Consumer   *nsq.Consumer
	Throughput int
}

func (w *Worker) Run() {
	// Set the message handler.
	w.Consumer.SetHandler(nsq.HandlerFunc(func(m *nsq.Message) error {
		job, err := w.parse(m)
		if err != nil {
			fmt.Println(err)
			return err
		}

		err = w.execute(job)
		if err != nil {
			fmt.Println(err)
			return err
		}
		return nil
	}))
}

func (w *Worker) parse(m *nsq.Message) (common.Job, error) {
	var job common.Job
	err := json.Unmarshal(m.Body, &job)
	if err != nil {
		return job, err
	}
	return job, nil
}

func (w *Worker) execute(job common.Job) error {
	var name string
	for _, script := range job.Scripts {
		// TODO: Make the strings safe. Somehow...
		name = fmt.Sprintf("./%s", script.Name)
		out, err := exec.Command(name, script.Params...).Output()
		if err != nil {
			return err
		}
		fmt.Println(string(out))
	}
	return nil
}
