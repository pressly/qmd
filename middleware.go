package main

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/zenazn/goji/web"
)

func RequestLogger(c *web.C, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		log.Info("sup...")
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func BasicAuth(c *web.C, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if config.Auth.Enabled {
			auth := r.Header.Get("Authorization")
			if !strings.HasPrefix(auth, "Basic ") {
				unauthorized(w)
				return
			}

			pass, err := decodeAuth(auth[6:])
			if err != nil || pass != config.Auth.AuthString {
				unauthorized(w)
				return
			}
		}
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func unauthorized(w http.ResponseWriter) {
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
		if len(r.URL.Path) > 1 {
			r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
		}
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
