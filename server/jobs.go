package server

import (
	"net/http"

	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/middleware"
)

func JobsHandler() http.Handler {
	s := web.New()
	s.Use(middleware.SubRouter)

	s.Get("/:id", Job)

	return s
}

func Job(c web.C, w http.ResponseWriter, r *http.Request) {
	resp, err := Qmd.Wait(c.URLParams["id"])
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}
	w.Write(resp)
}
