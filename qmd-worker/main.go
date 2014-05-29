package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/bitly/go-nsq"
	"github.com/nulayer/qmd/common"
)

var (
	queue    = flag.String("queue", "127.0.0.1:4161", "queue address")
	topic    = flag.String("topic", "jobs", "queue topic")
	channel  = flag.String("channel", "qmd-worker", "queue channel")
	consumer *nsq.Consumer
)

func init() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: qmd-worker [-queue=<queue>] [-topic=<topic>] [-channel=<channel>]")
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()
	fmt.Printf("Worker with topic: %s and channel: %s\n", *topic, *channel)

	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	// Create and configure the consumer.
	var err error
	consumer, err = nsq.NewConsumer(*topic, *channel, nsq.NewConfig())
	if err != nil {
		fmt.Println(err)
	}

	// Set the message handler.
	consumer.SetHandler(nsq.HandlerFunc(func(m *nsq.Message) error {
		job, err := parseMessage(m)
		if err != nil {
			fmt.Println(err)
			return err
		}

		err = runScript(job)
		if err != nil {
			fmt.Println(err)
			return err
		}
		return nil
	}))

	// Connect the consumer.
	fmt.Printf("Connecting to %s\n", *queue)
	err = consumer.ConnectToNSQLookupd(*queue)
	if err != nil {
		fmt.Println(err)
	}

	for {
		select {
		case <-consumer.StopChan:
			return
		case <-termChan:
			consumer.Stop()
		}
	}
}

func parseMessage(m *nsq.Message) (common.Job, error) {
	var job common.Job
	err := json.Unmarshal(m.Body, &job)
	if err != nil {
		return job, err
	}
	return job, nil
}

func runScript(job common.Job) error {
	for _, script := range job.Scripts {
		// TODO: Make the strings safe. Somehow...
		out, err := exec.Command(script.Name, script.Params...).Output()
		if err != nil {
			return err
		}
		fmt.Println(string(out))
	}
	return nil
}
