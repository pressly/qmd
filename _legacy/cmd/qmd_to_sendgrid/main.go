// This is a QMD client that reads job results from a specified channel
// and posts them to SendGrid

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bitly/go-nsq"
	"github.com/pressly/qmd"
	"github.com/sendgrid/sendgrid-go"
)

var (
	topic      = flag.String("topic", "result", "nsq topic")
	channel    = flag.String("channel", "qmd_to_sendgrid", "nsq channel")
	throughput = flag.Int("throughput", 100, "max handlers and messages allowed in flight")
	sender     = flag.String("sender", "qmd@example.com", "email of sender")
	recipients = qmd.StringFlagArray{}
	statuses   = qmd.StringFlagArray{}

	// SendGrid settings
	sgUser = flag.String("sendgrid-user", "", "username of Sendgrid account")
	sgKey  = flag.String("sendgrid-key", "", "key of Sendgrid account")

	nsqdAddrs    = qmd.StringFlagArray{}
	lookupdAddrs = qmd.StringFlagArray{}

	subjectTemplate = "Job: %s for script: %s with status: %s"
	msgTemplate     = "Job: %s for script: %s with args: %v\nStarted at %s, took %s seconds, and has a status of %s.\nSee full log below:\n---------------------------------\n\n%s"
	allowedStatuses = make(map[string]bool)
	sg              *sendgrid.SGClient
)

func init() {
	flag.Var(&nsqdAddrs, "nsqd-addresses", "nsqd address for consumption (may be given multiple times)")
	flag.Var(&lookupdAddrs, "lookupd-addresses", "lookupd address for consumption, takes precedence over nsqd (may be given multiple times)")
	flag.Var(&recipients, "recipients", "recipient(s) of email (may be given multiple times)")
	flag.Var(&statuses, "statuses", "status(es) to send email for (may be given multiple times)")
}

func main() {
	var err error

	flag.Parse()

	if *topic == "" || *channel == "" {
		log.Fatalf("--topic and --channel are required")
	}

	if *sender == "" {
		log.Fatalf("--sender required")
	}

	if len(recipients) == 0 {
		log.Fatalf("--recipients required")
	}

	if len(statuses) == 0 {
		log.Fatalf("--statuses required")
	}

	if len(nsqdAddrs) == 0 && len(lookupdAddrs) == 0 {
		log.Fatalf("--nsqd-addresses or --lookupd-addresses required")
	}

	if *sgUser == "" {
		log.Fatalf("--sendgrid-user required")
	}

	if *sgKey == "" {
		log.Fatalf("--sendgrid-key required")
	}

	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)

	sg = sendgrid.NewSendGridClient(*sgUser, *sgKey)

	for _, status := range statuses.Slice() {
		allowedStatuses[status] = true
	}
	log.Printf("Allowed statuses: %v", allowedStatuses)

	log.Printf("Recipients: %v", recipients.Slice())

	cfg := nsq.NewConfig()
	cfg.MaxInFlight = *throughput
	consumer, err := nsq.NewConsumer(*topic, *channel, cfg)
	if err != nil {
		log.Fatalf(err.Error())
	}

	consumer.AddConcurrentHandlers(nsq.HandlerFunc(sendgridHandler), *throughput)

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

func sendgridHandler(m *nsq.Message) error {
	var v map[string]interface{}
	var err error

	if err = json.Unmarshal(m.Body, &v); err != nil {
		return err
	}

	status := v["status"].(string)
	log.Printf("Received log with status %s", status)
	if _, exist := allowedStatuses[status]; !exist {
		return nil
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

	subject := fmt.Sprintf(
		subjectTemplate,
		id,
		script,
		status,
	)
	var logText bytes.Buffer
	json.Indent(&logText, m.Body, "", "  ")
	content := fmt.Sprintf(
		msgTemplate,
		id,
		script,
		args.String(),
		startTime.Format(time.UnixDate),
		duration,
		status,
		logText.Bytes(),
	)

	mail := sendgrid.NewMail()
	mail.SetFrom(*sender)
	for _, email := range recipients.Slice() {
		mail.AddTo(email)
	}
	mail.SetSubject(subject)
	mail.SetText(content)
	if r := sg.Send(mail); r == nil {
		log.Printf("Sent message for log %s to %s", id, recipients.String())
	} else {
		log.Print(r)
	}
	return nil
}
