package qmd

import (
	"crypto/sha1"
	"fmt"
	"strings"

	"code.google.com/p/go-uuid/uuid"
)

// Unique ID generator

func NewID() string {
	h := sha1.New()
	h.Write(uuid.NewRandom())
	return fmt.Sprintf("%x", h.Sum(nil))
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

func (a *StringFlagArray) Slice() []string {
	var tmp []string
	for _, s := range strings.Split(a.String(), ",") {
		tmp = append(tmp, s)
	}
	return tmp
}
