package api

import "time"

type ScriptsRequest struct {
	Args        []string          `json:"args,omitempty"`
	Files       map[string]string `json:"files,omitempty"`
	CallbackURL string            `json:"callback_url,omitempty"`
}

type ScriptsResponse struct {
	ID string `json:"id"`
	//TODO: We probably don't need those in response:
	// Script      string            `json:"script"`
	// Args        []string          `json:"args,omitempty"`
	// Files       map[string]string `json:"files,omitempty"`
	CallbackURL string    `json:"callback_url,omitempty"`
	Status      string    `json:"status"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Duration    string    `json:"duration"`
	QmdOut      string    `json:"output"`
	ExecLog     string    `json:"exec_log"`
}

type JobScriptsRequest struct {
	ScriptsRequest `json:",inline"`
	ID             string `json:"id"`
	File           string `json:"file"`
}
