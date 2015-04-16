package server

import (
	"encoding/json"
	"net/http"
	"os/exec"

	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/middleware"

	"github.com/pressly/qmd/api"
)

func ScriptsHandler() http.Handler {
	s := web.New()
	s.Use(middleware.SubRouter)
	s.Post("/:filename", CreateSyncJob)
	return s
}

func CreateSyncJob(c web.C, w http.ResponseWriter, r *http.Request) {
	// Decode request data.
	var req *api.ScriptsRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "parse request body: "+err.Error(), 422)
		return
	}

	//log.Printf("\n\nreq: %#v\n\n", req)

	// Get script path.
	script, err := Qmd.GetScript(c.URLParams["filename"])
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}

	// Create QMD job to run the command.
	cmd := exec.Command(script, req.Args...)
	job, err := Qmd.Job(cmd)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Enqueue job.
	Qmd.Enqueue(job)

	//TODO: This is Sync operation, so if the client closes the request,
	//      before getting the response, we should kill the job.
	//      Use w.(http.CloseNotifier).

	// Redirect to the actual /job/:id/result handler.
	// Post/Redirect/Get pattern would be too expensive.
	c.URLParams["id"] = job.ID
	JobResult(c, w, r)
}
