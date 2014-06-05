package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/bitly/go-nsq"
	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/feeds"
	"github.com/gorilla/mux"
)

type ScriptRequest struct {
	ID     string
	Script string   `json:"script"`
	Args   []string `json:"args"`
	Dir    string   `json:"dir"`
}

func GetAllScripts(w http.ResponseWriter, r *http.Request) {
	// Get a list of all the scripts in script folder.

	// How??
	// Send to queue and have it returned to me?
}

func RunScript(w http.ResponseWriter, r *http.Request) {
	// Send details to queue for execution.

	log.Printf("Received POST %s\n", r.URL)

	// Parse the request
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
	}
	var sr ScriptRequest
	err = json.Unmarshal(body, &sr)
	if err != nil {
		log.Println(err)
	}
	id := feeds.NewUUID().String()
	sr.ID = id

	// Queue up the request
	doneChan := make(chan *nsq.ProducerTransaction)
	data, err := json.Marshal(sr)
	if err != nil {
		log.Println(err)
	}
	err = producer.PublishAsync(config.Topic, data, doneChan)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	<-doneChan
	log.Printf("Request queued as %s\n", sr.ID)

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func GetAllLogs(w http.ResponseWriter, r *http.Request) {
	// Retrieve all logs for a specific script.

	log.Printf("Received GET %s\n", r.URL)

	conn := redisDB.Get()
	defer conn.Close()

	params := mux.Vars(r)

	reply, err := redis.Strings(conn.Do("LRANGE", params["name"], 0, -1))
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(reply)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func GetLog(w http.ResponseWriter, r *http.Request) {
	// Retrieve a specific log of a specific script.

	log.Printf("Received GET %s\n", r.URL)

	conn := redisDB.Get()
	defer conn.Close()

	params := mux.Vars(r)

	reply, err := conn.Do("GET", params["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(reply.([]byte))
}
