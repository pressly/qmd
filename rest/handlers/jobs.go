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
	lowActive, _ := Qmd.Queue.ActiveLen("low")
	highActive, _ := Qmd.Queue.ActiveLen("high")
	urgentActive, _ := Qmd.Queue.ActiveLen("urgent")
	cached, _ := Qmd.DB.Len()
	finished, _ := Qmd.DB.TotalLen()

	r.Header.Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "Queued: %v\n- %v (urgent)\n- %v (high)\n- %v (low)\n\n", urgent+high+low, urgent, high, low)
	fmt.Fprintf(w, "Running: %v\n- %v (urgent)\n- %v (high)\n- %v (low)\n\n", urgentActive+highActive+lowActive, urgentActive, highActive, lowActive)
	fmt.Fprintf(w, "Finished (in-cache): %v\n\n", cached)
	fmt.Fprintf(w, "Finished (total): %v", finished)
}
