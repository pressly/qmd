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
	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/middleware"
)

var (
	configPath = flag.String("config-file", "./config.toml", "path to qmd config file")
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

	// Setup and start worker.
	fmt.Println("Creating worker")
	worker, err := NewWorker(config)
	if err != nil {
		log.Println(err)
		return
	}
	go worker.Run()

	// Http server
	w := web.New()

	// w.Use(RequestLogger)
	w.Use(middleware.Logger)

	w.Get("/", ServiceRoot)
	w.Post("/", ServiceRoot)
	w.Get("/scripts", GetAllScripts)
	w.Put("/scripts", ReloadScripts)
	w.Post("/scripts/:name", RunScript)
	w.Get("/scripts/:name/logs", GetAllLogs)
	w.Get("/scripts/:name/logs/:id", GetLog)

	// Create and start server
	server := manners.NewServer()
	fmt.Printf("Listening on %s\n", config.ListenOnAddr)
	go server.ListenAndServe(config.ListenOnAddr, w)

	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

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

			fmt.Println("Goodbye!\n")
			return
		}
	}
}

// func ExampleMiddleware(h http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		log.Println("Request yooooooo")
// 		h.ServeHTTP(w, r)
// 	})
// }
