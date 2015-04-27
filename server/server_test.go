package server

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pressly/qmd/config"
)

func TestPing(t *testing.T) {
	conf, _ := config.New("../etc/qmd.conf.sample")

	apiHandler := New(conf)
	ts := httptest.NewServer(apiHandler)
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
