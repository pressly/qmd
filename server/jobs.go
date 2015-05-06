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
	low, _ := Qmd.Len("low")
	high, _ := Qmd.Len("high")
	urgent, _ := Qmd.Len("urgent")

	fmt.Fprintf(w, "Enqueued (total):\n- %v (urgent)\n- %v (high)\n- %v (low)\n\n", urgent, high, low)

}
