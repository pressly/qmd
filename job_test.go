package qmd

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestNewJob(t *testing.T) {
	data := []byte(`{"args": ["10"], "callback_url": "http://192.168.88.88:9090"}`)
	job, err := NewJob(data)
	if err != nil {
		t.Error("Wanted to catch error but passed instead")
	}
	if len(job.Args) != 1 {
		t.Error("Wanted arg length of 1, got %s instead", len(job.Args))
	}
}

func TestNewJobWithBadData(t *testing.T) {
	badData := []byte(`{"args: ["10"], "callback_url": "http://192.168.88.88:9090"}`)
	_, err := NewJob(badData)
	if err == nil {
		t.Error("Wanted to catch error but passed instead")
	}
}

func TestCleanArgs(t *testing.T) {
	var TestJob Job
	dirtyArgs := []string{"a"}
	expected := []string{"a"}

	TestJob.Args = dirtyArgs
	result, err := TestJob.CleanArgs()
	if err != nil {
		t.Error(err)
	}
	for i := 0; i < len(expected); i++ {
		if result[i] != expected[i] {
			t.Errorf("Wanted %s, got %s instead", expected, result)
		}
	}
}

func TestSaveFiles(t *testing.T) {
	var TestJob Job
	dir := "."
	dirtyName := "../../test.test.test/../.."
	data := "sudo rm -rf *"
	TestJob.Files = make(map[string]string, 1)
	TestJob.Files[dirtyName] = data
	expectedName := "./test.test.test"

	// Write file
	err := TestJob.SaveFiles(dir)
	if err != nil {
		t.Error(err)
	}

	// Check file was written
	if _, err := os.Stat(expectedName); os.IsNotExist(err) {
		t.Error(err)
	}

	// Check file's contents
	innards, err := ioutil.ReadFile(expectedName)
	if err != nil {
		t.Error(err)
	}
	if string(innards) != data {
		t.Errorf("Wanted %s, got %s instead", data, string(innards))
	}

	// Clean up
	err = os.Remove(expectedName)
	if err != nil {
		t.Error(err)
	}
}
