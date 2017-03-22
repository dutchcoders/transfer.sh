// Copyright 2013 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package main

import (
	"testing"
)

var isBrowseURLTests = []struct {
	s          string
	importPath string
	ok         bool
}{
	{"https://github.com/garyburd/gddo/blob/master/doc/code.go", "github.com/garyburd/gddo/doc", true},
	{"https://github.com/garyburd/go-oauth/blob/master/.gitignore", "github.com/garyburd/go-oauth", true},
	{"https://github.com/garyburd/gddo/issues/154", "github.com/garyburd/gddo", true},
	{"https://bitbucket.org/user/repo/src/bd0b661a263e/p1/p2?at=default", "bitbucket.org/user/repo/p1/p2", true},
	{"https://bitbucket.org/user/repo/src", "bitbucket.org/user/repo", true},
	{"https://bitbucket.org/user/repo", "bitbucket.org/user/repo", true},
	{"https://github.com/user/repo", "github.com/user/repo", true},
	{"https://github.com/user/repo/tree/master/p1", "github.com/user/repo/p1", true},
	{"http://code.google.com/p/project", "code.google.com/p/project", true},
}

func TestIsBrowseURL(t *testing.T) {
	for _, tt := range isBrowseURLTests {
		importPath, ok := isBrowseURL(tt.s)
		if tt.ok {
			if importPath != tt.importPath || ok != true {
				t.Errorf("IsBrowseURL(%q) = %q, %v; want %q %v", tt.s, importPath, ok, tt.importPath, true)
			}
		} else if ok {
			t.Errorf("IsBrowseURL(%q) = %q, %v; want _, false", tt.s, importPath, ok)
		}
	}
}
