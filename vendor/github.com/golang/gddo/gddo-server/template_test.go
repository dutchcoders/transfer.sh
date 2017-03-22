// Copyright 2014 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package main

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestFlashMessages(t *testing.T) {
	resp := httptest.NewRecorder()

	expectedMessages := []flashMessage{
		{ID: "a", Args: []string{"one"}},
		{ID: "b", Args: []string{"two", "three"}},
		{ID: "c", Args: []string{}},
	}

	setFlashMessages(resp, expectedMessages)
	req := &http.Request{Header: http.Header{"Cookie": {strings.Split(resp.Header().Get("Set-Cookie"), ";")[0]}}}

	actualMessages := getFlashMessages(resp, req)
	if !reflect.DeepEqual(actualMessages, expectedMessages) {
		t.Errorf("got messages %+v, want %+v", actualMessages, expectedMessages)
	}
}
