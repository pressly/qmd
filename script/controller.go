package script

import (
	"errors"
	"log"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/pressly/qmd/config"
)

var Ctl *Controller

// Controller watches scripts in ScriptDir.
type Controller struct {
	ScriptDir string

	*sync.Mutex                   // guards scripts
	scripts     map[string]string // Map of relative script paths to the actual files.
}

// NewController creates new instance of Script Controller.
func NewController(conf *config.Config) (*Controller, error) {
	info, err := os.Stat(conf.ScriptDir)
	if err != nil {
		return nil, errors.New("script_dir=\"" + conf.ScriptDir + "\": " + err.Error())
	}
	if !info.IsDir() {
		return nil, errors.New("script_dir=\"" + conf.ScriptDir + "\": not a directory")
	}

	Ctl = &Controller{
		ScriptDir: conf.ScriptDir,
		Mutex:     &sync.Mutex{},
	}

	return Ctl, nil
}

// Run runs the Controller loop.
func (c *Controller) Run() {
	for {
		err := c.FindScripts()
		if err != nil {
			log.Print(err)
		}
		time.Sleep(10 * time.Second)
	}
}

// FindScripts walks the ScriptDir and finds all the shell scripts.
func (c *Controller) FindScripts() error {
	log.Println("ScriptController: Walking script_dir..")

	scripts := map[string]string{}
	if err := filepath.Walk(c.ScriptDir, func(file string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			if path.Ext(file) == ".sh" {
				rel, err := filepath.Rel(c.ScriptDir, file)
				if err != nil {
					return err
				}
				scripts[rel] = file
				log.Printf("ScriptController: Found script \"%v\"\n", rel)
			}
		}
		return nil
	}); err != nil {
		return err
	}

	c.Lock()
	defer c.Unlock()

	c.scripts = scripts
	return nil
}

func (c *Controller) Get(file string) (string, error) {
	c.Lock()
	defer c.Unlock()

	script, ok := c.scripts[file]
	if !ok {
		return "", errors.New("script doesn't exist")
	}
	return script, nil
}