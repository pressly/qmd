package common

type JobRequest struct {
	ID      string   `json:"id"`
	Scripts []Script `json:"scripts"`
}

type JobResponse struct {
	ID      string   `json:"id"`
	Scripts []Script `json:"scripts"`
}

type Script struct {
	Name   string                   `json:"name"`
	Params []map[string]interface{} `json:"params"`
}
