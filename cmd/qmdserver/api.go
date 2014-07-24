package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"github.com/pressly/qmd"
	"github.com/zenazn/goji/web"
)

// Handle and reverse proxy the admin route to nsqadmin.
func adminProxy(c web.C, w http.ResponseWriter, r *http.Request) {
	targetURL := c.Env["adminAddr"].(string)
	if targetURL == "" {
		log.Error("No admin panel found")
		http.Error(w, "No admin panel found", http.StatusNotFound)
		return
	}
	if !strings.Contains(targetURL, "http://") {
		targetURL = fmt.Sprintf("%s%s", "http://", "0.0.0.0:4171")
	}
	u, err := url.Parse(targetURL)
	if err != nil {
		w.Write([]byte(err.Error()))
	}
	p := httputil.NewSingleHostReverseProxy(u)
	p.ServeHTTP(w, r)
}

// Reload the whitelist of scripts.
func reloadScripts(c web.C, w http.ResponseWriter, r *http.Request) {
	server := c.Env["server"].(*qmd.Server)

	err := server.Reload()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var buf bytes.Buffer
	buf.WriteString("Reload request sent")

	w.Header().Set("Content-Type", "application/json")
	w.Write(buf.Bytes())
}

// Send details to queue for execution.
func runScript(c web.C, w http.ResponseWriter, r *http.Request) {
	// Parse the request
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Queue up the request
	server := c.Env["server"].(*qmd.Server)
	ch, err := server.Queue(c.URLParams["name"], body)
	if err != nil {
		log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(<-ch)
}

// Retrieve all logs for a specific script.
func getAllLogs(c web.C, w http.ResponseWriter, r *http.Request) {
	var err error
	var limit, offset int

	server := c.Env["server"].(*qmd.Server)
	script := c.URLParams["name"]
	params := r.URL.Query()

	if len(params["limit"]) > 0 {
		limit, err = strconv.Atoi(params["limit"][0])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	fmt.Println(limit, offset)

	reply, err := server.DB.GetLogs(script, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	buf.WriteString("[")

	length := len(reply)
	if length > 0 {
		buf.WriteString(reply[length-1])
		for i := length - 2; i > 0; i-- {
			buf.WriteString(", ")
			buf.WriteString(reply[i])
		}
	}
	buf.WriteString("]")
	w.Header().Set("Content-Type", "application/json")
	w.Write(buf.Bytes())
}

// Retrieve a specific log of a specific script.
func getLog(c web.C, w http.ResponseWriter, r *http.Request) {
	server := c.Env["server"].(*qmd.Server)
	script := c.URLParams["name"]
	id := c.URLParams["id"]

	reply, err := server.DB.GetLog(script, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(reply))
}
