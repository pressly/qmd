package main

import (
	"flag"
	"fmt"

	"net/http"

	"github.com/BurntSushi/toml"
	"github.com/bitly/go-nsq"
	"github.com/gorilla/mux"
)

var (
	configPath = flag.String("config-path", "./qmd.toml", "path to config file")
	config     Config

	producer *nsq.Producer
	consumer *nsq.Consumer
)

func main() {
	flag.Parse()
	fmt.Printf("Using config file from: %s\n", *configPath)

	if _, err := toml.DecodeFile(*configPath, &config); err != nil {
		fmt.Println(err)
		return
	}
	producer = nsq.NewProducer(config.QueueAddr, nsq.NewConfig())

	// Setup and start worker.
	worker, err := NewWorker(config)
	if err != nil {
		fmt.Println(err)
	}
	worker.Run()

	// Register endpoints
	rtr := mux.NewRouter()
	pre := rtr.PathPrefix("/scripts").Subrouter()
	pre.HandleFunc("/", GetAllScripts).Methods("GET")
	pre.HandleFunc("/{name}/", RunScript).Methods("POST")
	pre.HandleFunc("/{name}/logs", GetAllLogs).Methods("GET")
	pre.HandleFunc("/{name}/logs/{id}/", GetLog).Methods("GET")

	http.ListenAndServe(config.ListenOnAddr, nil)
}
