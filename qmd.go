package qmd

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/goware/disque"
	"github.com/goware/lg"

	"github.com/pressly/qmd/config"
)

type Qmd struct {
	Config  *config.Config
	DB      *DB
	Queue   *disque.Pool
	Scripts Scripts
	Workers chan Worker
	Slack   *SlackNotifier

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

	// TODO: SlackNOP.
	slack := &SlackNotifier{
		WebhookURL: conf.Slack.WebhookURL,
		Channel:    conf.Slack.Channel,
		Prefix:     fmt.Sprintf("%v: ", conf.URL),
	}

	qmd := &Qmd{
		Config:             conf,
		DB:                 db,
		Queue:              queue,
		Workers:            make(chan Worker, conf.MaxJobs),
		ClosingListenQueue: make(chan struct{}),
		ClosingWorkers:     make(chan struct{}),
		Slack:              slack,
	}

	if err := lg.SetLevelString("debug"); err != nil {
		return nil, err
	}

	lg.AlertFn = func(level lg.Level, msg string) {
		switch level {
		case lg.ErrorLevel, lg.FatalLevel, lg.PanicLevel:
			qmd.Slack.Notify(msg)
		}
	}

	return qmd, nil
}

func (qmd *Qmd) Close() {
	lg.Debug("Closing")

	qmd.Closing = true

	close(qmd.ClosingListenQueue)
	qmd.WaitListenQueue.Wait()

	close(qmd.ClosingWorkers)
	qmd.WaitWorkers.Wait()

	qmd.DB.Close()
	qmd.Queue.Close()
}

func (qmd *Qmd) GetScript(file string) (string, error) {
	return qmd.Scripts.Get(file)
}

// TODO: Use fsnotify.
func (qmd *Qmd) WatchScripts() {
	for {
		if err := qmd.Scripts.Update(qmd.Config.ScriptDir); err != nil {
			lg.Error(err)
			continue
		}
		time.Sleep(10 * time.Second)
	}
}

func (qmd *Qmd) ClosingResponder(h http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if qmd.Closing {
			http.Error(w, http.StatusText(503), 503)
			return
		}
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(handler)
}
