package script

import (
	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/middleware"
)

func Router() *web.Mux {
	r := web.New()

	r.Use(middleware.SubRouter)

	r.Post("/:filename", CreateJob)

	return r
}
