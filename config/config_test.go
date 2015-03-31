package config

import (
	"testing"
)

func TestSampleConfigFile(t *testing.T) {
	_, err := New("this-file-not-exists.conf")
	if err == nil {
		t.Error("expected error")
	}

	_, err = New("../etc/qmd.conf.sample")
	if err != nil {
		t.Error(err)
	}
}
