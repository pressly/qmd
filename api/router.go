package api

import (
	"net/http"

	"github.com/goware/throttler"
	"github.com/pressly/gohttpware/heartbeat"
	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/middleware"

	"github.com/pressly/qmd/api/script"
	"github.com/pressly/qmd/config"
)

func New(conf *config.Config) http.Handler {
	r := web.New()

	r.Use(middleware.EnvInit)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.NoCache)

	if conf.Environment != "testing" {
		r.Use(middleware.Logger)
	}
	r.Use(middleware.Recoverer)

	r.Use(heartbeat.Route("/ping"))
	r.Use(heartbeat.Route("/"))

	scriptRoute := script.Router()
	scriptRoute.Use(throttler.Limit(conf.MaxJobs))
	r.Handle("/scripts/*", scriptRoute)

	return r
}
