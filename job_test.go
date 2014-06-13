package main

import (
	"io/ioutil"
	"os"
	"testing"
)

var TestJob Job

func TestCleanArgs(t *testing.T) {
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
