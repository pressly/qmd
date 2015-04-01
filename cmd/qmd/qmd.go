package main

import (
	"flag"
	"log"
	"os"
	"runtime"
	"syscall"

	//"github.com/pressly/qmd"
	"github.com/pressly/qmd/config"
	"github.com/pressly/qmd/server"
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

	// Create a QMD app.
	// app, err := qmd.New(conf)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	log.Printf("Starting QMD at %s\n", conf.Bind)

	graceful.AddSignal(syscall.SIGINT, syscall.SIGTERM)
	err = graceful.ListenAndServe(conf.Bind, server.New(conf))
	if err != nil {
		log.Fatal(err)
	}
	graceful.Wait()
}
