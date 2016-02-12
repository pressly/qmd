package handlers

import (
	"net/http"

	"github.com/pressly/qmd"
	"golang.org/x/net/context"
)

var Qmd *qmd.Qmd

func Index(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`¯\_(ツ)_/¯`))
}
