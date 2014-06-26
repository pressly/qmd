package main

import (
	"flag"
	"syscall"

	"github.com/bitly/go-nsq"
	"github.com/garyburd/redigo/redis"
	"github.com/op/go-logging"
	"github.com/pressly/gohttpware"
	"github.com/zenazn/goji/graceful"
	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/middleware"
)

var (
	configPath = flag.String("config-file", "./config.toml", "path to qmd config file")
	log        = logging.MustGetLogger("qmd")

	config   Config
	producer *nsq.Producer
	consumer *nsq.Consumer
	redisDB  *redis.Pool
)

const (
	VERSION = "0.1.0"
)

func main() {
	var err error

	flag.Parse()

	// Server config
	err = config.Load(*configPath)
	if err != nil {
		log.Fatal(err)
	}

	err = config.Setup()
	if err != nil {
		log.Fatal(err)
	}

	log.Info("=====> [ QMD v%s ] <=====", VERSION)

	// Setup facilities
	producer, err = nsq.NewProducer(config.QueueAddr, nsq.NewConfig())
	if err != nil {
		log.Fatal(err)
	}
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
	w.Use(middleware.Recoverer)
	if config.Auth.Enabled {
		w.Use(httpware.BasicAuth(config.Auth.Username, config.Auth.Password, "Restricted"))
	}
	w.Use(httpware.AllowSlash)

	w.Get("/", ServiceRoot)
	w.Post("/", ServiceRoot)
	w.Get("/scripts", GetAllScripts)
	w.Put("/scripts", ReloadScripts)
	w.Post("/scripts/:name", RunScript)
	w.Get("/scripts/:name/logs", GetAllLogs)
	w.Get("/scripts/:name/logs/:id", GetLog)

	// Spin up the server with graceful hooks
	graceful.PreHook(func() {
		log.Info("Stopping queue producer")
		producer.Stop()

		log.Info("Stopping queue workers")
		worker.Stop()

		log.Info("Closing redis connections")
		redisDB.Close()
	})

	graceful.AddSignal(syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	err = graceful.ListenAndServe(config.ListenOnAddr, w)
	if err != nil {
		log.Fatal(err)
	}
	graceful.Wait()
}
