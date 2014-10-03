package qmd

import (
	"net/http"

	"github.com/pressly/gohttpware/heartbeat"
	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/middleware"
)

func NewHTTPHandler() http.Handler {
	m := web.New()

	// Middleware
	m.Use(middleware.RequestID)
	m.Use(middleware.RealIP)
	m.Use(middleware.Logger)
	m.Use(middleware.Recoverer)
	// m.Use(middleware.AutomaticOptions)
	m.Use(heartbeat.Route("/ping"))

	// TODO: the auth handler ... + bruteforce thing.......

	// Routes
	m.Get("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(""))
	})
	m.Get("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Disallow: /\n")) // disallow all robots
	})
	m.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("."))
	})

	m.Get("/scripts", listScriptsHandler)
	m.Post("/scripts/:name", execScriptHandler)
	m.Get("/scripts/:name", execScriptHandler)

	return m
}

// TODO: would be cool to have a /console
// websocket thing where you can see the stdout of everything..
// maybe even do /console/<script>
// + filter by OK, ERR, etc... with a counter in some top area..
