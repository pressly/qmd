package main

import (
	"net/http"

	"github.com/bitly/go-nsq"
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
)

type Server struct {
	producer *nsq.Producer
}

func (s Server) Run() {
	// Register endpoints
	goji.Get("/scripts/*", GetScript)
	goji.Post("/scripts/:name", RunScript)
	goji.Get("/scripts/:name/log/*", GetLog)
	goji.Get("/scripts/:name/log/:id", GetLogByID)

	goji.Serve()
}

func GetScript(w http.ResponseWriter, r *http.Request) {
	// Get a list of all the scripts in script folder.
}

func RunScript(c web.C, w http.ResponseWriter, r *http.Request) {
	// Send details to queue for execution.
}

func GetLog(c web.C, w http.ResponseWriter, r *http.Request) {
	// Retrieve all logs for a specific script.
}

func GetLogByID(c web.C, w http.ResponseWriter, r *http.Request) {
	// Retrieve a specific log of a specific script.
}
