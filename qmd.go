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

var App *QMD

type QMD struct {
	Config    *config.Config
	JobCtl    *job.Controller
	ScriptCtl *script.Controller
}

func RunOrDie(conf *config.Config) {
	var err error

	App := &QMD{
		Config: conf,
	}

	// Create Job Controller.
	App.JobCtl, err = job.NewController(conf)
	if err != nil {
		log.Fatal(err)
	}

	// Create Script Controller.
	App.ScriptCtl, err = script.NewController(conf)
	if err != nil {
		log.Fatal(err)
	}

	// Run controllers.
	log.Print("Starting QMD Job Controller")
	go App.JobCtl.Run()

	log.Print("Starting QMD Script Controller")
	go App.ScriptCtl.Run()

	// Start the API server.
	log.Printf("Starting QMD API at %s\n", conf.Bind)
	graceful.AddSignal(syscall.SIGINT, syscall.SIGTERM)
	err = graceful.ListenAndServe(conf.Bind, api.New(conf))
	if err != nil {
		log.Fatal(err)
	}
	graceful.Wait()
}
