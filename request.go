package qmd

import (
	"encoding/json"
	"time"
)

type Request struct {
	ID          string            `json:"id"`
	Script      string            `json:"script"`
	Args        []string          `json:"args,omitempty"`
	Files       map[string]string `json:"files,omitempty"`
	CallbackURL string            `json:"callback_url"`
	Status      string            `json:"status"`
	StartTime   time.Time         `json:"start_time"`
	FinishTime  time.Time         `json:"end_time"`
	Duration    string            `json:"duration"`
}

func (r Request) WriteJSON() ([]byte, error) {
	return json.Marshal(r)
}

func (r Request) WritePrettyJSON() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}
