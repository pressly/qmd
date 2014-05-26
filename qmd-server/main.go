package main

import (
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
	listen = flag.String("listen", "127.0.0.1:4444", "listen address")
	queue  = flag.String("queue", "127.0.0.1:4445", "queue address")
	mux    *tigertonic.TrieServeMux
	writer nsq.Writer
)

func init() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: qmd-server [-listen=<listen>] [-queue=<queue>]")
		flag.PrintDefaults()
	}

	// Register the endpoints.
	mux = tigertonic.NewTrieServeMux()
	mux.Handle("POST", "/job", tigertonic.Marshaled(create))
	mux.Handle("GET", "/job/{id}", tigertonic.Marshaled(get))

	// writer := nsq.NewWriter(*queue)
}

func main() {
	flag.Parse()

	server := tigertonic.NewServer(*listen, nil)
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Println(err)
		}
	}()
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	log.Println(<-ch)
	server.Close()
}

// POST /job
func create(u *url.URL, h http.Header, rq *common.JobRequest) (int, http.Header, *common.JobResponse, error) {
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
	fmt.Println("fuck")
	return http.StatusOK, nil, &common.JobResponse{u.Query().Get("id"), "Here be scripts"}, nil
}
