package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/bitly/go-nsq"
	"github.com/gorilla/feeds"
)

type ScriptRequest struct {
	ID   string
	Body []byte
}

func GetAllScripts(w http.ResponseWriter, r *http.Request) {
	// Get a list of all the scripts in script folder.

	// How??
	// Send to queue and have it returned to me?
}

func RunScript(w http.ResponseWriter, r *http.Request) {
	// Send details to queue for execution.

	// Parse the request
	id := feeds.NewUUID().String()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
	}
	sr := ScriptRequest{id, body}

	// Queue up the request
	doneChan := make(chan *nsq.ProducerTransaction)
	data, err := json.Marshal(sr)
	if err != nil {
		fmt.Println(err)
	}
	err = producer.PublishAsync(config.Topic, data, doneChan)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println(<-doneChan)

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func GetAllLogs(w http.ResponseWriter, r *http.Request) {
	// Retrieve all logs for a specific script.

	// Fetch from Redis
}

func GetLog(w http.ResponseWriter, r *http.Request) {
	// Retrieve a specific log of a specific script.

	// Fetch from Redis
}
