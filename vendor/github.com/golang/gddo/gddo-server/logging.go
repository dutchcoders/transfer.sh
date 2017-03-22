// Copyright 2016 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package main

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/logging"
	"github.com/golang/gddo/database"
)

// newGCELogger returns a handler that wraps h but logs each request
// using Google Cloud Logging service.
func newGCELogger(cli *logging.Logger) *GCELogger {
	return &GCELogger{cli}
}

type GCELogger struct {
	cli *logging.Logger
}

// LogEvent creates an entry in Cloud Logging to record user's behavior. We should only
// use this to log events we are interested in. General request logs are handled by GAE
// automatically in request_log and stderr.
func (g *GCELogger) LogEvent(w http.ResponseWriter, r *http.Request, content interface{}) {
	const sessionCookieName = "GODOC_ORG_SESSION_ID"
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		// Generates a random session id and sends it in response.
		rs, err := randomString()
		if err != nil {
			log.Println("error generating a random session id: ", err)
			return
		}
		// This cookie is intentionally short-lived and contains no information
		// that might identify the user.  Its sole purpose is to tie query
		// terms and destination pages together to measure search quality.
		cookie = &http.Cookie{
			Name:    sessionCookieName,
			Value:   rs,
			Expires: time.Now().Add(time.Hour),
		}
		http.SetCookie(w, cookie)
	}

	// We must not record the client's IP address, or any other information
	// that might compromise the user's privacy.
	payload := map[string]interface{}{
		sessionCookieName: cookie.Value,
		"path":            r.URL.RequestURI(),
		"method":          r.Method,
		"referer":         r.Referer(),
	}
	if pkgs, ok := content.([]database.Package); ok {
		payload["packages"] = pkgs
	}

	// Log queues the entry to its internal buffer, or discarding the entry
	// if the buffer was full.
	g.cli.Log(logging.Entry{
		Payload: payload,
	})
}

func randomString() (string, error) {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	return hex.EncodeToString(b), err
}
