package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"

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

	job.CallbackURL = req.CallbackURL
	job.ExtraWorkDirFiles = req.Files

	// Enqueue job.
	Qmd.Enqueue(job)

	// Response.
	resp := api.ScriptsResponse{
		ID: job.ID,
		//TODO: These are only for backward-compatibility, we don't need them.
		Script: c.URLParams["filename"],
		Args:   req.Args,
		Files:  req.Files,
	}

	if req.CallbackURL == "" {
		// Sync job. Wait for result.

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

		log.Printf("wait")
		// Wait for the job to finish.
		_ = job.Wait()
		log.Printf("wait finished")

		if job.StatusCode == 0 {
			// "OK" for backward compatibility.
			resp.Status = "OK"
		} else {
			resp.Status = fmt.Sprintf("%v", job.StatusCode)
		}

		resp.EndTime = job.EndTime
		resp.Duration = fmt.Sprintf("%f", job.Duration.Seconds())
		//resp.QmdOut = job.QmdOut.String()
		qmdOut, _ := ioutil.ReadFile(job.QmdOutFile)
		resp.QmdOut = string(qmdOut)
		resp.ExecLog = job.CmdOut.String()
		resp.StartTime = job.StartTime

	} else {
		// Async job. Don't wait on anything.
		resp.Status = "QUEUED"
		resp.CallbackURL = req.CallbackURL
	}

	// Return response.
	enc := json.NewEncoder(w)
	err = enc.Encode(resp)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}
