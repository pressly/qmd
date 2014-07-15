package main

import (
	"crypto/sha1"
	"fmt"
	"time"
)

var idChan chan string = make(chan string)

func NewID() string {
	generateID()
	return <-idChan
}

func generateID() {
	h := sha1.New()
	c := []byte(time.Now().String())
	h.Write(c)
	idChan <- fmt.Sprintf("%x", h.Sum(nil))
}
