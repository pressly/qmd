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

type Worker struct {
	Consumer   *nsq.Consumer
	Throughput int
	QueueAddr  string
	WorkingDir string
	WhiteList  map[string]bool
}

func NewWorker(c Config) (Worker, error) {
	fmt.Printf("Creating consumer with topic: %s and channel: %s.\n", c.Topic, c.Worker.Channel)
	consumer, err := nsq.NewConsumer(c.Topic, c.Worker.Channel, nsq.NewConfig())
	if err != nil {
		log.Println(err)
		return Worker{}, err
	}

	// Generate whitelist of allowed scripts.
	path := path.Join(config.Worker.ScriptDir, config.Worker.WhiteList)
	fmt.Printf("Creating whitelist from %s\n", path)
	whiteList, err := ParseWhiteList(path)
	if err != nil {
		log.Println(err)
		return Worker{}, err
	}

	for k := range whiteList {
		fmt.Println(k)
	}

	fmt.Printf("Worker connecting to %s and running scripts in %s.\n", c.QueueAddr, c.Worker.Dir)
	return Worker{
		consumer,
		c.Worker.Throughput,
		c.QueueAddr,
		c.Worker.Dir,
		whiteList,
	}, nil
}

func ParseWhiteList(path string) (map[string]bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	whitelist := make(map[string]bool)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		whitelist[scanner.Text()] = true
	}
	return whitelist, scanner.Err()
}

func (w *Worker) Run() {
	// Set the message handler.
	w.Consumer.SetHandler(nsq.HandlerFunc(func(m *nsq.Message) error {
		var job Job
		err := json.Unmarshal(m.Body, &job)
		if err != nil {
			log.Println("Invalid JSON request", err)
			return nil
		}

		if w.WhiteList[job.Script] {
			job.Dir = w.WorkingDir
			log.Println("Dequeued request as Job", job.ID)

			_, err = job.Execute()
			if err != nil {
				job.Output = err.Error()
			}
		} else {
			msg := fmt.Sprintf("%s is not on script whitelist", job.Script)
			job.Output = msg
		}
		log.Println(job.Output)
		job.Log()
		return nil
	}))

	// Connect the queue.
	fmt.Println("Connecting to", w.QueueAddr)
	err := w.Consumer.ConnectToNSQD(w.QueueAddr)
	if err != nil {
		log.Println(err)
	}
}
