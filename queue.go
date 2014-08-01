package qmd

import (
	"fmt"

	"github.com/bitly/go-nsq"
)

type QueueConfig struct {
	HostNSQDAddr string   `toml:"host_nsqd_address"`
	NSQDAddrs    []string `toml:"nsqd_addresses"`
	LookupdAddrs []string `toml:"lookupd_addresses"`
}

func (qc *QueueConfig) Clean() error {
	if len(qc.LookupdAddrs) == 0 {
		if len(qc.NSQDAddrs) == 0 {
			return fmt.Errorf("Both LookupdAddresses and NSQDAddresses are missing")
		}
	}
	return nil
}

func ConnectConsumer(qc *QueueConfig, consumer *nsq.Consumer) error {
	var err error

	// Connect consumers to NSQLookupd
	if qc.LookupdAddrs != nil || len(qc.LookupdAddrs) != 0 {
		log.Info("Connecting Consumer to the following NSQLookupds %s", qc.LookupdAddrs)
		err = consumer.ConnectToNSQLookupds(qc.LookupdAddrs)
		if err != nil {
			return err
		}
	}
	// Connect consumers to NSQD
	fmt.Println(qc.NSQDAddrs)
	fmt.Println(qc.NSQDAddrs)
	fmt.Println(qc.NSQDAddrs)
	if qc.NSQDAddrs != nil || len(qc.NSQDAddrs) != 0 {
		log.Info("Connecting Consumer to the following NSQDs %s", qc.NSQDAddrs)
		err = consumer.ConnectToNSQDs(qc.NSQDAddrs)
		if err != nil {
			return err
		}
	}
	return nil
}
