package server

import (
	"net/http"

	"github.com/goware/throttler"
	"github.com/pressly/gohttpware/heartbeat"
	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/middleware"

	"github.com/pressly/qmd"
)

var App *qmd.Qmd

func APIHandler(app *qmd.Qmd) http.Handler {
	App = app
	r := web.New()

	r.Use(middleware.EnvInit)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.NoCache)

	if App.Config.Environment != "testing" {
		r.Use(middleware.Logger)
	}
	r.Use(middleware.Recoverer)

	r.Use(heartbeat.Route("/ping"))
	r.Use(heartbeat.Route("/"))

	s := web.New()
	s.Use(middleware.SubRouter)
	s.Post("/:filename", CreateSyncJob)
	s.Use(throttler.Limit(App.Config.MaxJobs))

	r.Handle("/scripts/*", s)

	return r
}
