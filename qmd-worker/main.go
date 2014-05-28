package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bitly/go-nsq"
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
		fmt.Println(string(m.Body))
		m.Finish()
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
