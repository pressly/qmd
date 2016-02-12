package rest

import (
	"net/http"

	"github.com/pressly/gohttpware/heartbeat"
	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/middleware"

	"github.com/pressly/qmd"
	"github.com/pressly/qmd/rest/handlers"
)

func Routes(qmd *qmd.Qmd) http.Handler {
	handlers.Qmd = qmd

	r := web.New()

	r.Use(middleware.EnvInit)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.NoCache)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)
	r.Use(qmd.ClosingResponder)

	r.Use(heartbeat.Route("/ping"))
	r.Use(heartbeat.Route("/"))

	r.Post("/:filename", handlers.CreateJob)

	r.Get("/jobs", handlers.Jobs)
	r.Get("/jobs/:id", handlers.Job)

	return r
}
