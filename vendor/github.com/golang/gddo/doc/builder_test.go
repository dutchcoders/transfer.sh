// Copyright 2013 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package doc

import (
	"go/ast"
	"testing"
)

var badSynopsis = []string{
	"+build !release",
	"COPYRIGHT Jimmy Bob",
	"### Markdown heading",
	"-*- indent-tabs-mode: nil -*-",
	"vim:set ts=2 sw=2 et ai ft=go:",
}

func TestBadSynopsis(t *testing.T) {
	for _, s := range badSynopsis {
		if synopsis(s) != "" {
			t.Errorf(`synopsis(%q) did not return ""`, s)
		}
	}
}

const readme = `
    $ go get github.com/user/repo/pkg1
    [foo](http://gopkgdoc.appspot.com/pkg/github.com/user/repo/pkg2)
    [foo](http://go.pkgdoc.org/github.com/user/repo/pkg3)
    [foo](http://godoc.org/github.com/user/repo/pkg4)
    <http://go.pkgdoc.org/github.com/user/repo/pkg5>
    [foo](http://godoc.org/github.com/user/repo/pkg6#Export)
    http://gowalker.org/github.com/user/repo/pkg7
    Build Status: [![Build Status](https://drone.io/github.com/user/repo1/status.png)](https://drone.io/github.com/user/repo1/latest)
    'go get example.org/package1' will install package1.
    (http://go.pkgdoc.org/example.org/package2 "Package2's documentation on GoPkgDoc").
    import "example.org/package3"
`

var expectedReferences = []string{
	"github.com/user/repo/pkg1",
	"github.com/user/repo/pkg2",
	"github.com/user/repo/pkg3",
	"github.com/user/repo/pkg4",
	"github.com/user/repo/pkg5",
	"github.com/user/repo/pkg6",
	"github.com/user/repo/pkg7",
	"github.com/user/repo1",
	"example.org/package1",
	"example.org/package2",
	"example.org/package3",
}

func TestReferences(t *testing.T) {
	references := make(map[string]bool)
	addReferences(references, []byte(readme))
	for _, r := range expectedReferences {
		if !references[r] {
			t.Errorf("missing %s", r)
		}
		delete(references, r)
	}
	for r := range references {
		t.Errorf("extra %s", r)
	}
}

var simpleImporterTests = []struct {
	path string
	name string
}{
	// Last element with .suffix removed.
	{"example.com/user/name.git", "name"},
	{"example.com/user/name.svn", "name"},
	{"example.com/user/name.hg", "name"},
	{"example.com/user/name.bzr", "name"},
	{"example.com/name.v0", "name"},
	{"example.com/user/repo/name.v11", "name"},

	// Last element with "go" prefix or suffix removed.
	{"github.com/user/go-name", "name"},
	{"github.com/user/go.name", "name"},
	{"github.com/user/name.go", "name"},
	{"github.com/user/name-go", "name"},

	// Special cases for popular repos.
	{"code.google.com/p/biogo.name", "name"},
	{"code.google.com/p/google-api-go-client/name/v3", "name"},

	// Use last element of path.
	{"example.com/user/name.other", "name.other"},
	{"example.com/.v0", ".v0"},
	{"example.com/user/repo.v2/name", "name"},
	{"github.com/user/namev0", "namev0"},
	{"github.com/user/goname", "goname"},
	{"github.com/user/namego", "namego"},
	{"github.com/user/name", "name"},
	{"name", "name"},
	{"user/name", "name"},
}

func TestSimpleImporter(t *testing.T) {
	for _, tt := range simpleImporterTests {
		m := make(map[string]*ast.Object)
		obj, _ := simpleImporter(m, tt.path)
		if obj.Name != tt.name {
			t.Errorf("simpleImporter(%q) = %q, want %q", tt.path, obj.Name, tt.name)
		}
	}
}
