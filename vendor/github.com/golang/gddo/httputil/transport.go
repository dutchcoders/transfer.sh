// Copyright 2015 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

// This file implements a http.RoundTripper that authenticates
// requests issued against api.github.com endpoint.

package httputil

import (
	"log"
	"net/http"
	"net/url"
	"os"

	"cloud.google.com/go/compute/metadata"
)

// AuthTransport is an implementation of http.RoundTripper that authenticates
// with the GitHub API.
//
// When both a token and client credentials are set, the latter is preferred.
type AuthTransport struct {
	UserAgent    string
	Token        string
	ClientID     string
	ClientSecret string
	Base         http.RoundTripper
}

// NewAuthTransport gives new AuthTransport created with GitHub credentials
// read from GCE metadata when the metadata server is accessible (we're on GCE)
// or read from environment varialbes otherwise.
func NewAuthTransport(base http.RoundTripper) *AuthTransport {
	if metadata.OnGCE() {
		return NewAuthTransportFromMetadata(base)
	}
	return NewAuthTransportFromEnvironment(base)
}

// NewAuthTransportFromEnvironment gives new AuthTransport created with GitHub
// credentials read from environment variables.
func NewAuthTransportFromEnvironment(base http.RoundTripper) *AuthTransport {
	return &AuthTransport{
		UserAgent:    os.Getenv("USER_AGENT"),
		Token:        os.Getenv("GITHUB_TOKEN"),
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		Base:         base,
	}
}

// NewAuthTransportFromMetadata gives new AuthTransport created with GitHub
// credentials read from GCE metadata.
func NewAuthTransportFromMetadata(base http.RoundTripper) *AuthTransport {
	return &AuthTransport{
		UserAgent:    gceAttr("user-agent"),
		Token:        gceAttr("github-token"),
		ClientID:     gceAttr("github-client-id"),
		ClientSecret: gceAttr("github-client-secret"),
		Base:         base,
	}
}

// RoundTrip implements the http.RoundTripper interface.
func (t *AuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var reqCopy *http.Request
	if t.UserAgent != "" {
		reqCopy = copyRequest(req)
		reqCopy.Header.Set("User-Agent", t.UserAgent)
	}
	if req.URL.Host == "api.github.com" {
		switch {
		case t.ClientID != "" && t.ClientSecret != "":
			if reqCopy == nil {
				reqCopy = copyRequest(req)
			}
			if reqCopy.URL.RawQuery == "" {
				reqCopy.URL.RawQuery = "client_id=" + t.ClientID + "&client_secret=" + t.ClientSecret
			} else {
				reqCopy.URL.RawQuery += "&client_id=" + t.ClientID + "&client_secret=" + t.ClientSecret
			}
		case t.Token != "":
			if reqCopy == nil {
				reqCopy = copyRequest(req)
			}
			reqCopy.Header.Set("Authorization", "token "+t.Token)
		}
	}
	if reqCopy != nil {
		return t.base().RoundTrip(reqCopy)
	}
	return t.base().RoundTrip(req)
}

// CancelRequest cancels an in-flight request by closing its connection.
func (t *AuthTransport) CancelRequest(req *http.Request) {
	type canceler interface {
		CancelRequest(req *http.Request)
	}
	if cr, ok := t.base().(canceler); ok {
		cr.CancelRequest(req)
	}
}

func (t *AuthTransport) base() http.RoundTripper {
	if t.Base != nil {
		return t.Base
	}
	return http.DefaultTransport
}

func gceAttr(name string) string {
	s, err := metadata.ProjectAttributeValue(name)
	if err != nil {
		log.Printf("error querying metadata for %q: %s", name, err)
		return ""
	}
	return s
}

func copyRequest(req *http.Request) *http.Request {
	req2 := new(http.Request)
	*req2 = *req
	req2.URL = new(url.URL)
	*req2.URL = *req.URL
	req2.Header = make(http.Header, len(req.Header))
	for k, s := range req.Header {
		req2.Header[k] = append([]string(nil), s...)
	}
	return req2
}
