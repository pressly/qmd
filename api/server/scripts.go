package server

import (
	"encoding/json"
	"io"
	"net/http"
	"os/exec"

	"github.com/zenazn/goji/web"

	"github.com/pressly/qmd"
	"github.com/pressly/qmd/api"
)

func CreateSyncJob(c web.C, w http.ResponseWriter, r *http.Request) {
	// Decode request data.
	var req *api.ScriptsRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "parse request body: "+err.Error(), 422)
		return
	}

	// Get script path.
	script, err := App.GetScript(c.URLParams["filename"])
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}

	// Create QMD job to run the command.
	cmd := exec.Command(script, req.Args...)
	job, err := qmd.Job(cmd)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Start the job.
	err = job.Start()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Send the job's STDOUT over HTTP as soon as possible.
	// Wait closes the job.Stdout, so the .
	go io.Copy(w, job.Stdout)

	err = job.Wait()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}
