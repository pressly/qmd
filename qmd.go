package qmd

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/goware/disque"
	"github.com/op/go-logging"

	"github.com/pressly/qmd/config"
)

var lg = logging.MustGetLogger("qmd")

type Qmd struct {
	Config  *config.Config
	DB      *DB
	Queue   *disque.Pool
	Scripts Scripts
	Workers chan Worker
	Logger  *logging.Logger

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
		Logger:             lg,
	}

	if err := qmd.Scripts.Update(qmd.Config.ScriptDir); err != nil {
		return nil, err
	}

	backends := []logging.Backend{
		logging.NewLogBackend(os.Stdout, "", log.LstdFlags),
	}

	if conf.Slack.Enabled {
		backends = append(backends, &SlackNotifier{
			WebhookURL: conf.Slack.WebhookURL,
			Channel:    conf.Slack.Channel,
			Prefix:     fmt.Sprintf("%v: ", conf.URL),
		})
	}

	logging.SetBackend(backends...)

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

	lg.Fatal("Closed")
}

func (qmd *Qmd) GetScript(file string) (string, error) {
	return qmd.Scripts.Get(file)
}

func (qmd *Qmd) WatchScripts() {
	for {
		err := qmd.Scripts.Update(qmd.Config.ScriptDir)
		if err != nil {
			lg.Error(err.Error())
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
