package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/bitly/go-nsq"
	"github.com/garyburd/redigo/redis"
	"github.com/zenazn/goji/web"
)

type ScriptRequest struct {
	ID          int               `json:"id"`
	Script      string            `json:"script"`
	Args        []string          `json:"args,omitempty"`
	Files       map[string]string `json:"files,omitempty"`
	CallbackURL string            `json:"callback_url"`
}

// Handle and reverse proxy the admin route to nsqadmin.
func AdminProxy(w http.ResponseWriter, r *http.Request) {
	targetURL := config.AdminAddr
	if targetURL == "" {
		http.Error(w, "No admin panel found", http.StatusNotFound)
		return
	}
	if !strings.Contains(targetURL, "http://") {
		targetURL = fmt.Sprintf("%s%s", "http://", "0.0.0.0:4171")
	}
	u, err := url.Parse(targetURL)
	if err != nil {
		w.Write([]byte(err.Error()))
	}
	p := httputil.NewSingleHostReverseProxy(u)
	p.ServeHTTP(w, r)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var sr ScriptRequest
	err = json.Unmarshal(body, &sr)
	if err != nil {
		log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id, err := getRedisID()
	if err != nil {
		log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sr.ID = id
	sr.Script = c.URLParams["name"]

	// Queue up the request
	doneChan := make(chan *nsq.ProducerTransaction)
	data, err := json.Marshal(sr)
	if err != nil {
		log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = producer.PublishAsync(config.Topic, data, doneChan)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	<-doneChan
	log.Info("Request queued as %d", sr.ID)

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

	length := len(reply)
	if length > 0 {
		buf.WriteString(reply[length-1])
		for i := length - 2; i > 0; i-- {
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
