package script

import (
	"net/http"

	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/middleware"
)

func Router() http.Handler {
	r := web.New()

	r.Use(middleware.SubRouter)

	r.Post("/:filename", CreateJob)

	return r
}
