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

	"github.com/nulayer/qmd/common"
	"github.com/rcrowley/go-tigertonic"
)

var (
	listen = flag.String("listen", "0.0.0.0:8080", "listen address")
	queue  = flag.String("queue", "0.0.0.0:8181", "queue address")
	mux    *tigertonic.TrieServeMux
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
}

func main() {
	flag.Parse()

	server := tigertonic.NewServer(*listen, mux)
	fmt.Printf("Listening on %s\n", *listen)
	go func() {
		var err error
		err = server.ListenAndServe()
		if nil != err {
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
	return http.StatusOK, nil, &common.JobResponse{u.Query().Get("id"), "Here be scripts"}, nil
}
