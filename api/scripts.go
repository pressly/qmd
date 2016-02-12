package api

import "time"

type ScriptsRequest struct {
	Script      string            `json:"script"`
	Args        []string          `json:"args,omitempty"`
	Files       map[string]string `json:"files,omitempty"`
	CallbackURL string            `json:"callback_url,omitempty"`
}

type ScriptsResponse struct {
	ID string `json:"id"`

	//TODO: These are only for backward-compatibility, we don't need them.
	Script string            `json:"script"`
	Args   []string          `json:"args,omitempty"`
	Files  map[string]string `json:"files,omitempty"`

	CallbackURL string    `json:"callback_url,omitempty"`
	Status      string    `json:"status"`
	StartTime   time.Time `json:"start_time,omitempty"`
	EndTime     time.Time `json:"end_time,omitempty"`
	Duration    string    `json:"duration,omitempty"`
	QmdOut      string    `json:"output,omitempty"`
	ExecLog     string    `json:"exec_log,omitempty"`
	Err         string    `json:"error,omitempty"`
}

type JobScriptsRequest struct {
	ScriptsRequest `json:",inline"`
	ID             string `json:"id"`
	File           string `json:"file"`
}
