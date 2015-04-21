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

	"github.com/pressly/qmd"
	"github.com/pressly/qmd/api"
)

func JobHandler() http.Handler {
	s := web.New()
	s.Use(middleware.SubRouter)

	s.Get("/", ListJobs)
	s.Get("/:id", Job)
	s.Get("/:id/", Job)
	s.Get("/:id/result", JobResult)
	s.Get("/:id/stdout", JobStdout)
	s.Get("/:id/stderr", JobStderr)
	s.Get("/:id/QMD_OUT", JobQmdOut)

	return s
}

func ListJobs(c web.C, w http.ResponseWriter, r *http.Request) {
	var running, enqueued, finished, orphans []*qmd.Job
	for _, job := range Qmd.Jobs {
		switch job.State {
		case qmd.Running:
			running = append(running, job)
		case qmd.Enqueued:
			enqueued = append(enqueued, job)
		case qmd.Finished:
			finished = append(finished, job)
		default:
			orphans = append(orphans, job)
		}
	}

	fmt.Fprintf(w, `<h1>Running (%v jobs)</h1>`, len(running))
	for _, job := range running {
		fmt.Fprintf(w, `<a href="/jobs/%v">Job %v</a> (`, job.ID, job.ID)
		fmt.Fprintf(w, `<a href="/jobs/%v/result">result</a>, `, job.ID)
		fmt.Fprintf(w, `<a href="/jobs/%v/stdout">stdout</a>, `, job.ID)
		fmt.Fprintf(w, `<a href="/jobs/%v/stderr">stderr</a>, `, job.ID)
		fmt.Fprintf(w, `<a href="/jobs/%v/QMD_OUT">QMD_OUT</a>)<br>`, job.ID)
	}

	fmt.Fprintf(w, `<h1>Enqueued(%v jobs)</h1>`, len(enqueued))
	for _, job := range enqueued {
		fmt.Fprintf(w, `<a href="/jobs/%v">Job %v</a> (`, job.ID, job.ID)
		fmt.Fprintf(w, `<a href="/jobs/%v/result">result</a>, `, job.ID)
		fmt.Fprintf(w, `<a href="/jobs/%v/stdout">stdout</a>, `, job.ID)
		fmt.Fprintf(w, `<a href="/jobs/%v/stderr">stderr</a>, `, job.ID)
		fmt.Fprintf(w, `<a href="/jobs/%v/QMD_OUT">QMD_OUT</a>)<br>`, job.ID)
	}

	fmt.Fprintf(w, `<h1>Finished (%v jobs)</h1>`, len(finished))
	for _, job := range finished {
		fmt.Fprintf(w, `<a href="/jobs/%v">Job %v</a> (`, job.ID, job.ID)
		fmt.Fprintf(w, `<a href="/jobs/%v/result">result</a>, `, job.ID)
		fmt.Fprintf(w, `<a href="/jobs/%v/stdout">stdout</a>, `, job.ID)
		fmt.Fprintf(w, `<a href="/jobs/%v/stderr">stderr</a>, `, job.ID)
		fmt.Fprintf(w, `<a href="/jobs/%v/QMD_OUT">QMD_OUT</a>)<br>`, job.ID)
	}

	fmt.Fprintf(w, `<h1>Orphans (%v jobs)</h1>`, len(orphans))
	for _, job := range orphans {
		fmt.Fprintf(w, `<a href="/jobs/%v">Job %v</a> (`, job.ID, job.ID)
		fmt.Fprintf(w, `<a href="/jobs/%v/result">result</a>, `, job.ID)
		fmt.Fprintf(w, `<a href="/jobs/%v/stdout">stdout</a>, `, job.ID)
		fmt.Fprintf(w, `<a href="/jobs/%v/stderr">stderr</a>, `, job.ID)
		fmt.Fprintf(w, `<a href="/jobs/%v/QMD_OUT">QMD_OUT</a>)<br>`, job.ID)
	}

}

func Job(c web.C, w http.ResponseWriter, r *http.Request) {
	job, err := Qmd.GetJob(c.URLParams["id"])
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}

	fmt.Fprintf(w, `<h1>Job %v [%s]</h1>`, job.ID, job.State)
	fmt.Fprintf(w, `<a href="/jobs/%v/result">result</a><br>`, job.ID)
	fmt.Fprintf(w, `<a href="/jobs/%v/stdout">stdout</a><br>`, job.ID)
	fmt.Fprintf(w, `<a href="/jobs/%v/stderr">stderr</a><br>`, job.ID)
	fmt.Fprintf(w, `<a href="/jobs/%v/QMD_OUT">QMD_OUT</a>`, job.ID)
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
	qmdOut, _ := ioutil.ReadFile(job.QmdOutFile)

	resp := api.ScriptsResponse{
		ID:          job.ID,
		Script:      job.Args[0],
		Args:        job.Args[1:],
		CallbackURL: job.CallbackURL,
		Status:      status,
		StartTime:   job.StartTime,
		EndTime:     job.EndTime,
		Duration:    fmt.Sprintf("%f", job.Duration.Seconds()),
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

	file, err := os.Open(job.QmdOutFile)
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
