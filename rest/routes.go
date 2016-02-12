package rest

import (
	"net/http"
	"strings"

	"golang.org/x/net/context"

	"github.com/pressly/chi"
	"github.com/pressly/chi/middleware"

	"github.com/pressly/qmd"
	"github.com/pressly/qmd/rest/handlers"
)

func Routes(qmd *qmd.Qmd) http.Handler {
	handlers.Qmd = qmd

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.NoCache)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)
	r.Use(qmd.ClosingResponder)

	r.Get("/", handlers.Index)
	r.Get("/ping", handlers.Ping)

	r.Post("/scripts/:filename", handlers.CreateJob)

	r.Get("/jobs", handlers.Jobs)
	r.Get("/jobs/*", GetLongID, handlers.Job)

	return r
}

func GetLongID(next chi.Handler) chi.Handler {
	fn := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		ctx = context.WithValue(ctx, "id", strings.TrimPrefix(r.RequestURI, "/jobs/"))

		next.ServeHTTPC(ctx, w, r)
	}

	return chi.HandlerFunc(fn)
}
