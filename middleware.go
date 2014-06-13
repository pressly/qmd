package main

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/zenazn/goji/web"
)

func BasicAuth(c *web.C, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Basic ") {
			pleaseAuth(w)
			return
		}

		pass, err := decodeAuth(auth[6:])

		if err != nil || pass != config.auth.authString {
			pleaseAuth(w)
			return
		}
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func pleaseAuth(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="qmd"`)
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte("Please authenticate with the proper API user and password!\n"))
}

// This function decodes the given string.
// Here is where we would put any decryption if required.
func decodeAuth(auth string) (string, error) {
	pass, err := base64.StdEncoding.DecodeString(auth)
	return string(pass), err
}

// Removes the last trailing slash if it exists.
// So /scripts/ == /scripts but /scripts// == /scripts/
func AllowSlash(c *web.C, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
