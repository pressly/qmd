package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
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
	s.Get("/:id/stderr", JobStderr)
	s.Get("/:id/qmd_out", JobQmdOut)

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

	stdout, _ := ioutil.ReadFile(job.StdoutFile)
	//stderr, _ := ioutil.ReadFile(job.StderrFile)
	qmdOut, _ := ioutil.ReadFile(job.ExtraOutFile)

	resp := api.ScriptsResponse{
		ID:          job.ID,
		Script:      job.Args[0],
		Args:        job.Args[1:],
		CallbackURL: job.CallbackURL,
		Status:      status,
		StartTime:   job.StartTime,
		EndTime:     job.EndTime,
		Duration:    fmt.Sprintf("%d", job.Duration),
		Output:      string(qmdOut),
		ExecLog:     string(stdout),
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

	job.WaitForStart()

	file, err := os.Open(job.StdoutFile)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Wait for the job to finish.
	//TODO: This should be at the end of this file to allow streaming.
	job.Wait()

	streamData(w, file)
}

func JobStderr(c web.C, w http.ResponseWriter, r *http.Request) {
	job, err := Qmd.GetJob(c.URLParams["id"])
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}

	job.WaitForStart()

	file, err := os.Open(job.StderrFile)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Wait for the job to finish.
	//TODO: This should be at the end of this file to allow streaming.
	job.Wait()

	streamData(w, file)
}

func JobQmdOut(c web.C, w http.ResponseWriter, r *http.Request) {
	job, err := Qmd.GetJob(c.URLParams["id"])
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}

	job.WaitForStart()

	file, err := os.Open(job.ExtraOutFile)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Wait for the job to finish.
	//TODO: This should be at the end of this file to allow streaming.
	job.Wait()

	streamData(w, file)
}

func streamData(w io.Writer, r io.Reader) {
	// Send the job's STDOUT over HTTP as soon as possible.
	// Make the HTTP streaming possible by flushing each line.
	doneStreaming := make(chan struct{}, 0)
	go func() {
		defer func() {
			close(doneStreaming)
		}()

		if f, ok := w.(http.Flusher); ok {
			r := bufio.NewReader(r)
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
			io.Copy(w, r)
		}
	}()

	<-doneStreaming
}
