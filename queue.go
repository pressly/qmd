package qmd

import (
	"log"
	"os"

	"github.com/bitly/go-nsq"
	"github.com/bitly/nsq/nsqd"
)

type QueueConf struct {
	Bind       string `toml:"bind"`
	DataPath   string `toml:"data_path"`
	Throughput int    `toml:"throughput"`
	KeepWork   bool   `toml:"keep_work"`
}

func NewQueueConf() *QueueConf {
	cf := &QueueConf{}
	cf.Setup()
	return cf
}

func (cf *QueueConf) Setup() {
	// TODO: set defaults..
	// cf.Bind = "0.0.0.0:5150"
	// cf.DataPath = "."
	if cf.Throughput == 0 {
		cf.Throughput = 10
	}
}

//--

const (
	qTopic   = "qmd"
	qChannel = "hrmmmm" // what do we need a channel here for..?
)

type Queue struct {
	conf     *Conf
	nsqd     *nsqd.NSQD
	producer *nsq.Producer
	consumer *nsq.Consumer
}

func NewQueue(cf *QueueConf) (*Queue, error) {
	q := &Queue{}

	log.Println("config:", cf)

	opts := nsqd.NewNSQDOptions()

	opts.TCPAddress = cf.Bind
	opts.HTTPAddress = "" // or "0.0.0.0:5151" ?
	opts.HTTPSAddress = ""
	opts.StatsdPrefix = "" // "qmd.%s"
	opts.DataPath = cf.DataPath
	opts.Logger = log.New(os.Stdout, "[nsqd] ", 0) // hrmm.. sup..?

	q.nsqd = nsqd.NewNSQD(opts)
	q.nsqd.LoadMetadata()
	err := q.nsqd.PersistMetadata()
	if err != nil {
		return nil, err
	}

	//--

	// hrmm.. put this in another function...?

	// Boot server
	// TODO: .. we need to run this with a channel and listen on signal to exit.. etc.
	go q.nsqd.Main()

	// Connect the producer
	producer, err := nsq.NewProducer(cf.Bind, nsq.NewConfig())
	if err != nil {
		// stop nsqd....?
		return nil, err
	}
	q.producer = producer

	// Connect the consumer
	cfg := nsq.NewConfig()
	cfg.Set("max_in_flight", cf.Throughput)
	consumer, err := nsq.NewConsumer(qTopic, qChannel, cfg)
	if err != nil {
		return nil, err
	}
	q.consumer = consumer

	// Consumer handler
	q.consumer.AddConcurrentHandlers(nsq.HandlerFunc(q.MsgHandler), cf.Throughput)
	// q.consumer.ConnectToNSQDs(addresses) // connect the consumer to all servers.. what happens if one disconnects..does it reconnect/recover?
	q.consumer.ConnectToNSQD("0.0.0.0:5150") // for now..

	// time.Sleep(1e9)

	// Spin up the worker listener........
	// q.Publish([]byte("suppp"))
	// q.Publish([]byte("okay thanks"))

	return q, nil
}

func (q *Queue) Publish(msg []byte) error {
	err := q.producer.Publish("qmd", msg)
	if err != nil {
		return err
	}
	return nil
}

func (q *Queue) MsgHandler(m *nsq.Message) error {
	// TODO: .. this should be sent to a channel of up to X go-routines...
	log.Println("Processing job...............:", string(m.Body))
	return nil
}
