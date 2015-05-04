package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/middleware"

	"github.com/pressly/qmd/api"
)

func ScriptsHandler() http.Handler {
	s := web.New()
	s.Use(middleware.SubRouter)
	s.Post("/:filename", CreateJob)
	return s
}

func CreateJob(c web.C, w http.ResponseWriter, r *http.Request) {
	// Decode request data.
	var req *api.ScriptsRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "parse request body: "+err.Error(), 422)
		return
	}
	req.Script = c.URLParams["filename"]

	// Enqueue the request.
	data, err := json.Marshal(req)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// "low", "high" and "urgent" priorities
	// "high" priority by default
	priority := r.URL.Query().Get("priority")
	switch priority {
	case "low", "high", "urgent":
	default:
		priority = "high"
	}

	job, err := Qmd.Enqueue(string(data), priority)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	log.Printf("Handler: Enqueue job %v (\"%v priority\")", job.ID, job.Queue)

	log.Printf("Handler: Wait for job %s", job.ID)
	resp, err := Qmd.Wait(job.ID)
	w.Write(resp)
	log.Printf("Handler: Responded with job %s result", job.ID)

	// 	// Kill the job, if client closes the connection before
	// 	// it receives the data.
	// 	done := make(chan struct{})
	// 	defer close(done)
	// 	connClosed := w.(http.CloseNotifier).CloseNotify()
	// 	go func() {
	// 		select {
	// 		case <-connClosed:
	// 			job.Kill()
	// 		case <-done:
	// 		}
	// 	}()
}
