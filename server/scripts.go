package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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

	job.ExtraWorkDirFiles = req.Files

	// Enqueue job.
	Qmd.Enqueue(job)

	// Kill the job, if client closes the connection before
	// it receives the data.
	done := make(chan struct{})
	defer close(done)
	connClosed := w.(http.CloseNotifier).CloseNotify()
	go func() {
		select {
		case <-connClosed:
			job.Kill()
		case <-done:
		}
	}()

	// Wait for the job to finish.
	_ = job.Wait()

	stdout, _ := ioutil.ReadFile(job.StdoutFile)
	//stderr, _ := ioutil.ReadFile(job.StderrFile)
	qmdOut, _ := ioutil.ReadFile(job.QmdOutFile)

	var status string
	if job.StatusCode == 0 {
		// "OK" for backward compatibility.
		status = "OK"
	} else {
		status = fmt.Sprintf("%v", job.StatusCode)
	}

	resp := api.ScriptsResponse{
		ID: job.ID,
		//TODO: We probably don't need those in response:
		// Script:      job.Args[0],
		// Args:        job.Args[1:],
		// Files:
		CallbackURL: job.CallbackURL,
		Status:      status,
		StartTime:   job.StartTime,
		EndTime:     job.EndTime,
		Duration:    fmt.Sprintf("%f", job.Duration.Seconds()),
		QmdOut:      string(qmdOut),
		ExecLog:     string(stdout),
	}

	enc := json.NewEncoder(w)
	err = enc.Encode(resp)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}
