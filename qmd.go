package qmd

import (
	"log"
	"syscall"

	"github.com/zenazn/goji/graceful"

	"github.com/pressly/qmd/api"
	"github.com/pressly/qmd/config"
	"github.com/pressly/qmd/job"
)

var App *QMD

type QMD struct {
	Config     *config.Config
	Controller *job.Controller
}

func Run(conf *config.Config) error {
	var err error

	App := &QMD{
		Config: conf,
	}

	log.Print("Starting QMD Job Controller")

	// Create Job Controller.
	App.Controller, err = job.NewController(conf)
	if err != nil {
		return err
	}

	log.Printf("Starting QMD API at %s\n", conf.Bind)

	// Start the API server.
	graceful.AddSignal(syscall.SIGINT, syscall.SIGTERM)
	err = graceful.ListenAndServe(conf.Bind, api.New(conf))
	if err != nil {
		return err
	}
	graceful.Wait()

	return nil
}
