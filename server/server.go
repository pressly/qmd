package server

import (
	_ "expvar"
	"net/http"

	"github.com/op/go-logging"
	"github.com/pressly/gohttpware/heartbeat"
	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/middleware"

	"github.com/pressly/qmd"
)

var (
	Qmd *qmd.Qmd
	lg  *logging.Logger
)

func APIHandler(qmd *qmd.Qmd) http.Handler {
	Qmd = qmd
	lg = qmd.Logger

	h := web.New()

	h.Use(middleware.EnvInit)
	h.Use(middleware.RequestID)
	h.Use(middleware.RealIP)
	h.Use(middleware.NoCache)
	h.Use(middleware.Recoverer)
	h.Use(middleware.Logger)
	h.Use(qmd.ClosingResponder)

	h.Use(heartbeat.Route("/ping"))
	h.Use(heartbeat.Route("/"))

	h.Handle("/scripts/*", ScriptsHandler())
	h.Handle("/jobs/*", JobsHandler())

	h.Handle("/debug/vars", http.DefaultServeMux)

	return h
}
