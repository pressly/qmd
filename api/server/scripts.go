package server

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
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

	log.Printf("%v starting job", c.URLParams["filename"])

	// Start the job.
	err = job.Start()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Send the job's STDOUT over HTTP as soon as possible.
	// Make the HTTP streaming possible by flushing each line.
	go func() {
		if f, ok := w.(http.Flusher); ok {
			r := bufio.NewReader(job.Stdout)
			for {
				line, err := r.ReadBytes('\n')
				w.Write(line)
				f.Flush()
				if err != nil {
					return
				}
			}
		} else {
			io.Copy(w, job.Stdout)
		}
	}()

	// Wait for the job to finish.
	err = job.Wait()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	log.Printf("%v finished in %v", c.URLParams["filename"], job.Duration)
}
