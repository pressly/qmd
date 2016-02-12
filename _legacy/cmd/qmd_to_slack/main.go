// This is a QMD client that reads job results from a specified channel
// and posts them to a Slack channel

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
	"time"

	"github.com/bitly/go-nsq"
	"github.com/pressly/qmd"
)

var (
	topic      = flag.String("topic", "result", "nsq topic")
	channel    = flag.String("channel", "qmd_to_slack", "nsq channel")
	throughput = flag.Int("throughput", 100, "max handlers and messages allowed in flight")
	logAddr    = flag.String("log-address", "", "base address of QMD logs")

	// Slack settings
	slackWebhook   = flag.String("webhook", "", "unique Slack webhook URL")
	slackChannel   = flag.String("slack-channel", "", "Slack channel name")
	slackBotName   = flag.String("slack-bot-name", "qmd_to_slack", "Slack bot name")
	slackIconURL   = flag.String("slack-icon-url", "", "Slack icon URL")
	slackIconEmoji = flag.String("slack-icon-emoji", ":wrench:", "Slack icon emoji")

	nsqdAddrs    = qmd.StringFlagArray{}
	lookupdAddrs = qmd.StringFlagArray{}

	msgTemplate = "Job: %s for script: %s with args: %v\nStarted at %s, took %s seconds, and has a status of %s.\nSee full log: <http://%s/scripts/%s/logs/%s>."
)

func init() {
	flag.Var(&nsqdAddrs, "nsqd-addresses", "nsqd address for consumption (may be given multiple times)")
	flag.Var(&lookupdAddrs, "lookupd-addresses", "lookupd address for consumption, takes precedence over nsqd (may be given multiple times)")
}

func slackHandler(m *nsq.Message) error {
	var msg slackMessage
	var v map[string]interface{}
	var err error

	if err = json.Unmarshal(m.Body, &v); err != nil {
		return err
	}

	id := v["id"].(string)
	script := v["script"].(string)
	var args bytes.Buffer
	args.WriteString("[ ")
	if list, ok := v["args"].([]interface{}); ok {
		length := len(list)
		for i, arg := range list {
			args.WriteString(arg.(string))
			if i < length-1 {
				args.WriteString(", ")
			}
		}
	}
	args.WriteString(" ]")
	startTime, err := time.Parse(
		time.RFC3339,
		v["start_time"].(string),
	)
	if err != nil {
		return err
	}
	duration := v["duration"].(string)
	status := v["status"].(string)

	msg.Channel = *slackChannel
	msg.Username = *slackBotName
	if *slackIconURL != "" {
		msg.IconURL = *slackIconURL
	} else {
		msg.IconEmoji = *slackIconEmoji
	}
	msg.Text = fmt.Sprintf(
		msgTemplate,
		id,
		script,
		args.String(),
		startTime.Format(time.UnixDate),
		duration,
		status,
		*logAddr,
		script,
		id,
	)
	data, err := json.Marshal(&msg)
	if err != nil {
		return err
	}

	resp, err := http.Post(*slackWebhook, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	log.Printf("Sent message for log %s to %s", id, *slackChannel)
	return nil
}

func main() {
	var err error

	flag.Parse()

	if *topic == "" || *channel == "" {
		log.Fatalf("--topic and --channel are required")
	}
	if *logAddr == "" {
		log.Fatalf("--log-address required")
	}
	if len(nsqdAddrs) == 0 && len(lookupdAddrs) == 0 {
		log.Fatalf("--nsqd-addresses or --lookupd-addresses required")
	}

	if *slackWebhook == "" {
		log.Fatalf("--webhook required")
	}
	if *slackChannel == "" {
		log.Fatalf("--slack-channel required")
	}
	if *slackBotName == "" {
		log.Fatalf("--slack-bot-name required")
	}

	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)

	cfg := nsq.NewConfig()
	cfg.MaxInFlight = *throughput
	consumer, err := nsq.NewConsumer(*topic, *channel, cfg)
	if err != nil {
		log.Fatalf(err.Error())
	}

	consumer.AddConcurrentHandlers(nsq.HandlerFunc(slackHandler), *throughput)

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

type slackMessage struct {
	Channel   string `json:"channel"`
	Username  string `json:"username"`
	IconURL   string `json:"icon_url"`
	IconEmoji string `json:"icon_emoji"`
	Text      string `json:"text"`
}
