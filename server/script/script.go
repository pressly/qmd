package script

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"

	"github.com/pressly/qmd/api/script"
	"github.com/zenazn/goji/web"
)

func CreateJob(c web.C, w http.ResponseWriter, r *http.Request) {
	dump, err := httputil.DumpRequest(r, true)
	log.Printf("%s", dump)

	var req *script.Request
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), 422)
		return
	}

	filename := c.URLParams["filename"]
	log.Print(filename, req.Args, req.Files, req.CallbackURL)

	//TODO: Do the actual work.

	//TODO: Return Response instead.
	w.Write([]byte("OK"))
}
