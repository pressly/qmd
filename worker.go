package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path"

	"github.com/bitly/go-nsq"
)

const (
	STATUS_OK  string = "OK"
	STATUS_ERR string = "ERR"
)

type Worker struct {
	Consumer       *nsq.Consumer
	ReloadConsumer *nsq.Consumer
	Throughput     int
	QueueAddr      string
	WhiteList      map[string]bool
}

func NewWorker(c Config) (worker Worker, err error) {
	worker.Throughput = c.Worker.Throughput
	worker.QueueAddr = c.QueueAddr

	conf := nsq.NewConfig()
	conf.Set("max_in_flight", worker.Throughput)
	consumer, err := nsq.NewConsumer(c.Topic, c.Worker.Channel, conf)
	if err != nil {
		log.Error(err.Error())
		return worker, err
	}

	rConsumer, err := nsq.NewConsumer("reload", c.Worker.Channel, nsq.NewConfig())
	if err != nil {
		log.Error(err.Error())
		return worker, err
	}

	worker.Consumer = consumer
	worker.ReloadConsumer = rConsumer

	// Generate whitelist of allowed scripts.
	path := path.Join(config.Worker.ScriptDir, config.Worker.WhiteList)
	err = worker.LoadWhiteList(path)
	if err != nil {
		log.Error(err.Error())
		return worker, err
	}

	log.Info("Worker connecting to %s and running scripts in %s.\n", c.QueueAddr, c.Worker.WorkingDir)
	return worker, nil
}

func (w *Worker) LoadWhiteList(path string) error {
	log.Debug("Loading scripts in %s", path)

	var buf bytes.Buffer
	whiteList := make(map[string]bool)
	buf.WriteString(fmt.Sprintf("Whitelist: "))

	fileInfo, err := os.Stat(path)
	if err != nil {
		return err
	}
	if fileInfo.IsDir() {
		buf.WriteString(fmt.Sprintf("*"))
	} else {
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			whiteList[scanner.Text()] = true
		}
		if err != nil {
			log.Error(err.Error())
			return err
		}

		err = scanner.Err()

		for script := range whiteList {
			buf.WriteString(fmt.Sprintf("%s ", script))
		}
	}

	w.WhiteList = whiteList
	log.Debug(buf.String())
	return nil
}

func (w *Worker) JobRequestHandler(m *nsq.Message) error {
	job, err := NewJob(m.Body)
	if err != nil {
		log.Error("Couldn't create Job: %s", err.Error())
		return err
	}

	// Try and run script
	if len(w.WhiteList) == 0 || w.WhiteList[job.Script] {
		log.Info("Dequeued request as Job %d", job.ID)

		resultChan := make(chan error, 1)
		go job.Execute(resultChan)
		err := <-resultChan
		if err != nil {
			job.Status = STATUS_ERR
		} else {
			job.Status = STATUS_OK
		}
	} else {
		msg := fmt.Sprintf("%s is not on script whitelist", job.Script)
		job.ExecLog = msg
	}

	log.Info(job.ExecLog)
	job.Log()
	job.Callback()
	return nil
}

func (w *Worker) ReloadRequestHandler(m *nsq.Message) error {
	whitelistPath := path.Clean(string(m.Body))
	err := w.LoadWhiteList(whitelistPath)
	if err != nil {
		log.Error("Failed to reload whitelist from %s", whitelistPath)
		return err
	}
	log.Info("Reloaded whitelist from %s", whitelistPath)
	return nil
}

func (w *Worker) Run() {
	// Set the message handler.
	w.Consumer.SetConcurrentHandlers(nsq.HandlerFunc(w.JobRequestHandler), w.Throughput)
	w.ReloadConsumer.SetHandler(nsq.HandlerFunc(w.ReloadRequestHandler))

	// Connect the queue.
	err := w.Consumer.ConnectToNSQD(w.QueueAddr)
	if err != nil {
		log.Error(err.Error())
	}

	err = w.ReloadConsumer.ConnectToNSQD(w.QueueAddr)
	if err != nil {
		log.Error(err.Error())
	}
}

func (w *Worker) Stop() {
	w.Consumer.Stop()
	w.ReloadConsumer.Stop()
	return
}
