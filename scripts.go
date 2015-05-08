package qmd

import (
	"errors"
	"os"
	"path"
	"path/filepath"
	"sync"
)

type Scripts struct {
	Running bool

	sync.Mutex                   // guards files
	files      map[string]string // Map of scripts names to the actual files.
}

// Update walks ScriptDir directory for shell scripts
// and updates the files cache.
func (s *Scripts) Update(dir string) error {
	info, err := os.Stat(dir)
	if err != nil {
		return errors.New("script_dir=\"" + dir + "\": no such directory")
	}
	if !info.IsDir() {
		return errors.New("script_dir=\"" + dir + "\": not a directory")
	}

	files := map[string]string{}
	if err := filepath.Walk(dir, func(file string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			if path.Ext(file) == ".sh" {
				rel, err := filepath.Rel(dir, file)
				if err != nil {
					return err
				}
				files[rel] = file
			}
		}
		return nil
	}); err != nil {
		return err
	}

	s.Lock()
	defer s.Unlock()

	s.files = files
	return nil
}

func (s *Scripts) Get(file string) (string, error) {
	s.Lock()
	defer s.Unlock()

	script, ok := s.files[file]
	if !ok {
		return "", errors.New("script doesn't exist")
	}
	return script, nil
}
