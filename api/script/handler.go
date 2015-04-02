package script

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"

	"github.com/zenazn/goji/web"

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

	//TODO: Do the actual work.

	//TODO: Return Response instead.
	w.Write([]byte(script))
}
