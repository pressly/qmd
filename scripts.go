package qmd

import (
	"errors"
	"log"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"
)

type Scripts struct {
	Running bool

	sync.Mutex                   // guards files
	files      map[string]string // Map of scripts names to the actual files.
}

// Watch watches a specific directory for files and updates the cache.
func (s *Scripts) Watch(dir string) {
	for {
		info, err := os.Stat(dir)
		if err != nil {
			log.Printf("script_dir=\"" + dir + "\": no such directory")
			time.Sleep(1 * time.Second)
			continue
		}
		if !info.IsDir() {
			log.Printf("script_dir=\"" + dir + "\": not a directory")
			time.Sleep(1 * time.Second)
			continue
		}

		err = s.Walk(dir)
		if err != nil {
			log.Print(err)
			time.Sleep(1 * time.Second)
			continue
		}

		time.Sleep(10 * time.Second)
	}
}

// Walk walks the ScriptDir and finds all the shell scripts.
func (s *Scripts) Walk(dir string) error {
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
