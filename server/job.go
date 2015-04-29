package server

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/middleware"

	"github.com/pressly/qmd"
)

func JobHandler() http.Handler {
	s := web.New()
	s.Use(middleware.SubRouter)

	s.Get("/", ListJobs)
	s.Get("/:id", Job)
	s.Get("/:id/", Job)

	return s
}

func ListJobs(c web.C, w http.ResponseWriter, r *http.Request) {
	var running, finished, terminated, invalidated, failed []string

	Qmd.MuJobs.Lock()
	defer Qmd.MuJobs.Unlock()
	for _, job := range Qmd.Jobs {
		switch job.State {
		case qmd.Running:
			running = append(running, job.ID)
		case qmd.Finished:
			finished = append(finished, job.ID)
		case qmd.Terminated:
			terminated = append(terminated, job.ID)
		case qmd.Invalidated:
			invalidated = append(invalidated, job.ID)
		case qmd.Failed:
			failed = append(failed, job.ID)
		default:
			panic("unreachable")
		}
	}
	sort.Strings(running)
	sort.Strings(finished)
	sort.Strings(terminated)
	sort.Strings(failed)
	sort.Strings(invalidated)

	fmt.Fprintf(w, `<table border="1"><tr><th>Running</th><th>Finished</th><th>Terminated</th><th>Invalidated</th><th>Failed to start</th></tr>`)
	fmt.Fprintf(w, `<tr><td>%d</td><td>%d</td><td>%d</td><td>%d</td><td>%d</td></tr></table>`, len(running), len(finished), len(terminated), len(invalidated), len(failed))

	fmt.Fprintf(w, `<h1>Running (%v jobs)</h1>`, len(running))
	for _, id := range running {
		job := Qmd.Jobs[id]
		fmt.Fprintf(w, `<a href="/jobs/%v">Job %v</a> (`, job.ID, job.ID)
		fmt.Fprintf(w, `<a href="/jobs/%v/result">result</a>, `, job.ID)
		fmt.Fprintf(w, `<a href="/jobs/%v/logs">logs</a>)<br>`, job.ID)
	}

	fmt.Fprintf(w, `<h1>Finished (%v jobs)</h1>`, len(finished))
	for _, id := range finished {
		job := Qmd.Jobs[id]
		fmt.Fprintf(w, `<a href="/jobs/%v">Job %v</a> (`, job.ID, job.ID)
		fmt.Fprintf(w, `<a href="/jobs/%v/result">result</a>, `, job.ID)
		fmt.Fprintf(w, `<a href="/jobs/%v/logs">logs</a>)<br>`, job.ID)
	}

	fmt.Fprintf(w, `<h1>Terminated (%v jobs)</h1>`, len(terminated))
	for _, id := range terminated {
		job := Qmd.Jobs[id]
		fmt.Fprintf(w, `<a href="/jobs/%v">Job %v</a> (`, job.ID, job.ID)
		fmt.Fprintf(w, `<a href="/jobs/%v/result">result</a>, `, job.ID)
		fmt.Fprintf(w, `<a href="/jobs/%v/logs">logs</a>)<br>`, job.ID)
	}

	fmt.Fprintf(w, `<h1>Terminated before start (%v jobs)</h1>`, len(invalidated))
	for _, id := range invalidated {
		job := Qmd.Jobs[id]
		fmt.Fprintf(w, `<a href="/jobs/%v">Job %v</a> (`, job.ID, job.ID)
		fmt.Fprintf(w, `<a href="/jobs/%v/result">result</a>, `, job.ID)
		fmt.Fprintf(w, `<a href="/jobs/%v/logs">logs</a>)<br>`, job.ID)
	}

	fmt.Fprintf(w, `<h1>Failed to start (%v jobs)</h1>`, len(failed))
	for _, id := range failed {
		job := Qmd.Jobs[id]
		fmt.Fprintf(w, `<a href="/jobs/%v">Job %v</a> (`, job.ID, job.ID)
		fmt.Fprintf(w, `<a href="/jobs/%v/result">result</a>, `, job.ID)
		fmt.Fprintf(w, `<a href="/jobs/%v/logs">logs</a>)<br>`, job.ID)
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
	fmt.Fprintf(w, `<a href="/jobs/%v/logs">logs</a>`, job.ID)
}
