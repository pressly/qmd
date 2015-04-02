package qmd

import (
	"log"
	"syscall"

	"github.com/zenazn/goji/graceful"

	"github.com/pressly/qmd/api"
	"github.com/pressly/qmd/config"
	"github.com/pressly/qmd/job"
	"github.com/pressly/qmd/script"
)

func RunOrDie(conf *config.Config) {
	// Create Job Controller.
	jobCtl, err := job.NewController(conf)
	if err != nil {
		log.Fatal(err)
	}

	// Create Script Controller.
	scriptCtl, err := script.NewController(conf)
	if err != nil {
		log.Fatal(err)
	}

	// Run controllers.
	log.Print("Starting QMD Job Controller")
	go jobCtl.Run()

	log.Print("Starting QMD Script Controller")
	go scriptCtl.Run()

	// Start the API server.
	log.Printf("Starting QMD API at %s\n", conf.Bind)
	graceful.AddSignal(syscall.SIGINT, syscall.SIGTERM)
	err = graceful.ListenAndServe(conf.Bind, api.New(conf))
	if err != nil {
		log.Fatal(err)
	}
	graceful.Wait()
}
