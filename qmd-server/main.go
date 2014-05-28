package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/bitly/go-nsq"
	"github.com/nulayer/qmd/common"
	"github.com/rcrowley/go-tigertonic"
)

var (
	listen   = flag.String("listen", "0.0.0.0:8080", "listen address")
	queue    = flag.String("queue", "127.0.0.1:4150", "queue address")
	topic    = flag.String("topic", "jobs", "queue topic")
	mux      *tigertonic.TrieServeMux
	producer *nsq.Producer
)

func init() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: qmd-server [-listen=<listen>] [-queue=<queue>] [-topic=<topic>]")
		flag.PrintDefaults()
	}

	// Register the endpoints.
	mux = tigertonic.NewTrieServeMux()
	mux.Handle("POST", "/job", tigertonic.Marshaled(create))
	mux.Handle("GET", "/job/{id}", tigertonic.Marshaled(get))
}

func main() {
	flag.Parse()

	producer = nsq.NewProducer(*queue, nsq.NewConfig())
	fmt.Printf("Sending to %s\n", producer.String())

	server := tigertonic.NewServer(*listen, mux)
	fmt.Printf("Listening on %s\n", *listen)
	go func() {
		var err error
		err = server.ListenAndServe()
		if err != nil {
			log.Println(err)
		}
	}()

	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	log.Println(<-termChan)
	fmt.Println("Safely shutting down server now...")
	server.Close()
}

// POST /job
func create(u *url.URL, h http.Header, rq *common.JobRequest) (int, http.Header, *common.JobResponse, error) {

	var err error

	data, err := json.Marshal(rq.Scripts)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(data)

	doneChan := make(chan *nsq.ProducerTransaction)
	err = producer.PublishAsync(*topic, data, doneChan, nil)
	if err != nil {
		fmt.Println(err)
	}
	log.Println(<-doneChan)

	return http.StatusCreated, http.Header{
		"Content-Location": {fmt.Sprintf(
			"%s://%s/1.0/job/%s",
			u.Scheme,
			u.Host,
			rq.ID,
		)},
	}, &common.JobResponse{rq.ID, rq.Scripts}, nil
}

// GET /job/{id}
func get(u *url.URL, h http.Header, rq *common.JobRequest) (int, http.Header, *common.JobResponse, error) {

	// TODO: Figure out how to request a job through the queue.

	return http.StatusOK, nil, &common.JobResponse{
		u.Query().Get("id"),
		nil,
	}, nil
}
