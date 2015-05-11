package qmd_test

import (
	"log"
	"os/exec"
	"testing"

	"github.com/pressly/qmd"
	"github.com/pressly/qmd/config"
)

func TestStartWaitJob(t *testing.T) {
	run := exec.Command("bash", "-c", "echo -n stdout; echo -n stderr >&2; echo -n result >$QMD_OUT")

	conf, err := config.New("./etc/qmd.conf.sample")
	if err != nil {
		log.Fatal(err)
	}

	Qmd := &qmd.Qmd{
		Config:             conf,
		DB:                 nil,
		Queue:              nil,
		ClosingListenQueue: make(chan struct{}),
		ClosingWorkers:     make(chan struct{}),
	}

	cmd, err := Qmd.Cmd(run)
	if err != nil {
		t.Error(err)
	}
	if cmd.State != qmd.Initialized {
		t.Error("unexpected value")
	}
	if cmd.Duration != 0 {
		t.Error("unexpected value")
	}

	err = cmd.Start()
	if err != nil {
		t.Error(err)
	}
	if cmd.State != qmd.Running {
		t.Error("unexpected value")
	}

	err = cmd.Wait()
	if err != nil {
		t.Error(err)
	}
	if cmd.State != qmd.Finished {
		t.Error("unexpected value")
	}
	if cmd.Duration == 0 {
		t.Error("unexpected value")
	}

	// Test the cmd's STDOUT and STDERR.
	if e := "stdoutstderr"; cmd.CmdOut.String() != e {
		t.Errorf(`expected "%s", got "%s"`, e, cmd.CmdOut.String())
	}
	if e := "result"; cmd.QmdOut.String() != e {
		t.Errorf(`expected "%s", got "%s"`, e, cmd.CmdOut.String())
	}
}
