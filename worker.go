package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bitly/go-nsq"
)

type Worker struct {
	Consumer   *nsq.Consumer
	Throughput int
	QueueAddr  string
}

func NewWorker(c Config) (Worker, error) {
	consumer, err := nsq.NewConsumer(c.Topic, c.Worker.Channel, nsq.NewConfig())
	if err != nil {
		return Worker{}, err
	}
	return Worker{consumer, c.Worker.Throughput, c.QueueAddr}, nil
}

func (w *Worker) Run() {
	// Set the message handler.
	w.Consumer.SetHandler(nsq.HandlerFunc(func(m *nsq.Message) error {
		var job Job
		err := json.Unmarshal(m.Body, &job)
		if err != nil {
			log.Println(err)
		}
		log.Printf("Dequeued request as Job #%s", job.ID)

		_, err = job.Execute()
		if err != nil {
			return err
		}

		go job.Log()
		return nil
	}))

	// Connect the queue.
	fmt.Printf("Connecting to %s\n", w.QueueAddr)
	err := w.Consumer.ConnectToNSQD(w.QueueAddr)
	if err != nil {
		log.Println(err)
	}

	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	for {
		select {
		case <-w.Consumer.StopChan:
			return
		case <-termChan:
			w.Consumer.Stop()
		}
	}
}
