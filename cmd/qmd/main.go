package main

import (
	"flag"
	"log"
	"os"
	"runtime"
	"syscall"

	"github.com/pressly/qmd"
	"github.com/pressly/qmd/config"
	"github.com/pressly/qmd/rest"
	"github.com/zenazn/goji/graceful"
)

var (
	flags    = flag.NewFlagSet("qmd", flag.ExitOnError)
	confFile = flags.String("config", "", "path to config file")
)

func main() {
	flags.Parse(os.Args[1:])

	// Override config file by the CONFIG env var, if specified.
	if os.Getenv("CONFIG") != "" {
		*confFile = os.Getenv("CONFIG")
	}

	// Read Config.
	conf, err := config.New(*confFile)
	if err != nil {
		log.Fatal(err)
	}

	// Limit number of OS threads.
	runtime.GOMAXPROCS(conf.MaxProcs)

	// Run QMD.
	app, err := qmd.New(conf)
	if err != nil {
		log.Fatal(err)
	}
	go app.WatchScripts()
	go app.StartWorkers()
	go app.ListenQueue()

	graceful.AddSignal(syscall.SIGINT, syscall.SIGTERM)
	graceful.PreHook(app.Close)

	// Start the API server.
	log.Printf("Starting QMD API at %s\n", conf.Bind)
	err = graceful.ListenAndServe(conf.Bind, rest.Routes(app))
	if err != nil {
		log.Fatal(err)
	}
	graceful.Wait()
}
