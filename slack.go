package qmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/op/go-logging"
)

type SlackNotifier struct {
	WebhookURL string
	Channel    string
	Prefix     string
}

type slackPayload struct {
	Channel  string `json:"channel"`
	Username string `json:"username"`
	Text     string `json:"text"`
}

func (s *SlackNotifier) Log(level logging.Level, calldepth int, rec *logging.Record) error {
	if level != logging.INFO {
		return nil
	}

	payload, err := json.Marshal(slackPayload{
		Channel:  s.Channel,
		Username: "QMD",
		Text:     s.Prefix + rec.Message(),
	})
	if err != nil {
		return err
	}

	resp, err := http.Post(s.WebhookURL, "application/json", bytes.NewReader(payload))
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("coudn't POST to slack webhook %v", s.WebhookURL)
	}

	return nil
}
