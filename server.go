package qmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/bitly/go-nsq"
)

var sMutex = &sync.Mutex{}

type Server struct {
	Name        string
	TTL         time.Duration
	RequestChan chan Request
	ResultChan  chan []byte

	Requests map[string]chan []byte

	producer *nsq.Producer
	consumer *nsq.Consumer
	DB       *DB

	queue *QueueConfig
}

func NewServer(sc *ServerConfig) (*Server, error) {
	var server Server
	var err error

	server.Name = sc.Name
	server.TTL = sc.TTL
	server.queue = sc.Queue
	server.Requests = make(map[string]chan []byte)

	producer, err := nsq.NewProducer(server.queue.HostNSQDAddr, nsq.NewConfig())
	if err != nil {
		return &server, err
	}
	server.producer = producer

	channelName := fmt.Sprintf("%s#ephemeral", server.Name)
	consumer, err := nsq.NewConsumer("result", channelName, nsq.NewConfig())
	if err != nil {
		return &server, err
	}
	server.consumer = consumer

	server.RequestChan = make(chan Request)
	server.ResultChan = make(chan []byte)
	server.DB = NewDB(sc.DBAddr)

	if err = SetupLogging(sc.Logging); err != nil {
		return &server, err
	}

	log.Info("Server created as %s", server.Name)
	return &server, nil
}

func (s Server) Reload() error {
	doneChan := make(chan *nsq.ProducerTransaction)
	err := s.producer.PublishAsync("command", []byte("reload"), doneChan)
	<-doneChan
	return err
}

func (s Server) Queue(script string, data []byte) (<-chan []byte, error) {
	var err error

	req := Request{
		ID:        NewID(),
		Script:    script,
		Status:    StatusQUEUED,
		StartTime: time.Now(),
	}
	if err = json.Unmarshal(data, &req); err != nil {
		return nil, err
	}

	ch := make(chan []byte, 1)
	sMutex.Lock()
	s.Requests[req.ID] = ch
	sMutex.Unlock()
	runtime.Gosched()
	s.RequestChan <- req

	// Asynchronous call
	if req.CallbackURL != "" {
		newCh := make(chan []byte, 1)
		defer close(newCh)
		reply := <-ch
		newCh <- reply
		go callback(req.CallbackURL, ch)
		return newCh, nil
	}
	return ch, nil
}

func (s Server) Run() error {
	var err error

	s.consumer.AddHandler(nsq.HandlerFunc(s.resultHandler))
	if err = ConnectConsumer(s.queue, s.consumer); err != nil {
		s.Exit()
		return err
	}

	go func() {
		for req := range s.RequestChan {
			go s.processRequest(req)
		}
	}()

	go func() {
		for res := range s.ResultChan {
			go s.processResult(res)
		}
	}()

	return nil
}

func (s Server) Exit() {
	s.producer.Stop()
	s.consumer.Stop()
	defer close(s.RequestChan)
	defer close(s.ResultChan)
	for _, ch := range s.Requests {
		go close(ch)
	}
}

// Message handler(s)

func (s Server) resultHandler(m *nsq.Message) error {
	s.ResultChan <- m.Body
	return nil
}

// Helper functions

func (s Server) processRequest(r Request) {
	var err error

	data, err := r.WriteJSON()
	if err != nil {
		log.Error(err.Error())
		return
	}

	doneChan := make(chan *nsq.ProducerTransaction)
	if err = s.producer.PublishAsync("job", data, doneChan); err != nil {
		log.Error(err.Error())
		return
	}
	<-doneChan
	log.Info("Request queued as %s", r.ID)
	go s.startTTL(r)

	sMutex.Lock()
	ch, exist := s.Requests[r.ID]
	sMutex.Unlock()
	runtime.Gosched()
	if exist && r.CallbackURL != "" {
		ch <- data
	}
}

func (s Server) processResult(data []byte) {
	var r Job
	if err := json.Unmarshal(data, &r); err != nil {
		log.Error(err.Error())
		return
	}
	s.finish(r.Script, r.ID, data)
}

func (s Server) startTTL(req Request) {
	req.FinishTime = <-time.After(s.TTL)
	sMutex.Lock()
	_, exists := s.Requests[req.ID]
	sMutex.Unlock()
	runtime.Gosched()
	if !exists {
		return
	}
	req.Duration = fmt.Sprintf("%f", req.FinishTime.Sub(req.StartTime).Seconds())
	req.Status = StatusTIMEOUT

	data, err := json.Marshal(&req)
	if err != nil {
		log.Error(err.Error())
	}
	go s.killJob(req.ID)
	doneChan := make(chan *nsq.ProducerTransaction)
	if err := s.producer.PublishAsync("result", data, doneChan); err != nil {
		log.Error(err.Error())
	}
	<-doneChan
}

func (s Server) finish(script, id string, data []byte) {
	sMutex.Lock()
	ch, exist := s.Requests[id]
	sMutex.Unlock()
	runtime.Gosched()
	if exist {
		defer func() {
			sMutex.Lock()
			delete(s.Requests, id)
			sMutex.Unlock()
			runtime.Gosched()
		}()
		defer close(ch)
		ch <- data
		s.DB.SetLog(script, id, string(data))
	}
}

func (s Server) killJob(id string) {
	doneChan := make(chan *nsq.ProducerTransaction)
	cmd := fmt.Sprintf("kill:%s", id)
	if err := s.producer.PublishAsync("command", []byte(cmd), doneChan); err != nil {
		log.Error(err.Error())
	}
	<-doneChan
	log.Info("Sent kill request for %s", id)
}

func callback(url string, ch chan []byte) {
	buf := bytes.NewBuffer(<-ch)
	_, err := http.Post(url, "application/json", buf)
	if err != nil {
		log.Error(err.Error())
	}
}
