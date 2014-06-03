package main

import (
	"flag"
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/bitly/go-nsq"
)

var (
	configPath = flag.String("config-path", "./qmd.toml", "path to config file")
	config     Config
)

func main() {
	flag.Parse()
	fmt.Printf("Using config file from: %s\n", *configPath)

	if _, err := toml.DecodeFile(*configPath, &config); err != nil {
		fmt.Println(err)
		return
	}

	producer := nsq.NewProducer(config.QueueAddr, nsq.NewConfig())
	server := Server{producer}
	server.Run()

	consumer, err := nsq.NewConsumer(config.Topic, config.Worker.Channel, nsq.NewConfig())
	if err != nil {
		fmt.Println(err)
		return
	}
	worker := Worker{consumer, config.Worker.Throughput}
	worker.Run()
}
