package handlers

import (
	"fmt"
	"net/http"

	"golang.org/x/net/context"
)

func Job(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	id, _ := ctx.Value("id").(string)

	resp, err := Qmd.GetResponse(id)
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}
	w.Write(resp)
}

func Jobs(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	low, _ := Qmd.Queue.Len("low")
	high, _ := Qmd.Queue.Len("high")
	urgent, _ := Qmd.Queue.Len("urgent")
	cached, _ := Qmd.DB.Len()
	finished, _ := Qmd.DB.TotalLen()

	r.Header.Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "Enqueued: %v\n- %v (urgent)\n- %v (high)\n- %v (low)\n\n", urgent+high+low, urgent, high, low)
	fmt.Fprintf(w, "Running: TODO(https://github.com/antirez/disque/issues/48)\n\n")
	fmt.Fprintf(w, "In-cache: %v\n\n", cached)
	fmt.Fprintf(w, "Finished: %v", finished)
}
