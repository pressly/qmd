package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
)

type Client struct {
	ServiceUrl string
	S3Bucket   string
	Async      bool
}

func New(ServiceUrl, S3Bucket string, Async bool) *Client {
	return &Client{config}
}

func (q *Client) Request(script, callbackUrl string, options map[string]interface{}) error {
	u, err := url.Parse(q.config.ServiceUrl)
	if err != nil {
		return err
	}
	u.Path = "/scripts/" + script

	if q.config.Async {
		if callbackUrl != "" {
			options["callback_url"] = callbackUrl
		}
	}

	jsonStr, err := json.Marshal(options)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", u.String(), bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if !q.config.Async && callbackUrl != "" {
		req, _ := http.NewRequest("POST", callbackUrl, resp.Body)
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{}
		r, err := client.Do(req)
		if err != nil {
			return err
		}

		if r.StatusCode != 200 {
			body, _ := ioutil.ReadAll(r.Body)
			return errors.New("callback failed: " + string(body))
		}
	}

	return nil
}
