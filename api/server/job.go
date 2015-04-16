package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os/exec"
	"syscall"

	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/middleware"

	"github.com/pressly/qmd/api"
)

func JobHandler() http.Handler {
	s := web.New()
	s.Use(middleware.SubRouter)

	s.Get("/:id/result", JobResult)
	s.Get("/:id/stdout", JobStdout)

	return s
}

func JobResult(c web.C, w http.ResponseWriter, r *http.Request) {
	job, err := Qmd.GetJob(c.URLParams["id"])
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}

	// Wait for the job to finish.
	err = job.Wait()
	status := "OK"
	if err != nil {
		if e, ok := err.(*exec.ExitError); ok {
			if s, ok := e.Sys().(syscall.WaitStatus); ok {
				if s.ExitStatus() != 0 {
					status = fmt.Sprintf("%d", s.ExitStatus())
				}
			}
		}
	}

	output, _ := ioutil.ReadAll(job.Stdout)

	resp := api.ScriptsResponse{
		ID:          job.ID,
		Script:      job.Args[0],
		Args:        job.Args[1:],
		CallbackURL: job.CallbackURL,
		Status:      status,
		StartTime:   job.StartTime,
		EndTime:     job.EndTime,
		Duration:    job.Duration,
		Output:      string(output),
		//TODO: ExecLog:
		ExecLog: "TODO",
	}

	enc := json.NewEncoder(w)
	err = enc.Encode(resp)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func JobStdout(c web.C, w http.ResponseWriter, r *http.Request) {
	job, err := Qmd.GetJob(c.URLParams["id"])
	if err != nil {
		http.Error(w, err.Error(), 404)
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
				if err != io.EOF {
					f.Flush()
				}
				if err != nil {
					return
				}
			}
		} else {
			io.Copy(w, job.Stdout)
		}
	}()

	// Wait for the job to finish.
	job.Wait()
}
