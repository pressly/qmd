package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/bitly/go-nsq"
	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/mux"
)

type ScriptRequest struct {
	ID          int
	Script      string            `json:"script"`
	Args        []string          `json:"args"`
	Files       map[string]string `json:"files"`
	CallbackURL string            `json:"callback_url"`
}

func GetAllScripts(w http.ResponseWriter, r *http.Request) {
	// Get a list of all the scripts in script folder.

	log.Printf("Received GET %s\n", r.URL)

	// Open and parse whitelist
	p := path.Join(config.Worker.ScriptDir, config.Worker.WhiteList)
	file, err := os.Open(p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	var buf bytes.Buffer
	buf.WriteString("[")

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	buf.WriteString(scanner.Text())
	for scanner.Scan() {
		buf.WriteString(", ")
		buf.WriteString(scanner.Text())
	}
	buf.WriteString("]")

	w.Header().Set("Content-Type", "application/json")
	w.Write(buf.Bytes())
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
	id, err := getRedisID()
	if err != nil {
		log.Println(err)
	}
	sr.ID = id

	vars := mux.Vars(r)
	sr.Script = vars["name"]

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

	// LRANGE returns an array of json strings
	reply, err := redis.Strings(conn.Do("ZRANGE", params["name"], 0, -1))
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	buf.WriteString("[")

	if len(reply) > 0 {
		buf.WriteString(reply[0])
		for i := 1; i < len(reply); i++ {
			buf.WriteString(", ")
			buf.WriteString(reply[i])
		}
	}
	buf.WriteString("]")
	w.Header().Set("Content-Type", "application/json")
	w.Write(buf.Bytes())
}

func GetLog(w http.ResponseWriter, r *http.Request) {
	// Retrieve a specific log of a specific script.

	log.Printf("Received GET %s\n", r.URL)

	conn := redisDB.Get()
	defer conn.Close()

	params := mux.Vars(r)
	script := params["name"]
	id := params["id"]

	reply, err := redis.Strings(conn.Do("ZRANGEBYSCORE", script, id, id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(reply[0]))
}
