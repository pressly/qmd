package server

import (
	"fmt"
	"net/http"

	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/middleware"
)

func JobsHandler() http.Handler {
	s := web.New()
	s.Use(middleware.SubRouter)

	s.Get("/", Jobs)
	s.Get("/:id", Job)

	return s
}

func Job(c web.C, w http.ResponseWriter, r *http.Request) {
	resp, err := Qmd.GetResponse(c.URLParams["id"])
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}
	w.Write(resp)
}

func Jobs(c web.C, w http.ResponseWriter, r *http.Request) {
	low, _ := Qmd.Queue.Len("low")
	high, _ := Qmd.Queue.Len("high")
	urgent, _ := Qmd.Queue.Len("urgent")
	cached, _ := Qmd.DB.Len()
	finished, _ := Qmd.DB.TotalLen()

	r.Header.Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "Enqueued: %v\n- %v (urgent)\n- %v (high)\n- %v (low)\n\n", urgent+high+low, urgent, high, low)
	fmt.Fprintf(w, "Running: TODO(https://github.com/antirez/disque/issues/48)\n\n")
	fmt.Fprintf(w, "In-cache: %v\n\n", cached)
	fmt.Fprintf(w, "Finished: %v", finished)
}
