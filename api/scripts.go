package api

import "time"

type ScriptsRequest struct {
	Args        []string          `json:"args,omitempty"`
	Files       map[string]string `json:"files,omitempty"`
	CallbackURL string            `json:"callback_url"`
}

type ScriptsResponse struct {
	Script    string     `json:"script"`
	StartTime *time.Time `json:"start_time"`
	EndTime   *time.Time `json:"end_time"`
	Duration  string     `json:"duration"`
	Status    string     `json:"status"`
	Output    string     `json:"output"`
	ExecLog   string     `json:"exec_log"`
}
