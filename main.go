package main

import (
	"flag"
	"fmt"
	"log"
	"syscall"

	"github.com/BurntSushi/toml"
	"github.com/bitly/go-nsq"
	"github.com/garyburd/redigo/redis"
	"github.com/zenazn/goji/graceful"
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

const (
	VERSION = "0.1.0"
)

func main() {
	flag.Parse()

	_, err := toml.DecodeFile(*configPath, &config)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Setting up producer")
	producer = nsq.NewProducer(config.QueueAddr, nsq.NewConfig())

	fmt.Println("Creating Redis connection pool")
	redisDB = newRedisPool(config.RedisAddr)

	// Setup and start worker.
	fmt.Println("Creating worker")
	worker, err := NewWorker(config)
	if err != nil {
		log.Fatal(err)
	}
	go worker.Run()

	// Http server
	w := web.New()

	// Register middleware
	w.Use(middleware.Logger)
	w.Use(BasicAuth)
	w.Use(AllowSlash)

	// Register endpoints
	w.Get("/", ServiceRoot)
	w.Post("/", ServiceRoot)
	w.Get("/scripts", GetAllScripts)
	w.Put("/scripts", ReloadScripts)
	w.Post("/scripts/:name", RunScript)
	w.Get("/scripts/:name/logs", GetAllLogs)
	w.Get("/scripts/:name/logs/:id", GetLog)

	// Spin up the server with graceful hooks
	graceful.PreHook(func() {
		fmt.Println("Shutting down producer")
		producer.Stop()

		fmt.Println("Shutting down worker consumers")
		worker.Stop()

		fmt.Println("Closing Redis connections")
		redisDB.Close()
	})

	graceful.AddSignal(syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	err = graceful.ListenAndServe(config.ListenOnAddr, w)
	if err != nil {
		log.Fatal(err)
	}
	graceful.Wait()
}
