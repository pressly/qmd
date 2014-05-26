package common

type JobRequest struct {
	ID      string      `json:"id"`
	Scripts interface{} `json:"scripts"`
}

type JobResponse struct {
	ID      string      `json:"id"`
	Scripts interface{} `json:"scripts"`
}
