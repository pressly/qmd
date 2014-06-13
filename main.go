package main

import (
	"flag"
	"fmt"
	"syscall"

	"github.com/BurntSushi/toml"
	"github.com/bitly/go-nsq"
	"github.com/garyburd/redigo/redis"
	"github.com/op/go-logging"
	"github.com/zenazn/goji/graceful"
	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/middleware"
)

var (
	configPath = flag.String("config-file", "./config.toml", "path to qmd config file")
	log        = logging.MustGetLogger("qmd")

	authString string
	config     Config
	producer   *nsq.Producer
	consumer   *nsq.Consumer
	redisDB    *redis.Pool
)

const (
	VERSION = "0.1.0"
)

func main() {
	flag.Parse()

	// Server config
	_, err := toml.DecodeFile(*configPath, &config)
	if err != nil {
		log.Fatal(err)
	}
	config.logging.setup()

	authString = fmt.Sprintf("%s:%s", config.Username, config.Password)

	// Setup facilities
	producer = nsq.NewProducer(config.QueueAddr, nsq.NewConfig())
	redisDB = newRedisPool(config.RedisAddr)

	// Script processing worker
	worker, err := NewWorker(config)
	if err != nil {
		log.Fatal(err)
	}
	go worker.Run()

	// Http server
	w := web.New()

	w.Use(middleware.Logger)
	w.Use(BasicAuth)
	w.Use(AllowSlash)

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
