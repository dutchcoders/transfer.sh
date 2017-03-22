// Copyright 2013 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package httputil

import (
	"net/http"
	"testing"
)

func TestCacheBusters(t *testing.T) {
	cbs := &CacheBusters{Handler: http.FileServer(http.Dir("."))}

	token := cbs.Get("/buster_test.go")
	if token == "" {
		t.Errorf("could not extract token from http.FileServer")
	}

	var ss StaticServer
	cbs = &CacheBusters{Handler: ss.FileHandler("buster_test.go")}

	token = cbs.Get("/xxx")
	if token == "" {
		t.Errorf("could not extract token from StaticServer FileHandler")
	}
}
