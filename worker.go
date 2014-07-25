package qmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/bitly/go-nsq"
)

var wMutex = &sync.Mutex{}

type Worker struct {
	Name       string
	Throughput int
	Jobs       map[string]*Job

	queue         *QueueConfig
	scriptDir     string
	workingDir    string
	storeDir      string
	whitelist     map[string]bool
	whitelistPath string
	keepTemp      bool

	jobConsumer     *nsq.Consumer
	commandConsumer *nsq.Consumer

	producer *nsq.Producer
	workChan chan *nsq.Message
}

func NewWorker(wc *WorkerConfig) (*Worker, error) {
	var err error
	var worker Worker

	worker.Name = wc.Name
	worker.Throughput = wc.Throughput
	worker.Jobs = make(map[string]*Job)

	worker.queue = wc.Queue
	worker.scriptDir = wc.ScriptDir
	worker.workingDir = wc.WorkingDir
	worker.storeDir = wc.StoreDir
	worker.keepTemp = wc.KeepTemp

	worker.whitelistPath = wc.Whitelist
	err = worker.loadWhitelist()
	if err != nil {
		return &worker, err
	}

	cfg := nsq.NewConfig()
	cfg.Set("max_in_flight", worker.Throughput)
	jobConsumer, err := nsq.NewConsumer("job", "qmd-worker", cfg)
	if err != nil {
		return &worker, err
	}
	worker.jobConsumer = jobConsumer

	channelName := fmt.Sprintf("%s#ephemeral", worker.Name)
	commandConsumer, err := nsq.NewConsumer("command", channelName, nsq.NewConfig())

	if err != nil {
		return &worker, err
	}
	worker.commandConsumer = commandConsumer

	producer, err := nsq.NewProducer(worker.queue.HostNSQDAddr, nsq.NewConfig())
	if err != nil {
		return &worker, err
	}
	worker.producer = producer

	worker.workChan = make(chan *nsq.Message)

	if err = SetupLogging(wc.Logging); err != nil {
		return &worker, err
	}

	log.Info("Worker created as %s watching %s", worker.Name, worker.whitelistPath)
	return &worker, nil
}

func (w *Worker) Run() error {
	var err error
	// Add and connect the message handlers.
	w.jobConsumer.AddConcurrentHandlers(nsq.HandlerFunc(w.jobHandler), w.Throughput)
	err = ConnectConsumer(w.queue, w.jobConsumer)
	if err != nil {
		w.Exit()
		return err
	}

	w.commandConsumer.AddHandler(nsq.HandlerFunc(w.commandHandler))
	err = ConnectConsumer(w.queue, w.commandConsumer)
	if err != nil {
		w.Exit()
		return err
	}

	go func() {
		for m := range w.workChan {
			go w.process(m)
		}
	}()
	return nil
}

func (w *Worker) Exit() {
	w.jobConsumer.Stop()
	w.commandConsumer.Stop()
	w.producer.Stop()
	close(w.workChan)
}

// Message handlers

func (w *Worker) jobHandler(m *nsq.Message) error {
	m.DisableAutoResponse()
	w.workChan <- m
	return nil
}

func (w *Worker) commandHandler(m *nsq.Message) error {
	var err error
	cmd := strings.Split(string(m.Body), ":")
	switch cmd[0] {
	case "reload":
		log.Info("Received reload request")
		if err = w.loadWhitelist(); err != nil {
			log.Error("Failed to reload whitelist from %s", w.whitelistPath)
			return err
		}
	case "kill":
		log.Info("Received kill request for %s", cmd[1])
		wMutex.Lock()
		job, exists := w.Jobs[cmd[1]]
		wMutex.Unlock()
		runtime.Gosched()
		if exists {
			defer func() {
				wMutex.Lock()
				delete(w.Jobs, cmd[1])
				wMutex.Unlock()
				runtime.Gosched()
			}()
			job.kill()
		}
	}
	return nil
}

// Helper functions

func (w *Worker) process(m *nsq.Message) {
	var err error

	// Tell NSQD to reset time until done channel is non-empty.
	// Runs every 30 seconds.
	done := make(chan int, 1)
	go func() {
		defer close(done)
		for {
			select {
			case <-done:
				return
			case <-time.After(30 * time.Second):
				m.Touch()
			}
		}
	}()

	// Start processing Job
	job, err := NewJob(m.Body)
	if err != nil {
		log.Error("Couldn't create Job: %s", err.Error())
		m.RequeueWithoutBackoff(-1)
		log.Info("Job %s requeued", job.ID)
		done <- 1
	}
	wMutex.Lock()
	w.Jobs[job.ID] = &job
	wMutex.Unlock()
	runtime.Gosched()
	defer func() {
		wMutex.Lock()
		delete(w.Jobs, job.ID)
		wMutex.Unlock()
		runtime.Gosched()
	}()
	log.Info("Dequeued Job %s", job.ID)

	// Try and run script
	if len(w.whitelist) == 0 || w.whitelist[job.Script] {
		if err = job.Execute(w.scriptDir, w.workingDir, w.storeDir, w.keepTemp); err != nil {
			log.Error(err.Error())
		}
	} else {
		job.Status = StatusERR
		msg := fmt.Sprintf("%s is not on script whitelist", job.Script)
		job.ExecLog = msg
	}

	if job.Status != StatusTIMEOUT {
		defer w.respond(&job)
	}
	log.Info(job.ExecLog)
	done <- 1
	m.Finish()
	log.Info("Job %s finished", job.ID)
}

func (w *Worker) respond(j *Job) {
	var err error

	doneChan := make(chan *nsq.ProducerTransaction)
	defer close(doneChan)

	result, err := json.Marshal(j)
	if err != nil {
		log.Error(err.Error())
	}
	err = w.producer.PublishAsync("result", result, doneChan)
	if err != nil {
		log.Error(err.Error())
	}
	<-doneChan
	log.Info("Log for Job #%s sent", j.ID)
}

func (w *Worker) loadWhitelist() error {
	log.Info("Using whitelist from %s", w.whitelistPath)

	var buf bytes.Buffer
	whitelist := make(map[string]bool)
	buf.WriteString(fmt.Sprintf("Whitelist: "))

	fileInfo, err := os.Stat(w.whitelistPath)
	if err != nil {
		return err
	}
	if fileInfo.IsDir() {
		buf.WriteString(fmt.Sprintf("*"))
	} else {
		file, err := os.Open(w.whitelistPath)
		if err != nil {
			return err
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			whitelist[scanner.Text()] = true
		}
		if err != nil {
			log.Error(err.Error())
			return err
		}

		err = scanner.Err()

		for script := range whitelist {
			buf.WriteString(fmt.Sprintf("%s ", script))
		}
	}

	w.whitelist = whitelist
	log.Info(buf.String())
	return nil
}
