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
	Consumer        *nsq.Consumer
	ReloadConsumer  *nsq.Consumer
	Throughput      int
	QueueAddresses  []string
	LookupAddresses []string
	WhiteList       map[string]bool
}

func NewWorker(c Config) (worker Worker, err error) {
	worker.Throughput = c.Worker.Throughput
	worker.QueueAddresses = c.Worker.QueueAddresses
	worker.LookupAddresses = c.Worker.LookupAddresses

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

	log.Info("Worker created watching scripts in %s", c.Worker.ScriptDir)
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
	log.Info("Dequeued Job %d", job.ID)

	// Check if Job is already being executed
	success, err := setRedisID(job.ID)
	if err != nil {
		log.Error("Couldn't continue with Job #%d, aborting: %s", job.ID, err.Error())
		if !success {
			unsetRedisID(job.ID)
		}
		return err
	}
	if !success {
		log.Info("Job #%d being handled, aborting!", job.ID)
		return err
	}

	// Try and run script
	if len(w.WhiteList) == 0 || w.WhiteList[job.Script] {
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

	var err error

	// Connect consumers to NSQLookupd
	if w.LookupAddresses != nil && len(w.LookupAddresses) != 0 {
		log.Info("Connecting Consumer to the following NSQLookupds %s", w.LookupAddresses)
		err = w.Consumer.ConnectToNSQLookupds(w.LookupAddresses)
		if err != nil {
			log.Error(err.Error())
		}

		log.Info("Connecting ReloadConsumer to the following NSQLookupds %s", w.LookupAddresses)
		err = w.ReloadConsumer.ConnectToNSQLookupds(w.LookupAddresses)
		if err != nil {
			log.Error(err.Error())
		}
	} else {
		// Connect consumers to NSQD
		if w.QueueAddresses != nil && len(w.QueueAddresses) != 0 {
			log.Info("Connecting Consumer to the following NSQDs %s", w.QueueAddresses)
			err = w.Consumer.ConnectToNSQDs(w.QueueAddresses)
			if err != nil {
				log.Error(err.Error())
			}

			log.Info("Connecting ReloadConsumer to the following NSQDs %s", w.QueueAddresses)
			err = w.ReloadConsumer.ConnectToNSQDs(w.QueueAddresses)
			if err != nil {
				log.Error(err.Error())
			}
		} else {
			log.Fatal("Couldn't connect to any NSQLookupd: %s or NSQD: %s nodes", w.LookupAddresses, w.QueueAddresses)
		}
	}
}

func (w *Worker) Stop() {
	w.Consumer.Stop()
	w.ReloadConsumer.Stop()
	return
}
