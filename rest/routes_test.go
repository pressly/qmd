package rest_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pressly/qmd"
	"github.com/pressly/qmd/config"
	"github.com/pressly/qmd/rest"
)

func TestPing(t *testing.T) {
	conf, _ := config.New("../etc/qmd.conf.sample")

	qmd := &qmd.Qmd{
		Config:             conf,
		DB:                 nil,
		Queue:              nil,
		ClosingListenQueue: make(chan struct{}),
		ClosingWorkers:     make(chan struct{}),
	}

	ts := httptest.NewServer(rest.Routes(qmd))
	defer ts.Close()

	res, err := http.Get(ts.URL + "/ping")
	if err != nil {
		t.Error(err)
	}

	dot, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Error(err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		t.Error("unexpected response status code")
	}

	if string(dot) != "." {
		t.Error("unexpected response body")
	}
}
