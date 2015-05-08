package qmd

import (
	"log"
	"sync"
	"time"

	"github.com/pressly/qmd/config"

	"github.com/goware/disque"
)

type Qmd struct {
	Config  *config.Config
	DB      *DB
	Queue   *disque.Pool
	Scripts Scripts
	Workers chan Worker

	Closing        chan struct{}
	ClosingWorkers chan struct{}
	Wait           sync.WaitGroup
}

func New(conf *config.Config) (*Qmd, error) {
	db, err := NewDB(conf.DB.RedisURI)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	queue, err := disque.New(conf.Queue.DisqueURI)
	if err != nil {
		return nil, err
	}
	queue.Use(disque.Config{
		RetryAfter: time.Duration(conf.MaxExecTime) * time.Second,
		Timeout:    time.Second,
	})

	if err := queue.Ping(); err != nil {
		return nil, err
	}

	qmd := &Qmd{
		Config:         conf,
		DB:             db,
		Queue:          queue,
		Closing:        make(chan struct{}),
		ClosingWorkers: make(chan struct{}),
	}

	if err := qmd.Scripts.Update(qmd.Config.ScriptDir); err != nil {
		return nil, err
	}

	return qmd, nil
}

func (qmd *Qmd) Close() {
	log.Printf("Closing")

	close(qmd.Closing)
	qmd.Wait.Wait()
	qmd.DB.Close()
	qmd.Queue.Close()

	log.Fatalf("Closed")
}

func (qmd *Qmd) GetScript(file string) (string, error) {
	return qmd.Scripts.Get(file)
}

func (qmd *Qmd) WatchScripts() {
	for {
		err := qmd.Scripts.Update(qmd.Config.ScriptDir)
		if err != nil {
			log.Print(err)
			time.Sleep(1 * time.Second)
			continue
		}
		time.Sleep(10 * time.Second)
	}
}
