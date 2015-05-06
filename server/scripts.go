package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/goware/urlx"
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
	// Low, high and urgent priorities only (high is default).
	priority := r.URL.Query().Get("priority")
	switch priority {
	case "low", "high", "urgent":
		// NOP.
	case "":
		priority = "high"
	default:
		http.Error(w, "unknown priority \""+priority+"\"", 422)
		return
	}

	// Decode request data.
	var req *api.ScriptsRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "parse request body: "+err.Error(), 422)
		return
	}
	req.Script = c.URLParams["filename"]

	// Make sure ASYNC callback is valid URL.
	if req.CallbackURL != "" {
		req.CallbackURL, err = urlx.NormalizeString(req.CallbackURL)
		if err != nil {
			http.Error(w, "parse request body: "+err.Error(), 422)
			return
		}
	}

	// Enqueue the request.
	data, err := json.Marshal(req)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	log.Printf("Handler:\tEnqueue request \"%v\"", priority)
	job, err := Qmd.Enqueue(string(data), priority)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Async.
	if req.CallbackURL != "" {
		resp, _ := Qmd.GetAsyncResponse(req, job.ID)
		w.Write(resp)
		log.Printf("Handler:\tResponded with job %s ASYNC result", job.ID)

		go func() {
			//TODO: Retry?
			err := Qmd.PostResponseCallback(req, job.ID)
			if err != nil {
				log.Printf("can't post callback to %v", err)
			}
		}()
		return
	}

	// Sync.
	log.Printf("Handler:\tWaiting for job %s", job.ID)

	resp, _ := Qmd.GetResponse(job.ID)
	w.Write(resp)

	log.Printf("Handler:\tResponded with job %s result", job.ID)

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
