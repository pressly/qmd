package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/bitly/go-nsq"
)

const (
	STATUS_OK  string = "OK"
	STATUS_ERR string = "ERR"
)

type Worker struct {
	Consumer   *nsq.Consumer
	Throughput int
	QueueAddr  string
	ScriptDir  string
	StoreDir   string
	WorkingDir string
	WhiteList  map[string]bool
}

func NewWorker(c Config) (Worker, error) {
	fmt.Printf("Creating consumer with topic: %s and channel: %s.\n", c.Topic, c.Worker.Channel)

	var worker Worker

	consumer, err := nsq.NewConsumer(c.Topic, c.Worker.Channel, nsq.NewConfig())
	if err != nil {
		log.Println(err)
		return worker, err
	}
	worker.Consumer = consumer
	worker.Throughput = c.Worker.Throughput
	worker.QueueAddr = c.QueueAddr
	worker.ScriptDir = c.Worker.ScriptDir
	worker.StoreDir = c.Worker.StoreDir
	worker.WorkingDir = c.Worker.WorkingDir

	// Generate whitelist of allowed scripts.
	path := path.Join(config.Worker.ScriptDir, config.Worker.WhiteList)
	err = worker.LoadWhiteList(path)
	if err != nil {
		log.Println(err)
		return worker, err
	}

	fmt.Printf("Worker connecting to %s and running scripts in %s.\n", c.QueueAddr, c.Worker.WorkingDir)
	return worker, nil
}

func (w *Worker) LoadWhiteList(path string) error {
	fmt.Printf("Creating whitelist from %s\n", path)

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	whiteList := make(map[string]bool)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		whiteList[scanner.Text()] = true
	}
	err = scanner.Err()
	if err != nil {
		log.Println(err)
		return err
	}

	fmt.Println("Whitelist:")
	for script := range whiteList {
		fmt.Println(script)
	}

	w.WhiteList = whiteList
	return nil
}

func (w *Worker) Run() {
	// Set the message handler.
	w.Consumer.SetHandler(nsq.HandlerFunc(func(m *nsq.Message) error {

		// Initialize Job from request
		var job Job
		job.Status = STATUS_ERR
		job.ScriptDir = w.ScriptDir
		job.WorkingDir = w.WorkingDir
		job.StoreDir = w.StoreDir

		err := json.Unmarshal(m.Body, &job)
		if err != nil {
			log.Println("Invalid JSON request", err)
			return nil
		}

		// Try and run script
		if w.WhiteList[job.Script] {
			log.Println("Dequeued request as Job", job.ID)

			_, err = job.Execute()
			if err != nil {
				job.ExecLog = err.Error()
			}
			job.Status = STATUS_OK
		} else {
			msg := fmt.Sprintf("%s is not on script whitelist", job.Script)
			job.ExecLog = msg
		}

		log.Println(job.ExecLog)
		job.Log()
		job.Callback()
		return nil
	}))

	// Connect the queue.
	fmt.Println("Connecting to", w.QueueAddr)
	err := w.Consumer.ConnectToNSQD(w.QueueAddr)
	if err != nil {
		log.Println(err)
	}
}
