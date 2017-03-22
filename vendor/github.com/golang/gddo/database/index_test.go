// Copyright 2013 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package database

import (
	"reflect"
	"sort"
	"testing"

	"github.com/golang/gddo/doc"
)

var indexTests = []struct {
	pdoc  *doc.Package
	terms []string
}{
	{&doc.Package{
		ImportPath:  "strconv",
		ProjectRoot: "",
		ProjectName: "Go",
		Name:        "strconv",
		Synopsis:    "Package strconv implements conversions to and from string representations of basic data types.",
		Doc:         "Package strconv implements conversions to and from string representations\nof basic data types.",
		Imports:     []string{"errors", "math", "unicode/utf8"},
		Funcs:       []*doc.Func{{}},
	},
		[]string{
			"bas",
			"convert",
			"dat",
			"import:errors",
			"import:math",
			"import:unicode/utf8",
			"project:go",
			"repres",
			"strconv",
			"string",
			"typ"},
	},
	{&doc.Package{
		ImportPath:  "github.com/user/repo/dir",
		ProjectRoot: "github.com/user/repo",
		ProjectName: "go-oauth",
		ProjectURL:  "https://github.com/user/repo/",
		Name:        "dir",
		Synopsis:    "Package dir implements a subset of the OAuth client interface as defined in RFC 5849.",
		Doc: "Package oauth implements a subset of the OAuth client interface as defined in RFC 5849.\n\n" +
			"This package assumes that the application writes request URL paths to the\nnetwork using " +
			"the encoding implemented by the net/url URL RequestURI method.\n" +
			"The HTTP client in the standard net/http package uses this encoding.",
		IsCmd: false,
		Imports: []string{
			"bytes",
			"crypto/hmac",
			"crypto/sha1",
			"encoding/base64",
			"encoding/binary",
			"errors",
			"fmt",
			"io",
			"io/ioutil",
			"net/http",
			"net/url",
			"regexp",
			"sort",
			"strconv",
			"strings",
			"sync",
			"time",
		},
		TestImports: []string{"bytes", "net/url", "testing"},
		Funcs:       []*doc.Func{{}},
	},
		[]string{
			"all:",
			"5849", "cly", "defin", "dir", "github.com", "go",
			"import:bytes", "import:crypto/hmac", "import:crypto/sha1",
			"import:encoding/base64", "import:encoding/binary", "import:errors",
			"import:fmt", "import:io", "import:io/ioutil", "import:net/http",
			"import:net/url", "import:regexp", "import:sort", "import:strconv",
			"import:strings", "import:sync", "import:time", "interfac",
			"oau", "project:github.com/user/repo", "repo", "rfc", "subset", "us",
		},
	},
}

func TestDocTerms(t *testing.T) {
	for _, tt := range indexTests {
		score := documentScore(tt.pdoc)
		terms := documentTerms(tt.pdoc, score)
		sort.Strings(terms)
		sort.Strings(tt.terms)
		if !reflect.DeepEqual(terms, tt.terms) {
			t.Errorf("documentTerms(%s) ->\n got: %#v\nwant: %#v", tt.pdoc.ImportPath, terms, tt.terms)
		}
	}
}

var vendorPatTests = []struct {
	path  string
	match bool
}{
	{"camlistore.org/third_party/github.com/user/repo", true},
	{"camlistore.org/third_party/dir", false},
	{"camlistore.org/third_party", false},
	{"camlistore.org/xthird_party/github.com/user/repo", false},
	{"camlistore.org/third_partyx/github.com/user/repo", false},

	{"example.org/_third_party/github.com/user/repo/dir", true},
	{"example.org/_third_party/dir", false},

	{"github.com/user/repo/Godeps/_workspace/src/github.com/user/repo", true},
	{"github.com/user/repo/Godeps/_workspace/src/dir", false},

	{"github.com/user/repo", false},
}

func TestVendorPat(t *testing.T) {
	for _, tt := range vendorPatTests {
		match := vendorPat.MatchString(tt.path)
		if match != tt.match {
			t.Errorf("match(%q) = %v, want %v", tt.path, match, match)
		}
	}
}

var synopsisTermTests = []struct {
	synopsis string
	terms    []string
}{
	{
		"Package foo implements bar.",
		[]string{"bar", "foo"},
	},
	{
		"Package foo provides bar.",
		[]string{"bar", "foo"},
	},
	{
		"The foo package provides bar.",
		[]string{"bar", "foo"},
	},
	{
		"Package foo contains an implementation of bar.",
		[]string{"bar", "foo", "impl"},
	},
	{
		"Package foo is awesome",
		[]string{"awesom", "foo"},
	},
	{
		"The foo package is awesome",
		[]string{"awesom", "foo"},
	},
	{
		"The foo command is awesome",
		[]string{"awesom", "foo"},
	},
	{
		"Command foo is awesome",
		[]string{"awesom", "foo"},
	},
	{
		"The foo package",
		[]string{"foo"},
	},
	{
		"Package foo",
		[]string{"foo"},
	},
	{
		"Command foo",
		[]string{"foo"},
	},
	{
		"Package",
		[]string{},
	},
	{
		"Command",
		[]string{},
	},
}

func TestSynopsisTerms(t *testing.T) {
	for _, tt := range synopsisTermTests {
		terms := make(map[string]bool)
		collectSynopsisTerms(terms, tt.synopsis)

		actual := termSlice(terms)
		expected := tt.terms
		sort.Strings(actual)
		sort.Strings(expected)
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("%q ->\n got: %#v\nwant: %#v", tt.synopsis, actual, expected)
		}
	}
}
