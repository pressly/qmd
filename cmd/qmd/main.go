package main

import (
	"flag"
	"log"
	"os"
	"runtime"
	"syscall"

	"github.com/pressly/qmd"
	"github.com/pressly/qmd/api/server"
	"github.com/pressly/qmd/config"
	"github.com/zenazn/goji/graceful"
)

var (
	flags    = flag.NewFlagSet("qmd", flag.ExitOnError)
	confFile = flags.String("config", "", "path to config file")

	bind     = flags.String("bind", "0.0.0.0:8484", "<addr>:<port> to bind HTTP server")
	maxProcs = flags.Int("max-procs", 0, "GOMAXPROCS, default is NumCpu()")
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
	app := qmd.New(conf)
	go app.WatchScripts()
	go app.ListenQueue()

	graceful.AddSignal(syscall.SIGINT, syscall.SIGTERM)
	graceful.PreHook(app.Close)

	// Start the API server.
	log.Printf("Starting QMD API at %s\n", conf.Bind)
	err = graceful.ListenAndServe(conf.Bind, server.APIHandler(app))
	if err != nil {
		log.Fatal(err)
	}
	graceful.Wait()
}
