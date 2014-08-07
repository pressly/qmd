package qmd

import (
	"crypto/sha1"
	"fmt"
	"strings"
	"time"
)

// Unique ID generator

var idChan chan string = make(chan string)

func NewID() string {
	go generateID()
	return <-idChan
}

func generateID() {
	h := sha1.New()
	c := []byte(time.Now().String())
	for {
		h.Write(c)
		idChan <- fmt.Sprintf("%x", h.Sum(nil))
	}
}

// Flag string arrays

type StringFlagArray []string

func (a *StringFlagArray) Set(s string) error {
	*a = append(*a, s)
	return nil
}

func (a *StringFlagArray) String() string {
	return strings.Join(*a, ",")
}
