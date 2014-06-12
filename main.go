package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/BurntSushi/toml"
	"github.com/bitly/go-nsq"
	"github.com/braintree/manners"
	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/mux"
)

var (
	configPath = flag.String("config-path", "./qmd.toml", "path to config file")
	config     Config

	producer *nsq.Producer
	consumer *nsq.Consumer
	redisDB  *redis.Pool
)

func main() {
	flag.Parse()
	fmt.Printf("Using config file from: %s\n", *configPath)

	var err error
	_, err = toml.DecodeFile(*configPath, &config)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println("Setting up producer")
	producer = nsq.NewProducer(config.QueueAddr, nsq.NewConfig())

	fmt.Println("Creating Redis connection pool")
	redisDB = newPool(config.RedisAddr)

	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	// Setup and start worker.
	fmt.Println("Creating worker")
	worker, err := NewWorker(config)
	if err != nil {
		log.Println(err)
		return
	}
	go worker.Run()

	// Register endpoints
	rtr := mux.NewRouter()
	rtr.HandleFunc("/scripts", GetAllScripts).Methods("GET")
	rtr.HandleFunc("/scripts", ReloadScripts).Methods("PUT")
	rtr.HandleFunc("/scripts/{name}", RunScript).Methods("POST")
	rtr.HandleFunc("/scripts/{name}/logs", GetAllLogs).Methods("GET")
	rtr.HandleFunc("/scripts/{name}/logs/{id}", GetLog).Methods("GET")

	// Create and start server
	server := manners.NewServer()
	fmt.Printf("Listening on %s\n", config.ListenOnAddr)
	go server.ListenAndServe(config.ListenOnAddr, rtr)

	// Gracefully shutdown all the connections
	for {
		select {
		case <-termChan:
			fmt.Println("Shutting down producer")
			producer.Stop()

			fmt.Println("Shutting down worker consumers")
			worker.Stop()

			fmt.Println("Closing Redis connections")
			redisDB.Close()

			fmt.Println("Shutting down server")
			server.Shutdown <- true

			return
		}
	}
}
