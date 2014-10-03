package qmd

type Job struct {
	Msg     []byte // low-level message...? body...?
	Output  string
	ExecLog string
}

func NewJob() (*Job, error) {
	j := &Job{}
	return j, nil
}

func (j *Job) Exec() error {
	return nil
}
