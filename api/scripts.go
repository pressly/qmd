package api

import "time"

type ScriptsRequest struct {
	Args        []string          `json:"args,omitempty"`
	Files       map[string]string `json:"files,omitempty"`
	CallbackURL string            `json:"callback_url,omitempty"`
}

type ScriptsResponse struct {
	ID          string        `json:"id"`
	Script      string        `json:"script"`
	Args        []string      `json:"args"`
	CallbackURL string        `json:"callback_url"`
	Status      string        `json:"status"`
	StartTime   time.Time     `json:"start_time"`
	EndTime     time.Time     `json:"end_time"`
	Duration    time.Duration `json:"duration"`
	Output      string        `json:"output"`
	ExecLog     string        `json:"exec_log"`
}
