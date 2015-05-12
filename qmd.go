package qmd

import (
	"log"
	"net/http"
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

	Closing            bool
	ClosingListenQueue chan struct{}
	WaitListenQueue    sync.WaitGroup
	ClosingWorkers     chan struct{}
	WaitWorkers        sync.WaitGroup
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
		Config:             conf,
		DB:                 db,
		Queue:              queue,
		Workers:            make(chan Worker, conf.MaxJobs),
		ClosingListenQueue: make(chan struct{}),
		ClosingWorkers:     make(chan struct{}),
	}

	if err := qmd.Scripts.Update(qmd.Config.ScriptDir); err != nil {
		return nil, err
	}

	return qmd, nil
}

func (qmd *Qmd) Close() {
	log.Printf("Closing")

	qmd.Closing = true

	close(qmd.ClosingListenQueue)
	qmd.WaitListenQueue.Wait()

	close(qmd.ClosingWorkers)
	qmd.WaitWorkers.Wait()

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

func (qmd *Qmd) ClosingResponder(h http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if qmd.Closing {
			http.Error(w, "Temporary unavailable", 503)
			return
		}
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(handler)
}
