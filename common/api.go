package common

type Job struct {
	ID      string   `json:"id"`
	Scripts []Script `json:"scripts"`
}

type Script struct {
	Name   string   `json:"name"`
	Params []string `json:"params"`
}
