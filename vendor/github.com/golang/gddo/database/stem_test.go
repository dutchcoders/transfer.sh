// Copyright 2013 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package database

import (
	"testing"
)

var stemTests = []struct {
	s, expected string
}{
	{"html", "html"},
	{"strings", "string"},
	{"ballroom", "ballroom"},
	{"mechanicalization", "mech"},
	{"pragmaticality", "pragm"},
	{"rationalistically", "rat"},
}

func TestStem(t *testing.T) {
	for _, tt := range stemTests {
		actual := stem(tt.s)
		if actual != tt.expected {
			t.Errorf("stem(%q) = %q, want %q", tt.s, actual, tt.expected)
		}
	}
}
