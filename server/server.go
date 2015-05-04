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
	h.Use(middleware.Recoverer)
	h.Use(middleware.Logger)

	h.Use(heartbeat.Route("/ping"))
	h.Use(heartbeat.Route("/"))

	h.Handle("/scripts/*", ScriptsHandler())
	h.Handle("/jobs/*", JobsHandler())

	return h
}
