// This is a QMD client that reads job results from a specified channel
// and posts them to the keen.io servers.

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/bitly/go-nsq"
	"github.com/pressly/qmd"
)

var (
	topic      = flag.String("topic", "result", "nsq topic")
	channel    = flag.String("channel", "qmd_to_keen", "nsq channel")
	throughput = flag.Int("throughput", 100, "max handlers and messages allowed in flight")

	// keen.io settings
	keenVersion         = flag.String("api-version", "3.0", "keen.io API version")
	keenProjectID       = flag.String("project-id", "", "keen.io project ID")
	keenEventCollection = flag.String("event-collection", "", "keen.io project event collection")
	keenAPIKey          = flag.String("api-key", "", "keen.io API write key")

	nsqdAddrs    = qmd.StringFlagArray{}
	lookupdAddrs = qmd.StringFlagArray{}

	keenAddress string
)

func init() {
	flag.Var(&nsqdAddrs, "nsqd-addresses", "nsqd address for consumption (may be given multiple times)")
	flag.Var(&lookupdAddrs, "lookupd-addresses", "lookupd address for consumption, takes precedence over nsqd (may be given multiple times)")
}

func keenHandler(m *nsq.Message) error {
	var v map[string]interface{}
	if err := json.Unmarshal(m.Body, &v); err != nil {
		return err
	}
	for k := range v {
		if k == "exec_log" {
			delete(v, k)
		}
		if k == "output" {
			delete(v, k)
		}
	}
	data, err := json.Marshal(&v)
	if err != nil {
		return err
	}
	resp, err := http.Post(keenAddress, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	script, id := v["script"], v["id"]
	log.Printf("Sent %s of %s to collection %s", id, script, *keenEventCollection)
	return nil
}

func main() {
	var err error

	flag.Parse()

	if *topic == "" || *channel == "" {
		log.Fatalf("--topic and --channel are required")
	}

	if len(nsqdAddrs) == 0 && len(lookupdAddrs) == 0 {
		log.Fatalf("--nsqd-addresses or --lookupd-addresses required")
	}

	if *keenProjectID == "" {
		log.Fatalf("--project-id required")
	}

	if *keenEventCollection == "" {
		log.Fatalf("--event-collection required")
	}

	if *keenAPIKey == "" {
		log.Fatalf("--api-key required")
	}

	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)

	keenAddress = fmt.Sprintf(
		"https://api.keen.io/%s/projects/%s/events/%s?api_key=%s",
		*keenVersion,
		*keenProjectID,
		*keenEventCollection,
		*keenAPIKey,
	)

	cfg := nsq.NewConfig()
	cfg.MaxInFlight = *throughput
	consumer, err := nsq.NewConsumer(*topic, *channel, cfg)
	if err != nil {
		log.Fatalf(err.Error())
	}

	consumer.AddConcurrentHandlers(nsq.HandlerFunc(keenHandler), *throughput)

	var qc qmd.QueueConfig
	qc.NSQDAddrs = nsqdAddrs
	qc.LookupdAddrs = lookupdAddrs

	if err = qmd.ConnectConsumer(&qc, consumer); err != nil {
		log.Fatalf(err.Error())
	}

	for {
		select {
		case <-consumer.StopChan:
			return
		case <-termChan:
			consumer.Stop()
		}
	}
}
