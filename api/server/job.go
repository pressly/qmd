package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sort"

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
	var running, enqueued, finished, interrupted, failed, initialized []string

	Qmd.MuJobs.Lock()
	defer Qmd.MuJobs.Unlock()
	for _, job := range Qmd.Jobs {
		switch job.State {
		case qmd.Initialized:
			initialized = append(initialized, job.ID)
		case qmd.Running:
			running = append(running, job.ID)
		case qmd.Enqueued:
			enqueued = append(enqueued, job.ID)
		case qmd.Finished:
			finished = append(finished, job.ID)
		case qmd.Failed:
			failed = append(failed, job.ID)
		case qmd.Interrupted:
			interrupted = append(interrupted, job.ID)
		default:
			panic("unreachable")
		}
	}
	sort.Strings(running)
	sort.Strings(enqueued)
	sort.Strings(finished)
	sort.Strings(interrupted)
	sort.Strings(failed)
	sort.Strings(initialized)

	fmt.Fprintf(w, `<table><tr><th>Running</th><th>Enqueued</th><th>Finished</th><th>Interrupted</th><th>Orhans</th></tr>`)
	fmt.Fprintf(w, `<tr><td>%d</td><td>%d</td><td>%d</td><td>%d</td><td>%d</td></tr></table>`, len(running), len(enqueued), len(finished), len(interrupted), len(initialized))

	fmt.Fprintf(w, `<h1>Running (%v jobs)</h1>`, len(running))
	for _, id := range running {
		job := Qmd.Jobs[id]
		fmt.Fprintf(w, `<a href="/jobs/%v">Job %v</a> (`, job.ID, job.ID)
		fmt.Fprintf(w, `<a href="/jobs/%v/result">result</a>, `, job.ID)
		fmt.Fprintf(w, `<a href="/jobs/%v/stdout">stdout</a>, `, job.ID)
		fmt.Fprintf(w, `<a href="/jobs/%v/stderr">stderr</a>, `, job.ID)
		fmt.Fprintf(w, `<a href="/jobs/%v/QMD_OUT">QMD_OUT</a>)<br>`, job.ID)
	}

	fmt.Fprintf(w, `<h1>Enqueued - waiting (%v jobs)</h1>`, len(enqueued))
	for _, id := range enqueued {
		job := Qmd.Jobs[id]
		fmt.Fprintf(w, `<a href="/jobs/%v">Job %v</a> (`, job.ID, job.ID)
		fmt.Fprintf(w, `<a href="/jobs/%v/result">result</a>, `, job.ID)
		fmt.Fprintf(w, `<a href="/jobs/%v/stdout">stdout</a>, `, job.ID)
		fmt.Fprintf(w, `<a href="/jobs/%v/stderr">stderr</a>, `, job.ID)
		fmt.Fprintf(w, `<a href="/jobs/%v/QMD_OUT">QMD_OUT</a>)<br>`, job.ID)
	}

	fmt.Fprintf(w, `<h1>Finished (%v jobs)</h1>`, len(finished))
	for _, id := range finished {
		job := Qmd.Jobs[id]
		fmt.Fprintf(w, `<a href="/jobs/%v">Job %v</a> (`, job.ID, job.ID)
		fmt.Fprintf(w, `<a href="/jobs/%v/result">result</a>, `, job.ID)
		fmt.Fprintf(w, `<a href="/jobs/%v/stdout">stdout</a>, `, job.ID)
		fmt.Fprintf(w, `<a href="/jobs/%v/stderr">stderr</a>, `, job.ID)
		fmt.Fprintf(w, `<a href="/jobs/%v/QMD_OUT">QMD_OUT</a>)<br>`, job.ID)
	}

	fmt.Fprintf(w, `<h1>Interrupted (%v jobs)</h1>`, len(interrupted))
	for _, id := range interrupted {
		job := Qmd.Jobs[id]
		fmt.Fprintf(w, `<a href="/jobs/%v">Job %v</a> (`, job.ID, job.ID)
		fmt.Fprintf(w, `<a href="/jobs/%v/result">result</a>, `, job.ID)
		fmt.Fprintf(w, `<a href="/jobs/%v/stdout">stdout</a>, `, job.ID)
		fmt.Fprintf(w, `<a href="/jobs/%v/stderr">stderr</a>, `, job.ID)
		fmt.Fprintf(w, `<a href="/jobs/%v/QMD_OUT">QMD_OUT</a>)<br>`, job.ID)
	}

	fmt.Fprintf(w, `<h1>Failed to start (%v jobs)</h1>`, len(failed))
	for _, id := range failed {
		job := Qmd.Jobs[id]
		fmt.Fprintf(w, `<a href="/jobs/%v">Job %v</a> (`, job.ID, job.ID)
		fmt.Fprintf(w, `<a href="/jobs/%v/result">result</a>, `, job.ID)
		fmt.Fprintf(w, `<a href="/jobs/%v/stdout">stdout</a>, `, job.ID)
		fmt.Fprintf(w, `<a href="/jobs/%v/stderr">stderr</a>, `, job.ID)
		fmt.Fprintf(w, `<a href="/jobs/%v/QMD_OUT">QMD_OUT</a>)<br>`, job.ID)
	}

	fmt.Fprintf(w, `<h1>Orphans (%v jobs)</h1>`, len(initialized))
	for _, id := range initialized {
		job := Qmd.Jobs[id]
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

func JobStdout(c web.C, w http.ResponseWriter, r *http.Request) {
	job, err := Qmd.GetJob(c.URLParams["id"])
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}

	<-job.Started

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

	<-job.Started

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

	<-job.Started

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
	doneStreaming := make(chan struct{})
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
