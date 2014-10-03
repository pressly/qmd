package qmd

import (
	"net/http"

	"github.com/zenazn/goji/web"
)

func listScriptsHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("listtt...."))
}

func execScriptHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	scriptName := c.URLParams["name"]

	queue.Publish([]byte(scriptName))

	w.Write([]byte(scriptName))
}
