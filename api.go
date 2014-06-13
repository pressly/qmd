package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/bitly/go-nsq"
	"github.com/garyburd/redigo/redis"
	"github.com/zenazn/goji/web"
)

type ScriptRequest struct {
	ID          int
	Script      string            `json:"script"`
	Args        []string          `json:"args"`
	Files       map[string]string `json:"files"`
	CallbackURL string            `json:"callback_url"`
}

// Handle the root route, also useful as a heartbeat.
func ServiceRoot(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		body, err := ioutil.ReadAll(r.Body)
		if err == nil {
			log.Info("Callback: %s", body)
		}
	}
	w.Write([]byte("."))
}

// Get a list of all the scripts in script folder.
func GetAllScripts(w http.ResponseWriter, r *http.Request) {

	// Open and parse whitelist
	p := path.Join(config.Worker.ScriptDir, config.Worker.WhiteList)
	file, err := os.Open(p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	list := make([]string, 0)
	for scanner.Scan() {
		list = append(list, fmt.Sprintf("\"%s\"", scanner.Text()))
	}
	var buf bytes.Buffer
	buf.WriteString("[")
	buf.WriteString(strings.Join(list, ", "))
	buf.WriteString("]")

	w.Header().Set("Content-Type", "application/json")
	w.Write(buf.Bytes())
}

// Reload the whitelist of scripts.
func ReloadScripts(w http.ResponseWriter, r *http.Request) {
	p := path.Join(config.Worker.ScriptDir, config.Worker.WhiteList)

	doneChan := make(chan *nsq.ProducerTransaction)
	err := producer.PublishAsync("reload", []byte(p), doneChan)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	<-doneChan

	var buf bytes.Buffer
	buf.WriteString("Reload request sent")

	w.Header().Set("Content-Type", "application/json")
	w.Write(buf.Bytes())
}

// Send details to queue for execution.
func RunScript(c web.C, w http.ResponseWriter, r *http.Request) {
	// Parse the request
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error(err.Error())
	}
	var sr ScriptRequest
	err = json.Unmarshal(body, &sr)
	if err != nil {
		log.Error(err.Error())
	}
	id, err := getRedisID()
	if err != nil {
		log.Error(err.Error())
	}
	sr.ID = id
	sr.Script = c.URLParams["name"]

	// Queue up the request
	doneChan := make(chan *nsq.ProducerTransaction)
	data, err := json.Marshal(sr)
	if err != nil {
		log.Error(err.Error())
	}
	err = producer.PublishAsync(config.Topic, data, doneChan)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	<-doneChan
	log.Debug("Request queued as %d", sr.ID)

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

// Retrieve all logs for a specific script.
func GetAllLogs(c web.C, w http.ResponseWriter, r *http.Request) {
	conn := redisDB.Get()
	defer conn.Close()

	// LRANGE returns an array of json strings
	reply, err := redis.Strings(conn.Do("ZRANGE", c.URLParams["name"], 0, -1))
	if err != nil {
		log.Error(err.Error())
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

// Retrieve a specific log of a specific script.
func GetLog(c web.C, w http.ResponseWriter, r *http.Request) {
	conn := redisDB.Get()
	defer conn.Close()

	script := c.URLParams["name"]
	id := c.URLParams["id"]

	reply, err := redis.Strings(conn.Do("ZRANGEBYSCORE", script, id, id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(reply[0]))
}
