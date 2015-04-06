package script

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"os/exec"

	"github.com/zenazn/goji/web"

	"github.com/pressly/qmd/job"
	"github.com/pressly/qmd/script"
)

func CreateJob(c web.C, w http.ResponseWriter, r *http.Request) {
	dump, err := httputil.DumpRequest(r, true)
	log.Printf("%s", dump)

	var req *Request
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), 422)
		return
	}

	filename := c.URLParams["filename"]
	log.Print(filename, req.Args, req.Files, req.CallbackURL)

	script, err := script.Ctl.Get(filename)
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}

	cmd := exec.Command(script, req.Args...)
	job, err := job.New(cmd)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	result := new(bytes.Buffer)
	go result.ReadFrom(job.Stdout)

	err = job.Run()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Write(result.Bytes())
}
