package server

import (
	"net/http"

	"github.com/pressly/gohttpware/heartbeat"
	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/middleware"

	"github.com/pressly/qmd"
)

var Qmd *qmd.Qmd

func APIHandler(qmd *qmd.Qmd) http.Handler {
	Qmd = qmd
	h := web.New()

	h.Use(middleware.EnvInit)
	h.Use(middleware.RequestID)
	h.Use(middleware.RealIP)
	h.Use(middleware.NoCache)

	if Qmd.Config.Environment != "testing" {
		h.Use(middleware.Logger)
	}
	h.Use(middleware.Recoverer)

	h.Use(heartbeat.Route("/ping"))
	h.Use(heartbeat.Route("/"))

	h.Handle("/scripts/*", ScriptsHandler())
	h.Get("/jobs", ListJobs)
	h.Handle("/jobs/*", JobHandler())

	return h
}
