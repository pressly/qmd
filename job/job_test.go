package job

import (
	"bytes"
	"os/exec"
	"testing"
)

func TestRunJob(t *testing.T) {
	// Bash command that prints "result" on STDOUT and "error" on STDERR.
	cmd := exec.Command("bash", "-c", "echo -n result; >&2 echo -n error")

	job, err := New(cmd)
	if err != nil {
		t.Error(err)
	}
	if job.Running {
		t.Error("unexpected value")
	}
	if job.Duration != 0 {
		t.Error("unexpected value")
	}

	// Copy job's STDOUT and STDERR to a buffer.
	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)
	go stdout.ReadFrom(job.Stdout)
	go stderr.ReadFrom(job.Stderr)

	err = job.Run()
	if err != nil {
		t.Error(err)
	}
	if !job.Running {
		t.Error("unexpected value")
	}

	err = job.Wait()
	if err != nil {
		t.Error(err)
	}
	if job.Running {
		t.Error("unexpected value")
	}
	if job.Duration == 0 {
		t.Error("unexpected value")
	}

	// Test the job's STDOUT and STDERR.
	if got := stdout.String(); got != "result" {
		t.Errorf("unexpected stdout \"%s\"", got)
	}
	if got := stderr.String(); got != "error" {
		t.Errorf("unexpected stderr \"%s\"", got)
	}
}
